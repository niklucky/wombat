$ErrorActionPreference = "Stop"

$repo = "niklucky/wombat"
$binary = "wombat"

$os = "windows"
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
        exit 1
    }
}

$ext = "zip"
$assetName = "$binary-$os-$arch.$ext"

Write-Host "Detecting latest release..."

# Try to get the version from GitHub redirect (avoids API rate limits).
$version = $null
$redirectUrl = $null
try {
    $response = Invoke-WebRequest -Uri "https://github.com/$repo/releases/latest" -Method Head -UseBasicParsing -MaximumRedirection 0 -ErrorAction SilentlyContinue
} catch {
    $redirectUrl = $_.Exception.Response.Headers.Location.AbsoluteUri
}

if ($redirectUrl -and $redirectUrl -match '/tag/(v\d+\.\d+\.\d+)') {
    $version = $Matches[1]
}

# Fallback to GitHub API if redirect approach fails.
if (-not $version) {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -UseBasicParsing
    $version = $release.tag_name
}

if (-not $version) {
    Write-Error "Failed to determine latest release version."
    exit 1
}

Write-Host "Latest release: $version"

$assetUrl = "https://github.com/$repo/releases/download/$version/$assetName"

$tmpDir = Join-Path $env:TEMP "wombat-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    $downloadPath = Join-Path $tmpDir $assetName
    Write-Host "Downloading $assetName..."
    Invoke-WebRequest -Uri $assetUrl -OutFile $downloadPath -UseBasicParsing

    Write-Host "Extracting..."
    Expand-Archive -Path $downloadPath -DestinationPath $tmpDir -Force

    $installDir = Join-Path (Join-Path $env:LOCALAPPDATA "Programs") $binary
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    $binaryPath = Join-Path $tmpDir "$binary.exe"
    $destPath = Join-Path $installDir "$binary.exe"
    Move-Item -Path $binaryPath -Destination $destPath -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
        $env:Path = "$env:Path;$installDir"
        Write-Host "Added $installDir to your PATH."
    }

    Write-Host ""
    Write-Host "$binary $version installed to $destPath"
    & $destPath version
} finally {
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
