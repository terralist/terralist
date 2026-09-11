package controllers

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TestMain silences the application logger so that the log lines emitted by
// the middlewares on expected failures do not pollute the test output.
func TestMain(m *testing.M) {
	log.Logger = zerolog.Nop()

	os.Exit(m.Run())
}
