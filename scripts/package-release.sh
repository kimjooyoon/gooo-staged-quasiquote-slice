#!/usr/bin/env bash
set -euo pipefail

version="${1:?version is required}"
output_dir="${2:?caller-owned release directory is required}"
repo_root="${GITHUB_WORKSPACE:-$(pwd)}"
mkdir -p "$output_dir"

prefix="gooo-staged-quasiquote-slice-${version}"
tar -czf "$output_dir/$prefix.tar.gz" \
	--exclude=.git \
	--exclude=out \
	--exclude=tmp \
	-C "$repo_root" .
cp "$repo_root/.gooo/staged-quasiquote.gooo" "$output_dir/$prefix-contract.gooo"
cp "$repo_root/contracts/terminal-record-v1.json" "$output_dir/$prefix-terminal-record-schema.json"
cp "$repo_root/contracts/oracle-lock-v1.json" "$output_dir/$prefix-oracle-lock.json"
jq -n \
	--arg version "$version" \
	--arg commit "${GITHUB_SHA:-unknown}" \
	--arg contract_digest "sha256:$(shasum -a 256 "$repo_root/.gooo/staged-quasiquote.gooo" | awk '{print $1}')" \
	'{schema:"gooo.release-manifest/v1", version:$version, commit:$commit, contract_digest:$contract_digest, authority:"github-actions"}' >"$output_dir/$prefix-manifest.json"
(
	cd "$output_dir"
	shasum -a 256 "$prefix.tar.gz" "$prefix-contract.gooo" "$prefix-terminal-record-schema.json" "$prefix-oracle-lock.json" "$prefix-manifest.json" > SHA256SUMS
)
