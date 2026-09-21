// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BlockBoxing.java

package ufo

// BlockBoxing is a utility class for laying block content.  It is called when
// a block box contains block level content.  BoxBuilder will have made sure
// that the block we're working on will either contain only inline or block
// content. If we're in a paged media environment, the various page break
// related properties are also handled here.  If a rule is violated, the
// affected run of boxes will be laid out again.  If the rule still cannot be
// satisfied, the rule will be dropped.

const blockBoxingNoPageTrim = -1

func BlockBoxingLayoutContent(c *LayoutContext, block BlockBoxI, contentStart int) {
	if blockBoxingLayoutMultiColumnContent(c, block, contentStart) {
		return
	}

	offset := -1

	// Java copies the list when it is not a RandomAccess list; a Go slice
	// always is.
	localChildren := block.GetChildren()

	childOffset := block.GetHeight() + contentStart

	var relayoutDataList *blockBoxingRelayoutDataList
	if c.IsPrint() {
		relayoutDataList = newBlockBoxingRelayoutDataList(len(localChildren))
	}

	pageCount := blockBoxingNoPageTrim
	var previousChildBox BlockBoxI
	for _, localChild := range localChildren {
		child := localChild.(BlockBoxI)
		offset++

		var relayoutData *blockBoxingRelayoutData

		mayCheckKeepTogether := false
		if c.IsPrint() {
			relayoutData = relayoutDataList.get(offset)
			relayoutData.setLayoutState(c.CopyStateForRelayout())
			relayoutData.setChildOffset(childOffset)
			pageCount = len(c.GetRootLayer().GetPages())

			child.SetNeedPageClear(false)

			if (child.GetStyle().IsAvoidPageBreakInside() || child.GetStyle().IsKeepWithInline()) &&
				c.IsMayCheckKeepTogether() {
				mayCheckKeepTogether = true
				c.SetMayCheckKeepTogether(false)
			}
		}

		var layoutState *LayoutState
		if relayoutData != nil {
			layoutState = relayoutData.getLayoutState()
		}
		blockBoxingLayoutBlockChild(
			c, block, child, false, 0, childOffset, blockBoxingNoPageTrim,
			layoutState)

		if c.IsPrint() {
			needPageClear := child.IsNeedPageClear()
			if needPageClear || mayCheckKeepTogether {
				if mayCheckKeepTogether {
					c.SetMayCheckKeepTogether(true)
				}
				tryToAvoidPageBreak := child.GetStyle().IsAvoidPageBreakInside() && child.CrossesPageBreak(c)
				keepWithInline := child.IsNeedsKeepWithInline(c)
				if tryToAvoidPageBreak || needPageClear || keepWithInline {
					c.RestoreStateForRelayout(relayoutData.getLayoutState())
					child.Reset(c)
					blockBoxingLayoutBlockChild(
						c, block, child, true, 0, childOffset, pageCount, relayoutData.getLayoutState())

					if tryToAvoidPageBreak && child.CrossesPageBreak(c) && !keepWithInline {
						c.RestoreStateForRelayout(relayoutData.getLayoutState())
						child.Reset(c)
						blockBoxingLayoutBlockChild(
							c, block, child, false, 0, childOffset, pageCount, relayoutData.getLayoutState())
					}
				}
			}
			c.GetRootLayer().EnsureHasPage(c, child)
		}

		relativeOffset := child.GetRelativeOffset()
		if relativeOffset == nil {
			childOffset = child.GetY() + child.GetHeight()
		} else {
			// Box will have been positioned by this point so calculate
			// relative to where it would have been if it hadn't been
			// moved
			childOffset = child.GetY() - relativeOffset.Height + child.GetHeight()
		}

		if childOffset > block.GetHeight() {
			block.SetHeight(childOffset)
		}

		if c.IsPrint() {
			if child.GetStyle().IsForcePageBreakAfter() {
				block.ForcePageBreakAfter(c, child.GetStyle().GetIdent(CSSNamePageBreakAfter))
				childOffset = block.GetHeight()
			}

			if previousChildBox != nil {
				relayoutDataList.markRun(offset, previousChildBox, child)
			}

			runResult :=
				blockBoxingProcessPageBreakAvoidRun(
					c, block, localChildren, offset, relayoutDataList, relayoutData)
			if runResult.isChanged() {
				childOffset = runResult.getChildOffset()
				if childOffset > block.GetHeight() {
					block.SetHeight(childOffset)
				}
			}
		}

		previousChildBox = child
	}
}

