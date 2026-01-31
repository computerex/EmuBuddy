package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorRed    = "\033[31m"
)

type EmulatorURL struct {
	Windows string
	Linux   string
	MacOS   string
}

type Emulator struct {
	Name        string
	URLs        EmulatorURL
	ArchiveName map[string]string // platform -> filename
	ExtractDir  string
}

type RetroArchCore struct {
	Name string
	URLs EmulatorURL
}

var emulators = []Emulator{
	{
		Name: "PCSX2 (PS2)",
		URLs: EmulatorURL{
			Windows: "https://github.com/PCSX2/pcsx2/releases/download/v2.2.0/pcsx2-v2.2.0-windows-x64-Qt.7z",
			Linux:   "https://github.com/PCSX2/pcsx2/releases/download/v2.2.0/pcsx2-v2.2.0-linux-appimage-x64-Qt.AppImage",
			MacOS:   "https://github.com/PCSX2/pcsx2/releases/download/v2.2.0/pcsx2-v2.2.0-macos-Qt.tar.xz",
		},
		ArchiveName: map[string]string{
			"windows": "pcsx2.7z",
			"linux":   "pcsx2.AppImage",
			"darwin":  "pcsx2.tar.xz",
		},
		ExtractDir: "PCSX2",
	},
	{
		Name: "PPSSPP (PSP)",
		URLs: EmulatorURL{
			Windows: "https://www.ppsspp.org/files/1_19_3/ppsspp_win.zip",
			Linux:   "https://github.com/hrydgard/ppsspp/releases/download/v1.19.3/PPSSPP-v1.19.3-anylinux-x86_64.AppImage",
			MacOS:   "https://www.ppsspp.org/files/1_19_3/PPSSPP_macOS.dmg",
		},
		ArchiveName: map[string]string{
			"windows": "ppsspp.zip",
			"linux":   "ppsspp.AppImage",
			"darwin":  "ppsspp.dmg",
		},
		ExtractDir: "PPSSPP",
	},
	{
		Name: "Dolphin (GameCube/Wii)",
		URLs: EmulatorURL{
			Windows: "https://dl.dolphin-emu.org/releases/2512/dolphin-2512-x64.7z",
			Linux:   "https://dl.dolphin-emu.org/releases/2512/dolphin-2512-x86_64.flatpak",
			MacOS:   "https://dl.dolphin-emu.org/releases/2512/dolphin-2512-universal.dmg",
		},
		ArchiveName: map[string]string{
			"windows": "dolphin.7z",
			"linux":   "dolphin.flatpak",
			"darwin":  "dolphin.dmg",
		},
		ExtractDir: "Dolphin",
	},
	{
		Name: "melonDS (Nintendo DS)",
		URLs: EmulatorURL{
			Windows: "https://github.com/melonDS-emu/melonDS/releases/download/1.1/melonDS-1.1-windows-x86_64.zip",
			Linux:   "https://github.com/melonDS-emu/melonDS/releases/download/1.1/melonDS-1.1-appimage-x86_64.zip",
			MacOS:   "https://github.com/melonDS-emu/melonDS/releases/download/1.1/melonDS-1.1-macOS-universal.zip",
		},
		ArchiveName: map[string]string{
			"windows": "melonds.zip",
			"linux":   "melonds.zip",
			"darwin":  "melonds.zip",
		},
		ExtractDir: "melonDS",
	},
	{
		Name: "Azahar (Nintendo 3DS)",
		URLs: EmulatorURL{
			Windows: "https://github.com/azahar-emu/azahar/releases/download/2124.3/azahar-2124.3-windows-msvc.zip",
			Linux:   "https://github.com/azahar-emu/azahar/releases/download/2124.3/azahar.AppImage",
			MacOS:   "https://github.com/azahar-emu/azahar/releases/download/2124.3/azahar-2124.3-macos-universal.zip",
		},
		ArchiveName: map[string]string{
			"windows": "azahar.zip",
			"linux":   "azahar.AppImage",
			"darwin":  "azahar.zip",
		},
		ExtractDir: "Azahar",
	},
	{
		Name: "mGBA (Game Boy Advance)",
		URLs: EmulatorURL{
			Windows: "https://github.com/mgba-emu/mgba/releases/download/0.10.5/mGBA-0.10.5-win64.7z",
			Linux:   "https://github.com/mgba-emu/mgba/releases/download/0.10.5/mGBA-0.10.5-appimage-x64.appimage",
			MacOS:   "https://github.com/mgba-emu/mgba/releases/download/0.10.5/mGBA-0.10.5-macos.dmg",
		},
		ArchiveName: map[string]string{
			"windows": "mgba.7z",
			"linux":   "mgba.AppImage",
			"darwin":  "mgba.dmg",
		},
		ExtractDir: "mGBA",
	},
	{
		Name: "RetroArch (Multi-System)",
		URLs: EmulatorURL{
			Windows: "https://buildbot.libretro.com/stable/1.19.1/windows/x86_64/RetroArch.7z",
			Linux:   "https://buildbot.libretro.com/nightly/linux/x86_64/RetroArch.7z", // Use nightly for Linux to fix Wayland compatibility
			MacOS:   "https://buildbot.libretro.com/stable/1.19.1/apple/osx/universal/RetroArch_Metal.dmg",
		},
		ArchiveName: map[string]string{
			"windows": "retroarch.7z",
			"linux":   "retroarch.7z",
			"darwin":  "retroarch.dmg",
		},
		ExtractDir: "RetroArch",
	},
	{
		Name: "Cemu (Wii U)",
		URLs: EmulatorURL{
			Windows: "https://github.com/cemu-project/Cemu/releases/download/v2.6/cemu-2.6-windows-x64.zip",
			Linux:   "https://github.com/cemu-project/Cemu/releases/download/v2.6/Cemu-2.6-x86_64.AppImage",
			MacOS:   "https://github.com/cemu-project/Cemu/releases/download/v2.6/cemu-2.6-macos-12-x64.dmg",
		},
		ArchiveName: map[string]string{
			"windows": "cemu.zip",
			"linux":   "Cemu.AppImage",
			"darwin":  "cemu.dmg",
		},
		ExtractDir: "Cemu",
	},
}

