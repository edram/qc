package cli

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/edram/qi/internal/http"
	"github.com/edram/qi/internal/models"
)

// The detail page loads basic information and the first page of related lists
// from /detail/basicAllDataAjax. Restricted fields remain masked or absent.
// https://xinpub.cdn.bcebos.com/aiqicha/static/1793259024/js/detailJudicialInfo.7b504469508.js
type aiqichaEnterpriseDetailResponse struct {
	Status  *int                        `json:"status"`
	Message string                      `json:"msg"`
	Data    aiqichaEnterpriseDetailData `json:"data"`
}

type aiqichaEnterpriseDetailData struct {
	Basic         aiqichaEnterpriseBasicData                      `json:"basicData"`
	LoginRequired bool                                            `json:"isDidiwei"`
	Shareholders  aiqichaDetailList[aiqichaEnterpriseShareholder] `json:"shareholdersData"`
	Employees     aiqichaDetailList[aiqichaEnterpriseEmployee]    `json:"directorsData"`
	Branches      aiqichaDetailList[aiqichaEnterpriseBranch]      `json:"branchsData"`
}

type aiqichaDetailList[T any] struct {
	List  []T `json:"list"`
	Limit *struct {
		Status   string `json:"limitStatus"`
		UserType string `json:"userType"`
	} `json:"limitForward"`
}

func (list aiqichaDetailList[T]) available() []T {
	if list.Limit != nil && (list.Limit.Status != "" || list.Limit.UserType == "nologin") {
		return nil
	}
	return list.List
}

type aiqichaEnterpriseShareholder struct {
	StockID               string `json:"stockId"`
	EnterpriseID          string `json:"pid"`
	PersonID              string `json:"personId"`
	Name                  string `json:"name"`
	Percent               string `json:"subRate"`
	SubscribedCapital     string `json:"subMoney"`
	SubscribedCapitalDate string `json:"subDate"`
	Labels                []struct {
		Value struct {
			Text string `json:"text"`
		} `json:"value"`
	} `json:"labels"`
}

type aiqichaEnterpriseEmployee struct {
	ID       string `json:"personId"`
	Name     string `json:"name"`
	Position string `json:"title"`
}

type aiqichaEnterpriseBranch struct {
	ID                  string `json:"pid"`
	Name                string `json:"entName"`
	LegalRepresentative string `json:"legalPerson"`
}

type aiqichaEnterpriseBasicData struct {
	Name                  string `json:"entName"`
	CreditCode            string `json:"unifiedCode"`
	RegistrationNumber    string `json:"regCode"`
	LegalRepresentative   string `json:"legalPerson"`
	Status                string `json:"openStatus"`
	EstablishedDate       string `json:"startDate"`
	Address               string `json:"regAddr"`
	RegisteredCapital     string `json:"regCapital"`
	Phone                 string `json:"telephone"`
	Email                 string `json:"email"`
	LogoURL               string `json:"entLogo"`
	BusinessScope         string `json:"scope"`
	CompanyType           string `json:"entType"`
	RegistrationAuthority string `json:"authority"`
	OrganizationCode      string `json:"orgNo"`
	TaxNumber             string `json:"taxNo"`
	OperatingTerm         string `json:"openTime"`
	CheckDate             string `json:"annualDate"`
	ActualCapital         string `json:"paidinCapital"`
	Industry              string `json:"industry"`
	InsuranceInfo         struct {
		ReportURL string `json:"reportUrl"`
	} `json:"insuranceInfo"`
}

func (s *searchAiqicha) GetEnterpriseDetail(ctx context.Context, id string) (models.EnterpriseDetail, error) {
	id = strings.TrimSpace(id)
	if id == "" || strings.IndexFunc(id, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return models.EnterpriseDetail{}, fmt.Errorf("aiqicha enterprise view: a numeric enterprise ID is required; Hong Kong details are not supported")
	}
	pageURL := "https://www.aiqicha.com/company_detail_" + id
	var response aiqichaEnterpriseDetailResponse
	if err := s.api.GetJSON(ctx, "/detail/basicAllDataAjax?"+url.Values{"pid": {id}}.Encode(), &response,
		http.WithRequestHeader("Referer", pageURL), http.WithRequestHeader("Zx-Open-Url", pageURL)); err != nil {
		return models.EnterpriseDetail{}, err
	}
	if response.Status == nil {
		return models.EnterpriseDetail{}, fmt.Errorf("aiqicha enterprise view: missing status")
	}
	if *response.Status != 0 {
		return models.EnterpriseDetail{}, fmt.Errorf("aiqicha enterprise view: status %d: %s", *response.Status, response.Message)
	}
	if response.Data.LoginRequired {
		return models.EnterpriseDetail{}, fmt.Errorf("aiqicha: login required; log in to Aiqicha and run 'qc auth import --provider aiqicha'")
	}
	if aiqichaDetailText(response.Data.Basic.Name) == "" {
		return models.EnterpriseDetail{}, fmt.Errorf("aiqicha enterprise view: missing company details")
	}
	// The API can return status 0 with another company's data when page context
	// or signing is rejected. Check the company reference when one is available.
	if report, err := url.Parse(response.Data.Basic.InsuranceInfo.ReportURL); err == nil {
		if reportID := report.Query().Get("pid"); reportID != "" && reportID != id {
			return models.EnterpriseDetail{}, fmt.Errorf("aiqicha enterprise view: response belongs to a different enterprise")
		}
	}
	return normalizeAiqichaEnterpriseDetail(response.Data, id), nil
}

