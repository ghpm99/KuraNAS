package tuyacloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const maxCloudReplyBytes = 1 << 20

func (client *Client) signedGet(ctx context.Context, credentials Credentials, accessToken string, pathWithQuery string) (envelope, error) {
	baseURL, err := client.baseURL(credentials.Region)
	if err != nil {
		return envelope{}, err
	}

	signing := signedRequest{
		method:         http.MethodGet,
		pathWithQuery:  pathWithQuery,
		body:           nil,
		accessID:       credentials.AccessID,
		accessSecret:   credentials.AccessSecret,
		accessToken:    accessToken,
		timestampMilli: strconv.FormatInt(client.now().UnixMilli(), 10),
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+pathWithQuery, bytes.NewReader(nil))
	if err != nil {
		return envelope{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	request.Header.Set("client_id", credentials.AccessID)
	request.Header.Set("sign", signing.signature())
	request.Header.Set("t", signing.timestampMilli)
	request.Header.Set("sign_method", "HMAC-SHA256")
	if accessToken != "" {
		request.Header.Set("access_token", accessToken)
	}

	body, err := client.readBody(request, maxCloudReplyBytes)
	if err != nil {
		return envelope{}, err
	}
	return decodeEnvelope(body)
}

func (client *Client) readBody(request *http.Request, maxBytes int64) ([]byte, error) {
	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreachable, withoutRequestURL(err))
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrInvalidResponse, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("%w: body too large", ErrInvalidResponse)
	}
	return body, nil
}

func withoutRequestURL(failure error) error {
	var urlFailure *url.Error
	if errors.As(failure, &urlFailure) {
		return urlFailure.Err
	}
	return failure
}
