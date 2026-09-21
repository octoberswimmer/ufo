// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/TableBox.java

package ufo

import (
	"sort"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

// Much of this code is directly inspired by (and even copied from)
// the equivalent code in KHTML (including the idea of "effective columns" to
// manage colspans and the details of the table layout algorithms).  Many kudos
// to the KHTML developers for making such an amazing piece of software!
type TableBox struct {
	BlockBox

	columns     []*ColumnData
	columnPos   []int
	tableLayout TableBoxTableLayout

	styleColumns []*TableColumn

	pageClearance int

	marginAreaRoot bool

	contentLimitContainer *ContentLimitContainer

	extraSpaceTop    int
	extraSpaceBottom int
}

// NewTableBox ports the constructor. Java's Box constructor calls the
// overridable setStyle, which for a table chooses the table layout strategy;
// here the BlockBox initializer cannot dispatch to TableBox.SetStyle (self is
// not set yet), so the constructor calls SetStyle itself once self is set.
func NewTableBox(element *dom.Element, style CalculatedStyleI, anonymous bool) *TableBox {
	t := &TableBox{}
	initBlockBox(&t.BlockBox, element, style, anonymous)
	t.SetSelf(t)
	t.SetStyle(style)
	return t
}

func (t *TableBox) IsMarginAreaRoot() bool {
	return t.marginAreaRoot
}

func (t *TableBox) SetMarginAreaRoot(marginAreaRoot bool) {
	t.marginAreaRoot = marginAreaRoot
}

func (t *TableBox) CopyOf() BlockBoxI {
	return NewTableBox(t.GetElement(), t.GetStyle(), t.IsAnonymous())
}

func (t *TableBox) AddStyleColumn(col *TableColumn) {
	t.styleColumns = append(t.styleColumns, col)
}

func (t *TableBox) GetStyleColumns() []*TableColumn {
	return t.styleColumns
}

func (t *TableBox) GetColumnPos() []int {
	return ArrayUtilCloneOrEmptyIntArray(t.columnPos)
}

func (t *TableBox) setColumnPos(columnPos []int) {
	t.columnPos = columnPos
}

func (t *TableBox) NumEffCols() int {
	return len(t.columns)
}

func (t *TableBox) SpanOfEffCol(effCol int) int {
	return t.columns[effCol].GetSpan()
}

func (t *TableBox) ColToEffCol(col int) int {
	c := 0
	i := 0
	for c < col && i < t.NumEffCols() {
		c += t.SpanOfEffCol(i)
		i++
	}
	return i
}

func (t *TableBox) EffColToCol(effCol int) int {
	c := 0
	for i := 0; i < effCol; i++ {
		c += t.SpanOfEffCol(i)
	}
	return c
}

func (t *TableBox) AppendColumn(span int) {
	data := NewColumnData()
	data.SetSpan(span)

	t.columns = append(t.columns, data)

	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		section.ExtendGridToColumnCount(len(t.columns))
	}
}

func (t *TableBox) SetStyle(style CalculatedStyleI) {
	t.BlockBox.SetStyle(style)

	if t.IsMarginAreaRoot() {
		t.tableLayout = newTableBoxMarginTableLayout(t)
	} else if style.IsIdent(CSSNameTableLayout, IdentValueAuto) || style.IsAutoWidth() {
		t.tableLayout = newTableBoxAutoTableLayout(t)
	} else {
		t.tableLayout = newTableBoxFixedTableLayout(t)
	}
}

func (t *TableBox) CalcMinMaxWidth(c *LayoutContext) {
	if !t.IsMinMaxCalculated() {
		t.recalcSections(c)
		if t.GetStyle().IsCollapseBorders() {
			t.calcBorders(c)
		}
		t.tableLayout.CalcMinMaxWidth(c)
		t.SetMinMaxCalculated(true)
	}
}

func (t *TableBox) SplitColumn(pos int, firstSpan int) {
	newColumn := NewColumnData()
	newColumn.SetSpan(firstSpan)
	t.columns = append(t.columns, nil)
	copy(t.columns[pos+1:], t.columns[pos:])
	t.columns[pos] = newColumn

	leftOver := t.columns[pos+1]
	leftOver.SetSpan(leftOver.GetSpan() - firstSpan)

	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		section.SplitColumn(pos)
	}
}

func (t *TableBox) MarginsBordersPaddingAndSpacing(c CssContext, ignoreAutoMargins bool) int {
	result := 0
	margin := t.GetMargin(c)
	if !ignoreAutoMargins || !t.GetStyle().IsAutoLeftMargin() {
		result += calculatedStyleFloatToInt(margin.Left())
	}
	if !ignoreAutoMargins || !t.GetStyle().IsAutoRightMargin() {
		result += calculatedStyleFloatToInt(margin.Right())
	}
	border := t.GetBorder(c)
	result += calculatedStyleFloatToInt(border.Left()) + calculatedStyleFloatToInt(border.Right())
	if !t.GetStyle().IsCollapseBorders() {
		padding := t.GetPadding(c)
		hSpacing := t.GetStyle().GetBorderHSpacing(c)
		result += calculatedStyleFloatToInt(padding.Left() + padding.Right() + float32((t.NumEffCols()+1)*hSpacing))
	}
	return result
}

func (t *TableBox) GetColumns() []*ColumnData {
	return t.columns
}

func (t *TableBox) recalcSections(c *LayoutContext) {
	t.EnsureChildren(c)
	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		section.RecalcCells(c)
	}
}

func (t *TableBox) calcBorders(c *LayoutContext) {
	t.EnsureChildren(c)
	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		section.CalcBorders(c)
	}
}

func (t *TableBox) IsAllowHeightToShrink() bool {
	return false
}

func (t *TableBox) Layout(c *LayoutContext) {
	t.CalcMinMaxWidth(c)
	t.CalcDimensions(c)
	t.calcWidth()
	t.calcPageClearance(c)

	// Recalc to pick up auto margins now that layout has been called on
	// containing block and the table has a content width
	if !t.IsAnonymous() {
		t.SetDimensionsCalculated(false)
		t.CalcDimensionsWithCssWidth(c, t.GetContentWidth())
	}

	t.tableLayout.Layout(c)

	t.setCellWidths(c)

	t.layoutTable(c)
}

func (t *TableBox) ResolveAutoMargins(c *LayoutContext, cssWidth int, padding RectPropertySetI,
	border *BorderPropertySet) {
	// If our minimum width is greater than the calculated CSS width,
	// don't try to allocate any margin space to auto margins.  It
	// will just confuse the issue later when we expand the effective
	// table width to its minimum width.
	if t.GetMinWidth() <= t.GetContentWidth()+t.MarginsBordersPaddingAndSpacing(c, true) {
		t.BlockBox.ResolveAutoMargins(c, cssWidth, padding, border)
	} else {
		if t.GetStyle().IsAutoLeftMargin() {
			t.SetMarginLeft(c, 0)
		}
		if t.GetStyle().IsAutoRightMargin() {
			t.SetMarginRight(c, 0)
		}
	}
}

func (t *TableBox) layoutTable(c *LayoutContext) {
	running := c.IsPrint() && t.GetStyle().IsPaginateTable()
	prevExtraTop := 0
	prevExtraBottom := 0

	if running {
		prevExtraTop = c.GetExtraSpaceTop()
		prevExtraBottom = c.GetExtraSpaceBottom()

		c.SetExtraSpaceTop(c.GetExtraSpaceTop() +
			calculatedStyleFloatToInt(t.GetPadding(c).Top()) +
			calculatedStyleFloatToInt(t.GetBorder(c).Top()) +
			t.GetStyle().GetBorderVSpacing(c))
		c.SetExtraSpaceBottom(c.GetExtraSpaceBottom() +
			calculatedStyleFloatToInt(t.GetPadding(c).Bottom()) +
			calculatedStyleFloatToInt(t.GetBorder(c).Bottom()) +
			t.GetStyle().GetBorderVSpacing(c))
	}

	t.BlockBox.Layout(c)

	if running {
		if t.isNeedAnalyzePageBreaks() {
			t.analyzePageBreaks(c)

			t.SetExtraSpaceTop(0)
			t.SetExtraSpaceBottom(0)
		} else {
			t.SetExtraSpaceTop(c.GetExtraSpaceTop() - prevExtraTop)
			t.SetExtraSpaceBottom(c.GetExtraSpaceBottom() - prevExtraBottom)
		}
		c.SetExtraSpaceTop(prevExtraTop)
		c.SetExtraSpaceBottom(prevExtraBottom)
	}
}

