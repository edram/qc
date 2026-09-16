package models

// Person is the provider-independent result of a people search.
type Person struct {
	ID                  string `json:"id"`
	DetailURL           string `json:"detailUrl,omitempty"`
	Name                string `json:"name"`
	MainCompanyID       string `json:"mainCompanyId,omitempty"`
	MainCompanyName     string `json:"mainCompanyName,omitempty"`
	Role                string `json:"role,omitempty"`
	RelatedCompanyCount int    `json:"relatedCompanyCount,omitempty"`
	PartnerCount        int    `json:"partnerCount,omitempty"`
	Introduction        string `json:"introduction,omitempty"`
	AvatarURL           string `json:"avatarUrl,omitempty"`
}