func blockBoxingProcessPageBreakAvoidRun(c *LayoutContext, block BlockBoxI,
	localChildren []BoxI, offset int,
	relayoutDataList *blockBoxingRelayoutDataList, relayoutData *blockBoxingRelayoutData) *blockBoxingRelayoutRunResult {
	result := &blockBoxingRelayoutRunResult{}
	if offset > 0 {
		mightNeedRelayout := false
		runEnd := -1
		if offset == len(localChildren)-1 && relayoutData.isEndsRun() {
			mightNeedRelayout = true
			runEnd = offset
		} else if offset > 0 {
			previousRelayoutData := relayoutDataList.get(offset - 1)
			if previousRelayoutData.isEndsRun() {
				mightNeedRelayout = true
				runEnd = offset - 1
			}
		}
		if mightNeedRelayout {
			runStart := relayoutDataList.getRunStart(runEnd)
			if blockBoxingIsPageBreakBetweenChildBoxes(runStart, runEnd, c, block) {
				result.setChanged()
				block.ResetChildrenWithStartEnd(c, runStart, offset)
				result.setChildOffset(blockBoxingRelayoutRun(c, localChildren, block,
					relayoutDataList, runStart, offset, true))
				if blockBoxingIsPageBreakBetweenChildBoxes(runStart, runEnd, c, block) {
					block.ResetChildrenWithStartEnd(c, runStart, offset)
					result.setChildOffset(blockBoxingRelayoutRun(c, localChildren, block,
						relayoutDataList, runStart, offset, false))
				}
			}
		}
	}
	return result
}

func blockBoxingIsPageBreakBetweenChildBoxes(runStart int, runEnd int, c *LayoutContext, block BlockBoxI) bool {
	for i := runStart; i < runEnd; i++ {
		prevChild := block.GetChild(i)
		nextChild := block.GetChild(i + 1)
		// if nextChild is made of several lines, then only the first line
		// is relevant for "page-break-before: avoid".
		var nextLine BoxI = nextChild
		if firstLine := blockBoxingGetFirstLine(nextChild); firstLine != nil {
			nextLine = firstLine
		}
		prevChildEnd := prevChild.GetAbsY() + prevChild.GetHeight()
		nextLineEnd := nextLine.GetAbsY() + nextLine.GetHeight()
		if c.GetRootLayer().CrossesPageBreak(c, prevChildEnd, nextLineEnd) {
			return true
		}
	}
	return false
}

// blockBoxingGetFirstLine returns nil when box has no first line.
func blockBoxingGetFirstLine(box BoxI) *LineBox {
	for child := box; child.GetChildCount() > 0; child = child.GetChild(0) {
		if lineBox, ok := child.(*LineBox); ok {
			return lineBox
		}
	}
	return nil
}

