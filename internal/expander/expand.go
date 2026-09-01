package expander

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"
)

const (
	irSchema       = "gooo.expanded-ast-ir/v1"
	terminalSchema = "gooo.terminal-record/v1"
)

type issue struct {
	status          contract.Status
	stage           string
	step            string
	reason          string
	unknownClass    string
	nextOperation   string
	blockedBy       []string
	counterexample  string
	captureDecision string
}

type evaluation struct {
	contract       contract.Contract
	scenario       contract.Scenario
	ast            []ASTNode
	proof          []ProofStep
	issues         []issue
	defaultCapture string
}

func RunScenario(c contract.Contract, contractDigest, scenarioID, outputDir string, replayBase *ExpandedIR) (ScenarioReport, ExpandedIR, error) {
	scenario, ok := c.Scenario(scenarioID)
	if !ok {
		return ScenarioReport{}, ExpandedIR{}, fmt.Errorf("scenario %q is not declared in .gooo", scenarioID)
	}
	result := evaluate(c, scenario)
	if replayBase != nil {
		sameAST := equalJSON(result.ast, replayBase.AST)
		sameCapture := captureDecisions(result.ast) == captureDecisions(replayBase.AST)
		replayStatus := contract.StatusClosed
		if !sameAST || !sameCapture {
			replayStatus = contract.StatusRefuted
			result.raise(issue{
				status:          contract.StatusRefuted,
				stage:           "replay",
				step:            "compare-expanded-ast",
				reason:          "replay changed AST structure or capture decision",
				counterexample:  scenario.ReplayOf,
				captureDecision: "REPLAY_DIVERGENCE",
			})
		}
		result.proof = append(result.proof, ProofStep{
			Step: "replay",
			From: "expanded-ast",
			To:   "replayed-ast",
			Rule: "replay preserves AST structure and capture decision",
		})
		result.defaultCapture = c.Semantics.CaptureJudgments.ReplayStable
		replay := &ReplayEvidence{
			SourceScenario: scenario.ReplayOf,
			ExpectedSame:   scenario.ExpectedReplay,
			SameAST:        sameAST,
			SameCapture:    sameCapture,
			Status:         replayStatus,
		}
		return writeResult(c, contractDigest, scenario, result, replay, outputDir)
	}
	return writeResult(c, contractDigest, scenario, result, nil, outputDir)
}

func RunConformance(c contract.Contract, contractDigest, outputDir string) (ConformanceReport, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return ConformanceReport{}, err
	}
	completed := map[string]ExpandedIR{}
	reports := make([]ScenarioReport, 0, len(c.Scenarios))
	for _, scenario := range c.Scenarios {
		var replayBase *ExpandedIR
		if scenario.ReplayOf != "" {
			base, ok := completed[scenario.ReplayOf]
			if !ok {
				return ConformanceReport{}, fmt.Errorf("replay source %q was not evaluated before %q", scenario.ReplayOf, scenario.ID)
			}
			replayBase = &base
		}
		report, ir, err := RunScenario(c, contractDigest, scenario.ID, outputDir, replayBase)
		if err != nil {
			return ConformanceReport{}, err
		}
		if report.ObservedStatus != scenario.ExpectedStatus {
			return ConformanceReport{}, fmt.Errorf("scenario %q got %s, contract expects %s", scenario.ID, report.ObservedStatus, scenario.ExpectedStatus)
		}
		completed[scenario.ID] = ir
		reports = append(reports, report)
	}
	expected := map[string]int{}
	observed := map[string]int{}
	corpus := make([]CorpusCase, 0, len(reports))
	for _, report := range reports {
		expected[string(report.ExpectedStatus)]++
		observed[string(report.ObservedStatus)]++
		corpus = append(corpus, CorpusCase{Scenario: report.Scenario, Status: report.ObservedStatus})
	}
	corpusResolution := reduceStatuses(reports, c.Semantics.StatusPrecedence)
	result := ConformanceReport{
		Schema:               "gooo.conformance/v1",
		ContractID:           c.ContractID,
		ContractDigest:       contractDigest,
		ContractDecision:     contract.StatusClosed,
		CorpusResolution:     corpusResolution,
		ScenarioCount:        len(reports),
		ExpectedStatusCounts: expected,
		ObservedStatusCounts: observed,
		Corpus:               corpus,
		Reports:              reports,
	}
	if err := writeJSON(filepath.Join(outputDir, "conformance.json"), result); err != nil {
		return ConformanceReport{}, err
	}
	return result, nil
}