func normalizeAiqichaEnterpriseDetail(data aiqichaEnterpriseDetailData, id string) models.EnterpriseDetail {
	detail := normalizeAiqichaEnterpriseBasicData(data.Basic, id)
	for _, item := range data.Shareholders.available() {
		shareholder := models.EnterpriseShareholder{
			ID: aiqichaText(item.StockID, item.EnterpriseID, item.PersonID), Name: aiqichaDetailText(item.Name),
			Percent: aiqichaDetailText(item.Percent), SubscribedCapital: aiqichaDetailText(item.SubscribedCapital),
			SubscribedCapitalDate: aiqichaDetailText(item.SubscribedCapitalDate),
		}
		for _, label := range item.Labels {
			if text := aiqichaDetailText(label.Value.Text); text != "" {
				shareholder.Tags = append(shareholder.Tags, text)
			}
		}
		detail.Shareholders = append(detail.Shareholders, shareholder)
	}
	for _, item := range data.Employees.available() {
		detail.Employees = append(detail.Employees, models.EnterpriseEmployee{
			ID: item.ID, Name: aiqichaDetailText(item.Name), Position: aiqichaDetailText(item.Position),
		})
	}
	for _, item := range data.Branches.available() {
		detail.Branches = append(detail.Branches, models.EnterpriseBranch{
			ID: item.ID, Name: aiqichaDetailText(item.Name), LegalRepresentative: aiqichaDetailText(item.LegalRepresentative),
		})
	}
	return detail
}

func normalizeAiqichaEnterpriseBasicData(basic aiqichaEnterpriseBasicData, id string) models.EnterpriseDetail {
	detail := models.EnterpriseDetail{
		Enterprise: models.Enterprise{
			ID: id, DetailURL: "https://www.aiqicha.com/company_detail_" + id,
			Name: aiqichaDetailText(basic.Name), CreditCode: aiqichaDetailText(basic.CreditCode),
			RegistrationNumber:  aiqichaDetailText(basic.RegistrationNumber),
			LegalRepresentative: aiqichaDetailText(basic.LegalRepresentative), Status: aiqichaDetailText(basic.Status),
			EstablishedDate: aiqichaDetailText(basic.EstablishedDate), Address: aiqichaDetailText(basic.Address),
			RegisteredCapital: aiqichaDetailText(basic.RegisteredCapital), Phone: aiqichaDetailText(basic.Phone),
			Email: aiqichaDetailText(basic.Email), LogoURL: aiqichaDetailText(basic.LogoURL),
		},
		BusinessScope: aiqichaDetailText(basic.BusinessScope), CompanyType: aiqichaDetailText(basic.CompanyType),
		RegistrationAuthority: aiqichaDetailText(basic.RegistrationAuthority), OrganizationCode: aiqichaDetailText(basic.OrganizationCode),
		TaxNumber: aiqichaDetailText(basic.TaxNumber), CheckDate: aiqichaDetailText(basic.CheckDate),
		ActualCapital: aiqichaDetailText(basic.ActualCapital),
	}
	if start, end, ok := strings.Cut(basic.OperatingTerm, "至"); ok {
		detail.TermStart, detail.TermEnd = aiqichaDetailText(start), aiqichaDetailText(end)
	}
	if industry := aiqichaDetailText(basic.Industry); industry != "" {
		detail.Industry = &models.EnterpriseIndustry{Name: industry}
	}
	return detail
}

func aiqichaDetailText(value string) string {
	value = strings.TrimSpace(aiqichaText(value))
	if value == "-" {
		return ""
	}
	return value
}