func (t *TableBox) LayoutChildren(c *LayoutContext, contentStart int) {
	t.EnsureChildren(c)
	// If we have a running footer, we need its dimensions right away
	running := c.IsPrint() && t.GetStyle().IsPaginateTable()
	if running {
		headerHeight := t.layoutRunningHeader(c)
		footerHeight := t.layoutRunningFooter(c)
		spacingHeight := 0
		if footerHeight != 0 {
			spacingHeight = t.GetStyle().GetBorderVSpacing(c)
		}

		first := c.GetRootLayer().GetFirstPage(c, t)
		if t.GetAbsY()+t.GetTy()+headerHeight+footerHeight+spacingHeight > first.GetBottom() {
			// XXX Performance problem here.  This forces the table
			// to move to the next page (which we want), but the initial
			// table layout run still completes (which we don't)
			t.SetNeedPageClear(true)
		}
	}
	t.BlockBox.LayoutChildren(c, contentStart)
}

func (t *TableBox) layoutRunningHeader(c *LayoutContext) int {
	result := 0
	if t.GetChildCount() > 0 {
		section := t.GetChild(0).(*TableSectionBox)
		if section.IsHeader() {
			c.SetNoPageBreak(c.GetNoPageBreak() + 1)

			section.InitContainingLayer(c)
			section.Layout(c)

			c.SetExtraSpaceTop(c.GetExtraSpaceTop() + section.GetHeight())

			result = section.GetHeight()

			section.Reset(c)

			c.SetNoPageBreak(c.GetNoPageBreak() - 1)
		}
	}

	return result
}

func (t *TableBox) layoutRunningFooter(c *LayoutContext) int {
	result := 0
	if t.GetChildCount() > 0 {
		section := t.GetChild(t.GetChildCount() - 1).(*TableSectionBox)
		if section.IsFooter() {
			c.SetNoPageBreak(c.GetNoPageBreak() + 1)

			section.InitContainingLayer(c)
			section.Layout(c)

			c.SetExtraSpaceBottom(c.GetExtraSpaceBottom() +
				section.GetHeight() +
				t.GetStyle().GetBorderVSpacing(c))

			result = section.GetHeight()

			section.Reset(c)

			c.SetNoPageBreak(c.GetNoPageBreak() - 1)
		}
	}
	return result
}

func (t *TableBox) isNeedAnalyzePageBreaks() bool {
	b := t.GetParent()
	for b != nil {
		if b.GetStyle().IsTable() && b.GetStyle().IsPaginateTable() {
			return false
		}

		b = b.GetParent()
	}

	return true
}

func (t *TableBox) analyzePageBreaks(c *LayoutContext) {
	t.AnalyzePageBreaks(c, nil)
}

func (t *TableBox) AnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer) {
	t.contentLimitContainer = t.BuildContainerAndAnalyzePageBreaks(c, container)

	if container != nil && t.contentLimitContainer.IsContainsMultiplePages() &&
		(t.GetExtraSpaceTop() > 0 || t.GetExtraSpaceBottom() > 0) {
		t.PropagateExtraSpace(c, container, t.contentLimitContainer, t.GetExtraSpaceTop(), t.GetExtraSpaceBottom())
	}
}

func (t *TableBox) PaintBackground(c *RenderingContext) {
	if t.contentLimitContainer == nil {
		t.BlockBox.PaintBackground(c)
	} else if t.GetStyle().IsVisible() {
		c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(
			c, t.GetStyle(), t.getContentLimitedBorderEdge(c), t.GetPaintingBorderEdge(c),
			t.GetStyle().GetBorder(c))
	}
}

func (t *TableBox) PaintBorder(c *RenderingContext) {
	if t.contentLimitContainer == nil {
		t.BlockBox.PaintBorder(c)
	} else if t.GetStyle().IsVisible() {
		c.GetOutputDevice().PaintBorderWithStyleEdgeSides(c, t.GetStyle(), t.getContentLimitedBorderEdge(c), t.GetBorderSides())
	}
}

func (t *TableBox) getContentLimitedBorderEdge(c *RenderingContext) *geom.Rectangle {
	result := t.GetPaintingBorderEdge(c)

	limit := t.contentLimitContainer.GetContentLimit(c.GetPageNo())

	if limit == nil {
		XRLogLayout(LevelWarning, "No content limit found")
	} else {
		if limit.GetTop() == ContentLimitUndefined ||
			limit.GetBottom() == ContentLimitUndefined {
			return result
		}

		padding := t.GetPadding(c)
		border := t.GetBorder(c)

		var top int
		if c.GetPageNo() == t.contentLimitContainer.GetInitialPageNo() {
			top = result.Y
		} else {
			top = limit.GetTop() - calculatedStyleFloatToInt(padding.Top()) -
				calculatedStyleFloatToInt(border.Top()) - t.GetStyle().GetBorderVSpacing(c)
			if t.GetChildCount() > 0 {
				section := t.GetChild(0).(*TableSectionBox)
				if section.IsHeader() {
					top -= section.GetHeight()
				}
			}
		}

		var bottom int
		if c.GetPageNo() == t.contentLimitContainer.GetLastPageNo() {
			bottom = result.Y + result.Height
		} else {
			bottom = limit.GetBottom() + calculatedStyleFloatToInt(padding.Bottom()) +
				calculatedStyleFloatToInt(border.Bottom()) + t.GetStyle().GetBorderVSpacing(c)
			if t.GetChildCount() > 0 {
				section := t.GetChild(t.GetChildCount() - 1).(*TableSectionBox)
				if section.IsFooter() {
					bottom += section.GetHeight()
				}
			}
		}

		result.Y = top
		result.Height = bottom - top

	}
	return result
}

func (t *TableBox) UpdateHeaderFooterPosition(c *RenderingContext) {
	limit := t.contentLimitContainer.GetContentLimit(c.GetPageNo())

	if limit != nil {
		t.updateHeaderPosition(c, limit)
		t.updateFooterPosition(c, limit)
	}
}

func (t *TableBox) updateHeaderPosition(c *RenderingContext, limit *ContentLimit) {
	if limit.GetTop() != ContentLimitUndefined ||
		c.GetPageNo() == t.contentLimitContainer.GetInitialPageNo() {
		if t.GetChildCount() > 0 {
			section := t.GetChild(0).(*TableSectionBox)
			if section.IsHeader() {
				if !section.IsCapturedOriginalAbsY() {
					section.SetOriginalAbsY(section.GetAbsY())
					section.SetCapturedOriginalAbsY(true)
				}

				var newAbsY int
				if c.GetPageNo() == t.contentLimitContainer.GetInitialPageNo() {
					newAbsY = section.GetOriginalAbsY()
				} else {
					newAbsY = limit.GetTop() -
						t.GetStyle().GetBorderVSpacing(c) -
						section.GetHeight()
				}

				diff := newAbsY - section.GetAbsY()

				if diff != 0 {
					section.SetY(section.GetY() + diff)
					section.CalcCanvasLocation()
					section.CalcChildLocations()
					section.CalcPaintingInfo(c, false)
				}
			}
		}
	}
}

