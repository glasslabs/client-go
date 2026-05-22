package client

import "encoding/xml"

// Widget is a node in a widget tree passed to Module.Content.
type Widget interface {
	widget()
}

// Text renders a single styled line of text.
type Text struct {
	XMLName   xml.Name `xml:"text"`
	Content   string   `xml:",chardata"`
	Color     string   `xml:"color,attr,omitempty"`
	FontSize  float32  `xml:"fontSize,attr,omitempty"`
	Condensed bool     `xml:"condensed,attr,omitempty"`
	Light     bool     `xml:"light,attr,omitempty"`
	Bold      bool     `xml:"bold,attr,omitempty"`
	Italic    bool     `xml:"italic,attr,omitempty"`
	Align     string   `xml:"align,attr,omitempty"`
}

func (t *Text) Equals(o *Text) bool {
	return t.Content == o.Content &&
		t.Color == o.Color &&
		t.FontSize == o.FontSize &&
		t.Condensed == o.Condensed &&
		t.Light == o.Light &&
		t.Bold == o.Bold &&
		t.Italic == o.Italic &&
		t.Align == o.Align
}

// SVG rasterizes raw SVG markup. Used for complex assets such as a floorplan.
type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Content string   `xml:",chardata"`
}

// VStack lays out children vertically, top to bottom.
type VStack struct {
	XMLName  xml.Name `xml:"vstack"`
	Children []Widget `xml:"-"`
}

// HStack lays out children horizontally, left to right.
type HStack struct {
	XMLName  xml.Name `xml:"hstack"`
	Children []Widget `xml:"-"`
}

// Spacer inserts flexible empty space inside a stack.
type Spacer struct {
	XMLName xml.Name `xml:"spacer"`
	Min     float32  `xml:"min,attr,omitempty"`
}

// Canvas renders draw operations within a fixed-size logical viewport scaled to
// fit the allocated space.
type Canvas struct {
	XMLName xml.Name `xml:"canvas"`
	Width   float32  `xml:"-"`
	Height  float32  `xml:"-"`
	Ops     []DrawOp `xml:"-"`
}

func (c *Canvas) Equals(o *Canvas) bool {
	if c.Width != o.Width || c.Height != o.Height || len(c.Ops) != len(o.Ops) {
		return false
	}
	for i, aOp := range c.Ops {
		if !drawOpEqual(aOp, o.Ops[i]) {
			return false
		}
	}
	return true
}

func drawOpEqual(a, b DrawOp) bool {
	switch av := a.(type) {
	case *Arc:
		bv, ok := b.(*Arc)
		return ok && *av == *bv
	case *Rect:
		bv, ok := b.(*Rect)
		return ok && *av == *bv
	case *Path:
		bv, ok := b.(*Path)
		return ok && *av == *bv
	case *Label:
		bv, ok := b.(*Label)
		if !ok || av.X != bv.X || av.Y != bv.Y || av.Align != bv.Align || len(av.Runs) != len(bv.Runs) {
			return false
		}
		for i, ar := range av.Runs {
			br := bv.Runs[i]
			if ar == nil || br == nil || *ar != *br {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (*Text) widget()   {}
func (*SVG) widget()    {}
func (*VStack) widget() {}
func (*HStack) widget() {}
func (*Spacer) widget() {}
func (*Canvas) widget() {}

// TextOption configures a Text widget.
type TextOption func(*Text)

// WithColor sets the text colour as an HTML hex string, e.g. "#ffffff".
func WithColor(c string) TextOption { return func(t *Text) { t.Color = c } }

// WithFontSize sets the text size in sp.
func WithFontSize(s float32) TextOption { return func(t *Text) { t.FontSize = s } }

// WithCondensed selects the Roboto Condensed typeface variant.
func WithCondensed() TextOption { return func(t *Text) { t.Condensed = true } }

// WithLight selects the light (300) weight variant.
func WithLight() TextOption { return func(t *Text) { t.Light = true } }

// WithBold selects the bold (700) weight variant.
func WithBold() TextOption { return func(t *Text) { t.Bold = true } }

// WithItalic selects italic style.
func WithItalic() TextOption { return func(t *Text) { t.Italic = true } }

// WithAlign sets the text alignment: "left", "center", or "right".
func WithAlign(a string) TextOption { return func(t *Text) { t.Align = a } }

// NewText returns a Text widget with the given content and optional style options.
func NewText(content string, opts ...TextOption) *Text {
	t := &Text{Content: content}
	for _, o := range opts {
		o(t)
	}
	return t
}

// NewSVG returns an SVG widget containing the given SVG markup.
func NewSVG(content string) *SVG {
	return &SVG{Content: content}
}

// NewVStack returns a VStack widget that lays out children top to bottom.
func NewVStack(children ...Widget) *VStack {
	return &VStack{Children: children}
}

// NewHStack returns an HStack widget that lays out children left to right.
func NewHStack(children ...Widget) *HStack {
	return &HStack{Children: children}
}

// SpacerOption configures a Spacer widget.
type SpacerOption func(*Spacer)

// WithMinSize sets the minimum size in dp that the spacer occupies along its primary axis.
func WithMinSize(min float32) SpacerOption { return func(s *Spacer) { s.Min = min } }

// NewSpacer returns a Spacer that expands to fill available space with a default minimum size of 8 dp.
func NewSpacer(opts ...SpacerOption) *Spacer {
	s := &Spacer{Min: 8}
	for _, o := range opts {
		o(s)
	}
	return s
}

// NewCanvas returns a Canvas widget with the given logical size and ordered draw list.
func NewCanvas(width, height float32, ops ...DrawOp) *Canvas {
	return &Canvas{Width: width, Height: height, Ops: ops}
}

// MarshalXML implements xml.Marshaler.
func (v *VStack) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "vstack"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, child := range v.Children {
		if child == nil {
			continue
		}
		if err := e.Encode(child); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MarshalXML implements xml.Marshaler.
func (h *HStack) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "hstack"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, child := range h.Children {
		if child == nil {
			continue
		}
		if err := e.Encode(child); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MarshalXML implements xml.Marshaler.
func (c *Canvas) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "canvas"}
	start.Attr = []xml.Attr{
		{Name: xml.Name{Local: "width"}, Value: formatFloat32(c.Width)},
		{Name: xml.Name{Local: "height"}, Value: formatFloat32(c.Height)},
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, op := range c.Ops {
		if op == nil {
			continue
		}
		if err := e.Encode(op); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}
