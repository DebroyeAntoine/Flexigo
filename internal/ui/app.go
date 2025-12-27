package ui

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DebroyeAntoine/flexigo/internal/http"
	"github.com/DebroyeAntoine/flexigo/internal/ir"
	"github.com/DebroyeAntoine/flexigo/internal/orchestration"
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

// GridState defines the current navigation/focus level in the scanning interface
type GridState int

const (
	StateIdle        GridState = iota // No active scan
	StateGroup                        // Scanning blocks of items
	StateRows                         // Scanning rows within a group
	StateItems                        // Scanning individual items within a row
	StateBrowserMode                  // Interface is hidden, browser is active
)

// UIManager manages the UI state, navigation stack, and the scanning selection logic
type UIManager struct {
	state            GridState
	window           fyne.Window
	app              fyne.App
	contentContainer *fyne.Container
	navigationStack  []types.Action
	currentContainer types.Action
	rows             [][]*ColorButton
	groups           [][][]*ColorButton
	selectedRow      []*ColorButton
	selectedGroup    [][]*ColorButton
	selectedItem     *ColorButton
	rowScanDone      chan bool
	groupScanDone    chan bool
	itemScanDone     chan bool
	timer            int // Scanning interval in milliseconds
	buttonToAction   map[*ColorButton]types.Action
	blocks           []types.Action
	keyboardLayout   []string
	textBuffer       string
	textInput        *widget.Entry
	orchestration    *orchestration.Orchestration
	voice            string

	// Browser mode management
	browserExecutor      *browser.BrowserExecutor
	browserControlWindow fyne.Window
	browserActive        bool
	currentCmd           *exec.Cmd
}

// customTheme defines visual overrides for the Fyne application
type customTheme struct {
	fyne.Theme
}

// Size returns custom sizes for specific text types to improve accessibility
func (t *customTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 78
	case theme.SizeNameHeadingText:
		return 64
	case theme.SizeNameSubHeadingText:
		return 56
	case theme.SizeNameCaptionText:
		return 32
	default:
		return t.Theme.Size(name)
	}
}

// Color returns custom colors for the application theme
func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground:
		return color.White
	case theme.ColorNameForeground:
		return color.Black
	default:
		return t.Theme.Color(name, variant)
	}
}

// NewUIManager initializes a new UI manager with default values
func NewUIManager(window fyne.Window, app fyne.App) *UIManager {
	return &UIManager{
		state:            StateIdle,
		window:           window,
		app:              app,
		contentContainer: container.NewStack(container.NewVBox()),
		browserActive:    false,
	}
}

// getBrowserPath returns the relative path to the browser binary based on the OS
func getBrowserPath() string {
	basePath := "./bin/browser/"
	switch runtime.GOOS {
	case "windows":
		return basePath + "win-unpacked/flexigo-browser.exe"
	case "darwin":
		return basePath + "mac-arm64/flexigo-browser.app/Contents/MacOS/flexigo-browser"
	case "linux":
		return basePath + "linux-unpacked/flexigo-browser"
	default:
		return ""
	}
}

// EnterBrowserMode hides the main UI and launches the external browser process
func (ui *UIManager) EnterBrowserMode(url string) {
	path := getBrowserPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Error: browser binary not found at %s", path)
		return
	}

	ui.window.Hide()
	ui.state = StateBrowserMode

	ui.currentCmd = exec.Command(path, url)
	err := ui.currentCmd.Start()
	if err != nil {
		log.Println("Error launching browser:", err)
		ui.window.Show()
		ui.state = StateIdle
		return
	}

	// Wait for browser exit in a goroutine to restore the UI
	go func() {
		ui.currentCmd.Wait()
		fyne.Do(func() {
			ui.state = StateIdle
			ui.currentCmd = nil
			ui.window.Show()
		})
	}()
}