func evaluate(c contract.Contract, scenario contract.Scenario) evaluation {
	result := evaluation{
		contract:       c,
		scenario:       scenario,
		ast:            make([]ASTNode, 0, len(scenario.Expressions)),
		proof:          proofFromContract(c),
		defaultCapture: c.Semantics.CaptureJudgments.NoCollision,
	}
	visible := map[string]string{}
	for _, expression := range scenario.Expressions {
		result.ast = append(result.ast, result.walk(expression, visible))
	}
	return result
}

func (e *evaluation) walk(expression contract.Expression, visible map[string]string) ASTNode {
	e.proof = append(e.proof, ProofStep{
		Step:      "parse-structured-expression",
		From:      "gooo-source",
		To:        "ast-node",
		Rule:      "Go parses declared expression fields without text substitution",
		NodeID:    expression.ID,
		FromStage: expression.Stage,
		ToStage:   expression.Stage,
	})
	decision := e.contract.Semantics.CaptureJudgments.NoCollision
	stableIdentity := ""
	if expression.OriginIdentity == "" {
		decision = e.contract.Semantics.CaptureJudgments.MissingOrigin
		e.raise(issue{
			status:          contract.StatusUnknown,
			stage:           "origin",
			step:            "prove-origin-identity",
			reason:          e.contract.Semantics.Unknown.Reason,
			unknownClass:    e.contract.Semantics.Unknown.UnknownClass,
			nextOperation:   e.contract.Semantics.Unknown.NextOperation,
			blockedBy:       e.contract.Semantics.Unknown.BlockedBy,
			counterexample:  expression.ID,
			captureDecision: decision,
		})
	} else {
		stableIdentity = stableID(e.contract.Semantics.OriginIdentity, expression.OriginIdentity, expression.Stage, expression.ID)
	}
	node := ASTNode{
		Kind:            expression.Op,
		ID:              expression.ID,
		OriginalName:    expression.Name,
		EffectiveName:   expression.Name,
		Value:           expression.Value,
		Stage:           expression.Stage,
		OriginIdentity:  expression.OriginIdentity,
		StableIdentity:  stableIdentity,
		CaptureDecision: decision,
	}

	switch expression.Op {
	case "quote":
		e.proof = append(e.proof, ProofStep{
			Step:      "quote",
			From:      "stage-level",
			To:        "ast-boundary",
			Rule:      e.contract.Semantics.Quote.StageRule,
			NodeID:    expression.ID,
			FromStage: expression.Stage,
			ToStage:   expression.Stage,
		})
		childVisible := cloneVisible(visible)
		if expression.Name != "" {
			if _, collision := childVisible[expression.Name]; collision && stableIdentity != "" {
				node.EffectiveName = expression.Name + "__gooo_" + stableIdentity[len(e.contract.Semantics.OriginIdentity.Prefix):]
				node.CaptureDecision = e.contract.Semantics.CaptureJudgments.FreshCollision
			} else if _, collision := childVisible[expression.Name]; collision {
				node.CaptureDecision = e.contract.Semantics.CaptureJudgments.MissingOrigin
			}
			childVisible[expression.Name] = node.EffectiveName
		}
		for _, child := range expression.Children {
			node.Children = append(node.Children, e.walk(child, childVisible))
		}
	case "splice":
		e.proof = append(e.proof, ProofStep{
			Step:      "splice",
			From:      "quoted-ast",
			To:        "ast-sequence",
			Rule:      e.contract.Semantics.Splice.StageRule,
			NodeID:    expression.ID,
			FromStage: expression.Stage,
			ToStage:   expression.TargetStage,
		})
		if expression.Stage != expression.TargetStage {
			node.CaptureDecision = e.contract.Semantics.CaptureJudgments.CrossStageReference
			e.raise(issue{
				status:          contract.StatusRefuted,
				stage:           "splice",
				step:            "check-stage-level",
				reason:          "splice crosses stage levels",
				counterexample:  expression.ID,
				captureDecision: node.CaptureDecision,
			})
		}
		for _, child := range expression.Children {
			node.Children = append(node.Children, e.walk(child, visible))
		}
	case "reference":
		e.proof = append(e.proof, ProofStep{
			Step:      "reference",
			From:      "reference-stage",
			To:        "target-stage",
			Rule:      "reference stage must equal target stage",
			NodeID:    expression.ID,
			FromStage: expression.Stage,
			ToStage:   expression.TargetStage,
		})
		if expression.Stage != expression.TargetStage {
			node.CaptureDecision = e.contract.Semantics.CaptureJudgments.CrossStageReference
			e.raise(issue{
				status:          contract.StatusRefuted,
				stage:           "reference",
				step:            "check-stage-level",
				reason:          "cross-stage reference is not visible at the target phase",
				counterexample:  expression.ID,
				captureDecision: node.CaptureDecision,
			})
		}
	case "effect":
		e.proof = append(e.proof, ProofStep{
			Step:      "phase-effect",
			From:      expression.PhaseEffect,
			To:        "expander",
			Rule:      "compile-time effect is checked before AST emission",
			NodeID:    expression.ID,
			FromStage: expression.Stage,
			ToStage:   expression.Stage,
		})
		if expression.PhaseEffect == "compile-time" && contains(e.contract.Semantics.PhaseEffects.ForbiddenCompileTime, expression.Effect) {
			node.CaptureDecision = e.contract.Semantics.CaptureJudgments.ForbiddenCompileEffect
			e.raise(issue{
				status:          contract.StatusRefuted,
				stage:           "effect",
				step:            "check-phase-effect",
				reason:          "forbidden compile-time effect would escape the pure expander",
				counterexample:  expression.Effect,
				captureDecision: node.CaptureDecision,
			})
		}
	case "atom":
		// Atom values remain data in the AST. They are never interpreted as Go source.
	}
	return node
}

