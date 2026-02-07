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

# Rename to avoid colliding with the built-in Write-Error cmdlet
function Write-ErrorMsg {
    param([string]$Message)
    Write-ColorOutput $Message "Red"
}

# Rename to avoid colliding with the built-in Write-Warning cmdlet
function Write-WarningMsg {
    param([string]$Message)
    Write-ColorOutput $Message "Yellow"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput $Message "Cyan"
}

# Get script directory — handle running from different contexts
if ($MyInvocation.MyCommand.Path) {
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $ProjectRoot = Split-Path -Parent $ScriptDir
} else {
    # Fallback: assume script is run from project root
    $ProjectRoot = Get-Location
}

$BinDir = Join-Path $ProjectRoot $OutputDir
$CmdDir = Join-Path $ProjectRoot "cmd"

Write-Info "Project root: $ProjectRoot"

# Clean build directory if requested
if ($Clean) {
    Write-Info "Cleaning build directory..."
    if (Test-Path $BinDir) {
        Remove-Item -Path $BinDir -Recurse -Force
        Write-Success "Build directory cleaned"
    } else {
        Write-Info "Build directory does not exist, nothing to clean"
    }

    # If only cleaning, exit early unless there's something else to do
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
        # $IsWindows, $IsMacOS, $IsLinux are automatic variables in PS Core (v6+)
        # For Windows PowerShell (v5.1), these don't exist — fall back safely
        if ($PSVersionTable.PSVersion.Major -ge 6) {
            $Platform = if ($IsWindows) { "windows" } elseif ($IsMacOS) { "darwin" } else { "linux" }
        } else {
            # Windows PowerShell 5.1 only runs on Windows
            $Platform = "windows"
        }
    }
}

if ([string]::IsNullOrEmpty($Arch)) {
    $Arch = $env:GOARCH
    if ([string]::IsNullOrEmpty($Arch)) {
        # Check runtime architecture more reliably
        $RuntimeArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
        $Arch = switch ($RuntimeArch) {
            "X64"   { "amd64" }
            "X86"   { "386" }
            "Arm64" { "arm64" }
            "Arm"   { "arm" }
            default {
                if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
            }
        }
    }
}

# Verify Go is installed before proceeding
try {
    $goVersion = & go version 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "go returned non-zero exit code"
    }
    Write-Info "Using: $goVersion"
} catch {
    Write-ErrorMsg "Go is not installed or not in PATH. Please install Go first."
    exit 1
}

# Set Go environment variables
$env:GOOS = $Platform
$env:GOARCH = $Arch

Write-Info "Building for: $Platform/$Arch"
Write-Info "Output directory: $BinDir"
Write-Host ""

# Build list of items to compile
$BuildItems = [System.Collections.ArrayList]@()

# Add cmd/ tools (if cmd/ directory exists)
if (Test-Path $CmdDir) {
    $Tools = Get-ChildItem -Path $CmdDir -Directory | Where-Object {
        Test-Path (Join-Path $_.FullName "main.go")
    }

    foreach ($tool in $Tools) {
        [void]$BuildItems.Add(@{
            Name      = $tool.Name
            # Point to the package directory, not the individual file —
            # this lets Go resolve multi-file packages correctly
            BuildPath = $tool.FullName
            Type      = "cmd"
        })
    }
} else {
    Write-WarningMsg "cmd/ directory not found at $CmdDir"
}

# Add root main.go as catwalk
$RootMain = Join-Path $ProjectRoot "main.go"
if (Test-Path $RootMain) {
    [void]$BuildItems.Add(@{
        Name      = "catwalk"
        BuildPath = $ProjectRoot
        Type      = "root"
    })
}

# Add examples (recursive search for main.go files)
$ExamplesDir = Join-Path $ProjectRoot "examples"
if (Test-Path $ExamplesDir) {
    $ExampleMains = Get-ChildItem -Path $ExamplesDir -Recurse -Filter "main.go" -File

    foreach ($exMain in $ExampleMains) {
        # Get relative path from examples directory
        $RelativePath = $exMain.Directory.FullName.Substring($ExamplesDir.Length + 1)
        # Replace path separators with hyphens for binary name
        $BinaryNamePart = $RelativePath -replace '[\\\/]', '-'

        [void]$BuildItems.Add(@{
            Name      = "example-$BinaryNamePart"
            BuildPath = $exMain.Directory.FullName
            Type      = "example"
        })
    }
}

