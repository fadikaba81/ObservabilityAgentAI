package sumologic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) CreateSearchJob(query, from, to string) (*SearchJobResponse, error) {
	payload := SearchJobRequest{
		Query: query,
		From: from, 
		To: to,
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

	var result SearchJobResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &result, nil
}	