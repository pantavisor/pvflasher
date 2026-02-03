# PvFlasher Windows Installation Script
param (
    [string]$Version
)

$ProjectName = "pvflasher"
$RepoUrl = "https://github.com/pantavisor/pvflasher"
$ApiUrl = "https://api.github.com/repos/pantavisor/pvflasher"
$InstallDir = Join-Path $env:LOCALAPPDATA "pvflasher"

# Detect Architecture
$Arch = "x86_64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
    $Arch = "aarch64"
}

Write-Host "Installing $ProjectName for $Arch..." -ForegroundColor Cyan

# 1. Get version if not specified
if (-not $Version) {
    Write-Host "Fetching latest release information..."
    $Release = Invoke-RestMethod -Uri "$ApiUrl/releases/latest"
    $Version = $Release.tag_name
}

if (-not $Version) {
    Write-Error "Could not find version. Please check $RepoUrl"
    return
}

Write-Host "Installing version: $Version"

# 2. Construct download URL for Windows zip
# Filename pattern: pvflasher-windows-x86_64.zip or pvflasher-windows-aarch64.zip
$ZipUrl = "$RepoUrl/releases/download/$Version/pvflasher-windows-$Arch.zip"
$ZipFile = Join-Path $env:TEMP "pvflasher.zip"

# 3. Download and Extract
Write-Host "Downloading $ZipUrl..."
try {
    Invoke-WebRequest -Uri $ZipUrl -OutFile $ZipFile -ErrorAction Stop
} catch {
    Write-Error "Failed to download $ZipUrl. Please verify the version and architecture."
    return
}

if (-not (Test-Path $InstallDir)) {
    New-Item -Path $InstallDir -ItemType Directory
}

Write-Host "Extracting to $InstallDir..."
Expand-Archive -Path $ZipFile -DestinationPath $InstallDir -Force
Remove-Item -Path $ZipFile

Write-Host ""
Write-Host "Done! Binary extracted to:" -ForegroundColor Green
Write-Host "  $InstallDir\pvflasher.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "Move it to your preferred location and add that directory to your PATH."
Write-Host "Note: Flashing physical drives on Windows requires Administrator privileges."
