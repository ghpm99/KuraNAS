package tuyacloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	tokenPath         = "/v1.0/token?grant_type=1"
	tokenSafetyMargin = 5 * time.Minute
)

type cachedToken struct {
	value     string
	expiresAt time.Time
}

func (token cachedToken) isUsableAt(instant time.Time) bool {
	return token.value != "" && instant.Before(token.expiresAt)
}

func (client *Client) accessToken(ctx context.Context, credentials Credentials) (string, error) {
	client.tokenMutex.Lock()
	defer client.tokenMutex.Unlock()

	if cached := client.tokens[credentials]; cached.isUsableAt(client.now()) {
		return cached.value, nil
	}

	fetched, err := client.fetchToken(ctx, credentials)
	if err != nil {
		return "", err
	}
	client.tokens[credentials] = fetched
	return fetched.value, nil
}

func (client *Client) forgetToken(credentials Credentials) {
	client.tokenMutex.Lock()
	defer client.tokenMutex.Unlock()
	delete(client.tokens, credentials)
}

func (client *Client) fetchToken(ctx context.Context, credentials Credentials) (cachedToken, error) {
	reply, err := client.signedGet(ctx, credentials, "", tokenPath)
	if err != nil {
		return cachedToken{}, err
	}
	if !reply.IsSuccess {
		return cachedToken{}, fmt.Errorf("%w: token refused with code %d", ErrCredentialsRejected, reply.Code)
	}

	var granted struct {
		AccessToken   string `json:"access_token"`
		ExpireSeconds int    `json:"expire_time"`
	}
	if err := json.Unmarshal(reply.Result, &granted); err != nil || granted.AccessToken == "" {
		return cachedToken{}, errors.Join(ErrInvalidResponse, errors.New("token reply has no access_token"))
	}

	lifetime := time.Duration(granted.ExpireSeconds)*time.Second - tokenSafetyMargin
	return cachedToken{value: granted.AccessToken, expiresAt: client.now().Add(lifetime)}, nil
}
