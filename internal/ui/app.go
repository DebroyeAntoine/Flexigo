package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/theme"
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
	navigationStack  [][]types.Action
	rows             [][]*widget.Button
	selectedRow      []*widget.Button
	selectedItem     *widget.Button
	rowScanDone      chan bool
	itemScanDone     chan bool
	timer            int
	buttonToAction   map[*widget.Button]types.Action
	blocks           []types.Action
	keyboardLayout   []string
	textBuffer       string
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
	//layers := []fyne.CanvasObject{ui.contentContainer}
	ui.window.SetContent(container.NewBorder(nil, nil, nil, nil, ui.contentContainer))

	//ui.window.SetContent(container.NewStack(layers...))
}

func (ui *UIManager) OpenVirtualKeyboard() {
	ui.navigationStack = append(ui.navigationStack, ui.blocks)
	ui.ShowVirtualKeyboardFromLayout()
	ui.setState(StateIdle)
	// letters := []types.Action{}
	//
	//	for _, c := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ " {
	//		label := string(c)
	//		letters = append(letters, types.Action{
	//			Label: label,
	//			Type:  "char",
	//			//Value: label,
	//		})
	//	}
	//
	//	letters = append(letters, types.Action{
	//		Label: "Effacer",
	//		Type:  "delete",
	//	}, types.Action{
	//
	//		Label: "Lire",
	//		Type:  "speak",
	//	})
	//
	// ui.navigationStack = append(ui.navigationStack, ui.blocks)
	// ui.updateView(letters)
	// ui.setState(StateIdle)
}

