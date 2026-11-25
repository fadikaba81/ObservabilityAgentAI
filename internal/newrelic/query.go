package newrelic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const nerdGraphQuery = `query($accountId: Int!, $nrql: Nrql!) {
  actor { account(id: $accountId) { nrql(query: $nrql) { results } } }
}`

// Window is the time range an alert covers.
type Window struct {
	Start time.Time
	End   time.Time
}

// RunNRQL executes an NRQL query through NerdGraph and returns the result rows.
// The account ID and NRQL are sent as GraphQL variables, not spliced into the query.
func (c *Client) RunNRQL(nrql string) ([]map[string]any, error) {
	accountID, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("NEWRELIC_ACCOUNTID_SANDBOX must be numeric: %w", err)
	}

	body, err := json.Marshal(NerdGraphRequest{
		Query:     nerdGraphQuery,
		Variables: map[string]any{"accountId": accountID, "nrql": nrql},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("new relic returned %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}

	var result NerdGraphResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(result.Errors) > 0 {
		msgs := make([]string, 0, len(result.Errors))
		for _, e := range result.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, errors.New("new relic query failed: " + strings.Join(msgs, "; "))
	}

	return result.Data.Actor.Account.NRQL.Results, nil
}

// ServiceMetrics returns request count, error percentage, and duration stats
// for an application within the window.
func (c *Client) ServiceMetrics(app string, w Window) ([]map[string]any, error) {
	return c.RunNRQL(fmt.Sprintf(
		"SELECT count(*), percentage(count(*), WHERE error IS true), average(duration), percentile(duration, 95), max(duration) "+
			"FROM Transaction WHERE appName = %s SINCE %d UNTIL %d",
		nrqlString(app), w.Start.UnixMilli(), w.End.UnixMilli(),
	))
}

// SlowTraces returns the slowest transactions, with their trace IDs, for an
// application within the window.
func (c *Client) SlowTraces(app string, w Window, limit int) ([]map[string]any, error) {
	return c.RunNRQL(fmt.Sprintf(
		"SELECT traceId, name, duration, error FROM Transaction "+
			"WHERE appName = %s AND traceId IS NOT NULL SINCE %d UNTIL %d ORDER BY duration DESC LIMIT %d",
		nrqlString(app), w.Start.UnixMilli(), w.End.UnixMilli(), limit,
	))
}

// nrqlString quotes a value as an NRQL string literal.
func nrqlString(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
