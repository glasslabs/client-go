package client

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
)

// widgetByTag maps XML element names to factory functions for Widget implementations.
// Unknown tags are skipped by callers; add new kinds here to extend the protocol.
var widgetByTag = map[string]func() Widget{
	"text":   func() Widget { return &Text{} },
	"svg":    func() Widget { return &SVG{} },
	"vstack": func() Widget { return &VStack{} },
	"hstack": func() Widget { return &HStack{} },
	"spacer": func() Widget { return &Spacer{} },
	"canvas": func() Widget { return &Canvas{} },
}

// drawOpByTag maps XML element names to factory functions for DrawOp implementations.
var drawOpByTag = map[string]func() DrawOp{
	"arc":   func() DrawOp { return &Arc{} },
	"rect":  func() DrawOp { return &Rect{} },
	"label": func() DrawOp { return &Label{} },
	"path":  func() DrawOp { return &Path{} },
}

// DecodeWidget decodes an XML-encoded Widget from b.
// The root element tag determines the concrete Widget type.
// Unknown root tags return an error; unknown child tags are silently skipped.
func DecodeWidget(b []byte) (Widget, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, fmt.Errorf("decoding widget: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		factory, ok := widgetByTag[start.Name.Local]
		if !ok {
			return nil, fmt.Errorf("unknown widget kind %q", start.Name.Local)
		}
		w := factory()
		if err = d.DecodeElement(w, &start); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", start.Name.Local, err)
		}
		return w, nil
	}
}

// UnmarshalXML decodes the VStack element and its Widget children.
func (v *VStack) UnmarshalXML(d *xml.Decoder, _ xml.StartElement) error {
	return decodeWidgetChildren(d, func(child Widget) {
		v.Children = append(v.Children, child)
	})
}

// UnmarshalXML decodes the HStack element and its Widget children.
func (h *HStack) UnmarshalXML(d *xml.Decoder, _ xml.StartElement) error {
	return decodeWidgetChildren(d, func(child Widget) {
		h.Children = append(h.Children, child)
	})
}

// UnmarshalXML decodes the Canvas element, its width/height attributes, and DrawOp children.
func (c *Canvas) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "width":
			f, _ := strconv.ParseFloat(attr.Value, 32)
			c.Width = float32(f)
		case "height":
			f, _ := strconv.ParseFloat(attr.Value, 32)
			c.Height = float32(f)
		}
	}
	return decodeDrawOpChildren(d, func(op DrawOp) {
		c.Ops = append(c.Ops, op)
	})
}

// decodeWidgetChildren reads child elements from d, instantiates a Widget for each
// known tag, and calls add for each decoded child. Unknown tags are skipped.
func decodeWidgetChildren(d *xml.Decoder, add func(Widget)) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			factory, ok := widgetByTag[t.Name.Local]
			if !ok {
				if err = d.Skip(); err != nil {
					return err
				}
				continue
			}
			child := factory()
			if err = d.DecodeElement(child, &t); err != nil {
				return err
			}
			add(child)
		case xml.EndElement:
			return nil
		}
	}
}

// decodeDrawOpChildren reads child elements from d, instantiates a DrawOp for each
// known tag, and calls add for each decoded op. Unknown tags are skipped.
func decodeDrawOpChildren(d *xml.Decoder, add func(DrawOp)) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			factory, ok := drawOpByTag[t.Name.Local]
			if !ok {
				if err = d.Skip(); err != nil {
					return err
				}
				continue
			}
			op := factory()
			if err = d.DecodeElement(op, &t); err != nil {
				return err
			}
			add(op)
		case xml.EndElement:
			return nil
		}
	}
}

// formatFloat32 formats a float32 as a decimal string with no trailing zeros.
func formatFloat32(f float32) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 32)
}