// HandleEnterKey processes the selection event (Enter key or switch click)
// based on the current scanning state.
func (ui *UIManager) HandleEnterKey() {
	if ui.state == StateBrowserMode {
		if ui.currentCmd != nil && ui.currentCmd.Process != nil {
			fmt.Println("Stopping Browser")
			ui.currentCmd.Process.Kill()
		}
		return
	}

	switch ui.state {
	case StateIdle:
		ui.state = StateGroup
		ui.groupScanDone = make(chan bool)
		ui.StartGroupScan()
	case StateGroup:
		ui.groupScanDone <- true
		ui.state = StateRows
		ui.rows = ui.selectedGroup
		ui.rowScanDone = make(chan bool)
		ui.StartRowsScan(func(t int) { fmt.Println(t) })
	case StateRows:
		ui.rowScanDone <- true
		ui.state = StateItems

		// Instant selection if the row contains only one item
		if len(ui.selectedRow) == 1 {
			ui.state = StateIdle
			action := ui.buttonToAction[ui.selectedRow[0]]
			unhighlightlastItem(ui.selectedRow[0])
			ui.ExecuteAction(action)
			break
		}
		ui.itemScanDone = make(chan bool)
		ui.StartItemScan()
	case StateItems:
		ui.itemScanDone <- true
		ui.state = StateIdle
		action := ui.buttonToAction[ui.selectedItem]
		unhighlightlastItem(ui.selectedItem)
		ui.ExecuteAction(action)
	}
}

func (ui *UIManager) setState(state GridState) {
	ui.state = state
	ui.refreshUI()
}

func (ui *UIManager) refreshUI() {
	ui.window.SetContent(container.NewBorder(nil, nil, nil, nil, ui.contentContainer))
}

// OpenVirtualKeyboard navigates to the virtual keyboard view
func (ui *UIManager) OpenVirtualKeyboard() {
	ui.navigationStack = append(ui.navigationStack, ui.currentContainer)
	ui.ShowVirtualKeyboardFromLayout()
	ui.setState(StateIdle)
}

// ExecuteKeyboardAction handles specific logic for virtual keyboard keys
func (ui *UIManager) ExecuteKeyboardAction(action types.Action) {
	switch action.Type {
	case "char":
		ui.textBuffer += action.Label
		ui.textInput.SetText(ui.textBuffer)
	case "space":
		ui.textBuffer += " "
		ui.textInput.SetText(ui.textBuffer)
	case "delete":
		if len(ui.textBuffer) > 0 {
			runes := []rune(ui.textBuffer)
			ui.textBuffer = string(runes[:len(runes)-1])
			ui.textInput.SetText(ui.textBuffer)
		}
	case "speak":
		if err := ui.orchestration.SayWithVoice(ui.textBuffer, ui.voice); err != nil {
			log.Printf("TTS error: %v", err)
		}
	default:
		ui.ExecuteAction(action)
	}
}

// updateView re-renders the grid based on a container action and updates the navigation
func (ui *UIManager) updateView(containerAction types.Action) {
	ui.currentContainer = containerAction
	ui.blocks = containerAction.Children
	var mainContent fyne.CanvasObject

	// Add a back button if we are not at the root level
	if len(ui.navigationStack) > 0 {
		backAction := types.Action{
			Label:    "← Retour",
			Type:     "back",
			Width:    containerAction.GridWidth,
			Height:   1,
			Position: types.Position{X: 0, Y: 0},
		}

		// Shift existing blocks down to accommodate the back button
		adjustedBlocks := []types.Action{backAction}
		for _, block := range containerAction.Children {
			adjustedBlock := block
			adjustedBlock.Position.Y += 1
			adjustedBlocks = append(adjustedBlocks, adjustedBlock)
		}

		adjustedContainer := containerAction
		adjustedContainer.Children = adjustedBlocks
		adjustedContainer.GridHeight += 1

		firstValue, rows, groups := ui.renderBlocks(adjustedContainer)
		ui.rows = rows
		ui.groups = groups
		mainContent = firstValue
	} else {
		firstValue, rows, groups := ui.renderBlocks(containerAction)
		ui.rows = rows
		ui.groups = groups
		mainContent = firstValue
	}

	bgColor := containerAction.Background
	background := canvas.NewRectangle(convertColor(bgColor))

	// Stack the background color behind the buttons
	ui.contentContainer.Objects = []fyne.CanvasObject{
		background,
		mainContent,
	}

	ui.contentContainer.Refresh()
}

