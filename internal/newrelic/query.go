package newrelic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) QueryErrorRate(service string, queryWindow string) (float64, error) {
	nrql := fmt.Sprintf(
		"SELECT percentage(count(*), WHERE error is true) FROM Transaction WHERE appName='%s' SINCE %s ago",
		service,
		queryWindow,
	)

	query := NerdGraphRequest{
		Query: fmt.Sprintf(
			`{ actor { account(id: "%s") { nrql(query: "%s") {results }}} }`,
			c.accountID,
			nrql,
		),
	}

	body, err := json.Marshal(query)
	if err != nil {
		return 0, fmt.Errorf("failed to Marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var result NerdGraphResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("Failed to decode response: %w", err)
	}

	results := result.Data.Actor.Account.NRQL.Results
	if len(results) == 0 {
		return 0, nil
	}

	errorRate, ok := results[0]["percentage"].(float64)
	if !ok {
		return 0, fmt.Errorf("failed to extract error rate from results")
	}

	return errorRate, nil

}
