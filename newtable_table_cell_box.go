// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/TableCellBox.java

package ufo

import (
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

type TableCellBox struct {
	BlockBox

	row int
	col int

	table   *TableBox
	section *TableSectionBox

	collapsedLayoutBorder   *BorderPropertySet
	collapsedPaintingBorder *BorderPropertySet

	collapsedBorderTop    *CollapsedBorderValue
	collapsedBorderRight  *CollapsedBorderValue
	collapsedBorderBottom *CollapsedBorderValue
	collapsedBorderLeft   *CollapsedBorderValue
}

var TableCellBoxSpanningCell = NewTableCellBox(nil, nil, false)

// 'double', 'solid', 'dashed', 'dotted', 'ridge', 'outset', 'groove', and the lowest: 'inset'.
var tableCellBoxBorderPriorities = func() []int {
	result := make([]int, IdentValueGetIdentCount())
	result[IdentValueDouble.FS_ID] = 1
	result[IdentValueSolid.FS_ID] = 2
	result[IdentValueDashed.FS_ID] = 3
	result[IdentValueDotted.FS_ID] = 4
	result[IdentValueRidge.FS_ID] = 5
	result[IdentValueOutset.FS_ID] = 6
	result[IdentValueGroove.FS_ID] = 7
	result[IdentValueInset.FS_ID] = 8
	return result
}()

const (
	tableCellBoxBcell     = 10
	tableCellBoxBrow      = 9
	tableCellBoxBrowgroup = 8
	tableCellBoxBcol      = 7
	tableCellBoxBtable    = 6
)

func NewTableCellBox(source *dom.Element, style CalculatedStyleI, anonymous bool) *TableCellBox {
	t := &TableCellBox{}
	initBlockBox(&t.BlockBox, source, style, anonymous)
	t.SetSelf(t)
	return t
}

// tableCellBoxCastTable is the Java cast (TableBox)box: nil stays nil, a box
// of another class panics.
func tableCellBoxCastTable(box BoxI) *TableBox {
	if box == nil {
		return nil
	}
	return box.(*TableBox)
}

// tableCellBoxCastSection is the Java cast (TableSectionBox)box: nil stays
// nil, a box of another class panics.
func tableCellBoxCastSection(box BoxI) *TableSectionBox {
	if box == nil {
		return nil
	}
	return box.(*TableSectionBox)
}

// tableCellBoxCastRow is the Java cast (TableRowBox)box: nil stays nil, a box
// of another class panics.
func tableCellBoxCastRow(box BoxI) *TableRowBox {
	if box == nil {
		return nil
	}
	return box.(*TableRowBox)
}

func (t *TableCellBox) CopyOf() BlockBoxI {
	return NewTableCellBox(t.GetElement(), t.GetStyle(), t.IsAnonymous())
}

func (t *TableCellBox) GetBorder(cssCtx CssContext) *BorderPropertySet {
	if t.GetTable().GetStyle().IsCollapseBorders() {
		// Should always be non-null, but might not be if layout code crashed
		if t.collapsedLayoutBorder == nil {
			return BorderPropertySetEmptyBorder
		}
		return t.collapsedLayoutBorder
	} else {
		return t.BlockBox.GetBorder(cssCtx)
	}
}

func (t *TableCellBox) CalcCollapsedBorder(c CssContext) {
	top := t.collapsedTopBorder(c)
	right := t.collapsedRightBorder(c)
	bottom := t.collapsedBottomBorder(c)
	left := t.collapsedLeftBorder(c)

	t.collapsedPaintingBorder = NewBorderPropertySetWithTopRightBottomLeft(top, right, bottom, left)

	// Give the extra pixel to top and left.
	t.collapsedBorderTop = top.WithWidth((top.Width() + 1) / 2)
	t.collapsedBorderRight = right.WithWidth(right.Width() / 2)
	t.collapsedBorderBottom = bottom.WithWidth(bottom.Width() / 2)
	t.collapsedBorderLeft = left.WithWidth((left.Width() + 1) / 2)
	t.collapsedLayoutBorder = NewBorderPropertySetWithTopRightBottomLeft(t.collapsedBorderTop, t.collapsedBorderRight, t.collapsedBorderBottom, t.collapsedBorderLeft)
}

func (t *TableCellBox) GetCol() int {
	return t.col
}

func (t *TableCellBox) SetCol(col int) {
	t.col = col
}

func (t *TableCellBox) GetRow() int {
	return t.row
}

func (t *TableCellBox) SetRow(row int) {
	t.row = row
}

func (t *TableCellBox) Layout(c *LayoutContext) {
	t.BlockBox.Layout(c)
}

// GetTable may return nil.
func (t *TableCellBox) GetTable() *TableBox {
	// cell -> row -> section -> table
	if t.table == nil {
		t.table = tableCellBoxCastTable(t.GetParent().GetParent().GetParent())
	}
	return t.table
}

func (t *TableCellBox) GetContainingBlockWidth() int {
	// A table row has no width of its own (it always spans the full table), so at the
	// point cell widths are calculated (TableSectionBox.setCellWidths()) the row's content
	// width is not yet established. Percentages (e.g. max-width: 10%) must resolve against
	// the table's width instead, per CSS 2.1 17.4.
	table := t.GetTable()
	if table == nil {
		return t.BlockBox.GetContainingBlockWidth()
	}
	return table.GetContentWidth()
}

// GetSection may return nil.
func (t *TableCellBox) GetSection() *TableSectionBox {
	if t.section == nil {
		t.section = tableCellBoxCastSection(t.GetParent().GetParent())
	}
	return t.section
}

func (t *TableCellBox) GetOuterStyleWidth(c CssContext) *Length {
	result := t.GetStyle().AsLength(c, CSSNameWidth)
	if result.IsVariable() || result.IsPercent() {
		return result
	}

	// When box-sizing: border-box, the CSS 'width' already includes padding and border,
	// so it IS the outer width. No need to add them again.
	if t.GetStyle().IsBorderBox() {
		return result
	}

	bordersAndPadding := 0
	border := t.GetBorder(c)
	bordersAndPadding += int(border.Left()) + int(border.Right())

	padding := t.GetPadding(c)
	bordersAndPadding += int(padding.Left()) + int(padding.Right())

	return NewLength(result.Value()+int64(bordersAndPadding), result.Type())
}

func (t *TableCellBox) GetOuterStyleOrColWidth(c CssContext) *Length {
	result := t.GetOuterStyleWidth(c)
	if t.GetStyle().GetColSpan() > 1 || !result.IsVariable() {
		return result
	}
	col := t.GetTable().ColElement(t.GetCol())
	if col != nil {
		// XXX Need to add in collapsed borders from cell (if collapsing borders)
		result = col.GetStyle().AsLength(c, CSSNameWidth)
	}
	return result
}

func (t *TableCellBox) SetLayoutWidth(c *LayoutContext, width int) {
	t.CalcDimensions(c)

	t.SetContentWidth(width - t.GetLeftMBP() - t.GetRightMBP())
	if t.IsFixedWidthAdvisoryOnly() && t.GetStyle().GetColSpan() <= 1 {
		t.ApplyCSSMinMaxWidth(c)
	}
}

func (t *TableCellBox) IsAutoHeight() bool {
	return t.GetStyle().IsAutoHeight() || !t.GetStyle().HasAbsoluteUnit(CSSNameHeight)
}

func (t *TableCellBox) CalcBaseline(c *LayoutContext) int {
	result := t.BlockBox.CalcBaseline(c)
	if result != BlockBoxNoBaseline {
		return result
	} else {
		contentArea := t.GetContentAreaEdge(t.GetAbsX(), t.GetAbsY(), c)
		return int(contentArea.GetY())
	}
}

func (t *TableCellBox) CalcBlockBaseline(c *LayoutContext) int {
	return t.BlockBox.CalcBaseline(c)
}

// tableCellBoxMoveContentOperation is the lambda that Java's moveContent
// passes to FloatManager.performFloatOperation.
type tableCellBoxMoveContentOperation struct {
	deltaY int
}

func (o *tableCellBoxMoveContentOperation) Operate(floater BoxI) {
	floater.SetY(floater.GetY() + o.deltaY)
}

func (t *TableCellBox) MoveContent(deltaY int) {
	for i := 0; i < t.GetChildCount(); i++ {
		b := t.GetChild(i)
		b.SetY(b.GetY() + deltaY)
	}

	t.GetPersistentBFC().GetFloatManager().PerformFloatOperation(
		&tableCellBoxMoveContentOperation{deltaY: deltaY})

	t.CalcChildLocations()
}

func (t *TableCellBox) IsPageBreaksChange(c *LayoutContext, posDeltaY int) bool {
	if !c.IsPageBreaksAllowed() {
		return false
	}

	page := c.GetRootLayer().GetFirstPage(c, t)

	bottomEdge := t.GetAbsY() + t.GetChildrenHeight()

	return page != nil && (bottomEdge >= page.GetBottom()-c.GetExtraSpaceBottom() ||
		bottomEdge+posDeltaY >= page.GetBottom()-c.GetExtraSpaceBottom())
}

func (t *TableCellBox) GetVerticalAlign() *IdentValue {
	val := t.GetStyle().GetIdent(CSSNameVerticalAlign)

	if val == IdentValueTop || val == IdentValueMiddle || val == IdentValueBottom {
		return val
	} else {
		return IdentValueBaseline
	}
}

func (t *TableCellBox) isPaintBackgroundsAndBorders() bool {
	showEmpty := t.GetStyle().IsShowEmptyCells()
	// XXX Not quite right, but good enough for now
	// (e.g. absolute boxes will be counted as content here when the spec
	// says the cell should be treated as empty).
	return showEmpty || t.GetChildrenContentType() != BlockBoxContentTypeEmpty
}

func (t *TableCellBox) PaintBackground(c *RenderingContext) {
	if t.isPaintBackgroundsAndBorders() && t.GetStyle().IsVisible() {
		var bounds *geom.Rectangle
		if c.IsPrint() && t.GetTable().GetStyle().IsPaginateTable() {
			bounds = t.getContentLimitedBorderEdge(c)
			bounds = t.adjustBoundsToAvoidTheadOverlap(c, bounds)
		} else {
			bounds = t.GetPaintingBorderEdge(c)
		}

		if bounds != nil && bounds.Height > 0 {
			t.paintBackgroundStack(c, bounds)
		}
	}
}

// getTheadBottom gets the bottom Y position of the thead section in the table.
// It returns the absolute Y position of thead's bottom, or -1 if no thead is
// found.
func (t *TableCellBox) getTheadBottom(table *TableBox) int {
	for i := 0; i < table.GetChildCount(); i++ {
		child := table.GetChild(i)
		if tableSection, ok := child.(*TableSectionBox); ok && tableSection.IsHeader() {
			return tableSection.GetAbsY() + tableSection.GetHeight()
		}
	}
	return -1
}

// adjustBoundsToAvoidTheadOverlap adjusts the bounds of a rowspan cell on
// continuation pages to avoid overlapping with thead. For rowspan cells that
// span multiple pages, this ensures the cell's rendering starts after the
// thead section on subsequent pages.
//
// It returns adjusted bounds that don't overlap with thead, or the original
// bounds if no adjustment is needed.
func (t *TableCellBox) adjustBoundsToAvoidTheadOverlap(c *RenderingContext, bounds *geom.Rectangle) *geom.Rectangle {
	// Only adjust for rowspan cells on subsequent pages
	if bounds == nil || t.GetStyle().GetRowSpan() <= 1 {
		return bounds
	}

	contentLimitContainer := tableCellBoxCastRow(t.GetParent()).GetContentLimitContainer()
	if contentLimitContainer == nil || c.GetPageNo() <= contentLimitContainer.GetInitialPageNo() {
		return bounds
	}

	// This is a continuation page - check for thead overlap
	table := t.GetTable()
	if table == nil {
		return bounds
	}

	for i := 0; i < table.GetChildCount(); i++ {
		child := table.GetChild(i)
		if tableSection, ok := child.(*TableSectionBox); ok && tableSection.IsHeader() {
			// Found thead - get its bottom position
			theadBottom := tableSection.GetAbsY() + tableSection.GetHeight()
			if bounds.Y < theadBottom {
				// Adjust bounds to start after thead
				overlap := theadBottom - bounds.Y
				bounds = geom.NewRectangleFromRectangle(bounds) // Create a copy to avoid modifying original
				bounds.Y = theadBottom
				bounds.Height = max(0, bounds.Height-overlap)
			}
			break
		}
	}

	return bounds
}

func (t *TableCellBox) paintBackgroundStack(c *RenderingContext, bounds *geom.Rectangle) {
	var imageContainer *geom.Rectangle

	border := t.GetStyle().GetBorder(c)
	column := t.GetTable().ColElement(t.GetCol())
	if column != nil {
		c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(
			c, column.GetStyle(),
			bounds, t.GetTable().GetColumnBounds(c, t.GetCol()),
			border)
	}

	row := t.GetParent()
	section := row.GetParent()

	tableStyle := t.GetTable().GetStyle()

	sectionStyle := section.GetStyle()

	imageContainer = section.GetPaintingBorderEdge(c)
	imageContainer.Y += tableStyle.GetBorderVSpacing(c)
	imageContainer.Height -= tableStyle.GetBorderVSpacing(c)
	imageContainer.X += tableStyle.GetBorderHSpacing(c)
	imageContainer.Width -= 2 * tableStyle.GetBorderHSpacing(c)

	c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, sectionStyle, bounds, imageContainer, sectionStyle.GetBorder(c))

	rowStyle := row.GetStyle()

	imageContainer = row.GetPaintingBorderEdge(c)
	imageContainer.X += tableStyle.GetBorderHSpacing(c)
	imageContainer.Width -= 2 * tableStyle.GetBorderHSpacing(c)

	c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, rowStyle, bounds, imageContainer, rowStyle.GetBorder(c))
	c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, t.GetStyle(), bounds, t.GetPaintingBorderEdge(c), border)
}