// ExecuteAction triggers the logic associated with a selected block
func (ui *UIManager) ExecuteAction(block types.Action) {
	if block.Type == "back" {
		if len(ui.navigationStack) > 0 {
			last := ui.navigationStack[len(ui.navigationStack)-1]
			ui.navigationStack = ui.navigationStack[:len(ui.navigationStack)-1]
			ui.updateView(last)
		}
		ui.setState(StateIdle)
		return
	}

	if block.Type == "container" {
		ui.timer = block.Timer
		ui.navigationStack = append(ui.navigationStack, ui.currentContainer)
		ui.updateView(block)
		ui.setState(StateIdle)
		return
	}

	if block.Type == "browser" {
		ui.EnterBrowserMode(block.BrowserURL)
		return
	}

	if block.Type == "keyboard" {
		ui.OpenVirtualKeyboard()
		return
	}

	if block.Type == "char" || block.Type == "speak" {
		ui.ExecuteKeyboardAction(block)
		return
	}

	// External hardware/service integrations
	if block.Type == "tts" {
		if err := ui.orchestration.ExecuteTTSAction(block); err != nil {
			log.Printf("TTS action failed: %v", err)
		}
	}
	if block.Type == "http" {
		if err := ui.orchestration.ExecuteHTTPAction(block); err != nil {
			log.Printf("HTTP action failed: %v", err)
		}
	}
	if block.Type == "ir" {
		if err := ui.orchestration.ExecuteIRAction(block); err != nil {
			log.Printf("IR action failed: %v", err)
		}
	}

	ui.setState(StateIdle)
	fmt.Println("Action triggered:", block.Label)
}

// --- SCANNING LOGIC ---

// StartGroupScan begins the automated highlighting of groups
func (ui *UIManager) StartGroupScan() {
	ticker := time.NewTicker(time.Duration(ui.timer) * time.Millisecond)
	currentGroup := 0
	if len(ui.groups) == 1 {
		ui.state = StateRows
		ui.rows = ui.groups[0]
		ui.rowScanDone = make(chan bool)
		ui.StartRowsScan(func(t int) {})
		return
	}

	go func() {
		for {
			select {
			case <-ui.groupScanDone:
				ticker.Stop()
				return
			case <-ticker.C:
				if currentGroup >= len(ui.groups) {
					ticker.Stop()
					ui.selectedGroup = nil
					fyne.Do(func() {
						unhighlightlastGroup(ui.groups[len(ui.groups)-1])
					})
					ui.state = StateIdle
					ui.groupScanDone <- true
					return
				}
				groupToHighlight := currentGroup
				fyne.Do(func() {
					highlightGroup(ui.groups, groupToHighlight)
				})
				ui.selectedGroup = ui.groups[currentGroup]
				currentGroup++
			}
		}
	}()
}

func unhighlightlastGroup(group [][]*ColorButton) {
	for _, row := range group {
		for _, btn := range row {
			btn.BGColor = btn.OriginalColor
			btn.Refresh()
		}
	}
}

func highlightGroup(group [][][]*ColorButton, index int) {
	for i, rows := range group {
		for _, row := range rows {
			for _, btn := range row {
				if i == index {
					btn.BGColor = color.RGBA{B: 255, A: 255}
				} else {
					btn.BGColor = btn.OriginalColor
				}
				btn.Refresh()
			}
		}
	}
}

// StartRowsScan begins the automated highlighting of rows
func (ui *UIManager) StartRowsScan(onRowSelected func(int)) {
	ticker := time.NewTicker(time.Duration(ui.timer) * time.Millisecond)
	currentRow := 0

	go func() {
		for {
			select {
			case <-ui.rowScanDone:
				ticker.Stop()
				return
			case <-ticker.C:
				if currentRow >= len(ui.rows) {
					ticker.Stop()
					ui.selectedRow = nil
					fyne.Do(func() {
						unhighlightlastRow(ui.rows[len(ui.rows)-1])
					})
					ui.state = StateIdle
					ui.rowScanDone <- true
					return
				}
				rowToHighlight := currentRow
				fyne.Do(func() {
					highlightRow(ui.rows, rowToHighlight)
				})
				ui.selectedRow = ui.rows[currentRow]
				currentRow++
			}
		}
	}()
}

