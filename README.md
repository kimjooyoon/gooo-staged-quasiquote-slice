# gooo-staged-quasiquote-slice

An executable Gooo language slice for staged quasiquotation and hygienic
splicing. The authoritative meaning is in
[`.gooo/staged-quasiquote.gooo`](.gooo/staged-quasiquote.gooo). It defines
`quote`, `splice`, stage levels, origin identity, phase effects, the
`REFUTED > UNKNOWN > CLOSED` precedence, the denominator, and the generation
plan.

The Go expander parses that contract, evaluates its declared expressions, and
emits an expanded IR plus generated Go containing typed AST structures only.
It does not replace source text. Every terminal record carries a phase
separation proof path and a capture decision. UNKNOWN records preserve
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.

## Executable corpus

The denominator is exactly six scenarios declared by the `.gooo` contract:

| Scenario | Expected decision |
| --- | --- |
| same-stage-splice | CLOSED |
| hygienic-nested-splice | CLOSED |
| cross-stage-reference | REFUTED |
| missing-origin | UNKNOWN |
| forbidden-compile-time-effect | REFUTED |
| replay | CLOSED |

The optional `gooo-hygienic-origin-resolver` v0.1.1 oracle is locked by
[`contracts/oracle-lock-v1.json`](contracts/oracle-lock-v1.json). It is
advisory-only and is not a required cross-project gate.

## Commands and evidence

Commands write all reports and generated files to a caller-owned output
directory:

```text
go run ./cmd/gooo-expander conformance \
  --contract .gooo/staged-quasiquote.gooo \
  --output-dir /caller-owned/output
go run ./cmd/gooo-verify verify \
  --contract .gooo/staged-quasiquote.gooo \
  --conformance-dir /caller-owned/output
```

GitHub Actions with Go 1.27 is the only success authority. The required CI
job formats, compiles, builds, vets, tests, runs conformance, verifies the
expanded IR and generated Go, executes every generated artifact, and uploads
exact integer inventory, physical-line, artifact, runtime, peak-RSS, and test
metrics. The root README is excluded from inventory. Local Go test, build,
vet, and conformance runs are intentionally not evidence.

The repository began with the bootstrap exception from
`gooo-repository-bootstrap` v0.1.1. Functional changes are PR-only. Releases
are annotated, immutable, digest-checked, and never overwritten.
