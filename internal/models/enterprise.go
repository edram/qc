package models

// Enterprise is the provider-independent result of a company search.
type Enterprise struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	RegistrationNumber  string `json:"registrationNumber,omitempty"`
	CreditCode          string `json:"creditCode,omitempty"`
	LegalRepresentative string `json:"legalRepresentative,omitempty"`
	Status              string `json:"status,omitempty"`
	EstablishedDate     string `json:"establishedDate,omitempty"`
	Address             string `json:"address,omitempty"`
	RegisteredCapital   string `json:"registeredCapital,omitempty"`
	Phone               string `json:"phone,omitempty"`
	Email               string `json:"email,omitempty"`
	LogoURL             string `json:"logoUrl,omitempty"`
}
