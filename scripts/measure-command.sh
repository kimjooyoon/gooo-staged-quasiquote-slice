#!/usr/bin/env bash
set -euo pipefail

metric_path="${1:?metric path is required}"
shift
metric_base="${metric_path%.json}"
mkdir -p "$(dirname "$metric_path")"
start_ms="$(date +%s%3N)"
set +e
/usr/bin/time -f '%e %M' -o "${metric_base}.time" "$@" >"${metric_base}.stdout" 2>"${metric_base}.stderr"
status=$?
set -e
end_ms="$(date +%s%3N)"
wall_ms=$((end_ms - start_ms))
read -r elapsed_seconds peak_rss_kib <"${metric_base}.time"
if [[ -z "${elapsed_seconds:-}" ]]; then
	elapsed_seconds="0"
fi
if [[ -z "${peak_rss_kib:-}" ]]; then
	peak_rss_kib="0"
fi
jq -n \
	--argjson wall_ms "$wall_ms" \
	--argjson peak_rss_kib "$peak_rss_kib" \
	--argjson exit_code "$status" \
	'{wall_ms: $wall_ms, peak_rss_kib: $peak_rss_kib, exit_code: $exit_code}' >"$metric_path"
exit "$status"
