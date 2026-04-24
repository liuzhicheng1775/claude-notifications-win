# Claude Notifications Win - Bootstrap Script
# Downloads/updates the notify.exe binary from GitHub Releases

param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$BINARY_NAME = "notify.exe"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BinaryPath = Join-Path $ScriptDir $BINARY_NAME

function Get-LatestRelease {
    try {
        $apiUrl = "https://api.github.com/repos/liuzhicheng1775/claude-notifications-win/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        return @{
            TagName = $response.tag_name
            Assets = $response.assets
        }
    } catch {
        [Console]::Error.WriteLine("[claude-notifications-win] ERROR: Failed to fetch latest release: $_")
        return $null
    }
}

function Get-DownloadUrl {
    param($Assets, [string]$Version)

    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

    # amd64 uses "notify.exe", 386 uses "notify-windows-386.exe"
    $downloadName = if ($arch -eq "amd64") { "notify.exe" } else { "notify-windows-$arch.exe" }

    foreach ($asset in $Assets) {
        if ($asset.name -eq $downloadName) {
            return $asset.browser_download_url
        }
    }

    # Fallback: construct URL
    return "https://github.com/liuzhicheng1775/claude-notifications-win/releases/download/$Version/$downloadName"
}

function Install-Binary {
    param([string]$Url, [string]$Destination)

    Write-Host "[claude-notifications-win] Downloading latest release..."
    try {
        Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing
        Write-Host "[claude-notifications-win] Binary installed at: $Destination"
    } catch {
        [Console]::Error.WriteLine("[claude-notifications-win] ERROR: Failed to download binary: $_")
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
    [Console]::Error.WriteLine("[claude-notifications-win] Bootstrap failed: $_")
    exit 1
}
