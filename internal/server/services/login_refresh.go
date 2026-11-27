package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"
	"terralist/pkg/secret"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// codeGrantScope seals an upstream grant while it waits in an authorization
// code. Once a refresh session exists, the grant is sealed for its ID instead.
const codeGrantScope = "oauth-code"

// RefreshOptions configures the refresh tokens issued by the login service.
type RefreshOptions struct {
	Repository repositories.OAuthRefreshTokenRepository

	// Sealer seals the upstream grants at rest.
	Sealer *secret.Sealer

	// AccessTokenExpirationSecs is the lifetime of the access tokens issued
	// together with a refresh token.
	AccessTokenExpirationSecs int

	// TokenExpiration is how long a refresh token stays valid without being
	// used. Each refresh starts it again.
	TokenExpiration time.Duration

	// MaxSessionAge bounds a refresh session from the login, however often it
	// is refreshed. Zero leaves it unbounded.
	MaxSessionAge time.Duration

	Now func() time.Time
}

func (o *RefreshOptions) clock() time.Time {
	if o.Now != nil {
		return o.Now()
	}

	return time.Now()
}

// expiry returns when a refresh token issued now expires, never past the end
// of its session.
func (o *RefreshOptions) expiry(now time.Time, sessionExpiresAt *time.Time) time.Time {
	expiresAt := now.Add(o.TokenExpiration)
	if sessionExpiresAt != nil && sessionExpiresAt.Before(expiresAt) {
		return *sessionExpiresAt
	}

	return expiresAt
}

func (s *DefaultLoginService) RefreshEnabled() bool {
	return s.Refresh != nil
}

// refreshesClient reports whether a login started by the given client keeps
// the upstream grant. Terraform and OpenTofu discard refresh tokens, so their
// logins never keep one.
func (s *DefaultLoginService) refreshesClient(clientID string) bool {
	if s.Refresh == nil || clientID == "" || clientID == oauth.TerraformClientID {
		return false
	}

	_, ok := s.Provider.(auth.Refresher)
	return ok
}

func (s *DefaultLoginService) unpackCodeWithGrant(code string, r *oauth.Request) (*oauth.CodeComponents, oauth.Error) {
	refresher, _ := s.Provider.(auth.Refresher)

	var userDetails auth.User
	grant, err := refresher.ExchangeCode(code, &userDetails)
	if err != nil {
		return nil, oauth.WrapError(err, oauth.AccessDenied)
	}

	components := &oauth.CodeComponents{
		CodeChallenge:       r.CodeChallenge,
		CodeChallengeMethod: r.CodeChallengeMethod,
		UserName:            userDetails.Name,
		UserEmail:           userDetails.Email,
		UserGroups:          userDetails.Groups,
		UserAuthority:       userDetails.Authority,
		UserAuthorityID:     userDetails.AuthorityID,
		ClientID:            r.ClientID,
	}

	if grant == nil {
		log.Warn().
			Str("provider", s.Provider.Name()).
			Str("client_id", r.ClientID).
			Msg("The provider returned no refresh token, so this login gets no Terralist refresh token")

		return components, nil
	}

	sealed, err := s.sealGrant(*grant, codeGrantScope)
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	components.UpstreamGrant = sealed

	return components, nil
}

func (s *DefaultLoginService) issueRefreshToken(components *oauth.CodeComponents, user auth.User) (*oauth.TokenResponse, oauth.Error) {
	grant, err := s.openGrant(components.UpstreamGrant, codeGrantScope)
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	id := uuid.NewString()

	sealed, err := s.sealGrant(grant, id)
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	token, hash, err := newRefreshToken()
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	now := s.Refresh.clock()

	entry := &oauth.RefreshToken{
		ID:            id,
		TokenHash:     hash,
		ClientID:      components.ClientID,
		Subject:       user.Name,
		UpstreamGrant: sealed,
		CreatedAt:     now,
	}

	if s.Refresh.MaxSessionAge > 0 {
		sessionExpiresAt := now.Add(s.Refresh.MaxSessionAge)
		entry.SessionExpiresAt = &sessionExpiresAt
	}

	entry.ExpiresAt = s.Refresh.expiry(now, entry.SessionExpiresAt)

	if err := s.Refresh.Repository.Create(entry); err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	return s.tokenResponse(user, token)
}

