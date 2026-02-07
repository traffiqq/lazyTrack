package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("$top") != "50" {
			t.Errorf("unexpected $top: %s", r.URL.Query().Get("$top"))
		}
		if r.URL.Query().Get("$skip") != "0" {
			t.Errorf("unexpected $skip: %s", r.URL.Query().Get("$skip"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"2-1","idReadable":"PROJ-1","summary":"Test issue","project":{"id":"0-0","name":"Project","shortName":"PROJ"}}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	issues, err := client.ListIssues("", 0, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].IDReadable != "PROJ-1" {
		t.Errorf("got ID %q, want %q", issues[0].IDReadable, "PROJ-1")
	}
	if issues[0].Project.ShortName != "PROJ" {
		t.Errorf("got project %q, want %q", issues[0].Project.ShortName, "PROJ")
	}
}

func TestClient_ListIssues_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "project: PROJ #Unresolved" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("query"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	issues, err := client.ListIssues("project: PROJ #Unresolved", 0, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("got %d issues, want 0", len(issues))
	}
}

func TestClient_ListIssues_Pagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("$skip") != "50" {
			t.Errorf("unexpected $skip: %s", r.URL.Query().Get("$skip"))
		}
		if r.URL.Query().Get("$top") != "50" {
			t.Errorf("unexpected $top: %s", r.URL.Query().Get("$top"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"2-51","idReadable":"PROJ-51","summary":"Page 2 issue"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	issues, err := client.ListIssues("", 50, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].IDReadable != "PROJ-51" {
		t.Errorf("got ID %q, want %q", issues[0].IDReadable, "PROJ-51")
	}
}

func TestClient_GetIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/PROJ-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"2-1","idReadable":"PROJ-1","summary":"Test issue","description":"A test","project":{"id":"0-0","name":"Project","shortName":"PROJ"},"comments":[{"id":"4-1","text":"Hello","author":{"login":"john","fullName":"John"}}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	issue, err := client.GetIssue("PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Summary != "Test issue" {
		t.Errorf("got summary %q, want %q", issue.Summary, "Test issue")
	}
	if len(issue.Comments) != 1 {
		t.Fatalf("got %d comments, want 1", len(issue.Comments))
	}
}

func TestClient_CreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		if payload["summary"] != "New issue" {
			t.Errorf("unexpected summary: %v", payload["summary"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"2-2","idReadable":"PROJ-2","summary":"New issue"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	issue, err := client.CreateIssue("0-0", "New issue", "Description here", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.IDReadable != "PROJ-2" {
		t.Errorf("got ID %q, want %q", issue.IDReadable, "PROJ-2")
	}
}

func TestClient_CreateIssue_WithCustomFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		cf, ok := payload["customFields"].([]any)
		if !ok {
			t.Fatal("customFields missing or wrong type")
		}
		if len(cf) != 2 {
			t.Errorf("got %d custom fields, want 2", len(cf))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"2-3","idReadable":"PROJ-3","summary":"With fields"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	customFields := []map[string]any{
		{
			"name":  "State",
			"$type": "StateIssueCustomField",
			"value": map[string]string{"name": "Open", "$type": "StateBundleElement"},
		},
		{
			"name":  "Type",
			"$type": "SingleEnumIssueCustomField",
			"value": map[string]string{"name": "Bug", "$type": "EnumBundleElement"},
		},
	}
	issue, err := client.CreateIssue("0-0", "With fields", "Desc", customFields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.IDReadable != "PROJ-3" {
		t.Errorf("got ID %q, want %q", issue.IDReadable, "PROJ-3")
	}
}

func TestClient_UpdateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/issues/PROJ-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"2-1","idReadable":"PROJ-1","summary":"Updated"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.UpdateIssue("PROJ-1", map[string]any{"summary": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/issues/PROJ-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.DeleteIssue("PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"0-0","name":"My Project","shortName":"MP"},{"id":"0-1","name":"Other","shortName":"OT"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	projects, err := client.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}
	if projects[0].ShortName != "MP" {
		t.Errorf("got shortName %q, want %q", projects[0].ShortName, "MP")
	}
}