func (t *TableCellBox) PaintBorder(c *RenderingContext) {
	if t.isPaintBackgroundsAndBorders() && !t.HasCollapsedPaintingBorder() {
		// Collapsed table borders are painted separately
		if c.IsPrint() && t.GetTable().GetStyle().IsPaginateTable() && t.GetStyle().IsVisible() {
			bounds := t.getContentLimitedBorderEdge(c)
			bounds = t.adjustBoundsToAvoidTheadOverlap(c, bounds)
			if bounds != nil && bounds.Height > 0 {
				c.GetOutputDevice().PaintBorderWithStyleEdgeSides(c, t.GetStyle(), bounds, t.GetBorderSides())
			}
		} else {
			t.BlockBox.PaintBorder(c)
		}
	}
}

func (t *TableCellBox) PaintCollapsedBorder(c *RenderingContext, side int) {
	c.GetOutputDevice().PaintCollapsedBorder(
		c, t.GetCollapsedPaintingBorder(), t.getCollapsedBorderBounds(c), side)
}

// getContentLimitedBorderEdge may return nil.
func (t *TableCellBox) getContentLimitedBorderEdge(c *RenderingContext) *geom.Rectangle {
	result := t.GetPaintingBorderEdge(c)

	section := t.GetSection()
	if section != nil && (section.IsHeader() || section.IsFooter()) {
		return result
	}

	contentLimitContainer := tableCellBoxCastRow(t.GetParent()).GetContentLimitContainer()
	var limit *ContentLimit
	if contentLimitContainer != nil {
		limit = contentLimitContainer.GetContentLimit(c.GetPageNo())
	}

	if limit == nil {
		return nil
	} else {
		if limit.GetTop() == ContentLimitUndefined ||
			limit.GetBottom() == ContentLimitUndefined {
			return result
		}

		var top int
		if c.GetPageNo() == contentLimitContainer.GetInitialPageNo() {
			top = result.Y
		} else {
			// For rowspan cells on continuation pages, start from thead bottom
			rowSpan := t.GetStyle().GetRowSpan()
			if rowSpan > 1 {
				table := t.GetTable()
				if table != nil {
					theadBottom := t.getTheadBottom(table)
					if theadBottom > 0 {
						top = theadBottom
					} else {
						top = limit.GetTop() - tableCellBoxCastRow(t.GetParent()).GetExtraSpaceTop()
					}
				} else {
					top = limit.GetTop() - tableCellBoxCastRow(t.GetParent()).GetExtraSpaceTop()
				}
			} else {
				top = limit.GetTop() - tableCellBoxCastRow(t.GetParent()).GetExtraSpaceTop()
			}
		}

		var bottom int
		if c.GetPageNo() == contentLimitContainer.GetLastPageNo() {
			bottom = result.Y + result.Height
		} else {
			// For rowspan cells, find the maximum bottom among all rows covered
			var maxBottom int
			if limit.GetBottom() != ContentLimitUndefined {
				maxBottom = limit.GetBottom() + tableCellBoxCastRow(t.GetParent()).GetExtraSpaceBottom()
			} else {
				maxBottom = result.Y + result.Height
			}

			// Check all rows covered by rowspan by traversing siblings
			rowSpan := t.GetStyle().GetRowSpan()
			currentBox := t.GetParent()
			for i := 1; i < rowSpan && currentBox != nil; i++ {
				currentBox = currentBox.GetNextSibling()
				if spannedRow, ok := currentBox.(*TableRowBox); ok {
					spannedContainer := spannedRow.GetContentLimitContainer()
					if spannedContainer != nil {
						spannedLimit := spannedContainer.GetContentLimit(c.GetPageNo())
						if spannedLimit != nil && spannedLimit.GetBottom() != ContentLimitUndefined {
							spannedBottom := spannedLimit.GetBottom() + spannedRow.GetExtraSpaceBottom()
							maxBottom = max(maxBottom, spannedBottom)
						}
					}
				}
			}

			bottom = min(result.Y+result.Height, maxBottom)
		}

		result.Y = top
		result.Height = bottom - top

		return result
	}
}

