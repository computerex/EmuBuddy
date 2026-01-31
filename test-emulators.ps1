#
# EmuBuddy Emulator Test Script for Windows
# Downloads small test ROMs and verifies each emulator works
#

param(
    [string]$InstallDir = ""
)

# ANSI color codes for PowerShell
$colors = @{
    Reset = "`e[0m"
    Red = "`e[31m"
    Green = "`e[32m"
    Yellow = "`e[33m"
    Blue = "`e[34m"
    Cyan = "`e[36m"
}

function Write-ColorMessage {
    param(
        [string]$Color,
        [string]$Message
    )
    Write-Host "$($colors[$Color])$Message$($colors.Reset)"
}

function Convert-SizeToBytes {
    param([string]$sizeStr)
    if ([string]::IsNullOrEmpty($sizeStr)) { return [double]::MaxValue }

    $sizeStr = $sizeStr.Trim().ToUpper()
    if ($sizeStr -match "([\d\.]+)\s*(KB|KI|MB|MI|GB|GI|B)?") {
        $value = [double]$matches[1]
        $unit = if ($matches[2]) { $matches[2] } else { "B" }

        switch -Regex ($unit) {
            "^(KB|KI)$" { return $value * 1KB }
            "^(MB|MI)$" { return $value * 1MB }
            "^(GB|GI)$" { return $value * 1GB }
            default { return $value }
        }
    }
    return [double]::MaxValue
}

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$TestDir = Join-Path $ScriptDir "test-roms"
$ReportFile = Join-Path $ScriptDir "emulator-test-report.txt"

# Create test directory
New-Item -ItemType Directory -Force -Path $TestDir | Out-Null

# Test results
$TestResults = [System.Collections.Generic.List[PSCustomObject]]::new()

# All 27 systems to test
$TestSystems = @(
    "nes", "snes", "n64",
    "gb", "gbc", "gba",
    "ds", "3ds",
    "gc", "wii", "wiiu",
    "genesis", "sms", "gamegear",
    "dreamcast",
    "psp", "ps1", "ps2",
    "atari2600", "atari7800", "lynx",
    "ngpc", "wonderswan", "wonderswancolor",
    "coleco", "intellivision", "virtualboy",
    "tg16"
)

# ROM JSON file mappings
$RomJsonFiles = @{
    "nes" = "nes.json"
    "snes" = "snes.json"
    "n64" = "n64.json"
    "gb" = "gb.json"
    "gbc" = "gbc.json"
    "gba" = "gba.json"
    "ds" = "ds.json"
    "3ds" = "3ds.json"
    "gc" = "games_1g1r_english_gc_full.json"
    "wii" = "games_1g1r_english_wii_full.json"
    "wiiu" = "wiiu.json"
    "genesis" = "games_1g1r_english_genesis.json"
    "sms" = "games_1g1r_english_sms.json"
    "gamegear" = "games_1g1r_english_gamegear.json"
    "dreamcast" = "dreamcast.json"
    "psp" = "games_1g1r_english_psp_full.json"
    "ps1" = "games_1g1r_english_ps1_full.json"
    "ps2" = "games_1g1r_english_ps2_full.json"
    "atari2600" = "games_1g1r_english_atari2600.json"
    "atari7800" = "games_1g1r_english_atari7800.json"
    "lynx" = "games_1g1r_english_lynx.json"
    "ngpc" = "games_1g1r_english_ngpc.json"
    "wonderswan" = "games_1g1r_english_wonderswan.json"
    "wonderswancolor" = "games_1g1r_english_wonderswancolor.json"
    "coleco" = "games_1g1r_english_coleco.json"
    "intellivision" = "games_1g1r_english_intellivision.json"
    "virtualboy" = "games_1g1r_english_virtualboy.json"
    "tg16" = "games_1g1r_english_tg16.json"
}

function Get-TestRom {
    param(
        [string]$System,
        [string]$BaseDir
    )

    $jsonFile = Get-RomJsonFile -System $System -BaseDir $BaseDir
    if ([string]::IsNullOrEmpty($jsonFile)) {
        Write-ColorMessage "Red" "  ROM JSON not found for $System"
        return $null
    }

    if (-not (Test-Path $jsonFile)) {
        Write-ColorMessage "Red" "  JSON file not found: $jsonFile"
        return $null
    }

    try {
        $jsonContent = Get-Content $jsonFile -Raw | ConvertFrom-Json

        # Find smallest clean ROM (skip Pirate, Proto, Beta, Unl, Sample, Demo, Test)
        $excludePatterns = @("Pirate", "Proto", "Beta", "Unl", "Sample", "Demo", "Test", "Program", "BIOS")
        $cleanRom = $jsonContent | Where-Object {
            $name = $_.name
            -not ($excludePatterns | Where-Object { $name -match $_ })
        } | Sort-Object { Convert-SizeToBytes -sizeStr $_.size } | Select-Object -First 1

        if ($null -eq $cleanRom) {
            # Fallback to first ROM
            $cleanRom = $jsonContent | Select-Object -First 1
        }

        if ($null -ne $cleanRom) {
            return @{
                Name = $cleanRom.name
                Url  = $cleanRom.url
                Size = $cleanRom.size
            }
        }
    }
    catch {
        Write-ColorMessage "Red" "  Failed to parse JSON: $_"
    }

    return $null
}

