package oauth

type Error string

const (
	InvalidRequest          = "invalid_request"
	AccessDenied            = "access_denied"
	UnauthorizedClient      = "unauthorized_client"
	UnsupportedResponseType = "unsupported_response_type"
	InvalidScope            = "invalid_scope"
	ServerError             = "server_error"
	TemporarilyUnavailable  = "temporarily_unavailable"
)