func (t *TableCellBox) GetChildrenClipEdge(c *RenderingContext) *geom.Rectangle {
	if c.IsPrint() && t.GetTable().GetStyle().IsPaginateTable() {
		bounds := t.getContentLimitedBorderEdge(c)
		if bounds != nil {
			border := t.GetBorder(c)
			padding := t.GetPadding(c)
			bounds.Y += int(border.Top()) + int(padding.Top())
			bounds.Height -= int(border.Height()) + int(padding.Height())
			return bounds
		}
	}

	return t.BlockBox.GetChildrenClipEdge(c)
}

func (t *TableCellBox) IsFixedWidthAdvisoryOnly() bool {
	return t.GetTable().GetStyle().IsIdent(CSSNameTableLayout, IdentValueAuto)
}

func (t *TableCellBox) IsSkipWhenCollapsingMargins() bool {
	return true
}

// The following rules apply for resolving conflicts and figuring out which
// border
// to use.
// (1) Borders with the 'border-style' of 'hidden' take precedence over all
// other conflicting
// borders. Any border with this value suppresses all borders at this
// location.
// (2) Borders with a style of 'none' have the lowest priority. Only if the
// border properties of all
// the elements meeting at this edge are 'none' will the border be omitted
// (but note that 'none' is
// the default value for the border style.)
// (3) If none of the styles are 'hidden' and at least one of them is not
// 'none', then narrow borders
// are discarded in favor of wider ones. If several have the same
// 'border-width' then styles are preferred
// in this order: 'double', 'solid', 'dashed', 'dotted', 'ridge', 'outset',
// 'groove', and the lowest: 'inset'.
// (4) If border styles differ only in color, then a style set on a cell
// wins over one on a row,
// which wins over a row group, column, column group and, lastly, table. It
// is undefined which color
// is used when two elements of the same type disagree.
//
// The result is nil when returnNullOnEqual is set and the two borders have the
// same width, style and precedence.
func TableCellBoxCompareBordersWithReturnNullOnEqual(
	border1 *CollapsedBorderValue, border2 *CollapsedBorderValue, returnNullOnEqual bool) *CollapsedBorderValue {
	// Sanity check the values passed in.  If either is null, return the other.
	if !border2.Defined() {
		return border1
	}

	if !border1.Defined() {
		return border2
	}

	// Rule #1 above.
	if border1.Style() == IdentValueHidden {
		return border1
	}
	if border2.Style() == IdentValueHidden {
		return border2
	}

	// Rule #2 above. A style of 'none' has the lowest priority and always loses
	// to any other border.
	if border2.Style() == IdentValueNone {
		return border1
	}

	if border1.Style() == IdentValueNone {
		return border2
	}

	// The first part of rule #3 above. Wider borders win.
	if border1.Width() != border2.Width() {
		if border1.Width() > border2.Width() {
			return border1
		}
		return border2
	}

	// The borders have equal width. Sort by border style.
	if border1.Style() != border2.Style() {
		if tableCellBoxBorderPriorities[border1.Style().FS_ID] >
			tableCellBoxBorderPriorities[border2.Style().FS_ID] {
			return border1
		}
		return border2
	}

	// The border have the same width and style. Rely on precedence (cell
	// over row group, etc.)
	if returnNullOnEqual && border1.Precedence() == border2.Precedence() {
		return nil
	} else {
		if border1.Precedence() >= border2.Precedence() {
			return border1
		}
		return border2
	}
}

