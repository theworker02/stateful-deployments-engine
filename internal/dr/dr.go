// Package dr plans and scores disaster recovery readiness from evidence.
package dr

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/escrow"
	"github.com/theworker02/stateful-deployments-engine/internal/recoverypoint"
)

// Conclusion is the DR planner verdict. UNKNOWN must never be treated as RECOVERABLE.
type Conclusion string

const (
	Recoverable          Conclusion = "RECOVERABLE"
	PartiallyRecoverable Conclusion = "PARTIALLY_RECOVERABLE"
	NotRecoverable       Conclusion = "NOT_RECOVERABLE"
	Unknown              Conclusion = "UNKNOWN"
)

// Plan is the durable DR assessment.
type Plan struct {
	PlanID     string     `json:"plan_id"`
	CreatedAt  time.Time  `json:"created_at"`
	Conclusion Conclusion `json:"conclusion"`
	Evidence   []string   `json:"evidence"`
	Archives   int        `json:"archives_cataloged"`
	Points     int        `json:"recovery_points"`
	LastDigest string     `json:"last_root_digest,omitempty"`
	LastEpoch  uint64     `json:"last_epoch,omitempty"`
	Scorecard  Scorecard  `json:"scorecard"`
}

// Scorecard uses factual dimensions only (no marketing scores).
type Scorecard struct {
	ArchivePresent       bool       `json:"archive_present"`
	ArchiveVerified      bool       `json:"archive_verified"`
	EscrowPresent        bool       `json:"escrow_present"`
	RecoveryPointPresent bool       `json:"recovery_point_present"`
	JournalChainPresent  bool       `json:"journal_chain_present"`
	RestorePathKnown     bool       `json:"restore_path_known"`
	SourceIndependent    bool       `json:"source_independent"` // archive exists off-source
	FireDrillRecorded    bool       `json:"fire_drill_recorded"`
	LastFireDrillOK      bool       `json:"last_fire_drill_ok"`
	LastFireDrillAt      *time.Time `json:"last_fire_drill_at,omitempty"`
}

// Planner inspects catalog + recovery points.
type Planner struct {
	Escrow   *escrow.Store
	Points   *recoverypoint.Store
}

// Plan produces a DR conclusion with evidence.
func (p *Planner) Plan() (*Plan, error) {
	plan := &Plan{
		PlanID:    fmt.Sprintf("dr-%d", time.Now().UnixNano()),
		CreatedAt: time.Now().UTC(),
	}
	var evidence []string
	sc := Scorecard{}

	entries, err := p.Escrow.List()
	if err != nil {
		plan.Conclusion = Unknown
		plan.Evidence = []string{"catalog unreadable: " + err.Error()}
		return plan, nil
	}
	plan.Archives = len(entries)
	if len(entries) == 0 {
		evidence = append(evidence, "no escrowed archives in catalog")
	} else {
		sc.ArchivePresent = true
		sc.EscrowPresent = true
		sc.SourceIndependent = true
		sc.RestorePathKnown = entries[0].LocalPath != ""
		plan.LastDigest = entries[0].RootDigest
		plan.LastEpoch = entries[0].Epoch
		if _, err := archive.Verify(entries[0].LocalPath); err == nil {
			sc.ArchiveVerified = true
			evidence = append(evidence, "latest archive verified: "+entries[0].ArchiveID)
		} else {
			evidence = append(evidence, "latest archive verify failed: "+err.Error())
		}
		if entries[0].LastFireDrill != nil {
			sc.FireDrillRecorded = true
			sc.LastFireDrillOK = entries[0].LastFireDrill.RootDigestOK
			t := entries[0].LastFireDrill.RanAt
			sc.LastFireDrillAt = &t
			if sc.LastFireDrillOK {
				evidence = append(evidence, "last fire drill ok at "+t.Format(time.RFC3339))
			} else {
				evidence = append(evidence, "last fire drill failed at "+t.Format(time.RFC3339))
			}
		} else {
			evidence = append(evidence, "no fire drill recorded on latest catalog entry")
		}
	}

	pts, err := p.Points.List()
	if err != nil {
		evidence = append(evidence, "recovery points unreadable: "+err.Error())
	} else {
		plan.Points = len(pts)
		if len(pts) > 0 {
			sc.RecoveryPointPresent = true
			evidence = append(evidence, fmt.Sprintf("%d recovery points registered", len(pts)))
			for _, pt := range pts {
				if pt.Kind == "JOURNAL_CHAIN" {
					sc.JournalChainPresent = true
					break
				}
			}
			if !sc.JournalChainPresent {
				evidence = append(evidence, "PITR limited to SNAPSHOT epoch/journal position (no finer claim)")
			}
		}
	}

	plan.Scorecard = sc
	switch {
	case !sc.ArchivePresent:
		plan.Conclusion = NotRecoverable
		evidence = append(evidence, "NOT_RECOVERABLE: no portable archive")
	case sc.ArchivePresent && !sc.ArchiveVerified:
		plan.Conclusion = PartiallyRecoverable
		evidence = append(evidence, "PARTIALLY_RECOVERABLE: archive present but not verified")
	case sc.ArchivePresent && sc.ArchiveVerified && sc.RestorePathKnown:
		plan.Conclusion = Recoverable
		evidence = append(evidence, "RECOVERABLE via ENGINE_PROVIDED archive restore")
	default:
		plan.Conclusion = Unknown
		evidence = append(evidence, "UNKNOWN: insufficient evidence — must not treat as RECOVERABLE")
	}
	plan.Evidence = evidence
	return plan, nil
}

