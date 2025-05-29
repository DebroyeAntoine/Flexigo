// ui/grid_renderer.go
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/DebroyeAntoine/flexigo/internal/types"
)

// Create a button for each block
func (ui *UIManager) renderBlocks(blocks []types.Action) (fyne.CanvasObject, [][]*widget.Button) {
	const buttonsPerRow = 3
	grid := container.NewGridWithColumns(buttonsPerRow)
	rows := [][]*widget.Button{}

	currentRow := []*widget.Button{}
	for i, block := range blocks {
		btn := widget.NewButton(block.Label, func() {
			ui.ExecuteAction(block)
		})
		ui.buttonToAction[btn] = block

		btn.Resize(fyne.NewSize(100, 40))
		btn.Importance = widget.MediumImportance
		grid.Add(btn)
		currentRow = append(currentRow, btn)

		if (i+1)%buttonsPerRow == 0 {
			rows = append(rows, currentRow)
			currentRow = []*widget.Button{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return grid, rows
}
