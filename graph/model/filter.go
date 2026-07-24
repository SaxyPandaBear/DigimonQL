package model

type Filter struct {
	Name        *string  `json:"name,omitempty"`
	Level       *string  `json:"level,omitempty"`
	DigimonType *string  `json:"type,omitempty"`
	Attribute   *string  `json:"attribute,omitempty"`
	Moves       []string `json:"moves,omitempty"`
	IsMode      *bool    `json:"isMode,omitempty"`
	IsXAntibody *bool    `json:"isXAntibody,omitempty"`
}
