package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Status string

const (
	StatusClosed  Status = "CLOSED"
	StatusUnknown Status = "UNKNOWN"
	StatusRefuted Status = "REFUTED"
)

type Contract struct {
	Schema     string     `json:"schema"`
	Authority  string     `json:"authority"`
	ContractID string     `json:"contract_id"`
	Toolchain  Toolchain  `json:"toolchain"`
	Semantics  Semantics  `json:"semantics"`
	Scenarios  []Scenario `json:"scenarios"`
}

type Toolchain struct {
	Go string `json:"go"`
}

type Semantics struct {
	Quote                    QuoteSemantics     `json:"quote"`
	Splice                   SpliceSemantics    `json:"splice"`
	StageLevels              StageLevels        `json:"stage_levels"`
	OriginIdentity           OriginIdentityRule `json:"origin_identity"`
	PhaseEffects             PhaseEffects       `json:"phase_effects"`
	CaptureJudgments         CaptureJudgments   `json:"capture_judgments"`
	StatusPrecedence         []Status           `json:"status_precedence"`
	Unknown                  UnknownRule        `json:"unknown"`
	Denominator              Denominator        `json:"denominator"`
	GenerationPlan           []string           `json:"generation_plan"`
	PhaseSeparation          PhaseSeparation    `json:"phase_separation"`
	NoArbitraryStringReplace bool               `json:"no_arbitrary_string_replacement"`
}

type QuoteSemantics struct {
	Construct string `json:"construct"`
	StageRule string `json:"stage_rule"`
	Output    string `json:"output"`
}

type SpliceSemantics struct {
	Construct  string `json:"construct"`
	SameStage  string `json:"same_stage"`
	CrossStage string `json:"cross_stage"`
	StageRule  string `json:"stage_rule"`
}

type StageLevels struct {
	Source    int `json:"source"`
	Macro     int `json:"macro"`
	Generated int `json:"generated"`
}

type OriginIdentityRule struct {
	Algorithm     string   `json:"algorithm"`
	Inputs        []string `json:"inputs"`
	Delimiter     string   `json:"delimiter"`
	Prefix        string   `json:"prefix"`
	IdentityChars int      `json:"identity_chars"`
}

type PhaseEffects struct {
	Allowed              []string `json:"allowed"`
	ForbiddenCompileTime []string `json:"forbidden_compile_time"`
	Judgment             string   `json:"judgment"`
}

type CaptureJudgments struct {
	FreshCollision         string `json:"fresh_collision"`
	NoCollision            string `json:"no_collision"`
	CrossStageReference    string `json:"cross_stage_reference"`
	MissingOrigin          string `json:"missing_origin"`
	ForbiddenCompileEffect string `json:"forbidden_compile_time_effect"`
	ReplayStable           string `json:"replay_stable"`
}

type UnknownRule struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Denominator struct {
	Scenarios                   []string       `json:"scenarios"`
	StatusCounts                map[string]int `json:"status_counts"`
	Improvement                 string         `json:"improvement"`
	RootReadmeExcludedInventory bool           `json:"root_readme_excluded_from_inventory"`
}

type PhaseSeparation struct {
	ProofPath []ProofRule `json:"proof_path"`
}

type ProofRule struct {
	Step string `json:"step"`
	From string `json:"from"`
	To   string `json:"to"`
	Rule string `json:"rule"`
}

type Scenario struct {
	ID             string       `json:"id"`
	Description    string       `json:"description"`
	Source         string       `json:"source"`
	ExpectedStatus Status       `json:"expected_status"`
	Expressions    []Expression `json:"expressions"`
	ReplayOf       string       `json:"replay_of,omitempty"`
	ExpectedReplay bool         `json:"expected_replay,omitempty"`
}

type Expression struct {
	Op             string       `json:"op"`
	ID             string       `json:"id"`
	Stage          int          `json:"stage"`
	TargetStage    int          `json:"target_stage,omitempty"`
	OriginIdentity string       `json:"origin_identity"`
	PhaseEffect    string       `json:"phase_effect"`
	Effect         string       `json:"effect,omitempty"`
	Name           string       `json:"name,omitempty"`
	Value          string       `json:"value,omitempty"`
	Children       []Expression `json:"children,omitempty"`
}

