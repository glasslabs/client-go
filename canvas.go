package client

import "encoding/xml"

// DrawOp is a drawing command within a Canvas.
type DrawOp interface {
	drawOp()
}

// Arc draws a circular arc stroke.
type Arc struct {
	XMLName     xml.Name `xml:"arc"`
	Cx          float32  `xml:"cx,attr"`
	Cy          float32  `xml:"cy,attr"`
	Radius      float32  `xml:"radius,attr"`
	StartAngle  float32  `xml:"startAngle,attr"` // degrees; 0 = right, clockwise
	SweepAngle  float32  `xml:"sweepAngle,attr"` // degrees; clockwise positive
	StrokeWidth float32  `xml:"strokeWidth,attr"`
	Color       string   `xml:"color,attr"`
}

// Rect draws a filled and/or stroked rectangle.
type Rect struct {
	XMLName      xml.Name `xml:"rect"`
	X            float32  `xml:"x,attr"`
	Y            float32  `xml:"y,attr"`
	W            float32  `xml:"w,attr"`
	H            float32  `xml:"h,attr"`
	CornerRadius float32  `xml:"cornerRadius,attr,omitempty"`
	Fill         string   `xml:"fill,attr,omitempty"`
	Stroke       string   `xml:"stroke,attr,omitempty"`
	StrokeWidth  float32  `xml:"strokeWidth,attr,omitempty"`
}

// Label draws baseline-aligned multi-run text at a canvas position.
type Label struct {
	XMLName xml.Name   `xml:"label"`
	X       float32    `xml:"x,attr"`
	Y       float32    `xml:"y,attr"`
	Align   string     `xml:"align,attr,omitempty"` // "start", "middle", or "end"
	Runs    []*TextRun `xml:"run"`
}

// TextRun is a single styled text segment within a Label.
type TextRun struct {
	XMLName       xml.Name `xml:"run"`
	Content       string   `xml:",chardata"`
	FontSize      float32  `xml:"fontSize,attr,omitempty"`
	BaselineShift float32  `xml:"baselineShift,attr,omitempty"`
	Color         string   `xml:"color,attr,omitempty"`
}

// Path draws an SVG path d-string at an offset with optional uniform scale.
type Path struct {
	XMLName xml.Name `xml:"path"`
	X       float32  `xml:"x,attr,omitempty"`
	Y       float32  `xml:"y,attr,omitempty"`
	Scale   float32  `xml:"scale,attr,omitempty"`
	D       string   `xml:"d,attr"`
	Fill    string   `xml:"fill,attr,omitempty"`
}

func (*Arc) drawOp()   {}
func (*Rect) drawOp()  {}
func (*Label) drawOp() {}
func (*Path) drawOp()  {}

// RectOption configures a Rect draw operation.
type RectOption func(*Rect)

// WithFill sets the rectangle fill colour.
func WithFill(c string) RectOption { return func(r *Rect) { r.Fill = c } }

// WithStroke sets the rectangle stroke colour and width.
func WithStroke(c string, w float32) RectOption {
	return func(r *Rect) { r.Stroke = c; r.StrokeWidth = w }
}

// WithCornerRadius sets the corner radius for a rounded rectangle.
func WithCornerRadius(cr float32) RectOption { return func(r *Rect) { r.CornerRadius = cr } }

// RunOption configures a TextRun within a Label.
type RunOption func(*TextRun)

// WithRunFontSize sets the run font size in canvas units.
func WithRunFontSize(s float32) RunOption { return func(r *TextRun) { r.FontSize = s } }

// WithRunBaselineShift shifts the run down from the shared baseline in canvas units.
func WithRunBaselineShift(s float32) RunOption { return func(r *TextRun) { r.BaselineShift = s } }

// WithRunColor sets the run text colour as an HTML hex string.
func WithRunColor(c string) RunOption { return func(r *TextRun) { r.Color = c } }

// NewArc returns an Arc draw operation.
func NewArc(cx, cy, radius, startAngle, sweepAngle, strokeWidth float32, color string) *Arc {
	return &Arc{
		Cx:          cx,
		Cy:          cy,
		Radius:      radius,
		StartAngle:  startAngle,
		SweepAngle:  sweepAngle,
		StrokeWidth: strokeWidth,
		Color:       color,
	}
}

// NewRect returns a Rect draw operation.
func NewRect(x, y, w, h float32, opts ...RectOption) *Rect {
	r := &Rect{X: x, Y: y, W: w, H: h}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewLabel returns a Label draw operation.
func NewLabel(x, y float32, align string, runs ...*TextRun) *Label {
	return &Label{X: x, Y: y, Align: align, Runs: runs}
}

// NewRun returns a TextRun for use inside a Label.
func NewRun(content string, opts ...RunOption) *TextRun {
	r := &TextRun{Content: content}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewPath returns a Path draw operation.
func NewPath(x, y, scale float32, d, fill string) *Path {
	return &Path{X: x, Y: y, Scale: scale, D: d, Fill: fill}
}
