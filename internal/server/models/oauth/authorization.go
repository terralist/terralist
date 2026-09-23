package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidPayload = errors.New("invalid payload")
)

type Payload string

func (p Payload) String() string {
	return string(p)
}

func (p Payload) ToRequest(key []byte) (Request, error) {
	signed, err := base64.StdEncoding.DecodeString(p.String())
	if err != nil {
		return Request{}, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	if len(signed) < sha256.Size {
		return Request{}, fmt.Errorf("%w: payload too short", ErrInvalidPayload)
	}

	signature, data := signed[:sha256.Size], signed[sha256.Size:]

	if !hmac.Equal(signature, sign(key, data)) {
		return Request{}, fmt.Errorf("%w: signature mismatch", ErrInvalidPayload)
	}

	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		return Request{}, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	return request, nil
}

type Request struct {
	ClientID            string `json:"client_id"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	State               string `json:"state"`
}

func (r Request) ToPayload(key []byte) (Payload, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	signed := append(sign(key, data), data...)

	return Payload(base64.StdEncoding.EncodeToString(signed)), nil
}

func sign(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)

	return mac.Sum(nil)
}

type CodeComponents struct {
	CodeChallenge       string   `json:"code_challenge"`
	CodeChallengeMethod string   `json:"code_challenge_method"`
	UserName            string   `json:"user_name"`
	UserEmail           string   `json:"user_email"`
	UserGroups          []string `json:"user_groups,omitempty"`
	UserAuthority       string   `json:"user_authority,omitempty"`
	UserAuthorityID     string   `json:"user_authority_id,omitempty"`
}
