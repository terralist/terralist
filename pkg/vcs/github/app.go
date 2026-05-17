package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

type Client struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey
	HTTP       *http.Client

	mu     sync.Mutex
	cache  map[int64]cached
	client *http.Client
}

type cached struct {
	token string
	until time.Time
}

func NewFromKeyFile(appID int64, keyPath string) (*Client, error) {
	if appID == 0 || keyPath == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read github app private key: %w", err)
	}
	key, err := parseRSAPrivateKey(raw)
	if err != nil {
		return nil, err
	}
	return &Client{
		AppID:      appID,
		PrivateKey: key,
		cache:      make(map[int64]cached),
	}, nil
}

func parseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("github app private key: no PEM block")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse github app private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github app private key is not RSA")
	}
	return rsaKey, nil
}

type appJWTClaims struct {
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
	Issuer    int64 `json:"iss"`
}

func (c *appJWTClaims) Valid() error {
	return nil
}

func (c *Client) appJWT() (string, error) {
	now := time.Now()
	claims := &appJWTClaims{
		IssuedAt:  now.Unix() - 60,
		ExpiresAt: now.Add(9 * time.Minute).Unix(),
		Issuer:    c.AppID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(c.PrivateKey)
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	if c.client != nil {
		return c.client
	}
	c.client = &http.Client{Timeout: 30 * time.Second}
	return c.client
}

type installationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) BearerToken(ctx context.Context, installationID int64) (string, error) {
	if installationID == 0 {
		return "", errors.New("github installation id is required for app authentication")
	}

	c.mu.Lock()
	if ent, ok := c.cache[installationID]; ok && time.Now().Before(ent.until) {
		tok := ent.token
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()

	appTok, err := c.appJWT()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+appTok)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github app access token: %s: %s", resp.Status, string(body))
	}

	var parsed installationTokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.Token == "" {
		return "", errors.New("empty installation token from github")
	}

	until := parsed.ExpiresAt.Add(-1 * time.Minute)
	if until.Before(time.Now()) {
		until = time.Now().Add(5 * time.Minute)
	}

	c.mu.Lock()
	c.cache[installationID] = cached{token: parsed.Token, until: until}
	c.mu.Unlock()

	return parsed.Token, nil
}
