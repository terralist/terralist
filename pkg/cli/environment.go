package cli

import (
	"os"
	"regexp"
)

var (
	environmentRegEx = regexp.MustCompile(`(?m)[$][{]([a-zA-Z_]+)(?::(.*))?[}]`)
)

func environmentLookup(value string) (string, bool) {
	matches := environmentRegEx.FindAllStringSubmatch(value, -1)

	if len(matches) == 0 {
		return "", false
	}

	groups := matches[0][1:]
	if len(groups) == 0 {
		return "", false
	}

	envName := groups[0]
	envDefault := groups[1]

	envValue, ok := os.LookupEnv(envName)
	if !ok {
		return envDefault, true
	}

	return envValue, true
}