func blockBoxingRelayoutRun(
	c *LayoutContext, localChildren []BoxI, block BlockBoxI,
	relayoutDataList *blockBoxingRelayoutDataList, start int, end int, onNewPage bool) int {
	childOffset := relayoutDataList.get(start).getChildOffset()

	if onNewPage {
		startBox := localChildren[start]
		startPageBox := c.GetRootLayer().GetFirstPage(c, startBox)
		childOffset += startPageBox.GetBottom() - startBox.GetAbsY()
	}

	// reset height of parent as it is used for Y-setting of children
	block.SetHeight(childOffset)

	for i := start; i <= end; i++ {
		child := localChildren[i].(BlockBoxI)

		relayoutData := relayoutDataList.get(i)

		pageCount := len(c.GetRootLayer().GetPages())

		//TODO:handle run-ins. For now, treat them as blocks

		c.RestoreStateForRelayout(relayoutData.getLayoutState())
		relayoutData.setChildOffset(childOffset)
		mayCheckKeepTogether := false
		if (child.GetStyle().IsAvoidPageBreakInside() || child.GetStyle().IsKeepWithInline()) &&
			c.IsMayCheckKeepTogether() {
			mayCheckKeepTogether = true
			c.SetMayCheckKeepTogether(false)
		}
		blockBoxingLayoutBlockChild(
			c, block, child, false, 0, childOffset, blockBoxingNoPageTrim, relayoutData.getLayoutState())

		if mayCheckKeepTogether {
			c.SetMayCheckKeepTogether(true)
			tryToAvoidPageBreak :=
				child.GetStyle().IsAvoidPageBreakInside() && child.CrossesPageBreak(c)
			needPageClear := child.IsNeedPageClear()
			keepWithInline := child.IsNeedsKeepWithInline(c)
			if tryToAvoidPageBreak || needPageClear || keepWithInline {
				c.RestoreStateForRelayout(relayoutData.getLayoutState())
				child.Reset(c)
				blockBoxingLayoutBlockChild(
					c, block, child, true, 0, childOffset, pageCount, relayoutData.getLayoutState())

				if tryToAvoidPageBreak && child.CrossesPageBreak(c) && !keepWithInline {
					c.RestoreStateForRelayout(relayoutData.getLayoutState())
					child.Reset(c)
					blockBoxingLayoutBlockChild(
						c, block, child, false, 0, childOffset, pageCount, relayoutData.getLayoutState())
				}
			}
		}

		c.GetRootLayer().EnsureHasPage(c, child)

		relativeOffset := child.GetRelativeOffset()
		if relativeOffset == nil {
			childOffset = child.GetY() + child.GetHeight()
		} else {
			childOffset = child.GetY() - relativeOffset.Height + child.GetHeight()
		}

		if childOffset > block.GetHeight() {
			block.SetHeight(childOffset)
		}

		if child.GetStyle().IsForcePageBreakAfter() {
			block.ForcePageBreakAfter(c, child.GetStyle().GetIdent(CSSNamePageBreakAfter))
			childOffset = block.GetHeight()
		}
	}

	return childOffset
}

func blockBoxingLayoutBlockChild(
	c *LayoutContext, parent BlockBoxI, child BlockBoxI,
	needPageClear bool, childX int, childOffset int, trimmedPageCount int, layoutState *LayoutState) {
	blockBoxingLayoutBlockChild0(c, parent, child, needPageClear, childX, childOffset, trimmedPageCount)
	bContext := child.CalcBreakAtLineContext(c)
	if bContext != nil {
		c.SetBreakAtLineContext(bContext)
		c.RestoreStateForRelayout(layoutState)
		child.Reset(c)
		blockBoxingLayoutBlockChild0(c, parent, child, needPageClear, childX, childOffset, trimmedPageCount)
		c.SetBreakAtLineContext(nil)
	}
}

func blockBoxingLayoutBlockChild0(c *LayoutContext, parent BlockBoxI, child BlockBoxI,
	needPageClear bool, childX int, childOffset int, trimmedPageCount int) {
	child.SetNeedPageClear(needPageClear)

	child.InitStaticPos(c, parent, childOffset)
	child.SetX(childX)

	child.InitContainingLayer(c)
	child.CalcCanvasLocation()

	c.Translate(childX, childOffset)
	blockBoxingRepositionBox(c, child, trimmedPageCount)
	child.Layout(c)
	c.Translate(-child.GetX(), -child.GetY())
}

func blockBoxingLayoutMultiColumnContent(c *LayoutContext, block BlockBoxI, contentStart int) bool {
	columnCount := blockBoxingResolveColumnCount(c, block)
	if columnCount <= 1 {
		return false
	}

	columnGap := blockBoxingResolveColumnGap(c, block)
	savedContentWidth := block.GetContentWidth()
	columnsAndGapsWidth := savedContentWidth - max(0, (columnCount-1)*columnGap)
	columnWidth := max(1, columnsAndGapsWidth/columnCount)
	columnTop := block.GetHeight() + contentStart

	children := block.GetChildren()
	func() {
		// Java: try { ... } finally { block.setContentWidth(savedContentWidth); }
		defer block.SetContentWidth(savedContentWidth)

		// Children must resolve widths against column width, not the multicol element's full width.
		block.SetContentWidth(columnWidth)

		columnHeight := blockBoxingResolveColumnHeightForMultiColumn(
			c, block, contentStart, columnCount, children, columnTop)

		columnIndex := 0
		childOffset := columnTop
		maxColumnBottom := columnTop

		for _, localChild := range children {
			child := localChild.(BlockBoxI)
			var state *LayoutState
			if c.IsPrint() {
				state = c.CopyStateForRelayout()
			}
			startOffset := childOffset
			childX := columnIndex * (columnWidth + columnGap)

			blockBoxingLayoutBlockChild(c, block, child, false, childX, childOffset, blockBoxingNoPageTrim, state)
			childOffset = blockBoxingNextBlockChildOffset(child)

			overflowedColumn := childOffset > columnTop+columnHeight
			canMoveToNextColumn := columnIndex < columnCount-1
			isFirstInColumn := startOffset == columnTop
			if overflowedColumn && canMoveToNextColumn && !isFirstInColumn {
				if state != nil {
					c.RestoreStateForRelayout(state)
				}
				child.Reset(c)
				columnIndex++
				childOffset = columnTop
				childX = columnIndex * (columnWidth + columnGap)
				blockBoxingLayoutBlockChild(c, block, child, false, childX, childOffset, blockBoxingNoPageTrim, state)
				childOffset = blockBoxingNextBlockChildOffset(child)
			}

			maxColumnBottom = max(maxColumnBottom, childOffset)
		}

		block.SetHeight(max(block.GetHeight(), maxColumnBottom))
	}()
	return true
}

