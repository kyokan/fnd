package testflags

import (
	"os"
	"testing"
)

func IntegrationTest(t *testing.T) {
	_, ok := os.LookupEnv("Footnote_ENABLE_INTEGRATION_TESTS")
	if !ok {
		t.SkipNow()
	}
	t.Parallel()
}

func HandshakeTest(t *testing.T) {
	_, ok := os.LookupEnv("Footnote_ENABLE_HANDSHAKE_TESTS")
	if !ok {
		t.SkipNow()
	}
	t.Parallel()
}
