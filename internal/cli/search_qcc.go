package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edram/qi/internal/industries"
	"github.com/edram/qi/internal/models"
	"github.com/edram/qi/internal/qcc"
	"github.com/edram/qi/internal/regions"
)

const searchProviderQCC searchProviderName = "qcc"

// QCC encodes calendar dates as Unix milliseconds at China-local midnight.
var qccDateLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

type searchQCC struct {
	api *qcc.Client
}

type qccEnterpriseSearchRequest struct {
	SearchKey   string `json:"searchKey"`
	PageIndex   int    `json:"pageIndex"`
	PageSize    int    `json:"pageSize"`
	SearchIndex string `json:"searchIndex"`
	Filter      string `json:"filter,omitempty"`
}

type qccEnterpriseSearchResponse struct {
	Status  int    `json:"Status"`
	Message string `json:"Message"`
	Paging  struct {
		TotalRecords int `json:"TotalRecords"`
	} `json:"Paging"`
	GroupItems []qccEnterpriseSearchGroup `json:"GroupItems"`
	Result     []struct {
		KeyNo         string `json:"KeyNo"`
		Name          string `json:"Name"`
		No            string `json:"No"`
		CreditCode    string `json:"CreditCode"`
		OperName      string `json:"OperName"`
		Status        string `json:"Status"`
		StartDate     int64  `json:"StartDate"`
		Address       string `json:"Address"`
		RegistCapi    string `json:"RegistCapi"`
		ContactNumber string `json:"ContactNumber"`
		Email         string `json:"Email"`
		ImageURL      string `json:"ImageUrl"`
		TagsInfoV2    []struct {
			Name string `json:"Name"`
		} `json:"TagsInfoV2"`
	} `json:"Result"`
}

type qccEnterpriseSearchGroup struct {
	Key   string `json:"key"`
	Items []struct {
		Value string `json:"value"`
		Count int    `json:"count"`
		Name  string `json:"desc"`
	} `json:"items"`
}

type qccPersonSearchRequest struct {
	Key          string `json:"key"`
	AreaInfo     string `json:"areaInfo"`
	IndustryInfo string `json:"industryInfo"`
	PageIndex    int    `json:"pageIndex"`
	Status       []int  `json:"status"`
	Name         string `json:"name"`
	PageSize     int    `json:"pageSize"`
}

type qccPersonSearchResponse struct {
	Status  int    `json:"Status"`
	Message string `json:"Message"`
	Result  []struct {
		ID          string `json:"Id"`
		Name        string `json:"Name"`
		MainCompany struct {
			KeyNo   string `json:"KeyNo"`
			Company string `json:"Company"`
		} `json:"MainCompany"`
		Introduction string `json:"Intro"`
		ImageURL     string `json:"ImageUrl"`
		Count        int    `json:"Count"`
		CountInfo    struct {
			PartnerCount int `json:"PartnerCount"`
		} `json:"CountInfo"`
		RoleDescription string `json:"RoleDesc"`
	} `json:"Result"`
}

func newSearchQCC(profile, userAgent string) search {
	return &searchQCC{api: qcc.New(qcc.Options{Profile: profile, UserAgent: userAgent})}
}