func (s *DefaultLoginService) RefreshToken(token string, clientID string) (*oauth.TokenResponse, oauth.Error) {
	if s.Refresh == nil {
		return nil, oauth.WrapError(fmt.Errorf("refresh tokens are not enabled"), oauth.UnsupportedGrantType)
	}

	refresher, ok := s.Provider.(auth.Refresher)
	if !ok {
		return nil, oauth.WrapError(fmt.Errorf("the provider cannot refresh sessions"), oauth.UnsupportedGrantType)
	}

	invalid := oauth.WrapError(fmt.Errorf("refresh token is invalid or expired"), oauth.InvalidGrant)

	hash := hashRefreshToken(token)

	entry, err := s.Refresh.Repository.FindByHash(hash)
	if errors.Is(err, repositories.ErrNotFound) {
		s.revokeReusedToken(hash)
		return nil, invalid
	}
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	if entry.ClientID != clientID {
		return nil, invalid
	}

	now := s.Refresh.clock()
	if now.After(entry.ExpiresAt) || (entry.SessionExpiresAt != nil && now.After(*entry.SessionExpiresAt)) {
		s.deleteSession(entry.ID)
		return nil, invalid
	}

	next, nextHash, err := newRefreshToken()
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	// Rotating before asking the provider makes concurrent refreshes with the
	// same token resolve to a single winner, which matters for providers that
	// rotate their own refresh tokens.
	claimed, err := s.Refresh.Repository.Rotate(entry.ID, hash, nextHash, hash, s.Refresh.expiry(now, entry.SessionExpiresAt))
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}
	if !claimed {
		return nil, invalid
	}

	grant, err := s.openGrant(entry.UpstreamGrant, entry.ID)
	if err != nil {
		s.deleteSession(entry.ID)
		return nil, invalid
	}

	var user auth.User
	nextGrant, err := refresher.Refresh(grant, &user)
	if err != nil {
		if isUpstreamUnavailable(err) {
			// The provider never answered, so the client keeps its token and
			// may retry.
			if _, rerr := s.Refresh.Repository.Rotate(entry.ID, nextHash, hash, entry.PreviousTokenHash, entry.ExpiresAt); rerr != nil {
				log.Error().Err(rerr).Str("session", entry.ID).Msg("Could not restore a refresh session after the provider was unavailable")
			}

			return nil, oauth.WrapError(err, oauth.TemporarilyUnavailable)
		}

		log.Info().
			Err(err).
			Str("session", entry.ID).
			Str("user", entry.Subject).
			Msg("The provider refused to refresh the session, revoking it")

		s.deleteSession(entry.ID)
		return nil, invalid
	}

	if user.Name != entry.Subject {
		log.Warn().
			Str("session", entry.ID).
			Str("user", entry.Subject).
			Str("upstream_user", user.Name).
			Msg("The provider refreshed the session for another user, revoking it")

		s.deleteSession(entry.ID)
		return nil, invalid
	}

	sealed, err := s.sealGrant(*nextGrant, entry.ID)
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	if err := s.Refresh.Repository.UpdateGrant(entry.ID, sealed); err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	return s.tokenResponse(user, next)
}

func (s *DefaultLoginService) RevokeToken(token string, clientID string) oauth.Error {
	if s.Refresh == nil {
		return nil
	}

	entry, err := s.Refresh.Repository.FindByHash(hashRefreshToken(token))
	if errors.Is(err, repositories.ErrNotFound) {
		return nil
	}
	if err != nil {
		return oauth.WrapError(err, oauth.ServerError)
	}

	if entry.ClientID != clientID {
		return nil
	}

	if err := s.Refresh.Repository.Delete(entry.ID); err != nil {
		return oauth.WrapError(err, oauth.ServerError)
	}

	return nil
}

// revokeReusedToken ends the session a rotated-out token belonged to. Only a
// copy of the token can present it again, so the session is treated as
// leaked.
func (s *DefaultLoginService) revokeReusedToken(hash string) {
	entry, err := s.Refresh.Repository.FindByPreviousHash(hash)
	if err != nil {
		return
	}

	log.Warn().
		Str("session", entry.ID).
		Str("user", entry.Subject).
		Str("client_id", entry.ClientID).
		Msg("A rotated-out refresh token was used again, revoking its session")

	s.deleteSession(entry.ID)
}

func (s *DefaultLoginService) deleteSession(id string) {
	if err := s.Refresh.Repository.Delete(id); err != nil {
		log.Error().Err(err).Str("session", id).Msg("Could not delete a refresh session")
	}
}

func (s *DefaultLoginService) tokenResponse(user auth.User, refreshToken string) (*oauth.TokenResponse, oauth.Error) {
	t, err := s.JWT.Build(user, s.Refresh.AccessTokenExpirationSecs)
	if err != nil {
		return nil, oauth.WrapError(err, oauth.ServerError)
	}

	return &oauth.TokenResponse{
		AccessToken:  t,
		TokenType:    "bearer",
		RefreshToken: refreshToken,
		ExpiresIn:    s.Refresh.AccessTokenExpirationSecs,
	}, nil
}

func (s *DefaultLoginService) sealGrant(grant auth.Grant, scope string) (string, error) {
	data, err := json.Marshal(grant)
	if err != nil {
		return "", fmt.Errorf("could not encode the upstream grant: %w", err)
	}

	return s.Refresh.Sealer.Seal(string(data), scope)
}

func (s *DefaultLoginService) openGrant(sealed, scope string) (auth.Grant, error) {
	var grant auth.Grant

	data, err := s.Refresh.Sealer.Open(sealed, scope)
	if err != nil {
		return grant, err
	}

	if err := json.Unmarshal([]byte(data), &grant); err != nil {
		return grant, fmt.Errorf("could not decode the upstream grant: %w", err)
	}

	return grant, nil
}

func isUpstreamUnavailable(err error) bool {
	var netErr net.Error
	return errors.Is(err, auth.ErrUpstreamUnavailable) || errors.As(err, &netErr)
}

func newRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("could not generate refresh token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(buf)

	return token, hashRefreshToken(token), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
