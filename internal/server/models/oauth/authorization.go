package oauth

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
