#!/usr/bin/env bash
set -euo pipefail

output_dir="${1:?caller-owned output directory is required}"
conformance_dir="$output_dir/conformance"
binary_dir="$output_dir/generated-binaries"
mkdir -p "$binary_dir"

index=0
while IFS= read -r -d '' generated_go; do
	binary="$binary_dir/generated-$index"
	go build -o "$binary" "$generated_go"
	actual="$($binary)"
	relative="${generated_go#"$conformance_dir"/}"
	expected="$(jq -r --arg path "$relative" '.reports[] | select(.generated_go_path == $path) | .observed_status' "$conformance_dir/conformance.json")"
	if [[ -z "$expected" || "$actual" != "$expected" ]]; then
		echo "generated artifact status mismatch for $relative: expected $expected, got $actual" >&2
		exit 1
	fi
	index=$((index + 1))
done < <(find "$conformance_dir/scenarios" -type f -name generated.go -print0 | sort -z)

if [[ "$index" -ne 6 ]]; then
	echo "expected six generated Go artifacts, got $index" >&2
	exit 1
fi
