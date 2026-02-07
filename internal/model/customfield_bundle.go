package model

// BundleValue represents a single value in a YouTrack custom field bundle
// (e.g., a state like "Open" or a type like "Bug").
type BundleValue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"$type"`
}

// ProjectCustomField represents a custom field configuration for a project,
// including its bundle of allowed values.
type ProjectCustomField struct {
	Field struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"$type"`
	} `json:"field"`
	Bundle struct {
		Values []BundleValue `json:"values"`
	} `json:"bundle"`
}
