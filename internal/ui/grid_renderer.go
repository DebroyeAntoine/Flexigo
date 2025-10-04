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

func (ui *UIManager) renderBlocks(containerAction types.Action) (fyne.CanvasObject, [][]*ColorButton, [][][]*ColorButton) {
	items := []GridItem{}
	objects := []fyne.CanvasObject{}
	rows := make([][]*ColorButton, containerAction.GridHeight)

	// Map pour organiser par groupes : groupID -> rowIndex -> buttons
	groupMap := make(map[int]map[int][]*ColorButton)

	for _, block := range containerAction.Children {
		if block.Width == 0 {
			block.Width = 1
		}
		if block.Height == 0 {
			block.Height = 1
		}

		borderedContainer, btn := ui.createBorderedButton(
			block.Label,
			func(b types.Action) func() {
				return func() { ui.ExecuteAction(b) }
			}(block),
			0, 0,
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

		// Organiser par groupe
		groupID := 0
		if block.GroupMembership != nil {
			groupID = *block.GroupMembership
		}

		// Initialiser la map du groupe si nécessaire
		if groupMap[groupID] == nil {
			groupMap[groupID] = make(map[int][]*ColorButton)
		}

		// Ajouter le bouton à la ligne du groupe
		if y >= 0 {
			groupMap[groupID][y] = append(groupMap[groupID][y], btn)
		}
	}

	// Convertir groupMap en [][][]*ColorButton
	// Format: groups[groupIndex][rowIndex][buttonIndex]
	groups := make([][][]*ColorButton, 0)

	// Trouver le nombre de groupes
	maxGroupID := 0
	for gid := range groupMap {
		if gid > maxGroupID {
			maxGroupID = gid
		}
	}

	// Construire les groupes dans l'ordre
	for groupID := 0; groupID <= maxGroupID; groupID++ {
		if groupRows, exists := groupMap[groupID]; exists {
			// Trouver le nombre de lignes dans ce groupe
			maxRow := 0
			for rowIdx := range groupRows {
				if rowIdx > maxRow {
					maxRow = rowIdx
				}
			}

			// Créer le slice de lignes pour ce groupe
			groupRowsSlice := make([][]*ColorButton, 0)
			for rowIdx := 0; rowIdx <= maxRow; rowIdx++ {
				if buttons, exists := groupRows[rowIdx]; exists && len(buttons) > 0 {
					groupRowsSlice = append(groupRowsSlice, buttons)
				}
			}

			// Ajouter ce groupe seulement s'il contient des lignes
			if len(groupRowsSlice) > 0 {
				groups = append(groups, groupRowsSlice)
			}
		}
	}

	gridContainer := NewContainerFromConfig(containerAction.GridWidth, containerAction.GridHeight, items, objects)

	return gridContainer, rows, groups
}
