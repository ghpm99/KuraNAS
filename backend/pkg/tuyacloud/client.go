package tuyacloud

import (
	"net/http"
	"sync"
	"time"
)

// Client reads sweeping robot maps from the Tuya cloud. It caches the access token per credentials.
type Client struct {
	httpClient      *http.Client
	baseURLOverride string
	now             func() time.Time

	tokenMutex sync.Mutex
	tokens     map[Credentials]cachedToken
}

// NewClient creates a Client that talks to the regional Tuya endpoints through httpClient.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient, now: time.Now, tokens: map[Credentials]cachedToken{}}
}

// NewClientWithBaseURL creates a Client that sends every cloud call to baseURL regardless of the region.
// It exists so tests can point the client at a local server.
func NewClientWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	client := NewClient(httpClient)
	client.baseURLOverride = baseURL
	return client
}

func (client *Client) baseURL(region string) (string, error) {
	if client.baseURLOverride != "" {
		return client.baseURLOverride, nil
	}
	return baseURLForRegion(region)
}
