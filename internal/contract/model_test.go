package contract

import (
	"path/filepath"
	"testing"
)

func TestDeclaredContractOwnsSixCaseDenominator(t *testing.T) {
	path := filepath.Join("..", "..", ".gooo", "staged-quasiquote.gooo")
	c, _, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Scenarios) != 6 {
		t.Fatalf("scenario count = %d, want 6", len(c.Scenarios))
	}
	if c.Semantics.Denominator.StatusCounts[string(StatusClosed)] != 3 || c.Semantics.Denominator.StatusCounts[string(StatusUnknown)] != 1 || c.Semantics.Denominator.StatusCounts[string(StatusRefuted)] != 2 {
		t.Fatalf("unexpected denominator counts: %#v", c.Semantics.Denominator.StatusCounts)
	}
}
