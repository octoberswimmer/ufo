// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/FontDescription.java

package pdf

import (
	"fmt"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

const fontDescriptionDefaultFontWeight = 400

type FontDescription struct {
	style          *ufo.IdentValue
	font           *writer.BaseFont
	decorations    *FontDescriptionDecorations
	isFromFontFace bool

	// fallback supplies the font that draws the characters font cannot
	// show; nil when there is none. It is not part of the Java class.
	fallback func() (*writer.BaseFont, error)
}

// fontRun is a piece of a text drawn with one font.
type fontRun struct {
	font *writer.BaseFont
	text string
}

func NewFontDescription(font *writer.BaseFont) *FontDescription {
	return NewFontDescriptionWithIsFromFontFace(font, false)
}

func NewFontDescriptionWithIsFromFontFace(font *writer.BaseFont, isFromFontFace bool) *FontDescription {
	return NewFontDescriptionWithIsFromFontFaceStyle(font, isFromFontFace, ufo.IdentValueNormal)
}

func NewFontDescriptionWithStyleWeight(font *writer.BaseFont, style *ufo.IdentValue, weight int) *FontDescription {
	return NewFontDescriptionWithIsFromFontFaceStyleDecorations(font, false, style, fontDescriptionDefaultDecorations(font, weight))
}

func NewFontDescriptionWithIsFromFontFaceStyle(font *writer.BaseFont, isFromFontFace bool, style *ufo.IdentValue) *FontDescription {
	return NewFontDescriptionWithIsFromFontFaceStyleDecorations(font, isFromFontFace, style,
		fontDescriptionDefaultDecorations(font, fontDescriptionDefaultFontWeight))
}

func NewFontDescriptionWithIsFromFontFaceStyleDecorations(font *writer.BaseFont, isFromFontFace bool, style *ufo.IdentValue, decorations *FontDescriptionDecorations) *FontDescription {
	return &FontDescription{
		font:           font,
		isFromFontFace: isFromFontFace,
		style:          style,
		decorations:    decorations,
	}
}

func (d *FontDescription) GetFont() *writer.BaseFont {
	return d.font
}

// SetFallbackFont names the font that draws the characters this
// description's font cannot show, such as Japanese text in a description of
// Helvetica. The other characters keep the description's font, and so their
// widths. fallback is called when a text first holds such a character; the
// error it returns stops the rendering. It is not part of the Java class,
// where a character the font cannot show is left out.
func (d *FontDescription) SetFallbackFont(fallback func() (*writer.BaseFont, error)) {
	d.fallback = fallback
}

// fontFor returns the font that draws r: the fallback font when the
// description's font cannot show r and the fallback font can, the
// description's font otherwise.
func (d *FontDescription) fontFor(r rune) *writer.BaseFont {
	if d.fallback == nil || d.font.CharExists(r) {
		return d.font
	}
	fallback, err := d.fallback()
	if err != nil {
		panic(fmt.Errorf("the fallback font of %s could not be loaded: %w", d.font.GetPostscriptFontName(), err))
	}
	if fallback.CharExists(r) {
		return fallback
	}
	return d.font
}

// fontRuns splits s into the pieces drawn with one font each, in order. A
// description without a fallback font draws s whole.
func (d *FontDescription) fontRuns(s string) []fontRun {
	if d.fallback == nil {
		return []fontRun{{font: d.font, text: s}}
	}
	var runs []fontRun
	start := 0
	var current *writer.BaseFont
	for i, r := range s {
		font := d.fontFor(r)
		if font == current {
			continue
		}
		if current != nil {
			runs = append(runs, fontRun{font: current, text: s[start:i]})
		}
		current, start = font, i
	}
	if current != nil {
		runs = append(runs, fontRun{font: current, text: s[start:]})
	}
	return runs
}

func (d *FontDescription) GetWeight() int {
	return d.decorations.Weight()
}

func (d *FontDescription) GetStyle() *ufo.IdentValue {
	return d.style
}

// GetUnderlinePosition refers to the top of the underline stroke
func (d *FontDescription) GetUnderlinePosition() float32 {
	return d.decorations.UnderlinePosition()
}

func (d *FontDescription) GetUnderlineThickness() float32 {
	return d.decorations.UnderlineThickness()
}

func (d *FontDescription) GetYStrikeoutPosition() float32 {
	return d.decorations.YStrikeoutPosition()
}

func (d *FontDescription) GetYStrikeoutSize() float32 {
	return d.decorations.YStrikeoutSize()
}

func fontDescriptionDefaultDecorations(font *writer.BaseFont, weight int) *FontDescriptionDecorations {
	var underlinePosition float32 = -50
	var underlineThickness float32 = 50

	box, ok := font.GetCharBBox('x')
	if ok {
		yStrikeoutPosition := float32(box[3])/2.0 + 50
		var yStrikeoutSize float32 = 100
		return NewFontDescriptionDecorations(weight, yStrikeoutSize, yStrikeoutPosition, underlinePosition, underlineThickness)
	} else {
		// Do what the JDK does, size will be calculated by ITextTextRenderer
		yStrikeoutPosition := font.GetFontDescriptor(writer.BaseFontBboxury, 1000.0) / 3.0
		return NewFontDescriptionDecorations(weight, 0, yStrikeoutPosition, underlinePosition, underlineThickness)
	}
}

func (d *FontDescription) IsFromFontFace() bool {
	return d.isFromFontFace
}

func (d *FontDescription) String() string {
	return fmt.Sprintf("Font %s:%d", d.font.GetPostscriptFontName(), d.GetWeight())
}

func (d *FontDescription) ToString() string {
	return d.String()
}

// FontDescriptionDecorations ports the record FontDescription.Decorations.
type FontDescriptionDecorations struct {
	weight             int
	yStrikeoutSize     float32
	yStrikeoutPosition float32
	underlinePosition  float32
	underlineThickness float32
}

func NewFontDescriptionDecorations(weight int, yStrikeoutSize float32, yStrikeoutPosition float32,
	underlinePosition float32, underlineThickness float32) *FontDescriptionDecorations {
	return &FontDescriptionDecorations{
		weight:             weight,
		yStrikeoutSize:     yStrikeoutSize,
		yStrikeoutPosition: yStrikeoutPosition,
		underlinePosition:  underlinePosition,
		underlineThickness: underlineThickness,
	}
}

func (d *FontDescriptionDecorations) Weight() int {
	return d.weight
}

func (d *FontDescriptionDecorations) YStrikeoutSize() float32 {
	return d.yStrikeoutSize
}

func (d *FontDescriptionDecorations) YStrikeoutPosition() float32 {
	return d.yStrikeoutPosition
}

func (d *FontDescriptionDecorations) UnderlinePosition() float32 {
	return d.underlinePosition
}

func (d *FontDescriptionDecorations) UnderlineThickness() float32 {
	return d.underlineThickness
}
