package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/edram/qi/internal/models"
)

const qccFirmPagePrefix = "/firm/"

func (s *searchQCC) GetEnterpriseDetail(ctx context.Context, id string) (models.EnterpriseDetail, error) {
	response, err := s.api.Get(ctx, qccFirmPagePrefix+url.PathEscape(id)+".html")
	if err != nil {
		return models.EnterpriseDetail{}, err
	}
	defer response.Body.Close()
	page, err := io.ReadAll(response.Body)
	if err != nil {
		return models.EnterpriseDetail{}, fmt.Errorf("read QCC firm page: %w", err)
	}
	state, err := qccEnterpriseInitialState(page)
	if err != nil {
		return models.EnterpriseDetail{}, err
	}
	detail, err := qccEnterpriseCompanyDetail(state)
	if err != nil {
		return models.EnterpriseDetail{}, err
	}
	return qccEnterpriseFromDetail(detail, id)
}

func qccEnterpriseCompanyDetail(state []byte) ([]byte, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(state, &root); err != nil {
		return nil, fmt.Errorf("qcc enterprise page: decode initial state: %w", err)
	}
	company, ok := qccInitialStateObject(root, "company")
	if !ok {
		return nil, errors.New("qcc enterprise page: company state not found")
	}
	detail, ok := qccInitialStateObject(company, "companyDetail")
	if !ok {
		return nil, errors.New("qcc enterprise page: company detail state not found")
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		return nil, fmt.Errorf("qcc enterprise page: encode company detail: %w", err)
	}
	return encoded, nil
}

func qccEnterpriseFromDetail(raw []byte, requestedID string) (models.EnterpriseDetail, error) {
	values, err := qccInitialStateMap(raw)
	if err != nil {
		return models.EnterpriseDetail{}, fmt.Errorf("qcc enterprise page: decode company detail: %w", err)
	}
	oper := qccInitialStateNestedObject(values, "Oper", "oper")
	area := qccInitialStateNestedObject(values, "Area", "area")
	contact := qccInitialStateNestedObject(values, "ContactInfo", "contactInfo")
	id := qccInitialStateString(values, "KeyNo", "keyNo")
	if id == "" {
		id = requestedID
	}
	name := qccInitialStateString(values, "Name", "name")
	if name == "" {
		return models.EnterpriseDetail{}, errors.New("qcc enterprise page: company name not found")
	}

	enterprise := models.Enterprise{
		ID:                  id,
		DetailURL:           "https://www.qcc.com/firm/" + id + ".html",
		Name:                name,
		RegistrationNumber:  qccInitialStateString(values, "No", "no"),
		CreditCode:          qccInitialStateString(values, "CreditCode", "creditCode"),
		LegalRepresentative: qccInitialStateString(values, "OperName", "operName"),
		Status:              qccInitialStateString(values, "Status", "status"),
		EstablishedDate:     qccInitialStateString(values, "StartDate", "startDate"),
		Address:             qccInitialStateString(values, "Address", "address"),
		RegisteredCapital:   qccInitialStateString(values, "RegistCapi", "registCapi"),
		LogoURL:             qccInitialStateString(values, "ImageUrl", "ImageURL", "imageUrl"),
		Phone:               qccInitialStateString(contact, "PhoneNumber", "phoneNumber"),
		Email:               qccInitialStateString(contact, "Email", "email"),
	}
	if enterprise.LegalRepresentative == "" {
		enterprise.LegalRepresentative = qccInitialStateString(oper, "Name", "name")
	}
	if divisions := qccEnterpriseAdministrativeDivisions(
		qccInitialStateString(area, "Province", "province"),
		qccInitialStateString(area, "City", "city"),
		qccInitialStateString(area, "County", "county"),
	); divisions != nil {
		enterprise.AdministrativeDivisions = divisions
	}
	enterprise.Tags = qccInitialStateTags(values)
	return qccEnterpriseDetailFields(values, enterprise), nil
}

