package models

type Alias struct {
	ID      int32  `json:"id,omitempty"`
	Alias   string `json:"alias"`
	Source  string `json:"source"`
	Primary bool   `json:"is_primary"`
}
