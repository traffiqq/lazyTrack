package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cf/lazytrack/internal/model"
)

func (c *Client) ListProjectCustomFields(projectID string) ([]model.ProjectCustomField, error) {
	params := url.Values{}
	params.Set("fields", "field(id,name,$type),bundle(values(id,name,$type))")

	resp, err := c.get("/api/admin/projects/"+url.PathEscape(projectID)+"/customFields", params)
	if err != nil {
		return nil, fmt.Errorf("listing custom fields for project %s: %w", projectID, err)
	}
	defer resp.Body.Close()

	var fields []model.ProjectCustomField
	if err := json.NewDecoder(resp.Body).Decode(&fields); err != nil {
		return nil, fmt.Errorf("decoding custom fields: %w", err)
	}

	return fields, nil
}
