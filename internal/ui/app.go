package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

// Create a button for each block
func renderBlocks(blocks []types.Action, onClick func(types.Action)) (fyne.CanvasObject, [][]*widget.Button) {
	const buttonsPerRow = 3
	grid := container.NewGridWithColumns(buttonsPerRow)
	rows := make([][]*widget.Button, 4)

	currentRow := []*widget.Button{}
	for i, block := range blocks {
		btn := widget.NewButton(block.Label, func(b types.Action) func() {
			return func() {
				onClick(b)
			}
		}(block))

		btn.Resize(fyne.NewSize(100, 40))
		btn.Importance = widget.MediumImportance

		grid.Add(btn)

		currentRow = append(currentRow, btn)

		// Exemple : 3 boutons par ligne
		if (i+1)%3 == 0 {
			rows = append(rows, currentRow)
			currentRow = []*widget.Button{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return grid, rows
}

// StartUI show the graphical interface with blocks defined in conf
func StartUI(cfg *types.Config) error {
	myApp := app.New()
	myWindow := myApp.NewWindow("Flexigo")

	contentContainer := container.NewVBox()

	var navigationStack [][]types.Action

	// Mandatory to declare it to use it recursively
	var updateView func(blocks []types.Action)

	updateView = func(blocks []types.Action) {
		content := []fyne.CanvasObject{}

		// Add a back button if the stack is non empty
		if len(navigationStack) > 0 {
			backBtn := widget.NewButton("Back", func() {
				// Show the last stack of blocs
				last := navigationStack[len(navigationStack)-1]
				navigationStack = navigationStack[:len(navigationStack)-1]
				updateView(last)
			})
			backBtn.Resize(fyne.NewSize(100, 40))
			backBtn.Importance = widget.MediumImportance
			content = append(content, backBtn)
		}

		// Render of blocs
		firstValue, _ := renderBlocks(blocks, func(block types.Action) {
			if block.Type == "container" {
				navigationStack = append(navigationStack, blocks)
				updateView(block.Children)
			} else {
				fmt.Println("Action lancée :", block.Label)
			}
		})
		content = append(content, firstValue)
		contentContainer.Objects = content
		contentContainer.Refresh()
	}

	// Start after the main bloc
	if len(cfg.Blocks) == 0 {
		fmt.Println("No bloc found.")
		return nil
	}
	updateView(cfg.Blocks[0].Children)

	myWindow.SetContent(contentContainer)
	myWindow.ShowAndRun()
	return nil
}
