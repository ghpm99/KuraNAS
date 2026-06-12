package email

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"time"
)

// googleScopes is exactly read-only mail access plus the identity claims
// (openid/email) needed to know which address was linked. No send/modify
// capability, ever (hard rule of the e-mail feature).
const googleScopes = "https://www.googleapis.com/auth/gmail.readonly openid email"

// googleRedirectURI uses the loopback host on the API port: Google's Device
// Flow does not accept Gmail scopes, so the link is done in a browser on the
// NAS machine itself (or through an SSH tunnel).
const googleRedirectURI = "http://localhost:8000/api/v1/email/oauth/google/callback"

const pkceStateTTL = 10 * time.Minute

// GoogleAuthURL builds the consent URL for the Authorization Code + PKCE flow
// and parks the verifier server-side under a one-time state.
func (s *Service) GoogleAuthURL() (GoogleAuthURLDto, error) {
	if s.config.GoogleClientID == "" || s.config.GoogleClientSecret == "" {
		return GoogleAuthURLDto{}, ErrProviderNotConfigured
	}

	state, err := randomToken(24)
	if err != nil {
		return GoogleAuthURLDto{}, err
	}
	verifier, err := randomToken(48)
	if err != nil {
		return GoogleAuthURLDto{}, err
	}

	s.mu.Lock()
	for key, pending := range s.pkceStates {
		if time.Now().After(pending.expiresAt) {
			delete(s.pkceStates, key)
		}
	}
	s.pkceStates[state] = pkceState{verifier: verifier, expiresAt: time.Now().Add(pkceStateTTL)}
	s.mu.Unlock()

	challenge := sha256.Sum256([]byte(verifier))

	params := url.Values{}
	params.Set("client_id", s.config.GoogleClientID)
	params.Set("redirect_uri", googleRedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", googleScopes)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("state", state)
	params.Set("code_challenge", base64.RawURLEncoding.EncodeToString(challenge[:]))
	params.Set("code_challenge_method", "S256")

	return GoogleAuthURLDto{AuthURL: s.endpoints.GoogleAuth + "?" + params.Encode()}, nil
}

// HandleGoogleCallback consumes the one-time state, exchanges the code for
// tokens and persists the linked account.
func (s *Service) HandleGoogleCallback(state string, code string) error {
	if s.config.GoogleClientID == "" || s.config.GoogleClientSecret == "" {
		return ErrProviderNotConfigured
	}

	s.mu.Lock()
	pending, found := s.pkceStates[state]
	delete(s.pkceStates, state)
	s.mu.Unlock()

	if !found || time.Now().After(pending.expiresAt) {
		return ErrInvalidOAuthState
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", s.config.GoogleClientID)
	form.Set("client_secret", s.config.GoogleClientSecret)
	form.Set("redirect_uri", googleRedirectURI)
	form.Set("code_verifier", pending.verifier)

	response, _, err := s.postForm(s.endpoints.GoogleToken, form)
	if err != nil {
		return err
	}
	if response.Error != "" || response.AccessToken == "" {
		return oauthError("google token exchange", response.Error)
	}

	address, err := emailFromIDToken(response.IDToken)
	if err != nil {
		return err
	}

	return s.persistLinkedAccount(ProviderGoogle, address, displayNameFromIDToken(response.IDToken), response.toTokenSet(""))
}

func (s *Service) refreshGoogleTokens(tokens TokenSet) (TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokens.RefreshToken)
	form.Set("client_id", s.config.GoogleClientID)
	form.Set("client_secret", s.config.GoogleClientSecret)

	response, _, err := s.postForm(s.endpoints.GoogleToken, form)
	if err != nil {
		return TokenSet{}, err
	}
	if response.Error != "" || response.AccessToken == "" {
		return TokenSet{}, oauthError("google refresh", response.Error)
	}

	return response.toTokenSet(tokens.RefreshToken), nil
}
