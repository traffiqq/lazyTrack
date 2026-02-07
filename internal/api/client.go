package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cf/lazytrack/internal/model"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return resp, nil
}

func (c *Client) get(path string, params url.Values) (*http.Response, error) {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return c.doRequest(http.MethodGet, path, nil)
}

func (c *Client) post(path string, body io.Reader) (*http.Response, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// doDelete performs a DELETE and closes the response body (DELETE returns no useful body).
func (c *Client) doDelete(path string) error {
	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) GetCurrentUser() (*model.User, error) {
	params := url.Values{}
	params.Set("fields", "id,login,fullName")

	resp, err := c.get("/api/users/me", params)
	if err != nil {
		return nil, fmt.Errorf("fetching current user: %w", err)
	}
	defer resp.Body.Close()

	var user model.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding user: %w", err)
	}

	return &user, nil
}
