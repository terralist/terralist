package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"terralist/pkg/auth"

	_jwt "github.com/golang-jwt/jwt"
)

var (
	ErrTokenExpired    = errors.New("token expired")
	ErrTokenNotActive  = errors.New("token not active")
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenGeneration = errors.New("could not generate token")
)

type Serializer interface{}

type Extractor = func(_jwt.Claims) Serializer

// JWT handles the creation and extraction of a jwt.
type JWT interface {
	// Build generates and sign a token for a given data object
	// The first parameter represents the data for which a token
	// should be granted
	// The second parameter represents the number of seconds after
	// which the token should expire
	Build(Serializer, int) (string, error)

	// Extract is the reverse method for Build, which extracts
	// the data from a given token
	// If the token is expired, it will return an error
	Extract(string) (Serializer, error)
}

// defaultJWT is the concrete implementation of JWT.
type defaultJWT struct {
	tokenSigningSecret []byte
}

func New(secret string) (JWT, error) {
	return &defaultJWT{
		tokenSigningSecret: []byte(secret),
	}, nil
}

type TokenClaims struct {
	_jwt.StandardClaims
	Data json.RawMessage `json:"data,omitempty"`
}

func (th *defaultJWT) Build(data Serializer, expireIn int) (string, error) {
	var exp int64
	if expireIn <= 0 {
		exp = 0
	} else {
		exp = time.Now().Add(time.Duration(expireIn) * time.Second).Unix()
	}

	if data == nil {
		data = NoData{}
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, &TokenClaims{
		_jwt.StandardClaims{
			ExpiresAt: exp,
		},
		payload,
	})

	tokenString, err := token.SignedString(th.tokenSigningSecret)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	return tokenString, nil
}

func (th *defaultJWT) Extract(t string) (Serializer, error) {
	token, err := _jwt.ParseWithClaims(t, &TokenClaims{}, func(token *_jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*_jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, token.Header["alg"])
		}

		return th.tokenSigningSecret, nil
	})

	if token == nil || !token.Valid {
		ve, ok := err.(*_jwt.ValidationError)

		if ok {
			if ve.Errors&_jwt.ValidationErrorMalformed != 0 {
				return nil, ErrInvalidToken
			} else if ve.Errors&_jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			} else if ve.Errors&_jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotActive
			}
		} else {
			return nil, fmt.Errorf("%w: unable to parse token: %v", ErrInvalidToken, err)
		}
	}

	claims, _ := token.Claims.(*TokenClaims)
	if claims == nil {
		return nil, ErrInvalidToken
	}

	if len(claims.Data) == 0 {
		return NoData{}, nil
	}

	// First try to decode as an auth user payload.
	var user auth.User
	if err := json.Unmarshal(claims.Data, &user); err == nil && (user.Name != "" || user.Email != "" || user.AuthorityID != "" || user.Authority != "" || len(user.Groups) > 0) {
		return &user, nil
	}

	return NoData{}, nil
}

type NoData struct{}

func (o NoData) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

func (o NoData) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &o)
}
