package sumologic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/pkg/models"
)

var (
	pollInterval = 5 * time.Second
	pollTimeout  = 5 * time.Minute
)

const messagePageSize = 100

// Search runs a search job over [from, to], waits for it to finish, and returns
// up to limit normalised events, oldest first.
func (c *Client) Search(query string, from, to time.Time, limit int) ([]models.Event, error) {
	job, err := c.CreateSearchJob(query, from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	status, err := c.WaitForSearchJob(job.ID)
	if err != nil {
		return nil, err
	}

	count := min(status.MessageCount, limit)
	if count == 0 {
		return nil, nil
	}

	messages, err := c.GetSearchJobMessages(job.ID, count)
	if err != nil {
		return nil, err
	}
	return NormalizeMessages(messages)
}

func (c *Client) CreateSearchJob(query, from, to string) (*SearchJobResponse, error) {
	body, err := json.Marshal(SearchJobRequest{
		Query:    query,
		From:     from,
		To:       to,
		TimeZone: "UTC",
	})
	if err != nil {
		return nil, fmt.Errorf("marshal search job request: %w", err)
	}

	var result SearchJobResponse
	if err := c.doJSON(http.MethodPost, c.endpoint+"/searchJobs", body, http.StatusAccepted, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetSearchJobStatus(jobID string) (*SearchJobStatus, error) {
	var status SearchJobStatus
	if err := c.doJSON(http.MethodGet, fmt.Sprintf("%s/searchJobs/%s", c.endpoint, jobID), nil, http.StatusOK, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// WaitForSearchJob polls until the job has finished gathering results.
func (c *Client) WaitForSearchJob(jobID string) (*SearchJobStatus, error) {
	ticker := time.NewTicker(pollInterval)
	timeout := time.After(pollTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("search job %s timed out after %s", jobID, pollTimeout)

		case <-ticker.C:
			status, err := c.GetSearchJobStatus(jobID)
			if err != nil {
				return nil, err
			}

			switch status.State {
			case "DONE GATHERING RESULTS":
				slog.Info("sumo.search.done", "job_id", jobID, "messages", status.MessageCount)
				return status, nil
			case "CANCELLED", "FAILED":
				return nil, fmt.Errorf("search job %s ended with state %s", jobID, status.State)
			}
		}
	}
}

// GetSearchJobMessages fetches up to totalCount messages from a finished job.
func (c *Client) GetSearchJobMessages(jobID string, totalCount int) ([]SearchJobMessage, error) {
	var all []SearchJobMessage

	for offset := 0; offset < totalCount; offset += messagePageSize {
		url := fmt.Sprintf("%s/searchJobs/%s/messages?offset=%d&limit=%d",
			c.endpoint, jobID, offset, messagePageSize)

		var page SearchJobMessageResponse
		if err := c.doJSON(http.MethodGet, url, nil, http.StatusOK, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Messages...)
		if len(page.Messages) == 0 {
			break
		}
	}

	if len(all) > totalCount {
		all = all[:totalCount]
	}
	return all, nil
}

// doJSON performs an authenticated request, checks the status code, and decodes
// the JSON response into out (if out is non-nil).
func (c *Client) doJSON(method, url string, body []byte, expectStatus int, out any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.SetBasicAuth(c.accessID, c.accessKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectStatus {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("sumo logic returned %d: %s", resp.StatusCode, snippet)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
