// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/AnonymousBlockBox.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// AnonymousBlockBox is an anonymous block box as defined in the CSS spec. This
// class is only used when wrapping inline content in a block box in order to
// ensure that a block box only ever contains either block or inline content.
// Other anonymous block boxes create a BlockBox directly with the anonymous
// property is true.
type AnonymousBlockBox struct {
	BlockBox

	openInlineBoxes []*InlineBox
}

func NewAnonymousBlockBox(element *dom.Element, style CalculatedStyleI, savedParents []*InlineBox,
	inlineContent []Styleable) *AnonymousBlockBox {
	a := &AnonymousBlockBox{}
	initBlockBox(&a.BlockBox, element, style, true)
	a.SetSelf(a)
	a.openInlineBoxes = savedParents
	a.SetChildrenContentType(BlockBoxContentTypeInline)
	a.SetInlineContent(inlineContent)
	return a
}

func (a *AnonymousBlockBox) Layout(c *LayoutContext) {
	a.LayoutInlineChildren(c, 0, a.CalcInitialBreakAtLine(c), true)
}

func (a *AnonymousBlockBox) GetContentWidth() int {
	return a.GetContainingBlock().GetContentWidth()
}

// Find may return nil.
func (a *AnonymousBlockBox) Find(cssCtx CssContext, absX int, absY int, findAnonymous bool) BoxI {
	result := a.BlockBox.Find(cssCtx, absX, absY, findAnonymous)
	if !findAnonymous && result == BoxI(a) {
		return a.GetParent()
	} else {
		return result
	}
}

func (a *AnonymousBlockBox) GetOpenInlineBoxes() []*InlineBox {
	return a.openInlineBoxes
}

func (a *AnonymousBlockBox) IsSkipWhenCollapsingMargins() bool {
	// An anonymous block will already have its children provided to it
	for _, styleable := range a.GetInlineContent() {
		style := styleable.GetStyle()
		if !(style.IsFloated() || style.IsAbsolute() || style.IsFixed() || style.IsRunning()) {
			return false
		}
	}
	return true
}

func (a *AnonymousBlockBox) ProvideSiblingMarginToFloats(margin int) {
	for _, styleable := range a.GetInlineContent() {
		if b, ok := styleable.(BlockBoxI); ok {
			if b.IsFloated() {
				b.GetFloatedBoxData().SetMarginFromSibling(margin)
			}
		}
	}
}

func (a *AnonymousBlockBox) IsMayCollapseMarginsWithChildren() bool {
	return false
}

func (a *AnonymousBlockBox) StyleText(c *LayoutContext) {
	a.StyleTextWithStyle(c, a.GetParent().GetStyle())
}

func (a *AnonymousBlockBox) CopyOf() BlockBoxI {
	panic(NewXRRuntimeException("cannot be copied"))
}