func blockBoxingNextBlockChildOffset(child BlockBoxI) int {
	relativeOffset := child.GetRelativeOffset()
	if relativeOffset == nil {
		return child.GetY() + child.GetHeight()
	}
	return child.GetY() - relativeOffset.Height + child.GetHeight()
}

// blockBoxingResolveColumnHeightForMultiColumn gives the column height for
// balancing: when the multicol block has auto height, split total content
// height across columns so columns fill sequentially (CSS multi-column balance
// behavior). When height is specified, use that as the column box height.
func blockBoxingResolveColumnHeightForMultiColumn(
	c *LayoutContext, block BlockBoxI, contentStart int, columnCount int,
	children []BoxI, columnTop int) int {
	if !block.GetStyle().IsAutoHeight() && block.GetStyle().HasAbsoluteUnit(CSSNameHeight) {
		specifiedHeight := calculatedStyleFloatToInt(block.GetStyle().GetFloatPropertyProportionalHeight(CSSNameHeight, 0, c))
		if specifiedHeight > 0 {
			return specifiedHeight
		}
	}

	measureOffset := columnTop
	for _, localChild := range children {
		child := localChild.(BlockBoxI)
		var state *LayoutState
		if c.IsPrint() {
			state = c.CopyStateForRelayout()
		}
		blockBoxingLayoutBlockChild(c, block, child, false, 0, measureOffset, blockBoxingNoPageTrim, state)
		measureOffset = blockBoxingNextBlockChildOffset(child)
	}
	totalHeight := measureOffset - columnTop
	for _, localChild := range children {
		localChild.Reset(c)
	}
	if totalHeight <= 0 || len(children) == 0 {
		return max(1, (totalHeight+columnCount-1)/columnCount)
	}

	lo := max(1, (totalHeight+columnCount-1)/columnCount) // ceil(total/n)
	hi := totalHeight
	bestH := hi // safe fallback

	for lo <= hi {
		mid := lo + (hi-lo)/2

		if blockBoxingContentFitsInColumns(c, block, children, columnTop, columnCount, mid) {
			bestH = mid // mid works → try smaller
			hi = mid - 1
		} else {
			lo = mid + 1 // mid is too small → try larger
		}
	}

	bestH = blockBoxingApplyFragmentationRules(c, block, children, columnTop, columnCount, bestH)
	return max(1, bestH)
}

func blockBoxingContentFitsInColumns(
	c *LayoutContext, block BlockBoxI, children []BoxI,
	columnTop int, columnCount int, columnH int) bool {

	usedColumns := 1
	columnUsed := 0 // height consumed in the current column

	measureOffset := columnTop
	for _, localChild := range children {
		child := localChild.(BlockBoxI)

		// Lay out the child to get its real height.
		var state *LayoutState
		if c.IsPrint() {
			state = c.CopyStateForRelayout()
		}
		blockBoxingLayoutBlockChild(c, block, child, false, 0, measureOffset, blockBoxingNoPageTrim, state)
		childHeight := child.GetHeight()
		child.Reset(c)

		// Would this child overflow the current column?
		if columnUsed+childHeight > columnH {
			// Move to the next column.
			usedColumns++
			if usedColumns > columnCount {
				return false // needs more columns than allowed → H too small
			}
			columnUsed = childHeight // child starts fresh in the new column
		} else {
			columnUsed += childHeight
		}

		measureOffset = columnTop + columnUsed // keep offset in sync
	}

	return true // all children fit within columnCount columns
}

