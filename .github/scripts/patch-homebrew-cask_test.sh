#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
PATCHER="$SCRIPT_DIR/patch-homebrew-cask.sh"
FIXTURE="$SCRIPT_DIR/testdata/replicator-v0.5.0.rb"
RENDERED_CASK=${1:-}
ARCHIVE_NAME="replicator_0.5.0_darwin_arm64.tar.gz"
LINUX_ARM64_SHA="d35cf51192f4bc3eb92d32c2a63304fdbc561243a2bb8e406d0a5c7f9d1a83f1"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

if [ ! -x "$PATCHER" ]; then
  fail "patcher is not executable: $PATCHER"
fi

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

new_case() {
  CASE_DIR="$WORK/$1"
  mkdir -p "$CASE_DIR"
  cp "$FIXTURE" "$CASE_DIR/replicator.rb"
  printf 'signed darwin archive fixture\n' > "$CASE_DIR/$ARCHIVE_NAME"
  ARCHIVE_SHA=$(sha256sum "$CASE_DIR/$ARCHIVE_NAME" | awk '{ print $1 }')
  printf '%s  %s\n' "$ARCHIVE_SHA" "$ARCHIVE_NAME" > "$CASE_DIR/checksums.txt"
}

assert_success() {
  "$PATCHER" \
    "$CASE_DIR/$ARCHIVE_NAME" \
    "$CASE_DIR/checksums.txt" \
    "$CASE_DIR/replicator.rb" >/dev/null
}

assert_failure_preserves_cask() {
  cp "$CASE_DIR/replicator.rb" "$CASE_DIR/before.rb"
  if "$PATCHER" \
    "$CASE_DIR/$ARCHIVE_NAME" \
    "$CASE_DIR/checksums.txt" \
    "$CASE_DIR/replicator.rb" >"$CASE_DIR/stdout" 2>"$CASE_DIR/stderr"; then
    fail "$1: expected failure"
  fi
  grep -q '^::error::' "$CASE_DIR/stderr" || \
    fail "$1: failure did not emit an ::error:: annotation"
  cmp -s "$CASE_DIR/before.rb" "$CASE_DIR/replicator.rb" || \
    fail "$1: original cask changed on failure"
}

new_case happy
assert_success
sed "7c\\      sha256 \"$ARCHIVE_SHA\"" "$FIXTURE" > "$CASE_DIR/expected.rb"
cmp -s "$CASE_DIR/expected.rb" "$CASE_DIR/replicator.rb" || \
  fail "happy: patched cask differs from exact expected fixture"
grep -q '#{staged_path}' "$CASE_DIR/replicator.rb" || \
  fail "happy: patched cask is missing Ruby interpolation #{staged_path}"

new_case stray-comment
printf '\n# note: darwin_arm64 builds are notarized\n' >> "$CASE_DIR/replicator.rb"
assert_success
grep -q "sha256 \"$LINUX_ARM64_SHA\"" "$CASE_DIR/replicator.rb" || \
  fail "stray-comment: linux_arm64 SHA changed"

new_case trailing-comment
sed -i '/linux_arm64.tar.gz"$/s/$/ # darwin_arm64/' "$CASE_DIR/replicator.rb"
assert_success
grep -q "sha256 \"$LINUX_ARM64_SHA\"" "$CASE_DIR/replicator.rb" || \
  fail "trailing-comment: linux_arm64 SHA changed"

new_case missing-darwin
sed -i '/darwin_arm64/d' "$CASE_DIR/replicator.rb"
assert_failure_preserves_cask "missing darwin URL"

new_case duplicate-darwin
printf '  url "https://example.invalid/replicator_darwin_arm64.tar.gz"\n' >> \
  "$CASE_DIR/replicator.rb"
assert_failure_preserves_cask "duplicate darwin URL"

new_case reordered
awk '
  NR == 7 { sha = $0; next }
  NR == 8 { print; print sha; next }
  { print }
' "$CASE_DIR/replicator.rb" > "$CASE_DIR/reordered.rb"
mv "$CASE_DIR/reordered.rb" "$CASE_DIR/replicator.rb"
assert_failure_preserves_cask "URL before sha256"

new_case stale-candidate
cat > "$CASE_DIR/replicator.rb" <<'CASK'
cask "replicator" do
  sha256 "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  url "https://example.invalid/replicator_linux_arm64.tar.gz"
  url "https://example.invalid/replicator_darwin_arm64.tar.gz"
  sha256 "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
