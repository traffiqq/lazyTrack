package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_ListProjectCustomFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/admin/projects/" + url.PathEscape("0-0") + "/customFields"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: %s, want %s", r.URL.Path, expectedPath)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"field": {"id":"f-1","name":"State","$type":"StateIssueCustomField"},
				"bundle": {"values":[
					{"id":"v-1","name":"Open","$type":"StateBundleElement"},
					{"id":"v-2","name":"In Progress","$type":"StateBundleElement"},
					{"id":"v-3","name":"Fixed","$type":"StateBundleElement"}
				]}
			},
			{
				"field": {"id":"f-2","name":"Type","$type":"SingleEnumIssueCustomField"},
				"bundle": {"values":[
					{"id":"v-4","name":"Bug","$type":"EnumBundleElement"},
					{"id":"v-5","name":"Task","$type":"EnumBundleElement"},
					{"id":"v-6","name":"Feature","$type":"EnumBundleElement"}
				]}
			}
		]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	fields, err := client.ListProjectCustomFields("0-0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("got %d fields, want 2", len(fields))
	}

	if fields[0].Field.Name != "State" {
		t.Errorf("got field name %q, want %q", fields[0].Field.Name, "State")
	}
	if fields[0].Field.Type != "StateIssueCustomField" {
		t.Errorf("got field type %q, want %q", fields[0].Field.Type, "StateIssueCustomField")
	}
	if len(fields[0].Bundle.Values) != 3 {
		t.Fatalf("got %d state values, want 3", len(fields[0].Bundle.Values))
	}
	if fields[0].Bundle.Values[0].Name != "Open" {
		t.Errorf("got state value %q, want %q", fields[0].Bundle.Values[0].Name, "Open")
	}
	if fields[0].Bundle.Values[0].Type != "StateBundleElement" {
		t.Errorf("got bundle type %q, want %q", fields[0].Bundle.Values[0].Type, "StateBundleElement")
	}

	if fields[1].Field.Name != "Type" {
		t.Errorf("got field name %q, want %q", fields[1].Field.Name, "Type")
	}
	if len(fields[1].Bundle.Values) != 3 {
		t.Fatalf("got %d type values, want 3", len(fields[1].Bundle.Values))
	}
}

func TestClient_ListProjectCustomFields_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	fields, err := client.ListProjectCustomFields("0-0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 0 {
		t.Fatalf("got %d fields, want 0", len(fields))
	}
}
