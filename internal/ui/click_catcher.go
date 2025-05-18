// ui/click_catcher.go
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type ClickCatcher struct {
	widget.BaseWidget
	OnClick func()
}

func (c *ClickCatcher) Tapped(_ *fyne.PointEvent) {
	if c.OnClick != nil {
		c.OnClick()
	}
}

func (c *ClickCatcher) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(rect)
}
