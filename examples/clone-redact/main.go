// Clone + redaction demo: export → clone archive → import → apply text hooks.
// Demonstrates ENGINE_PROVIDED safe cloning with honest (non-binary) redaction.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/theworker02/stateful-deployments-engine/internal/archive"
	"github.com/theworker02/stateful-deployments-engine/internal/sensitive"
)

func main() {
	base := filepath.Join(os.TempDir(), "sde-clone-redact")
	_ = os.RemoveAll(base)
	src := filepath.Join(base, "src")
	_ = os.MkdirAll(filepath.Join(src, "data"), 0o755)
	_ = os.WriteFile(filepath.Join(src, "data", "user.json"),
		[]byte(`{"email":"ops@example.com","api_key":"sk-live-secret","note":"ok"}`), 0o644)
	_ = os.WriteFile(filepath.Join(src, "data", "blob.bin"), []byte{0x00, 0x01, 0xff}, 0o644)

	arch := filepath.Join(base, "archive")
	m, _, err := archive.Export(src, arch, 1, 1, "")
	must(err)
	fmt.Printf("exported %s\n", m.ArchiveID)

	cloneArch := filepath.Join(base, "clone-arch")
	cm, err := archive.CloneReadOnly(arch, cloneArch)
	must(err)
	fmt.Printf("cloned archive %s (original %s untouched)\n", cm.ArchiveID, m.ArchiveID)

	stateOut := filepath.Join(base, "clone-state")
	_, err = archive.Import(cloneArch, stateOut)
	must(err)

	pipe := sensitive.Pipeline{Hooks: []sensitive.Hook{sensitive.RedactEmails{}, sensitive.TokenizeSecrets{}}}
	n, err := archive.ApplyHooksToState(stateOut, pipe.Apply)
	must(err)
	fmt.Printf("redacted %d text file(s)\n", n)

	out, _ := os.ReadFile(filepath.Join(stateOut, "data", "user.json"))
	fmt.Printf("user.json => %s\n", string(out))
	bin, _ := os.ReadFile(filepath.Join(stateOut, "data", "blob.bin"))
	fmt.Printf("blob.bin unchanged len=%d (no fake binary anonymization)\n", len(bin))
	fmt.Println("CLONE+REDACT DEMO COMPLETE")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
