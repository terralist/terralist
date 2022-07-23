package jwt

import (
	"fmt"
	"terralist/pkg/auth"
	"time"

	_jwt "github.com/golang-jwt/jwt"
)

// JWT handles the creation and extraction of a jwt
type JWT interface {
	// Build generates and sign a token for a given user
	// The first parameter represents the user for which a token
	// should be granted
	// The second parameter represents the number of seconds after
	// which the token should expire
	Build(auth.User, int) (string, error)

	// Extract is the reverse method for Build, which extracts
	// the user data from a given token
	// If the token is expired, it will return an error
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

func (th *defaultJWT) Build(user auth.User, expireIn int) (string, error) {
	// Allow no expiration date
	var exp int64
	if expireIn == 0 {
		exp = 0
	} else {
		exp = time.Now().Add(time.Duration(expireIn) * time.Second).Unix()
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, _jwt.MapClaims{
		"user": &user,
		"exp":  exp,
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

	exp := claims["exp"].(int64)
	if exp != 0 && time.Now().Unix() > exp {
		return nil, fmt.Errorf("token expired")
	}

	return claims["user"].(*auth.User), nil
}
