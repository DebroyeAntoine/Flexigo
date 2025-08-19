package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

// GridItem represents an item with its property in the grid
type GridItem struct {
	Object     fyne.CanvasObject `yaml:"-"` // Object to be placed
	Width      int               `yaml:"width,omitempty"`
	Height     int               `yaml:"height,omitempty"`
	Position   types.Position    `yaml:"position,omitempty"`
	GridWidth  int               `yaml:"grid_width,omitempty"`
	GridHeight int               `yaml:"grid_height,omitempty"`
}

// CustomGridLayout impléments fyne.layout
type CustomGridLayout struct {
	gridWidth  int
	gridHeight int
	items      []GridItem
}

func NewCustomGridLayout(gridWidth, gridHeight int, items []GridItem) *CustomGridLayout {
	return &CustomGridLayout{
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
		items:      items,
	}
}

// Layout calculate and move objects. I t will be automatically called by fyne
func (c *CustomGridLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if len(objects) == 0 || c.gridWidth == 0 || c.gridHeight == 0 {
		return
	}

	// Calcul of the unit size of an item in the grid
	cellWidth := containerSize.Width / float32(c.gridWidth)
	cellHeight := containerSize.Height / float32(c.gridHeight)

	for i, obj := range objects {
		if i >= len(c.items) {
			// Hide items in case of more items than expected containerSize
			// TODO: remove this by checking at loading the conf
			obj.Hide()
			continue
		}

		item := c.items[i]

		x := float32(item.Position.X) * cellWidth
		y := float32(item.Position.Y) * cellHeight

		width := float32(item.Width) * cellWidth
		height := float32(item.Height) * cellHeight

		obj.Move(fyne.NewPos(x, y))
		obj.Resize(fyne.NewSize(width, height))
		obj.Show()
	}
}

func (c *CustomGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}

	var minWidth, minHeight float32 = 100, 100

	for _, item := range c.items {
		requiredWidth := float32(item.Position.X+item.Width) * 50
		requiredHeight := float32(item.Position.Y+item.Height) * 50

		if requiredWidth > minWidth {
			minWidth = requiredWidth
		}
		if requiredHeight > minHeight {
			minHeight = requiredHeight
		}
	}

	return fyne.NewSize(minWidth, minHeight)
}

func (c *CustomGridLayout) UpdateGridSize(width, height int) {
	c.gridWidth = width
	c.gridHeight = height
}

func (c *CustomGridLayout) AddItem(item GridItem) {
	c.items = append(c.items, item)
}

func (c *CustomGridLayout) UpdateItem(index int, item GridItem) {
	if index >= 0 && index < len(c.items) {
		c.items[index] = item
	}
}

func (c *CustomGridLayout) GetItems() []GridItem {
	return c.items
}

func NewContainerFromConfig(gridWidth, gridHeight int, items []GridItem, objects []fyne.CanvasObject) *fyne.Container {
	for i := range items {
		if i < len(objects) {
			items[i].Object = objects[i]
		}
	}

	layout := NewCustomGridLayout(gridWidth, gridHeight, items)
	return container.New(layout, objects...)
}