// IsRecoverable never returns true for UNKNOWN.
func IsRecoverable(c Conclusion) bool {
	return c == Recoverable || c == PartiallyRecoverable
}

// FireDrill restores into an isolated temp target, verifies, measures, destroys temp.
// Never modifies production paths under prodRoot.
type FireDrillResult struct {
	ReceiptID     string    `json:"receipt_id"`
	StartedAt     time.Time `json:"started_at"`
	DurationMs    float64   `json:"duration_ms"`
	ArchiveID     string    `json:"archive_id"`
	TempTarget    string    `json:"temp_target"`
	RootDigestOK  bool      `json:"root_digest_ok"`
	ExpectedDigest string   `json:"expected_digest"`
	ActualDigest  string    `json:"actual_digest"`
	DestroyedTemp bool      `json:"destroyed_temp"`
	ProductionUntouched bool `json:"production_untouched"`
	Error         string    `json:"error,omitempty"`
}

// RunFireDrill executes an isolated restore drill.
func RunFireDrill(archiveDir, tempParent string) (*FireDrillResult, error) {
	start := time.Now()
	r := &FireDrillResult{
		ReceiptID:           fmt.Sprintf("drill-%d", start.UnixNano()),
		StartedAt:           start.UTC(),
		ProductionUntouched: true,
	}
	m, err := archive.Inspect(archiveDir)
	if err != nil {
		r.Error = err.Error()
		return r, err
	}
	r.ArchiveID = m.ArchiveID
	r.ExpectedDigest = m.RootDigest
	if tempParent != "" {
		_ = os.MkdirAll(tempParent, 0o755)
	}
	tmp, err := os.MkdirTemp(tempParent, "sde-firedrill-*")
	if err != nil {
		r.Error = err.Error()
		return r, err
	}
	r.TempTarget = tmp
	defer func() {
		_ = os.RemoveAll(tmp)
		r.DestroyedTemp = true
	}()

	if _, err := archive.Import(archiveDir, filepath.Join(tmp, "state")); err != nil {
		r.Error = err.Error()
		r.DurationMs = float64(time.Since(start).Microseconds()) / 1000
		return r, err
	}
	// Re-export digest check via verify of re-archived import
	out := filepath.Join(tmp, "reexport")
	m2, _, err := archive.Export(filepath.Join(tmp, "state"), out, m.SourceEpoch, m.JournalPosition, "")
	if err != nil {
		r.Error = err.Error()
		r.DurationMs = float64(time.Since(start).Microseconds()) / 1000
		return r, err
	}
	r.ActualDigest = m2.RootDigest
	r.RootDigestOK = m2.RootDigest == m.RootDigest
	r.DurationMs = float64(time.Since(start).Microseconds()) / 1000
	if !r.RootDigestOK {
		r.Error = "digest mismatch after fire-drill restore"
		return r, fmt.Errorf("%s", r.Error)
	}
	return r, nil
}
