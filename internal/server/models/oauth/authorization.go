package oauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/valentindeaconu/terralist/settings"
)

type AuthorizationRequest struct {
	ClientID            string `json:"client_id"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectUri         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	State               string `json:"state"`
}

type AuthorizationCodeComponents struct {
	Key                 string `json:"key"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	UserName            string `json:"user_name"`
	UserEmail           string `json:"user_email"`
}

func (m *AuthorizationRequest) ToPayload() (string, error) {
	data, err := json.Marshal(*m)

	if err != nil {
		return "", err
	}

	salted := fmt.Sprintf("%s/%s", settings.EncryptSalt, string(data))
	state := base64.StdEncoding.EncodeToString([]byte(salted))

	return state, nil
}

func AuthorizationRequestFromPayload(payload string) (AuthorizationRequest, error) {
	salted, err := base64.StdEncoding.DecodeString(payload)

	if err != nil {
		return AuthorizationRequest{}, err
	}

	saltedStr := string(salted)

	data := saltedStr[len(settings.EncryptSalt)+1:]

	var request AuthorizationRequest
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return AuthorizationRequest{}, err
	}

	return request, nil
}

func (m *AuthorizationCodeComponents) ToPayload() (string, error) {
	data, err := json.Marshal(*m)

	if err != nil {
		return "", err
	}

	salted := fmt.Sprintf("%s/%s", settings.EncryptSalt, string(data))
	state := base64.StdEncoding.EncodeToString([]byte(salted))

	return state, nil
}

func AuthorizationCodeFromPayload(payload string) (AuthorizationCodeComponents, error) {
	salted, err := base64.StdEncoding.DecodeString(payload)

	if err != nil {
		return AuthorizationCodeComponents{}, err
	}

	saltedStr := string(salted)

	data := saltedStr[len(settings.EncryptSalt)+1:]

	var request AuthorizationCodeComponents
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return AuthorizationCodeComponents{}, err
	}

	return request, nil
}
