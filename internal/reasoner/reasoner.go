package reasoner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultURL = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
	maxTokens  = 2048
)

const systemPrompt = `You are an SRE incident assistant. You receive an alert, New Relic metrics and slow traces, and Sumo Logic log lines for one service and time window.

Identify the most likely issue and recommend concrete next steps. Use only the evidence provided. If the evidence is insufficient, say so and set confidence to "low". Do not invent services, metrics or log lines.

Respond with only a JSON object, no prose, in this shape:
{"summary": "...", "probable_root_cause": "...", "confidence": "low|medium|high", "evidence": ["..."], "recommended_actions": ["..."]}`

// Analysis is the structured answer returned by the model.
type Analysis struct {
	Summary            string   `json:"summary"`
	ProbableRootCause  string   `json:"probable_root_cause"`
	Confidence         string   `json:"confidence"`
	Evidence           []string `json:"evidence"`
	RecommendedActions []string `json:"recommended_actions"`
}

// Client calls the Anthropic Messages API.
type Client struct {
	apiKey     string
	model      string
	url        string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		url:        defaultURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// Analyze sends the prompt to Claude and parses the JSON analysis. It makes a
// second attempt if the model's reply cannot be parsed.
func (c *Client) Analyze(ctx context.Context, prompt string) (*Analysis, error) {
	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		text, err := c.complete(ctx, prompt)
		if err != nil {
			return nil, err
		}
		analysis, err := parseAnalysis(text)
		if err == nil {
			return analysis, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("model did not return a valid analysis: %w", lastErr)
}

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) complete(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(request{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  []message{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var out response
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || out.Error != nil {
		msg := "no error detail"
		if out.Error != nil {
			msg = out.Error.Message
		}
		return "", fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, msg)
	}

	var sb strings.Builder
	for _, block := range out.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}

// parseAnalysis extracts the JSON object from the reply, tolerating code fences
// or stray text around it.
func parseAnalysis(text string) (*Analysis, error) {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return nil, errors.New("no JSON object in model reply")
	}

	var a Analysis
	if err := json.Unmarshal([]byte(text[start:end+1]), &a); err != nil {
		return nil, err
	}
	if a.Summary == "" {
		return nil, errors.New("analysis has an empty summary")
	}
	return &a, nil
}
