// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/TextRenderer.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// TextRenderer measures and draws text. The Java type parameters
// (OutputDeviceType, FontContextType, FontType) are the interfaces
// OutputDevice, FontContext and FSFont here; an implementation asserts its
// own concrete types.
type TextRenderer interface {
	Setup(context FontContext)

	DrawString(outputDevice OutputDevice, str string, x float32, y float32)
	DrawStringWithInfo(outputDevice OutputDevice, str string, x float32, y float32, info *JustificationInfo)

	DrawGlyphVector(outputDevice OutputDevice, vector FSGlyphVector, x float32, y float32)

	GetGlyphVector(outputDevice OutputDevice, font FSFont, str string) FSGlyphVector

	GetGlyphPositions(outputDevice OutputDevice, font FSFont, fsGlyphVector FSGlyphVector) []float32
	GetGlyphBounds(outputDevice OutputDevice, font FSFont, fsGlyphVector FSGlyphVector, index int, x float32, y float32) *geom.Rectangle

	GetFSFontMetrics(context FontContext, font FSFont, str string) FSFontMetrics

	GetWidth(context FontContext, font FSFont, str string) int

	SetFontScale(scale float32)

	GetFontScale() float32

	// SetSmoothingThreshold sets the smoothing threshold. This is a font size
	// above which all text will be anti-aliased. Text below this size will
	// not be anti-aliased. Set to -1 for no antialiasing. Set to 0 for all
	// antialiasing. Else, set to the threshold font size. does not take font
	// scaling into account.
	SetSmoothingThreshold(fontsize float32)
}