func qccEnterpriseDetailFields(values map[string]json.RawMessage, base models.Enterprise) models.EnterpriseDetail {
	contact := qccInitialStateNestedObject(values, "ContactInfo", "contactInfo")
	detail := models.EnterpriseDetail{
		Enterprise:            base,
		BusinessScope:         qccInitialStateString(values, "Scope", "scope", "BusinessScope", "businessScope"),
		CompanyType:           qccInitialStateString(values, "EconKind", "econKind", "CompanyType", "companyType"),
		RegistrationAuthority: qccInitialStateString(values, "BelongOrg", "belongOrg", "Authority", "authority"),
		OrganizationCode:      qccInitialStateString(values, "OrgNo", "orgNo", "OrgCode", "orgCode"),
		TaxNumber:             qccInitialStateString(values, "TaxNo", "taxNo", "TaxNumber", "taxNumber"),
		TermStart:             qccInitialStateString(values, "TermStart", "termStart"),
		TermEnd:               qccInitialStateString(values, "TermEnd", "termEnd"),
		CheckDate:             qccInitialStateString(values, "CheckDate", "checkDate", "IssueDate", "issueDate"),
		ActualCapital:         qccInitialStateString(values, "RecCap", "recCap", "RealCapi", "realCapi", "PaidUpCapital", "paidUpCapital"),
		EnglishName:           qccInitialStateString(values, "EnglishName", "englishName"),
		Industry:              qccEnterpriseIndustry(values),
		Websites:              qccEnterpriseWebsites(contact, values),
		Shareholders:          qccEnterpriseShareholders(values),
		Employees:             qccEnterpriseEmployees(values),
		Branches:              qccEnterpriseBranches(values),
		Changes:               qccEnterpriseChanges(values),
		PreviousNames:         qccEnterprisePreviousNames(values),
	}
	return detail
}

func qccEnterpriseIndustry(values map[string]json.RawMessage) *models.EnterpriseIndustry {
	industry := qccInitialStateNestedObject(values, "Industry", "industry", "QccIndustry", "qccIndustry")
	standard := qccInitialStateNestedObject(values, "IndustryV3", "industryV3")
	if industry == nil && standard == nil {
		return nil
	}
	result := &models.EnterpriseIndustry{
		Code:       qccInitialStateString(industry, "IndustryCode", "Code", "code", "Cn", "cn"),
		Name:       qccInitialStateString(industry, "Industry", "Name", "name", "Dn", "dn"),
		SubCode:    qccInitialStateString(industry, "SubIndustryCode", "SubCode", "subCode"),
		SubName:    qccInitialStateString(industry, "SubIndustry", "SubName", "subName", "Cn", "cn"),
		MiddleCode: qccInitialStateString(industry, "MiddleCategoryCode", "MiddleCode", "middleCode"),
		MiddleName: qccInitialStateString(industry, "MiddleCategory", "MiddleName", "middleName"),
		SmallCode:  qccInitialStateString(industry, "SmallCategoryCode", "SmallCode", "smallCode"),
		SmallName:  qccInitialStateString(industry, "SmallCategory", "SmallName", "smallName"),
	}
	if result.MiddleName == "" {
		result.MiddleName = qccInitialStateString(standard, "MiddleCategory", "MiddleName", "middleName")
	}
	if result.SmallName == "" {
		result.SmallName = qccInitialStateString(standard, "SmallCategory", "SmallName", "smallName")
	}
	if result.Name == "" {
		result.Name = qccInitialStateString(standard, "Industry", "Name", "name")
	}
	if *result == (models.EnterpriseIndustry{}) {
		return nil
	}
	return result
}