var retroarchCores = EmulatorURL{
	Windows: "https://buildbot.libretro.com/stable/1.19.1/windows/x86_64/RetroArch_cores.7z",
	Linux:   "https://buildbot.libretro.com/nightly/linux/x86_64/RetroArch_cores.7z", // Use nightly for Linux to match RetroArch version
	MacOS:   "", // macOS cores downloaded individually from nightly builds
}

// BIOS files URLs
var retroarchBIOSURL = "https://github.com/Abdess/retroarch_system/releases/download/v20220308/libretro_31-01-22.zip"

// PS2 BIOS - USA version for best compatibility
var ps2BIOSURL = "https://myrient.erista.me/files/Redump/Sony%20-%20PlayStation%202%20-%20BIOS%20Images%20%28DoM%20Version%29/ps2-0220a-20060905-125923.zip"

// Additional cores that need to be downloaded separately (not in the main cores pack)
var additionalCores = []RetroArchCore{
	{
		Name: "Citra (3DS)",
		URLs: EmulatorURL{
			Windows: "https://buildbot.libretro.com/nightly/windows/x86_64/latest/citra_libretro.dll.zip",
			Linux:   "https://buildbot.libretro.com/nightly/linux/x86_64/latest/citra_libretro.so.zip",
			MacOS:   "",
		},
	},
}

