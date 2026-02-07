package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cf/lazytrack/internal/model"
)

func (c *Client) SearchUsers(query string) ([]model.User, error) {
	params := url.Values{}
	params.Set("fields", "id,login,fullName")
	params.Set("query", query)
	params.Set("$top", "10")

	resp, err := c.get("/api/users", params)
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}
	defer resp.Body.Close()

	var users []model.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decoding users: %w", err)
	}

	return users, nil
}
