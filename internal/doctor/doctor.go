// Package doctor reports local install and environment readiness.
package doctor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/adapter/railway"
	"github.com/theworker02/stateful-deployments-engine/internal/version"
)

// Severity classifies a check outcome.
type Severity string

const (
	OK      Severity = "ok"
	Warn    Severity = "warn"
	Fail    Severity = "fail"
	Info    Severity = "info"
)

// Check is one readiness row.
type Check struct {
	Name     string   `json:"name"`
	Severity Severity `json:"severity"`
	Detail   string   `json:"detail"`
}

// Report is the full doctor output.
type Report struct {
	Version   string    `json:"version"`
	Codename  string    `json:"codename"`
	CheckedAt time.Time `json:"checked_at"`
	GOOS      string    `json:"goos"`
	GOARCH    string    `json:"goarch"`
	Checks    []Check   `json:"checks"`
	Ready     bool      `json:"ready"`
}

// Options control optional live probes.
type Options struct {
	Root       string
	ProbeLive  bool // if SDE_RAILWAY_LIVE=1, call Railway ValidateToken
	Timeout    time.Duration
}

// Run executes local readiness checks.
func Run(ctx context.Context, opt Options) Report {
	if opt.Timeout <= 0 {
		opt.Timeout = 8 * time.Second
	}
	if opt.Root == "" {
		opt.Root = ".sde"
	}
	r := Report{
		Version:   version.Version,
		Codename:  version.Codename,
		CheckedAt: time.Now().UTC(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		Ready:     true,
	}

	r.add(Check{Name: "engine", Severity: OK, Detail: version.String() + " (" + version.Codename + ")"})
	r.add(Check{Name: "runtime", Severity: OK, Detail: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)})

	if st, err := os.Stat(opt.Root); err != nil {
		r.add(Check{
			Name:     "workspace",
			Severity: Warn,
			Detail:   fmt.Sprintf("%s not initialized — run: sde init --root %s", opt.Root, opt.Root),
		})
	} else if !st.IsDir() {
		r.add(Check{Name: "workspace", Severity: Fail, Detail: opt.Root + " exists but is not a directory"})
		r.Ready = false
	} else {
		r.add(Check{Name: "workspace", Severity: OK, Detail: absOr(opt.Root)})
		for _, sub := range []string{"journal", "slots", "deployments"} {
			p := filepath.Join(opt.Root, sub)
			if _, err := os.Stat(p); err != nil {
				r.add(Check{Name: "workspace." + sub, Severity: Info, Detail: "missing (created on first use)"})
			}
		}
	}

	live := os.Getenv("SDE_RAILWAY_LIVE") == "1"
	token := firstNonEmpty(
		os.Getenv("RAILWAY_TOKEN"),
		os.Getenv("RAILWAY_API_TOKEN"),
		os.Getenv("RAILWAY_PROJECT_TOKEN"),
	)
	project := os.Getenv("RAILWAY_PROJECT_ID")
	envID := os.Getenv("RAILWAY_ENVIRONMENT_ID")
	svc := os.Getenv("RAILWAY_SERVICE_ID")

	if !live {
		r.add(Check{
			Name:     "railway.live",
			Severity: Info,
			Detail:   "offline / ENGINE_PROVIDED mode (set SDE_RAILWAY_LIVE=1 for GraphQL)",
		})
	} else {
		r.add(Check{Name: "railway.live", Severity: OK, Detail: "SDE_RAILWAY_LIVE=1"})
		if token == "" {
			r.add(Check{Name: "railway.token", Severity: Fail, Detail: "RAILWAY_TOKEN (or API/PROJECT token) required"})
			r.Ready = false
		} else {
			r.add(Check{Name: "railway.token", Severity: OK, Detail: "present (" + redactTail(token) + ")"})
		}
		if project == "" && os.Getenv("RAILWAY_PROJECT_TOKEN") == "" {
			r.add(Check{Name: "railway.project", Severity: Warn, Detail: "RAILWAY_PROJECT_ID empty"})
		} else if project != "" {
			r.add(Check{Name: "railway.project", Severity: OK, Detail: project})
		}
		if envID == "" {
			r.add(Check{Name: "railway.environment", Severity: Warn, Detail: "RAILWAY_ENVIRONMENT_ID empty (needed for health/deploy)"})
		} else {
			r.add(Check{Name: "railway.environment", Severity: OK, Detail: envID})
		}
		if svc == "" {
			r.add(Check{Name: "railway.service", Severity: Info, Detail: "RAILWAY_SERVICE_ID unset (optional until cutover)"})
		} else {
			r.add(Check{Name: "railway.service", Severity: OK, Detail: svc})
		}

		if opt.ProbeLive && token != "" {
			pctx, cancel := context.WithTimeout(ctx, opt.Timeout)
			defer cancel()
			a, err := railway.NewFromEnv()
			if err != nil {
				r.add(Check{Name: "railway.api", Severity: Fail, Detail: err.Error()})
				r.Ready = false
			} else if err := a.ValidateAPI(pctx); err != nil {
				r.add(Check{Name: "railway.api", Severity: Fail, Detail: err.Error()})
				r.Ready = false
			} else {
				r.add(Check{Name: "railway.api", Severity: OK, Detail: "GraphQL token validated"})
			}
		}
	}

	if _, err := os.Stat(".env.example"); err == nil {
		r.add(Check{Name: "docs.env", Severity: OK, Detail: ".env.example present — copy to .env for live mode"})
	}

	return r
}

func (r *Report) add(c Check) { r.Checks = append(r.Checks, c) }

// FormatHuman renders a terminal-friendly report.
func FormatHuman(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "SDE doctor — %s (%s)\n", r.Version, r.Codename)
	fmt.Fprintf(&b, "checked %s · %s/%s\n\n", r.CheckedAt.Format(time.RFC3339), r.GOOS, r.GOARCH)
	for _, c := range r.Checks {
		mark := "·"
		switch c.Severity {
		case OK:
			mark = "✓"
		case Warn:
			mark = "!"
		case Fail:
			mark = "✗"
		case Info:
			mark = "i"
		}
		fmt.Fprintf(&b, "  %s %-22s %s\n", mark, c.Name, c.Detail)
	}
	fmt.Fprintln(&b)
	if r.Ready {
		fmt.Fprintln(&b, "ready: yes — next: sde demo --root .sde   or   .\\evaluate.ps1")
	} else {
		fmt.Fprintln(&b, "ready: no — fix ✗ items above, then re-run: sde doctor")
	}
	return b.String()
}

func absOr(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func redactTail(s string) string {
	if len(s) <= 6 {
		return "***"
	}
	return "***" + s[len(s)-4:]
}
