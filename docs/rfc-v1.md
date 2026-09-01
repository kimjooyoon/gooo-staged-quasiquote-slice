# Staged quasiquote and splice RFC v1

## Authority

`.gooo/staged-quasiquote.gooo` is the executable semantic contract. Its
`semantics` object owns the language meaning, decision precedence, denominator,
phase proof rules, and generation plan. Go code is limited to structured-data
parsing, contract execution, and typed AST emission.

## Stages and forms

The contract fixes source level 0, macro level 1, and generated level 2.
`quote` creates an AST boundary. `splice` creates an AST sequence and is
closed only when its source and target stages are equal. References must also
remain at one stage; a cross-stage reference is refuted.

Every expression with an origin identity receives a stable identity derived by
SHA-256 from `origin_identity|stage_level|node_id`. A visible fresh-name
collision is alpha-renamed from that identity. Generated Go carries the
identity, original/effective names, and capture decision as AST fields.

Compile-time expansion accepts only the declared pure effect. Filesystem,
network, and process effects are refuted before emission. Missing origin proof
is UNKNOWN and retains all six required fields for the next operation.

## Phase separation proof

The runtime record begins with the contract proof path:

1. `.gooo` source → structured semantics: the contract owns meaning and the
   denominator.
2. stage-level expression → AST node: quote and same-stage splice preserve
   structure.
3. expanded AST IR → typed Go data: the emitter serializes fields and never
   performs arbitrary source-text replacement.

Runtime parse, quote, splice, reference, effect, and replay steps are appended
to that path in the terminal record.

## Decision and improvement boundary

The lattice is `REFUTED > UNKNOWN > CLOSED`. Corpus cases intentionally retain
their declared REFUTED and UNKNOWN decisions; conformance is CLOSED when every
observed decision matches the contract. Improvement is UNKNOWN unless matched
scenario, source, contract, and toolchain digests have exact integer before and
after values.
