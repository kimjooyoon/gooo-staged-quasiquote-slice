# Release v0.1.0

This release contains the staged quasiquote language slice, its six-case
executable corpus, expanded AST IR, generated Go emitter, replay proof, and
CI evidence contract. GitHub Actions is the verification authority; local Go
execution is not counted as evidence.

The release asset manifest is `SHA256SUMS`. The release workflow refuses an
existing tag or release, creates an annotated tag, and verifies the immutable
release object plus every asset digest after publication.