func tableCellBoxCompareBorders(
	border1 *CollapsedBorderValue, border2 *CollapsedBorderValue) *CollapsedBorderValue {
	return TableCellBoxCompareBordersWithReturnNullOnEqual(border1, border2, false)
}

func (t *TableCellBox) collapsedLeftBorder(c CssContext) *CollapsedBorderValue {
	border := t.GetStyle().GetBorder(c)
	// For border left, we need to check, in order of precedence:
	// (1) Our left border.
	result := CollapsedBorderValueBorderLeft(border, tableCellBoxBcell)

	// (2) The previous cell's right border.
	prevCell := t.GetTable().CellLeft(t)
	if prevCell != nil {
		result = tableCellBoxCompareBorders(
			result, CollapsedBorderValueBorderRight(prevCell.GetStyle().GetBorder(c), tableCellBoxBcell))
		if result.Hidden() {
			return result
		}
	} else if t.GetCol() == 0 {
		// (3) Our row's left border.
		result = tableCellBoxCompareBorders(
			result, CollapsedBorderValueBorderLeft(t.GetParent().GetStyle().GetBorder(c), tableCellBoxBrow))
		if result.Hidden() {
			return result
		}

		// (4) Our row group's left border.
		result = tableCellBoxCompareBorders(
			result, CollapsedBorderValueBorderLeft(t.GetSection().GetStyle().GetBorder(c), tableCellBoxBrowgroup))
		if result.Hidden() {
			return result
		}
	}

	// (5) Our column's left border.
	colElt := t.GetTable().ColElement(t.GetCol())
	if colElt != nil {
		result = tableCellBoxCompareBorders(
			result, CollapsedBorderValueBorderLeft(colElt.GetStyle().GetBorder(c), tableCellBoxBcol))
		if result.Hidden() {
			return result
		}
	}

	// (6) The previous column's right border.
	if t.GetCol() > 0 {
		colElt = t.GetTable().ColElement(t.GetCol() - 1)
		if colElt != nil {
			result = tableCellBoxCompareBorders(
				result, CollapsedBorderValueBorderRight(colElt.GetStyle().GetBorder(c), tableCellBoxBcol))
			if result.Hidden() {
				return result
			}
		}
	}

	if t.GetCol() == 0 {
		// (7) The table's left border.
		result = tableCellBoxCompareBorders(
			result, CollapsedBorderValueBorderLeft(t.GetTable().GetStyle().GetBorder(c), tableCellBoxBtable))
		if result.Hidden() {
			return result
		}
	}

	return result
}

