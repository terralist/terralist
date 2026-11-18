package repositories

import (
	"path/filepath"
	"testing"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOAuthRefreshTokenRepository(t *testing.T) *DefaultOAuthRefreshTokenRepository {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	require.NoError(t, err)
	require.NoError(t, engine.Handler().AutoMigrate(&oauth.RefreshToken{}))

	return &DefaultOAuthRefreshTokenRepository{Database: engine}
}

func TestOAuthRefreshTokenRepository_Rotate(t *testing.T) {
	repo := newTestOAuthRefreshTokenRepository(t)
	expiresAt := time.Now().Add(time.Hour)

	require.NoError(t, repo.Create(&oauth.RefreshToken{
		ID:            "session",
		TokenHash:     "hash-1",
		ClientID:      "oauth-client",
		Subject:       "jdoe",
		UpstreamGrant: "sealed",
		ExpiresAt:     expiresAt,
	}))

	claimed, err := repo.Rotate("session", "hash-1", "hash-2", "hash-1", expiresAt)
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = repo.Rotate("session", "hash-1", "hash-3", "hash-1", expiresAt)
	require.NoError(t, err)
	assert.False(t, claimed, "a second rotation from the same token must lose")

	current, err := repo.FindByHash("hash-2")
	require.NoError(t, err)
	assert.Equal(t, "session", current.ID)

	previous, err := repo.FindByPreviousHash("hash-1")
	require.NoError(t, err)
	assert.Equal(t, "session", previous.ID)

	_, err = repo.FindByPreviousHash("")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestOAuthRefreshTokenRepository_CreateRemovesExpired(t *testing.T) {
	repo := newTestOAuthRefreshTokenRepository(t)
	now := time.Now()
	repo.now = func() time.Time { return now }

	sessionEnd := now.Add(-time.Minute)
	for _, token := range []oauth.RefreshToken{
		{ID: "expired", TokenHash: "a", ExpiresAt: now.Add(-time.Minute)},
		{ID: "session-ended", TokenHash: "b", ExpiresAt: now.Add(time.Hour), SessionExpiresAt: &sessionEnd},
		{ID: "valid", TokenHash: "c", ExpiresAt: now.Add(time.Hour)},
	} {
		token.ClientID, token.Subject, token.UpstreamGrant = "oauth-client", "jdoe", "sealed"
		require.NoError(t, repo.Database.Handler().Create(&token).Error)
	}

	require.NoError(t, repo.Create(&oauth.RefreshToken{
		ID: "new", TokenHash: "d", ClientID: "oauth-client", Subject: "jdoe", UpstreamGrant: "sealed", ExpiresAt: now.Add(time.Hour),
	}))

	for hash, found := range map[string]bool{"a": false, "b": false, "c": true, "d": true} {
		_, err := repo.FindByHash(hash)
		assert.Equal(t, found, err == nil, hash)
	}

	require.NoError(t, repo.DeleteAll())
	_, err := repo.FindByHash("c")
	assert.ErrorIs(t, err, ErrNotFound)
}