func (t *TableBox) updateFooterPosition(c *RenderingContext, limit *ContentLimit) {
	if limit.GetBottom() != ContentLimitUndefined ||
		c.GetPageNo() == t.contentLimitContainer.GetLastPageNo() {
		if t.GetChildCount() > 0 {
			section := t.GetChild(t.GetChildCount() - 1).(*TableSectionBox)
			if section.IsFooter() {
				if !section.IsCapturedOriginalAbsY() {
					section.SetOriginalAbsY(section.GetAbsY())
					section.SetCapturedOriginalAbsY(true)
				}

				var newAbsY int
				if c.GetPageNo() == t.contentLimitContainer.GetLastPageNo() {
					newAbsY = section.GetOriginalAbsY()
				} else {
					newAbsY = limit.GetBottom()
				}

				diff := newAbsY - section.GetAbsY()

				if diff != 0 {
					section.SetY(section.GetY() + diff)
					section.CalcCanvasLocation()
					section.CalcChildLocations()
					section.CalcPaintingInfo(c, false)
				}
			}
		}
	}
}

func (t *TableBox) calcPageClearance(c *LayoutContext) {
	if c.IsPrint() && t.GetStyle().IsCollapseBorders() {
		page := c.GetRootLayer().GetFirstPage(c, t)
		if page != nil {
			row := t.GetFirstRow()
			if row != nil {
				spill := 0
				for _, box := range row.GetChildren() {
					cell := box.(*TableCellBox)
					collapsed := cell.GetCollapsedPaintingBorder()
					tmp := calculatedStyleFloatToInt(collapsed.Top()) / 2
					if tmp > spill {
						spill = tmp
					}
				}

				borderTop := t.GetAbsY() + calculatedStyleFloatToInt(t.GetMargin(c).Top()) - spill
				delta := page.GetTop() - borderTop
				if delta > 0 {
					t.SetY(t.GetY() + delta)
					t.SetPageClearance(delta)
					t.CalcCanvasLocation()
					c.Translate(0, delta)
				}
			}
		}
	}
}

func (t *TableBox) calcWidth() {
	if t.GetMinWidth() > t.GetWidth() {
		t.SetContentWidth(t.GetContentWidth() + t.GetMinWidth() - t.GetWidth())
	} else if t.GetStyle().IsIdent(CSSNameWidth, IdentValueAuto) &&
		t.GetMaxWidth() < t.GetWidth() {
		t.SetContentWidth(t.GetContentWidth() - (t.GetWidth() - t.GetMaxWidth()))
	}
}

// GetFirstRow may return nil.
func (t *TableBox) GetFirstRow() *TableRowBox {
	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		if section.GetChildCount() > 0 {
			return section.GetChild(0).(*TableRowBox)
		}
	}

	return nil
}

// GetFirstBodyRow may return nil.
func (t *TableBox) GetFirstBodyRow() *TableRowBox {
	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		if section.IsHeader() || section.IsFooter() {
			continue
		}
		if section.GetChildCount() > 0 {
			return section.GetChild(0).(*TableRowBox)
		}
	}

	return nil
}

func (t *TableBox) setCellWidths(c *LayoutContext) {
	for _, value := range t.GetChildren() {
		box := value.(BlockBoxI)
		if box.GetStyle().IsTableSection() {
			box.(*TableSectionBox).SetCellWidths(c)
		}
	}
}

func (t *TableBox) CalcLayoutHeight(c *LayoutContext, border *BorderPropertySet,
	margin RectPropertySetI, padding RectPropertySetI) {
	t.BlockBox.CalcLayoutHeight(c, border, margin, padding)

	if t.GetChildCount() > 0 {
		t.SetHeight(t.GetHeight() + t.GetStyle().GetBorderVSpacing(c))
	}
}

func (t *TableBox) Reset(c *LayoutContext) {
	t.BlockBox.Reset(c)

	t.contentLimitContainer = nil

	t.tableLayout.Reset()
}

func (t *TableBox) GetCSSWidth(c CssContext) int {
	if t.GetStyle().IsAutoWidth() {
		return -1
	} else {
		// XHTML 1.0 specifies that a table width refers to the border
		// width.  This can be removed if/when we support the box-sizing
		// property.
		result := calculatedStyleFloatToInt(t.GetStyle().GetFloatPropertyProportionalWidth(
			CSSNameWidth, float32(t.GetContainingBlock().GetContentWidth()), c))

		border := t.GetBorder(c)
		result -= calculatedStyleFloatToInt(border.Left()) + calculatedStyleFloatToInt(border.Right())
		if !t.GetStyle().IsCollapseBorders() {
			padding := t.GetPadding(c)
			result -= calculatedStyleFloatToInt(padding.Left()) + calculatedStyleFloatToInt(padding.Right())
		}

		if result >= 0 {
			return result
		}
		return -1
	}
}

// ColElement may return nil.
func (t *TableBox) ColElement(col int) *TableColumn {
	styleColumns := t.GetStyleColumns()
	if len(styleColumns) == 0 {
		return nil
	}
	cCol := 0
	for _, colElem := range styleColumns {
		span := colElem.GetStyle().GetColSpan()
		cCol += span
		if cCol > col {
			return colElem
		}
	}
	return nil
}

func (t *TableBox) GetColumnBounds(c CssContext, col int) *geom.Rectangle {
	effCol := t.ColToEffCol(col)

	hspacing := t.GetStyle().GetBorderHSpacing(c)
	vspacing := t.GetStyle().GetBorderVSpacing(c)

	result := t.GetContentAreaEdge(t.GetAbsX(), t.GetAbsY(), c)

	result.Y += vspacing
	result.Height -= vspacing * 2

	result.X += t.columnPos[effCol] + hspacing

	return result
}

func (t *TableBox) GetBorder(cssCtx CssContext) *BorderPropertySet {
	if t.GetStyle().IsCollapseBorders() {
		return BorderPropertySetEmptyBorder
	} else {
		return t.BlockBox.GetBorder(cssCtx)
	}
}

func (t *TableBox) CalcFixedHeightRowBottom(c CssContext) int {
	if !t.IsAnonymous() {
		cssHeight := t.GetCSSHeight(c)
		if cssHeight != -1 {
			return t.GetAbsY() + cssHeight -
				calculatedStyleFloatToInt(t.GetBorder(c).Bottom()) - calculatedStyleFloatToInt(t.GetPadding(c).Bottom()) -
				t.GetStyle().GetBorderVSpacing(c)
		}
	}

	return -1
}

func (t *TableBox) IsMayCollapseMarginsWithChildren() bool {
	return false
}

// tableBoxSectionOrNil is Java's (TableSectionBox) cast of a Box that may be
// null.
func tableBoxSectionOrNil(b BoxI) *TableSectionBox {
	if b == nil {
		return nil
	}
	return b.(*TableSectionBox)
}

// SectionAbove may return nil.
func (t *TableBox) SectionAbove(
	section *TableSectionBox, skipEmptySections bool) *TableSectionBox {
	prevSection := tableBoxSectionOrNil(section.GetPreviousSibling())

	if prevSection == nil {
		return nil
	}

	for prevSection != nil {
		if prevSection.NumRows() > 0 || !skipEmptySections {
			break
		}
		prevSection = tableBoxSectionOrNil(prevSection.GetPreviousSibling())
	}

	return prevSection
}

// SectionBelow may return nil.
func (t *TableBox) SectionBelow(
	section *TableSectionBox, skipEmptySections bool) *TableSectionBox {
	nextSection := tableBoxSectionOrNil(section.GetNextSibling())

	if nextSection == nil {
		return nil
	}

	for nextSection != nil {
		if nextSection.NumRows() > 0 || !skipEmptySections {
			break
		}
		nextSection = tableBoxSectionOrNil(nextSection.GetNextSibling())
	}

	return nextSection
}

