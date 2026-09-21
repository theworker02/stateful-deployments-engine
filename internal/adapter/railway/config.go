package railway

import (
	"fmt"
	"os"
)

// Config holds Railway project targeting.
type Config struct {
	Token           string
	ProjectID       string
	EnvironmentID   string
	ActiveServiceID string
	ShadowServiceID string
	PublicDomain    string
	UseProjectToken bool
	GraphQLURL      string
}

// NewFromEnv builds config from environment variables.
// Token resolution: RAILWAY_TOKEN, else RAILWAY_API_TOKEN, else RAILWAY_PROJECT_TOKEN.
func NewFromEnv() (*Adapter, error) {
	token := os.Getenv("RAILWAY_TOKEN")
	useProject := false
	if token == "" {
		token = os.Getenv("RAILWAY_API_TOKEN")
	}
	if token == "" {
		token = os.Getenv("RAILWAY_PROJECT_TOKEN")
		useProject = token != ""
	}
	cfg := Config{
		Token:           token,
		ProjectID:       os.Getenv("RAILWAY_PROJECT_ID"),
		EnvironmentID:   os.Getenv("RAILWAY_ENVIRONMENT_ID"),
		ActiveServiceID: os.Getenv("RAILWAY_SERVICE_ID"),
		ShadowServiceID: os.Getenv("RAILWAY_SHADOW_SERVICE_ID"),
		PublicDomain:    os.Getenv("RAILWAY_PUBLIC_DOMAIN"),
		UseProjectToken: useProject || os.Getenv("RAILWAY_USE_PROJECT_TOKEN") == "1",
		GraphQLURL:      os.Getenv("RAILWAY_GRAPHQL_URL"),
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("railway: RAILWAY_TOKEN, RAILWAY_API_TOKEN, or RAILWAY_PROJECT_TOKEN required")
	}
	if cfg.ProjectID == "" && !cfg.UseProjectToken {
		return nil, fmt.Errorf("railway: RAILWAY_PROJECT_ID required (unless using project token)")
	}
	client := NewClient(cfg.Token, cfg.GraphQLURL)
	client.ProjectToken = cfg.UseProjectToken
	return &Adapter{cfg: cfg, client: client}, nil
}

// NewWithConfig constructs an adapter from explicit config (tests).
func NewWithConfig(cfg Config) *Adapter {
	client := NewClient(cfg.Token, cfg.GraphQLURL)
	client.ProjectToken = cfg.UseProjectToken
	return &Adapter{cfg: cfg, client: client}
}
