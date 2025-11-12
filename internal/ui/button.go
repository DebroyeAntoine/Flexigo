package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ColorButton est un bouton à fond coloré avec support du highlight.
type ColorButton struct {
	widget.BaseWidget
	Text           string
	OnTapped       func()
	BGColor        color.Color
	OriginalColor  color.Color
	HighlightColor color.Color // Couleur lors du scan/highlight
	IsHighlighted  bool        // État actuel du highlight
}

// NewColorButton crée un ColorButton avec un label, un callback et une couleur de fond.
func NewColorButton(label string, onTapped func(), bgColor color.Color) *ColorButton {
	b := &ColorButton{
		Text:           label,
		OnTapped:       onTapped,
		BGColor:        bgColor,
		OriginalColor:  bgColor,
		HighlightColor: color.RGBA{R: 0, G: 0, B: 255, A: 255}, // Bleu par défaut
		IsHighlighted:  false,
	}
	b.ExtendBaseWidget(b)
	return b
}

// Highlight met le bouton en surbrillance
func (b *ColorButton) Highlight() {
	b.IsHighlighted = true
	b.BGColor = b.HighlightColor
	b.Refresh()
}

// Unhighlight retire la surbrillance
func (b *ColorButton) Unhighlight() {
	b.IsHighlighted = false
	b.BGColor = b.OriginalColor
	b.Refresh()
}

// SetHighlightColor définit la couleur de surbrillance
func (b *ColorButton) SetHighlightColor(c color.Color) {
	b.HighlightColor = c
}

// CreateRenderer définit comment dessiner notre bouton personnalisé.
func (b *ColorButton) CreateRenderer() fyne.WidgetRenderer {
	// Rectangle de fond
	bg := canvas.NewRectangle(b.BGColor)
	// Texte centré
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

// Tapped appelle le callback quand on clique.
func (b *ColorButton) Tapped(_ *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// TappedSecondary (clic droit) n'est pas utilisé ici.
func (b *ColorButton) TappedSecondary(_ *fyne.PointEvent) {}

// colorButtonRenderer gère le layout, le rafraîchissement et le sizing.
type colorButtonRenderer struct {
	button     *ColorButton
	background *canvas.Rectangle
	label      *canvas.Text
	objects    []fyne.CanvasObject
}

func (r *colorButtonRenderer) Layout(size fyne.Size) {
	// Fond qui occupe tout
	r.background.Resize(size)

	// Texte centré
	textSize := r.label.MinSize()
	x := (size.Width - textSize.Width) / 2
	y := (size.Height - textSize.Height) / 2
	r.label.Move(fyne.NewPos(x, y))
}

func (r *colorButtonRenderer) MinSize() fyne.Size {
	// On prend la taille minimale du texte + un peu de padding
	textSize := r.label.MinSize()
	return textSize.Add(fyne.NewSize(20, 20))
}

func (r *colorButtonRenderer) Refresh() {
	// Met à jour la couleur (au cas où) et le texte
	r.background.FillColor = r.button.BGColor
	r.background.Refresh()

	r.label.Text = r.button.Text
	r.label.Refresh()
}

func (r *colorButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *colorButtonRenderer) Destroy() {}
