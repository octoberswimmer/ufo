// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/MarkerData.java

package ufo

// MarkerData contains the information necessary to draw a list marker. This
// includes font information from the block (for selecting the correct font
// when drawing a text marker) or the data necessary to draw other types of
// markers. It also includes a reference to the first line box in the block box
// (which in turn may be nested inside of other block boxes). All markers are
// drawn relative to the baseline of this line box.
type MarkerData struct {
	structMetrics *StrutMetrics

	// At most one of textMarker, glyphMarker and imageMarker is not nil.
	textMarker  *MarkerDataTextMarker
	glyphMarker *MarkerDataGlyphMarker
	imageMarker *MarkerDataImageMarker

	referenceLine         *LineBox
	previousReferenceLine *LineBox
}

func NewMarkerData(structMetrics *StrutMetrics, imageMarker *MarkerDataImageMarker, glyphMarker *MarkerDataGlyphMarker, textMarker *MarkerDataTextMarker) *MarkerData {
	return &MarkerData{
		structMetrics: structMetrics,
		imageMarker:   imageMarker,
		glyphMarker:   glyphMarker,
		textMarker:    textMarker,
	}
}

// GetGlyphMarker may return nil.
func (m *MarkerData) GetGlyphMarker() *MarkerDataGlyphMarker {
	return m.glyphMarker
}

// GetTextMarker may return nil.
func (m *MarkerData) GetTextMarker() *MarkerDataTextMarker {
	return m.textMarker
}

// GetImageMarker may return nil.
func (m *MarkerData) GetImageMarker() *MarkerDataImageMarker {
	return m.imageMarker
}

func (m *MarkerData) GetStructMetrics() *StrutMetrics {
	return m.structMetrics
}

func (m *MarkerData) GetLayoutWidth() int {
	if m.textMarker != nil {
		return m.textMarker.GetLayoutWidth()
	} else if m.glyphMarker != nil {
		return m.glyphMarker.GetLayoutWidth()
	} else if m.imageMarker != nil {
		return m.imageMarker.GetLayoutWidth()
	} else {
		return 0
	}
}

// GetReferenceLine may return nil.
func (m *MarkerData) GetReferenceLine() *LineBox {
	return m.referenceLine
}

func (m *MarkerData) SetReferenceLine(referenceLine *LineBox) {
	m.previousReferenceLine = m.referenceLine
	m.referenceLine = referenceLine
}

func (m *MarkerData) RestorePreviousReferenceLine(current *LineBox) {
	if current == m.referenceLine {
		m.referenceLine = m.previousReferenceLine
	}
}

type MarkerDataImageMarker struct {
	image       FSImage
	layoutWidth int
}

func NewMarkerDataImageMarker(image FSImage, layoutWidth int) *MarkerDataImageMarker {
	return &MarkerDataImageMarker{image: image, layoutWidth: layoutWidth}
}

func (i *MarkerDataImageMarker) GetImage() FSImage {
	return i.image
}

func (i *MarkerDataImageMarker) GetLayoutWidth() int {
	return i.layoutWidth
}

type MarkerDataGlyphMarker struct {
	diameter    int
	layoutWidth int
}

func NewMarkerDataGlyphMarker(diameter int, layoutWidth int) *MarkerDataGlyphMarker {
	return &MarkerDataGlyphMarker{diameter: diameter, layoutWidth: layoutWidth}
}

func (g *MarkerDataGlyphMarker) GetDiameter() int {
	return g.diameter
}

func (g *MarkerDataGlyphMarker) GetLayoutWidth() int {
	return g.layoutWidth
}

type MarkerDataTextMarker struct {
	text        string
	layoutWidth int
}

func NewMarkerDataTextMarker(text string, width int) *MarkerDataTextMarker {
	return &MarkerDataTextMarker{text: text, layoutWidth: width}
}

func (t *MarkerDataTextMarker) GetText() string {
	return t.text
}

func (t *MarkerDataTextMarker) GetLayoutWidth() int {
	return t.layoutWidth
}
