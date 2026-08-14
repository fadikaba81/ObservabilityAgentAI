package newrelic

import (
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
	accountID  string
}

func NewClient(endpoint string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		endpoint:   "https://api.newrelic.com/graphql",
		apiKey:     os.Getenv("NEWRELIC_APIKEY_SANDBOX"),
		accountID:  os.Getenv("NEWRELIC_ACCOUNTID_SANDBOX"),
	}

}
