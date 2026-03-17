#!/bin/bash
# download-and-merge.sh — Downloads lists from URLs, merges with local additions,
# cleans up and deduplicates the result.
#
# Usage: download-and-merge.sh <source_dir> <output_file> [--strip-dot-prefix]
#   <source_dir>        Directory containing lists-to-download.txt and addition.txt
#   <output_file>       Path to write the cleaned, deduplicated list
#   --strip-dot-prefix  Optional: remove leading '+.' and '.' from lines (for geosite domains)

set -euo pipefail

SOURCE_DIR="$1"
OUTPUT_FILE="$2"
STRIP_DOT_PREFIX="${3:-}"

temp_list=$(mktemp)
trap 'rm -f "$temp_list"' EXIT

# Download lists
if [ -f "${SOURCE_DIR}/lists-to-download.txt" ]; then
  while IFS= read -r url || [[ -n "$url" ]]; do
    [[ -z "$url" ]] && continue
    echo "Downloading $url"
    curl -sSLf --retry 3 --max-filesize 10485760 "$url" >> "$temp_list"
    printf '\n' >> "$temp_list"
  done < "${SOURCE_DIR}/lists-to-download.txt"
fi

# Add local additions
if [ -f "${SOURCE_DIR}/addition.txt" ]; then
  cat "${SOURCE_DIR}/addition.txt" >> "$temp_list"
  printf '\n' >> "$temp_list"
fi

# Strip leading '+.' and '.' if requested (for geosite)
if [ "$STRIP_DOT_PREFIX" = "--strip-dot-prefix" ]; then
  sed -i -e 's/^+\.//g' -e 's/^\.//g' "$temp_list"
fi

# Clean: remove carriage returns, empty lines, sort unique
tr -d '\r' < "$temp_list" | awk 'NF' | sort -u > "$OUTPUT_FILE"

echo "Result: $(wc -l < "$OUTPUT_FILE") unique lines"