if ($BuildItems.Count -eq 0) {
    Write-ErrorMsg "No tools found to build (checked $CmdDir, root main.go, and examples/)"
    exit 1
}

Write-Info "Found $($BuildItems.Count) tool(s) to build:"
foreach ($item in $BuildItems) {
    Write-Host "  - $($item.Name) ($($item.Type))"
}
Write-Host ""

# Build each tool
$SuccessCount = 0
$FailedTools = [System.Collections.ArrayList]@()

foreach ($Item in $BuildItems) {
    $ToolName = $Item.Name
    $BuildPath = $Item.BuildPath

    # Determine output binary name
    $BinaryName = if ($Platform -eq "windows") { "$ToolName.exe" } else { $ToolName }
    $OutputPath = Join-Path $BinDir $BinaryName

    Write-Host "Building $ToolName..." -NoNewline

    # Build arguments as an array — avoids quoting/escaping issues with
    # Invoke-Expression and properly handles paths with spaces
    $goArgs = @("build", "-o", $OutputPath)

    if ($Verbose) {
        $goArgs += @("-v", "-x")
        Write-Host ""
        Write-Host "  Running: go $($goArgs -join ' ') `"$BuildPath`"" -ForegroundColor Gray
    }

    # Use ./... pattern if pointing at a directory, or the path directly
    $goArgs += $BuildPath

    try {
        # Use Start-Process for reliable exit code capture, or call go directly
        # Using & (call operator) is the idiomatic PowerShell approach
        if ($Verbose) {
            # Show output in real time
            & go @goArgs 2>&1 | ForEach-Object {
                if ($_ -is [System.Management.Automation.ErrorRecord]) {
                    Write-Host "  $_" -ForegroundColor Red
                } else {
                    Write-Host "  $_" -ForegroundColor Gray
                }
            }
        } else {
            $BuildOutput = & go @goArgs 2>&1
        }

        if ($LASTEXITCODE -eq 0) {
            Write-Success " OK"
            $SuccessCount++
        } else {
            Write-ErrorMsg " FAILED"
            Write-ErrorMsg "  Build failed for ${ToolName}:"
            if (-not $Verbose -and $BuildOutput) {
                foreach ($line in $BuildOutput) {
                    Write-Host "  $line" -ForegroundColor Red
                }
            }
            [void]$FailedTools.Add($ToolName)
        }
    } catch {
        Write-ErrorMsg " FAILED"
        Write-ErrorMsg "  Exception building ${ToolName}: $($_.Exception.Message)"
        [void]$FailedTools.Add($ToolName)
    }
}

Write-Host ""
Write-Info "Build Summary:"
Write-Host "  Total items:  $($BuildItems.Count)"
Write-Success "  Successful:   $SuccessCount"
if ($FailedTools.Count -gt 0) {
    Write-ErrorMsg "  Failed:       $($FailedTools.Count)"
    Write-ErrorMsg "  Failed tools: $($FailedTools -join ', ')"
}

Write-Host ""
if ($FailedTools.Count -eq 0) {
    Write-Success "All CLI tools built successfully!"
    Write-Info "Binaries are located in: $BinDir"

    # List built binaries with sizes
    Write-Host ""
    Write-Info "Built binaries:"
    Get-ChildItem -Path $BinDir -File | Sort-Object Name | ForEach-Object {
        $Size = if ($_.Length -ge 1MB) {
            "$([math]::Round($_.Length / 1MB, 2)) MB"
        } else {
            "$([math]::Round($_.Length / 1KB, 2)) KB"
        }
        Write-Host "  $($_.Name) ($Size)"
    }
    exit 0
} else {
    Write-ErrorMsg "Build completed with errors"
    exit 1
}