func (ui *UIManager) updateView(blocks []types.Action) {
	var backBtn *widget.Button

	// Add a back button if the stack is non empty
	if len(ui.navigationStack) > 0 {
		backBtn = widget.NewButton("Back", func() {
			// Show the last stack of blocs
			last := ui.navigationStack[len(ui.navigationStack)-1]
			ui.navigationStack = ui.navigationStack[:len(ui.navigationStack)-1]
			ui.updateView(last)
			ui.setState(StateIdle)
		})
		backBtn.Resize(fyne.NewSize(300, 300))
		backBtn.Importance = widget.MediumImportance
		//	content = append(content, backBtn)
	}

	// Render blocks (common logic)
	var firstValue fyne.CanvasObject
	ui.blocks = blocks
	firstValue, ui.rows = ui.renderBlocks(blocks)

	// Add back button to rows if it exists
	if backBtn != nil {
		ui.rows = append([][]*widget.Button{{backBtn}}, ui.rows...)
		ui.buttonToAction[backBtn] = types.Action{Label: "Back", Type: "back"}
	}

	var finalContent fyne.CanvasObject
	if backBtn != nil {
		finalContent = container.NewBorder(backBtn, nil, nil, nil, firstValue)
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
		ui.navigationStack = append(ui.navigationStack, ui.blocks)
		ui.updateView(block.Children)
		ui.setState(StateIdle)
	}
	if block.Type == "keyboard" {
		ui.OpenVirtualKeyboard()
		return
	}
	if block.Type == "char" {
		fmt.Println("bloub")
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

func unhighlightlastRow(row []*widget.Button) {
	for _, btn := range row {
		btn.Importance = widget.MediumImportance
		btn.Refresh()
	}
}

func highlightRow(rows [][]*widget.Button, index int) {
	for i, row := range rows {
		for _, btn := range row {
			if i == index {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
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

func unhighlightlastItem(btn *widget.Button) {
	btn.Importance = widget.MediumImportance
	btn.Refresh()
}

func highlightItem(items []*widget.Button, index int) {
	for i, item := range items {
		if i == index {
			item.Importance = widget.HighImportance // par exemple
		} else {
			item.Importance = widget.MediumImportance
		}
		item.Refresh()
	}
}

func (ui *UIManager) ShowCustomActionGrid(rows [][]types.Action) {
	buttonRows := [][]*widget.Button{}

	// Crée l'entrée de texte
	textInput := widget.NewEntry()
	textInput.SetText(ui.textBuffer)
	textInput.Wrapping = fyne.TextWrapWord
	textInput.MultiLine = true
	textInput.Disable()
	textInput.SetMinRowsVisible(10) // Augmenter la hauteur

	backBtn := widget.NewButton("← Retour", func() {
		if len(ui.navigationStack) > 0 {
			last := ui.navigationStack[len(ui.navigationStack)-1]
			ui.navigationStack = ui.navigationStack[:len(ui.navigationStack)-1]
			ui.updateView(last)
			ui.setState(StateIdle)
		}
	})

	topSection := container.NewVBox(
		backBtn,
		textInput,
	)

	keyboardContainer := container.NewVBox()

	maxCols := 0
	for _, actionRow := range rows {
		if len(actionRow) > maxCols {
			maxCols = len(actionRow)
		}
	}

	for _, actionRow := range rows {
		btnRow := []*widget.Button{}
		rowContainer := container.NewGridWithColumns(maxCols)

		for i := 0; i < maxCols; i++ {
			var action *types.Action
			if i < len(actionRow) {
				action = &actionRow[i]
			}

			var btn *widget.Button
			if action != nil {
				btn = widget.NewButton(action.Label, func(a types.Action) func() {
					return func() {
						ui.ExecuteKeyboardAction(a, textInput)
					}
				}(actionRow[i]))
				btn.Importance = widget.MediumImportance
				ui.buttonToAction[btn] = *action
			} else {
				btn = widget.NewButton("", nil)
				btn.Disable()
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
    backRow := []*widget.Button{backBtn}
    buttonRows = append([][]*widget.Button{backRow}, buttonRows...)

	ui.buttonToAction[backBtn] = types.Action{Label: "Retour", Type: "back"}
	ui.rows = buttonRows
}

//	buttonRows := [][]*widget.Button{}
//	grid := container.NewVBox() // One row per keyboard line
//
//	for _, actionRow := range rows {
//		btnRow := []*widget.Button{}
//		buttons := []fyne.CanvasObject{}
//
//		for _, action := range actionRow {
//			btn := widget.NewButton(action.Label, nil)
//			btn.Importance = widget.MediumImportance
//			ui.buttonToAction[btn] = action
//			btnRow = append(btnRow, btn)
//			buttons = append(buttons, btn)
//		}
//
//		buttonRows = append(buttonRows, btnRow)
//
//		row := container.NewGridWithColumns(len(buttons))
//		for _, btn := range buttons {
//			row.Add(btn)
//		}
//
//		grid.Add(row)
//	}
//
//	// Optionnel : ajouter zone de texte en haut
//	inputLabel := widget.NewLabel(ui.textBuffer)
//	inputLabel.Alignment = fyne.TextAlignCenter
//	inputLabel.TextStyle = fyne.TextStyle{Bold: true}
//	ui.contentContainer.Objects = []fyne.CanvasObject{inputLabel, grid}
//	ui.contentContainer.Refresh()
//
//	ui.rows = buttonRows
//	//ui.contentContainer.Objects = []fyne.CanvasObject{final}
//	ui.contentContainer.Refresh()
//}

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
    myApp.Settings().SetTheme(&customTheme{Theme: theme.DefaultTheme()})
	myWindow := myApp.NewWindow("Flexigo")
	myWindow.SetFullScreen(true)
	myUI := NewUIManager(myWindow)
	myUI.buttonToAction = make(map[*widget.Button]types.Action, 10)
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
	myUI.updateView(cfg.Blocks[0].Children)
	myUI.refreshUI()

	myWindow.SetContent(container.NewStack(
		myUI.contentContainer, // normal grid
	))
	fmt.Printf("%T\n", myUI.window.Canvas())

	myWindow.ShowAndRun()
	return nil
}
