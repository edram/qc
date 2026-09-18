package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type qccAreaFilterCode string

func (code qccAreaFilterCode) MarshalJSON() ([]byte, error) {
	if _, err := strconv.ParseInt(string(code), 10, 64); err == nil {
		return []byte(code), nil
	}
	return json.Marshal(string(code))
}

// These values mirror the visible filters and payload codes used by QCC's search page.
// See: https://www.qcc.com/web/search
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

var qccIndustryCodes = map[string]string{
	"农、林、牧、渔业": "A",
	"采矿业":      "B",
	"制造业":      "C",
	"电力、热力、燃气及水生产和供应业": "D",
	"建筑业":             "E",
	"批发和零售业":          "F",
	"交通运输、仓储和邮政业":     "G",
	"住宿和餐饮业":          "H",
	"信息传输、软件和信息技术服务业": "I",
	"金融业":             "J",
	"房地产业":            "K",
	"租赁和商务服务业":        "L",
	"科学研究和技术服务业":      "M",
	"水利、环境和公共设施管理业":   "N",
	"居民服务、修理和其他服务业":   "O",
	"教育":        "P",
	"卫生和社会工作":   "Q",
	"文化、体育和娱乐业": "R",
	"公共管理、社会保障和社会组织": "S",
	"国际组织": "T",
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
		selection, err := qccEnterpriseArea(area)
		if err != nil {
			return "", err
		}
		encodedArea := areaFilter{Province: selection.Province}
		if code := selection.code(); code != selection.Province {
			encodedArea.Codes = []qccAreaFilterCode{qccAreaFilterCode(code)}
		}
		encoded.Areas = append(encoded.Areas, encodedArea)
	}
	for _, industry := range filter.Industries {
		code, err := qccFilterCode(industry, qccIndustryCodes, "industry")
		if err != nil {
			return "", err
		}
		encoded.Industries = append(encoded.Industries, code)
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

func qccFilterCode(value string, values map[string]string, name string) (string, error) {
	value = strings.TrimSpace(value)
	if code, ok := values[value]; ok {
		return code, nil
	}
	code := strings.ToUpper(value)
	for _, known := range values {
		if code == known {
			return code, nil
		}
	}
	return "", fmt.Errorf("unsupported QCC %s %q; use a top-level page label or code", name, value)
}
