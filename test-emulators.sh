#!/bin/bash
#
# EmuBuddy Emulator Test Script
# Downloads small test ROMs and verifies each emulator works
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results arrays
declare -a TESTED_SYSTEMS
declare -a TEST_RESULTS
declare -a TEST_MESSAGES

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DIR="$SCRIPT_DIR/test-roms"
REPORT_FILE="$SCRIPT_DIR/emulator-test-report.txt"

# Create test directory
mkdir -p "$TEST_DIR"

echo "========================================="
echo "  EmuBuddy Emulator Test Suite"
echo "========================================="
echo ""
echo "This script will download small test ROMs"
echo "and verify each emulator launches correctly."
echo ""

# Test ROM definitions - use first small ROM from each system's JSON
# Format: system_id
# ROMs will be loaded from 1g1rsets/{system}.json
TEST_SYSTEMS=(
    "nes"
    "snes"
    "n64"
    "gb"
    "gbc"
    "gba"
    "ds"
    "3ds"
    "gc"
    "wii"
    "genesis"
    "sms"
    "gamegear"
    "dreamcast"
    "psp"
    "ps1"
    "ps2"
    "atari2600"
    "atari7800"
    "lynx"
    "ngpc"
    "wonderswan"
    "wonderswancolor"
    "coleco"
    "intellivision"
    "virtualboy"
    "tg16"
)

# Function to print colored message
print_msg() {
    local color=$1
    local msg=$2
    echo -e "${color}${msg}${NC}" >&2
}

# Function to get smallest clean ROM from JSON
get_test_rom() {
    local system=$1
    local json_file="$SCRIPT_DIR/1g1rsets/${system}.json"

    if [ ! -f "$json_file" ]; then
        print_msg "$RED" "  ROM JSON not found: $json_file"
        return 1
    fi

    # Get first clean ROM (skip Pirate, Proto, Beta, Unl, Sample, Demo)
    local rom_info=$(cat "$json_file" | jq -r '.[] | select(
        (.name | test("(Pirate|Proto|Beta|Unl|Sample|Demo|Test)"; "i")) | not
    ) | "\(.name)|\(.url)"' 2>/dev/null | head -1)

    if [ -z "$rom_info" ]; then
        # Fallback to first ROM if no clean ones found
        rom_info=$(cat "$json_file" | jq -r '.[0] | "\(.name)|\(.url)"' 2>/dev/null)
    fi

    if [ -z "$rom_info" ]; then
        print_msg "$RED" "  Failed to parse ROM JSON"
        return 1
    fi

    echo "$rom_info"
}

# Function to download test ROM
download_rom() {
    local system=$1
    local install_dir=$2

    local rom_info=$(get_test_rom "$system")
    if [ $? -ne 0 ]; then
        return 1
    fi

    IFS='|' read -r rom_name rom_url <<< "$rom_info"

    local rom_file="$install_dir/roms/${system}/$rom_name"
    local rom_dir="$install_dir/roms/${system}"

    mkdir -p "$rom_dir"

    if [ -f "$rom_file" ]; then
        print_msg "$YELLOW" "  ROM already exists: $rom_name"
        echo "$rom_file"
        return 0
    fi

    print_msg "$BLUE" "  Downloading: $rom_name"
    if wget -q --show-progress -O "$rom_file" "$rom_url" 2>&1; then
        print_msg "$GREEN" "  ✓ Downloaded successfully"
        echo "$rom_file"
        return 0
    else
        print_msg "$RED" "  ✗ Download failed"
        rm -f "$rom_file"
        return 1
    fi
}

