package railway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter"
)

func TestGraphQLValidateAndGetService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("auth=%q", r.Header.Get("Authorization"))
		}
		var req gqlRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(req.Query, "me {"):
			_, _ = w.Write([]byte(`{"data":{"me":{"id":"u1","name":"T","email":"t@example.com"}}}`))
		case strings.Contains(req.Query, "service(id"):
			_, _ = w.Write([]byte(`{"data":{"service":{"id":"svc1","name":"api","projectId":"p1"}}}`))
		case strings.Contains(req.Query, "serviceInstance"):
			_, _ = w.Write([]byte(`{"data":{"serviceInstance":{"latestDeployment":{"id":"d1","status":"SUCCESS"}}}}`))
		case strings.Contains(req.Query, "serviceCreate"):
			_, _ = w.Write([]byte(`{"data":{"serviceCreate":{"id":"svc-new","name":"shadow"}}}`))
		default:
			_, _ = w.Write([]byte(`{"data":{}}`))
		}
	}))
	defer srv.Close()

	c := NewClient("test-token", srv.URL)
	c.HTTP = srv.Client()
	if err := c.ValidateToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	id, name, err := c.GetService(context.Background(), "svc1")
	if err != nil || id != "svc1" || name != "api" {
		t.Fatalf("id=%s name=%s err=%v", id, name, err)
	}
	st, ok, err := c.GetServiceInstanceHealth(context.Background(), "svc1", "env1")
	if err != nil || !ok || st != "SUCCESS" {
		t.Fatalf("st=%s ok=%v err=%v", st, ok, err)
	}
	nid, err := c.CreateService(context.Background(), "p1", "shadow", "nginx:latest")
	if err != nil || nid != "svc-new" {
		t.Fatalf("nid=%s err=%v", nid, err)
	}
}

func TestRequireLiveDisabled(t *testing.T) {
	t.Setenv("SDE_RAILWAY_LIVE", "")
	a := NewWithConfig(Config{Token: "x", ProjectID: "p"})
	if err := a.requireLive(); err != ErrNotWired {
		t.Fatalf("err=%v", err)
	}
}

func TestLiveCreateCandidateMock(t *testing.T) {
	t.Setenv("SDE_RAILWAY_LIVE", "1")
	t.Setenv("RAILWAY_STATE_STAGING", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(req.Query, "serviceCreate") {
			_, _ = w.Write([]byte(`{"data":{"serviceCreate":{"id":"shadow-1","name":"s"}}}`))
			return
		}
		if strings.Contains(req.Query, "volumeCreate") {
			_, _ = w.Write([]byte(`{"data":{"volumeCreate":{"id":"vol-1"}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	a := NewWithConfig(Config{Token: "tok", ProjectID: "proj", EnvironmentID: "env", GraphQLURL: srv.URL})
	a.client.HTTP = srv.Client()
	slot, err := a.CreateCandidate(context.Background(), adapter.CreateShadowRequest{ImageRef: "img:v2"})
	if err != nil {
		t.Fatal(err)
	}
	if slot.ID != "shadow-1" {
		t.Fatalf("%+v", slot)
	}
}
