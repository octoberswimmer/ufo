// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/CssContext.java

package ufo

type CssContext interface {
	GetMmPerDot() float32

	GetDotsPerPixel() int

	GetFontSize2D(font *FontSpecification) float32

	GetXHeight(parentFont *FontSpecification) float32

	// GetFont may return nil.
	GetFont(font *FontSpecification) FSFont

	// FIXME Doesn't really belong here, but this is
	// the only common interface of LayoutContext
	// and RenderingContext
	GetCss() *StyleReference

	GetTextRenderer() TextRenderer

	GetFontContext() FontContext

	GetFSFontMetrics(font FSFont) FSFontMetrics
}
