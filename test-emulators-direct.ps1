#
# EmuBuddy Emulator Test Script for Windows (Direct Mode)
# Tests emulators directly without using EmuBuddyLauncher
#

param(
    [string]$InstallDir = ""
)

# ANSI color codes
$colors = @{
    Reset = "`e[0m"
    Red = "`e[31m"
    Green = "`e[32m"
    Yellow = "`e[33m"
    Blue = "`e[34m"
    Cyan = "`e[36m"
}

function Write-ColorMessage {
    param([string]$Color, [string]$Message)
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

# Emulator configurations for Windows (direct paths)
$EmulatorConfigs = @{
    "nes" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\nestopia_libretro.dll")
    }
    "snes" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\snes9x_libretro.dll")
    }
    "n64" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mupen64plus_next_libretro.dll")
    }
    "gb" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\gambatte_libretro.dll")
    }
    "gbc" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\gambatte_libretro.dll")
    }
    "gba" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mgba_libretro.dll")
    }
    "ds" = @{
        Emulator = "Emulators\melonDS\melonDS.exe"
        Args = @()
    }
    "3ds" = @{
        Emulator = "Emulators\Azahar\azahar.exe"
        Args = @()
    }
    "gc" = @{
        Emulator = "Emulators\Dolphin\Dolphin-x64\Dolphin.exe"
        Args = @("-e")
    }
    "wii" = @{
        Emulator = "Emulators\Dolphin\Dolphin-x64\Dolphin.exe"
        Args = @("-e")
    }
    "wiiu" = @{
        Emulator = "Emulators\Cemu\Cemu.exe"
        Args = @("-g")
    }
    "genesis" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\genesis_plus_gx_libretro.dll")
    }
    "sms" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\genesis_plus_gx_libretro.dll")
    }
    "gamegear" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\genesis_plus_gx_libretro.dll")
    }
    "dreamcast" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\flycast_libretro.dll")
    }
    "psp" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\ppsspp_libretro.dll")
    }
    "ps1" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mednafen_psx_hw_libretro.dll")
    }
    "ps2" = @{
        Emulator = "Emulators\PCSX2\pcsx2-qt.exe"
        Args = @()
    }
    "atari2600" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\stella_libretro.dll")
    }
    "atari7800" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\prosystem_libretro.dll")
    }
    "lynx" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\handy_libretro.dll")
    }
    "ngpc" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mednafen_ngp_libretro.dll")
    }
    "wonderswan" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mednafen_wswan_libretro.dll")
    }
    "wonderswancolor" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mednafen_wswan_libretro.dll")
    }
    "coleco" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\bluemsx_libretro.dll")
    }
    "intellivision" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\freeintv_libretro.dll")
    }
    "virtualboy" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mednafen_vb_libretro.dll")
    }
    "tg16" = @{
        Emulator = "Emulators\RetroArch\RetroArch-Win64\retroarch.exe"
        Args = @("-L", "Emulators\RetroArch\RetroArch-Win64\cores\mednafen_pce_fast_libretro.dll")
    }
}

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

$TestSystems = @(
    "nes", "snes", "n64", "gb", "gbc", "gba", "ds", "3ds",
    "gc", "wii", "wiiu", "genesis", "sms", "gamegear", "dreamcast",
    "psp", "ps1", "ps2", "atari2600", "atari7800", "lynx",
    "ngpc", "wonderswan", "wonderswancolor", "coleco", "intellivision",
    "virtualboy", "tg16"
)

$TestResults = [System.Collections.Generic.List[PSCustomObject]]::new()
$ReportFile = Join-Path $ScriptDir "emulator-test-report.txt"

function Get-TestRom {
    param([string]$System, [string]$BaseDir)

    $jsonFileName = $RomJsonFiles[$System]
    if ([string]::IsNullOrEmpty($jsonFileName)) { return $null }

    $jsonPath = Join-Path $BaseDir "1g1rsets\$jsonFileName"
    if (-not (Test-Path $jsonPath)) { return $null }

    try {
        $jsonContent = Get-Content $jsonPath -Raw | ConvertFrom-Json
        $excludePatterns = @("Pirate", "Proto", "Beta", "Unl", "Sample", "Demo", "Test", "Program", "BIOS")
        $cleanRom = $jsonContent | Where-Object {
            $name = $_.name
            -not ($excludePatterns | Where-Object { $name -match $_ })
        } | Sort-Object { Convert-SizeToBytes -sizeStr $_.size } | Select-Object -First 1

        if ($null -eq $cleanRom) {
            $cleanRom = $jsonContent | Select-Object -First 1
        }

        return @{
            Name = $cleanRom.name
            Url  = $cleanRom.url
            Size = $cleanRom.size
        }
    }
    catch {
        return $null
    }
}

