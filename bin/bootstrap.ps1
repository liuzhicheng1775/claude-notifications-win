# Claude Notifications Win - Bootstrap Script
# Downloads/updates the notify.exe binary from GitHub Releases

param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$REPO_OWNER = "liuzhicheng1775"
$REPO_NAME = "claude-notifications-win"
$API_URL = "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
$BINARY_NAME = "notify.exe"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BinaryPath = Join-Path $ScriptDir $BINARY_NAME

function Get-LatestRelease {
    try {
        $response = Invoke-RestMethod -Uri $API_URL -UseBasicParsing
        return @{
            TagName = $response.tag_name
            Assets = $response.assets
        }
    } catch {
        Write-Host "[claude-notifications-win] ERROR: Failed to fetch latest release: $_" >&2
        return $null
    }
}

function Get-DownloadUrl {
    param($Assets, [string]$Version)

    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $downloadName = "notify-windows-$arch.exe"

    foreach ($asset in $Assets) {
        if ($asset.name -eq $downloadName) {
            return $asset.browser_download_url
        }
    }

    # Fallback: construct URL
    return "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$Version/$downloadName"
}

function Install-Binary {
    param([string]$Url, [string]$Destination)

    Write-Host "[claude-notifications-win] Downloading latest release..."
    try {
        Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing
        Write-Host "[claude-notifications-win] Binary installed at: $Destination"
    } catch {
        Write-Host "[claude-notifications-win] ERROR: Failed to download binary: $_" >&2
        throw
    }
}

function Get-CurrentVersion {
    if (Test-Path $BinaryPath) {
        try {
            $version = & $BinaryPath version 2>$null
            return $version
        } catch {
            return $null
        }
    }
    return $null
}

# Main
try {
    $release = Get-LatestRelease
    if (-not $release) {
        Write-Host "[claude-notifications-win] Skipping update (could not fetch release info)"
        exit 0
    }

    $version = $release.TagName -replace "^v", ""
    $currentVersion = Get-CurrentVersion

    if ($currentVersion -eq $version -and -not $Force) {
        Write-Host "[claude-notifications-win] Already up to date (v$version)"
        exit 0
    }

    $downloadUrl = Get-DownloadUrl -Assets $release.Assets -Version $release.TagName
    Install-Binary -Url $downloadUrl -Destination $BinaryPath

} catch {
    Write-Host "[claude-notifications-win] Bootstrap failed: $_" >&2
    exit 1
}
