package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type ROM struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size string `json:"size"`
	Date string `json:"date"`
}

type EmulatorConfig struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
	Name string   `json:"name"`
}

type SystemConfig struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Dir                string           `json:"dir"`
	RomJsonFile        string           `json:"romJsonFile"`
	Emulator           EmulatorConfig   `json:"emulator"`
	StandaloneEmulator *EmulatorConfig  `json:"standaloneEmulator"`
	FileExtensions     []string         `json:"fileExtensions"`
	NeedsExtract       bool             `json:"needsExtract"`
}

type SystemsConfig struct {
	Systems []SystemConfig `json:"systems"`
}

var systems map[string]SystemConfig
var systemsListCache []SystemListItem
var favorites map[string]map[string]bool

var baseDir string
var romsDir string
var romgetPath string
var favoritesPath string

func init() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	
	// Get the directory containing the executable
	exeDir := filepath.Dir(exe)
	
	// Check if we're running from the root (EmuBuddyLauncher.exe is in root)
	// or from a subdirectory during development
	if fileExists(filepath.Join(exeDir, "1g1rsets")) {
		// Executable is in the root directory
		baseDir = exeDir
	} else if fileExists(filepath.Join(filepath.Dir(exeDir), "1g1rsets")) {
		// One level up (e.g., launcher/EmuBuddyLauncher.exe)
		baseDir = filepath.Dir(exeDir)
	} else if fileExists(filepath.Join(filepath.Dir(filepath.Dir(exeDir)), "1g1rsets")) {
		// Two levels up (e.g., launcher/gui/EmuBuddyLauncher.exe)
		baseDir = filepath.Dir(filepath.Dir(exeDir))
	} else {
		// Fallback: assume exe is in root
		baseDir = exeDir
	}
	
	romsDir = filepath.Join(baseDir, "roms")
	romgetPath = filepath.Join(baseDir, "Tools", "romget", "romget")
	if runtime.GOOS == "windows" {
		romgetPath += ".exe"
	}
	favoritesPath = filepath.Join(baseDir, "favorites.json")

	// Load systems configuration and favorites
	loadSystemsConfig()
	loadFavorites()
}

func loadSystemsConfig() {
	configPath := filepath.Join(baseDir, "systems.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load systems.json: %v", err))
	}

	var config SystemsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		panic(fmt.Sprintf("Failed to parse systems.json: %v", err))
	}

	// Build the systems map and list
	systems = make(map[string]SystemConfig)
	systemsListCache = make([]SystemListItem, 0, len(config.Systems))
	for _, sys := range config.Systems {
		systems[sys.ID] = sys
		systemsListCache = append(systemsListCache, SystemListItem{
			Key:  sys.ID,
			Name: sys.Name,
		})
	}
}

func loadFavorites() {
	favorites = make(map[string]map[string]bool)
	data, err := os.ReadFile(favoritesPath)
	if err != nil {
		// File doesn't exist yet, that's ok
		return
	}

	if err := json.Unmarshal(data, &favorites); err != nil {
		fmt.Println("Failed to parse favorites.json:", err)
	}
}

func saveFavorites() {
	data, err := json.Marshal(favorites)
	if err != nil {
		fmt.Println("Failed to save favorites:", err)
		return
	}

	if err := os.WriteFile(favoritesPath, data, 0644); err != nil {
		fmt.Println("Failed to write favorites.json:", err)
	}
}

func (a *App) isFavorite(gameName string) bool {
	if favorites[a.currentSystem] == nil {
		return false
	}
	return favorites[a.currentSystem][gameName]
}

