// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/ViewportBox.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// ViewportBox is a dummy box representing the viewport
type ViewportBox struct {
	BlockBox

	viewport *geom.Rectangle
}

func NewViewportBox(viewport *geom.Rectangle) *ViewportBox {
	v := &ViewportBox{}
	initBlockBox(&v.BlockBox, nil, nil, false)
	v.SetSelf(v)
	v.viewport = viewport
	return v
}

func (v *ViewportBox) GetWidth() int {
	return v.viewport.Width
}

func (v *ViewportBox) GetHeight() int {
	return v.viewport.Height
}

func (v *ViewportBox) GetContentWidth() int {
	return v.viewport.Width
}

func (v *ViewportBox) GetContentAreaEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	return geom.NewRectangle(-v.viewport.X, -v.viewport.Y, v.viewport.Width, v.viewport.Height)
}

func (v *ViewportBox) GetPaddingEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	return geom.NewRectangle(-v.viewport.X, -v.viewport.Y, v.viewport.Width, v.viewport.Height)
}

func (v *ViewportBox) GetPaddingWidth(cssCtx CssContext) int {
	return v.viewport.Width
}

func (v *ViewportBox) CopyOf() BlockBoxI {
	panic(NewXRRuntimeException("cannot be copied"))
}

func (v *ViewportBox) IsAutoHeight() bool {
	return false
}

func (v *ViewportBox) GetCSSHeight(c CssContext) int {
	return v.viewport.Height
}

func (v *ViewportBox) IsInitialContainingBlock() bool {
	return true
}

func (v *ViewportBox) GetExtents() *geom.Rectangle {
	return v.viewport
}
