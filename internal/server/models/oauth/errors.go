package oauth

type ErrorKind = string

var (
	InvalidRequest          = "invalid_request"
	AccessDenied            = "access_denied"
	UnauthorizedClient      = "unauthorized_client"
	UnsupportedResponseType = "unsupported_response_type"
	InvalidScope            = "invalid_scope"
	ServerError             = "server_error"
	TemporarilyUnavailable  = "temporarily_unavailable"
)

type Error interface {
	error
	Kind() ErrorKind
}

type defaultError struct {
	err  error
	kind ErrorKind
}

func (e *defaultError) Error() string {
	return e.err.Error()
}

func (e *defaultError) Kind() string {
	return e.kind
}

func WrapError(err error, kind ErrorKind) Error {
	return &defaultError{
		err:  err,
		kind: kind,
	}
}
