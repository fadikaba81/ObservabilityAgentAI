package sumologic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fadikaba81/ObservabilityAgentAI/internal/telemetry"
)

var (
	pollInterval = 5 * time.Second
	pollTimeout  = 5 * time.Minute
)

func (c *Client) CreateSearchJob(query, from, to string) (*SearchJobResponse, error) {
	payload := SearchJobRequest{
		Query:    query,
		From:     from,
		To:       to,
		TimeZone: "UTC",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marchel search job request: %w", err)
	}

	req, err := http.NewRequest("POST", c.endpoint+"/searchJobs", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("Failed to create a request %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.accessID, c.accessKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var result SearchJobResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetSearchJobStatus(jobID string) (*SearchJobStatus, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/searchJobs/%s", c.endpoint, jobID), nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.accessID, c.accessKey)

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var status SearchJobStatus

	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil

}
func (c *Client) WaitForSearchJob(jobID string) (*SearchJobStatus, error) {
	start := time.Now()

	ticker := time.NewTicker(pollInterval)
	timeout := time.After(pollTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			slog.Error("search job timed out", "jobID", jobID, "timeout", pollTimeout)
			return nil, fmt.Errorf("search job %s timed out", jobID)

		case <-ticker.C:
			status, err := c.GetSearchJobStatus(jobID)
			if err != nil {
				slog.Error("failed to get search job status", "jobID", jobID, "error", err)
				return nil, err
			}

			switch status.State {
			case "DONE GATHERING RESULTS":
				slog.Info("search job completed", "jobID", jobID, "messageCount", status.MessageCount)
				telemetry.RecordSumoLogicDuration(context.Background(), time.Since(start))
				return status, nil
			case "CANCELLED", "FAILED":
				slog.Error("search job failed", "jobID", jobID, "state", status.State)
				return nil, fmt.Errorf("search job %s ended with state: %s", jobID, status.State)
			}
		}
	}

}

func (c *Client) GetSearchJobMessages(jobID string, totalCount int) ([]SearchJobMessage, error) {
	var (
		allMessages []SearchJobMessage
		offset      = 0
		limit       = 100
	)

	for offset < totalCount {

		url := fmt.Sprintf("%s/searchJobs/%s/messages?offset=%d&limit=%d",
			c.endpoint, jobID, offset, limit)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to create the request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(c.accessID, c.accessKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		
		var result SearchJobMessageResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode the response: %w", err)
		}
		resp.Body.Close()

		allMessages = append(allMessages, result.Messages...)

		offset += limit
	}

	return allMessages, nil
}
