package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SearchUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "john" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("$top") != "10" {
			t.Errorf("unexpected $top: %s", r.URL.Query().Get("$top"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"1-1","login":"john.doe","fullName":"John Doe"},{"id":"1-2","login":"johnny","fullName":"Johnny Smith"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	users, err := client.SearchUsers("john")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d users, want 2", len(users))
	}
	if users[0].Login != "john.doe" {
		t.Errorf("got login %q, want %q", users[0].Login, "john.doe")
	}
	if users[0].ID != "1-1" {
		t.Errorf("got ID %q, want %q", users[0].ID, "1-1")
	}
}

func TestClient_SearchUsers_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	users, err := client.SearchUsers("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("got %d users, want 0", len(users))
	}
}