// CellAbove may return nil.
func (t *TableBox) CellAbove(cell *TableCellBox) *TableCellBox {
	// Find the section and row to look in
	r := cell.GetRow()
	var section *TableSectionBox
	rAbove := 0
	if r > 0 {
		// cell is not in the first row, so use the above row in its own
		// section
		section = cell.GetSection()
		rAbove = r - 1
	} else {
		section = t.SectionAbove(cell.GetSection(), true)
		if section != nil {
			rAbove = section.NumRows() - 1
		}
	}

	// Look up the cell in the section's grid, which requires effective col
	// index
	return t.getTableCellBox(cell, section, rAbove)
}

// CellBelow may return nil.
func (t *TableBox) CellBelow(cell *TableCellBox) *TableCellBox {
	// Find the section and row to look in
	r := cell.GetRow() + cell.GetStyle().GetRowSpan() - 1
	var section *TableSectionBox
	rBelow := 0
	if r < cell.GetSection().NumRows()-1 {
		// The cell is not in the last row, so use the next row in the
		// section.
		section = cell.GetSection()
		rBelow = r + 1
	} else {
		section = t.SectionBelow(cell.GetSection(), true)
		if section != nil {
			rBelow = 0
		}
	}

	// Look up the cell in the section's grid, which requires effective col
	// index
	return t.getTableCellBox(cell, section, rBelow)
}

func (t *TableBox) getTableCellBox(cell *TableCellBox, section *TableSectionBox, rBelow int) *TableCellBox {
	if section == nil {
		return nil
	}

	effCol := t.ColToEffCol(cell.GetCol())
	var belowCell *TableCellBox
	// If we hit a colspan back up to a real cell.
	for {
		belowCell = section.CellAt(rBelow, effCol)
		effCol--
		if !(belowCell == TableCellBoxSpanningCell && effCol >= 0) {
			break
		}
	}
	if belowCell == TableCellBoxSpanningCell {
		return nil
	}
	return belowCell
}

// CellLeft may return nil.
func (t *TableBox) CellLeft(cell *TableCellBox) *TableCellBox {
	section := cell.GetSection()
	effCol := t.ColToEffCol(cell.GetCol())
	if effCol == 0 {
		return nil
	}

	// If we hit a colspan back up to a real cell.
	var prevCell *TableCellBox
	for {
		prevCell = section.CellAt(cell.GetRow(), effCol-1)
		effCol--
		if !(prevCell == TableCellBoxSpanningCell && effCol >= 0) {
			break
		}
	}
	if prevCell == TableCellBoxSpanningCell {
		return nil
	}
	return prevCell
}

// CellRight may return nil.
func (t *TableBox) CellRight(cell *TableCellBox) *TableCellBox {
	effCol := t.ColToEffCol(cell.GetCol() + cell.GetStyle().GetColSpan())
	if effCol >= t.NumEffCols() {
		return nil
	}
	result := cell.GetSection().CellAt(cell.GetRow(), effCol)
	if result == TableCellBoxSpanningCell {
		return nil
	}
	return result
}

func (t *TableBox) CalcInlineBaseline(c CssContext) int {
	result := 0
	found := false
OUTER:
	for _, box := range t.GetChildren() {
		section := box.(*TableSectionBox)
		for _, value := range section.GetChildren() {
			row := value.(*TableRowBox)
			found = true
			result = row.GetAbsY() + row.GetBaseline() - t.GetAbsY()
			break OUTER
		}
	}

	if !found {
		result = t.GetHeight()
	}

	return result
}

func (t *TableBox) GetPageClearance() int {
	return t.pageClearance
}

func (t *TableBox) SetPageClearance(pageClearance int) {
	t.pageClearance = pageClearance
}

func (t *TableBox) HasContentLimitContainer() bool {
	return t.contentLimitContainer != nil
}

func (t *TableBox) GetExtraSpaceTop() int {
	return t.extraSpaceTop
}

func (t *TableBox) SetExtraSpaceTop(extraSpaceTop int) {
	t.extraSpaceTop = extraSpaceTop
}

func (t *TableBox) GetExtraSpaceBottom() int {
	return t.extraSpaceBottom
}

func (t *TableBox) SetExtraSpaceBottom(extraSpaceBottom int) {
	t.extraSpaceBottom = extraSpaceBottom
}

// TableBoxTableLayout ports the private interface TableBox.TableLayout.
type TableBoxTableLayout interface {
	CalcMinMaxWidth(c *LayoutContext)
	Layout(c *LayoutContext)
	Reset()
}

// TableBoxMarginTableLayout is a specialization of TableBoxAutoTableLayout used for laying out the
// tables used to approximate the margin box layout algorithm from CSS3
// GCPM.
type TableBoxMarginTableLayout struct {
	TableBoxAutoTableLayout
}

func newTableBoxMarginTableLayout(table *TableBox) *TableBoxMarginTableLayout {
	m := &TableBoxMarginTableLayout{}
	m.table = table
	m.self = m
	return m
}

func (m *TableBoxMarginTableLayout) GetMinColWidth() int {
	return 0
}

func (m *TableBoxMarginTableLayout) CalcMinMaxWidth(c *LayoutContext) {
	m.TableBoxAutoTableLayout.CalcMinMaxWidth(c)

	layoutStruct := m.GetLayoutStruct()

	if len(layoutStruct) == 3 {
		center := layoutStruct[1]

		if !(center.Width().IsVariable() && center.MaxWidth() == 0) {
			if layoutStruct[0].MinWidth() > layoutStruct[2].MinWidth() {
				layoutStruct[2] = layoutStruct[0]
			} else if layoutStruct[2].MinWidth() > layoutStruct[0].MinWidth() {
				layoutStruct[0] = layoutStruct[2]
			} else {
				l := NewTableBoxAutoTableLayoutLayout()
				l.SetMinWidth(max(layoutStruct[0].MinWidth(), layoutStruct[2].MinWidth()))
				l.SetEffMinWidth(l.MinWidth())
				l.SetMaxWidth(max(layoutStruct[0].MaxWidth(), layoutStruct[2].MaxWidth()))
				l.SetEffMaxWidth(l.MaxWidth())

				layoutStruct[0] = l
				layoutStruct[2] = l
			}
		}
	}
}

// TableBoxFixedTableLayout ports the private class TableBox.FixedTableLayout.
//
// In the two layout strategies a Java int local is an int32 and a Java long an
// int64, so that the narrowing (int) casts and 32-bit overflow of the Java
// arithmetic give the same column widths.
type TableBoxFixedTableLayout struct {
	table  *TableBox
	widths []*Length
}

func newTableBoxFixedTableLayout(table *TableBox) *TableBoxFixedTableLayout {
	return &TableBoxFixedTableLayout{table: table}
}

func (f *TableBoxFixedTableLayout) Reset() {
	f.widths = nil
}

func (f *TableBoxFixedTableLayout) initWidths() {
	f.widths = make([]*Length, 0, f.table.NumEffCols())
	for i := 0; i < f.table.NumEffCols(); i++ {
		f.widths = append(f.widths, LengthZero)
	}
}