// downloadMacOSCores downloads essential RetroArch cores for macOS from the buildbot
func downloadMacOSCores(coresDir, downloadDir string) int {
	// List of essential cores for all supported systems
	essentialCores := []string{
		// NES
		"nestopia_libretro.dylib.zip",
		"fceumm_libretro.dylib.zip",
		"mesen_libretro.dylib.zip",
		// SNES
		"snes9x_libretro.dylib.zip",
		"bsnes_libretro.dylib.zip",
		// N64
		"mupen64plus_next_libretro.dylib.zip",
		"parallel_n64_libretro.dylib.zip",
		// GB/GBC/GBA
		"gambatte_libretro.dylib.zip",
		"sameboy_libretro.dylib.zip",
		"mgba_libretro.dylib.zip",
		"vbam_libretro.dylib.zip",
		"vba_next_libretro.dylib.zip",
		// DS
		"melonds_libretro.dylib.zip",
		"desmume_libretro.dylib.zip",
		// 3DS
		"panda3ds_libretro.dylib.zip",
		// PSP
		"ppsspp_libretro.dylib.zip",
		// PS1
		"mednafen_psx_hw_libretro.dylib.zip",
		"pcsx_rearmed_libretro.dylib.zip",
		"swanstation_libretro.dylib.zip",
		// Dreamcast
		"flycast_libretro.dylib.zip",
		// Genesis/MD
		"genesis_plus_gx_libretro.dylib.zip",
		"picodrive_libretro.dylib.zip",
		"blastem_libretro.dylib.zip",
		// Game Gear
		"gearsystem_libretro.dylib.zip",
		// TurboGrafx-16
		"mednafen_pce_fast_libretro.dylib.zip",
		"mednafen_supergrafx_libretro.dylib.zip",
		// Virtual Boy
		"mednafen_vb_libretro.dylib.zip",
		// Atari
		"stella_libretro.dylib.zip",
		"stella2014_libretro.dylib.zip",
		"prosystem_libretro.dylib.zip",
		"handy_libretro.dylib.zip",
		"mednafen_lynx_libretro.dylib.zip",
		// Neo Geo Pocket
		"mednafen_ngp_libretro.dylib.zip",
		"race_libretro.dylib.zip",
		// ColecoVision
		"bluemsx_libretro.dylib.zip",
		"gearcoleco_libretro.dylib.zip",
		// Intellivision
		"freeintv_libretro.dylib.zip",
		// WonderSwan
		"mednafen_wswan_libretro.dylib.zip",
	}

	baseURL := "https://buildbot.libretro.com/nightly/apple/osx/x86_64/latest"
	downloadedCount := 0
	total := len(essentialCores)

	printInfo(fmt.Sprintf("Downloading %d cores...", total))

	for i, coreZip := range essentialCores {
		coreName := strings.TrimSuffix(coreZip, "_libretro.dylib.zip")
		coreURL := fmt.Sprintf("%s/%s", baseURL, coreZip)
		coreArchive := filepath.Join(downloadDir, coreZip)

		fmt.Printf("\r  [%d/%d] %s...", i+1, total, coreName)

		// Download core
		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("GET", coreURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		out, err := os.Create(coreArchive)
		if err != nil {
			resp.Body.Close()
			continue
		}

		_, err = io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()

		if err != nil {
			continue
		}

		// Extract core
		if err := extractZipToDir(coreArchive, coresDir); err == nil {
			downloadedCount++
		}
		os.Remove(coreArchive)
	}

	fmt.Println()
	return downloadedCount
}

func main() {
	printHeader()

	// Detect OS
	platform := runtime.GOOS
	platformName := getPlatformName(platform)

	printInfo(fmt.Sprintf("Detected platform: %s", platformName))
	fmt.Println()

	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		printError("Failed to get executable path: " + err.Error())
		waitForExit(1)
		return
	}
	baseDir := filepath.Dir(exePath)

	// Create necessary directories
	emuDir := filepath.Join(baseDir, "Emulators")
	downloadDir := filepath.Join(baseDir, "Downloads")

	for _, dir := range []string{emuDir, downloadDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			printError("Failed to create directory " + dir + ": " + err.Error())
			waitForExit(1)
			return
		}
	}

	// Download and setup 7-Zip for all platforms
	printSection("Step 1: Setting up 7-Zip")
	extractorPath := get7ZipPath(baseDir)
	if !fileExists(extractorPath) {
		if err := setup7Zip(baseDir); err != nil {
			printError("Failed to setup 7-Zip: " + err.Error())
			waitForExit(1)
			return
		}
	} else {
		printSuccess("7-Zip already installed")
	}
	
	// On non-Windows, also check for tar (needed for .tar.xz files)
	if platform != "windows" {
		if !commandExists("tar") {
			printError("'tar' command not found. Please install tar utilities.")
			waitForExit(1)
			return
		}
	}

	// Download emulators
	printSection("Step 2: Downloading Emulators")
	installedCount := 0
	skippedCount := 0
	failedEmulators := []string{}
	linuxManualInstalls := []string{}

	for i, emu := range emulators {
		fmt.Printf("[%d/%d] %s\n", i+1, len(emulators), emu.Name)

		// Get platform-specific URL
		url := getURLForPlatform(emu.URLs, platform)
		if url == "" {
			printWarning("  Not available for " + platformName)
			if platform == "linux" {
				linuxManualInstalls = append(linuxManualInstalls, emu.Name)
			} else {
				failedEmulators = append(failedEmulators, emu.Name)
			}
			skippedCount++
			continue
		}

		archiveName := emu.ArchiveName[platform]
		downloadPath := filepath.Join(downloadDir, archiveName)
		extractPath := filepath.Join(emuDir, emu.ExtractDir)

		// Skip if already extracted/installed
		skipInstall := false
		if platform == "darwin" && strings.HasSuffix(archiveName, ".dmg") {
			// For DMG files, check if .app bundle exists inside the directory
			if fileExists(extractPath) {
				entries, err := os.ReadDir(extractPath)
				if err == nil {
					for _, entry := range entries {
						if strings.HasSuffix(entry.Name(), ".app") {
							skipInstall = true
							break
						}
					}
				}
			}
		} else if platform == "linux" && strings.HasSuffix(archiveName, ".AppImage") {
			// For AppImage files, check if the .AppImage file exists inside the directory
			if fileExists(extractPath) {
				entries, err := os.ReadDir(extractPath)
				if err == nil {
					for _, entry := range entries {
						if strings.HasSuffix(strings.ToLower(entry.Name()), ".appimage") {
							skipInstall = true
							break
						}
					}
				}
			}
		} else if platform == "linux" && strings.HasSuffix(archiveName, ".7z") && emu.ExtractDir == "RetroArch" {
			// For Linux RetroArch, check if the extracted directory structure exists
			retroarchBinary := filepath.Join(extractPath, "RetroArch-Linux-x86_64", "retroarch")
			if fileExists(retroarchBinary) {
				skipInstall = true
			}
		} else if fileExists(extractPath) {
			// For other archives, check if the extract directory has content
			entries, err := os.ReadDir(extractPath)
			if err == nil && len(entries) > 0 {
				skipInstall = true
			}
		}

		if skipInstall {
			printInfo("  Already installed, skipping...")
			installedCount++
			continue
		}

		// Download
		if !fileExists(downloadPath) {
			printInfo("  Downloading...")
			if err := downloadFile(url, downloadPath); err != nil {
				printWarning("  Download failed: " + err.Error())
				printWarning("  Skipping " + emu.Name)
				failedEmulators = append(failedEmulators, emu.Name)
				continue
			}
		} else {
			printInfo("  Archive already downloaded")
		}

		// Extract/Install based on file type
		printInfo("  Installing...")
		if err := extractFile(extractorPath, downloadPath, extractPath, platform); err != nil {
			printWarning("  Installation failed: " + err.Error())
			failedEmulators = append(failedEmulators, emu.Name)
			continue
		}

		printSuccess("  ✓ Installed")
		installedCount++
	}

	fmt.Println()
	printInfo(fmt.Sprintf("Successfully installed: %d/%d emulators", installedCount, len(emulators)))
	if skippedCount > 0 {
		printInfo(fmt.Sprintf("Skipped (already installed): %d", skippedCount-len(failedEmulators)))
	}
	if len(failedEmulators) > 0 {
		printWarning("Failed to install: " + strings.Join(failedEmulators, ", "))
	}
	if len(linuxManualInstalls) > 0 {
		fmt.Println()
		printInfo("Linux: Some emulators downloaded as Flatpak packages:")
		printInfo("  Install with: flatpak install <path-to-flatpak-file>")
	}

	// Download RetroArch cores
	printSection("Step 3: Downloading RetroArch Cores")

	// Determine cores directory based on platform
	var coresDir string
	if platform == "darwin" {
		// macOS stores cores in ~/Library/Application Support/RetroArch/cores/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			printWarning("Failed to get home directory: " + err.Error())
		} else {
			coresDir = filepath.Join(homeDir, "Library", "Application Support", "RetroArch", "cores")
		}
	} else if platform == "linux" {
		coresDir = filepath.Join(emuDir, "RetroArch", "RetroArch-Linux-x86_64", "cores")
	} else {
		coresDir = filepath.Join(emuDir, "RetroArch", "RetroArch-Win64", "cores")
	}

	if coresDir != "" {
		os.MkdirAll(coresDir, 0755)

		// Check if cores already exist
		entries, _ := os.ReadDir(coresDir)
		hasCores := false
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".dll") || strings.HasSuffix(entry.Name(), ".so") || strings.HasSuffix(entry.Name(), ".dylib") {
				hasCores = true
				break
			}
		}

		if !hasCores {
			if platform == "darwin" {
				// macOS: Download individual cores from nightly builds
				printInfo("Downloading essential cores for macOS...")
				downloadedCount := downloadMacOSCores(coresDir, downloadDir)
				if downloadedCount > 0 {
					printSuccess(fmt.Sprintf("✓ %d RetroArch cores installed", downloadedCount))
				} else {
					printWarning("Failed to download cores")
				}
			} else {
				// Windows/Linux: Download cores bundle
				coresURL := getURLForPlatform(retroarchCores, platform)
				if coresURL != "" {
					coresArchive := filepath.Join(downloadDir, "RetroArch_cores.7z")
					printInfo("Downloading RetroArch cores package...")
					if err := downloadFile(coresURL, coresArchive); err != nil {
						printWarning("Failed to download cores: " + err.Error())
					} else {
						printInfo("Extracting cores...")
						retroarchDir := filepath.Join(emuDir, "RetroArch")
						if err := extractFile(extractorPath, coresArchive, retroarchDir, platform); err != nil {
							printWarning("Failed to extract cores: " + err.Error())
						} else {
							// On Linux, both the main RetroArch 7z and cores 7z extract to nested structures
							// Move all cores from nested locations to our portable cores directory
							if platform == "linux" {
								// Check both possible nested locations
								nestedPaths := []string{
									filepath.Join(retroarchDir, "RetroArch-Linux-x86_64.AppImage.home", ".config", "retroarch", "cores"),
									filepath.Join(retroarchDir, "RetroArch-Linux-x86_64", "RetroArch-Linux-x86_64.AppImage.home", ".config", "retroarch", "cores"),
								}
								
								for _, nestedCoresDir := range nestedPaths {
									if entries, err := os.ReadDir(nestedCoresDir); err == nil && len(entries) > 0 {
										printInfo("Moving cores to portable location...")
										for _, entry := range entries {
											if strings.HasSuffix(entry.Name(), ".so") {
												srcPath := filepath.Join(nestedCoresDir, entry.Name())
												dstPath := filepath.Join(coresDir, entry.Name())
												// Only move if not already there
												if _, err := os.Stat(dstPath); os.IsNotExist(err) {
													os.Rename(srcPath, dstPath)
												}
											}
										}
									}
								}
								
								// Clean up nested directory structures
								os.RemoveAll(filepath.Join(retroarchDir, "RetroArch-Linux-x86_64.AppImage.home"))
								os.RemoveAll(filepath.Join(retroarchDir, "RetroArch-Linux-x86_64", "RetroArch-Linux-x86_64.AppImage.home"))
							}
							printSuccess("✓ RetroArch cores installed")
						}
					}
				}
			}
		} else {
			printSuccess("RetroArch cores already installed")
		}
	}

	// Download additional cores (like Citra) that aren't in the main pack
	if platform != "darwin" && coresDir != "" {
		printInfo("Downloading additional cores...")

		for _, core := range additionalCores {
			coreURL := getURLForPlatform(core.URLs, platform)
			if coreURL == "" {
				continue
			}

			// Determine expected dll/so name from URL
			coreName := filepath.Base(coreURL)
			coreName = strings.TrimSuffix(coreName, ".zip")
			coreFile := filepath.Join(coresDir, coreName)

			if fileExists(coreFile) {
				printSuccess(fmt.Sprintf("  ✓ %s already installed", core.Name))
				continue
			}

			printInfo(fmt.Sprintf("  Downloading %s core...", core.Name))
			coreArchive := filepath.Join(downloadDir, filepath.Base(coreURL))
			if err := downloadFile(coreURL, coreArchive); err != nil {
				printWarning(fmt.Sprintf("  Failed to download %s: %s", core.Name, err.Error()))
				continue
			}

			// Extract the core zip directly to cores folder
			if err := extractZipToDir(coreArchive, coresDir); err != nil {
				printWarning(fmt.Sprintf("  Failed to extract %s: %s", core.Name, err.Error()))
			} else {
				printSuccess(fmt.Sprintf("  ✓ %s core installed", core.Name))
			}
		}
	}

	// Download BIOS files
	printSection("Step 4: Downloading BIOS Files")
	biosDir := filepath.Join(emuDir, "RetroArch", "RetroArch-Win64", "system")
	if platform == "linux" {
		biosDir = filepath.Join(emuDir, "RetroArch", "RetroArch-Linux-x86_64", "system")
	} else if platform == "darwin" {
		biosDir = filepath.Join(emuDir, "RetroArch", "system")
	}
	os.MkdirAll(biosDir, 0755)

	// Download RetroArch system/BIOS files
	printInfo("Downloading RetroArch BIOS/System files...")
	retroarchBiosArchive := filepath.Join(downloadDir, "retroarch_bios.zip")
	if !fileExists(retroarchBiosArchive) {
		if err := downloadFile(retroarchBIOSURL, retroarchBiosArchive); err != nil {
			printWarning("Failed to download RetroArch BIOS: " + err.Error())
		} else {
			printInfo("Extracting RetroArch BIOS files...")
			if err := extractZip(retroarchBiosArchive, biosDir); err != nil {
				printWarning("Failed to extract RetroArch BIOS: " + err.Error())
			} else {
				printSuccess("✓ RetroArch BIOS files installed")
				// Copy BIOS files from subfolders to main system folder for core compatibility
				copyBIOSFilesToSystemFolder(biosDir)
			}
		}
	} else {
		printSuccess("RetroArch BIOS files already downloaded")
	}

	// Download PS2 BIOS
	printInfo("Downloading PS2 BIOS...")
	ps2BiosArchive := filepath.Join(downloadDir, "ps2_bios.zip")
	pcsx2BiosDir := filepath.Join(emuDir, "PCSX2", "bios")
	os.MkdirAll(pcsx2BiosDir, 0755)

	if !fileExists(ps2BiosArchive) {
		if err := downloadFromMyrient(ps2BIOSURL, ps2BiosArchive); err != nil {
			printWarning("Failed to download PS2 BIOS: " + err.Error())
		} else {
			printInfo("Extracting PS2 BIOS files...")
			if err := extractZip(ps2BiosArchive, pcsx2BiosDir); err != nil {
				printWarning("Failed to extract PS2 BIOS: " + err.Error())
			} else {
				printSuccess("✓ PS2 BIOS files installed")
			}
		}
	} else {
		printSuccess("PS2 BIOS files already downloaded")
	}

	// Configure PCSX2 to use the BIOS directory (portable mode)
	// On Linux, PCSX2 AppImage also supports portable mode with portable.txt
	printInfo("Configuring PCSX2...")
	if err := configurePCSX2(emuDir, pcsx2BiosDir, platform); err != nil {
		printWarning("Failed to configure PCSX2: " + err.Error())
	} else {
		printSuccess("✓ PCSX2 configured")
	}

	// Configure RetroArch system directory
	printInfo("Configuring RetroArch...")
	if err := configureRetroArch(emuDir, biosDir, platform); err != nil {
		printWarning("Failed to configure RetroArch: " + err.Error())
	} else {
		printSuccess("✓ RetroArch configured")
	}

	// Cleanup
	printSection("Step 5: Cleanup")
	printInfo("Removing downloaded archives...")
	os.RemoveAll(downloadDir)
	printSuccess("✓ Cleanup complete")

	// Final summary
	fmt.Println()
	if len(failedEmulators) == 0 && len(linuxManualInstalls) == 0 {
		printSuccess("═══════════════════════════════════════")
		printSuccess("  Installation Complete!")
		printSuccess("═══════════════════════════════════════")
		fmt.Println()
		printInfo("All emulators installed successfully!")
	} else {
		printSuccess("═══════════════════════════════════════")
		printSuccess("  Installation Mostly Complete!")
		printSuccess("═══════════════════════════════════════")
		fmt.Println()
		printInfo(fmt.Sprintf("Installed %d/%d emulators successfully.", installedCount, len(emulators)))
		if len(linuxManualInstalls) > 0 {
			printInfo("Some emulators require manual installation (see above).")
		}
		if len(failedEmulators) > 0 {
			printWarning("Some emulators failed to download.")
		}
	}

	fmt.Println()
	printInfo("Next steps:")
	if platform == "windows" {
		printInfo("  Launching EmuBuddy...")
	} else if platform == "linux" {
		printInfo("  Run: ./start-emubuddy.sh")
		printInfo("  Or double-click EmuBuddyLauncher-linux")
	} else if platform == "darwin" {
		printInfo("  Double-click 'Start EmuBuddy.command'")
		printInfo("  Or run: ./EmuBuddyLauncher-macos")
	}
	fmt.Println()

	// Launch the GUI on Windows
	if platform == "windows" {
		launcherPath := filepath.Join(baseDir, "EmuBuddyLauncher.exe")
		if fileExists(launcherPath) {
			exec.Command(launcherPath).Start()
		}
	}

	os.Exit(0)
}