func (e *evaluation) raise(candidate issue) {
	e.issues = append(e.issues, candidate)
}

func (e evaluation) terminal(replay *ReplayEvidence) TerminalRecord {
	decision := contract.StatusClosed
	winning := issue{}
	for _, candidate := range e.issues {
		if candidate.status.Rank(e.contract.Semantics.StatusPrecedence) > decision.Rank(e.contract.Semantics.StatusPrecedence) {
			decision = candidate.status
			winning = candidate
		}
	}
	terminal := TerminalRecord{
		Schema:                   terminalSchema,
		Decision:                 decision,
		Stage:                    "emit",
		Step:                     "emit-ast",
		Reason:                   "all selected expressions expanded as structured AST nodes",
		UnknownClass:             "",
		NextOperation:            "",
		BlockedBy:                []string{},
		Counterexample:           "",
		CounterexampleDigest:     "",
		PhaseSeparationProofPath: e.proof,
		CaptureDecision:          aggregateCapture(e.ast, e.defaultCapture),
	}
	if replay != nil && decision == contract.StatusClosed {
		terminal.Stage = "replay"
		terminal.Step = "compare-expanded-ast"
		terminal.Reason = "replay preserved AST structure and capture decisions"
		terminal.CaptureDecision = e.contract.Semantics.CaptureJudgments.ReplayStable
	}
	if winning.status != "" {
		terminal.Stage = winning.stage
		terminal.Step = winning.step
		terminal.Reason = winning.reason
		terminal.CaptureDecision = winning.captureDecision
		terminal.Counterexample = winning.counterexample
		terminal.CounterexampleDigest = digestText(winning.counterexample)
		if decision == contract.StatusUnknown {
			terminal.UnknownClass = winning.unknownClass
			terminal.NextOperation = winning.nextOperation
			terminal.BlockedBy = append([]string{}, winning.blockedBy...)
		}
	}
	return terminal
}