func (f *TableBoxFixedTableLayout) calcWidthArray(c *LayoutContext) int {
	f.initWidths()

	table := f.table

	cCol := 0
	nEffCols := table.NumEffCols()
	var usedWidth int32 = 0

	for _, col := range table.GetStyleColumns() {
		span := col.GetStyle().GetColSpan()
		w := col.GetStyle().AsLength(c, CSSNameWidth)
		if w.IsVariable() && col.GetParent() != nil {
			w = col.GetParent().GetStyle().AsLength(c, CSSNameWidth)
		}

		var effWidth int64 = 0
		if w.IsFixed() && w.Value() > 0 {
			effWidth = w.Value()
			effWidth = min(effWidth, int64(LengthMaxWidth))
		}

		usedSpan := 0
		i := 0
		for usedSpan < span {
			if cCol+i >= nEffCols {
				table.AppendColumn(span - usedSpan)
				nEffCols++
				f.widths = append(f.widths, LengthZero)
			}
			eSpan := table.SpanOfEffCol(cCol + i)
			if (w.IsFixed() || w.IsPercent()) && w.Value() > 0 {
				f.widths[cCol+i] = NewLength(w.Value()*int64(eSpan), w.Type())
				usedWidth += int32(effWidth * int64(eSpan))
			}
			usedSpan += eSpan
			i++
		}
		cCol += i
	}

	cCol = 0
	firstRow := f.table.GetFirstRow()
	if firstRow != nil {
		for _, box := range firstRow.GetChildren() {
			cell := box.(*TableCellBox)
			w := cell.GetOuterStyleWidth(c)
			span := cell.GetStyle().GetColSpan()
			var effWidth int64 = 0
			if w.IsFixed() && w.Value() > 0 {
				effWidth = w.Value()
			}

			usedSpan := 0
			i := 0
			for usedSpan < span {
				eSpan := f.table.SpanOfEffCol(cCol + i)

				columnWidth := f.widths[cCol+i]
				// only set if no col element has already set it.
				if columnWidth.IsVariable() && !w.IsVariable() {
					f.widths[cCol+i] = NewLength(w.Value()*int64(eSpan), w.Type())
					usedWidth += int32(effWidth * int64(eSpan))
				}

				usedSpan += eSpan
				i++
			}

			cCol += i
		}
	}

	return int(usedWidth)
}

func (f *TableBoxFixedTableLayout) CalcMinMaxWidth(c *LayoutContext) {
	bs := f.table.MarginsBordersPaddingAndSpacing(c, true)

	f.table.CalcDimensions(c)

	// Reset to allow layout to have another crack at this.  If we're
	// participating in a nested max/min-width calculation, the values
	// calculated above may be wrong and may need updating once our
	// parent has a width.
	f.table.SetDimensionsCalculated(false)

	mw := int(int32(f.calcWidthArray(c)) + int32(bs))
	f.table.SetMinWidth(max(mw, f.table.GetWidth()))
	f.table.SetMaxWidth(f.table.GetMinWidth())

	haveNonFixed := false
	for _, w := range f.widths {
		if !w.IsFixed() {
			haveNonFixed = true
			break
		}
	}

	if haveNonFixed {
		f.table.SetMaxWidth(LengthMaxWidth)
	}
}

func (f *TableBoxFixedTableLayout) Layout(c *LayoutContext) {
	tableWidth := int32(f.table.GetWidth()) - int32(f.table.MarginsBordersPaddingAndSpacing(c, false))
	available := tableWidth
	nEffCols := f.table.NumEffCols()

	calcWidth := make([]int64, nEffCols)
	for i := range calcWidth {
		calcWidth[i] = -1
	}

	// first assign fixed width
	for i := 0; i < nEffCols; i++ {
		l := f.widths[i]
		if l.IsFixed() {
			calcWidth[i] = l.Value()
			available -= int32(l.Value())
		}
	}

	// assign percent width
	if available > 0 {
		var totalPercent int32 = 0
		for i := 0; i < nEffCols; i++ {
			l := f.widths[i]
			if l.IsPercent() {
				totalPercent += int32(l.Value())
			}
		}

		// calculate how much to distribute to percent cells.
		base := tableWidth * totalPercent / 100
		if base > available {
			base = available
		}

		for i := 0; available > 0 && i < nEffCols; i++ {
			l := f.widths[i]
			if l.IsPercent() {
				w := int64(base) * l.Value() / int64(totalPercent)
				available -= int32(w)
				calcWidth[i] = w
			}
		}
	}

	// assign variable width
	if available > 0 {
		var totalVariable int32 = 0
		for i := 0; i < nEffCols; i++ {
			l := f.widths[i]
			if l.IsVariable() {
				totalVariable++
			}
		}

		for i := 0; available > 0 && i < nEffCols; i++ {
			l := f.widths[i]
			if l.IsVariable() {
				w := available / totalVariable
				available -= w
				calcWidth[i] = int64(w)
				totalVariable--
			}
		}
	}

	for i := 0; i < nEffCols; i++ {
		if calcWidth[i] < 0 {
			calcWidth[i] = 0 // IE gives min 1 px...
		}
	}

	// spread extra space over columns
	if available > 0 {
		total := int32(nEffCols)
		// still have some width to spread
		i := nEffCols
		for i > 0 {
			i--
			w := available / total
			available -= w
			total--
			calcWidth[i] += int64(w)
		}
	}

	var pos int32 = 0
	hspacing := int32(f.table.GetStyle().GetBorderHSpacing(c))
	columnPos := make([]int, nEffCols+1)
	for i := 0; i < nEffCols; i++ {
		columnPos[i] = int(pos)
		pos += int32(calcWidth[i] + int64(hspacing))
	}

	columnPos[len(columnPos)-1] = int(pos)

	f.table.setColumnPos(columnPos)
}

// tableBoxAutoTableLayoutSelf is the part of TableBoxAutoTableLayout that
// TableBoxMarginTableLayout overrides and the base class calls.
type tableBoxAutoTableLayoutSelf interface {
	GetMinColWidth() int
}

// TableBoxAutoTableLayout ports the private class TableBox.AutoTableLayout.
type TableBoxAutoTableLayout struct {
	table        *TableBox
	layoutStruct []*TableBoxAutoTableLayoutLayout
	spanCells    []*TableCellBox

	self tableBoxAutoTableLayoutSelf
}

func newTableBoxAutoTableLayout(table *TableBox) *TableBoxAutoTableLayout {
	a := &TableBoxAutoTableLayout{table: table}
	a.self = a
	return a
}

func (a *TableBoxAutoTableLayout) Reset() {
	a.layoutStruct = nil
	a.spanCells = nil
}

// GetLayoutStruct may return nil.
func (a *TableBoxAutoTableLayout) GetLayoutStruct() []*TableBoxAutoTableLayoutLayout {
	return a.layoutStruct
}

func (a *TableBoxAutoTableLayout) fullRecalc(c *LayoutContext) {
	a.layoutStruct = make([]*TableBoxAutoTableLayoutLayout, a.table.NumEffCols())
	for i := 0; i < len(a.layoutStruct); i++ {
		a.layoutStruct[i] = NewTableBoxAutoTableLayoutLayout()
		a.layoutStruct[i].SetMinWidth(int64(a.self.GetMinColWidth()))
		a.layoutStruct[i].SetMaxWidth(int64(a.self.GetMinColWidth()))
	}

	a.spanCells = []*TableCellBox{}

	table := a.table
	nEffCols := table.NumEffCols()

	cCol := 0
	for _, col := range table.GetStyleColumns() {
		span := col.GetStyle().GetColSpan()
		w := col.GetStyle().AsLength(c, CSSNameWidth)
		if w.IsVariable() && col.GetParent() != nil {
			w = col.GetParent().GetStyle().AsLength(c, CSSNameWidth)
		}

		if w.IsFixed() && w.Value() == 0 || w.IsPercent() && w.Value() == 0 {
			w = LengthZero
		}
		cEffCol := table.ColToEffCol(cCol)
		if !w.IsVariable() && span == 1 && cEffCol < nEffCols {
			if table.SpanOfEffCol(cEffCol) == 1 {
				a.layoutStruct[cEffCol].SetWidth(w)
				if w.IsFixed() && a.layoutStruct[cEffCol].MaxWidth() < w.Value() {
					a.layoutStruct[cEffCol].SetMaxWidth(w.Value())
				}
			}
		}
		cCol += span

	}

	for i := 0; i < nEffCols; i++ {
		a.recalcColumn(c, i)
	}
}

func (a *TableBoxAutoTableLayout) GetMinColWidth() int {
	return 1
}

