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
func renderBlocks(blocks []types.Action, onClick func(types.Action)) fyne.CanvasObject {
	grid := container.NewAdaptiveGrid(len(blocks))
	for _, block := range blocks {
		btn := widget.NewButton(block.Label, func(b types.Action) func() {
			return func() {
				onClick(b)
			}
		}(block)) // closure propre
		grid.Add(btn)
	}
	return grid
}

// StartUI show the graphical interface with blocks defined in conf
func StartUI(cfg *types.Config) error {
	myApp := app.New()
	myWindow := myApp.NewWindow("Flexigo")

	contentContainer := container.NewVBox()

	// Mandatory to declare it to use it recursively
	var updateView func(blocks []types.Action)

	updateView = func(blocks []types.Action) {
		contentContainer.Objects = []fyne.CanvasObject{
			renderBlocks(blocks, func(block types.Action) {
				if block.Type == "container" {
					updateView(block.Children)
				} else {
					fmt.Println("Action lancée :", block.Label)
				}
			}),
		}
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
