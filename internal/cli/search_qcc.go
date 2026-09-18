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

	"github.com/edram/qi/internal/models"
	"github.com/edram/qi/internal/qcc"
)

const searchSourceQCC searchSourceName = "qcc"

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
	Result  []struct {
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
	} `json:"Result"`
}

type qccPersonSearchRequest struct {
	Key       string `json:"key"`
	AreaInfo  string `json:"areaInfo"`
	PageIndex int    `json:"pageIndex"`
	Status    []int  `json:"status"`
	Name      string `json:"name"`
	PageSize  int    `json:"pageSize"`
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

func (s *searchQCC) SearchEnterprises(ctx context.Context, query string, filter enterpriseSearchFilter) ([]models.Enterprise, error) {
	searchKey, err := qccEnterpriseSearchKey(query, filter.Fields)
	if err != nil {
		return nil, err
	}
	encodedFilter, err := qccEnterpriseSearchFilter(filter)
	if err != nil {
		return nil, err
	}
	request := qccEnterpriseSearchRequest{
		SearchKey:   searchKey,
		PageIndex:   1,
		PageSize:    20,
		SearchIndex: "multicondition",
		Filter:      encodedFilter,
	}

	var result qccEnterpriseSearchResponse
	if err := s.api.PostJSON(ctx, "/api/search/searchMulti", request, &result); err != nil {
		return nil, err
	}
	if err := qccStatusError("enterprise search", result.Status, result.Message); err != nil {
		return nil, err
	}

	enterprises := make([]models.Enterprise, 0, len(result.Result))
	for _, enterprise := range result.Result {
		var establishedDate string
		if enterprise.StartDate != 0 {
			establishedDate = time.UnixMilli(enterprise.StartDate).In(qccDateLocation).Format(time.DateOnly)
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
		})
	}
	return enterprises, nil
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
	request := qccPersonSearchRequest{
		Key:       query,
		AreaInfo:  areaInfo,
		PageIndex: 1,
		Status:    []int{0, 1},
		Name:      query,
		PageSize:  18,
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
		Areas    []areaFilter `json:"r,omitempty"`
		Statuses []string     `json:"s,omitempty"`
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
	for _, status := range filter.Statuses {
		codes, ok := qccStatusCodes[status]
		if !ok {
			return "", fmt.Errorf("unsupported enterprise status %q", status)
		}
		encoded.Statuses = append(encoded.Statuses, codes...)
	}
	if len(encoded.Areas) == 0 && len(encoded.Statuses) == 0 {
		return "", nil
	}
	data, err := json.Marshal(encoded)
	return string(data), err
}