func (s *searchQCC) SearchEnterprises(ctx context.Context, query string, filter enterpriseSearchFilter) (models.EnterpriseSearchResult, error) {
	searchKey, err := qccEnterpriseSearchKey(query, filter.Fields)
	if err != nil {
		return models.EnterpriseSearchResult{}, err
	}
	encodedFilter, err := qccEnterpriseSearchFilter(filter)
	if err != nil {
		return models.EnterpriseSearchResult{}, err
	}
	request := qccEnterpriseSearchRequest{
		SearchKey:   searchKey,
		PageIndex:   1,
		PageSize:    20,
		SearchIndex: "multicondition",
		Filter:      encodedFilter,
	}

	var response qccEnterpriseSearchResponse
	if err := s.api.PostJSON(ctx, "/api/search/searchMulti", request, &response); err != nil {
		return models.EnterpriseSearchResult{}, err
	}
	if err := qccStatusError("enterprise search", response.Status, response.Message); err != nil {
		return models.EnterpriseSearchResult{}, err
	}

	enterprises := make([]models.Enterprise, 0, len(response.Result))
	for _, enterprise := range response.Result {
		var establishedDate string
		if enterprise.StartDate != 0 {
			establishedDate = time.UnixMilli(enterprise.StartDate).In(qccDateLocation).Format(time.DateOnly)
		}
		tags := make([]string, 0, len(enterprise.TagsInfoV2))
		for _, tag := range enterprise.TagsInfoV2 {
			tags = append(tags, tag.Name)
		}
		enterprises = append(enterprises, models.Enterprise{
			ID:                  enterprise.KeyNo,
			DetailURL:           "https://www.qcc.com/firm/" + enterprise.KeyNo + ".html",
			Name:                qccText(enterprise.Name),
			RegistrationNumber:  enterprise.No,
			CreditCode:          qccText(enterprise.CreditCode),
			LegalRepresentative: enterprise.OperName,
			Status:              enterprise.Status,
			EstablishedDate:     establishedDate,
			Address:             enterprise.Address,
			RegisteredCapital:   enterprise.RegistCapi,
			Phone:               enterprise.ContactNumber,
			Email:               enterprise.Email,
			LogoURL:             enterprise.ImageURL,
			Tags:                tags,
		})
	}
	aggregations, err := qccEnterpriseAggregations(response.GroupItems)
	if err != nil {
		return models.EnterpriseSearchResult{}, err
	}
	return models.EnterpriseSearchResult{
		Total:        response.Paging.TotalRecords,
		Enterprises:  enterprises,
		Aggregations: aggregations,
	}, nil
}

func qccEnterpriseAggregations(groups []qccEnterpriseSearchGroup) (models.EnterpriseSearchAggregations, error) {
	areaCatalog, err := regions.Load(regions.ProviderQCC)
	if err != nil {
		return models.EnterpriseSearchAggregations{}, err
	}
	industryCatalog, err := industries.Load(industries.ProviderQCC)
	if err != nil {
		return models.EnterpriseSearchAggregations{}, err
	}

	var aggregations models.EnterpriseSearchAggregations
	for _, group := range groups {
		for _, item := range group.Items {
			name := item.Name
			switch group.Key {
			case "province":
				if area, resolveErr := areaCatalog.Resolve(item.Value); resolveErr == nil {
					name = area.Name
				}
				aggregations.Provinces = append(aggregations.Provinces, models.EnterpriseSearchAggregation{
					Code: item.Value, Name: name, Count: item.Count,
				})
			case "industrycode":
				if industry, resolveErr := industryCatalog.Resolve(item.Value); resolveErr == nil {
					name = industry.Name
				}
				aggregations.Industries = append(aggregations.Industries, models.EnterpriseSearchAggregation{
					Code: item.Value, Name: name, Count: item.Count,
				})
			}
		}
	}
	return aggregations, nil
}

func (s *searchQCC) SearchPeople(ctx context.Context, query string, filter personSearchFilter) ([]models.Person, error) {
	areaInfo := ""
	if strings.TrimSpace(filter.Area) != "" {
		area, err := qcc.ResolveArea(filter.Area)
		if err != nil {
			return nil, err
		}
		areaInfo = area.Code
	}
	industryInfo := ""
	if strings.TrimSpace(filter.Industry) != "" {
		catalog, err := industries.Load(industries.ProviderQCC)
		if err != nil {
			return nil, err
		}
		industry, err := catalog.Resolve(filter.Industry)
		if err != nil {
			return nil, err
		}
		industryInfo = industry.Path
	}
	request := qccPersonSearchRequest{
		Key:          query,
		AreaInfo:     areaInfo,
		IndustryInfo: industryInfo,
		PageIndex:    1,
		Status:       []int{0, 1},
		Name:         query,
		PageSize:     18,
	}

	var result qccPersonSearchResponse
	if err := s.api.PostJSON(ctx, "/api/bigsearch/searchPerson", request, &result); err != nil {
		return nil, err
	}
	if err := qccStatusError("people search", result.Status, result.Message); err != nil {
		return nil, err
	}

	people := make([]models.Person, 0, len(result.Result))
	for _, person := range result.Result {
		people = append(people, models.Person{
			ID:                  person.ID,
			DetailURL:           "https://www.qcc.com/pl/" + person.ID + ".html",
			Name:                qccText(person.Name),
			MainCompanyID:       person.MainCompany.KeyNo,
			MainCompanyName:     qccText(person.MainCompany.Company),
			Role:                qccText(person.RoleDescription),
			RelatedCompanyCount: person.Count,
			PartnerCount:        person.CountInfo.PartnerCount,
			Introduction:        qccText(person.Introduction),
			AvatarURL:           person.ImageURL,
		})
	}
	return people, nil
}

