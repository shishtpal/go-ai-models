#!/usr/bin/env pwsh
# Build script for Catwalk CLI tools
# Builds all CLI tools from cmd/ directory into bin/ directory

param(
    [string]$OutputDir = "bin",
    [string]$Platform = "",
    [string]$Arch = "",
    [switch]$Clean,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

# Color output functions
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput $Message "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput $Message "Red"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput $Message "Yellow"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput $Message "Cyan"
}

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$BinDir = Join-Path $ProjectRoot $OutputDir
$CmdDir = Join-Path $ProjectRoot "cmd"

# Clean build directory if requested
if ($Clean) {
    Write-Info "Cleaning build directory..."
    if (Test-Path $BinDir) {
        Remove-Item -Path $BinDir -Recurse -Force
        Write-Success "Build directory cleaned"
    } else {
        Write-Info "Build directory does not exist, nothing to clean"
    }
}

# Create build directory
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    Write-Success "Created build directory: $BinDir"
}

# Determine OS and architecture
if ([string]::IsNullOrEmpty($Platform)) {
    $Platform = $env:GOOS
    if ([string]::IsNullOrEmpty($Platform)) {
        $Platform = if ($IsWindows) { "windows" } elseif ($IsMacOS) { "darwin" } else { "linux" }
    }
}

if ([string]::IsNullOrEmpty($Arch)) {
    $Arch = $env:GOARCH
    if ([string]::IsNullOrEmpty($Arch)) {
        $Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    }
}

# Set Go environment variables
$env:GOOS = $Platform
$env:GOARCH = $Arch

Write-Info "Building for: $Platform/$Arch"
Write-Info "Output directory: $BinDir"
Write-Host ""

# Get all CLI tool directories
$Tools = Get-ChildItem -Path $CmdDir -Directory | Where-Object {
    Test-Path (Join-Path $_.FullName "main.go")
}

if ($Tools.Count -eq 0) {
    Write-Error "No CLI tools found in $CmdDir"
    exit 1
}

Write-Info "Found $($Tools.Count) CLI tool(s) to build:"
$Tools | ForEach-Object { Write-Host "  - $($_.Name)" }
Write-Host ""

# Build each tool
$SuccessCount = 0
$FailedTools = @()

foreach ($Tool in $Tools) {
    $ToolName = $Tool.Name
    $ToolPath = Join-Path $Tool.FullName "main.go"

    # Determine output binary name
    $BinaryName = if ($Platform -eq "windows") { "$ToolName.exe" } else { $ToolName }
    $OutputPath = Join-Path $BinDir $BinaryName

    Write-Host "Building $ToolName..." -NoNewline

    # Build command
    if ($Verbose) {
        $BuildCmd = "go build -v -x -o `"$OutputPath`" `"$ToolPath`""
        Write-Host ""
        Write-Host "  Running: $BuildCmd" -ForegroundColor Gray
    } else {
        $BuildCmd = "go build -o `"$OutputPath`" `"$ToolPath`""
    }

    try {
        $BuildOutput = Invoke-Expression $BuildCmd 2>&1

        if ($LASTEXITCODE -eq 0) {
            if (-not $Verbose) {
                Write-Success " OK"
            } else {
                Write-Success " OK"
                if ($BuildOutput) {
                    $BuildOutput | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
                }
            }
            $SuccessCount++
        } else {
            if (-not $Verbose) {
                Write-Error " FAILED"
            } else {
                Write-Error " FAILED"
            }
            Write-Error "Build failed for $ToolName"
            if ($Verbose -and $BuildOutput) {
                $BuildOutput | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
            }
            $FailedTools += $ToolName
        }
    } catch {
        Write-Error " FAILED"
        Write-Error "Exception building $ToolName`: $_"
        $FailedTools += $ToolName
    }
}

Write-Host ""
Write-Info "Build Summary:"
Write-Host "  Total tools: $($Tools.Count)"
Write-Success "  Successful: $SuccessCount"
if ($FailedTools.Count -gt 0) {
    Write-Error "  Failed: $($FailedTools.Count)"
    Write-Error "  Failed tools: $($FailedTools -join ', ')"
}

Write-Host ""
if ($FailedTools.Count -eq 0) {
    Write-Success "All CLI tools built successfully!"
    Write-Info "Binaries are located in: $BinDir"

    # List built binaries
    Write-Host ""
    Write-Info "Built binaries:"
    Get-ChildItem -Path $BinDir -File | ForEach-Object {
        $Size = [math]::Round($_.Length / 1KB, 2)
        Write-Host "  $($_.Name) ($Size KB)"
    }
    exit 0
} else {
    Write-Error "Build completed with errors"
    exit 1
}
