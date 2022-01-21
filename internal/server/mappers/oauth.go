package mappers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	models "github.com/valentindeaconu/terralist/internal/server/models/oauth"
	"github.com/valentindeaconu/terralist/internal/server/utils"
)

type OAuthMapper struct {
	Keychain *utils.Keychain
}

func (o *OAuthMapper) AuthorizationRequestToPayload(authorizationRequest models.AuthorizationRequest) (string, error) {
	data, err := json.Marshal(authorizationRequest)

	if err != nil {
		return "", err
	}

	salted := fmt.Sprintf("%s/%s", o.Keychain.EncryptSalt, string(data))
	state := base64.StdEncoding.EncodeToString([]byte(salted))

	return state, nil
}

func (o *OAuthMapper) PayloadToAuthorizationRequest(payload string) (models.AuthorizationRequest, error) {
	salted, err := base64.StdEncoding.DecodeString(payload)

	if err != nil {
		return models.AuthorizationRequest{}, err
	}

	saltedStr := string(salted)

	data := saltedStr[len(o.Keychain.EncryptSalt)+1:]

	var request models.AuthorizationRequest
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return models.AuthorizationRequest{}, err
	}

	return request, nil
}

func (o *OAuthMapper) AuthorizationCodeComponentsToPayload(codeComponents models.AuthorizationCodeComponents) (string, error) {
	data, err := json.Marshal(codeComponents)

	if err != nil {
		return "", err
	}

	salted := fmt.Sprintf("%s/%s", o.Keychain.EncryptSalt, string(data))
	state := base64.StdEncoding.EncodeToString([]byte(salted))

	return state, nil
}

func (o *OAuthMapper) PayloadToAuthorizationCodeComponents(payload string) (models.AuthorizationCodeComponents, error) {
	salted, err := base64.StdEncoding.DecodeString(payload)

	if err != nil {
		return models.AuthorizationCodeComponents{}, err
	}

	saltedStr := string(salted)

	data := saltedStr[len(o.Keychain.EncryptSalt)+1:]

	var request models.AuthorizationCodeComponents
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return models.AuthorizationCodeComponents{}, err
	}

	return request, nil
}
