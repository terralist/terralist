package oauth

type AccessResponse struct {
	AccessToken string `json:"access_token"`
}

type TokenValidationRequest struct {
	ClientID     string `form:"client_id"`
	Code         string `form:"code"`
	CodeVerifier string `form:"code_verifier"`
	GrantType    string `form:"grant_type"`
	RedirectUri  string `form:"redirect_uri"`
}
