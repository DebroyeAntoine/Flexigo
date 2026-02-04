// ui/click_catcher.go
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ClickCatcher is a transparent widget that captures taps and invokes OnClick.
type ClickCatcher struct {
	widget.BaseWidget
	OnClick func()
}

// Tapped triggers the click handler when defined.
func (c *ClickCatcher) Tapped(_ *fyne.PointEvent) {
	if c.OnClick != nil {
		c.OnClick()
	}
}

// CreateRenderer draws a transparent rectangle to receive pointer events.
func (c *ClickCatcher) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(rect)
}
