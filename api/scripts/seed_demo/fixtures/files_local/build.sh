#!/usr/bin/env bash
# build.sh — convert every .md under sources/ to the MIME declared in its
# frontmatter, writing outputs to generated/<filename>.
#
# Requires: pandoc 3.x + a LaTeX engine (pdflatex / xelatex) on PATH.
#
# Usage:
#   cd api/scripts/seed_demo/fixtures/files_local
#   ./build.sh
#
# Exit code 0 on full success; non-zero if any source fails to convert.

set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
SOURCES="$HERE/sources"
OUT="$HERE/generated"

mkdir -p "$OUT"

total=0
failed=0

extract_frontmatter_field() {
    # Read the first YAML frontmatter block and grep a single `key: value` line.
    # Keeps dependencies to bash + awk + sed.
    local file="$1"
    local key="$2"
    awk '
        /^---[[:space:]]*$/ { fm++; next }
        fm == 1 { print }
        fm >= 2 { exit }
    ' "$file" | sed -n "s/^${key}:[[:space:]]*\"\{0,1\}\([^\"]*\)\"\{0,1\}[[:space:]]*$/\1/p" | head -n 1
}

convert_one() {
    local src="$1"
    local mime filename out
    mime=$(extract_frontmatter_field "$src" mime)
    filename=$(extract_frontmatter_field "$src" filename)
    if [[ -z "$mime" || -z "$filename" ]]; then
        echo "[build.sh] SKIP $src (missing mime or filename frontmatter)"
        return 1
    fi
    out="$OUT/$filename"

    case "$mime" in
        application/pdf)
            pandoc "$src" -o "$out" --pdf-engine=pdflatex --metadata title:"$filename" 2>/dev/null \
                || pandoc "$src" -o "$out" --pdf-engine=xelatex --metadata title:"$filename"
            ;;
        application/vnd.openxmlformats-officedocument.wordprocessingml.document)
            pandoc "$src" -o "$out" --to docx
            ;;
        application/vnd.openxmlformats-officedocument.presentationml.presentation)
            # Pandoc slides: every level-1 header (# or ##) becomes a new slide.
            pandoc "$src" -o "$out" --to pptx --slide-level=2
            ;;
        application/epub+zip)
            pandoc "$src" -o "$out" --to epub3 --metadata title:"$filename"
            ;;
        text/plain)
            # Pandoc 'plain' writer strips markdown formatting.
            pandoc "$src" -o "$out" --to plain --wrap=preserve
            ;;
        *)
            echo "[build.sh] SKIP $src (unsupported mime '$mime')"
            return 1
            ;;
    esac
}

while IFS= read -r -d '' src; do
    total=$((total + 1))
    rel="${src#$SOURCES/}"
    if convert_one "$src"; then
        echo "[build.sh] OK   $rel"
    else
        failed=$((failed + 1))
        echo "[build.sh] FAIL $rel"
    fi
done < <(find "$SOURCES" -type f -name '*.md' -print0 | LC_ALL=C sort -z)

echo "[build.sh] done — $total sources, $failed failures"
if [[ "$failed" -gt 0 ]]; then
    exit 1
fi