function Download-Rom {
    param([string]$System, [string]$BaseDir)

    $romInfo = Get-TestRom -System $System -BaseDir $BaseDir
    if ($null -eq $romInfo) { return $null }

    $romDir = Join-Path $BaseDir "roms\$System"
    $romFile = Join-Path $romDir $romInfo.Name
    New-Item -ItemType Directory -Force -Path $romDir | Out-Null

    if (Test-Path $romFile) {
        Write-ColorMessage "Yellow" "  ROM exists: $($romInfo.Name)"
        return $romFile
    }

    Write-ColorMessage "Blue" "  Downloading: $($romInfo.Name)"
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $romInfo.Url -OutFile $romFile -UseBasicParsing
        $ProgressPreference = 'Continue'
        Write-ColorMessage "Green" "  Downloaded"
        return $romFile
    }
    catch {
        Write-ColorMessage "Red" "  Download failed: $_"
        return $null
    }
}

function Test-Emulator {
    param([string]$System, [string]$RomPath, [string]$BaseDir)

    Write-ColorMessage "Cyan" "`n[Testing $System]"

    $config = $EmulatorConfigs[$System]
    if ($null -eq $config) {
        return @{ System = $System; Result = "SKIP"; Message = "No emulator config" }
    }

    $emuPath = Join-Path $BaseDir $config.Emulator
    if (-not (Test-Path $emuPath)) {
        return @{ System = $System; Result = "SKIP"; Message = "Emulator not found: $($config.Emulator)" }
    }

    $args = $config.Args + $RomPath
    $argsStr = $args -join ' '

    Write-ColorMessage "Yellow" "  Launching: $emuPath"
    Write-ColorMessage "Yellow" "  Args: $argsStr"
    Write-ColorMessage "Yellow" "  (Close emulator when done testing)"

    try {
        $process = Start-Process -FilePath $emuPath -ArgumentList $args -PassThru
        $process.WaitForExit()
    }
    catch {
        return @{ System = $System; Result = "FAIL"; Message = "Launch failed: $_" }
    }

    $response = Read-Host "  Did it work? (y/n)"
    if ($response -match '^[Yy]') {
        Write-ColorMessage "Green" "  PASSED"
        return @{ System = $System; Result = "PASS"; Message = "Launched successfully" }
    }
    else {
        Write-ColorMessage "Red" "  FAILED"
        $issue = Read-Host "  Describe issue (or Enter to skip)"
        $msg = if ([string]::IsNullOrWhiteSpace($issue)) { "User reported failure" } else { $issue }
        return @{ System = $System; Result = "FAIL"; Message = $msg }
    }
}

function Get-InstallDir {
    if (-not [string]::IsNullOrEmpty($InstallDir)) {
        if (Test-Path $InstallDir) { return $InstallDir }
    }

    $possiblePaths = @($ScriptDir, (Split-Path -Parent $ScriptDir))
    foreach ($path in $possiblePaths) {
        if (Test-Path (Join-Path $path "Emulators")) { return $path }
    }
    return $ScriptDir
}

Clear-Host
Write-ColorMessage "Cyan" "========================================="
Write-ColorMessage "Cyan" "  EmuBuddy Direct Emulator Test"
Write-ColorMessage "Cyan" "  Windows Edition"
Write-ColorMessage "Cyan" "========================================="
Write-Host ""

$installDir = Get-InstallDir
Write-ColorMessage "Green" "Installation: $installDir`n"

Write-ColorMessage "Blue" "Testing $($TestSystems.Count) systems..."
Write-Host ""

foreach ($system in $TestSystems) {
    Write-ColorMessage "Cyan" "========================================="
    Write-ColorMessage "Cyan" "System: $system"
    Write-ColorMessage "Cyan" "========================================="

    $romPath = Download-Rom -System $system -BaseDir $installDir

    if ($null -ne $romPath) {
        $result = Test-Emulator -System $system -RomPath $romPath -BaseDir $installDir
        $TestResults.Add([PSCustomObject]@{
            System  = $result.System
            Result  = $result.Result
            Message = $result.Message
        })
    }
    else {
        $TestResults.Add([PSCustomObject]@{
            System  = $system
            Result  = "SKIP"
            Message = "ROM download failed"
        })
    }
}

Write-ColorMessage "Green" "`n`n========================================="
Write-ColorMessage "Green" "       TEST REPORT"
Write-ColorMessage "Green" "========================================="

$passed = 0; $failed = 0; $skipped = 0
$report = "EmuBuddy Emulator Test Report`r`nGenerated: $(Get-Date -Format 's')`r`n`r`n"

foreach ($result in $TestResults) {
    $line = "{0,-20} : {1,-6} : {2}" -f $result.System, $result.Result, $result.Message
    $report += $line + "`r`n"

    switch ($result.Result) {
        "PASS" { $passed++; Write-ColorMessage "Green" "[PASS] $($result.System)" }
        "FAIL" { $failed++; Write-ColorMessage "Red" "[FAIL] $($result.System): $($result.Message)" }
        "SKIP" { $skipped++; Write-ColorMessage "Yellow" "[SKIP] $($result.System): $($result.Message)" }
    }
}

$report += "`r`nSummary: Passed=$passed, Failed=$failed, Skipped=$skipped, Total=$($TestResults.Count)"
$report | Out-File -FilePath $ReportFile -Encoding UTF8

Write-Host ""
Write-ColorMessage "Green" "Summary: Passed=$passed, Failed=$failed, Skipped=$skipped, Total=$($TestResults.Count)"
Write-ColorMessage "Green" "Report: $ReportFile"