# Function to get emulator path for system
get_emulator_info() {
    local system=$1
    local base_dir="$2"

    # Determine RetroArch path based on OS
    local retroarch_path=""
    if [[ "$OSTYPE" == "darwin"* ]]; then
        retroarch_path="RetroArch/RetroArch.app/Contents/MacOS/RetroArch"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        retroarch_path="RetroArch/RetroArch-Linux-x86_64/RetroArch-Linux-x86_64.AppImage"
    else
        retroarch_path="RetroArch/RetroArch-Win64/retroarch.exe"
    fi

    case "$system" in
        # RetroArch systems
        nes|snes|n64|gb|gbc|gba|genesis|sms|gamegear|tg16|atari2600|atari7800|coleco|intellivision|virtualboy|psp|ps1|lynx|ngpc|wonderswan|wonderswancolor)
            echo "$retroarch_path"
            ;;

        # Standalone emulators
        ds)
            if [[ "$OSTYPE" == "darwin"* ]]; then
                echo "melonDS/melonDS.app/Contents/MacOS/melonDS"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                echo "melonDS/melonDS-x86_64.AppImage"
            else
                echo "melonDS/melonDS.exe"
            fi
            ;;
        3ds)
            if [[ "$OSTYPE" == "darwin"* ]]; then
                echo "Citra/Citra.app/Contents/MacOS/citra-qt"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                echo "Citra/citra-qt.AppImage"
            else
                echo "Citra/citra-qt.exe"
            fi
            ;;
        gc|wii)
            if [[ "$OSTYPE" == "darwin"* ]]; then
                echo "Dolphin/Dolphin.app/Contents/MacOS/Dolphin"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                echo "Dolphin/dolphin-emu.AppImage"
            else
                echo "Dolphin/Dolphin.exe"
            fi
            ;;
        wiiu)
            if [[ "$OSTYPE" == "darwin"* ]]; then
                echo "Cemu/Cemu.app/Contents/MacOS/Cemu"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                echo "Cemu/Cemu-x86_64.AppImage"
            else
                echo "Cemu/Cemu.exe"
            fi
            ;;
        ps2)
            if [[ "$OSTYPE" == "darwin"* ]]; then
                echo "PCSX2/PCSX2.app/Contents/MacOS/PCSX2-qt"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                echo "PCSX2/pcsx2-qt.AppImage"
            else
                echo "PCSX2/pcsx2-qt.exe"
            fi
            ;;
        dreamcast)
            if [[ "$OSTYPE" == "darwin"* ]]; then
                echo "Flycast/flycast.app/Contents/MacOS/flycast"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                echo "Flycast/flycast.AppImage"
            else
                echo "Flycast/flycast.exe"
            fi
            ;;
        *)
            echo ""
            ;;
    esac
}

# Function to get core for RetroArch
get_core_for_system() {
    local system=$1
    local base_dir="$2"

    local core_ext=".dll"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        core_ext=".dylib"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        core_ext=".so"
    fi

    case "$system" in
        # Nintendo
        nes) echo "cores/fceumm_libretro${core_ext}" ;;
        snes) echo "cores/snes9x_libretro${core_ext}" ;;
        n64) echo "cores/parallel_n64_libretro${core_ext}" ;;
        gb|gbc) echo "cores/gambatte_libretro${core_ext}" ;;
        gba) echo "cores/mgba_libretro${core_ext}" ;;

        # Sega
        genesis) echo "cores/genesis_plus_gx_libretro${core_ext}" ;;
        sms) echo "cores/genesis_plus_gx_libretro${core_ext}" ;;
        gamegear) echo "cores/genesis_plus_gx_libretro${core_ext}" ;;
        dreamcast) echo "cores/flycast_libretro${core_ext}" ;;

        # Sony
        psp) echo "cores/ppsspp_libretro${core_ext}" ;;
        ps1) echo "cores/beetle_psx_libretro${core_ext}" ;;

        # Atari
        atari2600) echo "cores/stella_libretro${core_ext}" ;;
        atari7800) echo "cores/prosystem_libretro${core_ext}" ;;
        lynx) echo "cores/handy_libretro${core_ext}" ;;

        # Other
        tg16) echo "cores/mednafen_pce_libretro${core_ext}" ;;
        virtualboy) echo "cores/beetle_vb_libretro${core_ext}" ;;
        coleco) echo "cores/bluemsx_libretro${core_ext}" ;;
        intellivision) echo "cores/freeintv_libretro${core_ext}" ;;
        ngpc) echo "cores/mednafen_ngp_libretro${core_ext}" ;;
        wonderswan|wonderswancolor) echo "cores/mednafen_wswan_libretro${core_ext}" ;;

        *) echo "" ;;
    esac
}