func getPlatformName(platform string) string {
	switch platform {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	default:
		return platform
	}
}

func getURLForPlatform(urls EmulatorURL, platform string) string {
	switch platform {
	case "windows":
		return urls.Windows
	case "linux":
		return urls.Linux
	case "darwin":
		return urls.MacOS
	default:
		return ""
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func printHeader() {
	platform := getPlatformName(runtime.GOOS)
	fmt.Println(colorCyan + "╔═══════════════════════════════════════╗" + colorReset)
	fmt.Println(colorCyan + "║   EmuBuddy Installer v2.1            ║" + colorReset)
	fmt.Println(colorCyan + "║   Cross-Platform Edition             ║" + colorReset)
	fmt.Println(colorCyan + "╚═══════════════════════════════════════╝" + colorReset)
	fmt.Println()
	fmt.Println("Platform: " + platform)
	fmt.Println()
	fmt.Println("This installer will download and set up:")
	fmt.Println("  • 8 Emulators (~375 MB)")
	fmt.Println("  • RetroArch Cores (~468 MB)")
	fmt.Println("  • BIOS Files (~600 MB)")
	fmt.Println("  • Total download: ~1.4 GB")
	fmt.Println()
	fmt.Println("Press Ctrl+C to cancel, or Enter to continue...")
	fmt.Scanln()
	fmt.Println()
}

func printSection(title string) {
	fmt.Println()
	fmt.Println(colorCyan + "═══════════════════════════════════════" + colorReset)
	fmt.Println(colorCyan + "  " + title + colorReset)
	fmt.Println(colorCyan + "═══════════════════════════════════════" + colorReset)
}

func printSuccess(msg string) {
	fmt.Println(colorGreen + msg + colorReset)
}

func printInfo(msg string) {
	fmt.Println(msg)
}

func printWarning(msg string) {
	fmt.Println(colorYellow + msg + colorReset)
}

func printError(msg string) {
	fmt.Println(colorRed + "ERROR: " + msg + colorReset)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func downloadFile(url, destPath string) error {
	return downloadFileWithReferer(url, destPath, "")
}

// downloadFileWithReferer downloads a file with an optional Referer header
func downloadFileWithReferer(url, destPath, referer string) error {
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	// Set headers to avoid rate limiting
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	totalSize := resp.ContentLength
	downloaded := int64(0)
	lastPrint := time.Now()

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if time.Since(lastPrint) > time.Second {
				if totalSize > 0 {
					pct := float64(downloaded) / float64(totalSize) * 100
					fmt.Printf("\r  Progress: %.1f%% (%s / %s)", pct, formatBytes(downloaded), formatBytes(totalSize))
				} else {
					fmt.Printf("\r  Downloaded: %s", formatBytes(downloaded))
				}
				lastPrint = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	fmt.Println()
	return nil
}

// downloadFromMyrient downloads a file from Myrient with proper headers to avoid rate limiting
func downloadFromMyrient(url, destPath string) error {
	return downloadFileWithReferer(url, destPath, "https://myrient.erista.me/")
}

func extractFile(extractorPath, archivePath, destDir string, platform string) error {
	ext := filepath.Ext(archivePath)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, destDir)

	case strings.HasSuffix(archivePath, ".7z"):
		// Use our bundled 7-Zip on all platforms
		return extract7z(extractorPath, archivePath, destDir)

	case strings.HasSuffix(archivePath, ".tar.xz"):
		return extractTarXz(archivePath, destDir)

	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGz(archivePath, destDir)

	case strings.HasSuffix(archivePath, ".AppImage") || strings.HasSuffix(archivePath, ".appimage"):
		// Make AppImage executable and move to destination
		if err := os.Chmod(archivePath, 0755); err != nil {
			return err
		}
		finalPath := filepath.Join(destDir, filepath.Base(archivePath))
		return os.Rename(archivePath, finalPath)

	case strings.HasSuffix(archivePath, ".dmg"):
		return extractDMG(archivePath, destDir)

	case strings.HasSuffix(archivePath, ".flatpak"):
		printInfo("  Flatpak downloaded. Install with: flatpak install " + archivePath)
		return nil

	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

func extract7z(sevenZipPath, archivePath, destDir string) error {
	cmd := exec.Command(sevenZipPath, "x", archivePath, "-o"+destDir, "-y")
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extractZipToDir extracts a zip file directly to destDir without stripping root folders
// Used for simple core zip files that contain just the dll/so file
func extractZipToDir(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(destDir, 0755)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		
		// Just use the base filename, ignore any folder structure in the zip
		destPath := filepath.Join(destDir, filepath.Base(f.Name))
		
		rc, err := f.Open()
		if err != nil {
			return err
		}
		
		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return err
		}
		
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		
		if err != nil {
			return err
		}
	}
	return nil
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Clean and normalize destDir for consistent path handling on Windows
	destDir = filepath.Clean(destDir)

	// Ensure destination directory exists first
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %v", destDir, err)
	}

	// Find if there's a common root folder
	var rootFolder string
	if len(r.File) > 0 {
		// Check first file's path
		firstPath := r.File[0].Name
		parts := strings.Split(filepath.ToSlash(firstPath), "/")
		if len(parts) > 1 {
			// Potential root folder
			potentialRoot := parts[0] + "/"
			hasRoot := true
			for _, f := range r.File {
				if !strings.HasPrefix(filepath.ToSlash(f.Name), potentialRoot) {
					hasRoot = false
					break
				}
			}
			if hasRoot {
				rootFolder = potentialRoot
			}
		}
	}

	// Helper function to check if a ZIP entry is a directory
	isDir := func(f *zip.File) bool {
		// Check the mode flag
		if f.FileInfo().IsDir() {
			return true
		}
		// Also check for trailing slash (some ZIPs mark dirs this way)
		if strings.HasSuffix(f.Name, "/") || strings.HasSuffix(f.Name, "\\") {
			return true
		}
		// Check if uncompressed size is 0 and name looks like a directory
		if f.UncompressedSize64 == 0 && !strings.Contains(filepath.Base(f.Name), ".") {
			return true
		}
		return false
	}

	// First pass: collect all directories that need to be created
	dirsToCreate := make(map[string]bool)
	for _, f := range r.File {
		name := filepath.ToSlash(f.Name)
		if rootFolder != "" {
			name = strings.TrimPrefix(name, rootFolder)
		}
		if name == "" {
			continue
		}

		// Remove trailing slashes before processing
		name = strings.TrimSuffix(name, "/")
		if name == "" {
			continue
		}

		// Security check
		if strings.Contains(name, "..") {
			continue
		}

		// Convert to OS-specific path and clean it
		name = filepath.Clean(filepath.FromSlash(name))
		fpath := filepath.Join(destDir, name)

		if isDir(f) {
			dirsToCreate[fpath] = true
		} else {
			// Add parent directory
			parentDir := filepath.Dir(fpath)
			if parentDir != destDir {
				dirsToCreate[parentDir] = true
			}
		}
	}

	// Create all directories upfront, sorted by depth (shortest paths first)
	var sortedDirs []string
	for dir := range dirsToCreate {
		sortedDirs = append(sortedDirs, dir)
	}
	// Sort by path length to ensure parent dirs are created first
	for i := 0; i < len(sortedDirs); i++ {
		for j := i + 1; j < len(sortedDirs); j++ {
			if len(sortedDirs[i]) > len(sortedDirs[j]) {
				sortedDirs[i], sortedDirs[j] = sortedDirs[j], sortedDirs[i]
			}
		}
	}

	for _, dir := range sortedDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %v", dir, err)
		}
	}

	// Second pass: extract files
	for _, f := range r.File {
		// Skip directories (already created)
		if isDir(f) {
			continue
		}

		name := filepath.ToSlash(f.Name)
		if rootFolder != "" {
			name = strings.TrimPrefix(name, rootFolder)
		}

		if name == "" {
			continue
		}

		// Security check
		if strings.Contains(name, "..") {
			continue
		}

		// Convert to OS-specific path and clean it
		name = filepath.Clean(filepath.FromSlash(name))
		fpath := filepath.Join(destDir, name)

		// Create file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create file %s: %v", fpath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, copyErr := io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if copyErr != nil {
			return fmt.Errorf("write file %s: %v", fpath, copyErr)
		}
		
		// Make AppImage files executable
		if strings.HasSuffix(strings.ToLower(fpath), ".appimage") {
			os.Chmod(fpath, 0755)
		}
	}

	return nil
}

