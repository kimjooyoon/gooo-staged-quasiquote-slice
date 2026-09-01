#!/usr/bin/env bash
set -euo pipefail

repo_root="${1:?repository root is required}"
artifact_root="${2:?generated artifact root is required}"
timing_root="${3:?timing root is required}"
test_json="${4:?go test JSON output is required}"
conformance_json="${5:?conformance JSON is required}"
output_path="${6:?evidence path is required}"

go_files=0
gooo_files=0
go_lines=0
gooo_lines=0
regular_files=0
subdirectories=0

physical_lines() {
	awk 'END { print NR + 0 }' "$1"
}

file_size() {
	stat -c %s "$1" 2>/dev/null || stat -f %z "$1"
}

while IFS= read -r -d '' path; do
	if [[ -d "$path" ]]; then
		if [[ "$path" != "$repo_root" ]]; then
			subdirectories=$((subdirectories + 1))
		fi
		continue
	fi
	if [[ ! -f "$path" || "$path" == "$repo_root/README.md" ]]; then
		continue
	fi
	regular_files=$((regular_files + 1))
	case "$path" in
		*.go)
			go_files=$((go_files + 1))
			go_lines=$((go_lines + $(physical_lines "$path")))
			;;
		*.gooo)
			gooo_files=$((gooo_files + 1))
			gooo_lines=$((gooo_lines + $(physical_lines "$path")))
			;;
	esac
done < <(find "$repo_root" \( -path "$repo_root/.git" -o -path "$repo_root/.git/*" \) -prune -o -print0)

artifact_count=0
artifact_bytes=0
while IFS= read -r -d '' path; do
	if [[ -f "$path" ]]; then
		artifact_count=$((artifact_count + 1))
		artifact_bytes=$((artifact_bytes + $(file_size "$path")))
	fi
done < <(find "$artifact_root" -type f -print0)

test_metrics="$(jq -s '{
	total: ([.[] | select(.Action == "run" and (.Test // "") != "")] | length),
	selected: ([.[] | select(.Action == "run" and (.Test // "") != "")] | length),
	executed: ([.[] | select(.Action == "run" and (.Test // "") != "")] | length),
	reused: ([.[] | select(.Action == "output" and ((.Output // "") | contains("(cached)")))] | length),
	failed: ([.[] | select(.Action == "fail" and (.Test // "") != "")] | length),
	unknown: ([.[] | select(.Action == "skip" and (.Test // "") != "")] | length)
}' "$test_json")"

inventory="$(jq -n \
	--argjson go_files "$go_files" \
	--argjson gooo_files "$gooo_files" \
	--argjson go_physical_lines "$go_lines" \
	--argjson gooo_physical_lines "$gooo_lines" \
	--argjson subdirectories "$subdirectories" \
	--argjson regular_files "$regular_files" \
	'{go_files:$go_files, gooo_files:$gooo_files, go_physical_lines:$go_physical_lines, gooo_physical_lines:$gooo_physical_lines, subdirectories:$subdirectories, regular_files:$regular_files, root_readme_excluded:true}')"

runtime="$(jq -n \
	--slurpfile compile "$timing_root/compile.json" \
	--slurpfile build "$timing_root/build.json" \
	--slurpfile test "$timing_root/test.json" \
	--slurpfile conformance "$timing_root/conformance.json" \
	--slurpfile integration "$timing_root/integration.json" \
	'{compile:$compile[0], build:$build[0], test:$test[0], conformance:$conformance[0], integration:$integration[0]}')"

corpus="$(jq '{scenario_count, contract_decision, corpus_resolution, expected_status_counts, observed_status_counts, cases: [.corpus[] | {scenario, status}]}' "$conformance_json")"
contract_digest="$(shasum -a 256 "$repo_root/.gooo/staged-quasiquote.gooo" | awk '{print "sha256:"$1}')"

jq -n \
	--arg contract_digest "$contract_digest" \
	--argjson inventory "$inventory" \
	--argjson runtime "$runtime" \
	--argjson tests "$test_metrics" \
	--argjson corpus "$corpus" \
	--argjson artifact_count "$artifact_count" \
	--argjson artifact_bytes "$artifact_bytes" \
	'{schema:"gooo.ci-evidence/v1", authority:"github-actions", contract_digest:$contract_digest, inventory:$inventory, runtime:$runtime, tests:$tests, generated_artifacts:{count:$artifact_count, bytes:$artifact_bytes}, corpus:$corpus, improvement:{status:"UNKNOWN", matched_scenario:false, reason:"no matched scenario/source/contract/toolchain before/after integer pair", before:null, after:null}}' >"$output_path"