func writeResult(c contract.Contract, contractDigest string, scenario contract.Scenario, evaluation evaluation, replay *ReplayEvidence, outputDir string) (ScenarioReport, ExpandedIR, error) {
	terminal := evaluation.terminal(replay)
	if terminal.Decision == contract.StatusUnknown && (terminal.Stage == "" || terminal.Step == "" || terminal.Reason == "" || terminal.UnknownClass == "" || terminal.NextOperation == "" || len(terminal.BlockedBy) == 0) {
		return ScenarioReport{}, ExpandedIR{}, fmt.Errorf("UNKNOWN terminal record is missing one of six required fields")
	}
	sourceDigest := digestJSON(scenario)
	ir := ExpandedIR{
		Schema:         irSchema,
		ContractID:     c.ContractID,
		ContractDigest: contractDigest,
		Scenario:       scenario.ID,
		SourceDigest:   sourceDigest,
		Decision:       terminal.Decision,
		AST:            evaluation.ast,
		Terminal:       terminal,
		Replay:         replay,
	}
	scenarioDir := filepath.Join(outputDir, "scenarios", scenario.ID)
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		return ScenarioReport{}, ExpandedIR{}, err
	}
	irPath := filepath.Join(scenarioDir, "expanded-ir.json")
	terminalPath := filepath.Join(scenarioDir, "terminal-record.json")
	generatedPath := filepath.Join(scenarioDir, "generated.go")
	if err := writeJSON(irPath, ir); err != nil {
		return ScenarioReport{}, ExpandedIR{}, err
	}
	if err := writeJSON(terminalPath, terminal); err != nil {
		return ScenarioReport{}, ExpandedIR{}, err
	}
	generated := EmitGenerated(ir)
	if err := os.WriteFile(generatedPath, generated, 0o644); err != nil {
		return ScenarioReport{}, ExpandedIR{}, err
	}
	return ScenarioReport{
		Scenario:          scenario.ID,
		ExpectedStatus:    scenario.ExpectedStatus,
		ObservedStatus:    terminal.Decision,
		SourceDigest:      sourceDigest,
		ExpandedIRPath:    filepath.Join("scenarios", scenario.ID, "expanded-ir.json"),
		GeneratedGoPath:   filepath.Join("scenarios", scenario.ID, "generated.go"),
		TerminalPath:      filepath.Join("scenarios", scenario.ID, "terminal-record.json"),
		ExpandedIRDigest:  digestJSON(ir),
		GeneratedGoDigest: digestBytes(generated),
		TerminalDigest:    digestJSON(terminal),
		Terminal:          terminal,
		Replay:            replay,
	}, ir, nil
}

func proofFromContract(c contract.Contract) []ProofStep {
	proof := make([]ProofStep, 0, len(c.Semantics.PhaseSeparation.ProofPath))
	for _, rule := range c.Semantics.PhaseSeparation.ProofPath {
		proof = append(proof, ProofStep{Step: rule.Step, From: rule.From, To: rule.To, Rule: rule.Rule})
	}
	return proof
}

func stableID(rule contract.OriginIdentityRule, origin string, stage int, id string) string {
	payload := strings.Join([]string{origin, strconv.Itoa(stage), id}, rule.Delimiter)
	sum := sha256.Sum256([]byte(payload))
	encoded := hex.EncodeToString(sum[:])
	if len(encoded) > rule.IdentityChars {
		encoded = encoded[:rule.IdentityChars]
	}
	return rule.Prefix + encoded
}

func digestText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func cloneVisible(values map[string]string) map[string]string {
	copyValues := make(map[string]string, len(values)+1)
	for key, value := range values {
		copyValues[key] = value
	}
	return copyValues
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func equalJSON(left, right any) bool {
	leftData, _ := json.Marshal(left)
	rightData, _ := json.Marshal(right)
	return string(leftData) == string(rightData)
}

func captureDecisions(nodes []ASTNode) string {
	var decisions []string
	for _, node := range nodes {
		decisions = append(decisions, node.CaptureDecision)
		decisions = append(decisions, captureDecisions(node.Children))
	}
	return strings.Join(decisions, "|")
}

func aggregateCapture(nodes []ASTNode, fallback string) string {
	best := fallback
	priority := map[string]int{
		"CAPTURE_FREE":                  1,
		"REPLAY_STABLE":                 2,
		"ALPHA_RENAME":                  3,
		"MISSING_ORIGIN":                4,
		"CROSS_STAGE_REFUTED":           5,
		"FORBIDDEN_COMPILE_TIME_EFFECT": 6,
	}
	for _, node := range nodes {
		if priority[node.CaptureDecision] > priority[best] {
			best = node.CaptureDecision
		}
		child := aggregateCapture(node.Children, best)
		if priority[child] > priority[best] {
			best = child
		}
	}
	return best
}

func reduceStatuses(reports []ScenarioReport, precedence []contract.Status) contract.Status {
	decision := contract.StatusClosed
	for _, report := range reports {
		if report.ObservedStatus.Rank(precedence) > decision.Rank(precedence) {
			decision = report.ObservedStatus
		}
	}
	return decision
}