func (a *TableBoxAutoTableLayout) recalcColumn(c *LayoutContext, effCol int) {
	l := a.layoutStruct[effCol]

	// first we iterate over all rows.
	for _, box := range a.table.GetChildren() {
		section := box.(*TableSectionBox)
		numRows := section.NumRows()
		for i := 0; i < numRows; i++ {
			cell := section.CellAt(i, effCol)
			if cell == TableCellBoxSpanningCell || cell == nil {
				continue
			}
			if cell.GetStyle().GetColSpan() == 1 {
				// A cell originates in this column. Ensure we have
				// a min/max width of at least 1px for this column now.
				l.SetMinWidth(max(l.MinWidth(), int64(a.self.GetMinColWidth())))
				l.SetMaxWidth(max(l.MaxWidth(), int64(a.self.GetMinColWidth())))

				cell.CalcMinMaxWidth(c)
				if int64(cell.GetMinWidth()) > l.MinWidth() {
					l.SetMinWidth(int64(cell.GetMinWidth()))
				}
				if int64(cell.GetMaxWidth()) > l.MaxWidth() {
					l.SetMaxWidth(int64(cell.GetMaxWidth()))
				}

				outerLength := cell.GetOuterStyleOrColWidth(c)
				w := NewLength(min(int64(LengthMaxWidth), max(0, outerLength.Value())), outerLength.Type())

				switch w.Type() {
				case LengthLengthTypeFixed:
					if w.Value() > 0 && !l.Width().IsPercent() {
						if l.Width().IsFixed() {
							if w.Value() > l.Width().Value() {
								l.SetWidth(w)
							}
						} else {
							l.SetWidth(w)
						}
						if w.Value() > l.MaxWidth() {
							l.SetMaxWidth(w.Value())
						}
					}
				case LengthLengthTypePercent:
					if w.Value() > 0 &&
						(!l.Width().IsPercent() || w.Value() > l.Width().Value()) {
						l.SetWidth(w)
					}
				}
			} else {
				if effCol == 0 || section.CellAt(i, effCol-1) != cell {
					// This spanning cell originates in this column.
					// Ensure we have a min/max width of at least 1px for this column now.
					l.SetMinWidth(max(l.MinWidth(), int64(a.self.GetMinColWidth())))
					l.SetMaxWidth(max(l.MaxWidth(), int64(a.self.GetMinColWidth())))

					a.spanCells = append(a.spanCells, cell)
				}
			}
		}
	}

	l.SetMaxWidth(max(l.MaxWidth(), l.MinWidth()))
}

// calcEffectiveWidth takes care of colspans. effWidth is the same as width for
// cells without colspans. If we have colspans, they get modified.
func (a *TableBoxAutoTableLayout) calcEffectiveWidth(c *LayoutContext) int64 {
	var tMaxWidth int64 = 0

	layoutStruct := a.layoutStruct

	nEffCols := len(layoutStruct)
	hspacing := int32(a.table.GetStyle().GetBorderHSpacing(c))

	for _, layout := range layoutStruct {
		layout.SetEffWidth(layout.Width())
		layout.SetEffMinWidth(layout.MinWidth())
		layout.SetEffMaxWidth(layout.MaxWidth())
	}

	// List.sort is a stable sort.
	sort.SliceStable(a.spanCells, func(i, j int) bool {
		return a.spanCells[i].GetStyle().GetColSpan() < a.spanCells[j].GetStyle().GetColSpan()
	})

	for _, cell := range a.spanCells {
		cell.CalcMinMaxWidth(c)

		span := cell.GetStyle().GetColSpan()
		w := cell.GetOuterStyleOrColWidth(c)
		if w.Value() == 0 {
			w = LengthZero // make it Variable
		}

		col := a.table.ColToEffCol(cell.GetCol())
		lastCol := col
		cMinWidth := int32(cell.GetMinWidth()) + hspacing
		cMaxWidth := int32(cell.GetMaxWidth()) + hspacing
		var totalPercent int32 = 0
		var minWidth int32 = 0
		var maxWidth int32 = 0
		allColsArePercent := true
		allColsAreFixed := true
		haveVariable := false
		var fixedWidth int32 = 0

		for lastCol < nEffCols && span > 0 {
			switch layoutStruct[lastCol].Width().Type() {
			case LengthLengthTypePercent:
				totalPercent += int32(layoutStruct[lastCol].Width().Value())
				allColsAreFixed = false
			case LengthLengthTypeFixed:
				if layoutStruct[lastCol].Width().Value() > 0 {
					fixedWidth += int32(layoutStruct[lastCol].Width().Value())
					allColsArePercent = false
					break
				}
				fallthrough
			case LengthLengthTypeVariable:
				haveVariable = true
				fallthrough
			default:
				// If the column is a percentage width, do not let the spanning cell overwrite the
				// width value.  This caused a mis-rendering on amazon.com.
				// Sample snippet:
				// <table border=2 width=100%><
				//   <tr><td>1</td><td colspan=2>2-3</tr>
				//   <tr><td>1</td><td colspan=2 width=100%>2-3</td></tr>
				// </table>
				if !layoutStruct[lastCol].EffWidth().IsPercent() {
					layoutStruct[lastCol].SetEffWidth(LengthZero)
					allColsArePercent = false
				} else {
					totalPercent += int32(layoutStruct[lastCol].EffWidth().Value())
				}
				allColsAreFixed = false
			}

			span -= a.table.SpanOfEffCol(lastCol)
			minWidth += int32(layoutStruct[lastCol].EffMinWidth())
			maxWidth += int32(layoutStruct[lastCol].EffMaxWidth())
			lastCol++
			cMinWidth -= hspacing
			cMaxWidth -= hspacing
		}

		// adjust table max width if needed
		if w.IsPercent() {
			if int64(totalPercent) > w.Value() || allColsArePercent {
				// can't satisfy this condition, treat as variable
				w = LengthZero
			} else {
				spanMax := max(maxWidth, cMaxWidth)
				tMaxWidth = max(tMaxWidth, int64(spanMax)*100/w.Value())

				// all non-percent columns in the span get percent
				// values to sum up correctly.
				percentMissing := w.Value() - int64(totalPercent)
				var totalWidth int32 = 0
				for pos := col; pos < lastCol; pos++ {
					if !layoutStruct[pos].Width().IsPercent() {
						totalWidth += int32(layoutStruct[pos].EffMaxWidth())
					}
				}

				for pos := col; pos < lastCol && totalWidth > 0; pos++ {
					if !layoutStruct[pos].Width().IsPercent() {
						percent := percentMissing * layoutStruct[pos].EffMaxWidth() /
							int64(totalWidth)
						totalWidth -= int32(layoutStruct[pos].EffMaxWidth())
						percentMissing -= percent
						if percent > 0 {
							layoutStruct[pos].SetEffWidth(NewLength(percent, LengthLengthTypePercent))
						} else {
							layoutStruct[pos].SetEffWidth(LengthZero)
						}
					}
				}
			}
		}

		// make sure minWidth and maxWidth of the spanning cell are honoured
		if cMinWidth > minWidth {
			if allColsAreFixed {
				for pos := col; fixedWidth > 0 && pos < lastCol; pos++ {
					cWidth := max(layoutStruct[pos].EffMinWidth(), int64(cMinWidth)*
						layoutStruct[pos].Width().Value()/int64(fixedWidth))
					fixedWidth -= int32(layoutStruct[pos].Width().Value())
					cMinWidth -= int32(cWidth)
					layoutStruct[pos].SetEffMinWidth(cWidth)
				}
			} else if allColsArePercent {
				maxw := maxWidth
				minw := minWidth
				cminw := cMinWidth

				for pos := col; maxw > 0 && pos < lastCol; pos++ {
					if layoutStruct[pos].EffWidth().IsPercent() &&
						layoutStruct[pos].EffWidth().Value() > 0 &&
						fixedWidth <= cMinWidth {
						cWidth := layoutStruct[pos].EffMinWidth()
						cWidth = max(cWidth, int64(cminw)*
							layoutStruct[pos].EffWidth().Value()/int64(totalPercent))
						cWidth = min(layoutStruct[pos].EffMinWidth()+
							int64(cMinWidth-minw), cWidth)
						maxw -= int32(layoutStruct[pos].EffMaxWidth())
						minw -= int32(layoutStruct[pos].EffMinWidth())
						cMinWidth -= int32(cWidth)
						layoutStruct[pos].SetEffMinWidth(cWidth)
					}
				}
			} else {
				maxw := maxWidth
				minw := minWidth

				// Give min to variable first, to fixed second, and to
				// others third.
				for pos := col; maxw > 0 && pos < lastCol; pos++ {
					if layoutStruct[pos].Width().IsFixed() && haveVariable &&
						fixedWidth <= cMinWidth {
						cWidth := max(layoutStruct[pos].EffMinWidth(),
							layoutStruct[pos].Width().Value())
						fixedWidth -= int32(layoutStruct[pos].Width().Value())
						minw -= int32(layoutStruct[pos].EffMinWidth())
						maxw -= int32(layoutStruct[pos].EffMaxWidth())
						cMinWidth -= int32(cWidth)
						layoutStruct[pos].SetEffMinWidth(cWidth)
					}
				}

				for pos := col; maxw > 0 && pos < lastCol && minw < cMinWidth; pos++ {
					if !(layoutStruct[pos].Width().IsFixed() && haveVariable && fixedWidth <= cMinWidth) {
						cWidth := max(layoutStruct[pos].EffMinWidth(), int64(cMinWidth)*
							layoutStruct[pos].EffMaxWidth()/int64(maxw))
						cWidth = min(layoutStruct[pos].EffMinWidth()+
							int64(cMinWidth-minw), cWidth)

						maxw -= int32(layoutStruct[pos].EffMaxWidth())
						minw -= int32(layoutStruct[pos].EffMinWidth())
						cMinWidth -= int32(cWidth)
						layoutStruct[pos].SetEffMinWidth(cWidth)
					}
				}
			}
		}

		if !w.IsPercent() {
			if cMaxWidth > maxWidth {
				for pos := col; maxWidth > 0 && pos < lastCol; pos++ {
					cWidth := max(layoutStruct[pos].EffMaxWidth(), int64(cMaxWidth)*
						layoutStruct[pos].EffMaxWidth()/int64(maxWidth))
					maxWidth -= int32(layoutStruct[pos].EffMaxWidth())
					cMaxWidth -= int32(cWidth)
					layoutStruct[pos].SetEffMaxWidth(cWidth)
				}
			}
		} else {
			for pos := col; pos < lastCol; pos++ {
				layoutStruct[pos].SetMaxWidth(max(layoutStruct[pos].MaxWidth(),
					layoutStruct[pos].MinWidth()))
			}
		}
	}

	return tMaxWidth
}

