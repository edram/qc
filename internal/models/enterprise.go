package models

// Enterprise is the provider-independent result of a company search.
type Enterprise struct {
	ID                      string                             `json:"id"`
	DetailURL               string                             `json:"detailUrl,omitempty"`
	Name                    string                             `json:"name"`
	RegistrationNumber      string                             `json:"registrationNumber,omitempty"`
	CreditCode              string                             `json:"creditCode,omitempty"`
	LegalRepresentative     string                             `json:"legalRepresentative,omitempty"`
	Status                  string                             `json:"status,omitempty"`
	EstablishedDate         string                             `json:"establishedDate,omitempty"`
	Address                 string                             `json:"address,omitempty"`
	AdministrativeDivisions *EnterpriseAdministrativeDivisions `json:"administrativeDivisions,omitempty"`
	RegisteredCapital       string                             `json:"registeredCapital,omitempty"`
	Phone                   string                             `json:"phone,omitempty"`
	Email                   string                             `json:"email,omitempty"`
	LogoURL                 string                             `json:"logoUrl,omitempty"`
	Tags                    []string                           `json:"tags,omitempty"`
	Risk                    *EnterpriseRiskSummary             `json:"risk,omitempty"`
}

// EnterpriseDetail contains the normalized fields available on an enterprise firm page.
type EnterpriseDetail struct {
	Enterprise
	BusinessScope         string                   `json:"businessScope,omitempty"`
	CompanyType           string                   `json:"companyType,omitempty"`
	RegistrationAuthority string                   `json:"registrationAuthority,omitempty"`
	OrganizationCode      string                   `json:"organizationCode,omitempty"`
	TaxNumber             string                   `json:"taxNumber,omitempty"`
	TermStart             string                   `json:"termStart,omitempty"`
	TermEnd               string                   `json:"termEnd,omitempty"`
	CheckDate             string                   `json:"checkDate,omitempty"`
	ActualCapital         string                   `json:"actualCapital,omitempty"`
	EnglishName           string                   `json:"englishName,omitempty"`
	Industry              *EnterpriseIndustry      `json:"industry,omitempty"`
	Websites              []EnterpriseWebsite      `json:"websites,omitempty"`
	Shareholders          []EnterpriseShareholder  `json:"shareholders,omitempty"`
	Employees             []EnterpriseEmployee     `json:"employees,omitempty"`
	Branches              []EnterpriseBranch       `json:"branches,omitempty"`
	Changes               []EnterpriseChange       `json:"changes,omitempty"`
	PreviousNames         []EnterprisePreviousName `json:"previousNames,omitempty"`
}

// EnterpriseIndustry contains the industry hierarchy returned by QCC.
type EnterpriseIndustry struct {
	Code       string `json:"code,omitempty"`
	Name       string `json:"name,omitempty"`
	SubCode    string `json:"subCode,omitempty"`
	SubName    string `json:"subName,omitempty"`
	MiddleCode string `json:"middleCode,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
	SmallCode  string `json:"smallCode,omitempty"`
	SmallName  string `json:"smallName,omitempty"`
}

// EnterpriseWebsite is a website published for an enterprise.
type EnterpriseWebsite struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// EnterpriseShareholder contains one shareholder and its contribution data.
type EnterpriseShareholder struct {
	ID                    string   `json:"id,omitempty"`
	Name                  string   `json:"name,omitempty"`
	Type                  string   `json:"type,omitempty"`
	Percent               string   `json:"percent,omitempty"`
	SubscribedCapital     string   `json:"subscribedCapital,omitempty"`
	SubscribedCapitalDate string   `json:"subscribedCapitalDate,omitempty"`
	PaidUpCapital         string   `json:"paidUpCapital,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
}

// EnterpriseEmployee contains one key employee.
type EnterpriseEmployee struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Position string `json:"position,omitempty"`
}

// EnterpriseBranch contains one branch office.
type EnterpriseBranch struct {
	ID                  string `json:"id,omitempty"`
	RegistrationNumber  string `json:"registrationNumber,omitempty"`
	Name                string `json:"name,omitempty"`
	Authority           string `json:"authority,omitempty"`
	CreditCode          string `json:"creditCode,omitempty"`
	LegalRepresentative string `json:"legalRepresentative,omitempty"`
}

// EnterpriseChange contains one registration change.
type EnterpriseChange struct {
	Project string `json:"project,omitempty"`
	Before  string `json:"before,omitempty"`
	After   string `json:"after,omitempty"`
	Date    string `json:"date,omitempty"`
}

// EnterprisePreviousName contains one historical enterprise name.
type EnterprisePreviousName struct {
	Name       string `json:"name,omitempty"`
	ChangeDate string `json:"changeDate,omitempty"`
}

// EnterpriseAdministrativeDivisions contains the administrative address hierarchy returned for an enterprise.
type EnterpriseAdministrativeDivisions struct {
	Province string `json:"province,omitempty"`
	City     string `json:"city,omitempty"`
	Area     string `json:"area,omitempty"`
}

// EnterpriseRiskSummary contains counts for risks directly belonging to an
// enterprise and risks belonging to its associated entities.
type EnterpriseRiskSummary struct {
	Direct     *EnterpriseRiskScope `json:"direct,omitempty"`
	Associated *EnterpriseRiskScope `json:"associated,omitempty"`
}

// EnterpriseRiskScope contains the count for one risk scope.
type EnterpriseRiskScope struct {
	Count int `json:"count"`
}

// EnterpriseSearchAggregation is one counted value in an enterprise search facet.
type EnterpriseSearchAggregation struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// EnterpriseSearchAggregations contains the supported enterprise search facets.
type EnterpriseSearchAggregations struct {
	Provinces  []EnterpriseSearchAggregation `json:"provinces,omitempty"`
	Industries []EnterpriseSearchAggregation `json:"industries,omitempty"`
}

// EnterpriseSearchResult contains one page of enterprises and counts for the full result set.
type EnterpriseSearchResult struct {
	Total        int                          `json:"total"`
	Enterprises  []Enterprise                 `json:"enterprises"`
	Aggregations EnterpriseSearchAggregations `json:"aggregations"`
}