func blockBoxingApplyFragmentationRules(
	c *LayoutContext, block BlockBoxI, children []BoxI,
	columnTop int, columnCount int, initialH int) int {

	columnH := initialH
	changed := true

	// Iterate until no more adjustments are needed (usually 1–2 passes).
	for changed {
		changed = false
		columnUsed := 0

		measureOffset := columnTop
		for _, localChild := range children {
			child := localChild.(BlockBoxI)

			// Lay out child to get its real height.
			var state *LayoutState
			if c.IsPrint() {
				state = c.CopyStateForRelayout()
			}
			blockBoxingLayoutBlockChild(c, block, child, false, 0, measureOffset, blockBoxingNoPageTrim, state)
			childHeight := child.GetHeight()
			child.Reset(c)

			avoidBreakInside := blockBoxingIsBreakInsideAvoid(child)

			if columnUsed+childHeight > columnH {
				// A column break would occur inside this child.
				if avoidBreakInside && childHeight <= columnH {
					// The child fits in one column on its own but is being split.
					// Bump H so the break happens *before* this child (i.e. the
					// previous column is extended to exactly columnH, and this
					// child starts the next column whole).
					// We only need to ensure the previous column's used space
					// equals columnH — no change to columnH is needed here;
					// the child will simply start in the next column.
					// BUT if the child itself is taller than columnH we cannot
					// avoid the split — leave it as-is.
					columnUsed = childHeight // child starts fresh in next column
				} else if avoidBreakInside && childHeight > columnH {
					// Child is taller than one column; we must expand columnH
					// to at least childHeight so it is never split.
					newH := childHeight
					if newH > columnH {
						columnH = newH
						changed = true
						break // restart the pass with the new columnH
					}
					columnUsed = childHeight
				} else {
					// No avoid constraint — normal column break.
					columnUsed = childHeight
				}
			} else {
				columnUsed += childHeight
			}

			measureOffset = columnTop + columnUsed
		}
	}

	return columnH
}

func blockBoxingIsBreakInsideAvoid(child BlockBoxI) bool {
	style := child.GetStyle()

	// CSS 2.1 legacy: page-break-inside
	if style.IsIdent(CSSNamePageBreakInside, IdentValueAvoid) {
		return true
	}

	return false
}

func blockBoxingResolveColumnCount(c *LayoutContext, block BlockBoxI) int {
	if !block.GetStyle().IsIdent(CSSNameColumnCount, IdentValueAuto) {
		return max(1, calculatedStyleFloatToInt(block.GetStyle().AsFloat(CSSNameColumnCount)))
	}

	if block.GetStyle().IsIdent(CSSNameColumnWidth, IdentValueAuto) {
		return 1
	}

	width := max(1, calculatedStyleFloatToInt(block.GetStyle().GetFloatPropertyProportionalWidth(
		CSSNameColumnWidth, float32(block.GetContentWidth()), c)))
	gap := blockBoxingResolveColumnGap(c, block)
	return max(1, (block.GetContentWidth()+gap)/(width+gap))
}

func blockBoxingResolveColumnGap(c *LayoutContext, block BlockBoxI) int {
	if block.GetStyle().IsIdent(CSSNameColumnGap, IdentValueNormal) {
		return 16
	}
	return max(0, calculatedStyleFloatToInt(block.GetStyle().GetFloatPropertyProportionalWidth(
		CSSNameColumnGap, float32(block.GetContentWidth()), c)))
}

func blockBoxingRepositionBox(c *LayoutContext, child BlockBoxI, trimmedPageCount int) {
	moved := false
	if child.GetStyle().IsRelative() {
		delta := child.PositionRelative(c)
		c.Translate(delta.Width, delta.Height)
		moved = true
	}
	if c.IsPrint() {
		pageClear := child.IsNeedPageClear() ||
			child.GetStyle().IsForcePageBreakBefore()
		needNewPageContext := child.CheckPageContext(c)

		if needNewPageContext && trimmedPageCount != blockBoxingNoPageTrim {
			c.GetRootLayer().TrimPageCount(trimmedPageCount)
		}

		if pageClear || needNewPageContext {
			delta := child.ForcePageBreakBefore(
				c,
				child.GetStyle().GetIdent(CSSNamePageBreakBefore),
				needNewPageContext)
			c.Translate(0, delta)
			moved = true
			child.SetNeedPageClear(false)
		}
	}
	if moved {
		child.CalcCanvasLocation()
	}
}

