// ui/grid_renderer.go
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	//"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"

	"github.com/DebroyeAntoine/flexigo/internal/types"
)

func (ui *UIManager) createBorderedButton(
	label string,
	onTapped func(),
	width, height float32,
	bgColor color.Color,
) (fyne.CanvasObject, *ColorButton) {
	btn := NewColorButton(label, onTapped, bgColor)

	// Créer la bordure avec un rectangle
	borderWidth := float32(2)
	border := canvas.NewRectangle(color.White)
	border.StrokeColor = color.RGBA{R: 120, G: 120, B: 120, A: 255}
	border.StrokeWidth = borderWidth

	// Utiliser un container Border pour un positionnement automatique
	//	padding := float32(4) // padding interne

	// Créer un container avec padding
	paddedBtn := container.NewPadded(btn)

	// Stack le rectangle de bordure avec le bouton paddé
	borderedContainer := container.NewStack(border, paddedBtn)
	borderedContainer.Resize(fyne.NewSize(width, height))

	return borderedContainer, btn
}

// Version mise à jour de renderBlocks
func (ui *UIManager) renderBlocks(blocks []types.Action) (fyne.CanvasObject, [][]*ColorButton) {
	const buttonsPerRow = 3
	grid := container.NewGridWithColumns(buttonsPerRow)
	rows := [][]*ColorButton{}
	currentRow := []*ColorButton{}

	// Obtenir la taille de la fenêtre
	windowSize := ui.window.Canvas().Size()
	buttonSpacing := float32(40) // Espacement entre boutons

	buttonWidth := (windowSize.Width - buttonSpacing*(buttonsPerRow+1)) / buttonsPerRow

	buttonHeight := windowSize.Height / float32(len(blocks)/buttonsPerRow+1)

	for i, block := range blocks {
		// Utiliser la méthode de bordure choisie
		borderedContainer, btn := ui.createBorderedButton(
			block.Label,
			func(b types.Action) func() {
				return func() { ui.ExecuteAction(b) }
			}(block),
			buttonWidth-30,
			buttonHeight-30,
			color.RGBA{R: 255, G: 0, B: 0, A: 255},
		)

		ui.buttonToAction[btn] = block

		grid.Add(borderedContainer)
		currentRow = append(currentRow, btn)

		if (i+1)%buttonsPerRow == 0 {
			rows = append(rows, currentRow)
			currentRow = []*ColorButton{}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return grid, rows
}
