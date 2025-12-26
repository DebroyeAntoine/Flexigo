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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DebroyeAntoine/flexigo/internal/browser"
	"github.com/DebroyeAntoine/flexigo/internal/http"
	"github.com/DebroyeAntoine/flexigo/internal/ir"
	"github.com/DebroyeAntoine/flexigo/internal/orchestration"
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type GridState int

const (
	StateIdle GridState = iota
	StateGroup
	StateRows
	StateItems
	StateBrowserMode
)

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
	timer            int
	buttonToAction   map[*ColorButton]types.Action
	blocks           []types.Action
	keyboardLayout   []string
	textBuffer       string
	textInput        *widget.Entry
	orchestration    *orchestration.Orchestration
	voice            string

	// Browser mode avec nouvelle fenêtre
	browserExecutor      *browser.BrowserExecutor
	browserControlWindow fyne.Window // Nouvelle petite fenêtre de contrôle
	browserActive        bool
}

// Thème personnalisé pour contrôler les couleurs
type customTheme struct {
	fyne.Theme
}

func (t *customTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 78 // Taille du texte normal (votre Entry)
	case theme.SizeNameHeadingText:
		return 64 // Taille des titres
	case theme.SizeNameSubHeadingText:
		return 56 // Taille des sous-titres
	case theme.SizeNameCaptionText:
		return 32 // Taille des petits textes
	default:
		return t.Theme.Size(name)
	}
}

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

func NewUIManager(window fyne.Window, app fyne.App) *UIManager {
	return &UIManager{
		state:            StateIdle,
		window:           window,
		app:              app,
		contentContainer: container.NewStack(container.NewVBox()),
		browserActive:    false,
	}
}

func getBrowserPath() string {
	// En développement, on peut pointer vers le dossier bin
	// En production, le binaire sera à côté de l'app
	basePath := "./bin/browser/"

	switch runtime.GOOS {
	case "windows":
		return basePath + "win-unpacked/flexigo-browser.exe"
	case "darwin":
		return basePath + "mac-arm64/flexigo-browser.app/Contents/MacOS/flexigo-browser"
		// Pour macOS, le binaire est à l'intérieur du .app
		return basePath + "mac/flexigo-browser.app/Contents/MacOS/flexigo-browser"
	case "linux":
		return basePath + "linux-unpacked/flexigo-browser"
	default:
		return ""
	}
}

func (ui *UIManager) EnterBrowserMode(url string) {
	path := getBrowserPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Erreur : binaire browser introuvable à %s", path)
		return
	}

	ui.window.Hide()

	// On lance et on attend
	cmd := exec.Command(path, url)
	err := cmd.Run()
	if err != nil {
		log.Println("Le browser s'est arrêté avec une erreur:", err)
	}

	ui.window.Show()
}

//func (ui *UIManager) EnterBrowserMode(browserPath, url string) error {
//	log.Println("Entering browser mode...")
//
//	// Lance le navigateur
//	ui.browserExecutor = browser.NewBrowserExecutor(browserPath)
//	if err := ui.browserExecutor.Launch(url); err != nil {
//		return fmt.Errorf("failed to launch browser: %w", err)
//	}
//
//	ui.browserActive = true
//	ui.state = StateBrowserMode
//
//	// Attend que le navigateur démarre
//	time.Sleep(500 * time.Millisecond)
//
//	// Cache la fenêtre principale
//	//	ui.window.Hide()
//
//	// Crée une NOUVELLE petite fenêtre pour le contrôle
//	ui.browserControlWindow = ui.app.NewWindow("Contrôle Navigateur")
//
//	// Applique le contenu stylisé
//	content := ui.createBrowserControlContent()
//	ui.browserControlWindow.SetContent(content)
//
//	// Calcule la taille minimale du contenu
//	minSize := content.MinSize()
//	ui.browserControlWindow.Resize(minSize)
//
//	// Positionne en haut à gauche de l'écran
//	ui.browserControlWindow.SetFixedSize(true) // Empêche le redimensionnement
//
//	// Gestion de la fermeture de la fenêtre de contrôle
//	ui.browserControlWindow.SetOnClosed(func() {
//		log.Println("Control window closed, exiting browser mode")
//		ui.ExitBrowserMode()
//	})
//
//	// Affiche la fenêtre de contrôle
//	ui.browserControlWindow.Show()
//
//	// Positionne en haut à gauche après l'affichage
//	// (nécessaire car la position n'est accessible qu'après Show())
//	canvas := ui.browserControlWindow.Canvas()
//	if canvas != nil {
//		// Petite astuce : on force la position via un refresh
//		ui.browserControlWindow.RequestFocus()
//	}
//
//	ui.PositionWindow("Contrôle Navigateur", 0, 0)
//	log.Println("Browser mode active - control window shown")
//	return nil
//}

// createBrowserControlContent crée le contenu stylisé de la fenêtre de contrôle
func (ui *UIManager) createBrowserControlContent() fyne.CanvasObject {
	// Utilise vos ColorButton au lieu des boutons standards
	quitBtn := NewColorButton("Exit", func() {
		log.Println("Quit button clicked")
		if err := ui.ExitBrowserMode(); err != nil {
			log.Printf("Error exiting browser mode: %v", err)
		}
	}, color.RGBA{R: 220, G: 53, B: 69, A: 255}) // Rouge pour le bouton de fermeture

	// Ajuste la taille du texte si nécessaire
	quitBtn.MinSize()

	// Disposition verticale avec espacement
	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(
			container.NewVBox(
				quitBtn,
			),
		),
		layout.NewSpacer(),
	)

	// Optionnel : ajouter un fond de couleur pour toute la fenêtre
	// bgRect := canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 240, A: 255})

	return container.NewStack(content)
}