func unhighlightlastRow(row []*ColorButton) {
	for _, btn := range row {
		btn.BGColor = btn.OriginalColor
		btn.Refresh()
	}
}

func highlightRow(rows [][]*ColorButton, index int) {
	for i, row := range rows {
		for _, btn := range row {
			if i == index {
				btn.BGColor = color.RGBA{B: 255, A: 255}
			} else {
				btn.BGColor = btn.OriginalColor
			}
			btn.Refresh()
		}
	}
}

// StartItemScan begins the automated highlighting of individual buttons in a row
func (ui *UIManager) StartItemScan() {
	ticker := time.NewTicker(time.Duration(ui.timer) * time.Millisecond)
	currentCol := 0

	go func() {
		for {
			select {
			case <-ui.itemScanDone:
				ticker.Stop()
				return
			case <-ticker.C:
				if currentCol >= len(ui.selectedRow) {
					ticker.Stop()
					fyne.Do(func() {
						unhighlightlastItem(ui.selectedRow[len(ui.selectedRow)-1])
					})
					ui.selectedItem = nil
					ui.state = StateIdle
					ui.itemScanDone <- true
					return
				}
				itemToHighlight := currentCol
				fyne.Do(func() {
					highlightItem(ui.selectedRow, itemToHighlight)
				})
				ui.selectedItem = ui.selectedRow[currentCol]
				currentCol++
			}
		}
	}()
}

func unhighlightlastItem(btn *ColorButton) {
	btn.BGColor = btn.OriginalColor
	btn.Refresh()
}

func highlightItem(items []*ColorButton, index int) {
	for i, item := range items {
		if i == index {
			item.BGColor = color.RGBA{B: 255, A: 255}
		} else {
			item.BGColor = item.OriginalColor
		}
		item.Refresh()
	}
}

// ShowCustomActionGrid generates a grid of actions, used primarily for the virtual keyboard
func (ui *UIManager) ShowCustomActionGrid(rows [][]types.Action) {
	buttonRows := [][]*ColorButton{}

	ui.textInput = widget.NewMultiLineEntry()
	ui.textInput.SetText(ui.textBuffer)
	ui.textInput.OnChanged = func(t string) {
		ui.textBuffer = t
	}
	ui.textInput.Wrapping = fyne.TextWrapWord
	ui.textInput.SetMinRowsVisible(3)
	ui.textInput.TextStyle = fyne.TextStyle{}

	backBtn := NewColorButton("← Retour", func() {
		if len(ui.navigationStack) > 0 {
			last := ui.navigationStack[len(ui.navigationStack)-1]
			ui.navigationStack = ui.navigationStack[:len(ui.navigationStack)-1]
			ui.updateView(last)
			ui.setState(StateIdle)
		}
	}, color.RGBA{R: 255, B: 255, A: 255})

	topSection := container.NewVBox(
		backBtn,
		ui.textInput,
	)

	keyboardContainer := container.NewVBox()

	maxCols := 0
	for _, actionRow := range rows {
		if len(actionRow) > maxCols {
			maxCols = len(actionRow)
		}
	}

	for _, actionRow := range rows {
		btnRow := []*ColorButton{}
		rowContainer := container.NewGridWithColumns(maxCols)

		for i := 0; i < maxCols; i++ {
			var action *types.Action
			if i < len(actionRow) {
				action = &actionRow[i]
			}

			var btn *ColorButton
			if action != nil {
				btn = NewColorButton(action.Label, func(a types.Action) func() {
					return func() {
						ui.ExecuteKeyboardAction(a)
					}
				}(actionRow[i]), color.RGBA{R: 255, A: 255})
				ui.buttonToAction[btn] = *action
			} else {
				btn = NewColorButton("", nil, color.Transparent)
			}

			btnRow = append(btnRow, btn)
			rowContainer.Add(container.NewVBox(
				layout.NewSpacer(),
				btn,
				layout.NewSpacer(),
			))
		}

		buttonRows = append(buttonRows, btnRow)
		keyboardContainer.Add(rowContainer)
	}

	scrollable := container.NewVScroll(container.NewVBox(
		topSection,
		keyboardContainer,
	))

	ui.contentContainer.Objects = []fyne.CanvasObject{scrollable}
	ui.contentContainer.Refresh()

	// Make the back button scannable by adding it as the first row
	backRow := []*ColorButton{backBtn}
	buttonRows = append([][]*ColorButton{backRow}, buttonRows...)

	ui.buttonToAction[backBtn] = types.Action{Label: "Retour", Type: "back"}
	ui.groups = [][][]*ColorButton{buttonRows}
	// ui.rows = buttonRows
}

