package ui

import "testing"

func TestBuildFinderQuery(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		bug      bool
		task     bool
		expected string
	}{
		{
			name:     "text only, all types checked",
			text:     "server migration",
			bug:      true,
			task:     true,
			expected: "Type: Bug,Task server migration",
		},
		{
			name:     "text only, no types checked",
			text:     "server",
			bug:      false,
			task:     false,
			expected: "server",
		},
		{
			name:     "text with bug filter",
			text:     "crash",
			bug:      true,
			task:     false,
			expected: "Type: Bug crash",
		},
		{
			name:     "text with task filter",
			text:     "login",
			bug:      false,
			task:     true,
			expected: "Type: Task login",
		},
		{
			name:     "no text, bug filter",
			text:     "",
			bug:      true,
			task:     false,
			expected: "Type: Bug",
		},
		{
			name:     "no text, all types",
			text:     "",
			bug:      true,
			task:     true,
			expected: "Type: Bug,Task",
		},
		{
			name:     "no text, no types",
			text:     "",
			bug:      false,
			task:     false,
			expected: "",
		},
		{
			name:     "text with whitespace trimmed",
			text:     "  search  ",
			bug:      false,
			task:     false,
			expected: "search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFinderQuery(tt.text, tt.bug, tt.task)
			if got != tt.expected {
				t.Errorf("buildFinderQuery(%q, %v, %v) = %q, want %q",
					tt.text, tt.bug, tt.task, got, tt.expected)
			}
		})
	}
}
