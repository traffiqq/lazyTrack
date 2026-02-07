package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected accept header: %s", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1-1","login":"john","fullName":"John Doe","$type":"Me"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	user, err := client.GetCurrentUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Login != "john" {
		t.Errorf("got login %q, want %q", user.Login, "john")
	}
	if user.FullName != "John Doe" {
		t.Errorf("got fullName %q, want %q", user.FullName, "John Doe")
	}
}

func TestClient_GetCurrentUser_AuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized","error_description":"Invalid token"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	_, err := client.GetCurrentUser()
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestClient_GetCurrentUser_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1-1","login":"john","fullName":"John Doe"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	user, err := client.GetCurrentUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Login != "john" {
		t.Errorf("got login %q, want %q", user.Login, "john")
	}
}
