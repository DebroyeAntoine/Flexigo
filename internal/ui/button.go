package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ColorButton defines a custom widget that displays a button with a colored background
// and supports visual highlighting (e.g., for scanning interfaces).
type ColorButton struct {
	widget.BaseWidget
	Text           string
	OnTapped       func()
	BGColor        color.Color
	OriginalColor  color.Color
	HighlightColor color.Color // Color used when the button is highlighted
	IsHighlighted  bool        // Current highlight state
}

// NewColorButton creates a new ColorButton with the given label, tap handler, and background color.
func NewColorButton(label string, onTapped func(), bgColor color.Color) *ColorButton {
	b := &ColorButton{
		Text:           label,
		OnTapped:       onTapped,
		BGColor:        bgColor,
		OriginalColor:  bgColor,
		HighlightColor: color.RGBA{R: 0, G: 0, B: 255, A: 255}, // Default blue highlight
		IsHighlighted:  false,
	}
	b.ExtendBaseWidget(b)
	return b
}

// Highlight triggers the visual highlight state by changing the background color.
func (b *ColorButton) Highlight() {
	b.IsHighlighted = true
	b.BGColor = b.HighlightColor
	b.Refresh()
}

// Unhighlight restores the button's background to its original color.
func (b *ColorButton) Unhighlight() {
	b.IsHighlighted = false
	b.BGColor = b.OriginalColor
	b.Refresh()
}

// SetHighlightColor updates the color used when the button is in a highlighted state.
func (b *ColorButton) SetHighlightColor(c color.Color) {
	b.HighlightColor = c
}

// CreateRenderer implements the fyne.Widget interface to define how the button is drawn.
func (b *ColorButton) CreateRenderer() fyne.WidgetRenderer {
	// Background rectangle
	bg := canvas.NewRectangle(b.BGColor)

	// Centered text label
	txt := canvas.NewText(b.Text, color.White)
	txt.TextSize = 30
	txt.Alignment = fyne.TextAlignCenter

	objs := []fyne.CanvasObject{bg, txt}
	return &colorButtonRenderer{
		button:     b,
		background: bg,
		label:      txt,
		objects:    objs,
	}
}

// Tapped handles the primary tap/click event.
func (b *ColorButton) Tapped(_ *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// TappedSecondary handles right-click events (not implemented).
func (b *ColorButton) TappedSecondary(_ *fyne.PointEvent) {}

// colorButtonRenderer manages the layout and synchronization of the ColorButton's components.
type colorButtonRenderer struct {
	button     *ColorButton
	background *canvas.Rectangle
	label      *canvas.Text
	objects    []fyne.CanvasObject
}

// Layout positions the background and centers the text within the widget's size.
func (r *colorButtonRenderer) Layout(size fyne.Size) {
	// Background fills the entire widget area
	r.background.Resize(size)

	// Center the text label
	textSize := r.label.MinSize()
	x := (size.Width - textSize.Width) / 2
	y := (size.Height - textSize.Height) / 2
	r.label.Move(fyne.NewPos(x, y))
}

// MinSize returns the minimum size required to display the text with padding.
func (r *colorButtonRenderer) MinSize() fyne.Size {
	textSize := r.label.MinSize()
	return textSize.Add(fyne.NewSize(20, 20)) // Adds 20px padding
}

// Refresh updates the visual state of the renderer's objects to match the button's properties.
func (r *colorButtonRenderer) Refresh() {
	r.background.FillColor = r.button.BGColor
	r.background.Refresh()

	r.label.Text = r.button.Text
	r.label.Refresh()
}

// Objects returns the underlying canvas objects for the Fyne painter.
func (r *colorButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy cleans up resources (no-op in this implementation).
func (r *colorButtonRenderer) Destroy() {}

