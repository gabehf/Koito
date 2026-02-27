package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestInjectAndFromContext(t *testing.T) {
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Inject logger
	req = Inject(req, &testLogger)

	// Retrieve logger
	l := FromContext(req.Context())

	l.Info().Msg("hello")

	output := buf.String()

	if output == "" {
		t.Fatal("expected log output, got empty string")
	}

	if !bytes.Contains([]byte(output), []byte("hello")) {
		t.Errorf("expected log to contain 'hello', got %s", output)
	}
}
