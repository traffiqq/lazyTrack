package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/PROJ-1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"4-1","text":"First comment","author":{"login":"john","fullName":"John"},"created":1700000000000}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	comments, err := client.ListComments("PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("got %d comments, want 1", len(comments))
	}
	if comments[0].Text != "First comment" {
		t.Errorf("got text %q, want %q", comments[0].Text, "First comment")
	}
}

func TestClient_AddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/issues/PROJ-1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)
		if payload["text"] != "New comment" {
			t.Errorf("unexpected text: %v", payload["text"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"4-2","text":"New comment","author":{"login":"john","fullName":"John"},"created":1700100000000}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	comment, err := client.AddComment("PROJ-1", "New comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.Text != "New comment" {
		t.Errorf("got text %q, want %q", comment.Text, "New comment")
	}
}