// blockBoxingRelayoutDataList is the private class BlockBoxing.RelayoutDataList.
type blockBoxingRelayoutDataList struct {
	hints []*blockBoxingRelayoutData
}

func newBlockBoxingRelayoutDataList(size int) *blockBoxingRelayoutDataList {
	l := &blockBoxingRelayoutDataList{hints: make([]*blockBoxingRelayoutData, 0, size)}
	for i := 0; i < size; i++ {
		l.hints = append(l.hints, &blockBoxingRelayoutData{})
	}
	return l
}

func (l *blockBoxingRelayoutDataList) get(index int) *blockBoxingRelayoutData {
	return l.hints[index]
}

func (l *blockBoxingRelayoutDataList) markRun(offset int, previous BlockBoxI, current BlockBoxI) {
	previousData := l.get(offset - 1)
	currentData := l.get(offset)

	previousAfter :=
		previous.GetStyle().GetIdent(CSSNamePageBreakAfter)
	currentBefore :=
		current.GetStyle().GetIdent(CSSNamePageBreakBefore)

	if previousAfter == IdentValueAvoid && currentBefore == IdentValueAuto ||
		previousAfter == IdentValueAuto && currentBefore == IdentValueAvoid ||
		previousAfter == IdentValueAvoid && currentBefore == IdentValueAvoid {
		if !previousData.isInRun() {
			previousData.setStartsRun()
		}
		previousData.setInRun()
		currentData.setInRun()

		if offset == len(l.hints)-1 {
			currentData.setEndsRun()
		}
	} else {
		if previousData.isInRun() {
			previousData.setEndsRun()
		}
	}
}

func (l *blockBoxingRelayoutDataList) getRunStart(runEnd int) int {
	offset := runEnd
	current := l.get(offset)
	if !current.isEndsRun() {
		panic(NewXRRuntimeException("Not the end of a run"))
	}
	for !current.isStartsRun() {
		offset--
		current = l.get(offset)
	}
	return offset
}

// blockBoxingRelayoutRunResult is the private class BlockBoxing.RelayoutRunResult.
type blockBoxingRelayoutRunResult struct {
	changed     bool
	childOffset int
}

func (r *blockBoxingRelayoutRunResult) isChanged() bool {
	return r.changed
}

func (r *blockBoxingRelayoutRunResult) setChanged() {
	r.changed = true
}

func (r *blockBoxingRelayoutRunResult) getChildOffset() int {
	return r.childOffset
}

func (r *blockBoxingRelayoutRunResult) setChildOffset(childOffset int) {
	r.childOffset = childOffset
}

// blockBoxingRelayoutData is the private class BlockBoxing.RelayoutData.
type blockBoxingRelayoutData struct {
	layoutState *LayoutState

	startsRun bool
	endsRun   bool
	inRun     bool

	childOffset int
}

func (d *blockBoxingRelayoutData) isEndsRun() bool {
	return d.endsRun
}

func (d *blockBoxingRelayoutData) setEndsRun() {
	d.endsRun = true
}

func (d *blockBoxingRelayoutData) isInRun() bool {
	return d.inRun
}

func (d *blockBoxingRelayoutData) setInRun() {
	d.inRun = true
}

func (d *blockBoxingRelayoutData) getLayoutState() *LayoutState {
	return d.layoutState
}

func (d *blockBoxingRelayoutData) setLayoutState(layoutState *LayoutState) {
	d.layoutState = layoutState
}

func (d *blockBoxingRelayoutData) isStartsRun() bool {
	return d.startsRun
}

func (d *blockBoxingRelayoutData) setStartsRun() {
	d.startsRun = true
}

func (d *blockBoxingRelayoutData) getChildOffset() int {
	return d.childOffset
}

func (d *blockBoxingRelayoutData) setChildOffset(childOffset int) {
	d.childOffset = childOffset
}