func qccEnterpriseWebsites(contact, values map[string]json.RawMessage) []models.EnterpriseWebsite {
	items := qccInitialStateArray(contact, "WebSite", "Website", "websites")
	if len(items) == 0 {
		items = qccInitialStateArray(values, "WebSite", "Website", "websites")
	}
	if len(items) == 0 {
		raw := qccInitialStateValue(contact, "WebSite", "Website", "websites")
		if raw == nil {
			raw = qccInitialStateValue(values, "WebSite", "Website", "websites")
		}
		var website string
		if json.Unmarshal(raw, &website) == nil && strings.TrimSpace(website) != "" {
			return []models.EnterpriseWebsite{{URL: qccText(strings.TrimSpace(website))}}
		}
	}
	result := make([]models.EnterpriseWebsite, 0, len(items))
	for _, item := range items {
		website := models.EnterpriseWebsite{
			Name: qccInitialStateString(item, "Name", "name"),
			URL:  qccInitialStateString(item, "Url", "URL", "url"),
		}
		if website.Name != "" || website.URL != "" {
			result = append(result, website)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func qccEnterpriseShareholders(values map[string]json.RawMessage) []models.EnterpriseShareholder {
	items := qccInitialStateArray(values, "Partners", "partners", "ShareHolderList", "shareHolderList", "Shareholders", "shareholders")
	result := make([]models.EnterpriseShareholder, 0, len(items))
	for _, item := range items {
		shareholder := models.EnterpriseShareholder{
			ID:                    qccInitialStateString(item, "KeyNo", "keyNo", "Id", "id"),
			Name:                  qccInitialStateString(item, "StockName", "stockName", "Name", "name"),
			Type:                  qccInitialStateString(item, "StockType", "stockType", "Type", "type"),
			Percent:               qccInitialStateString(item, "StockPercent", "stockPercent", "Ratio", "ratio", "Percent", "percent"),
			SubscribedCapital:     qccInitialStateString(item, "SubscribedCapital", "subscribedCapital", "ShouldCapi", "shouldCapi", "SubConAmount", "subConAmount"),
			SubscribedCapitalDate: qccInitialStateString(item, "ShoudDate", "ShouldDate", "subscribedCapitalDate"),
			PaidUpCapital:         qccInitialStateString(item, "PaidUpCapital", "paidUpCapital", "RealCapi", "realCapi"),
			Tags:                  qccInitialStateTags(item),
		}
		if shareholder.ID != "" || shareholder.Name != "" {
			result = append(result, shareholder)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func qccEnterpriseEmployees(values map[string]json.RawMessage) []models.EnterpriseEmployee {
	items := qccInitialStateArray(values, "Employees", "employees", "EmployeeList", "employeeList", "KeyPersonList", "keyPersonList")
	result := make([]models.EnterpriseEmployee, 0, len(items))
	for _, item := range items {
		employee := models.EnterpriseEmployee{
			ID:       qccInitialStateString(item, "KeyNo", "keyNo", "Id", "id"),
			Name:     qccInitialStateString(item, "Name", "name"),
			Position: qccInitialStateString(item, "Job", "job", "Position", "position"),
		}
		if employee.ID != "" || employee.Name != "" {
			result = append(result, employee)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func qccEnterpriseBranches(values map[string]json.RawMessage) []models.EnterpriseBranch {
	items := qccInitialStateArray(values, "Branches", "branches", "BranchList", "branchList")
	result := make([]models.EnterpriseBranch, 0, len(items))
	for _, item := range items {
		branch := models.EnterpriseBranch{
			ID:                  qccInitialStateString(item, "CompanyId", "companyId", "KeyNo", "keyNo", "Id", "id"),
			RegistrationNumber:  qccInitialStateString(item, "RegNo", "regNo", "CompanyCode", "companyCode", "registrationNumber"),
			Name:                qccInitialStateString(item, "Name", "name", "CompanyName", "companyName"),
			Authority:           qccInitialStateString(item, "BelongOrg", "belongOrg", "Authority", "authority"),
			CreditCode:          qccInitialStateString(item, "CreditCode", "creditCode"),
			LegalRepresentative: qccInitialStateString(item, "OperName", "operName", "LegalPerson", "legalPerson"),
		}
		if branch.ID != "" || branch.Name != "" {
			result = append(result, branch)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func qccEnterpriseChanges(values map[string]json.RawMessage) []models.EnterpriseChange {
	items := qccInitialStateArray(values, "ChangeRecords", "changeRecords", "Changes", "changes", "ChangeList", "changeList")
	result := make([]models.EnterpriseChange, 0, len(items))
	for _, item := range items {
		change := models.EnterpriseChange{
			Project: qccInitialStateString(item, "ProjectName", "projectName", "ChangeField", "changeField", "Project", "project"),
			Before:  qccInitialStateString(item, "BeforeContent", "beforeContent", "ChangeBefore", "changeBefore", "Before", "before"),
			After:   qccInitialStateString(item, "AfterContent", "afterContent", "ChangeAfter", "changeAfter", "After", "after"),
			Date:    qccInitialStateString(item, "ChangeDate", "changeDate", "Date", "date"),
		}
		if change.Project != "" || change.Before != "" || change.After != "" || change.Date != "" {
			result = append(result, change)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func qccEnterprisePreviousNames(values map[string]json.RawMessage) []models.EnterprisePreviousName {
	items := qccInitialStateArray(values, "OriginalName", "originalName", "PreviousNames", "previousNames")
	result := make([]models.EnterprisePreviousName, 0, len(items))
	for _, item := range items {
		name := models.EnterprisePreviousName{
			Name:       qccInitialStateString(item, "Name", "name"),
			ChangeDate: qccInitialStateString(item, "ChangeDate", "changeDate"),
		}
		if name.Name != "" {
			result = append(result, name)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func qccInitialStateMap(raw []byte) (map[string]json.RawMessage, error) {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func qccInitialStateNestedObject(values map[string]json.RawMessage, keys ...string) map[string]json.RawMessage {
	raw := qccInitialStateValue(values, keys...)
	if raw == nil {
		return nil
	}
	result, err := qccInitialStateMap(raw)
	if err != nil {
		return nil
	}
	return result
}

func qccInitialStateValue(values map[string]json.RawMessage, keys ...string) json.RawMessage {
	for _, key := range keys {
		for actual, value := range values {
			if strings.EqualFold(actual, key) {
				return value
			}
		}
	}
	return nil
}

func qccInitialStateArray(values map[string]json.RawMessage, keys ...string) []map[string]json.RawMessage {
	raw := qccInitialStateValue(values, keys...)
	if raw == nil {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	result := make([]map[string]json.RawMessage, 0, len(items))
	for _, item := range items {
		object, err := qccInitialStateMap(item)
		if err == nil {
			result = append(result, object)
		}
	}
	return result
}

func qccInitialStateString(values map[string]json.RawMessage, keys ...string) string {
	raw := qccInitialStateValue(values, keys...)
	if raw == nil {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return qccText(strings.TrimSpace(value))
	}
	return ""
}

func qccInitialStateTags(values map[string]json.RawMessage) []string {
	raw := qccInitialStateValue(values, "TagList", "tagList", "Tags", "tags")
	if raw == nil {
		return nil
	}
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil {
		return nil
	}
	tags := make([]string, 0, len(items))
	for _, item := range items {
		var text string
		if json.Unmarshal(item, &text) == nil {
			if text = qccText(strings.TrimSpace(text)); text != "" {
				tags = append(tags, text)
			}
			continue
		}
		object, err := qccInitialStateMap(item)
		if err != nil {
			continue
		}
		if text = qccInitialStateString(object, "Name", "name"); text != "" {
			tags = append(tags, text)
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func qccEnterpriseInitialState(page []byte) (json.RawMessage, error) {
	state, err := extractQCCInitialState(page)
	if err != nil {
		return nil, err
	}
	if !json.Valid(state) {
		return nil, errors.New("qcc enterprise page: initial state is not valid JSON")
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(state, &root); err != nil {
		return nil, fmt.Errorf("qcc enterprise page: decode initial state: %w", err)
	}
	company, ok := qccInitialStateObject(root, "company")
	if !ok {
		return nil, errors.New("qcc enterprise page: company state not found")
	}
	if _, ok := qccInitialStateObject(company, "companyDetail"); !ok {
		return nil, errors.New("qcc enterprise page: company detail state not found")
	}
	return json.RawMessage(append([]byte(nil), state...)), nil
}

func qccInitialStateObject(values map[string]json.RawMessage, key string) (map[string]json.RawMessage, bool) {
	raw, ok := values[key]
	if !ok {
		return nil, false
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, false
	}
	return object, true
}

func extractQCCInitialState(page []byte) ([]byte, error) {
	const marker = "window.__INITIAL_STATE__"
	start := 0
	for {
		index := bytes.Index(page[start:], []byte(marker))
		if index < 0 {
			return nil, errors.New("qcc enterprise page: window.__INITIAL_STATE__ not found")
		}
		index += start + len(marker)
		end := bytes.Index(page[index:], []byte("</script>"))
		if end < 0 {
			end = len(page) - index
		}
		end += index
		if state, ok := scanInitialStateJSON(page[index:end]); ok {
			return state, nil
		}
		start = index
		if start >= len(page) {
			return nil, errors.New("qcc enterprise page: initial state JSON is invalid")
		}
	}
}

func scanInitialStateJSON(script []byte) ([]byte, bool) {
	start := bytes.IndexByte(script, '{')
	if start < 0 {
		return nil, false
	}

	depth := 0
	inString := false
	escaped := false
	for index := start; index < len(script); index++ {
		character := script[index]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
				continue
			}
			if character == '"' {
				inString = false
			}
			continue
		}
		if character == '"' {
			inString = true
			continue
		}
		switch character {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return script[start : index+1], true
			}
		}
	}
	return nil, false
}

var _ enterpriseDetailSearch = (*searchQCC)(nil)
