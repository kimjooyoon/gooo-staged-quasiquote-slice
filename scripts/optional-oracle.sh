#!/usr/bin/env bash
set -euo pipefail

output_dir="${1:?caller-owned output directory is required}"
lock="contracts/oracle-lock-v1.json"
asset_name="$(jq -r '.asset.name' "$lock")"
asset_digest="$(jq -r '.asset.sha256' "$lock" | sed 's/^sha256://')"
asset_url="https://github.com/kimjooyoon/gooo-hygienic-origin-resolver/releases/download/v0.1.1/$asset_name"
mkdir -p "$output_dir"
curl --fail --location --silent --show-error "$asset_url" --output "$output_dir/$asset_name"
actual="$(shasum -a 256 "$output_dir/$asset_name" | awk '{print $1}')"
if [[ "$actual" != "$asset_digest" ]]; then
	echo "optional oracle asset digest mismatch" >&2
	exit 1
fi
jq -n \
	--arg repository "kimjooyoon/gooo-hygienic-origin-resolver" \
	--arg tag "v0.1.1" \
	--arg asset "$asset_name" \
	--arg digest "sha256:$actual" \
	'{schema:"gooo.optional-oracle-consumption/v1", optional:true, required_cross_project_gate:false, repository:$repository, tag:$tag, asset:$asset, digest:$digest, advisory_only:true}' >"$output_dir/oracle-consumption.json"
