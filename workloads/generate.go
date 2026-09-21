package workloads

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

// Profile names used by benches and chaos fixtures.
const (
	ProfileSmall        = "SMALL"
	ProfileMedium       = "MEDIUM"
	ProfileDatabaseLike = "DATABASE-LIKE"
	ProfileMediaLike    = "MEDIA-LIKE"
	ProfileLogLike      = "LOG-LIKE"
)

// Spec describes a synthetic state tree.
type Spec struct {
	Name       string
	FileCount  int
	FileSize   int
	HotFraction float64 // fraction of files that receive frequent writes
}

// Specs returns built-in workload profiles.
func Specs() map[string]Spec {
	return map[string]Spec{
		ProfileSmall:        {Name: ProfileSmall, FileCount: 20, FileSize: 4 * 1024, HotFraction: 0.2},
		ProfileMedium:       {Name: ProfileMedium, FileCount: 200, FileSize: 64 * 1024, HotFraction: 0.15},
		ProfileDatabaseLike: {Name: ProfileDatabaseLike, FileCount: 8, FileSize: 1 * 1024 * 1024, HotFraction: 0.5},
		ProfileMediaLike:    {Name: ProfileMediaLike, FileCount: 50, FileSize: 512 * 1024, HotFraction: 0.05},
		ProfileLogLike:      {Name: ProfileLogLike, FileCount: 5, FileSize: 256 * 1024, HotFraction: 1.0},
	}
}

// Generate writes a synthetic tree under root for the named profile.
func Generate(root, profile string) (Spec, error) {
	specs := Specs()
	spec, ok := specs[profile]
	if !ok {
		return Spec{}, fmt.Errorf("unknown workload profile %q", profile)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return Spec{}, err
	}
	buf := make([]byte, spec.FileSize)
	for i := 0; i < spec.FileCount; i++ {
		if _, err := rand.Read(buf); err != nil {
			return Spec{}, err
		}
		dir := "cold"
		if float64(i)/float64(spec.FileCount) < spec.HotFraction {
			dir = "hot"
		}
		path := filepath.Join(root, dir, fmt.Sprintf("obj-%04d.bin", i))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Spec{}, err
		}
		if err := os.WriteFile(path, buf, 0o644); err != nil {
			return Spec{}, err
		}
	}
	return spec, nil
}