func (t *TableCellBox) collapsedRightBorder(c CssContext) *CollapsedBorderValue {
	tableElt := t.GetTable()
	inLastColumn := false
	effCol := tableElt.ColToEffCol(t.GetCol() + t.GetStyle().GetColSpan() - 1)
	if effCol == tableElt.NumEffCols()-1 {
		inLastColumn = true
	}

	// For border right, we need to check, in order of precedence:
	// (1) Our right border.
	result := CollapsedBorderValueBorderRight(t.GetStyle().GetBorder(c), tableCellBoxBcell)

	// (2) The next cell's left border.
	if !inLastColumn {
		nextCell := tableElt.CellRight(t)
		if nextCell != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderLeft(nextCell.GetStyle().GetBorder(c), tableCellBoxBcell))
			if result.Hidden() {
				return result
			}
		}
	} else {
		// (3) Our row's right border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderRight(t.GetParent().GetStyle().GetBorder(c), tableCellBoxBrow))
		if result.Hidden() {
			return result
		}

		// (4) Our row group's right border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderRight(t.GetSection().GetStyle().GetBorder(c), tableCellBoxBrowgroup))
		if result.Hidden() {
			return result
		}
	}

	// (5) Our column's right border.
	colElt := t.GetTable().ColElement(t.GetCol() + t.GetStyle().GetColSpan() - 1)
	if colElt != nil {
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderRight(colElt.GetStyle().GetBorder(c), tableCellBoxBcol))
		if result.Hidden() {
			return result
		}
	}

	// (6) The next column's left border.
	if !inLastColumn {
		colElt = tableElt.ColElement(t.GetCol() + t.GetStyle().GetColSpan())
		if colElt != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderLeft(colElt.GetStyle().GetBorder(c), tableCellBoxBcol))
			if result.Hidden() {
				return result
			}
		}
	} else {
		// (7) The table's right border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderRight(tableElt.GetStyle().GetBorder(c), tableCellBoxBtable))
		if result.Hidden() {
			return result
		}
	}

	return result
}

