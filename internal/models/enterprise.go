package models

// Enterprise is the provider-independent result of a company search.
type Enterprise struct {
	ID                  string   `json:"id"`
	DetailURL           string   `json:"detailUrl,omitempty"`
	Name                string   `json:"name"`
	RegistrationNumber  string   `json:"registrationNumber,omitempty"`
	CreditCode          string   `json:"creditCode,omitempty"`
	LegalRepresentative string   `json:"legalRepresentative,omitempty"`
	Status              string   `json:"status,omitempty"`
	EstablishedDate     string   `json:"establishedDate,omitempty"`
	Address             string   `json:"address,omitempty"`
	RegisteredCapital   string   `json:"registeredCapital,omitempty"`
	Phone               string   `json:"phone,omitempty"`
	Email               string   `json:"email,omitempty"`
	LogoURL             string   `json:"logoUrl,omitempty"`
	Tags                []string `json:"tags,omitempty"`
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
