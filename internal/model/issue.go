package model

import "encoding/json"

type Issue struct {
	ID           string        `json:"id"`
	IDReadable   string        `json:"idReadable"`
	Summary      string        `json:"summary"`
	Description  string        `json:"description"`
	Created      int64         `json:"created"`
	Updated      int64         `json:"updated"`
	Resolved     *int64        `json:"resolved"`
	Reporter     *User         `json:"reporter"`
	Project      *Project      `json:"project"`
	Comments     []Comment     `json:"comments"`
	CustomFields []CustomField `json:"customFields"`
}

type CustomField struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Type  string          `json:"$type"`
	Value json.RawMessage `json:"value"`
}

// customFieldValueName extracts "name" from a custom field value JSON object.
// Returns "" if the value is null, not an object, or has no "name" key.
func customFieldValueName(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	if name, ok := obj["name"].(string); ok {
		return name
	}
	return ""
}

// customFieldValueUser extracts a User from a custom field value JSON object.
// Returns nil if the value is null or not a valid user object.
func customFieldValueUser(raw json.RawMessage) *User {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil
	}
	u := &User{}
	if login, ok := obj["login"].(string); ok {
		u.Login = login
	}
	if name, ok := obj["fullName"].(string); ok {
		u.FullName = name
	}
	if u.Login == "" && u.FullName == "" {
		return nil
	}
	return u
}

// StateValue extracts the "State" custom field value name, returns "" if not found.
func (i *Issue) StateValue() string {
	for _, cf := range i.CustomFields {
		if cf.Name == "State" {
			return customFieldValueName(cf.Value)
		}
	}
	return ""
}

// AssigneeValue extracts the "Assignee" custom field, returns nil if not found.
func (i *Issue) AssigneeValue() *User {
	for _, cf := range i.CustomFields {
		if cf.Name == "Assignee" {
			return customFieldValueUser(cf.Value)
		}
	}
	return nil
}
