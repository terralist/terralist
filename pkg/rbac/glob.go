package rbac

import (
	"strings"

	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
)

// globMatch is a custom function for Casbin to support glob pattern matching.
// Values match regardless of case, as the artifacts they name are looked up.
func globMatch(args ...any) (any, error) {
	if len(args) < 2 {
		return false, nil
	}

	val, ok := args[0].(string)
	if !ok {
		return false, nil
	}

	pattern, ok := args[1].(string)
	if !ok {
		return false, nil
	}

	compiledGlob, err := glob.Compile(strings.ToLower(pattern))
	if err != nil {
		log.Warn().Err(err).Str("pattern", pattern).Msg("failed to compile glob pattern")
		return false, nil
	}

	return compiledGlob.Match(strings.ToLower(val)), nil
}