// ExitBrowserMode ferme le navigateur et la fenêtre de contrôle
func (ui *UIManager) ExitBrowserMode() error {
	log.Println("Exiting browser mode...")

	if !ui.browserActive {
		return nil
	}

	// Ferme le navigateur
	if ui.browserExecutor != nil {
		if err := ui.browserExecutor.Close(); err != nil {
			log.Printf("Warning: error closing browser: %v", err)
		}
		ui.browserExecutor = nil
	}

	// Ferme la fenêtre de contrôle
	if ui.browserControlWindow != nil {
		ui.browserControlWindow.Close()
		ui.browserControlWindow = nil
	}

	ui.browserActive = false
	ui.state = StateIdle

	// Restaure la fenêtre principale si elle était cachée
	// ui.window.Show()

	log.Println("Browser mode exited")
	return nil
}

func (ui *UIManager) HandleEnterKey() {
	if ui.state == StateBrowserMode {
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

		// Special case for lines with only one element in
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

func (ui *UIManager) OpenVirtualKeyboard() {
	ui.navigationStack = append(ui.navigationStack, ui.currentContainer)
	ui.ShowVirtualKeyboardFromLayout()
	ui.setState(StateIdle)
}

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

func (ui *UIManager) updateView(containerAction types.Action) {
	ui.currentContainer = containerAction
	ui.blocks = containerAction.Children

	// if this is not the root container add a back button
	if len(ui.navigationStack) > 0 {
		backAction := types.Action{
			Label:    "← Retour",
			Type:     "back",
			Width:    containerAction.GridWidth,
			Height:   1,
			Position: types.Position{X: 0, Y: 0},
		}

		// Do a shift +1 on all other blocks below
		adjustedBlocks := []types.Action{backAction}
		for _, block := range containerAction.Children {
			adjustedBlock := block
			adjustedBlock.Position.Y += 1
			adjustedBlocks = append(adjustedBlocks, adjustedBlock)
		}

		// Change the container by adding one extra row
		adjustedContainer := containerAction
		adjustedContainer.Children = adjustedBlocks
		adjustedContainer.GridHeight += 1

		firstValue, rows, groups := ui.renderBlocks(adjustedContainer)
		ui.rows = rows
		ui.groups = groups
		ui.contentContainer.Objects = []fyne.CanvasObject{firstValue}
	} else {
		firstValue, rows, groups := ui.renderBlocks(containerAction)
		ui.rows = rows
		ui.groups = groups
		ui.contentContainer.Objects = []fyne.CanvasObject{firstValue}
	}

	ui.contentContainer.Refresh()
}

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
		browserPath := block.BrowserPath
		if browserPath == "" {
			browserPath = browser.GetDefaultBrowserPath()
		}

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
	fmt.Println("Action lancée :", block.Label)
}

// [... Le reste des fonctions de scan restent identiques ...]
func (ui *UIManager) StartGroupScan() {
	ticker := time.NewTicker(time.Duration(ui.timer) * time.Millisecond)
	currentGroup := 0

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

	// Ajouter le bouton retour comme première ligne pour qu'il soit scannable
	backRow := []*ColorButton{backBtn}
	buttonRows = append([][]*ColorButton{backRow}, buttonRows...)

	ui.buttonToAction[backBtn] = types.Action{Label: "Retour", Type: "back"}
	ui.rows = buttonRows
}

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

	// Ajoute les boutons spéciaux à la fin
	rows = append(rows, []types.Action{
		{Label: "Espace", Type: "space"},
		{Label: "Effacer", Type: "delete"},
		{Label: "Lire", Type: "speak"},
	})

	ui.ShowCustomActionGrid(rows)
}

func (ui *UIManager) LoadKeyboard(actions *[]types.Action) {
	for _, action := range *actions {
		if action.Type == "keyboard" {
			if len(action.Layout) != 0 {
				ui.keyboardLayout = action.Layout
				fmt.Println("coucou")
				return
			}
		}
		if action.Type == "container" {
			ui.LoadKeyboard(&action.Children)
		}
	}
}

// StartUI show the graphical interface with blocks defined in conf
func StartUI(cfg *types.Config) error {
	myApp := app.New()
	myWindow := myApp.NewWindow("Flexigo")

	// Applique le thème personnalisé
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
			// Ici on s'arrête et on retourne l'erreur avant de lancer l'UI
			fmt.Println("test")
			return fmt.Errorf("impossible d'initialiser le module IR sur %s : %w", cfg.IRSerialPort, irErr)
		}

	}

	orchestration := orchestration.Orchestration{TTS: localTTS, Cfg: cfg, HTTP: httpClient, IR: irSender}
	myWindow.SetFullScreen(true)

	// IMPORTANT: Passer l'app en plus du window
	myUI := NewUIManager(myWindow, myApp)
	myUI.orchestration = &orchestration
	myUI.buttonToAction = make(map[*ColorButton]types.Action, 10)
	myUI.voice = cfg.Voice
	myWindow.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyReturn {
			myUI.HandleEnterKey()
		}
	})

	if s, ok := irSender.(*ir.SerialIRSender); ok {
		s.ListenForEvents(func(msg string) {
			if msg == "BTN:CLICK" {
				// Utiliser fyne.Do pour s'assurer que l'action s'exécute sur le thread UI
				fyne.Do(func() {
					myUI.HandleEnterKey()
				})
			}
		})
	}

	// Start after the main bloc
	if len(cfg.Blocks) == 0 {
		fmt.Println("No bloc found.")
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

	// Cleanup
	if myUI.browserExecutor != nil {
		myUI.browserExecutor.Close()
	}

	return nil
}
