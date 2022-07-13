package jwt

import (
	"fmt"
	"terralist/pkg/auth"
	"time"

	_jwt "github.com/golang-jwt/jwt"
)

const (
	oneDay = 24 * time.Hour
)

// JWT handles the creation and extraction of a jwt
type JWT interface {
	Build(auth.User) (string, error)
	Extract(string) (*auth.User, error)
}

// defaultJWT is the concrete implementation of JWT
type defaultJWT struct {
	tokenSigningSecret []byte
}

func New(secret string) (JWT, error) {
	return &defaultJWT{
		tokenSigningSecret: []byte(secret),
	}, nil
}

func (th *defaultJWT) Build(user auth.User) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, _jwt.MapClaims{
		"name":  user.Name,
		"email": user.Email,
		"exp":   time.Now().Add(oneDay).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(th.tokenSigningSecret)
	if err != nil {
		return "", fmt.Errorf("unable to generate token: %w", err)
	}

	return tokenString, nil
}

func (th *defaultJWT) Extract(t string) (*auth.User, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := _jwt.Parse(t, func(token *_jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*_jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return th.tokenSigningSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to parse token: %w", err)
	}

	claims, ok := token.Claims.(_jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unable to get claims from token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().Unix() > int64(claims["exp"].(float64)) {
		return nil, fmt.Errorf("token expired")
	}

	return &auth.User{
		Name:  claims["name"].(string),
		Email: claims["email"].(string),
	}, nil
}
