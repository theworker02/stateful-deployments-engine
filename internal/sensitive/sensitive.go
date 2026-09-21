// Package sensitive provides redaction/tokenization/sanitization hooks.
// It does NOT pretend to auto-anonymize arbitrary binary blobs.
package sensitive

import (
	"regexp"
	"strings"
)

// Hook transforms text-ish payloads during export/clone when registered.
type Hook interface {
	Name() string
	Apply(path string, data []byte) ([]byte, error)
}

// RedactEmails replaces email-like tokens in UTF-8 text files.
type RedactEmails struct{}

func (RedactEmails) Name() string { return "redact-emails" }

var emailRE = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

func (RedactEmails) Apply(path string, data []byte) ([]byte, error) {
	if !looksText(path, data) {
		return data, nil // refuse to invent binary anonymization
	}
	return emailRE.ReplaceAll(data, []byte("[REDACTED_EMAIL]")), nil
}

// TokenizeSecrets replaces common key=value secret patterns in text.
type TokenizeSecrets struct{}

func (TokenizeSecrets) Name() string { return "tokenize-secrets" }

var secretRE = regexp.MustCompile(`(?i)("?(?:api[_-]?key|password|secret|token)"?\s*[:=]\s*)("?)[^"\s,}\]]+("?)`)

func (TokenizeSecrets) Apply(path string, data []byte) ([]byte, error) {
	if !looksText(path, data) {
		return data, nil
	}
	return secretRE.ReplaceAll(data, []byte("${1}${2}[TOKENIZED]${3}")), nil
}

// Pipeline applies hooks in order.
type Pipeline struct {
	Hooks []Hook
}

func (p Pipeline) Apply(path string, data []byte) ([]byte, error) {
	var err error
	for _, h := range p.Hooks {
		data, err = h.Apply(path, data)
		if err != nil {
			return data, err
		}
	}
	return data, nil
}

func looksText(path string, data []byte) bool {
	lower := strings.ToLower(path)
	for _, ext := range []string{".txt", ".json", ".yaml", ".yml", ".md", ".env", ".csv", ".log"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	// heuristic: no NUL in first 512
	n := len(data)
	if n > 512 {
		n = 512
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return false
		}
	}
	return true
}
