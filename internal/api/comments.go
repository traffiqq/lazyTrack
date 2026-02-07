package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cf/lazytrack/internal/model"
)

const commentFields = "id,text,author(login,fullName),created,updated"

func (c *Client) ListComments(issueID string) ([]model.Comment, error) {
	params := url.Values{}
	params.Set("fields", commentFields)

	resp, err := c.get("/api/issues/"+url.PathEscape(issueID)+"/comments", params)
	if err != nil {
		return nil, fmt.Errorf("listing comments for %s: %w", issueID, err)
	}
	defer resp.Body.Close()

	var comments []model.Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("decoding comments: %w", err)
	}

	return comments, nil
}

func (c *Client) AddComment(issueID, text string) (*model.Comment, error) {
	payload := map[string]string{"text": text}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling comment: %w", err)
	}

	params := url.Values{}
	params.Set("fields", commentFields)

	resp, err := c.post("/api/issues/"+url.PathEscape(issueID)+"/comments?"+params.Encode(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("adding comment to %s: %w", issueID, err)
	}
	defer resp.Body.Close()

	var comment model.Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, fmt.Errorf("decoding comment: %w", err)
	}

	return &comment, nil
}
