package config

import (
	"strings"
	"testing"
)

func TestNormalizeDSN(t *testing.T) {
	raw := "mysql://4ER1REbBspwxNjK.root:i2Ny6gcH4uuXF1GP@gateway01.us-east-1.prod.aws.tidbcloud.com:4000/workshop"
	normalized := normalizeDSN(raw)

	expectedPrefix := "4ER1REbBspwxNjK.root:i2Ny6gcH4uuXF1GP@tcp(gateway01.us-east-1.prod.aws.tidbcloud.com:4000)/workshop?"
	if !strings.HasPrefix(normalized, expectedPrefix) {
		t.Fatalf("Expected DSN to start with %s, got: %s", expectedPrefix, normalized)
	}
	if !strings.Contains(normalized, "parseTime=true") {
		t.Fatalf("Expected parseTime=true in DSN, got: %s", normalized)
	}
	if !strings.Contains(normalized, "tls=true") {
		t.Fatalf("Expected tls=true in DSN, got: %s", normalized)
	}
}
