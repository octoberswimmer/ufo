// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextFSFontMetrics.java

package pdf

import "github.com/octoberswimmer/ufo"

type ITextFSFontMetrics struct {
	ascent                 float32
	descent                float32
	strikethroughOffset    float32
	strikethroughThickness float32
	underlineOffset        float32
	underlineThickness     float32
}

var _ ufo.FSFontMetrics = (*ITextFSFontMetrics)(nil)

func NewITextFSFontMetrics(ascent float32, descent float32, strikethroughOffset float32, strikethroughThickness float32,
	underlineOffset float32, underlineThickness float32) *ITextFSFontMetrics {
	return &ITextFSFontMetrics{
		ascent:                 ascent,
		descent:                descent,
		strikethroughOffset:    strikethroughOffset,
		strikethroughThickness: strikethroughThickness,
		underlineOffset:        underlineOffset,
		underlineThickness:     underlineThickness,
	}
}

func (m *ITextFSFontMetrics) GetAscent() float32 {
	return m.ascent
}

func (m *ITextFSFontMetrics) GetDescent() float32 {
	return m.descent
}

func (m *ITextFSFontMetrics) GetStrikethroughOffset() float32 {
	return m.strikethroughOffset
}

func (m *ITextFSFontMetrics) GetStrikethroughThickness() float32 {
	return m.strikethroughThickness
}

func (m *ITextFSFontMetrics) GetUnderlineOffset() float32 {
	return m.underlineOffset
}

func (m *ITextFSFontMetrics) GetUnderlineThickness() float32 {
	return m.underlineThickness
}
