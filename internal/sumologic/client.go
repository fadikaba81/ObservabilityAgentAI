package sumologic

import (
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
	accessID   string
	accessKey  string
}

func NewClient(endpoint string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		endpoint:   endpoint,
		accessID:   os.Getenv("SUMOLOGIC_ACCESSID_SANDBOX"),
		accessKey:  os.Getenv("SUMOLOGIC_ACCESSKEY_SANDBOX"),
	}

}