function Get-RomJsonFile {
    param(
        [string]$System,
        [string]$BaseDir
    )

    $jsonFileName = $RomJsonFiles[$System]
    if ([string]::IsNullOrEmpty($jsonFileName)) {
        return $null
    }

    # Try multiple locations
    $paths = @(
        (Join-Path $BaseDir "1g1rsets\$jsonFileName"),
        (Join-Path $ScriptDir "1g1rsets\$jsonFileName")
    )

    foreach ($path in $paths) {
        if (Test-Path $path) {
            return $path
        }
    }

    return $null
}

function Download-Rom {
    param(
        [string]$System,
        [string]$BaseDir
    )

    $romInfo = Get-TestRom -System $System -BaseDir $BaseDir
    if ($null -eq $romInfo) {
        return $null
    }

    $romDir = Join-Path $BaseDir "roms\$System"
    $romFile = Join-Path $romDir $romInfo.Name

    New-Item -ItemType Directory -Force -Path $romDir | Out-Null

    if (Test-Path $romFile) {
        Write-ColorMessage "Yellow" "  ROM already exists: $($romInfo.Name)"
        return $romFile
    }

    Write-ColorMessage "Blue" "  Downloading: $($romInfo.Name) ($($romInfo.Size))"

    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $romInfo.Url -OutFile $romFile -UseBasicParsing
        $ProgressPreference = 'Continue'
        Write-ColorMessage "Green" "  Downloaded successfully"
        return $romFile
    }
    catch {
        Write-ColorMessage "Red" "  Download failed: $_"
        Remove-Item -Path $romFile -Force -ErrorAction SilentlyContinue
        return $null
    }
}

function Test-Emulator {
    param(
        [string]$System,
        [string]$RomPath,
        [string]$BaseDir
    )

    Write-ColorMessage "Cyan" "`n[Testing $System]"

    # Use console version for testing (has stdout/stderr)
    $launcherPath = Join-Path $BaseDir "EmuBuddyLauncher-console.exe"
    
    # Fall back to regular launcher if console version doesn't exist
    if (-not (Test-Path $launcherPath)) {
        $launcherPath = Join-Path $BaseDir "EmuBuddyLauncher.exe"
    }

    if (-not (Test-Path $launcherPath)) {
        Write-ColorMessage "Red" "  Launcher not found: $launcherPath"
        return @{
            System  = $System
            Result  = "SKIP"
            Message = "Launcher not found"
        }
    }

    Write-ColorMessage "Yellow" "  Launching with EmuBuddy launcher..."
    Write-ColorMessage "Yellow" "  Command: $launcherPath --launch $System $RomPath"
    Write-ColorMessage "Yellow" "  (Close the emulator when you've verified it works)"

    try {
        # Use direct command invocation to ensure arguments pass correctly
        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = $launcherPath
        $psi.Arguments = "--launch `"$System`" `"$RomPath`""
        $psi.UseShellExecute = $false
        $psi.RedirectStandardOutput = $true
        $psi.RedirectStandardError = $true
        $psi.CreateNoWindow = $false  # Allow console window for debugging

        $process = [System.Diagnostics.Process]::Start($psi)

        # Capture output for debugging
        $output = $process.StandardOutput.ReadToEnd()
        $error = $process.StandardError.ReadToEnd()
        if ($output) { Write-Host $output }
        if ($error) { Write-Host $error -ForegroundColor Red }

        $process.WaitForExit()
        $exitCode = $process.ExitCode
    }
    catch {
        Write-ColorMessage "Red" "  Failed to launch: $_"
        return @{
            System  = $System
            Result  = "FAIL"
            Message = "Failed to launch: $_"
        }
    }

    Write-Host ""
    $response = Read-Host "  Did the emulator launch and display the game correctly? (y/n)"

    if ($response -match '^[Yy]') {
        Write-ColorMessage "Green" "  Test PASSED"
        return @{
            System  = $System
            Result  = "PASS"
            Message = "Emulator launched successfully"
        }
    }
    else {
        Write-ColorMessage "Red" "  Test FAILED"
        $issue = Read-Host "  Please describe the issue (or press Enter to skip)"
        $msg = if ([string]::IsNullOrWhiteSpace($issue)) { "User reported failure" } else { $issue }
        return @{
            System  = $System
            Result  = "FAIL"
            Message = $msg
        }
    }
}