func qccStatusError(operation string, status int, message string) error {
	if status == http.StatusOK {
		return nil
	}
	return fmt.Errorf("qcc %s: status %d: %s", operation, status, message)
}

func qccText(value string) string {
	value = strings.ReplaceAll(value, "<em>", "")
	value = strings.ReplaceAll(value, "</em>", "")
	return html.UnescapeString(value)
}

type qccAreaFilterCode string

func (code qccAreaFilterCode) MarshalJSON() ([]byte, error) {
	if _, err := strconv.ParseInt(string(code), 10, 64); err == nil {
		return []byte(code), nil
	}
	return json.Marshal(string(code))
}

var qccEnterpriseFieldCodes = map[enterpriseSearchField]string{
	enterpriseSearchFieldName:                "onlyname",
	enterpriseSearchFieldScope:               "scope",
	enterpriseSearchFieldIntroduction:        "introduction",
	enterpriseSearchFieldAddress:             "address",
	enterpriseSearchFieldBrand:               "product",
	enterpriseSearchFieldLegalRepresentative: "opername",
	enterpriseSearchFieldPatent:              "patent",
	enterpriseSearchFieldTrademark:           "featurelist",
	enterpriseSearchFieldShareholder:         "promoterlist",
	enterpriseSearchFieldKeyPersonnel:        "employeelist",
}

var qccStatusCodes = map[enterpriseSearchStatus][]string{
	enterpriseSearchStatusActive:       {"20", "10", "50"},
	enterpriseSearchStatusMoved:        {"60"},
	enterpriseSearchStatusEstablishing: {"117"},
	enterpriseSearchStatusCancelled:    {"99", "91"},
	enterpriseSearchStatusRevoked:      {"90"},
}

func qccEnterpriseSearchKey(query string, fields []enterpriseSearchField) (string, error) {
	if len(fields) == 0 {
		fields = []enterpriseSearchField{enterpriseSearchFieldName}
	}
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		code, ok := qccEnterpriseFieldCodes[field]
		if !ok {
			return "", fmt.Errorf("unsupported enterprise match field %q", field)
		}
		values[code] = query
	}
	data, err := json.Marshal(values)
	return string(data), err
}

func qccEnterpriseSearchFilter(filter enterpriseSearchFilter) (string, error) {
	type areaFilter struct {
		Province string              `json:"pr"`
		Codes    []qccAreaFilterCode `json:"cc,omitempty"`
	}
	encoded := struct {
		Areas      []areaFilter `json:"r,omitempty"`
		Industries []string     `json:"i,omitempty"`
		Statuses   []string     `json:"s,omitempty"`
	}{}
	for _, area := range filter.Areas {
		selection, err := qcc.ResolveArea(area)
		if err != nil {
			return "", err
		}
		encodedArea := areaFilter{Province: selection.ProvinceCode}
		if selection.Code != selection.ProvinceCode {
			encodedArea.Codes = []qccAreaFilterCode{qccAreaFilterCode(selection.Code)}
		}
		encoded.Areas = append(encoded.Areas, encodedArea)
	}
	if len(filter.Industries) > 0 {
		catalog, err := industries.Load(industries.ProviderQCC)
		if err != nil {
			return "", err
		}
		for _, name := range filter.Industries {
			industry, err := catalog.Resolve(name)
			if err != nil {
				return "", err
			}
			encoded.Industries = append(encoded.Industries, industry.ProviderCode)
		}
	}
	for _, status := range filter.Statuses {
		codes, ok := qccStatusCodes[status]
		if !ok {
			return "", fmt.Errorf("unsupported enterprise status %q", status)
		}
		encoded.Statuses = append(encoded.Statuses, codes...)
	}
	if len(encoded.Areas) == 0 && len(encoded.Industries) == 0 && len(encoded.Statuses) == 0 {
		return "", nil
	}
	data, err := json.Marshal(encoded)
	return string(data), err
}
