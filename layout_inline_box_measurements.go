// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/InlineBoxMeasurements.java

package ufo

// InlineBoxMeasurements is a bean which tracks various characteristics of an
// inline box.  It is used when calculating the vertical position of boxes in a
// line.
type InlineBoxMeasurements struct {
	textTop      int
	textBottom   int
	baseline     int
	inlineTop    int
	inlineBottom int

	paintingTop    int
	paintingBottom int
}

func NewInlineBoxMeasurements(baseline int,
	textTop int, textBottom int,
	inlineTop int, inlineBottom int,
	paintingTop int, paintingBottom int) *InlineBoxMeasurements {
	return &InlineBoxMeasurements{
		textTop:        textTop,
		textBottom:     textBottom,
		baseline:       baseline,
		inlineTop:      inlineTop,
		inlineBottom:   inlineBottom,
		paintingTop:    paintingTop,
		paintingBottom: paintingBottom,
	}
}

func (m *InlineBoxMeasurements) GetBaseline() int {
	return m.baseline
}

func (m *InlineBoxMeasurements) GetInlineBottom() int {
	return m.inlineBottom
}

func (m *InlineBoxMeasurements) GetInlineTop() int {
	return m.inlineTop
}

func (m *InlineBoxMeasurements) GetTextBottom() int {
	return m.textBottom
}

func (m *InlineBoxMeasurements) GetTextTop() int {
	return m.textTop
}

func (m *InlineBoxMeasurements) GetPaintingBottom() int {
	return m.paintingBottom
}

func (m *InlineBoxMeasurements) GetPaintingTop() int {
	return m.paintingTop
}