func (a *App) toggleFavorite(gameName string) {
	if favorites[a.currentSystem] == nil {
		favorites[a.currentSystem] = make(map[string]bool)
	}
	
	if favorites[a.currentSystem][gameName] {
		delete(favorites[a.currentSystem], gameName)
	} else {
		favorites[a.currentSystem][gameName] = true
	}
	
	saveFavorites()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type App struct {
	window          fyne.Window
	currentSystem   string
	allGames        []ROM
	filteredGames   []ROM
	gamesList       *widget.List
	searchEntry     *widget.Entry
	statusLabel     *widget.Label
	showFavoritesOnly bool
	favoritesToggle *widget.Check
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("EmuBuddy Launcher")
	myWindow.Resize(fyne.NewSize(1200, 700))

	appState := &App{
		window:    myWindow,
		allGames:  []ROM{},
		statusLabel: widget.NewLabel("Select a system to get started"),
	}

	// System list
	systemsList := widget.NewList(
		func() int {
			return len(getSystemList())
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("System Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			systems := getSystemList()
			obj.(*widget.Label).SetText(systems[id].Name)
		},
	)

	systemsList.OnSelected = func(id widget.ListItemID) {
		systems := getSystemList()
		appState.selectSystem(systems[id].Key)
	}

	// Search bar
	appState.searchEntry = widget.NewEntry()
	appState.searchEntry.SetPlaceHolder("Search games...")
	appState.searchEntry.OnChanged = func(query string) {
		appState.filterGames(query)
	}

	// Favorites toggle
	appState.favoritesToggle = widget.NewCheck("Show Favorites Only", func(checked bool) {
		appState.showFavoritesOnly = checked
		appState.filterGames(appState.searchEntry.Text)
	})

	// Games list
	appState.gamesList = widget.NewList(
		func() int {
			return len(appState.filteredGames)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("Game Name")
			nameLabel.Wrapping = fyne.TextTruncate
			sizeLabel := widget.NewLabel("Size")
			statusBadge := widget.NewLabel("Status")
			favBtn := widget.NewButton("+", nil)

			downloadBtn := widget.NewButton("Download", nil)
			playBtn := widget.NewButton("Play", nil)
			buttonsBox := container.NewHBox(downloadBtn, playBtn)

			return container.NewBorder(
				nil, nil,
				container.NewVBox(statusBadge, sizeLabel, favBtn),
				buttonsBox,
				nameLabel,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(appState.filteredGames) {
				return
			}

			game := appState.filteredGames[id]
			border := obj.(*fyne.Container)

			nameLabel := border.Objects[0].(*widget.Label)
			leftBox := border.Objects[1].(*fyne.Container)
			buttonsBox := border.Objects[2].(*fyne.Container)

			statusBadge := leftBox.Objects[0].(*widget.Label)
			sizeLabel := leftBox.Objects[1].(*widget.Label)
			favBtn := leftBox.Objects[2].(*widget.Button)

			// Set content
			nameLabel.SetText(game.Name)
			sizeLabel.SetText(game.Size)

			// Update favorite button
			isFav := appState.isFavorite(game.Name)
			if isFav {
				favBtn.SetText("-")
			} else {
				favBtn.SetText("+")
			}
			favBtn.OnTapped = func() {
				appState.toggleFavorite(game.Name)
				appState.gamesList.Refresh()
			}

			// Check ROM status
			exists := appState.checkROMExists(game.Name)
			if exists {
				statusBadge.SetText("Downloaded")
				// Only show play button when downloaded
				playBtn := widget.NewButton("Play", func() {
					appState.launchGame(game)
				})
				buttonsBox.Objects = []fyne.CanvasObject{playBtn}
			} else {
				statusBadge.SetText("Not Downloaded")
				// Only show download button when not downloaded
				downloadBtn := widget.NewButton("Download", func() {
					appState.downloadGame(game)
				})
				buttonsBox.Objects = []fyne.CanvasObject{downloadBtn}
			}
			buttonsBox.Refresh()
		},
	)

	// Layout
	leftPanel := container.NewBorder(
		widget.NewLabelWithStyle("Systems", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		systemsList,
	)

	rightPanel := container.NewBorder(
		container.NewVBox(
			appState.statusLabel,
			appState.searchEntry,
			appState.favoritesToggle,
		),
		nil, nil, nil,
		appState.gamesList,
	)

	content := container.NewHSplit(
		leftPanel,
		rightPanel,
	)
	content.SetOffset(0.25)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

type SystemListItem struct {
	Key  string
	Name string
}

func getSystemList() []SystemListItem {
	return systemsListCache
}

func (a *App) selectSystem(systemKey string) {
	a.currentSystem = systemKey
	config := systems[systemKey]
	a.statusLabel.SetText(fmt.Sprintf("Loading %s...", config.Name))

	// Load the ROM JSON file specified in config
	jsonFile := filepath.Join(baseDir, "1g1rsets", config.RomJsonFile)
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	if err := json.Unmarshal(data, &a.allGames); err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	a.filteredGames = a.allGames
	a.gamesList.Refresh()
	a.statusLabel.SetText(fmt.Sprintf("%s - %d games", systems[systemKey].Name, len(a.allGames)))
	a.searchEntry.SetText("")
}

func (a *App) filterGames(query string) {
	filtered := []ROM{}
	
	for _, game := range a.allGames {
		// Apply search filter
		if query != "" {
			queryLower := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(game.Name), queryLower) {
				continue
			}
		}
		
		// Apply favorites filter
		if a.showFavoritesOnly && !a.isFavorite(game.Name) {
			continue
		}
		
		filtered = append(filtered, game)
	}
	
	a.filteredGames = filtered
	a.gamesList.Refresh()
	a.statusLabel.SetText(fmt.Sprintf("%s - %d games", systems[a.currentSystem].Name, len(a.filteredGames)))
}

func (a *App) checkROMExists(romName string) bool {
	config := systems[a.currentSystem]
	romDir := filepath.Join(romsDir, config.Dir)

	// If system needs extraction, check for extracted ROM files (not ZIP)
	if config.NeedsExtract {
		// Get base name without .zip extension
		baseName := strings.TrimSuffix(romName, ".zip")

		// Check for any matching ROM file with valid extensions
		for _, ext := range config.FileExtensions {
			romPath := filepath.Join(romDir, baseName+ext)
			if _, err := os.Stat(romPath); err == nil {
				return true
			}
		}

		// Also check for any file starting with baseName (handles multiple files in ZIP)
		entries, err := os.ReadDir(romDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					name := entry.Name()
					// Check if file starts with game name and has valid extension
					if strings.HasPrefix(name, baseName) {
						for _, ext := range config.FileExtensions {
							if strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext)) {
								return true
							}
						}
					}
				}
			}
		}
		return false
	}

	// For systems that support ZIP directly
	romPath := filepath.Join(romDir, romName)
	_, err := os.Stat(romPath)
	return err == nil
}