func (a *TableBoxAutoTableLayout) CalcMinMaxWidth(c *LayoutContext) {
	table := a.table

	a.fullRecalc(c)

	layoutStruct := a.layoutStruct

	spanMaxWidth := a.calcEffectiveWidth(c)
	var minWidth int64 = 0
	var maxWidth int64 = 0
	var maxPercent int64 = 0
	var maxNonPercent int64 = 0

	var remainingPercent int32 = 100
	for _, layout := range layoutStruct {
		minWidth += layout.EffMinWidth()
		maxWidth += layout.EffMaxWidth()
		if layout.EffWidth().IsPercent() {
			percent := min(layout.EffWidth().Value(), int64(remainingPercent))
			pw := layout.EffMaxWidth() * 100 / max(percent, 1)
			remainingPercent -= int32(percent)
			maxPercent = max(pw, maxPercent)
		} else {
			maxNonPercent += layout.EffMaxWidth()
		}
	}

	maxNonPercent = (maxNonPercent*100 + 50) / int64(max(remainingPercent, 1))
	maxWidth = max(maxNonPercent, maxWidth)
	maxWidth = max(maxWidth, maxPercent)
	maxWidth = max(maxWidth, spanMaxWidth)

	bs := table.MarginsBordersPaddingAndSpacing(c, true)
	minWidth += int64(bs)
	maxWidth += int64(bs)

	tw := table.GetStyle().AsLength(c, CSSNameWidth)
	if tw.IsFixed() && tw.Value() > 0 {
		table.CalcDimensions(c)
		width := int32(table.GetContentWidth()) + int32(table.MarginsBordersPaddingAndSpacing(c, true))
		minWidth = max(minWidth, int64(width))
		maxWidth = minWidth
	}

	table.SetMaxWidth(int(int32(min(maxWidth, int64(LengthMaxWidth)))))
	table.SetMinWidth(int(int32(min(minWidth, int64(LengthMaxWidth)))))
}

