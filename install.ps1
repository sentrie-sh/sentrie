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

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -eq "AMD64" -or $arch -eq "x86_64") {
	$target = "windows_amd64.zip"
} elseif ($arch -eq "ARM64") {
	$target = "windows_arm64.zip"
} else {
	Write-Host "Error: Architecture '$arch' is not supported yet." -ForegroundColor Red
	exit 1
}

if ($Version -eq "") {
	$version_response = Invoke-RestMethod -Uri "https://api.github.com/repos/sentrie-sh/sentrie/releases/latest" -Method Get
	$Version = $version_response.tag_name
	if ($Version -eq "") {
		Write-Host "Error: Could not determine latest version" -ForegroundColor Red
		exit 1
	}
}

$version_no_v = $Version -replace '^v', ''
$sentrie_uri = "https://github.com/sentrie-sh/sentrie/releases/download/${Version}/sentrie_${version_no_v}_${target}"
$checksums_uri = "https://github.com/sentrie-sh/sentrie/releases/download/${Version}/checksums.txt"
$signature_uri = "https://github.com/sentrie-sh/sentrie/releases/download/${Version}/sentrie_${version_no_v}_${target}.signature.bundle.json"

Write-Host "Detected version: $Version"
Write-Host "Detected target: $target"

# Check if sentrie is already installed and use that location if writable
$bin_dir = ""
if (Get-Command sentrie -ErrorAction SilentlyContinue) {
	$existing_path = (Get-Command sentrie).Source
	$existing_dir = Split-Path $existing_path -Parent
	if (Test-Path $existing_dir -PathType Container) {
		try {
			$test_file = Join-Path $existing_dir ".test_write"
			[System.IO.File]::WriteAllText($test_file, "test")
			Remove-Item $test_file -ErrorAction SilentlyContinue
			$bin_dir = $existing_dir
		} catch {
			# Directory not writable, continue
		}
	}
}

# If not found, use default location
if ($bin_dir -eq "") {
	$sentrie_install = if ($env:SENTRIE_INSTALL) { $env:SENTRIE_INSTALL } else { "$env:LOCALAPPDATA\sentrie" }
	$bin_dir = "$sentrie_install\bin"
	# Create directory if it doesn't exist
	if (!(Test-Path $bin_dir)) {
		New-Item -ItemType Directory -Path $bin_dir -Force | Out-Null
	}
}

$exe = "$bin_dir\sentrie.exe"

$tmp_dir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

Write-Host "Created temporary directory at $tmp_dir. Changing to $tmp_dir"
Set-Location "$tmp_dir"

$archive_location = "$tmp_dir\sentrie.zip"
$checksums_location = "$tmp_dir\checksums.txt"
$signature_location = "$tmp_dir\sentrie_signature.bundle.json"

try {
	Write-Host "Downloading from $sentrie_uri"
	$ProgressPreference = 'SilentlyContinue'
	try {
		Invoke-WebRequest -Uri $sentrie_uri -OutFile $archive_location -UseBasicParsing
	} catch {
		Write-Host "Could not find version $Version" -ForegroundColor Red
		exit 1
	}

	Write-Host "Downloading checksums"
	if (!(Invoke-WebRequest -Uri $checksums_uri -OutFile $checksums_location -UseBasicParsing)) {
		Write-Host "Could not download checksums" -ForegroundColor Red
		exit 1
	}

	Write-Host "Downloading signature"
	try {
		Invoke-WebRequest -Uri $signature_uri -OutFile $signature_location -UseBasicParsing -ErrorAction Stop
	} catch {
		Write-Host "Could not download signature" -ForegroundColor Red
		exit 1
	}

	Write-Host "Verifying checksum"
	$archive_name = Split-Path $sentrie_uri -Leaf
	$checksums_content = Get-Content $checksums_location
	$expected_hash = ""
	foreach ($line in $checksums_content) {
    if ($line -match '^([a-f0-9]{64})\s+\*?(\S+)$') {
			$hash = $matches[1]
			$file = $matches[2]
			if ($file -eq $archive_name) {
				$expected_hash = $hash
				break
			}
		}
	}

	if ($expected_hash -eq "") {
		Write-Host "Error: Checksum not found for $archive_name" -ForegroundColor Red
		exit 1
	}

	$actual_hash = (Get-FileHash -Path $archive_location -Algorithm SHA256).Hash.ToLower()

	if ($expected_hash -ne $actual_hash) {
		Write-Host "Error: Checksum verification failed" -ForegroundColor Red
		Write-Host "Expected: $expected_hash" -ForegroundColor Red
		Write-Host "Actual:   $actual_hash" -ForegroundColor Red
		exit 1
	}

	Write-Host "Checksum verification successful"

	if (Get-Command cosign -ErrorAction SilentlyContinue) {
		Write-Host "Verifying artifact signature"
		$verify_result = & cosign verify-blob --bundle $signature_location $archive_location --certificate-identity="https://github.com/sentrie-sh/sentrie/.github/workflows/release.yml@refs/tags/${Version}" --certificate-oidc-issuer="https://token.actions.githubusercontent.com" 2>&1
		if ($LASTEXITCODE -ne 0) {
			Write-Host "Error: Artifact signature verification failed" -ForegroundColor Red
			exit 1
		}
	}

	Write-Host "Deflating downloaded archive"
	$extract_path = "$tmp_dir\extract"
	Expand-Archive -Path $archive_location -DestinationPath $extract_path -Force

	Write-Host "Installing"
	if (!(Test-Path $bin_dir)) {
		New-Item -ItemType Directory -Path $bin_dir -Force | Out-Null
	}
	Copy-Item -Path "$extract_path\sentrie.exe" -Destination $exe -Force

	Write-Host "Removing downloaded archive"
	Remove-Item $archive_location
	Remove-Item $checksums_location
	Remove-Item $signature_location -ErrorAction SilentlyContinue

	Write-Host "Sentrie was installed successfully to $exe"

	if (!(Get-Command sentrie -ErrorAction SilentlyContinue)) {
		Write-Host ""
		Write-Host "Note: 'sentrie' is not in your PATH."
		Write-Host "Add it to your PATH by running:"
		Write-Host "  `$env:PATH += `";$bin_dir`""
		Write-Host "Or add it permanently to your user environment variables."
		Write-Host ""
		Write-Host "You can also run Sentrie directly:"
		Write-Host "  $exe"
		Write-Host ""
	}

	# Verify the binary can be executed
	try {
		$null = & $exe --version 2>&1
	} catch {
		Write-Host "Sentrie was installed, but could not be executed. Are you sure '$exe' has the necessary permissions?" -ForegroundColor Red
		exit 1
	}
} finally {
	Remove-Item -Path $tmp_dir -Recurse -Force -ErrorAction SilentlyContinue
}
