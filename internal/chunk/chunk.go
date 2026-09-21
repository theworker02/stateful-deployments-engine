// Package chunk implements optional fixed-size block differential transfer.
package chunk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DefaultBlockSize = 64 * 1024

// Fingerprint is a content hash of one block.
type Fingerprint struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	Length int    `json:"length"`
	Hash   string `json:"hash"`
}

// Index maps path:offset -> hash for a tree.
type Index map[string]Fingerprint

func key(path string, off int64) string { return fmt.Sprintf("%s@%d", path, off) }

// BuildIndex fingerprints all files under root with fixed block size.
func BuildIndex(root string, blockSize int) (Index, int64, error) {
	if blockSize <= 0 {
		blockSize = DefaultBlockSize
	}
	idx := make(Index)
	var total int64
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, ".sde/") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		buf := make([]byte, blockSize)
		var off int64
		for {
			n, rerr := io.ReadFull(f, buf)
			if n > 0 {
				sum := sha256.Sum256(buf[:n])
				fp := Fingerprint{Path: rel, Offset: off, Length: n, Hash: hex.EncodeToString(sum[:])}
				idx[key(rel, off)] = fp
				total += int64(n)
				off += int64(n)
			}
			if rerr == io.EOF || rerr == io.ErrUnexpectedEOF {
				break
			}
			if rerr != nil {
				return rerr
			}
		}
		return nil
	})
	return idx, total, err
}

// Diff reports blocks present in src but missing/different in dst.
func Diff(src, dst Index) []Fingerprint {
	var out []Fingerprint
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s := src[k]
		d, ok := dst[k]
		if !ok || d.Hash != s.Hash {
			out = append(out, s)
		}
	}
	return out
}

// TransferChanged copies only differing blocks from srcRoot to dstRoot.
// Open file handles are cached per path to avoid reopening on every block.
func TransferChanged(srcRoot, dstRoot string, changed []Fingerprint) (bytesCopied int64, err error) {
	type pair struct {
		in  *os.File
		out *os.File
	}
	open := map[string]*pair{}
	defer func() {
		for _, p := range open {
			if p.in != nil {
				_ = p.in.Close()
			}
			if p.out != nil {
				_ = p.out.Close()
			}
		}
	}()

	get := func(rel string) (*pair, error) {
		if p, ok := open[rel]; ok {
			return p, nil
		}
		src := filepath.Join(srcRoot, filepath.FromSlash(rel))
		dst := filepath.Join(dstRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, err
		}
		in, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			_ = in.Close()
			return nil, err
		}
		p := &pair{in: in, out: out}
		open[rel] = p
		return p, nil
	}

	for _, fp := range changed {
		p, err := get(fp.Path)
		if err != nil {
			return bytesCopied, err
		}
		buf := make([]byte, fp.Length)
		if _, err := p.in.ReadAt(buf, fp.Offset); err != nil && err != io.EOF {
			return bytesCopied, err
		}
		if _, err := p.out.WriteAt(buf, fp.Offset); err != nil {
			return bytesCopied, err
		}
		bytesCopied += int64(fp.Length)
	}
	return bytesCopied, nil
}

// FullCopy copies every file (benchmark baseline).
func FullCopy(srcRoot, dstRoot string) (int64, error) {
	var total int64
	err := filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, info.Mode()); err != nil {
			return err
		}
		total += int64(len(b))
		return nil
	})
	return total, err
}
