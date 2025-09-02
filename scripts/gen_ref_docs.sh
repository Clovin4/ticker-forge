#!/usr/bin/env bash
set -euo pipefail

OUT_DIR="docs/reference"
mkdir -p "$OUT_DIR"

# Optional: turn your Git remote into a browsable HTTPS URL for source links
REPO_URL="${REPO_URL:-$(git config --get remote.origin.url || true)}"
REPO_URL="${REPO_URL/git@github.com:/https://github.com/}"
REPO_URL="${REPO_URL/.git/}"
REPO_URL="${REPO_URL/ssh:/https:}"

# Ensure gomarkdoc is installed
command -v gomarkdoc >/dev/null 2>&1 || {
  echo "gomarkdoc not found. Install with:"
  echo "  go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest"
  exit 1
}

# Enumerate all packages (skip vendor if present)
pkgs=$(go list ./... | grep -v '/vendor/')

# Generate one Markdown file per package under docs/reference/
for pkg in $pkgs; do
  out="${OUT_DIR}/${pkg//\//_}.md"   # turn "internal/chart" into "internal_chart.md"
  if [[ -n "$REPO_URL" ]]; then
    gomarkdoc "$pkg" --output "$out" --repository.url "$REPO_URL"
  else
    gomarkdoc "$pkg" --output "$out"
  fi
  echo "✓ $pkg -> $out"
done

# Build an index page
{
  echo "# API Reference"
  echo
  echo "> Generated with \`gomarkdoc\`. Run \`scripts/gen_ref_docs.sh\` after code changes."
  echo
  for f in $(ls "$OUT_DIR" | grep -v README.md | sort); do
    title="${f%.md}"
    echo "- [${title}](${f})"
  done
} > "$OUT_DIR/README.md"

echo "All done → ${OUT_DIR}"
