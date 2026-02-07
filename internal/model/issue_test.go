package model

import (
	"encoding/json"
	"testing"
)

func TestIssue_StateValue(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "state present",
			json:     `{"customFields":[{"name":"State","$type":"StateIssueCustomField","value":{"name":"Open","$type":"StateBundleElement"}}]}`,
			expected: "Open",
		},
		{
			name:     "no state field",
			json:     `{"customFields":[{"name":"Priority","$type":"SingleEnumIssueCustomField","value":{"name":"Normal"}}]}`,
			expected: "",
		},
		{
			name:     "no custom fields",
			json:     `{}`,
			expected: "",
		},
		{
			name:     "null value",
			json:     `{"customFields":[{"name":"State","$type":"StateIssueCustomField","value":null}]}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var issue Issue
			if err := json.Unmarshal([]byte(tt.json), &issue); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := issue.StateValue(); got != tt.expected {
				t.Errorf("StateValue() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIssue_TypeValue(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "type present",
			json:     `{"customFields":[{"name":"Type","$type":"SingleEnumIssueCustomField","value":{"name":"Bug","$type":"EnumBundleElement"}}]}`,
			expected: "Bug",
		},
		{
			name:     "no type field",
			json:     `{"customFields":[{"name":"State","$type":"StateIssueCustomField","value":{"name":"Open"}}]}`,
			expected: "",
		},
		{
			name:     "no custom fields",
			json:     `{}`,
			expected: "",
		},
		{
			name:     "null value",
			json:     `{"customFields":[{"name":"Type","$type":"SingleEnumIssueCustomField","value":null}]}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var issue Issue
			if err := json.Unmarshal([]byte(tt.json), &issue); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := issue.TypeValue(); got != tt.expected {
				t.Errorf("TypeValue() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIssue_AssigneeValue(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantNil   bool
		wantLogin string
		wantName  string
	}{
		{
			name:      "assignee present",
			json:      `{"customFields":[{"name":"Assignee","$type":"SingleUserIssueCustomField","value":{"login":"john","fullName":"John Doe","$type":"User"}}]}`,
			wantLogin: "john",
			wantName:  "John Doe",
		},
		{
			name:    "no assignee",
			json:    `{"customFields":[]}`,
			wantNil: true,
		},
		{
			name:    "null value",
			json:    `{"customFields":[{"name":"Assignee","$type":"SingleUserIssueCustomField","value":null}]}`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var issue Issue
			if err := json.Unmarshal([]byte(tt.json), &issue); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			user := issue.AssigneeValue()
			if tt.wantNil {
				if user != nil {
					t.Errorf("AssigneeValue() = %v, want nil", user)
				}
				return
			}
			if user == nil {
				t.Fatal("AssigneeValue() = nil, want non-nil")
			}
			if user.Login != tt.wantLogin {
				t.Errorf("login = %q, want %q", user.Login, tt.wantLogin)
			}
			if user.FullName != tt.wantName {
				t.Errorf("fullName = %q, want %q", user.FullName, tt.wantName)
			}
		})
	}
}
