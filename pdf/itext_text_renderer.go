// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextTextRenderer.java

package pdf

import (
	"math"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/geom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

const iTextTextRendererTextMeasuringDelta = 0.01

// ITextTextRenderer implements ufo.TextRenderer. Java's interface is generic
// in the output device, font context and font; here the arguments arrive as
// the core interfaces and are asserted to *ITextOutputDevice and *ITextFSFont.
type ITextTextRenderer struct {
}

var _ ufo.TextRenderer = (*ITextTextRenderer)(nil)

func NewITextTextRenderer() *ITextTextRenderer {
	return &ITextTextRenderer{}
}

func (r *ITextTextRenderer) Setup(context ufo.FontContext) {
}

func (r *ITextTextRenderer) DrawString(outputDevice ufo.OutputDevice, str string, x float32, y float32) {
	outputDevice.(*ITextOutputDevice).DrawString(str, x, y, nil)
}

func (r *ITextTextRenderer) DrawStringWithInfo(outputDevice ufo.OutputDevice, str string, x float32, y float32, info *ufo.JustificationInfo) {
	outputDevice.(*ITextOutputDevice).DrawString(str, x, y, info)
}

func (r *ITextTextRenderer) GetFSFontMetrics(context ufo.FontContext, font ufo.FSFont, str string) ufo.FSFontMetrics {
	description := font.(*ITextFSFont).GetFontDescription()
	bf := description.GetFont()
	size := font.GetSize2D()
	var strikethroughThickness float32
	if description.GetYStrikeoutSize() != 0 {
		strikethroughThickness = description.GetYStrikeoutSize() / 1000.0 * size
	} else {
		strikethroughThickness = size / 12.0
	}

	return NewITextFSFontMetrics(
		bf.GetFontDescriptor(writer.BaseFontBboxury, size),
		-bf.GetFontDescriptor(writer.BaseFontBboxlly, size),
		-description.GetYStrikeoutPosition()/1000.0*size,
		strikethroughThickness,
		-description.GetUnderlinePosition()/1000.0*size,
		description.GetUnderlineThickness()/1000.0*size,
	)
}

func (r *ITextTextRenderer) GetWidth(context ufo.FontContext, font ufo.FSFont, str string) int {
	itextFont := font.(*ITextFSFont)
	bf := itextFont.GetFontDescription().GetFont()
	result := bf.GetWidthPoint(str, itextFont.GetSize2D())
	if float64(result)-math.Floor(float64(result)) < iTextTextRendererTextMeasuringDelta {
		return int(result)
	} else {
		return int(math.Ceil(float64(result)))
	}
}

func (r *ITextTextRenderer) SetFontScale(scale float32) {
}

func (r *ITextTextRenderer) GetFontScale() float32 {
	return 1.0
}

func (r *ITextTextRenderer) SetSmoothingThreshold(fontsize float32) {
}

func (r *ITextTextRenderer) GetGlyphBounds(outputDevice ufo.OutputDevice, font ufo.FSFont, fsGlyphVector ufo.FSGlyphVector, index int, x float32, y float32) *geom.Rectangle {
	panic(ufo.NewXRRuntimeException("Unsupported operation: getGlyphBounds"))
}

func (r *ITextTextRenderer) GetGlyphPositions(outputDevice ufo.OutputDevice, font ufo.FSFont, fsGlyphVector ufo.FSGlyphVector) []float32 {
	panic(ufo.NewXRRuntimeException("Unsupported operation: getGlyphPositions"))
}

func (r *ITextTextRenderer) GetGlyphVector(outputDevice ufo.OutputDevice, font ufo.FSFont, str string) ufo.FSGlyphVector {
	panic(ufo.NewXRRuntimeException("Unsupported operation: getGlyphVector"))
}

func (r *ITextTextRenderer) DrawGlyphVector(outputDevice ufo.OutputDevice, vector ufo.FSGlyphVector, x float32, y float32) {
	panic(ufo.NewXRRuntimeException("Unsupported operation: drawGlyphVector"))
}
