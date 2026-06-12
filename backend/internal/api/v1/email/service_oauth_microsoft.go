package email

import (
	"encoding/json"
	"net/url"
	"time"
)

// microsoftScopes is exactly read-only mail access plus offline_access (for
// the refresh token) and the identity claims needed to know which address was
// linked. No send/modify capability, ever (hard rule of the e-mail feature).
const microsoftScopes = "https://graph.microsoft.com/Mail.Read offline_access openid email profile"

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Error           string `json:"error"`
}

// StartMicrosoftDeviceCode begins a Device Code Flow link: it asks Microsoft
// for a user code, returns it for display, and polls the token endpoint in a
// goroutine until the user finishes (or the code expires). Starting a new link
// replaces a previous unfinished one.
func (s *Service) StartMicrosoftDeviceCode() (DeviceCodeDto, error) {
	if s.config.MicrosoftClientID == "" {
		return DeviceCodeDto{}, ErrProviderNotConfigured
	}

	form := url.Values{}
	form.Set("client_id", s.config.MicrosoftClientID)
	form.Set("scope", microsoftScopes)

	body, _, err := s.postFormRaw(s.endpoints.MicrosoftDeviceCode, form)
	if err != nil {
		return DeviceCodeDto{}, err
	}

	var response deviceCodeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return DeviceCodeDto{}, oauthError("microsoft device code", "invalid_response")
	}
	if response.Error != "" || response.DeviceCode == "" {
		return DeviceCodeDto{}, oauthError("microsoft device code", response.Error)
	}

	s.setDeviceLinkStatus(DeviceCodePending)

	interval := time.Duration(response.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	expiresAt := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	go s.pollMicrosoftToken(response.DeviceCode, interval, expiresAt)

	return DeviceCodeDto{
		UserCode:        response.UserCode,
		VerificationURI: response.VerificationURI,
		ExpiresIn:       response.ExpiresIn,
	}, nil
}

// MicrosoftDeviceCodeStatus reports the progress of the in-flight (or last)
// device-code link.
func (s *Service) MicrosoftDeviceCodeStatus() DeviceCodeStatusDto {
	s.mu.Lock()
	defer s.mu.Unlock()
	return DeviceCodeStatusDto{Status: s.deviceLink.status}
}

func (s *Service) setDeviceLinkStatus(status string) {
	s.mu.Lock()
	s.deviceLink.status = status
	s.mu.Unlock()
}

// pollMicrosoftToken drives the device-code grant to completion. RFC 8628:
// authorization_pending → keep polling; slow_down → stretch the interval;
// expired_token → give up.
func (s *Service) pollMicrosoftToken(deviceCode string, interval time.Duration, expiresAt time.Time) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("client_id", s.config.MicrosoftClientID)
	form.Set("device_code", deviceCode)

	for {
		if time.Now().After(expiresAt) {
			s.setDeviceLinkStatus(DeviceCodeExpired)
			return
		}

		response, _, err := s.postForm(s.endpoints.MicrosoftToken, form)
		if err != nil {
			s.setDeviceLinkStatus(DeviceCodeError)
			return
		}

		switch {
		case response.AccessToken != "":
			s.finishMicrosoftLink(response)
			return
		case response.Error == "authorization_pending":
			// user has not finished yet — keep waiting
		case response.Error == "slow_down":
			interval += 5 * time.Second
		case response.Error == "expired_token":
			s.setDeviceLinkStatus(DeviceCodeExpired)
			return
		default:
			s.setDeviceLinkStatus(DeviceCodeError)
			return
		}

		time.Sleep(interval)
	}
}

func (s *Service) finishMicrosoftLink(response tokenResponse) {
	address, err := emailFromIDToken(response.IDToken)
	if err != nil {
		s.setDeviceLinkStatus(DeviceCodeError)
		return
	}

	if err := s.persistLinkedAccount(ProviderMicrosoft, address, displayNameFromIDToken(response.IDToken), response.toTokenSet("")); err != nil {
		s.setDeviceLinkStatus(DeviceCodeError)
		return
	}

	s.setDeviceLinkStatus(DeviceCodeLinked)
}

func (s *Service) refreshMicrosoftTokens(tokens TokenSet) (TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokens.RefreshToken)
	form.Set("client_id", s.config.MicrosoftClientID)
	form.Set("scope", microsoftScopes)

	response, _, err := s.postForm(s.endpoints.MicrosoftToken, form)
	if err != nil {
		return TokenSet{}, err
	}
	if response.Error != "" || response.AccessToken == "" {
		return TokenSet{}, oauthError("microsoft refresh", response.Error)
	}

	return response.toTokenSet(tokens.RefreshToken), nil
}
