package jwt

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"terralist/pkg/auth"

	"github.com/google/uuid"
	"github.com/mazen160/go-random"
)

var (
	nextUser = func() auth.User {
		user, _ := random.String(12)
		emailHost, _ := random.String(6)
		emailDomain, _ := random.String(3)

		email := fmt.Sprintf("%s@%s.%s", user, emailHost, emailDomain)

		id, _ := uuid.NewRandom()

		return auth.User{
			Name:        user,
			Email:       email,
			AuthorityID: id.String(),
		}
	}

	nextExpire = func() int {
		e, _ := random.IntRange(0, 3600)
		return e
	}

	nextSecret = func() string {
		secret, _ := random.String(16)
		return secret
	}
)

func TestJWT_User(t *testing.T) {
	j, _ := New(nextSecret())
	user := nextUser()
	expireIn := nextExpire()

	token, err := j.Build(user, expireIn)
	if err != nil {
		t.Fatalf("build returned with error: %v", err)
	}

	data, err := j.Extract(token)
	if err != nil {
		t.Fatalf("extract returned with error: %v", err)
	}

	u, ok := data.(auth.User)
	if !ok {
		t.Fatalf("token data was not of type auth.User, but %T", data)
	}

	if u != user {
		t.Fatalf("user mismatch, expected = %v, got = %v", user, u)
	}
}

func TestJWT_TokenExpired(t *testing.T) {
	j, _ := New(nextSecret())
	user := nextUser()
	expireIn := 1 // 1 second

	token, err := j.Build(user, expireIn)
	if err != nil {
		t.Errorf("build returned with error: %v", err)
	}

	time.Sleep(2 * time.Second) // wait for token to expire

	_, err = j.Extract(token)
	if err == nil {
		t.Fatal("extract returned with no error, expected token expiration error")
	}

	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("extract returned with error: %v, expected token expiration error", err)
	}
}

func TestJWT_InfiniteToken(t *testing.T) {
	j, _ := New(nextSecret())
	user := nextUser()
	expireIn := 0 // no expiration

	token, err := j.Build(user, expireIn)
	if err != nil {
		t.Errorf("build returned with error: %v", err)
	}

	_, err = j.Extract(token)
	if err != nil {
		t.Fatalf("extract returned with error: %v, expected no error", err)
	}
}
