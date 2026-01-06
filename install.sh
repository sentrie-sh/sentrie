#!/bin/sh
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

# TODO(everyone): Keep this script simple and easily auditable.

set -e

if [ "$OS" = "Windows_NT" ]; then
	echo "Error: This installer is for macOS, Linux, and WSL2. Windows is not supported. Use install.ps1 instead." 1>&2
	exit 1
else
	case $(uname -sm) in
	"Darwin x86_64") target="darwin_amd64.tar.gz" ;;
	"Darwin arm64") target="darwin_arm64.tar.gz" ;;
	"Linux x86_64") target="linux_amd64.tar.gz" ;;
	"Linux aarch64") target="linux_arm64.tar.gz" ;;
	*) echo "Error: '$(uname -sm)' is not supported yet." 1>&2; exit 1 ;;
	esac
fi

if ! command -v tar >/dev/null; then
	echo "Error: 'tar' is required to install Sentrie." 1>&2
	exit 1
fi

if ! command -v gzip >/dev/null; then
	echo "Error: 'gzip' is required to install Sentrie." 1>&2
	exit 1
fi

if ! command -v install >/dev/null; then
	echo "Error: 'install' is required to install Sentrie." 1>&2
	exit 1
fi

# Utility function to download a file from a URL to a local file
function download_file() {
	local url="$1"
	local output="$2"
	if command -v wget >/dev/null; then
		wget --help | grep -q '\--showprogress' && _FORCE_PROGRESS_BAR="--no-verbose --show-progress" || _FORCE_PROGRESS_BAR=""
		if ! wget --prefer-family=IPv4 --progress=bar:force:noscroll $_FORCE_PROGRESS_BAR -O "$output" "$url"; then
			echo "Could not download $url" 1>&2
			exit 1
		fi
	elif command -v curl >/dev/null; then
		if ! curl --fail --location --progress-bar --output "$output" "$url"; then
			echo "Could not download $url" 1>&2
			exit 1
		fi
	fi
  
  echo "Downloaded $url to $output"
}

test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
mkdir -p "${tmp_dir}"
tmp_dir="${tmp_dir%/}"

echo "Created temporary directory at $tmp_dir."
echo "Changing to $tmp_dir"
cd "$tmp_dir"

if [ $# -eq 0 ]; then
  # If no version is provided, get the latest version
	echo "Determining latest version"
  download_file "https://api.github.com/repos/sentrie-sh/sentrie/releases/latest" "$tmp_dir/latest-release.json"
	version=$(grep '"tag_name"' "$tmp_dir/latest-release.json" | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
	if [ -z "$version" ] || [ "$version" = "null" ]; then
		echo "Error: Could not determine latest version" 1>&2
		exit 1
	fi
else
	version="$1"
fi

version_no_v=$(echo "$version" | sed 's/^v//')
sentrie_uri="https://github.com/sentrie-sh/sentrie/releases/download/${version}/sentrie_${version_no_v}_${target}"
checksums_uri="https://github.com/sentrie-sh/sentrie/releases/download/${version}/checksums.txt"
signature_uri="https://github.com/sentrie-sh/sentrie/releases/download/${version}/sentrie_${version_no_v}_${target}.signature.bundle.json"

echo "Detected version: $version"
echo "Detected target: $target"

# Check if sentrie is already installed and use that location if writable
bin_dir=""
if command -v sentrie >/dev/null 2>&1; then
	existing_path=$(command -v sentrie)
	existing_dir=$(dirname "$existing_path")
	if [ -w "$existing_dir" ] 2>/dev/null; then
		bin_dir="$existing_dir"
	fi
fi

# If not found, use default location
if [ -z "$bin_dir" ]; then
	sentrie_install="${SENTRIE_INSTALL:-$HOME/.local}"
	bin_dir="$sentrie_install/bin"
	# Create directory if it doesn't exist
	mkdir -p "$bin_dir"
fi

exe="$bin_dir/sentrie"

# set a trap for a clean exit - even in failures
trap 'rm -rf $tmp_dir' EXIT

archive_location="$tmp_dir/sentrie.tar.gz"
checksums_location="$tmp_dir/checksums.txt"
signature_location="$tmp_dir/sentrie_signature.bundle.json"

echo "Downloading from $sentrie_uri"
download_file "$sentrie_uri" "$archive_location"
download_file "$checksums_uri" "$checksums_location"
download_file "$signature_uri" "$signature_location"

echo "Verifying checksum"
archive_name=$(basename "$sentrie_uri")
expected_hash=$(grep "$archive_name" "$checksums_location" | awk '{print $1}')
if [ -z "$expected_hash" ]; then
	echo "Error: Checksum not found for $archive_name" 1>&2
	exit 1
fi

if command -v sha256sum >/dev/null; then
	actual_hash=$(sha256sum "$archive_location" | awk '{print $1}')
elif command -v shasum >/dev/null; then
	actual_hash=$(shasum -a 256 "$archive_location" | awk '{print $1}')
else
	echo "Error: No SHA256 checksum tool available" 1>&2
	exit 1
fi

if [ "$expected_hash" != "$actual_hash" ]; then
	echo "Error: Checksum verification failed" 1>&2
	echo "Expected: $expected_hash" 1>&2
	echo "Actual:   $actual_hash" 1>&2
	exit 1
fi

echo "Checksum verification successful"

if command -v cosign >/dev/null; then
  echo "Downloading archive signature bundle"
  echo "Verifying artifact signature"
  if ! cosign verify-blob --bundle "$signature_location" "$archive_location" --certificate-identity="https://github.com/sentrie-sh/sentrie/.github/workflows/release.yml@refs/tags/${version}" --certificate-oidc-issuer="https://token.actions.githubusercontent.com"; then
    echo "Error: Artifact signature verification failed" 1>&2
    exit 1
  fi
fi

echo "Deflating downloaded archive"
tar -xf "$archive_location" -C "$tmp_dir"

echo "Installing"
cp "$tmp_dir/sentrie" "$exe"

echo "Applying necessary permissions"
chmod +x "$exe"

echo "Removing downloaded archive"
rm "$archive_location"
rm "$checksums_location"

echo "Sentrie was installed successfully to $exe"

if ! command -v sentrie >/dev/null; then
	echo ""
	echo "Note: 'sentrie' is not in your PATH."
	echo "Add it to your PATH by running:"
	echo "  export PATH=\"\$PATH:$bin_dir\""
	echo "Or add it permanently to your shell profile (~/.bashrc, ~/.zshrc, etc.)"
	echo ""
	echo "You can also run Sentrie directly:"
	echo "  $exe"
	echo ""
fi

# Verify the binary can be executed
if ! "$exe" --version >/dev/null 2>&1; then
	echo "Sentrie was installed, but could not be executed. Are you sure '$exe' has the necessary permissions?" 1>&2
	exit 1
fi

rm -rf "$tmp_dir"