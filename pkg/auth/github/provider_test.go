package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextPageURL(t *testing.T) {
	t.Run("empty header returns empty string", func(t *testing.T) {
		assert.Equal(t, "", nextPageURL(""))
	})

	t.Run("header with next link returns the URL", func(t *testing.T) {
		header := `<https://api.github.com/orgs/my-org/teams?per_page=100&page=2>; rel="next", <https://api.github.com/orgs/my-org/teams?per_page=100&page=5>; rel="last"`
		assert.Equal(t, "https://api.github.com/orgs/my-org/teams?per_page=100&page=2", nextPageURL(header))
	})

	t.Run("header without next link returns empty string", func(t *testing.T) {
		header := `<https://api.github.com/orgs/my-org/teams?per_page=100&page=4>; rel="prev", <https://api.github.com/orgs/my-org/teams?per_page=100&page=1>; rel="first"`
		assert.Equal(t, "", nextPageURL(header))
	})

	t.Run("header with only next link", func(t *testing.T) {
		header := `<https://api.github.com/orgs/my-org/teams?per_page=100&page=2>; rel="next"`
		assert.Equal(t, "https://api.github.com/orgs/my-org/teams?per_page=100&page=2", nextPageURL(header))
	})
}
