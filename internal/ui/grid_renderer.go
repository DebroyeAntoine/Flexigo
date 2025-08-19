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

	borderWidth := float32(2)
	border := canvas.NewRectangle(color.White)
	border.StrokeColor = color.RGBA{R: 120, G: 120, B: 120, A: 255}
	border.StrokeWidth = borderWidth
	paddedBtn := container.NewPadded(btn)

	borderedContainer := container.NewStack(border, paddedBtn)
	borderedContainer.Resize(fyne.NewSize(width, height))

	return borderedContainer, btn
}

func (ui *UIManager) renderBlocks(containerAction types.Action) (fyne.CanvasObject, [][]*ColorButton) {
	items := []GridItem{}
	objects := []fyne.CanvasObject{}
	rows := make([][]*ColorButton, containerAction.GridHeight)

	for _, block := range containerAction.Children {
		if block.Width == 0 {
			block.Width = 1
		}
		if block.Height == 0 {
			block.Height = 1
		}

		btn := NewColorButton(block.Label, func(b types.Action) func() {
			return func() { ui.ExecuteAction(b) }
		}(block), color.RGBA{R: 255, G: 0, B: 0, A: 255})

		borderedContainer, _ := ui.createBorderedButton(
			block.Label,
			func(b types.Action) func() {
				return func() { ui.ExecuteAction(b) }
			}(block),
			0, 0, // tailles calculées par le layout, donc 0 ici
			color.RGBA{R: 255, G: 0, B: 0, A: 255},
		)

		item := GridItem{
			Object:   borderedContainer,
			Width:    block.Width,
			Height:   block.Height,
			Position: block.Position,
		}

		items = append(items, item)
		objects = append(objects, borderedContainer)
		ui.buttonToAction[btn] = block

		y := block.Position.Y
		if y >= 0 && y < len(rows) {
			rows[y] = append(rows[y], btn)
		}
	}

	gridContainer := NewContainerFromConfig(containerAction.GridWidth, containerAction.GridHeight, items, objects)

	return gridContainer, rows
}
