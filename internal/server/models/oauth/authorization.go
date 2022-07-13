package oauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Payload string

func (p Payload) String() string {
	return string(p)
}

func (p Payload) ToRequest(salt string) (Request, error) {
	salted, err := base64.StdEncoding.DecodeString(p.String())

	if err != nil {
		return Request{}, err
	}

	saltedStr := string(salted)

	data := saltedStr[len(salt)+1:]

	var request Request
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return Request{}, err
	}

	return request, nil
}

func (p Payload) ToCodeComponents(salt string) (CodeComponents, error) {
	salted, err := base64.StdEncoding.DecodeString(p.String())

	if err != nil {
		return CodeComponents{}, err
	}

	saltedStr := string(salted)

	data := saltedStr[len(salt)+1:]

	var codeComponents CodeComponents
	if err := json.Unmarshal([]byte(data), &codeComponents); err != nil {
		return CodeComponents{}, err
	}

	return codeComponents, nil
}

type Request struct {
	ClientID            string `json:"client_id"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectUri         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	State               string `json:"state"`
}

func (r Request) ToPayload(salt string) (Payload, error) {
	data, err := json.Marshal(r)

	if err != nil {
		return "", err
	}

	salted := fmt.Sprintf("%s/%s", salt, string(data))
	state := base64.StdEncoding.EncodeToString([]byte(salted))

	return Payload(state), nil
}

type CodeComponents struct {
	Key                 string `json:"key"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	UserName            string `json:"user_name"`
	UserEmail           string `json:"user_email"`
}

func (c CodeComponents) ToPayload(salt string) (Payload, error) {
	data, err := json.Marshal(c)

	if err != nil {
		return "", err
	}

	salted := fmt.Sprintf("%s/%s", salt, string(data))
	state := base64.StdEncoding.EncodeToString([]byte(salted))

	return Payload(state), nil
}
