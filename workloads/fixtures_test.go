package workloads_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theworker02/stateful-deployments-engine/workloads"
)

func TestMaterializeAllFixtures(t *testing.T) {
	for _, k := range workloads.AllFixtureKinds() {
		dir := t.TempDir()
		meta, err := workloads.MaterializeFixture(dir, k)
		if err != nil {
			t.Fatalf("%s: %v", k, err)
		}
		if meta.Files <= 0 {
			t.Fatalf("%s: no files", k)
		}
		if _, err := os.Stat(filepath.Join(dir, "FIXTURE.json")); err != nil {
			t.Fatal(err)
		}
	}
}
