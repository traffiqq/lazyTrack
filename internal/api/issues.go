package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/cf/lazytrack/internal/model"
)

const issueListFields = "id,idReadable,summary,description,created,updated,resolved,reporter(login,fullName),project(id,name,shortName),customFields(id,name,$type,value(id,name,login,fullName))"

const issueDetailFields = issueListFields + ",comments(id,text,author(login,fullName),created,updated)"

func (c *Client) ListIssues(query string, skip, top int) ([]model.Issue, error) {
	params := url.Values{}
	params.Set("fields", issueListFields)
	params.Set("$skip", strconv.Itoa(skip))
	params.Set("$top", strconv.Itoa(top))
	if query != "" {
		params.Set("query", query)
	}

	resp, err := c.get("/api/issues", params)
	if err != nil {
		return nil, fmt.Errorf("listing issues: %w", err)
	}
	defer resp.Body.Close()

	var issues []model.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("decoding issues: %w", err)
	}

	return issues, nil
}

func (c *Client) GetIssue(issueID string) (*model.Issue, error) {
	params := url.Values{}
	params.Set("fields", issueDetailFields)

	resp, err := c.get("/api/issues/"+url.PathEscape(issueID), params)
	if err != nil {
		return nil, fmt.Errorf("getting issue %s: %w", issueID, err)
	}
	defer resp.Body.Close()

	var issue model.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("decoding issue: %w", err)
	}

	return &issue, nil
}

func (c *Client) CreateIssue(projectID, summary, description string) (*model.Issue, error) {
	payload := map[string]any{
		"project":     map[string]string{"id": projectID},
		"summary":     summary,
		"description": description,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling issue: %w", err)
	}

	params := url.Values{}
	params.Set("fields", issueListFields)

	resp, err := c.post("/api/issues?"+params.Encode(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating issue: %w", err)
	}
	defer resp.Body.Close()

	var issue model.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("decoding created issue: %w", err)
	}

	return &issue, nil
}

func (c *Client) UpdateIssue(issueID string, fields map[string]any) error {
	body, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("marshaling update: %w", err)
	}

	params := url.Values{}
	params.Set("fields", "id,idReadable")

	resp, err := c.post("/api/issues/"+url.PathEscape(issueID)+"?"+params.Encode(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("updating issue %s: %w", issueID, err)
	}
	resp.Body.Close()

	return nil
}

func (c *Client) DeleteIssue(issueID string) error {
	if err := c.doDelete("/api/issues/" + url.PathEscape(issueID)); err != nil {
		return fmt.Errorf("deleting issue %s: %w", issueID, err)
	}
	return nil
}

func (c *Client) ListProjects() ([]model.Project, error) {
	params := url.Values{}
	params.Set("fields", "id,name,shortName")

	resp, err := c.get("/api/admin/projects", params)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer resp.Body.Close()

	var projects []model.Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("decoding projects: %w", err)
	}

	return projects, nil
}