function Get-InstallDir {
    if (-not [string]::IsNullOrEmpty($InstallDir)) {
        if (Test-Path $InstallDir) {
            return $InstallDir
        }
    }

    # Auto-detect installation directory
    $possiblePaths = @(
        $ScriptDir,
        (Split-Path -Parent $ScriptDir),
        "${env:LOCALAPPDATA}\EmuBuddy",
        "${env:PROGRAMFILES}\EmuBuddy"
    )

    foreach ($path in $possiblePaths) {
        $launcherPath = Join-Path $path "EmuBuddyLauncher.exe"
        if (Test-Path $launcherPath) {
            return $path
        }
        $romsPath = Join-Path $path "roms"
        if (Test-Path $romsPath) {
            return $path
        }
    }

    return $ScriptDir
}

function Write-Header {
    Clear-Host
    Write-ColorMessage "Cyan" "========================================="
    Write-ColorMessage "Cyan" "  EmuBuddy Emulator Test Suite"
    Write-ColorMessage "Cyan" "  Windows Edition"
    Write-ColorMessage "Cyan" "========================================="
    Write-Host ""
    Write-Host "This script will download small test ROMs"
    Write-Host "and verify each emulator launches correctly."
    Write-Host ""
}

function Write-Report {
    param(
        [string]$InstallDir
    )

    Write-ColorMessage "Green" "`n`n========================================="
    Write-ColorMessage "Green" "       TEST REPORT"
    Write-ColorMessage "Green" "========================================="

    $reportContent = "EmuBuddy Emulator Test Report`r`n"
    $reportContent += "Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')`r`n"
    $reportContent += "Installation: $InstallDir`r`n"
    $reportContent += "Platform: Windows`r`n"
    $reportContent += "PowerShell Version: $($PSVersionTable.PSVersion)`r`n"
    $reportContent += "`r`n"
    $reportContent += "=========================================`r`n"
    $reportContent += "`r`n"

    $passed = 0
    $failed = 0
    $skipped = 0

    foreach ($result in $TestResults) {
        $line = "{0,-20} : {1,-6} : {2}" -f $result.System, $result.Result, $result.Message
        $reportContent += $line + "`r`n"

        switch ($result.Result) {
            "PASS" {
                $passed++
                Write-ColorMessage "Green" "[PASS] $($result.System): $($result.Message)"
            }
            "FAIL" {
                $failed++
                Write-ColorMessage "Red" "[FAIL] $($result.System): $($result.Message)"
            }
            "SKIP" {
                $skipped++
                Write-ColorMessage "Yellow" "[SKIP] $($result.System): $($result.Message)"
            }
        }
    }

    $reportContent += "`r`n"
    $reportContent += "=========================================`r`n"
    $reportContent += "Summary:`r`n"
    $reportContent += "  Passed:  $passed`r`n"
    $reportContent += "  Failed:  $failed`r`n"
    $reportContent += "  Skipped: $skipped`r`n"
    $reportContent += "  Total:   $($TestResults.Count)`r`n"
    $reportContent += "=========================================`r`n"

    $reportContent | Out-File -FilePath $ReportFile -Encoding UTF8

    Write-Host ""
    Write-ColorMessage "Green" "Summary: Passed=$passed, Failed=$failed, Skipped=$skipped, Total=$($TestResults.Count)"
    Write-ColorMessage "Green" "Report saved to: $ReportFile"
}

# Main execution
Write-Header

$installDir = Get-InstallDir
Write-ColorMessage "Green" "Found installation: $installDir`n"

$launcherPath = Join-Path $installDir "EmuBuddyLauncher.exe"
if (-not (Test-Path $launcherPath)) {
    Write-ColorMessage "Red" "ERROR: EmuBuddyLauncher.exe not found!"
    Write-ColorMessage "Yellow" "Please specify the installation directory with -InstallDir parameter."
    Write-Host ""
    Write-Host "Example: .\test-emulators.ps1 -InstallDir 'C:\Path\To\EmuBuddy'"
    exit 1
}

Write-ColorMessage "Blue" "Testing $($TestSystems.Count) emulator systems..."
Write-Host ""

foreach ($system in $TestSystems) {
    Write-ColorMessage "Cyan" "========================================="
    Write-ColorMessage "Cyan" "System: $system"
    Write-ColorMessage "Cyan" "========================================="

    # Download ROM
    $romPath = Download-Rom -System $system -BaseDir $installDir

    if ($null -ne $romPath) {
        # Test emulator
        $result = Test-Emulator -System $system -RomPath $romPath -BaseDir $installDir
        $TestResults.Add($result)
    }
    else {
        $TestResults.Add([PSCustomObject]@{
            System  = $system
            Result  = "SKIP"
            Message = "ROM download failed"
        })
    }

    Start-Sleep -Milliseconds 500
}

Write-Report -InstallDir $installDir

Write-ColorMessage "Blue" "`n`nTest complete! Check $ReportFile for full results."
