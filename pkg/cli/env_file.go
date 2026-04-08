package cli

import (
	"fmt"
	"os"
)

// LoadEnvFiles copies the contents of <ENV>_FILE into <ENV>.
func LoadEnvFiles(envNames ...string) error {
	for _, envName := range envNames {
		fileEnvName := envName + "_FILE"

		filePath, ok := os.LookupEnv(fileEnvName)
		if !ok {
			continue
		}

		if _, ok := os.LookupEnv(envName); ok {
			return fmt.Errorf("%s and %s cannot both be set", envName, fileEnvName)
		}

		value, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", fileEnvName, err)
		}

		if err := os.Setenv(envName, string(value)); err != nil {
			return fmt.Errorf("set %s from %s: %w", envName, fileEnvName, err)
		}
	}

	return nil
}
