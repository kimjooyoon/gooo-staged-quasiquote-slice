package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"
	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/expander"
)

func Verify(c contract.Contract, contractDigest, conformanceDir string) error {
	conformancePath := filepath.Join(conformanceDir, "conformance.json")
	data, err := os.ReadFile(conformancePath)
	if err != nil {
		return err
	}
	var report expander.ConformanceReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("parse conformance report: %w", err)
	}
	if report.Schema != "gooo.conformance/v1" || report.ContractID != c.ContractID || report.ContractDigest != contractDigest {
		return fmt.Errorf("conformance report does not bind to the declared .gooo contract")
	}
	if report.ContractDecision != contract.StatusClosed || report.ScenarioCount != len(c.Scenarios) || len(report.Reports) != len(c.Scenarios) {
		return fmt.Errorf("conformance contract decision or denominator count is not closed")
	}
	for _, scenario := range c.Scenarios {
		found := false
		for _, scenarioReport := range report.Reports {
			if scenarioReport.Scenario != scenario.ID {
				continue
			}
			found = true
			if scenarioReport.ExpectedStatus != scenario.ExpectedStatus || scenarioReport.ObservedStatus != scenario.ExpectedStatus {
				return fmt.Errorf("scenario %q does not match its .gooo judgment", scenario.ID)
			}
			if err := verifyScenario(conformanceDir, scenarioReport); err != nil {
				return err
			}
		}
		if !found {
			return fmt.Errorf("scenario %q is absent from conformance report", scenario.ID)
		}
	}
	return nil
}

func verifyScenario(root string, report expander.ScenarioReport) error {
	irPath := filepath.Join(root, report.ExpandedIRPath)
	generatedPath := filepath.Join(root, report.GeneratedGoPath)
	terminalPath := filepath.Join(root, report.TerminalPath)
	irData, err := os.ReadFile(irPath)
	if err != nil {
		return err
	}
	generatedData, err := os.ReadFile(generatedPath)
	if err != nil {
		return err
	}
	terminalData, err := os.ReadFile(terminalPath)
	if err != nil {
		return err
	}
	if digestBytes(irData) != report.ExpandedIRDigest || digestBytes(generatedData) != report.GeneratedGoDigest || digestBytes(terminalData) != report.TerminalDigest {
		return fmt.Errorf("scenario %q has an artifact digest mismatch", report.Scenario)
	}
	var ir expander.ExpandedIR
	if err := json.Unmarshal(irData, &ir); err != nil {
		return fmt.Errorf("parse expanded IR for %q: %w", report.Scenario, err)
	}
	var terminal expander.TerminalRecord
	if err := json.Unmarshal(terminalData, &terminal); err != nil {
		return fmt.Errorf("parse terminal record for %q: %w", report.Scenario, err)
	}
	if ir.Decision != report.ObservedStatus || terminal.Decision != report.ObservedStatus || terminal.CaptureDecision == "" || len(terminal.PhaseSeparationProofPath) == 0 {
		return fmt.Errorf("scenario %q lost decision, capture, or phase-proof evidence", report.Scenario)
	}
	if report.ObservedStatus == contract.StatusUnknown {
		if terminal.Stage == "" || terminal.Step == "" || terminal.Reason == "" || terminal.UnknownClass == "" || terminal.NextOperation == "" || len(terminal.BlockedBy) == 0 {
			return fmt.Errorf("UNKNOWN scenario %q is missing one of six required fields", report.Scenario)
		}
	}
	generatedText := string(generatedData)
	for _, required := range []string{"type ASTNode struct", "var Expanded = []ASTNode", "type TerminalRecord struct", "var Terminal = TerminalRecord"} {
		if !strings.Contains(generatedText, required) {
			return fmt.Errorf("generated Go for %q is not an AST-only artifact: missing %q", report.Scenario, required)
		}
	}
	if strings.Contains(generatedText, "strings.Replace") || strings.Contains(generatedText, "strings.ReplaceAll") {
		return fmt.Errorf("generated Go for %q contains arbitrary string replacement", report.Scenario)
	}
	return nil
}

func digestBytes(data []byte) string {
	// Kept local to make the verifier's artifact check independent of the
	// expander's write path while using the same published digest shape.
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
