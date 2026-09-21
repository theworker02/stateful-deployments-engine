package fence

import (
	"testing"
)

func TestIssueAdvanceCheck(t *testing.T) {
	root := t.TempDir()
	m, err := Open(root, "coord")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := m.Issue("coord")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Check(tok.Epoch, tok.Token); err != nil {
		t.Fatal(err)
	}
	if err := m.Check(tok.Epoch, tok.Token-1); err == nil {
		t.Fatal("expected stale token")
	}
	adv, err := m.AdvanceEpoch("coord-2")
	if err != nil {
		t.Fatal(err)
	}
	if adv.Epoch <= tok.Epoch {
		t.Fatalf("%+v vs %+v", adv, tok)
	}
	m2, err := Open(root, "x")
	if err != nil {
		t.Fatal(err)
	}
	if m2.Current().Epoch != adv.Epoch {
		t.Fatalf("%+v", m2.Current())
	}
}
