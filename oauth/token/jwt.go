package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/valentindeaconu/terralist/oauth"
	"github.com/valentindeaconu/terralist/settings"
)

var (
	oneDay = 24 * time.Hour
)

func Generate(userDetails oauth.UserDetails) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name":  userDetails.Name,
		"email": userDetails.Email,
		"exp":   time.Now().Add(oneDay).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(settings.TokenSigningSecret)
	if err != nil {
		return "", fmt.Errorf("unable to generate token: %w", err)
	}

	return tokenString, nil
}

func Validate(t string) (oauth.UserDetails, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return settings.TokenSigningSecret, nil
	})

	if err != nil {
		return oauth.UserDetails{}, fmt.Errorf("unable to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return oauth.UserDetails{}, fmt.Errorf("unable to get claims from token")
	}

	if !token.Valid {
		return oauth.UserDetails{}, fmt.Errorf("invalid token")
	}

	if time.Now().Unix() > int64(claims["exp"].(float64)) {
		return oauth.UserDetails{}, fmt.Errorf("token expired")
	}

	userDetails := oauth.UserDetails{
		Name:  claims["name"].(string),
		Email: claims["email"].(string),
	}

	return userDetails, nil
}
