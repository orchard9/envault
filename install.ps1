# envault installer script for Windows
# Usage: irm https://raw.githubusercontent.com/orchard9/envault/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "orchard9/envault"
$BinaryName = "envault"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:USERPROFILE\.local\bin" }

function Write-Info {
    param([string]$Message)
    Write-Host "==> " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "Warning: " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Err {
    param([string]$Message)
    Write-Host "Error: " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

function Get-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "x86_64" }
    } else {
        Write-Err "32-bit Windows is not supported"
    }

    return "Windows_$arch"
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        $version = $response.tag_name -replace '^v', ''
        return $version
    } catch {
        Write-Warn "Could not fetch latest release version: $_"
        return $null
    }
}

function Install-Binary {
    param(
        [string]$Version,
        [string]$Platform
    )

    $downloadUrl = "https://github.com/$Repo/releases/download/v$Version/${BinaryName}_${Version}_${Platform}.zip"
    $tmpDir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "envault-install-$(Get-Random)")

    try {
        Write-Info "Downloading $BinaryName v$Version for $Platform..."

        $zipPath = Join-Path $tmpDir "$BinaryName.zip"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

        Write-Info "Extracting archive..."
        Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

        # Create install directory if it doesn't exist
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        Write-Info "Installing to $InstallDir\$BinaryName.exe..."
        $binaryPath = Get-ChildItem -Path $tmpDir -Filter "$BinaryName.exe" -Recurse | Select-Object -First 1
        if ($binaryPath) {
            Move-Item -Path $binaryPath.FullName -Destination "$InstallDir\$BinaryName.exe" -Force
        } else {
            # Try without .exe extension (in case archive has unix binary)
            $binaryPath = Get-ChildItem -Path $tmpDir -Filter $BinaryName -Recurse | Select-Object -First 1
            if ($binaryPath) {
                Move-Item -Path $binaryPath.FullName -Destination "$InstallDir\$BinaryName.exe" -Force
            } else {
                throw "Binary not found in archive"
            }
        }

        return $true
    } catch {
        Write-Warn "Failed to download binary: $_"
        return $false
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Install-WithGo {
    $goPath = Get-Command go -ErrorAction SilentlyContinue
    if (-not $goPath) {
        Write-Err "Go is not installed and binary download failed. Please install Go or download binary manually from https://github.com/$Repo/releases"
    }

    Write-Info "Installing via 'go install'..."
    & go install "github.com/$Repo/cmd/$BinaryName@latest"

    $goBin = if ($env:GOBIN) { $env:GOBIN } else { Join-Path (go env GOPATH) "bin" }
    Write-Info "Installed to $goBin\$BinaryName.exe"

    # Check if GOPATH/bin is in PATH
    if ($env:PATH -notlike "*$goBin*") {
        Write-Warn "Add $goBin to your PATH to use $BinaryName"
        Write-Host ""
        Write-Host "Run this command to add to your PATH for this session:"
        Write-Host "  `$env:PATH += `";$goBin`""
        Write-Host ""
        Write-Host "To add permanently, run:"
        Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$goBin', 'User')"
    }
}

function Test-Installation {
    # Check in INSTALL_DIR first
    $installedPath = Join-Path $InstallDir "$BinaryName.exe"
    if (Test-Path $installedPath) {
        $version = & $installedPath version 2>$null
        Write-Info "Successfully installed: $installedPath"
        Write-Host $version
        return $true
    }

    # Check in PATH
    $inPath = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if ($inPath) {
        $version = & $BinaryName version 2>$null
        Write-Info "Successfully installed: $($inPath.Source)"
        Write-Host $version
        return $true
    }

    return $false
}

function Add-ToPath {
    param([string]$Directory)

    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$Directory*") {
        Write-Warn "Add $Directory to your PATH to use $BinaryName"
        Write-Host ""
        Write-Host "Run this command to add to your PATH permanently:"
        Write-Host "  [Environment]::SetEnvironmentVariable('PATH', [Environment]::GetEnvironmentVariable('PATH', 'User') + ';$Directory', 'User')"
        Write-Host ""
        Write-Host "Then restart your terminal or run:"
        Write-Host "  `$env:PATH = [Environment]::GetEnvironmentVariable('PATH', 'User')"
    }
}

function Main {
    Write-Host ""
    Write-Info "Installing $BinaryName..."
    Write-Host ""

    # Detect platform
    $platform = Get-Platform
    Write-Info "Detected platform: $platform"

    # Try to download binary from releases
    $version = Get-LatestVersion
    if ($version) {
        Write-Info "Latest version: v$version"

        if (Install-Binary -Version $version -Platform $platform) {
            Write-Host ""
            if (Test-Installation) {
                Write-Host ""
                Write-Info "Installation complete!"

                Add-ToPath -Directory $InstallDir

                Write-Host ""
                Write-Info "Next steps:"
                Write-Host "  1. Install age: scoop install age  (or: go install filippo.io/age/cmd/...@latest)"
                Write-Host "  2. Run: $BinaryName --help"
                exit 0
            }
        }
    }

    # Fallback to go install
    Write-Warn "Binary download failed, trying 'go install' as fallback..."
    Install-WithGo

    Write-Host ""
    if (Test-Installation) {
        Write-Host ""
        Write-Info "Installation complete!"
        Write-Host ""
        Write-Info "Next steps:"
        Write-Host "  1. Install age: scoop install age  (or: go install filippo.io/age/cmd/...@latest)"
        Write-Host "  2. Run: $BinaryName --help"
    } else {
        Write-Err "Installation verification failed"
    }
}

Main