func (a *App) downloadGame(game ROM) {
	config := systems[a.currentSystem]
	romDir := filepath.Join(romsDir, config.Dir)
	os.MkdirAll(romDir, 0755)

	outputPath := filepath.Join(romDir, game.Name)

	// Create progress bar UI
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressLabel := widget.NewLabel("Starting download...")
	progressContent := container.NewVBox(
		widget.NewLabel(game.Name),
		progressBar,
		progressLabel,
	)

	progressDialog := dialog.NewCustom("Downloading", "Cancel", progressContent, a.window)
	cancelled := false
	progressDialog.SetOnClosed(func() {
		cancelled = true
	})
	progressDialog.Show()

	go func() {
		// Download with progress using HTTP directly
		err := downloadWithProgress(game.URL, outputPath, func(downloaded, total int64) {
			if cancelled {
				return
			}
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				progressBar.SetValue(percent)
				progressLabel.SetText(fmt.Sprintf("%s / %s (%.1f%%)",
					formatBytes(downloaded), formatBytes(total), percent))
			} else {
				progressLabel.SetText(fmt.Sprintf("Downloaded: %s", formatBytes(downloaded)))
			}
		})

		if cancelled {
			os.Remove(outputPath)
			os.Remove(outputPath + ".tmp")
			return
		}

		if err != nil {
			progressDialog.Hide()
			dialog.ShowError(fmt.Errorf("Download failed: %v", err), a.window)
			return
		}

		// Extract ZIP if needed
		if config.NeedsExtract && strings.HasSuffix(strings.ToLower(game.Name), ".zip") {
			progressLabel.SetText("Extracting...")
			progressBar.SetValue(0)

			extractedFile, err := extractROMZip(outputPath, romDir)
			if err != nil {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("Extraction failed: %v", err), a.window)
				return
			}

			// Remove the ZIP after extraction
			os.Remove(outputPath)

			progressDialog.Hide()
			dialog.ShowInformation("Done", fmt.Sprintf("Installed: %s", filepath.Base(extractedFile)), a.window)
		} else {
			progressDialog.Hide()
			dialog.ShowInformation("Done", fmt.Sprintf("Downloaded: %s", game.Name), a.window)
		}

		a.gamesList.Refresh()
	}()
}

func downloadWithProgress(url, outputPath string, progress func(downloaded, total int64)) error {
	// Create HTTP client
	client := &http.Client{
		Timeout: 0, // No timeout for large downloads
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Set headers for myrient compatibility
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")

	// Infer referer from URL
	if strings.Contains(url, "myrient.erista.me") {
		dir := filepath.Dir(strings.ReplaceAll(url, "https://myrient.erista.me", ""))
		req.Header.Set("Referer", "https://myrient.erista.me"+dir+"/")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	total := resp.ContentLength

	// Create temp file
	tempPath := outputPath + ".tmp"
	out, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	// Download with progress updates
	var downloaded int64
	buf := make([]byte, 64*1024)
	lastUpdate := time.Now()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				out.Close()
				os.Remove(tempPath)
				return writeErr
			}
			downloaded += int64(n)

			// Update progress every 100ms
			if time.Since(lastUpdate) > 100*time.Millisecond {
				progress(downloaded, total)
				lastUpdate = time.Now()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			out.Close()
			os.Remove(tempPath)
			return readErr
		}
	}

	out.Close()
	progress(downloaded, total)

	// Move to final location
	return os.Rename(tempPath, outputPath)
}

func extractROMZip(zipPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var extractedFile string

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Skip macOS metadata files
		if strings.HasPrefix(f.Name, "__MACOSX") || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}

		// Get just the filename, ignore any directory structure in ZIP
		name := filepath.Base(f.Name)
		fpath := filepath.Join(destDir, name)

		// Create file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}

		// Keep track of the first extracted file (usually the ROM)
		if extractedFile == "" {
			extractedFile = fpath
		}
	}

	return extractedFile, nil
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

