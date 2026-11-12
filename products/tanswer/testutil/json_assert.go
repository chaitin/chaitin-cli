package testutil

import (
	"encoding/json"
	"testing"
)

func AssertValidJSON(t *testing.T, data []byte) {
	t.Helper()
	if !json.Valid(data) {
		t.Fatalf("invalid JSON: %s", string(data))
	}
}