func moveDir(src, dst string) error {
	// Ensure destination exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := moveDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := os.Rename(srcPath, dstPath); err != nil {
				// If rename fails, try copy
				if copyErr := copyFile(srcPath, dstPath); copyErr != nil {
					return copyErr
				}
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func extractTarXz(tarXzPath, destDir string) error {
	f, err := os.Open(tarXzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	xzReader, err := xz.NewReader(f)
	if err != nil {
		return err
	}

	return extractTar(xzReader, destDir)
}

func extractTarGz(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	return extractTar(gzReader, destDir)
}

func extractTar(reader io.Reader, destDir string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// extractDMG mounts a DMG file, copies the .app bundle to destDir, and unmounts
func extractDMG(dmgPath, destDir string) error {
	// Create temp mount point
	mountPoint := filepath.Join(os.TempDir(), fmt.Sprintf("emubuddy_mount_%d", time.Now().Unix()))
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}
	defer os.RemoveAll(mountPoint)

	// Mount the DMG
	cmd := exec.Command("hdiutil", "attach", dmgPath, "-mountpoint", mountPoint, "-nobrowse", "-quiet")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount DMG: %v", err)
	}

	// Ensure unmount on exit
	defer func() {
		exec.Command("hdiutil", "detach", mountPoint, "-quiet").Run()
	}()

	// Find .app bundles in the mounted volume
	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		return fmt.Errorf("failed to read mount point: %v", err)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination: %v", err)
	}

	// Copy all .app bundles
	copied := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".app") {
			srcPath := filepath.Join(mountPoint, entry.Name())
			dstPath := filepath.Join(destDir, entry.Name())

			// Copy the .app bundle recursively
			if err := copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %v", entry.Name(), err)
			}
			copied = true
		}
	}

	if !copied {
		return fmt.Errorf("no .app bundles found in DMG")
	}

	return nil
}