end
CASK
assert_failure_preserves_cask "stale checksum candidate"

new_case missing-manifest
: > "$CASE_DIR/checksums.txt"
assert_failure_preserves_cask "missing manifest entry"

new_case duplicate-manifest
cat "$CASE_DIR/checksums.txt" >> "$CASE_DIR/checksums.txt.copy"
cat "$CASE_DIR/checksums.txt" >> "$CASE_DIR/checksums.txt.copy"
mv "$CASE_DIR/checksums.txt.copy" "$CASE_DIR/checksums.txt"
assert_failure_preserves_cask "duplicate manifest entry"

new_case mismatched-manifest
printf '%064d  %s\n' 0 "$ARCHIVE_NAME" > "$CASE_DIR/checksums.txt"
assert_failure_preserves_cask "mismatched manifest entry"

# ---------------------------------------------------------------------------
# Rendered-cask regression: validate the custom_block in .goreleaser.yaml
# emits correct Homebrew postflight_steps DSL without running goreleaser.
# ---------------------------------------------------------------------------
GORELEASER="$SCRIPT_DIR/../../.goreleaser.yaml"
CI_WORKFLOW="$SCRIPT_DIR/../../.github/workflows/ci.yml"
if [ ! -f "$GORELEASER" ]; then
  fail "rendered-cask: .goreleaser.yaml not found at $GORELEASER"
fi
if [ ! -f "$CI_WORKFLOW" ]; then
  fail "rendered-cask: CI workflow not found at $CI_WORKFLOW"
fi

# goreleaser-action must use install-only mode so the binary is available
# on PATH for subsequent steps (the snapshot render below).
awk '
  /uses: goreleaser\/goreleaser-action@/ { in_action = 1; next }
  in_action && /install-only: true/ { found = 1 }
  in_action && /^[[:space:]]*-[[:space:]]/ { in_action = 0 }
  END { exit !found }
' "$CI_WORKFLOW" || fail "rendered-cask: GoReleaser action must set install-only: true"

# Extract the custom_block literal block scalar value from .goreleaser.yaml.
# The block starts on the line after "custom_block: |" and continues while
# lines are indented deeper than the key's column.
CUSTOM_BLOCK=$(awk '
  /^[[:space:]]*custom_block:[[:space:]]*\|/ {
    # Determine the indentation of the key itself
    match($0, /^[[:space:]]*/); key_indent = RLENGTH
    capturing = 1
    next
  }
  capturing {
    # Lines in the block must be indented more than the key
    match($0, /^[[:space:]]*/); line_indent = RLENGTH
    if (line_indent > key_indent || $0 ~ /^[[:space:]]*$/) {
      print
    } else {
      exit
    }
  }
' "$GORELEASER")

assert_current_postflight_dsl() {
  local cask=$1
  local source=$2

  [ -s "$cask" ] || fail "$source: cask is missing or empty"
  grep -q 'postflight_steps do' "$cask" || \
    fail "$source: cask is missing 'postflight_steps do' (new Homebrew DSL)"
  if grep -Eq '^[[:space:]]*postflight[[:space:]]+do' "$cask"; then
    fail "$source: cask contains legacy 'postflight do' (should be 'postflight_steps do')"
  fi
  # postflight_steps takes a literal {{staged_path}} placeholder; Ruby
  # interpolation there makes Homebrew 7 reject the whole cask as
  # unreadable. (In .goreleaser.yaml it is escaped as
  # {{ `{{staged_path}}` }}, which contains the same text.)
  grep -qF '{{staged_path}}' "$cask" || \
    fail "$source: cask is missing the postflight_steps placeholder '{{staged_path}}'"
  if grep -qF '#{staged_path}' "$cask"; then
    fail "$source: cask uses Ruby interpolation '#{staged_path}', which Homebrew 7 can't read in postflight_steps"
  fi
}

CONFIG_CASK="$WORK/custom-block.rb"
printf '%s\n' "$CUSTOM_BLOCK" > "$CONFIG_CASK"
assert_current_postflight_dsl "$CONFIG_CASK" "goreleaser custom_block"

if [ -n "$RENDERED_CASK" ]; then
  assert_current_postflight_dsl "$RENDERED_CASK" "rendered cask"
fi

echo "PASS: Homebrew cask integrity regression suite"
