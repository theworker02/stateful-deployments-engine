package local

import (
	"path/filepath"
)

// Paths returns conventional directories under the adapter root.
type Paths struct {
	Root     string
	Slots    string
	Journal  string
	MetaJSON string
	Receipts string
	FSM      string
}

// Layout returns the local adapter path layout.
func (a *Adapter) Layout() Paths {
	return Paths{
		Root:     a.root,
		Slots:    filepath.Join(a.root, "slots"),
		Journal:  filepath.Join(a.root, "journal"),
		MetaJSON: filepath.Join(a.root, "meta.json"),
		Receipts: filepath.Join(a.root, "receipts"),
		FSM:      filepath.Join(a.root, "fsm"),
	}
}