func (t *TableCellBox) collapsedTopBorder(c CssContext) *CollapsedBorderValue {
	// For border top, we need to check, in order of precedence:
	// (1) Our top border.
	result := CollapsedBorderValueBorderTop(t.GetStyle().GetBorder(c), tableCellBoxBcell)

	prevCell := t.GetTable().CellAbove(t)
	if prevCell != nil {
		// (2) A previous cell's bottom border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderBottom(prevCell.GetStyle().GetBorder(c), tableCellBoxBcell))
		if result.Hidden() {
			return result
		}
	}

	// (3) Our row's top border.
	result = tableCellBoxCompareBorders(result,
		CollapsedBorderValueBorderTop(t.GetParent().GetStyle().GetBorder(c), tableCellBoxBrow))
	if result.Hidden() {
		return result
	}

	// (4) The previous row's bottom border.
	if prevCell != nil {
		var prevRow *TableRowBox
		if prevCell.GetSection() == t.GetSection() {
			prevRow = tableCellBoxCastRow(t.GetParent().GetPreviousSibling())
		} else {
			prevRow = prevCell.GetSection().GetLastRow()
		}

		if prevRow != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderBottom(prevRow.GetStyle().GetBorder(c), tableCellBoxBrow))
			if result.Hidden() {
				return result
			}
		}
	}

	// Now check row groups.
	currSection := t.GetSection()
	if t.GetRow() == 0 {
		// (5) Our row group's top border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderTop(currSection.GetStyle().GetBorder(c), tableCellBoxBrowgroup))
		if result.Hidden() {
			return result
		}

		// (6) Previous row group's bottom border.
		currSection = t.GetTable().SectionAbove(currSection, false)
		if currSection != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderBottom(currSection.GetStyle().GetBorder(c), tableCellBoxBrowgroup))
			if result.Hidden() {
				return result
			}
		}
	}

	if currSection == nil {
		// (8) Our column's top border.
		colElt := t.GetTable().ColElement(t.GetCol())
		if colElt != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderTop(colElt.GetStyle().GetBorder(c), tableCellBoxBcol))
			if result.Hidden() {
				return result
			}
		}

		// (9) The table's top border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderTop(t.GetTable().GetStyle().GetBorder(c), tableCellBoxBtable))
		if result.Hidden() {
			return result
		}
	}

	return result
}

