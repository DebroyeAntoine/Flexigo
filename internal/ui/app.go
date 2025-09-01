package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/DebroyeAntoine/flexigo/internal/orchestration"
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type GridState int

const (
	StateIdle GridState = iota
	StateRows
	StateItems
)

type UIManager struct {
	state            GridState
	window           fyne.Window
	contentContainer *fyne.Container
	navigationStack  []types.Action
	currentContainer types.Action
	rows             [][]*ColorButton
	selectedRow      []*ColorButton
	selectedItem     *ColorButton
	rowScanDone      chan bool
	itemScanDone     chan bool
	timer            int
	buttonToAction   map[*ColorButton]types.Action
	blocks           []types.Action
	keyboardLayout   []string
	textBuffer       string
	textInput        *widget.Entry
	orchestration    *orchestration.Orchestration
}

func NewUIManager(window fyne.Window) *UIManager {
	return &UIManager{
		state:            StateIdle,
		window:           window,
		contentContainer: container.NewStack(container.NewVBox()),
	}
}

func (ui *UIManager) HandleEnterKey() {
	switch ui.state {
	case StateIdle:
		ui.state = StateRows
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
	// layers := []fyne.CanvasObject{ui.contentContainer}
	ui.window.SetContent(container.NewBorder(nil, nil, nil, nil, ui.contentContainer))

	// ui.window.SetContent(container.NewStack(layers...))
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
			ui.textBuffer = ui.textBuffer[:len(ui.textBuffer)-1]
			ui.textInput.SetText(ui.textBuffer)
		}
	case "speak":
		fmt.Println("Lecture du texte:", ui.textBuffer)
		ui.orchestration.Say(ui.textBuffer)
	default:
		ui.ExecuteAction(action)
	}
}

func (ui *UIManager) updateView(containerAction types.Action) {
	var backBtn *ColorButton
	ui.currentContainer = containerAction

	// Add a back button if the stack is non empty
	if len(ui.navigationStack) > 0 {
		backBtn = NewColorButton("Back", func() {
			// Show the last stack of blocs
			last := ui.navigationStack[len(ui.navigationStack)-1]
			ui.navigationStack = ui.navigationStack[:len(ui.navigationStack)-1]
			ui.updateView(last)
			ui.setState(StateIdle)
		}, color.White)
		backBtn.Resize(fyne.NewSize(300, 300))
		//	content = append(content, backBtn)
	}

	// Render blocks (common logic)
	var firstValue fyne.CanvasObject
	ui.blocks = containerAction.Children
	firstValue, ui.rows = ui.renderBlocks(containerAction)

	// Add back button to rows if it exists
	if backBtn != nil {
		ui.rows = append([][]*ColorButton{{backBtn}}, ui.rows...)
		ui.buttonToAction[backBtn] = types.Action{Label: "Back", Type: "back"}
	}

	if backBtn != nil {
		ui.rows = append([][]*ColorButton{{backBtn}}, ui.rows...)
		ui.buttonToAction[backBtn] = types.Action{Label: "Back", Type: "back"}
	}

	var finalContent fyne.CanvasObject
	if backBtn != nil {
		finalContent = container.NewVBox(
			container.NewHBox(backBtn),
			firstValue,
		)
	} else {
		finalContent = firstValue
	}
	ui.contentContainer.Objects = []fyne.CanvasObject{finalContent}
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
	}
	if block.Type == "keyboard" {
		ui.OpenVirtualKeyboard()
		return
	}
	if block.Type == "char" {
		ui.ExecuteKeyboardAction(block)
		ui.textBuffer = block.Label
		fmt.Println(ui.textBuffer)
	} else {
		ui.setState(StateIdle)
		fmt.Println("Action lancée :", block.Label)
		return
	}
}

func (ui *UIManager) StartRowsScan(onRowSelected func(int)) {
	ticker := time.NewTicker(time.Duration(ui.timer) * time.Millisecond)

	currentRow := 0

	go func() {
		for {
			select {
			case <-ui.rowScanDone:
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
				itemToHighlight := currentCol // Capturer la valeur actuelle
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
	// btn.Importance = widget.MediumImportance
	btn.Refresh()
}

func highlightItem(items []*ColorButton, index int) {
	for i, item := range items {
		if i == index {
			item.BGColor = color.RGBA{B: 255, A: 255}
			// item.Importance = widget.HighImportance // par exemple
		} else {
			item.BGColor = item.OriginalColor
			//@item.Importance = widget.MediumImportance
		}
		item.Refresh()
	}
}

func (ui *UIManager) ShowCustomActionGrid(rows [][]types.Action) {
	buttonRows := [][]*ColorButton{}

	// Crée l'entrée de texte
	ui.textInput = widget.NewEntry()
	ui.textInput.SetText(ui.textBuffer)
	ui.textInput.Wrapping = fyne.TextWrapWord
	ui.textInput.MultiLine = true
	ui.textInput.Disable()
	ui.textInput.SetMinRowsVisible(10) // Augmenter la hauteur

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
				}(actionRow[i]), color.White)
				// btn.Importance = widget.MediumImportance
				ui.buttonToAction[btn] = *action
			} else {
				btn = NewColorButton("", nil, color.Transparent)
				// btn.Disable()
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

	// Scrollable layout pour tout rendre accessible (y compris Retour)
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

	// Affiche ce clavier
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
	localTTS, err := tts.NewTTSProvider("local")
	if err != nil {
		return err
	}
	orchestration := orchestration.Orchestration{TTS: localTTS, Cfg: cfg}
	myWindow.SetFullScreen(true)
	myUI := NewUIManager(myWindow)
	myUI.orchestration = &orchestration
	myUI.buttonToAction = make(map[*ColorButton]types.Action, 10)
	myWindow.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyReturn {
			myUI.HandleEnterKey()
		}
	})

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
		myUI.contentContainer, // normal grid
	))

	myWindow.ShowAndRun()
	return nil
}
