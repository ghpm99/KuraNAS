package tuyacloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	realtimeMapPathTemplate = "/v1.0/users/sweepers/file/%s/realtime-map"
	layoutMapType           = 0
	maxLayoutFileBytes      = 4 << 20
)

// ReadMapLayout fetches the current layout map of the sweeping robot deviceID from the cloud.
// It returns ErrUnreachable, ErrCredentialsRejected, ErrServiceNotAuthorized or ErrInvalidResponse
// (wrapped) on failure.
func (client *Client) ReadMapLayout(ctx context.Context, credentials Credentials, deviceID string) (MapLayout, error) {
	layoutURL, err := client.findLayoutURL(ctx, credentials, deviceID)
	if err != nil {
		return MapLayout{}, err
	}
	return client.downloadLayout(ctx, layoutURL)
}

func (client *Client) findLayoutURL(ctx context.Context, credentials Credentials, deviceID string) (string, error) {
	accessToken, err := client.accessToken(ctx, credentials)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf(realtimeMapPathTemplate, url.PathEscape(deviceID))
	reply, err := client.signedGet(ctx, credentials, accessToken, path)
	if err != nil {
		return "", err
	}
	if err := reply.failure(); err != nil {
		client.forgetTokenIfRejected(credentials, err)
		return "", err
	}
	return layoutURLFromResult(reply.Result)
}

func (client *Client) forgetTokenIfRejected(credentials Credentials, failure error) {
	if errors.Is(failure, ErrCredentialsRejected) {
		client.forgetToken(credentials)
	}
}

func layoutURLFromResult(result json.RawMessage) (string, error) {
	var maps []struct {
		MapType int    `json:"map_type"`
		MapURL  string `json:"map_url"`
	}
	if err := json.Unmarshal(result, &maps); err != nil {
		return "", fmt.Errorf("%w: realtime-map result is not a list", ErrInvalidResponse)
	}
	for _, candidate := range maps {
		if candidate.MapType != layoutMapType {
			continue
		}
		if !strings.HasPrefix(candidate.MapURL, "https://") {
			return "", fmt.Errorf("%w: map url is not https", ErrInvalidResponse)
		}
		return candidate.MapURL, nil
	}
	return "", fmt.Errorf("%w: no layout map in the reply", ErrInvalidResponse)
}

func (client *Client) downloadLayout(ctx context.Context, layoutURL string) (MapLayout, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, layoutURL, nil)
	if err != nil {
		return MapLayout{}, fmt.Errorf("%w: map url: %v", ErrInvalidResponse, err)
	}
	body, err := client.readBody(request, maxLayoutFileBytes)
	if err != nil {
		return MapLayout{}, err
	}
	return decodeMapLayout(body)
}
