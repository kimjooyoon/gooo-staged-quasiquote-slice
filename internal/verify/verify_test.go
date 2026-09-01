package verify

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"
	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/expander"
)

func TestVerifierAcceptsContractBoundArtifacts(t *testing.T) {
	contractPath := filepath.Join("..", "..", ".gooo", "staged-quasiquote.gooo")
	c, _, digest, err := contract.Load(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if _, err := expander.RunConformance(c, digest, output); err != nil {
		t.Fatal(err)
	}
	if err := Verify(c, digest, output); err != nil {
		t.Fatal(err)
	}
}
