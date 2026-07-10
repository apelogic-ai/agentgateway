package curl

import (
	"strings"
	"testing"
)

func TestExecuteRequestRejectsMetadataServiceTargets(t *testing.T) {
	_, err := ExecuteRequest(WithHost("169.254.169.254"), WithPath("latest/meta-data"))
	if err == nil {
		t.Fatal("expected metadata service target to be rejected")
	}
	if !strings.Contains(err.Error(), "disallowed metadata service target") {
		t.Fatalf("expected metadata service rejection, got %v", err)
	}
}

func TestExecuteRequestRejectsUnsupportedScheme(t *testing.T) {
	_, err := ExecuteRequest(WithScheme("file"), WithHost("example.com"))
	if err == nil {
		t.Fatal("expected unsupported scheme to be rejected")
	}
	if !strings.Contains(err.Error(), "unsupported request scheme") {
		t.Fatalf("expected unsupported scheme rejection, got %v", err)
	}
}
