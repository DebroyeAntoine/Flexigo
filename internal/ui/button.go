package ui

import (
    "image/color"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
)

// ColorButton est un bouton à fond coloré.
type ColorButton struct {
    widget.BaseWidget
    Text     string
    OnTapped func()
    BGColor  color.Color
}

// NewColorButton crée un ColorButton avec un label, un callback et une couleur de fond.
func NewColorButton(label string, onTapped func(), bgColor color.Color) *ColorButton {
    b := &ColorButton{
        Text:     label,
        OnTapped: onTapped,
        BGColor:  bgColor,
    }
    b.ExtendBaseWidget(b)
    return b
}

// CreateRenderer définit comment dessiner notre bouton personnalisé.
func (b *ColorButton) CreateRenderer() fyne.WidgetRenderer {
    // Rectangle de fond
    bg := canvas.NewRectangle(b.BGColor)
    // Texte centré
    txt := canvas.NewText(b.Text, theme.TextColor())
    txt.Alignment = fyne.TextAlignCenter

    objs := []fyne.CanvasObject{bg, txt}
    return &colorButtonRenderer{
        button:    b,
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

// TappedSecondary (clic droit) n’est pas utilisé ici.
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

