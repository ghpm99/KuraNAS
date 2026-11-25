package tuyacloud

import (
	"encoding/json"
	"fmt"
)

const codeServiceNotAuthorized = 28841101

var credentialRejectionCodes = map[int]bool{
	1004: true,
	1010: true,
	1106: true,
	2009: true,
}

type envelope struct {
	IsSuccess bool            `json:"success"`
	Code      int             `json:"code"`
	Message   string          `json:"msg"`
	Result    json.RawMessage `json:"result"`
}

func decodeEnvelope(body []byte) (envelope, error) {
	var decoded envelope
	if err := json.Unmarshal(body, &decoded); err != nil {
		return envelope{}, fmt.Errorf("%w: body is not JSON", ErrInvalidResponse)
	}
	return decoded, nil
}

func (reply envelope) failure() error {
	if reply.IsSuccess {
		return nil
	}
	if reply.Code == codeServiceNotAuthorized {
		return fmt.Errorf("%w: %s", ErrServiceNotAuthorized, reply.Message)
	}
	if credentialRejectionCodes[reply.Code] {
		return fmt.Errorf("%w: code %d: %s", ErrCredentialsRejected, reply.Code, reply.Message)
	}
	return fmt.Errorf("%w: code %d: %s", ErrInvalidResponse, reply.Code, reply.Message)
}
