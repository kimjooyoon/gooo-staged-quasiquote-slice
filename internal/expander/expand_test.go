package expander

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"
)

func TestCorpusDecisionsAndReplay(t *testing.T) {
	contractPath := filepath.Join("..", "..", ".gooo", "staged-quasiquote.gooo")
	c, _, digest, err := contract.Load(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	report, err := RunConformance(c, digest, output)
	if err != nil {
		t.Fatal(err)
	}
	if report.ContractDecision != contract.StatusClosed || report.ScenarioCount != 6 {
		t.Fatalf("unexpected conformance result: %#v", report)
	}
	want := map[string]contract.Status{
		"same-stage-splice":             contract.StatusClosed,
		"hygienic-nested-splice":        contract.StatusClosed,
		"cross-stage-reference":         contract.StatusRefuted,
		"missing-origin":                contract.StatusUnknown,
		"forbidden-compile-time-effect": contract.StatusRefuted,
		"replay":                        contract.StatusClosed,
	}
	for _, scenario := range report.Reports {
		if scenario.ObservedStatus != want[scenario.Scenario] {
			t.Fatalf("scenario %s = %s, want %s", scenario.Scenario, scenario.ObservedStatus, want[scenario.Scenario])
		}
		if len(scenario.Terminal.PhaseSeparationProofPath) < 3 || scenario.Terminal.CaptureDecision == "" {
			t.Fatalf("scenario %s lost terminal proof or capture evidence", scenario.Scenario)
		}
	}
	for _, scenario := range report.Reports {
		if scenario.Scenario == "hygienic-nested-splice" && scenario.Terminal.CaptureDecision != "ALPHA_RENAME" {
			t.Fatalf("nested splice capture decision = %s, want ALPHA_RENAME", scenario.Terminal.CaptureDecision)
		}
	}
}

func TestUnknownCarriesSixFields(t *testing.T) {
	contractPath := filepath.Join("..", "..", ".gooo", "staged-quasiquote.gooo")
	c, _, digest, err := contract.Load(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	report, _, err := RunScenario(c, digest, "missing-origin", t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	unknown := report.Terminal
	if unknown.Decision != contract.StatusUnknown || unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || len(unknown.BlockedBy) == 0 {
		t.Fatalf("UNKNOWN terminal record is incomplete: %#v", unknown)
	}
}
