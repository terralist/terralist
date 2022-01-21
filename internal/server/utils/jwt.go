package utils

import (
	"fmt"
	"time"

	_jwt "github.com/golang-jwt/jwt"
	models "github.com/valentindeaconu/terralist/internal/server/models/oauth"
)

type JWT struct {
	Keychain *Keychain
}

const (
	oneDay = 24 * time.Hour
)

func (th *JWT) Generate(userDetails models.UserDetails) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, _jwt.MapClaims{
		"name":  userDetails.Name,
		"email": userDetails.Email,
		"exp":   time.Now().Add(oneDay).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(th.Keychain.TokenSigningSecret)
	if err != nil {
		return "", fmt.Errorf("unable to generate token: %w", err)
	}

	return tokenString, nil
}

func (th *JWT) Validate(t string) (models.UserDetails, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := _jwt.Parse(t, func(token *_jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*_jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return th.Keychain.TokenSigningSecret, nil
	})

	if err != nil {
		return models.UserDetails{}, fmt.Errorf("unable to parse token: %w", err)
	}

	claims, ok := token.Claims.(_jwt.MapClaims)
	if !ok {
		return models.UserDetails{}, fmt.Errorf("unable to get claims from token")
	}

	if !token.Valid {
		return models.UserDetails{}, fmt.Errorf("invalid token")
	}

	if time.Now().Unix() > int64(claims["exp"].(float64)) {
		return models.UserDetails{}, fmt.Errorf("token expired")
	}

	userDetails := models.UserDetails{
		Name:  claims["name"].(string),
		Email: claims["email"].(string),
	}

	return userDetails, nil
}
