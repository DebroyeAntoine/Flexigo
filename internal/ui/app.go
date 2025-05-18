package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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
	catcher          *ClickCatcher
	navigationStack  [][]types.Action
}

func NewUIManager(window fyne.Window) *UIManager {
	return &UIManager{
		state:            StateIdle,
		window:           window,
		contentContainer: container.NewVBox(),
		catcher:          &ClickCatcher{},
	}
}

func (ui *UIManager) setState(state GridState) {
	ui.state = state
	ui.refreshUI()
}

func (ui *UIManager) refreshUI() {
	layers := []fyne.CanvasObject{ui.contentContainer}
	if ui.state == StateIdle {
		layers = append(layers, ui.catcher)
	}
	ui.window.SetContent(container.NewStack(layers...))
}

func (ui *UIManager) updateView(blocks []types.Action) {
	content := []fyne.CanvasObject{}

	// Add a back button if the stack is non empty
	if len(ui.navigationStack) > 0 {
		backBtn := widget.NewButton("Back", func() {
			// Show the last stack of blocs
			last := ui.navigationStack[len(ui.navigationStack)-1]
			ui.navigationStack = ui.navigationStack[:len(ui.navigationStack)-1]
			ui.updateView(last)
			ui.setState(StateIdle)
		})
		backBtn.Resize(fyne.NewSize(100, 40))
		backBtn.Importance = widget.MediumImportance
		content = append(content, backBtn)
	}

	// Render of blocs
	firstValue, _ := renderBlocks(blocks, func(block types.Action) {
		if block.Type == "container" {
			ui.navigationStack = append(ui.navigationStack, blocks)
			ui.updateView(block.Children)
		} else {
			ui.setState(StateIdle)
			fmt.Println("Action lancée :", block.Label)
		}
	})
	content = append(content, firstValue)
	ui.contentContainer.Objects = content
	ui.contentContainer.Refresh()
}

// StartUI show the graphical interface with blocks defined in conf
func StartUI(cfg *types.Config) error {
	myApp := app.New()
	myWindow := myApp.NewWindow("Flexigo")
	myUI := NewUIManager(myWindow)

	myUI.catcher.OnClick = func() {
		if myUI.state == StateIdle {
			fmt.Println("Lancement du scan des lignes")
			myUI.setState(StateRows)
		}
	}

	// Start after the main bloc
	if len(cfg.Blocks) == 0 {
		fmt.Println("No bloc found.")
		return nil
	}
	myUI.updateView(cfg.Blocks[0].Children)
	myUI.refreshUI()

	myWindow.SetContent(container.NewStack(
		myUI.contentContainer, // normal grid
		myUI.catcher,          // global click catcher
	))

	myWindow.ShowAndRun()
	return nil
}
