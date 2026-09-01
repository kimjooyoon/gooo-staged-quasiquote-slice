package expander

import "github.com/kimjooyoon/gooo-staged-quasiquote-slice/internal/contract"

type ASTNode struct {
	Kind            string    `json:"kind"`
	ID              string    `json:"id"`
	OriginalName    string    `json:"original_name,omitempty"`
	EffectiveName   string    `json:"effective_name,omitempty"`
	Value           string    `json:"value,omitempty"`
	Stage           int       `json:"stage"`
	OriginIdentity  string    `json:"origin_identity,omitempty"`
	StableIdentity  string    `json:"stable_identity,omitempty"`
	CaptureDecision string    `json:"capture_decision"`
	Children        []ASTNode `json:"children,omitempty"`
}

type ProofStep struct {
	Step      string `json:"step"`
	From      string `json:"from"`
	To        string `json:"to"`
	Rule      string `json:"rule"`
	NodeID    string `json:"node_id,omitempty"`
	FromStage int    `json:"from_stage,omitempty"`
	ToStage   int    `json:"to_stage,omitempty"`
}

type TerminalRecord struct {
	Schema                   string          `json:"schema"`
	Decision                 contract.Status `json:"decision"`
	Stage                    string          `json:"stage"`
	Step                     string          `json:"step"`
	Reason                   string          `json:"reason"`
	UnknownClass             string          `json:"unknown_class"`
	NextOperation            string          `json:"next_operation"`
	BlockedBy                []string        `json:"blocked_by"`
	Counterexample           string          `json:"counterexample"`
	CounterexampleDigest     string          `json:"counterexample_digest"`
	PhaseSeparationProofPath []ProofStep     `json:"phase_separation_proof_path"`
	CaptureDecision          string          `json:"capture_decision"`
}

type ReplayEvidence struct {
	SourceScenario string          `json:"source_scenario"`
	ExpectedSame   bool            `json:"expected_same"`
	SameAST        bool            `json:"same_ast"`
	SameCapture    bool            `json:"same_capture_decision"`
	Status         contract.Status `json:"status"`
}

type ExpandedIR struct {
	Schema         string          `json:"schema"`
	ContractID     string          `json:"contract_id"`
	ContractDigest string          `json:"contract_digest"`
	Scenario       string          `json:"scenario"`
	SourceDigest   string          `json:"source_digest"`
	Decision       contract.Status `json:"decision"`
	AST            []ASTNode       `json:"ast"`
	Terminal       TerminalRecord  `json:"terminal"`
	Replay         *ReplayEvidence `json:"replay,omitempty"`
}

type ScenarioReport struct {
	Scenario          string          `json:"scenario"`
	ExpectedStatus    contract.Status `json:"expected_status"`
	ObservedStatus    contract.Status `json:"observed_status"`
	SourceDigest      string          `json:"source_digest"`
	ExpandedIRPath    string          `json:"expanded_ir_path"`
	GeneratedGoPath   string          `json:"generated_go_path"`
	TerminalPath      string          `json:"terminal_record_path"`
	ExpandedIRDigest  string          `json:"expanded_ir_digest"`
	GeneratedGoDigest string          `json:"generated_go_digest"`
	TerminalDigest    string          `json:"terminal_digest"`
	Terminal          TerminalRecord  `json:"terminal"`
	Replay            *ReplayEvidence `json:"replay,omitempty"`
}

type CorpusCase struct {
	Scenario string          `json:"scenario"`
	Status   contract.Status `json:"status"`
}

type ConformanceReport struct {
	Schema               string           `json:"schema"`
	ContractID           string           `json:"contract_id"`
	ContractDigest       string           `json:"contract_digest"`
	ContractDecision     contract.Status  `json:"contract_decision"`
	CorpusResolution     contract.Status  `json:"corpus_resolution"`
	ScenarioCount        int              `json:"scenario_count"`
	ExpectedStatusCounts map[string]int   `json:"expected_status_counts"`
	ObservedStatusCounts map[string]int   `json:"observed_status_counts"`
	Corpus               []CorpusCase     `json:"corpus"`
	Reports              []ScenarioReport `json:"reports"`
}
