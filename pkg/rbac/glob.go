package rbac

import (
	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
)

// globMatch is a custom function for Casbin to support glob pattern matching.
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

	compiledGlob, err := glob.Compile(pattern)
	if err != nil {
		log.Warn().Err(err).Str("pattern", pattern).Msg("failed to compile glob pattern")
		return false, nil
	}

	return compiledGlob.Match(val), nil
}
