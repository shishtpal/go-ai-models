# Build Scripts

This directory contains build scripts for the Catwalk project.

## build.ps1

PowerShell script to build all CLI tools from the `cmd/` directory into the `bin/` directory.

### Usage

```powershell
# Basic build (builds for current platform)
.\scripts\build.ps1

# Clean build (remove bin directory first)
.\scripts\build.ps1 -Clean

# Build for specific platform/architecture
.\scripts\build.ps1 -Platform linux -Arch amd64
.\scripts\build.ps1 -Platform darwin -Arch arm64
.\scripts\build.ps1 -Platform windows -Arch amd64

# Verbose output (shows build commands and details)
.\scripts\build.ps1 -Verbose

# Custom output directory
.\scripts\build.ps1 -OutputDir dist

# Combine options
.\scripts\build.ps1 -Clean -Verbose -Platform linux -Arch amd64
```

### Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|----------|
| `-OutputDir` | String | Output directory for binaries | `bin` |
| `-Platform` | String | Target platform (`windows`, `darwin`, `linux`) | Current OS |
| `-Arch` | String | Target architecture (`amd64`, `386`, `arm64`) | Current arch |
| `-Clean` | Switch | Clean build directory before building | `false` |
| `-Verbose` | Switch | Show verbose build output | `false` |

### Examples

#### Build for Windows (default)
```powershell
.\scripts\build.ps1
```

#### Build for macOS (Intel)
```powershell
.\scripts\build.ps1 -Platform darwin -Arch amd64
```

#### Build for macOS (Apple Silicon)
```powershell
.\scripts\build.ps1 -Platform darwin -Arch arm64
```

#### Build for Linux
```powershell
.\scripts\build.ps1 -Platform linux -Arch amd64
```

#### Clean rebuild with verbose output
```powershell
.\scripts\build.ps1 -Clean -Verbose
```

### CLI Tools Built

The script automatically discovers and builds all CLI tools from the `cmd/` directory:

- **copilot.exe** - Generate GitHub Copilot provider configuration
- **huggingface.exe** - Generate Hugging Face provider configuration
- **openrouter.exe** - Generate OpenRouter provider configuration
- **synthetic.exe** - Generate synthetic/test provider configuration
- **vercel.exe** - Generate Vercel AI provider configuration

### Output

All built binaries are placed in the `bin/` directory (or custom directory specified with `-OutputDir`).

On Windows, binaries have the `.exe` extension. On other platforms, they are executable without extension.

### Requirements

- Go 1.25.5 or later
- PowerShell 5.1 or later
- Appropriate Go compiler for target platform

### Exit Codes

- `0` - All tools built successfully
- `1` - One or more tools failed to build

### Notes

- The script automatically detects your current OS and architecture
- It sets `GOOS` and `GOARCH` environment variables for the build
- Failed builds are reported at the end with a summary
- The script uses colored output for better readability

### Troubleshooting

#### "go: command not found"
Make sure Go is installed and available in your PATH.

#### Build fails with "invalid import path"
Ensure you're running the script from the project root directory.

#### Permission denied on Linux/macOS
Make the script executable:
```bash
chmod +x scripts/build.ps1
```

#### Can't run script on Windows
Set execution policy:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Then run:
```powershell
pwsh -ExecutionPolicy Bypass -File scripts\build.ps1
```