func (t *TableCellBox) collapsedBottomBorder(c CssContext) *CollapsedBorderValue {
	// For border top, we need to check, in order of precedence:
	// (1) Our bottom border.
	result := CollapsedBorderValueBorderBottom(t.GetStyle().GetBorder(c), tableCellBoxBcell)

	nextCell := t.GetTable().CellBelow(t)
	if nextCell != nil {
		// (2) A following cell's top border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderTop(nextCell.GetStyle().GetBorder(c), tableCellBoxBcell))
		if result.Hidden() {
			return result
		}
	}

	// (3) Our row's bottom border. (FIXME: Deal with rowspan!)
	result = tableCellBoxCompareBorders(result,
		CollapsedBorderValueBorderBottom(t.GetParent().GetStyle().GetBorder(c), tableCellBoxBrow))
	if result.Hidden() {
		return result
	}

	// (4) The next row's top border.
	if nextCell != nil {
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderTop(nextCell.GetParent().GetStyle().GetBorder(c), tableCellBoxBrow))
		if result.Hidden() {
			return result
		}
	}

	// Now check row groups.
	currSection := t.GetSection()
	if t.GetRow()+t.GetStyle().GetRowSpan() >= currSection.NumRows() {
		// (5) Our row group's bottom border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderBottom(currSection.GetStyle().GetBorder(c), tableCellBoxBrowgroup))
		if result.Hidden() {
			return result
		}

		// (6) Following row group's top border.
		currSection = t.GetTable().SectionBelow(currSection, false)
		if currSection != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderTop(currSection.GetStyle().GetBorder(c), tableCellBoxBrowgroup))
			if result.Hidden() {
				return result
			}
		}
	}

	if currSection == nil {
		// (8) Our column's bottom border.
		colElt := t.GetTable().ColElement(t.GetCol())
		if colElt != nil {
			result = tableCellBoxCompareBorders(result,
				CollapsedBorderValueBorderBottom(colElt.GetStyle().GetBorder(c), tableCellBoxBcol))
			if result.Hidden() {
				return result
			}
		}

		// (9) The table's bottom border.
		result = tableCellBoxCompareBorders(result,
			CollapsedBorderValueBorderBottom(t.GetTable().GetStyle().GetBorder(c), tableCellBoxBtable))
		if result.Hidden() {
			return result
		}
	}

	return result
}

// getCollapsedBorderBounds may return nil.
func (t *TableCellBox) getCollapsedBorderBounds(c CssContext) *geom.Rectangle {
	border := t.GetCollapsedPaintingBorder()
	var bounds *geom.Rectangle

	// Use content-limited border edge for paginated tables to prevent borders
	// from extending beyond table boundaries when spanning multiple pages
	if renderingContext, ok := c.(*RenderingContext); ok && renderingContext.IsPrint() &&
		t.GetTable() != nil && t.GetTable().GetStyle() != nil &&
		t.GetTable().GetStyle().IsPaginateTable() {
		bounds = t.getContentLimitedBorderEdge(renderingContext)
		if bounds == nil {
			bounds = t.GetPaintingBorderEdge(c)
		}
		bounds = t.adjustBoundsToAvoidTheadOverlap(renderingContext, bounds)
	} else {
		bounds = t.GetPaintingBorderEdge(c)
	}

	if bounds == nil {
		return nil
	}

	bounds.X -= int(border.Left()) / 2
	bounds.Y -= int(border.Top()) / 2
	bounds.Width += int(border.Left())/2 + (int(border.Right())+1)/2
	bounds.Height += int(border.Top())/2 + (int(border.Bottom())+1)/2

	return bounds
}

