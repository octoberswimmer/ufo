// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/InlinePaintable.java

package ufo

// InlinePaintable indicates that a box is able to paint itself in an
// inline context.  This includes lines and laid out inline content, but also
// block content which participates in an inline formatting context.
type InlinePaintable interface {
	PaintInline(c *RenderingContext)
}