// copyDir recursively copies a directory, handling symlinks properly
func copyDir(src, dst string) error {
	// Get source directory info (use Lstat to not follow symlinks)
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// If source is a symlink, copy the symlink itself
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(linkTarget, dst)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Check if it's a symlink
		srcEntryInfo, err := os.Lstat(srcPath)
		if err != nil {
			return err
		}

		if srcEntryInfo.Mode()&os.ModeSymlink != 0 {
			// Copy the symlink itself
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				return err
			}
		} else if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}

			// Preserve executable permissions
			if info, err := os.Stat(srcPath); err == nil {
				os.Chmod(dstPath, info.Mode())
			}
		}
	}

	return nil
}

func setup7Zip(baseDir string) error {
	toolsDir := filepath.Join(baseDir, "Tools", "7zip")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return err
	}

	platform := runtime.GOOS
	
	switch platform {
	case "windows":
		sevenZipPath := filepath.Join(toolsDir, "7za.exe")
		printInfo("Downloading 7-Zip for Windows...")
		url := "https://www.7-zip.org/a/7zr.exe"
		if err := downloadFile(url, sevenZipPath); err != nil {
			return err
		}
		
	case "linux":
		printInfo("Downloading 7-Zip for Linux...")
		tarPath := filepath.Join(toolsDir, "7z-linux.tar.xz")
		url := "https://github.com/ip7z/7zip/releases/download/25.01/7z2501-linux-x64.tar.xz"
		if err := downloadFile(url, tarPath); err != nil {
			return err
		}
		// Extract tar.xz using system tar (more reliable for complex xz files)
		cmd := exec.Command("tar", "-xf", tarPath, "-C", toolsDir)
		if err := cmd.Run(); err != nil {
			// Fallback to Go implementation
			if err := extractTarXz(tarPath, toolsDir); err != nil {
				return fmt.Errorf("failed to extract 7-Zip: %v", err)
			}
		}
		// Make 7zz executable
		sevenZipPath := filepath.Join(toolsDir, "7zz")
		if err := os.Chmod(sevenZipPath, 0755); err != nil {
			return err
		}
		// Clean up tarball
		os.Remove(tarPath)
		
	case "darwin":
		printInfo("Downloading 7-Zip for macOS...")
		tarPath := filepath.Join(toolsDir, "7z-mac.tar.xz")
		url := "https://github.com/ip7z/7zip/releases/download/25.01/7z2501-mac.tar.xz"
		if err := downloadFile(url, tarPath); err != nil {
			return err
		}
		// Extract tar.xz using system tar (more reliable for complex xz files)
		cmd := exec.Command("tar", "-xf", tarPath, "-C", toolsDir)
		if err := cmd.Run(); err != nil {
			// Fallback to Go implementation
			if err := extractTarXz(tarPath, toolsDir); err != nil {
				return fmt.Errorf("failed to extract 7-Zip: %v", err)
			}
		}
		// Make 7zz executable
		sevenZipPath := filepath.Join(toolsDir, "7zz")
		if err := os.Chmod(sevenZipPath, 0755); err != nil {
			return err
		}
		// Clean up tarball
		os.Remove(tarPath)
		
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	printSuccess("✓ 7-Zip installed")
	return nil
}

