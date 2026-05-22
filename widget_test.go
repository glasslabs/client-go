package client_test

import (
	"encoding/xml"
	"testing"

	"github.com/glasslabs/client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWidget_DecodeRoundtripsText(t *testing.T) {
	t.Parallel()

	original := client.NewText("15:04",
		client.WithColor("#ffffff"),
		client.WithFontSize(72),
		client.WithCondensed(),
		client.WithLight(),
		client.WithAlign("center"),
	)
	data, err := xml.Marshal(original)
	require.NoError(t, err)

	got, err := client.DecodeWidget(data)
	require.NoError(t, err)

	text, ok := got.(*client.Text)
	require.True(t, ok)
	assert.Equal(t, "15:04", text.Content)
	assert.Equal(t, "#ffffff", text.Color)
	assert.Equal(t, float32(72), text.FontSize)
	assert.True(t, text.Condensed)
	assert.True(t, text.Light)
	assert.Equal(t, "center", text.Align)
}

func TestWidget_DecodeRoundtripsVStack(t *testing.T) {
	t.Parallel()

	original := client.NewVStack(
		client.NewText("15:04", client.WithFontSize(24)),
		client.NewText("Friday 15 May", client.WithFontSize(24)),
	)
	data, err := xml.Marshal(original)
	require.NoError(t, err)

	got, err := client.DecodeWidget(data)
	require.NoError(t, err)

	vstack, ok := got.(*client.VStack)
	require.True(t, ok)
	require.Len(t, vstack.Children, 2)
	text, ok := vstack.Children[0].(*client.Text)
	require.True(t, ok)
	assert.Equal(t, "15:04", text.Content)
}

func TestWidget_DecodeRoundtripsSVG(t *testing.T) {
	t.Parallel()

	svgContent := `<svg viewBox="0 0 100 100"><circle r="40" fill="#fff"/></svg>`
	original := client.NewSVG(svgContent)
	data, err := xml.Marshal(original)
	require.NoError(t, err)

	got, err := client.DecodeWidget(data)
	require.NoError(t, err)

	svg, ok := got.(*client.SVG)
	require.True(t, ok)
	assert.Equal(t, svgContent, svg.Content)
}

func TestWidget_DecodeRoundtripsCanvas(t *testing.T) {
	t.Parallel()

	original := client.NewCanvas(300, 300,
		client.NewArc(150, 150, 125, 123, 293.9, 45, "#282828"),
		client.NewRect(105, 240, 90, 28, client.WithStroke("#646464", 1), client.WithCornerRadius(2)),
		client.NewLabel(150, 150, "middle",
			client.NewRun("3", client.WithRunFontSize(70), client.WithRunColor("#ffffff")),
		),
	)
	data, err := xml.Marshal(original)
	require.NoError(t, err)

	got, err := client.DecodeWidget(data)
	require.NoError(t, err)

	canvas, ok := got.(*client.Canvas)
	require.True(t, ok)
	assert.Equal(t, float32(300), canvas.Width)
	assert.Equal(t, float32(300), canvas.Height)
	require.Len(t, canvas.Ops, 3)

	arc, ok := canvas.Ops[0].(*client.Arc)
	require.True(t, ok)
	assert.Equal(t, float32(150), arc.Cx)
	assert.Equal(t, float32(125), arc.Radius)
	assert.Equal(t, "#282828", arc.Color)
}

func TestWidget_DecodeReturnsErrorForUnknownKind(t *testing.T) {
	t.Parallel()

	_, err := client.DecodeWidget([]byte(`<future_type content="x"/>`))
	assert.Error(t, err)
}
