// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/LayoutUtil.java

package ufo

// LayoutUtil contains utility methods to layout floated and absolute content.
//
// XXX Could/should be folded into BlockBox

func LayoutUtilLayoutAbsolute(
	c *LayoutContext, currentLine *LineBox, box BlockBoxI) {
	markerData := c.GetCurrentMarkerData()
	c.SetCurrentMarkerData(nil)

	if box.GetStyle().IsFixed() {
		box.SetContainingBlock(c.GetRootLayer().GetMaster().GetContainingBlock())
	} else {
		box.SetContainingBlock(c.GetLayer().GetMaster())
	}
	box.SetStaticEquivalent(currentLine)

	// If printing, don't lay out until we know where it's going
	if !c.IsPrint() {
		box.Layout(c)
	} else {
		c.PushLayerBox(box)
		c.GetLayer().SetRequiresLayout(true)
		c.PopLayer()
	}

	c.SetCurrentMarkerData(markerData)
}

// LayoutUtilLayoutFloated distinguishes a nil pendingFloats (Java null: the
// float is laid out for good) from an empty non-nil slice (Java's empty list:
// the float may be marked pending). A caller that holds a list of pending
// floats therefore has to pass a non-nil slice even when the list is empty.
func LayoutUtilLayoutFloated(
	c *LayoutContext, currentLine *LineBox, block BlockBoxI,
	avail int, pendingFloats []*FloatLayoutResult) *FloatLayoutResult {
	markerData := c.GetCurrentMarkerData()
	c.SetCurrentMarkerData(nil)

	block.SetContainingBlock(currentLine.GetParent())
	block.SetContainingLayer(currentLine.GetContainingLayer())
	block.SetStaticEquivalent(currentLine)

	if pendingFloats != nil {
		block.SetY(currentLine.GetY() + block.GetFloatedBoxData().GetMarginFromSibling())
	} else {
		block.SetY(currentLine.GetY() + currentLine.GetHeight())
	}

	block.CalcInitialFloatedCanvasLocation(c)

	initialY := block.GetY()

	block.Layout(c)

	c.GetBlockFormattingContext().FloatBox(c, block)

	pending := false
	if pendingFloats != nil &&
		(len(pendingFloats) != 0 || block.GetWidth() > avail) &&
		currentLine.IsContainsContent() {
		block.Reset(c)
		pending = true
	} else {
		if c.IsPrint() {
			layoutUtilPositionFloatOnPage(c, currentLine, block, initialY != block.GetY())
			c.GetRootLayer().EnsureHasPage(c, block)
		}
	}

	c.SetCurrentMarkerData(markerData)

	return NewFloatLayoutResult(pending, block)
}

func layoutUtilPositionFloatOnPage(
	c *LayoutContext, currentLine *LineBox, block BlockBoxI,
	movedVertically bool) {
	if block.GetStyle().IsForcePageBreakBefore() {
		block.ForcePageBreakBefore(c, block.GetStyle().GetIdent(CSSNamePageBreakBefore), false)
		block.CalcCanvasLocation()
		layoutUtilResetAndFloatBlock(c, currentLine, block)
	} else if block.GetStyle().IsAvoidPageBreakInside() && block.CrossesPageBreak(c) {
		clearDelta := block.ForcePageBreakBefore(c, block.GetStyle().GetIdent(CSSNamePageBreakBefore), false)

		block.CalcCanvasLocation()
		layoutUtilResetAndFloatBlock(c, currentLine, block)

		if block.CrossesPageBreak(c) {
			block.SetY(block.GetY() - clearDelta)
			block.CalcCanvasLocation()
			layoutUtilResetAndFloatBlock(c, currentLine, block)
		}
	} else if movedVertically {
		layoutUtilResetAndFloatBlock(c, currentLine, block)
	}
}

func layoutUtilResetAndFloatBlock(c *LayoutContext, currentLine *LineBox, block BlockBoxI) {
	block.Reset(c)
	block.SetContainingLayer(currentLine.GetContainingLayer())
	block.Layout(c)
	c.GetBlockFormattingContext().FloatBox(c, block)
}