func (t *TableCellBox) GetPaintingClipEdge(c CssContext) *geom.Rectangle {
	if t.HasCollapsedPaintingBorder() {
		return t.getCollapsedBorderBounds(c)
	} else {
		return t.BlockBox.GetPaintingClipEdge(c)
	}
}

func (t *TableCellBox) HasCollapsedPaintingBorder() bool {
	return t.collapsedPaintingBorder != nil
}

func (t *TableCellBox) GetCollapsedPaintingBorder() *BorderPropertySet {
	return t.collapsedPaintingBorder
}

func (t *TableCellBox) GetCollapsedBorderBottom() *CollapsedBorderValue {
	return t.collapsedBorderBottom
}

func (t *TableCellBox) GetCollapsedBorderLeft() *CollapsedBorderValue {
	return t.collapsedBorderLeft
}

func (t *TableCellBox) GetCollapsedBorderRight() *CollapsedBorderValue {
	return t.collapsedBorderRight
}

func (t *TableCellBox) GetCollapsedBorderTop() *CollapsedBorderValue {
	return t.collapsedBorderTop
}

// AddCollapsedBorders appends to the list that Java's borders parameter refers
// to and returns the extended list. CollapsedBorderValue has no equals method
// in Java, so the set all is keyed by object identity, which is the pointer.
func (t *TableCellBox) AddCollapsedBorders(all map[*CollapsedBorderValue]struct{}, borders []*CollapsedBorderSide) []*CollapsedBorderSide {
	if _, contains := all[t.collapsedBorderTop]; t.collapsedBorderTop.Exists() && !contains {
		all[t.collapsedBorderTop] = struct{}{}
		borders = append(borders, NewCollapsedBorderSide(t, BorderPainterTop))
	}

	if _, contains := all[t.collapsedBorderRight]; t.collapsedBorderRight.Exists() && !contains {
		all[t.collapsedBorderRight] = struct{}{}
		borders = append(borders, NewCollapsedBorderSide(t, BorderPainterRight))
	}

	if _, contains := all[t.collapsedBorderBottom]; t.collapsedBorderBottom.Exists() && !contains {
		all[t.collapsedBorderBottom] = struct{}{}
		borders = append(borders, NewCollapsedBorderSide(t, BorderPainterBottom))
	}

	if _, contains := all[t.collapsedBorderLeft]; t.collapsedBorderLeft.Exists() && !contains {
		all[t.collapsedBorderLeft] = struct{}{}
		borders = append(borders, NewCollapsedBorderSide(t, BorderPainterLeft))
	}

	return borders
}

// Treat height as if it specifies border height (i.e.
// box-sizing: border-box in CSS3).  There doesn't seem to be any
// justification in the spec for this, but everybody does it
// (in standards mode) so I guess we will too
func (t *TableCellBox) GetCSSHeight(c CssContext) int {
	if t.GetStyle().IsAutoHeight() {
		return -1
	} else {
		result := calculatedStyleFloatToInt(t.GetStyle().GetFloatPropertyProportionalWidth(
			CSSNameHeight, float32(t.GetContainingBlock().GetContentWidth()), c))

		if !t.GetStyle().IsBorderBox() {
			// Legacy XHTML 1.0 behavior: treat CSS 'height' as border-box height
			// and convert to content height here.
			// When box-sizing: border-box is explicitly set, calcDimensions() in
			// BlockBox will perform this subtraction correctly — avoid double-subtracting.
			border := t.GetBorder(c)
			result -= int(border.Top()) + int(border.Bottom())

			padding := t.GetPadding(c)
			result -= int(padding.Top()) + int(padding.Bottom())
		}

		if result >= 0 {
			return result
		}
		return -1
	}
}

func (t *TableCellBox) IsAllowHeightToShrink() bool {
	return false
}

func (t *TableCellBox) IsNeedsClipOnPaint(c *RenderingContext) bool {
	result := t.BlockBox.IsNeedsClipOnPaint(c)
	if result {
		return result
	}
	contentLimitContainer := tableCellBoxCastRow(t.GetParent()).GetContentLimitContainer()
	if contentLimitContainer == nil {
		return false
	}
	return c.IsPrint() && t.GetTable().GetStyle().IsPaginateTable() &&
		contentLimitContainer.IsContainsMultiplePages()
}
