# PvFlasher Windows Installation Script
param (
    [string]$Version
)

$ProjectName = "pvflasher"
$RepoUrl = "https://gitlab.com/pantacor/pvflasher"
$ApiUrl = "https://gitlab.com/api/v4/projects/pantacor%2Fpvflasher"
$InstallDir = Join-Path $env:LOCALAPPDATA "pvflasher"

# Detect Architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
    $Arch = "arm64"
}

Write-Host "Installing $ProjectName for $Arch..." -ForegroundColor Cyan

# 1. Get version if not specified
if (-not $Version) {
    Write-Host "Fetching latest release information..."
    $Releases = Invoke-RestMethod -Uri "$ApiUrl/releases"
    $Version = $Releases[0].tag_name
}

if (-not $Version) {
    Write-Error "Could not find version. Please check $RepoUrl"
    return
}

Write-Host "Installing version: $Version"

# 2. Construct download URL for Windows zip
# Filename pattern: pvflasher-windows-amd64.zip or pvflasher-windows-arm64.zip
$ZipUrl = "$ApiUrl/packages/generic/$ProjectName/$Version/pvflasher-windows-$Arch.zip"
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

# 4. Add to User PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to User PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path += ";$InstallDir"
}

Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "You may need to restart your terminal to use 'pvflasher'."
Write-Host "Note: Flashing physical drives on Windows requires Administrator privileges."