func (a *TableBoxAutoTableLayout) Layout(c *LayoutContext) {
	table := a.table
	// table layout based on the values collected in the layout
	// structure.
	tableWidth := int32(table.GetWidth()) - int32(table.MarginsBordersPaddingAndSpacing(c, false))
	available := tableWidth
	nEffCols := table.NumEffCols()

	havePercent := false
	var numVariable int32 = 0
	var numFixed int32 = 0
	var totalVariable int32 = 0
	var totalFixed int32 = 0
	var totalPercent int32 = 0
	var allocVariable int32 = 0

	layoutStruct := a.layoutStruct

	// fill up every cell with it's minWidth
	for i := 0; i < nEffCols; i++ {
		w := layoutStruct[i].EffMinWidth()
		layoutStruct[i].SetCalcWidth(w)
		available -= int32(w)
		width := layoutStruct[i].EffWidth()
		switch width.Type() {
		case LengthLengthTypePercent:
			havePercent = true
			totalPercent += int32(width.Value())
		case LengthLengthTypeFixed:
			numFixed++
			totalFixed += int32(layoutStruct[i].EffMaxWidth())
		case LengthLengthTypeVariable:
			numVariable++
			totalVariable += int32(layoutStruct[i].EffMaxWidth())
			allocVariable += int32(w)
		}
	}

	// allocate width to percent cols
	if available > 0 && havePercent {
		for i := 0; i < nEffCols; i++ {
			width := layoutStruct[i].EffWidth()
			if width.IsPercent() {
				w := max(layoutStruct[i].EffMinWidth(), width.MinWidth(int(tableWidth)))
				available += int32(layoutStruct[i].CalcWidth() - w)
				layoutStruct[i].SetCalcWidth(w)
			}
		}
		if totalPercent > 100 {
			// remove over-allocated space from the last columns
			excess := tableWidth * (totalPercent - 100) / 100
			for i := nEffCols - 1; i >= 0; i-- {
				if layoutStruct[i].EffWidth().IsPercent() {
					w := layoutStruct[i].CalcWidth()
					reduction := min(w, int64(excess))
					// the lines below might look inconsistent, but
					// that's the way it's handled in mozilla
					excess -= int32(reduction)
					newWidth := max(layoutStruct[i].EffMinWidth(), w-reduction)
					available += int32(w - newWidth)
					layoutStruct[i].SetCalcWidth(newWidth)
					// qDebug("col %d: reducing to %d px
					// (reduction=%d)", i, newWidth, reduction );
				}
			}
		}
	}

	// then allocate width to fixed cols
	if available > 0 {
		for i := 0; i < nEffCols; i++ {
			width := layoutStruct[i].EffWidth()
			if width.IsFixed() && width.Value() > layoutStruct[i].CalcWidth() {
				available += int32(layoutStruct[i].CalcWidth() - width.Value())
				layoutStruct[i].SetCalcWidth(width.Value())
			}
		}
	}

	// now satisfy variable
	if available > 0 && numVariable > 0 {
		available += allocVariable // this gets redistributed
		// qDebug("redistributing %dpx to %d variable columns.
		// totalVariable=%d", available, numVariable, totalVariable );
		for i := 0; i < nEffCols; i++ {
			width := layoutStruct[i].EffWidth()
			if width.IsVariable() && totalVariable != 0 {
				w := max(layoutStruct[i].CalcWidth(), int64(available)*
					layoutStruct[i].EffMaxWidth()/int64(totalVariable))
				available -= int32(w)
				totalVariable -= int32(layoutStruct[i].EffMaxWidth())
				layoutStruct[i].SetCalcWidth(w)
			}
		}
	}

	// spread over fixed columns
	if available > 0 && numFixed > 0 {
		// still have some width to spread, distribute to fixed columns
		for i := 0; i < nEffCols; i++ {
			width := layoutStruct[i].EffWidth()
			if width.IsFixed() {
				w := int64(available) * layoutStruct[i].EffMaxWidth() / int64(totalFixed)
				available -= int32(w)
				totalFixed -= int32(layoutStruct[i].EffMaxWidth())
				layoutStruct[i].SetCalcWidth(layoutStruct[i].CalcWidth() + w)
			}
		}
	}

	// spread over percent columns
	if available > 0 && havePercent && totalPercent < 100 {
		// still have some width to spread, distribute weighted to
		// percent columns
		for i := 0; i < nEffCols; i++ {
			width := layoutStruct[i].EffWidth()
			if width.IsPercent() {
				w := int64(available) * width.Value() / int64(totalPercent)
				available -= int32(w)
				totalPercent -= int32(width.Value())
				layoutStruct[i].SetCalcWidth(layoutStruct[i].CalcWidth() + w)
				if available == 0 || totalPercent == 0 {
					break
				}
			}
		}
	}

	// spread over the rest
	if available > 0 {
		total := int32(nEffCols)
		// still have some width to spread
		i := nEffCols
		for i > 0 {
			i--
			w := available / total
			available -= w
			total--
			layoutStruct[i].SetCalcWidth(layoutStruct[i].CalcWidth() + int64(w))
		}
	}

	// if we have over-allocated, reduce every cell according to the
	// difference between desired width and min-width
	// this seems to produce to the pixel exact results with IE. Wonder
	// is some of this also holds for width distributing.
	if available < 0 {
		// Need to reduce cells with the following prioritization:
		// (1) Variable
		// (2) Relative
		// (3) Fixed
		// (4) Percent
		// This is basically the reverse of how we grew the cells.
		if available < 0 {
			var mw int32 = 0
			for i := nEffCols - 1; i >= 0; i-- {
				width := layoutStruct[i].EffWidth()
				if width.IsVariable() {
					mw += int32(layoutStruct[i].CalcWidth() - layoutStruct[i].EffMinWidth())
				}
			}

			for i := nEffCols - 1; i >= 0 && mw > 0; i-- {
				width := layoutStruct[i].EffWidth()
				if width.IsVariable() {
					minMaxDiff := layoutStruct[i].CalcWidth() -
						layoutStruct[i].EffMinWidth()
					reduce := int64(available) * minMaxDiff / int64(mw)
					layoutStruct[i].SetCalcWidth(layoutStruct[i].CalcWidth() + reduce)
					available -= int32(reduce)
					mw -= int32(minMaxDiff)
					if available >= 0 {
						break
					}
				}
			}
		}

		if available < 0 {
			var mw int32 = 0
			for i := nEffCols - 1; i >= 0; i-- {
				width := layoutStruct[i].EffWidth()
				if width.IsFixed() {
					mw += int32(layoutStruct[i].CalcWidth() - layoutStruct[i].EffMinWidth())
				}
			}

			for i := nEffCols - 1; i >= 0 && mw > 0; i-- {
				width := layoutStruct[i].EffWidth()
				if width.IsFixed() {
					minMaxDiff := layoutStruct[i].CalcWidth() -
						layoutStruct[i].EffMinWidth()
					reduce := int64(available) * minMaxDiff / int64(mw)
					layoutStruct[i].SetCalcWidth(layoutStruct[i].CalcWidth() + reduce)
					available -= int32(reduce)
					mw -= int32(minMaxDiff)
					if available >= 0 {
						break
					}
				}
			}
		}

		if available < 0 {
			var mw int32 = 0
			for i := nEffCols - 1; i >= 0; i-- {
				width := layoutStruct[i].EffWidth()
				if width.IsPercent() {
					mw += int32(layoutStruct[i].CalcWidth() - layoutStruct[i].EffMinWidth())
				}
			}

			for i := nEffCols - 1; i >= 0 && mw > 0; i-- {
				width := layoutStruct[i].EffWidth()
				if width.IsPercent() {
					minMaxDiff := layoutStruct[i].CalcWidth() -
						layoutStruct[i].EffMinWidth()
					reduce := int64(available) * minMaxDiff / int64(mw)
					layoutStruct[i].SetCalcWidth(layoutStruct[i].CalcWidth() + reduce)
					available -= int32(reduce)
					mw -= int32(minMaxDiff)
					if available >= 0 {
						break
					}
				}
			}
		}
	}

	var pos int32 = 0
	hspacing := int32(a.table.GetStyle().GetBorderHSpacing(c))
	columnPos := make([]int, nEffCols+1)
	for i := 0; i < nEffCols; i++ {
		columnPos[i] = int(pos)
		pos += int32(layoutStruct[i].CalcWidth() + int64(hspacing))
	}

	columnPos[len(columnPos)-1] = int(pos)

	a.table.setColumnPos(columnPos)
}

// TableBoxAutoTableLayoutLayout ports the class
// TableBox.AutoTableLayout.Layout.
type TableBoxAutoTableLayoutLayout struct {
	width       *Length
	effWidth    *Length
	minWidth    int64
	maxWidth    int64
	effMinWidth int64
	effMaxWidth int64
	calcWidth   int64
}

func NewTableBoxAutoTableLayoutLayout() *TableBoxAutoTableLayoutLayout {
	return &TableBoxAutoTableLayoutLayout{
		width:    LengthZero,
		effWidth: LengthZero,
		minWidth: 1,
		maxWidth: 1,
	}
}

func (l *TableBoxAutoTableLayoutLayout) Width() *Length {
	return l.width
}

func (l *TableBoxAutoTableLayoutLayout) SetWidth(w *Length) {
	l.width = w
}

func (l *TableBoxAutoTableLayoutLayout) EffWidth() *Length {
	return l.effWidth
}

func (l *TableBoxAutoTableLayoutLayout) SetEffWidth(w *Length) {
	l.effWidth = w
}

func (l *TableBoxAutoTableLayoutLayout) MinWidth() int64 {
	return l.minWidth
}

func (l *TableBoxAutoTableLayoutLayout) SetMinWidth(i int64) {
	l.minWidth = i
}

func (l *TableBoxAutoTableLayoutLayout) MaxWidth() int64 {
	return l.maxWidth
}

func (l *TableBoxAutoTableLayoutLayout) SetMaxWidth(i int64) {
	l.maxWidth = i
}

func (l *TableBoxAutoTableLayoutLayout) EffMinWidth() int64 {
	return l.effMinWidth
}

func (l *TableBoxAutoTableLayoutLayout) SetEffMinWidth(i int64) {
	l.effMinWidth = i
}

func (l *TableBoxAutoTableLayoutLayout) EffMaxWidth() int64 {
	return l.effMaxWidth
}

func (l *TableBoxAutoTableLayoutLayout) SetEffMaxWidth(i int64) {
	l.effMaxWidth = i
}

func (l *TableBoxAutoTableLayoutLayout) CalcWidth() int64 {
	return l.calcWidth
}

func (l *TableBoxAutoTableLayoutLayout) SetCalcWidth(i int64) {
	l.calcWidth = i
}
