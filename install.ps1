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

$bin_dir = "$env:LOCALAPPDATA\sentrie\bin"
$exe = "$bin_dir\sentrie.exe"

$tmp_dir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

Write-Host "Created temporary directory at $tmp_dir. Changing to $tmp_dir"
Set-Location "$tmp_dir"

$archive_location = "$tmp_dir\sentrie.zip"
$checksums_location = "$tmp_dir\checksums.txt"

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

	Write-Host "Verifying checksum"
	$archive_name = Split-Path $sentrie_uri -Leaf
	$checksums_content = Get-Content $checksums_location
	$expected_hash = ""
	foreach ($line in $checksums_content) {
		if ($line -match "^([a-f0-9]{64})\s+$([^\s]+)$") {
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

	if (Get-Command cosign -ErrorAction SilentlyContinue) {
		Write-Host "Verifying archive attestation"
		$attestation_bundle_uri = "https://github.com/sentrie-sh/sentrie/releases/download/${Version}/${archive_name}.attestation.bundle"
		$attestation_bundle_location = "$tmp_dir\${archive_name}.attestation.bundle"
		
		try {
			Invoke-WebRequest -Uri $attestation_bundle_uri -OutFile $attestation_bundle_location -UseBasicParsing -ErrorAction SilentlyContinue
			if (Test-Path $attestation_bundle_location) {
				$verify_result = & cosign verify-blob --bundle $attestation_bundle_location $archive_location 2>&1
				if ($LASTEXITCODE -ne 0) {
					Write-Host "Error: Archive attestation verification failed" -ForegroundColor Red
					exit 1
				}
				Remove-Item $attestation_bundle_location
			}
		} catch {
			# Bundle not available, skip Cosign verification
		}
	}

	Write-Host "Deflating downloaded archive"
	$extract_path = "$tmp_dir\extract"
	Expand-Archive -Path $archive_location -DestinationPath $extract_path -Force

	if (Get-Command cosign -ErrorAction SilentlyContinue) {
		Write-Host "Verifying binary signature"
		$binary_bundle_location = "$extract_path\sentrie.bundle"
		if (Test-Path $binary_bundle_location) {
			$verify_result = & cosign verify-blob --bundle $binary_bundle_location "$extract_path\sentrie.exe" 2>&1
			if ($LASTEXITCODE -ne 0) {
				Write-Host "Error: Binary signature verification failed" -ForegroundColor Red
				exit 1
			}
		}
	}

	Write-Host "Installing"
	if (!(Test-Path $bin_dir)) {
		New-Item -ItemType Directory -Path $bin_dir -Force | Out-Null
	}
	Copy-Item -Path "$extract_path\sentrie.exe" -Destination $exe -Force

	Write-Host "Removing downloaded archive"
	Remove-Item $archive_location
	Remove-Item $checksums_location

	Write-Host "Sentrie was installed successfully to $exe"

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
