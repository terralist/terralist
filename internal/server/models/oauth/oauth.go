package oauth

type AccessResponse struct {
	AccessToken string `json:"access_token"`
}

type TokenValidationRequest struct {
	ClientID     string `form:"client_id"`
	Code         string `form:"code"`
	CodeVerifier string `form:"code_verifier"`
	GrantType    string `form:"grant_type"`
	RedirectURI  string `form:"redirect_uri"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
