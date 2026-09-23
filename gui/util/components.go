package util

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SurfaceRadius is the corner radius used for panels on the main window.
const SurfaceRadius = 10

// Surface is a rounded, bordered panel. Its colours are read from the active
// theme on every refresh, so switching light/dark mode needs no rebuild.
type Surface struct {
	widget.BaseWidget
	Content fyne.CanvasObject
}

// NewSurface wraps content in a themed panel.
func NewSurface(content fyne.CanvasObject) *Surface {
	s := &Surface{Content: content}
	s.ExtendBaseWidget(s)
	return s
}

func (s *Surface) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = SurfaceRadius
	bg.StrokeWidth = 1
	r := &surfaceRenderer{bg: bg, s: s, pad: container.New(&insetLayout{inset: 14}, s.Content)}
	r.Refresh()
	return r
}

type surfaceRenderer struct {
	s   *Surface
	bg  *canvas.Rectangle
	pad *fyne.Container
}

func (r *surfaceRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.pad.Resize(size)
}

func (r *surfaceRenderer) MinSize() fyne.Size { return r.pad.MinSize() }

func (r *surfaceRenderer) Refresh() {
	r.bg.FillColor = CurrentCardBackground()
	r.bg.StrokeColor = CurrentBorderColor()
	r.bg.Refresh()
	r.pad.Refresh()
}

func (r *surfaceRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.pad}
}

func (r *surfaceRenderer) Destroy() {}

// insetLayout pads its single child by a fixed amount on every side.
type insetLayout struct{ inset float32 }

func (l *insetLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objs {
		o.Move(fyne.NewPos(l.inset, l.inset))
		o.Resize(fyne.NewSize(size.Width-2*l.inset, size.Height-2*l.inset))
	}
}

func (l *insetLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, o := range objs {
		min = min.Max(o.MinSize())
	}
	return min.Add(fyne.NewSize(2*l.inset, 2*l.inset))
}

// Inset pads content by a fixed amount on every side.
func Inset(inset float32, content fyne.CanvasObject) *fyne.Container {
	return container.New(&insetLayout{inset: inset}, content)
}

// StepHeader renders a numbered badge followed by a bold title.
func StepHeader(number, title string) fyne.CanvasObject {
	circle := canvas.NewCircle(ColorPrimary)
	num := canvas.NewText(number, color.White)
	num.TextStyle = fyne.TextStyle{Bold: true}
	num.TextSize = 13
	num.Alignment = fyne.TextAlignCenter
	badge := container.NewGridWrap(fyne.NewSize(24, 24), container.NewStack(circle, container.NewCenter(num)))

	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.SizeName = theme.SizeNameSubHeadingText
	return container.NewBorder(nil, nil, container.NewCenter(badge), nil, label)
}

// BigIcon shows a theme icon at a fixed, larger size.
func BigIcon(res fyne.Resource, size float32) (*widget.Icon, fyne.CanvasObject) {
	icon := widget.NewIcon(res)
	return icon, container.NewCenter(container.NewGridWrap(fyne.NewSize(size, size), icon))
}

// SelectionTile is the large "what is selected" block shown in each step.
type SelectionTile struct {
	Icon   *widget.Icon
	Title  *widget.Label
	Detail *widget.Label
	Object fyne.CanvasObject
}

// NewSelectionTile builds a tile showing an icon, a title and a detail line.
func NewSelectionTile(res fyne.Resource, title, detail string) *SelectionTile {
	icon, iconBox := BigIcon(res, 48)

	t := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	t.Truncation = fyne.TextTruncateEllipsis

	d := widget.NewLabel(detail)
	d.Alignment = fyne.TextAlignCenter
	d.Truncation = fyne.TextTruncateEllipsis
	d.Importance = widget.LowImportance
	d.SizeName = theme.SizeNameCaptionText

	return &SelectionTile{
		Icon:   icon,
		Title:  t,
		Detail: d,
		Object: container.NewVBox(iconBox, t, d),
	}
}

// Set updates every part of the tile at once.
func (t *SelectionTile) Set(res fyne.Resource, title, detail string) {
	t.Icon.SetResource(res)
	t.Title.SetText(title)
	t.Detail.SetText(detail)
}

// Notice is an inline message with a coloured icon aligned to its first line.
func Notice(res fyne.Resource, text string) (*widget.Label, fyne.CanvasObject) {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.SizeName = theme.SizeNameCaptionText

	// Match the height of one line of label text so the icon centres on it.
	line := widget.NewLabel("X")
	line.SizeName = theme.SizeNameCaptionText
	lineHeight := line.MinSize().Height

	icon := widget.NewIcon(res)
	iconBox := container.NewGridWrap(fyne.NewSize(18, lineHeight), container.NewCenter(
		container.NewGridWrap(fyne.NewSize(18, 18), icon)))
	return label, container.NewBorder(nil, nil, container.NewVBox(iconBox), nil, label)
}

// MinWidth forces a minimum width on content.
func MinWidth(width float32, content fyne.CanvasObject) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(width, 0))
	return container.NewStack(spacer, content)
}

// TallButton gives a button extra height so it reads as the main action.
func TallButton(btn *widget.Button, height float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, height))
	return container.NewStack(spacer, btn)
}
