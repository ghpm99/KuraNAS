package tuyacloud

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type signedRequest struct {
	method         string
	pathWithQuery  string
	body           []byte
	accessID       string
	accessSecret   string
	accessToken    string
	timestampMilli string
}

func (request signedRequest) signature() string {
	bodyDigest := sha256.Sum256(request.body)
	stringToSign := request.method + "\n" + hex.EncodeToString(bodyDigest[:]) + "\n" + "\n" + request.pathWithQuery

	signer := hmac.New(sha256.New, []byte(request.accessSecret))
	signer.Write([]byte(request.accessID + request.accessToken + request.timestampMilli + stringToSign))
	return strings.ToUpper(hex.EncodeToString(signer.Sum(nil)))
}
