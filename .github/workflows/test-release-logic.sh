#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
#
# Copyright 2025 Binaek Sarkar
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# Test helper functions
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

PASSED=0
FAILED=0

test_case() {
  local name="$1"
  shift
  local test_cmd="$@"
  
  echo -n "Testing: $name ... "
  if eval "$test_cmd" 2>&1; then
    echo "✓ PASSED"
    ((PASSED++)) || true
    return 0
  else
    local exit_code=$?
    echo "✗ FAILED (exit code: $exit_code)"
    ((FAILED++)) || true
    return 0
  fi
}

# Test version parsing logic (from workflow)
test_version_parsing() {
  local tag="$1"
  local expected_full_us="$2"
  
  local v="${tag#v}"
  local full_us=$(printf '%s' "$v" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/_/g; s/^_+|_+$//g')
  
  if [ "$full_us" = "$expected_full_us" ]; then
    return 0
  else
    echo "  Expected: '$expected_full_us', Got: '$full_us'"
    return 1
  fi
}

# Test file existence check logic
test_file_existence_check() {
  local tag="$1"
  local should_exist="$2"  # "true" or "false"
  
  local v="${tag#v}"
  local full_us=$(printf '%s' "$v" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/_/g; s/^_+|_+$//g')
  local cask_file="Casks/sentrie@${full_us}.rb"
  local formula_file="Formula/sentrie.rb"
  
  rm -rf "$TEST_DIR"/* 2>/dev/null || true
  mkdir -p "$TEST_DIR/Casks" "$TEST_DIR/Formula"
  
  if [ "$should_exist" = "true" ]; then
    touch "$TEST_DIR/$cask_file"
  fi
  
  (cd "$TEST_DIR" && if [ -f "$cask_file" ] || [ -f "$formula_file" ]; then
    [ "$should_exist" = "true" ] && return 0 || return 1
  else
    [ "$should_exist" = "false" ] && return 0 || return 1
  fi)
}

# Test prerelease validation logic
test_prerelease_validation() {
  local prerelease="$1"
  shift
  local cask_files_str="$1"
  shift
  local formula_files_str="$1"
  shift
  local should_fail="$1"  # "true" if validation should fail
  
  rm -rf "$TEST_DIR"/* 2>/dev/null || true
  mkdir -p "$TEST_DIR/Casks" "$TEST_DIR/Formula"
  
  # Create test files
  IFS=' ' read -ra cask_files <<< "$cask_files_str"
  for file in "${cask_files[@]}"; do
    [ -n "$file" ] && touch "$TEST_DIR/Casks/$file"
  done
  IFS=' ' read -ra formula_files <<< "$formula_files_str"
  for file in "${formula_files[@]}"; do
    [ -n "$file" ] && touch "$TEST_DIR/Formula/$file"
  done
  
  # Simulate git add - but for testing failures, we need to actually add the wrong files
  (cd "$TEST_DIR" && rm -rf .git 2>/dev/null || true && \
    git init -q && \
    git config user.name "Test" && \
    git config user.email "test@test.com" && \
    if [ "$prerelease" = "true" ]; then
      # For prerelease, only add versioned files (correct behavior)
      # But if should_fail is true, we're testing the case where default files exist
      # In that case, we need to simulate what would happen if they were accidentally added
      if [ "$should_fail" = "true" ] && echo "$cask_files_str" | grep -q "sentrie.rb"; then
        # Simulate the error case: default files were added
        git add Casks/sentrie.rb Casks/sentrie@*.rb 2>/dev/null || true
        git add Formula/sentrie.rb Formula/sentrie@*.rb 2>/dev/null || true
      else
        git add Casks/sentrie@*.rb 2>/dev/null || true
        git add Formula/sentrie@*.rb 2>/dev/null || true
      fi
    else
      # For stable, add all files
      # But if should_fail is true and count is wrong, only add the wrong number
      if [ "$should_fail" = "true" ]; then
        # Count how many files we have
        local cask_file_count=$(echo "$cask_files_str" | wc -w | tr -d ' ')
        if [ "$cask_file_count" -lt 3 ]; then
          # Only add the files that exist (less than 3)
          git add Casks/*.rb 2>/dev/null || true
          git add Formula/*.rb 2>/dev/null || true
        else
          # More than 4 - add all
          git add Casks/*.rb 2>/dev/null || true
          git add Formula/*.rb 2>/dev/null || true
        fi
      else
        git add Casks/sentrie.rb Casks/sentrie@*.rb 2>/dev/null || true
        git add Formula/sentrie.rb Formula/sentrie@*.rb 2>/dev/null || true
      fi
    fi && \
    git diff --cached --name-only | grep -E '^(Casks|Formula)/' || true) > "$TEST_DIR/git_output.txt"
  
  local changed=$(cat "$TEST_DIR/git_output.txt")
  local cask_changed=$(echo "$changed" | grep '^Casks/' || true)
  local formula_changed=$(echo "$changed" | grep '^Formula/' || true)
  local cask_count=$(printf '%s\n' "$cask_changed" | sed '/^$/d' | wc -l | tr -d ' ')
  local formula_count=$(printf '%s\n' "$formula_changed" | sed '/^$/d' | wc -l | tr -d ' ')
  
  local validation_passed=true
  
  if [ "$prerelease" = "true" ]; then
    # Check default files should not exist
    if echo "$cask_changed" | grep -Fxq 'Casks/sentrie.rb'; then
      validation_passed=false
    fi
    if echo "$formula_changed" | grep -Fxq 'Formula/sentrie.rb'; then
      validation_passed=false
    fi
    # Should have exactly 1 of each
    if [ "$cask_count" -ne 1 ] || [ "$formula_count" -ne 1 ]; then
      validation_passed=false
    fi
  else
    # Stable should have 3-4 of each
    if [ "$cask_count" -lt 3 ] || [ "$cask_count" -gt 4 ]; then
      validation_passed=false
    fi
    if [ "$formula_count" -lt 3 ] || [ "$formula_count" -gt 4 ]; then
      validation_passed=false
    fi
  fi
  
  # If validation should fail, then validation_passed should be false
  # If validation should pass, then validation_passed should be true
  if [ "$should_fail" = "true" ]; then
    [ "$validation_passed" = "false" ] && return 0 || return 1
  else
    [ "$validation_passed" = "true" ] && return 0 || return 1
  fi
}

# Test artifact file finding logic
test_artifact_finding() {
  local structure="$1"  # "flat", "nested", "deep", "none"
  local should_find="$2"  # "true" or "false"
  
  rm -rf "$TEST_DIR"/* 2>/dev/null || true
  mkdir -p "$TEST_DIR/tmp-cask"
  
  case "$structure" in
    "flat")
      touch "$TEST_DIR/tmp-cask/sentrie.rb.tmpl"
      ;;
    "nested")
      mkdir -p "$TEST_DIR/tmp-cask/Casks"
      touch "$TEST_DIR/tmp-cask/Casks/sentrie.rb.tmpl"
      ;;
    "deep")
      mkdir -p "$TEST_DIR/tmp-cask/homebrew/Casks"
      touch "$TEST_DIR/tmp-cask/homebrew/Casks/sentrie.rb.tmpl"
      ;;
    "none")
      # Don't create any file
      ;;
  esac
  
  local found=$(cd "$TEST_DIR" && find tmp-cask -type f -name "sentrie.rb.tmpl" 2>/dev/null | head -1)
  
  if [ -n "$found" ]; then
    [ "$should_find" = "true" ] && return 0 || return 1
  else
    [ "$should_find" = "false" ] && return 0 || return 1
  fi
}

echo "=== Testing Release Workflow Logic ==="
echo ""

# Test version parsing
echo "--- Version Parsing Tests ---"
test_case "Simple version v1.2.3" \
  "test_version_parsing 'v1.2.3' '1_2_3'"

test_case "Prerelease version v1.2.3-alpha.1" \
  "test_version_parsing 'v1.2.3-alpha.1' '1_2_3_alpha_1'"

test_case "Version with plus v0.0.2+git.abc" \
  "test_version_parsing 'v0.0.2+git.abc' '0_0_2_git_abc'"

test_case "Version without v prefix" \
  "test_version_parsing '2.5.8' '2_5_8'"

test_case "Version with underscores" \
  "test_version_parsing 'v1.0.0-test_release' '1_0_0_test_release'"

test_case "Version with mixed case" \
  "test_version_parsing 'v1.2.3-ALPHA.Beta' '1_2_3_alpha_beta'"

# Test file existence checks
echo ""
echo "--- File Existence Check Tests ---"
test_case "File should not exist (new release)" \
  "test_file_existence_check 'v1.2.3' 'false'"

test_case "File should exist (rerun prevention)" \
  "test_file_existence_check 'v1.2.3' 'true'"

# Test prerelease validation
echo ""
echo "--- Prerelease Validation Tests ---"
test_case "Prerelease with correct files (1 versioned each)" \
  "test_prerelease_validation 'true' 'sentrie@1_2_3_alpha.rb' 'sentrie@1_2_3_alpha.rb' 'false'"

test_case "Prerelease with default files (should fail)" \
  "test_prerelease_validation 'true' 'sentrie.rb sentrie@1_2_3_alpha.rb' 'sentrie.rb sentrie@1_2_3_alpha.rb' 'true'"

test_case "Prerelease with wrong count (should fail)" \
  "test_prerelease_validation 'true' 'sentrie@1_2_3_alpha.rb sentrie@1_2_3_beta.rb' 'sentrie@1_2_3_alpha.rb' 'true'"

test_case "Stable release with correct files (3-4 versioned each)" \
  "test_prerelease_validation 'false' 'sentrie.rb sentrie@1_2_3.rb sentrie@1_2.rb sentrie@1.rb' 'sentrie.rb sentrie@1_2_3.rb sentrie@1_2.rb sentrie@1.rb' 'false'"

test_case "Stable release with only 2 versioned (should fail)" \
  "test_prerelease_validation 'false' 'sentrie.rb sentrie@1_2_3.rb' 'sentrie.rb sentrie@1_2_3.rb' 'true'"

test_case "Stable release with 5 versioned (should fail)" \
  "test_prerelease_validation 'false' 'sentrie.rb sentrie@1_2_3.rb sentrie@1_2.rb sentrie@1.rb sentrie@2.rb sentrie@3.rb' 'sentrie.rb sentrie@1_2_3.rb sentrie@1_2.rb sentrie@1.rb' 'true'"

# Test artifact finding
echo ""
echo "--- Artifact Finding Tests ---"
test_case "Find artifact in flat structure" \
  "test_artifact_finding 'flat' 'true'"

test_case "Find artifact in nested structure" \
  "test_artifact_finding 'nested' 'true'"

test_case "Find artifact in deep structure" \
  "test_artifact_finding 'deep' 'true'"

test_case "No artifact found" \
  "test_artifact_finding 'none' 'false'"

# Test edge cases
echo ""
echo "--- Edge Case Tests ---"

# Test version with special characters
test_case "Version with dots, dashes, and plus" \
  "test_version_parsing 'v1.2.3-alpha.1+git.abc123' '1_2_3_alpha_1_git_abc123'"

# Test empty version (should be caught by validation)
test_case "Empty version string" \
  "test_version_parsing '' ''"

# Test very long version
test_case "Very long prerelease version" \
  "test_version_parsing 'v1.0.0-rc.1.build.12345.commit.abcdef' '1_0_0_rc_1_build_12345_commit_abcdef'"

# Summary
echo ""
echo "=== Test Results ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Total: $((PASSED + FAILED))"

if [ $FAILED -gt 0 ]; then
  exit 1
fi

exit 0
