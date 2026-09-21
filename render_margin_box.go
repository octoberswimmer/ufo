// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/MarginBox.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// MarginBox is a dummy box representing one side of the margin area of a page.
type MarginBox struct {
	BlockBox

	bounds *geom.Rectangle
}

func NewMarginBox(bounds *geom.Rectangle) *MarginBox {
	m := &MarginBox{}
	initBlockBox(&m.BlockBox, nil, nil, false)
	m.SetSelf(m)
	m.bounds = bounds
	return m
}

func (m *MarginBox) GetWidth() int {
	return m.bounds.Width
}

func (m *MarginBox) GetHeight() int {
	return m.bounds.Height
}

func (m *MarginBox) GetContentWidth() int {
	return m.bounds.Width
}

func (m *MarginBox) GetContentAreaEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	return geom.NewRectangle(-m.bounds.X, -m.bounds.Y, m.bounds.Width, m.bounds.Height)
}

func (m *MarginBox) GetPaddingEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	return geom.NewRectangle(-m.bounds.X, -m.bounds.Y, m.bounds.Width, m.bounds.Height)
}

func (m *MarginBox) GetContainingBlockWidth() int {
	return m.bounds.Width
}

func (m *MarginBox) GetPaddingWidth(cssCtx CssContext) int {
	return m.bounds.Width
}

func (m *MarginBox) CopyOf() BlockBoxI {
	panic(NewXRRuntimeException("cannot be copied"))
}