func (a *App) launchGame(game ROM) {
	config := systems[a.currentSystem]

	// Check if standalone emulator option is available
	hasStandalone := config.StandaloneEmulator != nil && config.StandaloneEmulator.Path != ""

	if hasStandalone {
		// Show choice dialog
		a.showEmulatorChoice(game, config)
		return
	}

	// Launch with default emulator
	a.launchWithEmulator(game, config.Emulator.Path, config.Emulator.Args)
}

func (a *App) showEmulatorChoice(game ROM, config SystemConfig) {
	// Get emulator names
	defaultName := config.Emulator.Name
	if defaultName == "" {
		defaultName = "RetroArch"
	}

	standaloneName := ""
	if config.StandaloneEmulator != nil {
		standaloneName = config.StandaloneEmulator.Name
	}
	if standaloneName == "" {
		standaloneName = "Standalone"
	}

	// Create choice dialog
	message := widget.NewLabel(fmt.Sprintf("Choose emulator for %s:", game.Name))
	
	var dlg dialog.Dialog
	defaultBtn := widget.NewButton(defaultName, func() {
		a.launchWithEmulator(game, config.Emulator.Path, config.Emulator.Args)
		dlg.Hide()
	})
	standaloneBtn := widget.NewButton(standaloneName, func() {
		a.launchWithEmulator(game, config.StandaloneEmulator.Path, config.StandaloneEmulator.Args)
		dlg.Hide()
	})

	content := container.NewVBox(
		message,
		defaultBtn,
		standaloneBtn,
	)

	dlg = dialog.NewCustom("Choose Emulator", "Cancel", content, a.window)
	dlg.Show()
}

func (a *App) launchWithEmulator(game ROM, emuPath string, emuArgs []string) {
	config := systems[a.currentSystem]
	romDir := filepath.Join(romsDir, config.Dir)
	emuPath = filepath.Join(baseDir, emuPath)

	// Find the ROM file
	var romPath string

	if config.NeedsExtract {
		// Look for extracted ROM file
		baseName := strings.TrimSuffix(game.Name, ".zip")

		// First, try exact match with known extensions
		for _, ext := range config.FileExtensions {
			testPath := filepath.Join(romDir, baseName+ext)
			if _, err := os.Stat(testPath); err == nil {
				romPath = testPath
				break
			}
		}

		// If not found, scan directory for matching files
		if romPath == "" {
			entries, err := os.ReadDir(romDir)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						name := entry.Name()
						// Check if file starts with game name and has valid extension
						if strings.HasPrefix(name, baseName) {
							for _, ext := range config.FileExtensions {
								if strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext)) {
									romPath = filepath.Join(romDir, name)
									break
								}
							}
						}
						if romPath != "" {
							break
						}
					}
				}
			}
		}
	} else {
		// For systems that support ZIP directly
		romPath = filepath.Join(romDir, game.Name)
	}

	if romPath == "" {
		dialog.ShowError(fmt.Errorf("ROM not found. Try downloading it again."), a.window)
		return
	}

	if _, err := os.Stat(romPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("ROM not found: %s", romPath), a.window)
		return
	}

	if _, err := os.Stat(emuPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("Emulator not found: %s", emuPath), a.window)
		return
	}

	// Build arguments - resolve relative paths for RetroArch cores
	var args []string
	emuDir := filepath.Dir(emuPath)
	for _, arg := range emuArgs {
		if strings.HasSuffix(arg, ".dll") || strings.HasSuffix(arg, ".so") {
			// This is a core path - make it absolute
			corePath := filepath.Join(emuDir, arg)
			if _, err := os.Stat(corePath); os.IsNotExist(err) {
				dialog.ShowError(fmt.Errorf("Emulator core not found: %s\n\nRetroArch cores need to be installed separately.", corePath), a.window)
				return
			}
			args = append(args, corePath)
		} else {
			args = append(args, arg)
		}
	}
	args = append(args, romPath)

	cmd := exec.Command(emuPath, args...)
	cmd.Dir = emuDir

	// Capture stderr for debugging
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to start: %v", err), a.window)
		return
	}

	// Check for early exit errors in background
	go func() {
		errOutput, _ := io.ReadAll(stderr)
		err := cmd.Wait()
		if err != nil {
			// Process exited with error
			errMsg := string(errOutput)
			if errMsg == "" {
				errMsg = err.Error()
			}
			dialog.ShowError(fmt.Errorf("Emulator error: %s", errMsg), a.window)
		}
	}()

	dialog.ShowInformation("Launched", filepath.Base(romPath), a.window)
}