# Function to launch emulator and test
test_emulator() {
    local system=$1
    local rom_path=$2
    local base_dir=$3

    print_msg "$BLUE" "\n[Testing $system]"

    # Use the launcher's CLI mode
    local launcher_path="$base_dir/EmuBuddyLauncher-linux"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        launcher_path="$base_dir/EmuBuddyLauncher-macos"
    elif [[ "$OSTYPE" != "linux-gnu"* ]]; then
        launcher_path="$base_dir/EmuBuddyLauncher.exe"
    fi

    if [ ! -f "$launcher_path" ]; then
        print_msg "$RED" "  ✗ Launcher not found: $launcher_path"
        TESTED_SYSTEMS+=("$system")
        TEST_RESULTS+=("SKIP")
        TEST_MESSAGES+=("Launcher not installed")
        return 1
    fi

    print_msg "$YELLOW" "  Launching with EmuBuddy launcher..."
    print_msg "$YELLOW" "  (Close the emulator when you've verified it works)"

    # Launch using the launcher CLI
    "$launcher_path" --launch "$system" "$rom_path"
    local exit_code=$?

    # Ask user if it worked
    print_msg "$YELLOW" "\n  Did the emulator launch and display the game correctly? (y/n): "
    read -r response

    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_msg "$GREEN" "  ✓ Test PASSED"
        TESTED_SYSTEMS+=("$system")
        TEST_RESULTS+=("PASS")
        TEST_MESSAGES+=("Emulator launched successfully")
    else
        print_msg "$RED" "  ✗ Test FAILED"
        TESTED_SYSTEMS+=("$system")
        TEST_RESULTS+=("FAIL")
        print_msg "$YELLOW" "  Please describe the issue (or press Enter to skip): "
        read -r issue
        if [ -z "$issue" ]; then
            TEST_MESSAGES+=("User reported failure")
        else
            TEST_MESSAGES+=("$issue")
        fi
    fi
}

# Main test loop
main() {
    # Detect installation directory
    local install_dir=""

    if [ -d "$HOME/EmuBuddy-Linux-v1.0.0" ]; then
        install_dir="$HOME/EmuBuddy-Linux-v1.0.0"
    elif [ -d "$HOME/EmuBuddy-Windows-v1.0.0" ]; then
        install_dir="$HOME/EmuBuddy-Windows-v1.0.0"
    elif [ -d "$HOME/EmuBuddy-macOS-v1.0.0" ]; then
        install_dir="$HOME/EmuBuddy-macOS-v1.0.0"
    else
        print_msg "$RED" "Error: Could not find EmuBuddy installation directory"
        exit 1
    fi

    print_msg "$GREEN" "Found installation: $install_dir\n"

    # Download and test each system
    for system in "${TEST_SYSTEMS[@]}"; do
        print_msg "$BLUE" "\n========================================="
        print_msg "$BLUE" "System: $system"
        print_msg "$BLUE" "========================================="

        # Download ROM
        rom_path=$(download_rom "$system" "$install_dir")

        if [ $? -eq 0 ]; then
            # Test emulator
            test_emulator "$system" "$rom_path" "$install_dir"
        else
            TESTED_SYSTEMS+=("$system")
            TEST_RESULTS+=("SKIP")
            TEST_MESSAGES+=("ROM download failed")
        fi

        sleep 1
    done

    # Generate report
    print_msg "$GREEN" "\n\n========================================="
    print_msg "$GREEN" "       TEST REPORT"
    print_msg "$GREEN" "========================================="

    {
        echo "EmuBuddy Emulator Test Report"
        echo "Generated: $(date)"
        echo "Installation: $install_dir"
        echo "Platform: $OSTYPE"
        echo ""
        echo "========================================="
        echo ""

        local passed=0
        local failed=0
        local skipped=0

        for i in "${!TESTED_SYSTEMS[@]}"; do
            local system="${TESTED_SYSTEMS[$i]}"
            local result="${TEST_RESULTS[$i]}"
            local message="${TEST_MESSAGES[$i]}"

            printf "%-20s : %-6s : %s\n" "$system" "$result" "$message"

            case "$result" in
                PASS) ((passed++)) ;;
                FAIL) ((failed++)) ;;
                SKIP) ((skipped++)) ;;
            esac
        done

        echo ""
        echo "========================================="
        echo "Summary:"
        echo "  Passed:  $passed"
        echo "  Failed:  $failed"
        echo "  Skipped: $skipped"
        echo "  Total:   ${#TESTED_SYSTEMS[@]}"
        echo "========================================="

    } | tee "$REPORT_FILE"

    print_msg "$GREEN" "\nReport saved to: $REPORT_FILE"

    # Show summary with colors
    echo ""
    for i in "${!TESTED_SYSTEMS[@]}"; do
        local system="${TESTED_SYSTEMS[$i]}"
        local result="${TEST_RESULTS[$i]}"
        local message="${TEST_MESSAGES[$i]}"

        case "$result" in
            PASS)
                print_msg "$GREEN" "✓ $system: $message"
                ;;
            FAIL)
                print_msg "$RED" "✗ $system: $message"
                ;;
            SKIP)
                print_msg "$YELLOW" "⊘ $system: $message"
                ;;
        esac
    done
}

# Run main function
main

print_msg "$BLUE" "\n\nTest complete! Check $REPORT_FILE for full results."
