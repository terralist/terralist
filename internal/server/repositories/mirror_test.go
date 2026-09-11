package repositories

import (
	"path/filepath"
	"testing"

	"terralist/internal/server/models/mirror"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMirrorRepository(t *testing.T) *DefaultMirrorRepository {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := engine.Handler().AutoMigrate(
		&mirror.Provider{},
		&mirror.Version{},
		&mirror.Platform{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return &DefaultMirrorRepository{
		Database: engine,
	}
}

func newMirrorProvider(hostname, namespace, name string, versions ...string) mirror.Provider {
	p := mirror.Provider{
		Hostname:  hostname,
		Namespace: namespace,
		Name:      name,
	}

	for _, v := range versions {
		p.Versions = append(p.Versions, mirror.Version{
			Version: v,
			Platforms: []mirror.Platform{
				{
					System:       "linux",
					Architecture: "amd64",
					Location:     "mirror/" + hostname + "/" + namespace + "/" + name + "/" + v + "/linux_amd64.zip",
					Hashes:       "h1:abc",
				},
			},
		})
	}

	return p
}

func TestMirrorRepository_Find(t *testing.T) {
	t.Run("returns error when provider does not exist", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		p, err := repo.Find("registry.terraform.io", "hashicorp", "null")

		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("returns provider with versions sorted and platforms preloaded", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		_, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4", "3.1.0"))
		require.NoError(t, err)

		p, err := repo.Find("registry.terraform.io", "hashicorp", "null")
		require.NoError(t, err)

		require.Len(t, p.Versions, 2)
		assert.Equal(t, "3.1.0", p.Versions[0].Version)
		assert.Equal(t, "3.2.4", p.Versions[1].Version)
		require.Len(t, p.Versions[0].Platforms, 1)
		assert.Equal(t, "linux", p.Versions[0].Platforms[0].System)
		assert.Equal(t, "h1:abc", p.Versions[0].Platforms[0].Hashes)
	})

	t.Run("matches hostname, namespace and name case-insensitively", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		_, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"))
		require.NoError(t, err)

		p, err := repo.Find("Registry.Terraform.IO", "HashiCorp", "Null")
		require.NoError(t, err)
		assert.Equal(t, "null", p.Name)
	})
}

func TestMirrorRepository_Upsert(t *testing.T) {
	t.Run("adds a version to an existing provider", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		created, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"))
		require.NoError(t, err)

		created.Versions = append(created.Versions, newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.5").Versions[0])

		_, err = repo.Upsert(*created)
		require.NoError(t, err)

		p, err := repo.Find("registry.terraform.io", "hashicorp", "null")
		require.NoError(t, err)
		assert.Equal(t, created.ID, p.ID)
		require.Len(t, p.Versions, 2)
	})

	t.Run("adds a platform to an existing version", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		created, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"))
		require.NoError(t, err)

		created.Versions[0].Platforms = append(created.Versions[0].Platforms, mirror.Platform{
			System:       "darwin",
			Architecture: "arm64",
			Location:     "mirror/registry.terraform.io/hashicorp/null/3.2.4/darwin_arm64.zip",
			Hashes:       "h1:def",
		})

		_, err = repo.Upsert(*created)
		require.NoError(t, err)

		p, err := repo.Find("registry.terraform.io", "hashicorp", "null")
		require.NoError(t, err)
		require.Len(t, p.Versions, 1)
		assert.Len(t, p.Versions[0].Platforms, 2)
	})

	t.Run("keeps providers with the same namespace and name on different hostnames apart", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		_, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"))
		require.NoError(t, err)
		_, err = repo.Upsert(newMirrorProvider("registry.opentofu.org", "hashicorp", "null", "3.2.4"))
		require.NoError(t, err)

		p, err := repo.Find("registry.opentofu.org", "hashicorp", "null")
		require.NoError(t, err)
		assert.Equal(t, "registry.opentofu.org", p.Hostname)
	})
}

func TestMirrorRepository_FindByNamespace(t *testing.T) {
	repo := newTestMirrorRepository(t)

	for _, p := range []mirror.Provider{
		newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"),
		newMirrorProvider("registry.terraform.io", "hashicorp", "random", "3.6.0"),
		newMirrorProvider("registry.terraform.io", "integrations", "github", "6.0.0"),
		newMirrorProvider("registry.opentofu.org", "hashicorp", "null", "3.2.4"),
	} {
		_, err := repo.Upsert(p)
		require.NoError(t, err)
	}

	t.Run("returns only providers under the namespace", func(t *testing.T) {
		providers, err := repo.FindByNamespace("registry.terraform.io", "hashicorp")
		require.NoError(t, err)
		require.Len(t, providers, 2)
		require.Len(t, providers[0].Versions, 1)
		assert.Len(t, providers[0].Versions[0].Platforms, 1)
	})

	t.Run("returns empty list for an unknown namespace", func(t *testing.T) {
		providers, err := repo.FindByNamespace("registry.terraform.io", "unknown")
		require.NoError(t, err)
		assert.Empty(t, providers)
	})
}

func TestMirrorRepository_FindByHostname(t *testing.T) {
	repo := newTestMirrorRepository(t)

	for _, p := range []mirror.Provider{
		newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"),
		newMirrorProvider("registry.terraform.io", "integrations", "github", "6.0.0"),
		newMirrorProvider("registry.opentofu.org", "hashicorp", "null", "3.2.4"),
	} {
		_, err := repo.Upsert(p)
		require.NoError(t, err)
	}

	t.Run("returns only providers under the hostname", func(t *testing.T) {
		providers, err := repo.FindByHostname("registry.terraform.io")
		require.NoError(t, err)
		assert.Len(t, providers, 2)
	})

	t.Run("returns empty list for an unknown hostname", func(t *testing.T) {
		providers, err := repo.FindByHostname("example.com")
		require.NoError(t, err)
		assert.Empty(t, providers)
	})
}

func TestMirrorRepository_Delete(t *testing.T) {
	repo := newTestMirrorRepository(t)

	created, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4", "3.2.5"))
	require.NoError(t, err)

	require.NoError(t, repo.Delete(created))

	_, err = repo.Find("registry.terraform.io", "hashicorp", "null")
	assert.Error(t, err)

	var versions int64
	require.NoError(t, repo.Database.Handler().Model(&mirror.Version{}).Count(&versions).Error)
	assert.Zero(t, versions)

	var platforms int64
	require.NoError(t, repo.Database.Handler().Model(&mirror.Platform{}).Count(&platforms).Error)
	assert.Zero(t, platforms)
}

func TestMirrorRepository_DeleteVersion(t *testing.T) {
	t.Run("removes the version and its platforms only", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		created, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4", "3.2.5"))
		require.NoError(t, err)

		require.NoError(t, repo.DeleteVersion(created, "3.2.4"))

		p, err := repo.Find("registry.terraform.io", "hashicorp", "null")
		require.NoError(t, err)
		require.Len(t, p.Versions, 1)
		assert.Equal(t, "3.2.5", p.Versions[0].Version)

		var platforms int64
		require.NoError(t, repo.Database.Handler().Model(&mirror.Platform{}).Count(&platforms).Error)
		assert.EqualValues(t, 1, platforms)
	})

	t.Run("returns error when version does not exist", func(t *testing.T) {
		repo := newTestMirrorRepository(t)

		created, err := repo.Upsert(newMirrorProvider("registry.terraform.io", "hashicorp", "null", "3.2.4"))
		require.NoError(t, err)

		assert.Error(t, repo.DeleteVersion(created, "9.9.9"))
	})
}