func get7ZipPath(baseDir string) string {
	toolsDir := filepath.Join(baseDir, "Tools", "7zip")
	platform := runtime.GOOS
	
	switch platform {
	case "windows":
		return filepath.Join(toolsDir, "7za.exe")
	case "linux", "darwin":
		return filepath.Join(toolsDir, "7zz")
	default:
		return ""
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func waitForExit(code int) {
	if runtime.GOOS == "windows" {
		fmt.Println()
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
	os.Exit(code)
}

// configurePCSX2 sets up PCSX2 to use the provided BIOS directory in portable mode
func configurePCSX2(emuDir, biosDir, platform string) error {
	pcsx2Dir := filepath.Join(emuDir, "PCSX2")
	
	// On Linux, PCSX2 is an AppImage, which supports portable mode via environment variable
	// or by placing portable.txt next to the extracted AppImage files
	// Since we're running the AppImage directly, we need a different approach
	// The AppImage will look for portable.txt in its extraction directory
	// But for simplicity, we'll just ensure the bios folder exists with BIOS files
	// PCSX2 AppImage on Linux will prompt for BIOS location on first run
	
	// Create portable.txt to make PCSX2 use local config (Qt version uses portable.txt)
	portableFile := filepath.Join(pcsx2Dir, "portable.txt")
	if err := os.WriteFile(portableFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create portable.txt: %v", err)
	}
	
	// Create the inis directory for config files
	inisDir := filepath.Join(pcsx2Dir, "inis")
	if err := os.MkdirAll(inisDir, 0755); err != nil {
		return fmt.Errorf("failed to create inis directory: %v", err)
	}
	
	// Create necessary directories for portable mode
	dirsToCreate := []string{"bios", "snaps", "sstates", "memcards", "logs", "cheats", "patches", "cache", "textures", "inputprofiles", "covers", "gamesettings"}
	for _, dir := range dirsToCreate {
		os.MkdirAll(filepath.Join(pcsx2Dir, dir), 0755)
	}
	
	// PCSX2 Qt version uses relative paths in portable mode
	// The bios folder is relative to the PCSX2 directory
	pcsx2Config := `[UI]
SettingsVersion = 1
InhibitScreensaver = true
StartFullscreen = false
SetupWizardIncomplete = false

[Folders]
Bios = bios
Snapshots = snaps
Savestates = sstates
MemoryCards = memcards
Logs = logs
Cheats = cheats
Patches = patches
Cache = cache
Textures = textures
InputProfiles = inputprofiles
Covers = covers

[EmuCore]
EnablePatches = true
EnableFastBoot = true
EnableGameFixes = true

[BIOS]
SearchDirectory = bios
`
	
	configPath := filepath.Join(inisDir, "PCSX2.ini")
	if err := os.WriteFile(configPath, []byte(pcsx2Config), 0644); err != nil {
		return fmt.Errorf("failed to write PCSX2.ini: %v", err)
	}
	
	return nil
}

// copyBIOSFilesToSystemFolder copies BIOS files from system-specific subfolders
// to the main system folder for RetroArch core compatibility
func copyBIOSFilesToSystemFolder(systemDir string) {
	// Map of subfolder names to the files that should be copied to the main system folder
	// Many RetroArch cores expect BIOS files directly in the system folder
	biosMapping := map[string][]string{
		// PlayStation - SwanStation, Beetle PSX, PCSX ReARMed
		"Sony - PlayStation": {
			"scph5500.bin", "scph5501.bin", "scph5502.bin", // Essential PS1 BIOS (JP, US, EU)
			"scph1001.bin", "scph7001.bin", "scph101.bin",  // Alternative BIOS versions
		},
		// Sega CD / Mega CD - Genesis Plus GX, PicoDrive
		"Sega - Mega CD - Sega CD": {
			"bios_CD_U.bin", "bios_CD_E.bin", "bios_CD_J.bin", // Sega CD BIOS (US, EU, JP)
		},
		// Sega Saturn - Beetle Saturn, Yabause
		"Sega - Saturn": {
			"sega_101.bin", "mpr-17933.bin", // Saturn BIOS (US/EU, JP)
			"saturn_bios.bin",               // Generic name some cores use
		},
		// Sega Dreamcast - Flycast
		"Sega - Dreamcast": {
			"dc_boot.bin", "dc_flash.bin", // Dreamcast BIOS and flash
		},
		// Also check the "dc" folder (common alternate location)
		"dc": {
			"dc_boot.bin", "dc_flash.bin",
		},
		// NEC PC Engine CD / TurboGrafx-CD - Beetle PCE
		"NEC - PC Engine - TurboGrafx 16 - SuperGrafx": {
			"syscard3.pce", "syscard2.pce", "syscard1.pce", // System Card BIOS
			"gexpress.pce", // Game Express CD Card
		},
		// NEC PC-FX
		"NEC - PC-FX": {
			"pcfx.rom", "pcfxbios.bin",
		},
		// Atari Lynx - Handy, Beetle Lynx
		"Atari - Lynx": {
			"lynxboot.img",
		},
		// Atari 5200
		"Atari - 5200": {
			"5200.rom", "ATARIXL.ROM",
		},
		// Atari 7800
		"Atari - 7800": {
			"7800 BIOS (U).rom", "7800 BIOS (E).rom",
		},
		// ColecoVision - blueMSX, Gearcoleco
		"Coleco - ColecoVision": {
			"colecovision.rom", "coleco.rom",
		},
		// Intellivision - FreeIntv
		"Mattel - Intellivision": {
			"exec.bin", "grom.bin",
		},
		// MSX - blueMSX, fMSX
		"Microsoft - MSX": {
			"MSX.ROM", "MSX2.ROM", "MSX2EXT.ROM", "MSX2P.ROM", "MSX2PEXT.ROM",
			"DISK.ROM", "FMPAC.ROM", "MSXDOS2.ROM", "KANJI.ROM",
		},
		// Nintendo Famicom Disk System
		"Nintendo - Famicom Disk System": {
			"disksys.rom",
		},
		// Nintendo Game Boy Advance - mGBA, VBA-M
		"Nintendo - Game Boy Advance": {
			"gba_bios.bin",
		},
		// Nintendo Game Boy / Color
		"Nintendo - Gameboy": {
			"gb_bios.bin", "dmg_boot.bin",
		},
		"Nintendo - Gameboy Color": {
			"gbc_bios.bin", "cgb_boot.bin",
		},
		// Nintendo DS - DeSmuME, melonDS
		"Nintendo - Nintendo DS": {
			"bios7.bin", "bios9.bin", "firmware.bin",
		},
		// Nintendo Super Game Boy
		"Nintendo - Super Game Boy": {
			"sgb_bios.bin", "SGB1.sfc", "SGB2.sfc",
		},
		// SNK Neo Geo CD
		"SNK - NeoGeo CD": {
			"neocd_f.rom", "neocd_sf.rom", "neocd_t.rom", "neocd_st.rom",
			"neocd_z.rom", "neocd.bin", "uni-bioscd.rom",
		},
		// 3DO
		"3DO Company, The - 3DO": {
			"panafz1.bin", "panafz10.bin", "panafz10-norsa.bin",
			"goldstar.bin", "sanyotry.bin", "3do_arcade_saot.bin",
		},
		// Magnavox Odyssey2 / Philips Videopac
		"Magnavox - Odyssey2": {
			"o2rom.bin",
		},
		"Phillips - Videopac+": {
			"c52.bin", "g7400.bin",
		},
		// Sharp X68000
		"Sharp - X68000": {
			"iplrom.dat", "cgrom.dat",
		},
		// Commodore Amiga - PUAE
		"Commodore - Amiga": {
			"kick34005.A500", "kick40063.A600", "kick40068.A1200",
			"kick33180.A500", "kick37175.A500",
		},
		// PSP - PPSSPP (needs specific folder structure but also check here)
		"Sony - PlayStation Portable": {
			"ppge_atlas.zim",
		},
	}

	totalCopied := 0

	for subfolderName, biosFiles := range biosMapping {
		subfolderPath := filepath.Join(systemDir, subfolderName)
		if _, err := os.Stat(subfolderPath); os.IsNotExist(err) {
			continue // Subfolder doesn't exist
		}

		for _, biosFile := range biosFiles {
			srcPath := filepath.Join(subfolderPath, biosFile)
			dstPath := filepath.Join(systemDir, biosFile)

			// Skip if source doesn't exist
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				continue
			}
			// Skip if destination already exists
			if _, err := os.Stat(dstPath); err == nil {
				continue
			}

			// Copy the file
			if err := copyFile(srcPath, dstPath); err == nil {
				totalCopied++
			}
		}
	}

	if totalCopied > 0 {
		printInfo(fmt.Sprintf("Copied %d BIOS files to system folder for core compatibility", totalCopied))
	}
}

// configureRetroArch sets up RetroArch to use the provided system/BIOS directory
func configureRetroArch(emuDir, systemDir string, platform string) error {
	var retroarchDir string
	switch platform {
	case "windows":
		retroarchDir = filepath.Join(emuDir, "RetroArch", "RetroArch-Win64")
	case "linux":
		retroarchDir = filepath.Join(emuDir, "RetroArch", "RetroArch-Linux-x86_64")
	case "darwin":
		retroarchDir = filepath.Join(emuDir, "RetroArch")
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
	
	configPath := filepath.Join(retroarchDir, "retroarch.cfg")
	
	// Convert paths to use forward slashes (RetroArch prefers this even on Windows)
	systemPath := filepath.ToSlash(systemDir)
	
	// Check if config already exists
	existingConfig := ""
	if data, err := os.ReadFile(configPath); err == nil {
		existingConfig = string(data)
	}
	
	// Key settings to ensure BIOS is found
	settings := map[string]string{
		"system_directory":        `"` + systemPath + `"`,
		"systemfiles_in_content_dir": `"false"`,
	}
	
	// Update or add settings
	lines := strings.Split(existingConfig, "\n")
	settingsFound := make(map[string]bool)
	
	for i, line := range lines {
		for key := range settings {
			if strings.HasPrefix(strings.TrimSpace(line), key+" ") || strings.HasPrefix(strings.TrimSpace(line), key+"=") {
				lines[i] = key + " = " + settings[key]
				settingsFound[key] = true
			}
		}
	}
	
	// Append any settings that weren't found
	for key, value := range settings {
		if !settingsFound[key] {
			lines = append(lines, key+" = "+value)
		}
	}
	
	// Write the updated config
	if err := os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write retroarch.cfg: %v", err)
	}
	
	return nil
}