// ShowVirtualKeyboardFromLayout prepares the keyboard actions based on the configuration layout
func (ui *UIManager) ShowVirtualKeyboardFromLayout() {
	if len(ui.keyboardLayout) == 0 {
		return
	}

	rows := [][]types.Action{}
	for _, line := range ui.keyboardLayout {
		row := []types.Action{}
		for _, char := range line {
			row = append(row, types.Action{
				Label: string(char),
				Type:  "char",
			})
		}
		rows = append(rows, row)
	}

	// Add special function keys
	rows = append(rows, []types.Action{
		{Label: "Espace", Type: "space"},
		{Label: "Effacer", Type: "delete"},
		{Label: "Lire", Type: "speak"},
	})

	ui.ShowCustomActionGrid(rows)
}

// LoadKeyboard recursively searches for a keyboard action in the config blocks
func (ui *UIManager) LoadKeyboard(actions *[]types.Action) {
	for _, action := range *actions {
		if action.Type == "keyboard" {
			if len(action.Layout) != 0 {
				ui.keyboardLayout = action.Layout
				return
			}
		}
		if action.Type == "container" {
			ui.LoadKeyboard(&action.Children)
		}
	}
}

// StartUI initializes the Fyne application and main window components
func StartUI(cfg *types.Config) error {
	myApp := app.New()
	myWindow := myApp.NewWindow("Flexigo")

	// Apply accessibility theme
	myApp.Settings().SetTheme(&customTheme{
		Theme: myApp.Settings().Theme(),
	})

	localTTS, err := tts.NewTTSProvider("local")
	if err != nil {
		return err
	}
	httpClient := http.NewHTTPClient()

	var irSender ir.IRSender
	if cfg.IRBackend != "" {
		var irErr error
		irConfig := ir.IRConfig{
			SerialPort:   cfg.IRSerialPort,
			BaudRate:     cfg.IRBaudRate,
			CommandsFile: "ir_commands.yaml",
			Timeout:      5000,
		}

		irSender, irErr = ir.NewIRSender(cfg.IRBackend, irConfig)
		if irErr != nil {
			return fmt.Errorf("failed to initialize IR module on %s: %w", cfg.IRSerialPort, irErr)
		}
	}

	orchestration := orchestration.Orchestration{
		TTS:  localTTS,
		Cfg:  cfg,
		HTTP: httpClient,
		IR:   irSender,
	}

	myWindow.SetFullScreen(true)

	myUI := NewUIManager(myWindow, myApp)
	myUI.orchestration = &orchestration
	myUI.buttonToAction = make(map[*ColorButton]types.Action, 10)
	myUI.voice = cfg.Voice

	// Register keyboard input
	myWindow.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyReturn {
			myUI.HandleEnterKey()
		}
	})

	// Register external switch input (via IR/Serial)
	if s, ok := irSender.(*ir.SerialIRSender); ok {
		s.ListenForEvents(func(msg string) {
			if msg == "BTN:CLICK" {
				// Use fyne.Do to ensure UI updates happen on the main thread
				fyne.Do(func() {
					myUI.HandleEnterKey()
				})
			}
		})
	}

	if len(cfg.Blocks) == 0 {
		fmt.Println("No blocks found in config.")
		return nil
	}

	myUI.timer = cfg.Blocks[0].Timer
	myUI.LoadKeyboard(&cfg.Blocks)
	myUI.updateView(cfg.Blocks[0])
	myUI.refreshUI()

	myWindow.SetContent(container.NewStack(
		myUI.contentContainer,
	))

	myWindow.ShowAndRun()

	// Clean up resources on exit
	if myUI.browserExecutor != nil {
		myUI.browserExecutor.Close()
	}

	return nil
}
