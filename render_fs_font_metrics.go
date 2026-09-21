// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/FSFontMetrics.java

package ufo

type FSFontMetrics interface {
	GetAscent() float32

	// GetDescent follows the JDK java.awt.font.LineMetrics convention: this
	// number is positive for values below the baseline.
	GetDescent() float32

	GetStrikethroughOffset() float32

	GetStrikethroughThickness() float32

	// GetUnderlineOffset follows the JDK java.awt.font.LineMetrics convention:
	// this number is positive for values below the baseline.
	GetUnderlineOffset() float32

	GetUnderlineThickness() float32
}
