package domain

// GroupModelsListConfig controls the optional custom /v1/models response list.
type GroupModelsListConfig struct {
	Enabled     bool     `json:"enabled"`
	Models      []string `json:"models,omitempty"`
	DisplayIcon string   `json:"display_icon,omitempty"`
}
