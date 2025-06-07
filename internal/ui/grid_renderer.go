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

	// Obtenir la taille de la fenêtre
	windowSize := ui.window.Canvas().Size()
	buttonWidth := windowSize.Width / buttonsPerRow
	buttonHeight := windowSize.Height / float32(len(blocks)/buttonsPerRow+1) // +1 pour gérer les rangées incomplètes

	for i, block := range blocks {
		btn := widget.NewButton(block.Label, func() {
			ui.ExecuteAction(block)
		})
		ui.buttonToAction[btn] = block

		// Définir une taille calculée
		btn.Resize(fyne.NewSize(buttonWidth-10, buttonHeight-10)) // -10 pour les marges
		//@btn.Resize(fyne.NewSize(300, 450))
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
