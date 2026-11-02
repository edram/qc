package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"

	"github.com/edram/qi/internal/aiqicha"
	"github.com/edram/qi/internal/http"
	"github.com/edram/qi/internal/models"
)

const searchProviderAiqicha searchProviderName = "aiqicha"

type searchAiqicha struct {
	api *aiqicha.Client
}

func newSearchAiqicha(profile, userAgent string) *searchAiqicha {
	return &searchAiqicha{api: aiqicha.New(aiqicha.Options{Profile: profile, UserAgent: userAgent})}
}

type aiqichaEnterpriseSearchResponse struct {
	Status  *int   `json:"status"`
	Message string `json:"msg"`
	Data    struct {
		Total       *int                `json:"totalNum"`
		TotalFound  json.Number         `json:"totalNumFound"`
		Enterprises []aiqichaEnterprise `json:"resultList"`
		Limit       *struct {
			Status   string `json:"limitStatus"`
			UserType string `json:"userType"`
		} `json:"limitForward"`
	} `json:"data"`
}

type aiqichaEnterprise struct {
	ID                  string `json:"pid"`
	HKID                string `json:"hkid"`
	IsHKCompany         int    `json:"isHkComp"`
	Name                string `json:"entName"`
	TitleName           string `json:"titleName"`
	LegalRepresentative string `json:"legalPerson"`
	TitleLegal          string `json:"titleLegal"`
	Status              string `json:"openStatus"`
	EstablishedDate     string `json:"validityFrom"`
	CreditCode          string `json:"regNo"`
	RegisteredCapital   string `json:"regCap"`
	Address             string `json:"domicile"`
	TitleAddress        string `json:"titleDomicile"`
	Phone               string `json:"telephone"`
	Email               string `json:"email"`
	LogoURL             string `json:"entLogo"`
	Labels              []struct {
		Key   string `json:"key"`
		Value struct {
			Text string `json:"text"`
		} `json:"value"`
	} `json:"newLabels"`
	Risk struct {
		Direct     *int `json:"selfRiskTotal"`
		Associated *int `json:"unionRiskTotal"`
	} `json:"risk"`
}

// Request parameters follow the website's enterpriseSearch operation:
// https://xinpub.cdn.bcebos.com/aiqicha/static/1793259024/js/s.b337ba3a42b.js
func (s *searchAiqicha) SearchEnterprises(ctx context.Context, query string, filter enterpriseSearchFilter) (models.EnterpriseSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: query is required")
	}
	for _, option := range []struct {
		name  string
		count int
	}{
		{"--match", len(filter.Fields)}, {"--area", len(filter.Areas)},
		{"--industry", len(filter.Industries)}, {"--status", len(filter.Statuses)},
	} {
		if option.count > 0 {
			return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: %s is not supported; use keyword search without filters", option.name)
		}
	}
	params := url.Values{"q": {query}, "t": {""}, "p": {"1"}, "s": {"10"}, "o": {"0"}, "f": {"{}"}}
	referer := "https://www.aiqicha.com/s?" + url.Values{"q": {query}, "t": {"0"}}.Encode()
	var response aiqichaEnterpriseSearchResponse
	if err := s.api.GetJSON(ctx, "/s/advanceFilterAjax?"+params.Encode(), &response, http.WithRequestHeader("Referer", referer)); err != nil {
		return models.EnterpriseSearchResult{}, err
	}
	if response.Status == nil {
		return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: missing status")
	}
	if *response.Status != 0 {
		return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: status %d: %s", *response.Status, response.Message)
	}
	if limit := response.Data.Limit; limit != nil {
		if limit.UserType == "nologin" || limit.Status == "0" {
			return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha: login required; log in to Aiqicha and run 'qc auth import --provider aiqicha'")
		}
		if limit.Status != "" {
			return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: access restricted (limitStatus %s)", limit.Status)
		}
	}
	var total int
	if response.Data.Total != nil {
		total = *response.Data.Total
	} else {
		var err error
		total, err = strconv.Atoi(string(response.Data.TotalFound))
		if err != nil {
			return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: missing search results")
		}
	}
	if total > 0 && response.Data.Enterprises == nil {
		return models.EnterpriseSearchResult{}, fmt.Errorf("aiqicha enterprise search: missing search results")
	}
	result := models.EnterpriseSearchResult{Total: total, Enterprises: make([]models.Enterprise, 0, len(response.Data.Enterprises))}
	for _, enterprise := range response.Data.Enterprises {
		result.Enterprises = append(result.Enterprises, normalizeAiqichaEnterprise(enterprise))
	}
	return result, nil
}

func normalizeAiqichaEnterprise(enterprise aiqichaEnterprise) models.Enterprise {
	result := models.Enterprise{
		ID:                  enterprise.ID,
		DetailURL:           "https://www.aiqicha.com/company_detail_" + enterprise.ID,
		Name:                aiqichaText(enterprise.TitleName, enterprise.Name),
		LegalRepresentative: aiqichaText(enterprise.TitleLegal, enterprise.LegalRepresentative),
		Status:              enterprise.Status,
		EstablishedDate:     enterprise.EstablishedDate,
		CreditCode:          aiqichaText(enterprise.CreditCode),
		RegisteredCapital:   enterprise.RegisteredCapital,
		Address:             aiqichaText(enterprise.TitleAddress, enterprise.Address),
		Phone:               enterprise.Phone,
		Email:               enterprise.Email,
		LogoURL:             enterprise.LogoURL,
	}
	if enterprise.HKID != "" {
		result.ID = enterprise.HKID
	}
	if enterprise.HKID != "" || enterprise.IsHKCompany == 1 {
		result.DetailURL = "https://www.aiqicha.com/hkcompany_detail_" + result.ID
	}
	if len(result.CreditCode) != 18 {
		result.RegistrationNumber, result.CreditCode = result.CreditCode, ""
	}
	for _, label := range enterprise.Labels {
		if label.Key != "opening" && label.Value.Text != "" {
			result.Tags = append(result.Tags, aiqichaText(label.Value.Text))
		}
	}
	if enterprise.Risk.Direct != nil || enterprise.Risk.Associated != nil {
		result.Risk = &models.EnterpriseRiskSummary{}
		if enterprise.Risk.Direct != nil {
			result.Risk.Direct = &models.EnterpriseRiskScope{Count: *enterprise.Risk.Direct}
		}
		if enterprise.Risk.Associated != nil {
			result.Risk.Associated = &models.EnterpriseRiskScope{Count: *enterprise.Risk.Associated}
		}
	}
	return result
}

func aiqichaText(values ...string) string {
	for _, value := range values {
		if value != "" {
			value = strings.ReplaceAll(value, "<em>", "")
			value = strings.ReplaceAll(value, "</em>", "")
			return html.UnescapeString(value)
		}
	}
	return ""
}

func (*searchAiqicha) SearchPeople(context.Context, string, personSearchFilter) ([]models.Person, error) {
	return nil, nil
}