func Load(path string) (Contract, []byte, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, nil, "", err
	}
	var c Contract
	if err := json.Unmarshal(raw, &c); err != nil {
		return Contract{}, nil, "", fmt.Errorf("parse .gooo contract: %w", err)
	}
	if err := c.Validate(); err != nil {
		return Contract{}, nil, "", err
	}
	sum := sha256.Sum256(raw)
	return c, raw, "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (c Contract) Validate() error {
	if c.Schema != "gooo.staged-quasiquote/v1" || c.Authority != "metacode" || c.ContractID == "" {
		return errors.New(".gooo contract must be authoritative and versioned")
	}
	if c.Toolchain.Go != "1.27" {
		return fmt.Errorf("contract toolchain must be Go 1.27, got %q", c.Toolchain.Go)
	}
	if c.Semantics.Quote.Construct != "AST" || c.Semantics.Quote.Output != "ast_node" || c.Semantics.Quote.StageRule == "" {
		return errors.New("quote semantics must construct AST nodes")
	}
	if c.Semantics.Splice.Construct != "AST_SEQUENCE" || c.Semantics.Splice.SameStage != string(StatusClosed) || c.Semantics.Splice.CrossStage != string(StatusRefuted) || c.Semantics.Splice.StageRule == "" {
		return errors.New("splice semantics must close same-stage and refute cross-stage expansion")
	}
	if c.Semantics.StageLevels.Source != 0 || c.Semantics.StageLevels.Macro != 1 || c.Semantics.StageLevels.Generated != 2 {
		return errors.New("stage levels must be source=0, macro=1, generated=2")
	}
	identity := c.Semantics.OriginIdentity
	if identity.Algorithm != "sha256" || !equalStrings(identity.Inputs, []string{"origin_identity", "stage_level", "node_id"}) || identity.Delimiter != "|" || identity.Prefix != "origin_" || identity.IdentityChars < 12 {
		return errors.New("origin identity semantics are incomplete")
	}
	if !containsString(c.Semantics.PhaseEffects.Allowed, "pure-expand") || c.Semantics.PhaseEffects.Judgment != string(StatusRefuted) || len(c.Semantics.PhaseEffects.ForbiddenCompileTime) == 0 {
		return errors.New("phase effect semantics are incomplete")
	}
	if !equalStatuses(c.Semantics.StatusPrecedence, []Status{StatusRefuted, StatusUnknown, StatusClosed}) {
		return errors.New("status precedence must be REFUTED > UNKNOWN > CLOSED")
	}
	u := c.Semantics.Unknown
	if u.Stage == "" || u.Step == "" || u.Reason == "" || u.UnknownClass == "" || u.NextOperation == "" || len(u.BlockedBy) == 0 {
		return errors.New("UNKNOWN must preserve stage, step, reason, unknown_class, next_operation, and blocked_by")
	}
	if c.Semantics.CaptureJudgments.FreshCollision == "" || c.Semantics.CaptureJudgments.NoCollision == "" || c.Semantics.CaptureJudgments.CrossStageReference == "" || c.Semantics.CaptureJudgments.MissingOrigin == "" || c.Semantics.CaptureJudgments.ForbiddenCompileEffect == "" || c.Semantics.CaptureJudgments.ReplayStable == "" {
		return errors.New("capture judgments are incomplete")
	}
	if len(c.Semantics.GenerationPlan) < 5 || !c.Semantics.NoArbitraryStringReplace || len(c.Semantics.PhaseSeparation.ProofPath) < 3 {
		return errors.New("generation plan, phase proof, or structured-emitter guard is incomplete")
	}
	if !c.Semantics.Denominator.RootReadmeExcludedInventory || c.Semantics.Denominator.Improvement == "" {
		return errors.New("denominator must define improvement uncertainty and README exclusion")
	}
	if len(c.Scenarios) == 0 || !equalStrings(c.Semantics.Denominator.Scenarios, scenarioIDs(c.Scenarios)) {
		return errors.New(".gooo denominator must enumerate the executable corpus")
	}
	counts := map[string]int{}
	seen := map[string]bool{}
	for _, scenario := range c.Scenarios {
		if scenario.ID == "" || seen[scenario.ID] || scenario.Description == "" || scenario.Source == "" || len(scenario.Expressions) == 0 {
			return fmt.Errorf("invalid or duplicate scenario %q", scenario.ID)
		}
		seen[scenario.ID] = true
		if scenario.ExpectedStatus != StatusClosed && scenario.ExpectedStatus != StatusUnknown && scenario.ExpectedStatus != StatusRefuted {
			return fmt.Errorf("scenario %q has invalid expected status %q", scenario.ID, scenario.ExpectedStatus)
		}
		counts[string(scenario.ExpectedStatus)]++
		for _, expression := range scenario.Expressions {
			if err := validateExpression(expression); err != nil {
				return fmt.Errorf("scenario %q: %w", scenario.ID, err)
			}
		}
		if scenario.ReplayOf != "" && scenario.ReplayOf == scenario.ID {
			return fmt.Errorf("scenario %q cannot replay itself", scenario.ID)
		}
	}
	for _, status := range []Status{StatusClosed, StatusUnknown, StatusRefuted} {
		if c.Semantics.Denominator.StatusCounts[string(status)] != counts[string(status)] {
			return fmt.Errorf("denominator status count for %s does not match corpus", status)
		}
	}
	for _, scenario := range c.Scenarios {
		if scenario.ReplayOf != "" && !seen[scenario.ReplayOf] {
			return fmt.Errorf("scenario %q replays unknown scenario %q", scenario.ID, scenario.ReplayOf)
		}
	}
	return nil
}

func validateExpression(expression Expression) error {
	if expression.ID == "" || expression.Stage < 0 || expression.PhaseEffect == "" {
		return errors.New("expression needs id, non-negative stage, and phase_effect")
	}
	switch expression.Op {
	case "quote", "splice", "atom", "reference", "effect":
	default:
		return fmt.Errorf("unsupported expression op %q", expression.Op)
	}
	if expression.Op == "splice" && len(expression.Children) == 0 {
		return fmt.Errorf("splice %q must contain a structured child", expression.ID)
	}
	if expression.Op == "quote" && len(expression.Children) == 0 && expression.Name == "" && expression.Value == "" {
		return fmt.Errorf("quote %q must carry a child or AST value", expression.ID)
	}
	if expression.Op == "effect" && expression.Effect == "" {
		return fmt.Errorf("effect %q must name an effect", expression.ID)
	}
	for _, child := range expression.Children {
		if err := validateExpression(child); err != nil {
			return err
		}
	}
	return nil
}

func (c Contract) Scenario(id string) (Scenario, bool) {
	for _, scenario := range c.Scenarios {
		if scenario.ID == id {
			return scenario, true
		}
	}
	return Scenario{}, false
}

func scenarioIDs(scenarios []Scenario) []string {
	ids := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		ids = append(ids, scenario.ID)
	}
	return ids
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalStatuses(left, right []Status) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func (s Status) Rank(precedence []Status) int {
	for index, status := range precedence {
		if s == status {
			return len(precedence) - index
		}
	}
	return 0
}

func NormalizeText(value string) string {
	return strings.TrimSpace(value)
}
