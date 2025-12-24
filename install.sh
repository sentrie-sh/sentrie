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

if [ "$OS" = "Windows_NT" ]; then
	echo "Error: Windows is not supported. Use install.ps1 instead." 1>&2
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

if [ $# -eq 0 ]; then
	version=$(curl -sSL https://api.github.com/repos/sentrie-sh/sentrie/releases/latest | grep -oP '"tag_name":\s*"\K[^"]+' | head -1)
	if [ -z "$version" ]; then
		echo "Error: Could not determine latest version" 1>&2
		exit 1
	fi
else
	version="$1"
fi

version_no_v=$(echo "$version" | sed 's/^v//')
sentrie_uri="https://github.com/sentrie-sh/sentrie/releases/download/${version}/sentrie_${version_no_v}_${target}"
checksums_uri="https://github.com/sentrie-sh/sentrie/releases/download/${version}/checksums.txt"

bin_dir="/usr/local/bin"
exe="$bin_dir/sentrie"

test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
mkdir -p "${tmp_dir}"
tmp_dir="${tmp_dir%/}"

echo "Created temporary directory at $tmp_dir. Changing to $tmp_dir"
cd "$tmp_dir"

# set a trap for a clean exit - even in failures
trap 'rm -rf $tmp_dir' EXIT

archive_location="$tmp_dir/sentrie.tar.gz"
checksums_location="$tmp_dir/checksums.txt"

echo "Downloading from $sentrie_uri"
if command -v wget >/dev/null; then
	# because --show-progress was introduced in 1.16.
	wget --help | grep -q '\--showprogress' && _FORCE_PROGRESS_BAR="--no-verbose --show-progress" || _FORCE_PROGRESS_BAR=""
	# prefer an IPv4 connection, since github.com does not handle IPv6 connections properly.
	if ! wget --prefer-family=IPv4 --progress=bar:force:noscroll $_FORCE_PROGRESS_BAR -O "$archive_location" "$sentrie_uri"; then
		echo "Could not find version $version" 1>&2
		exit 1
	fi
	if ! wget --prefer-family=IPv4 --progress=bar:force:noscroll $_FORCE_PROGRESS_BAR -O "$checksums_location" "$checksums_uri"; then
		echo "Could not download checksums" 1>&2
		exit 1
	fi
elif command -v curl >/dev/null; then
	# curl uses HappyEyeball for connections, therefore, no preference is required
	if ! curl --fail --location --progress-bar --output "$archive_location" "$sentrie_uri"; then
		echo "Could not find version $version" 1>&2
		exit 1
	fi
	if ! curl --fail --location --progress-bar --output "$checksums_location" "$checksums_uri"; then
		echo "Could not download checksums" 1>&2
		exit 1
	fi
else
	echo "Unable to find wget or curl. Cannot download." 1>&2
	exit 1
fi

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

if command -v cosign >/dev/null; then
	echo "Verifying Cosign signature"
	bundle_uri="https://github.com/sentrie-sh/sentrie/releases/download/${version}/${archive_name}.bundle"
	bundle_location="$tmp_dir/${archive_name}.bundle"
	
	if command -v wget >/dev/null; then
		wget --help | grep -q '\--showprogress' && _FORCE_PROGRESS_BAR="--no-verbose --show-progress" || _FORCE_PROGRESS_BAR=""
		if wget --prefer-family=IPv4 --progress=bar:force:noscroll $_FORCE_PROGRESS_BAR -O "$bundle_location" "$bundle_uri" 2>/dev/null; then
			if ! cosign verify-blob --bundle "$bundle_location" "$archive_location"; then
				echo "Error: Cosign signature verification failed" 1>&2
				exit 1
			fi
			rm "$bundle_location"
		fi
	elif command -v curl >/dev/null; then
		if curl --fail --location --progress-bar --output "$bundle_location" "$bundle_uri" 2>/dev/null; then
			if ! cosign verify-blob --bundle "$bundle_location" "$archive_location"; then
				echo "Error: Cosign signature verification failed" 1>&2
				exit 1
			fi
			rm "$bundle_location"
		fi
	fi
fi

echo "Deflating downloaded archive"
tar -xf "$archive_location" -C "$tmp_dir"

echo "Installing"
install -d "$bin_dir"
install "$tmp_dir/sentrie" "$bin_dir"

echo "Applying necessary permissions"
chmod +x "$exe"

echo "Removing downloaded archive"
rm "$archive_location"
rm "$checksums_location"

echo "Sentrie was installed successfully to $exe"

if ! command -v "$bin_dir/sentrie" >/dev/null; then
	echo "Sentrie was installed, but could not be executed. Are you sure '$bin_dir/sentrie' has the necessary permissions?" 1>&2
	exit 1
fi
