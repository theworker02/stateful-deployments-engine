package sensitive_test

import (
	"strings"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/sensitive"
)

func TestPipelineRedactsTextLeavesBinary(t *testing.T) {
	pipe := sensitive.Pipeline{Hooks: []sensitive.Hook{sensitive.RedactEmails{}, sensitive.TokenizeSecrets{}}}
	out, err := pipe.Apply("user.json", []byte(`{"email":"a@b.co","api_key":"sk-secret"}`))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "a@b.co") || strings.Contains(s, "sk-secret") {
		t.Fatalf("expected redaction, got %s", s)
	}
	if !strings.Contains(s, "[REDACTED_EMAIL]") || !strings.Contains(s, "[TOKENIZED]") {
		t.Fatalf("missing tokens: %s", s)
	}
	bin := []byte{0, 1, 2, 3}
	bout, err := pipe.Apply("blob.bin", bin)
	if err != nil {
		t.Fatal(err)
	}
	if string(bout) != string(bin) {
		t.Fatal("binary must be unchanged")
	}
}
