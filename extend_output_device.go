// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/OutputDevice.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// OutputDevice is the drawing surface. The Java type parameters
// <T extends FSImage, FontType extends FSFont> are the interfaces FSImage and
// FSFont here; an implementation asserts its own concrete types.
//
// getRenderingHint(RenderingHints.Key) and setRenderingHint(RenderingHints.Key,
// Object) are not ported: java.awt.RenderingHints has no counterpart and the
// PDF output device implements both as no-ops.
type OutputDevice interface {
	DrawText(c *RenderingContext, inlineText *InlineText)
	DrawSelection(c *RenderingContext, inlineText *InlineText)

	DrawTextDecoration(c *RenderingContext, lineBox *LineBox)
	DrawTextDecorationWithIBDecoration(c *RenderingContext, iB *InlineLayoutBox, decoration *TextDecoration)

	PaintBorder(c *RenderingContext, box BoxI)
	PaintBorderWithStyleEdgeSides(c *RenderingContext, style CalculatedStyleI, edge *geom.Rectangle, sides int)
	PaintCollapsedBorder(c *RenderingContext, border *BorderPropertySet, bounds *geom.Rectangle, side int)

	PaintBackground(c *RenderingContext, box BoxI)
	PaintBackgroundWithStyleBoundsBgImageContainerBorder(
		c *RenderingContext, style CalculatedStyleI,
		bounds *geom.Rectangle, bgImageContainer *geom.Rectangle,
		border *BorderPropertySet)

	PaintReplacedElement(c *RenderingContext, box BlockBoxI)

	DrawDebugOutline(c *RenderingContext, box BoxI, color FSColor)

	SetFont(font FSFont)

	SetColor(color FSColor)

	SetOpacity(opacity float32)

	DrawRect(x int, y int, width int, height int)
	DrawOval(x int, y int, width int, height int)

	DrawBorderLine(bounds geom.Shape, side int, width int, solid bool)

	DrawImage(image FSImage, x int, y int)

	DrawLinearGradient(gradient *FSLinearGradient, x int, y int, width int, height int)

	Draw(s geom.Shape)
	Fill(s geom.Shape)
	FillRect(x int, y int, width int, height int)
	FillOval(x int, y int, width int, height int)

	Clip(s geom.Shape)
	// GetClip returns nil when no clip is set.
	GetClip() geom.Shape
	SetClip(s geom.Shape)

	Translate(tx float64, ty float64)

	// PushTransform applies the box's CSS transform, if any, around
	// subsequent paint calls for this box and its descendants; must be paired
	// with a matching PopTransform. Java supplies a default implementation
	// that does nothing, so CSS transform only has a visual effect on
	// backends that implement this (currently PDF output only). A Go
	// implementation without transform support implements both methods with
	// an empty body.
	PushTransform(c *RenderingContext, box BoxI)

	// PopTransform restores the graphics state saved by the matching
	// PushTransform.
	PopTransform()

	SetStroke(s geom.Stroke)
	GetStroke() geom.Stroke

	IsSupportsSelection() bool

	IsSupportsCMYKColors() bool
}
