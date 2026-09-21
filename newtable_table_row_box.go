// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/TableRowBox.java

package ufo

import (
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

type TableRowBox struct {
	BlockBox

	baseline              int
	haveBaseline          bool
	heightOverride        int
	contentLimitContainer *ContentLimitContainer

	extraSpaceTop    int
	extraSpaceBottom int
}

func NewTableRowBox(element *dom.Element, style CalculatedStyleI, anonymous bool) *TableRowBox {
	t := &TableRowBox{}
	initBlockBox(&t.BlockBox, element, style, anonymous)
	t.SetSelf(t)
	return t
}

func (t *TableRowBox) CopyOf() BlockBoxI {
	return NewTableRowBox(t.GetElement(), t.GetStyle(), t.IsAnonymous())
}

func (t *TableRowBox) IsAutoHeight() bool {
	return t.GetStyle().IsAutoHeight() || !t.GetStyle().HasAbsoluteUnit(CSSNameHeight)
}

func (t *TableRowBox) getTable() *TableBox {
	// row -> section -> table
	return tableCellBoxCastTable(t.GetParent().GetParent())
}

func (t *TableRowBox) getSection() *TableSectionBox {
	return tableCellBoxCastSection(t.GetParent())
}

func (t *TableRowBox) LayoutWithContentStart(c *LayoutContext, contentStart int) {
	running := c.IsPrint() && t.getTable().GetStyle().IsPaginateTable()
	prevExtraTop := 0
	prevExtraBottom := 0

	if running {
		prevExtraTop = c.GetExtraSpaceTop()
		prevExtraBottom = c.GetExtraSpaceBottom()

		t.calcExtraSpaceTop(c)
		t.calcExtraSpaceBottom(c)

		c.SetExtraSpaceTop(c.GetExtraSpaceTop() + t.GetExtraSpaceTop())
		c.SetExtraSpaceBottom(c.GetExtraSpaceBottom() + t.GetExtraSpaceBottom())
	}

	t.BlockBox.LayoutWithContentStart(c, contentStart)

	if running {
		if t.isShouldMoveToNextPage(c) {
			if t.getTable().GetFirstBodyRow() == t {
				// XXX Performance problem here.  This forces the table
				// to move to the next page (which we want), but the initial
				// table layout run still completes (which we don't)
				t.getTable().SetNeedPageClear(true)
			} else {
				t.SetNeedPageClear(true)
			}
		}
		c.SetExtraSpaceTop(prevExtraTop)
		c.SetExtraSpaceBottom(prevExtraBottom)
	}
}

func (t *TableRowBox) isShouldMoveToNextPage(c *LayoutContext) bool {
	page := c.GetRootLayer().GetFirstPage(c, t)

	if t.GetAbsY()+t.GetHeight() < page.GetBottom() {
		return false
	}

	for _, box := range t.GetChildren() {
		cell := box.(*TableCellBox)
		baseline := cell.CalcBlockBaseline(c)
		if baseline != BlockBoxNoBaseline && baseline < page.GetBottom() {
			return false
		}
	}

	return true
}

func (t *TableRowBox) AnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer) {
	if t.getTable().GetStyle().IsPaginateTable() {
		t.contentLimitContainer = t.BuildContainerAndAnalyzePageBreaks(c, container)

		if container != nil && t.contentLimitContainer.IsContainsMultiplePages() {
			t.PropagateExtraSpace(c, container, t.contentLimitContainer, t.GetExtraSpaceTop(), t.GetExtraSpaceBottom())
		}
	} else {
		t.BlockBox.AnalyzePageBreaks(c, container)
	}
}

func (t *TableRowBox) calcExtraSpaceTop(c *LayoutContext) {
	maxBorderAndPadding := 0

	for _, box := range t.GetChildren() {
		cell := box.(*TableCellBox)

		borderAndPadding := int(cell.GetPadding(c).Top()) + int(cell.GetBorder(c).Top())
		if borderAndPadding > maxBorderAndPadding {
			maxBorderAndPadding = borderAndPadding
		}
	}

	t.extraSpaceTop = maxBorderAndPadding
}

func (t *TableRowBox) calcExtraSpaceBottom(c *LayoutContext) {
	maxBorderAndPadding := 0

	cRow := t.GetIndex()
	totalRows := t.getSection().NumRows()
	grid := t.getSection().GetGrid()
	if len(grid) != 0 && cRow < len(grid) {
		row := grid[cRow].GetRow()
		for cCol := 0; cCol < len(row); cCol++ {
			cell := row[cCol]

			if cell == nil || cell == TableCellBoxSpanningCell {
				continue
			}
			if cRow < totalRows-1 && t.getSection().CellAt(cRow+1, cCol) == cell {
				continue
			}

			borderAndPadding := int(cell.GetPadding(c).Bottom()) + int(cell.GetBorder(c).Bottom())
			if borderAndPadding > maxBorderAndPadding {
				maxBorderAndPadding = borderAndPadding
			}
		}
	}

	t.extraSpaceBottom = maxBorderAndPadding
}

func (t *TableRowBox) LayoutChildren(c *LayoutContext, contentStart int) {
	t.SetState(BoxStateChildrenFlux)
	t.EnsureChildren(c)

	section := t.getSection()
	if section.IsNeedCellWidthCalc() {
		section.SetCellWidths(c)
		section.SetNeedCellWidthCalc(false)
	}

	if t.GetChildrenContentType() != BlockBoxContentTypeEmpty {
		for _, box := range t.GetChildren() {
			cell := box.(*TableCellBox)

			t.layoutCell(c, cell, 0)

		}
	}

	t.SetState(BoxStateDone)
}

func (t *TableRowBox) alignBaselineAlignedCells(c *LayoutContext) {
	baselines := make([]int, t.GetChildCount())
	lowest := math.MinInt32
	found := false
	for i := 0; i < t.GetChildCount(); i++ {
		cell := t.GetChild(i).(*TableCellBox)

		if cell.GetVerticalAlign() == IdentValueBaseline {
			baseline := cell.CalcBaseline(c)
			baselines[i] = baseline
			if baseline > lowest {
				lowest = baseline
			}
			found = true
		}
	}

	if found {
		for i := 0; i < t.GetChildCount(); i++ {
			cell := t.GetChild(i).(*TableCellBox)

			if cell.GetVerticalAlign() == IdentValueBaseline {
				deltaY := lowest - baselines[i]
				if deltaY != 0 {
					if c.IsPrint() && cell.IsPageBreaksChange(c, deltaY) {
						t.relayoutCell(c, cell, deltaY)
					} else {
						cell.MoveContent(deltaY)
						cell.SetHeight(cell.GetHeight() + deltaY)
					}
				}
			}
		}

		t.SetBaseline(lowest - t.GetAbsY())
		t.SetHaveBaseline(true)
	}
}

func (t *TableRowBox) alignMiddleAndBottomAlignedCells(c *LayoutContext) bool {
	needRowHeightRecalc := false

	cRow := t.GetIndex()
	totalRows := t.getSection().NumRows()
	grid := t.getSection().GetGrid()
	if len(grid) != 0 && cRow < len(grid) {
		row := grid[cRow].GetRow()
		for cCol := 0; cCol < len(row); cCol++ {
			cell := row[cCol]

			if cell == nil || cell == TableCellBoxSpanningCell {
				continue
			}
			if cRow < totalRows-1 && t.getSection().CellAt(cRow+1, cCol) == cell {
				continue
			}

			val := cell.GetVerticalAlign()
			if val == IdentValueMiddle || val == IdentValueBottom {
				deltaY := t.calcMiddleBottomDeltaY(cell, val)
				if deltaY > 0 {
					if c.IsPrint() && cell.IsPageBreaksChange(c, deltaY) {
						oldCellHeight := cell.GetHeight()
						t.relayoutCell(c, cell, deltaY)
						if oldCellHeight+deltaY != cell.GetHeight() {
							needRowHeightRecalc = true
						}
					} else {
						cell.MoveContent(deltaY)
						// Set a provisional height in case we need to calculate
						// a default baseline
						cell.SetHeight(cell.GetHeight() + deltaY)
					}
				}
			}
		}
	}

	return needRowHeightRecalc
}

func (t *TableRowBox) calcMiddleBottomDeltaY(cell *TableCellBox, verticalAlign *IdentValue) int {
	var result int
	if cell.GetStyle().GetRowSpan() == 1 {
		result = t.GetHeight() - cell.GetChildrenHeight()
	} else {
		result = t.GetAbsY() + t.GetHeight() - (cell.GetAbsY() + cell.GetChildrenHeight())
	}

	if verticalAlign == IdentValueMiddle {
		return result / 2
	} else { /* verticalAlign == IdentValue.BOTTOM */
		return result
	}
}

func (t *TableRowBox) CalcLayoutHeight(
	c *LayoutContext, border *BorderPropertySet,
	margin RectPropertySetI, padding RectPropertySetI) {
	if t.GetHeightOverride() > 0 {
		t.SetHeight(t.GetHeightOverride())
	}

	t.alignBaselineAlignedCells(c)

	t.calcRowHeight(c)

	recalcRowHeight := t.alignMiddleAndBottomAlignedCells(c)

	if recalcRowHeight {
		t.calcRowHeight(c)
	}

	if !t.IsHaveBaseline() {
		t.calcDefaultBaseline(c)
	}

	t.setCellHeights()
}

func (t *TableRowBox) calcRowHeight(c CssContext) {
	y1 := t.GetAbsY()
	var y2 int

	if t.GetHeight() != 0 {
		y2 = y1 + t.GetHeight()
	} else {
		y2 = y1
	}

	if t.isLastRow() {
		bottom := t.getTable().CalcFixedHeightRowBottom(c)
		if bottom > 0 && bottom > y2 {
			y2 = bottom
		}
	}

	cRow := t.GetIndex()
	totalRows := t.getSection().NumRows()
	grid := t.getSection().GetGrid()
	if len(grid) != 0 && cRow < len(grid) {
		row := grid[cRow].GetRow()
		for cCol := 0; cCol < len(row); cCol++ {
			cell := row[cCol]

			if cell == nil || cell == TableCellBoxSpanningCell {
				continue
			}
			if cRow < totalRows-1 && t.getSection().CellAt(cRow+1, cCol) == cell {
				continue
			}

			bottomCellEdge := cell.GetAbsY() + cell.GetHeight()
			if bottomCellEdge > y2 {
				y2 = bottomCellEdge
			}
		}
	}

	t.SetHeight(y2 - y1)
}

func (t *TableRowBox) isLastRow() bool {
	table := t.getTable()
	section := t.getSection()
	if table.SectionBelow(section, true) == nil {
		return section.GetChild(section.GetChildCount()-1) == BoxI(t)
	} else {
		return false
	}
}

func (t *TableRowBox) calcDefaultBaseline(c *LayoutContext) {
	lowestCellEdge := 0
	cRow := t.GetIndex()
	totalRows := t.getSection().NumRows()
	grid := t.getSection().GetGrid()
	if len(grid) != 0 && cRow < len(grid) {
		row := grid[cRow].GetRow()
		for cCol := 0; cCol < len(row); cCol++ {
			cell := row[cCol]

			if cell == nil || cell == TableCellBoxSpanningCell {
				continue
			}
			if cRow < totalRows-1 && t.getSection().CellAt(cRow+1, cCol) == cell {
				continue
			}

			contentArea := cell.GetContentAreaEdge(cell.GetAbsX(), cell.GetAbsY(), c)
			bottomCellEdge := contentArea.Y + contentArea.Height
			if bottomCellEdge > lowestCellEdge {
				lowestCellEdge = bottomCellEdge
			}
		}
	}
	if lowestCellEdge > 0 {
		t.SetBaseline(lowestCellEdge - t.GetAbsY())
	}
	t.SetHaveBaseline(true)
}

func (t *TableRowBox) setCellHeights() {
	cRow := t.GetIndex()
	totalRows := t.getSection().NumRows()
	grid := t.getSection().GetGrid()
	if len(grid) != 0 && cRow < len(grid) {
		row := grid[cRow].GetRow()
		for cCol := 0; cCol < len(row); cCol++ {
			cell := row[cCol]

			if cell == nil || cell == TableCellBoxSpanningCell {
				continue
			}
			if cRow < totalRows-1 && t.getSection().CellAt(cRow+1, cCol) == cell {
				continue
			}

			if cell.GetStyle().GetRowSpan() == 1 {
				cell.SetHeight(t.GetHeight())
			} else {
				cell.SetHeight(t.GetAbsY() + t.GetHeight() - cell.GetAbsY())
			}
		}
	}
}

func (t *TableRowBox) relayoutCell(c *LayoutContext, cell *TableCellBox, contentStart int) {
	width := cell.GetWidth()
	cell.Reset(c)
	cell.SetLayoutWidth(c, width)
	t.layoutCell(c, cell, contentStart)
}

func (t *TableRowBox) layoutCell(c *LayoutContext, cell *TableCellBox, contentStart int) {
	cell.InitContainingLayer(c)
	cell.CalcCanvasLocation()

	cell.LayoutWithContentStart(c, contentStart)
}

func (t *TableRowBox) InitStaticPos(c *LayoutContext, parent BlockBoxI, childOffset int) {
	t.SetX(0)

	table := t.getTable()
	t.SetY(parent.GetHeight() + table.GetStyle().GetBorderVSpacing(c))
	c.Translate(0, t.GetY()-childOffset)
}

func (t *TableRowBox) GetBaseline() int {
	return t.baseline
}

func (t *TableRowBox) SetBaseline(baseline int) {
	t.baseline = baseline
}

func (t *TableRowBox) IsSkipWhenCollapsingMargins() bool {
	return true
}

func (t *TableRowBox) PaintBorder(c *RenderingContext) {
	// rows never have borders
}

func (t *TableRowBox) PaintBackground(c *RenderingContext) {
	// painted at the cell level
}

func (t *TableRowBox) Reset(c *LayoutContext) {
	t.BlockBox.Reset(c)
	t.SetHaveBaseline(false)
	t.getSection().SetNeedCellWidthCalc(true)
	t.SetContentLimitContainer(nil)
}

func (t *TableRowBox) IsHaveBaseline() bool {
	return t.haveBaseline
}

func (t *TableRowBox) SetHaveBaseline(haveBaseline bool) {
	t.haveBaseline = haveBaseline
}

func (t *TableRowBox) GetExtraBoxDescription() string {
	if t.IsHaveBaseline() {
		return "(baseline=" + strconv.Itoa(t.GetBaseline()) + ") "
	} else {
		return ""
	}
}

func (t *TableRowBox) GetHeightOverride() int {
	return t.heightOverride
}

func (t *TableRowBox) SetHeightOverride(heightOverride int) {
	t.heightOverride = heightOverride
}

func (t *TableRowBox) ExportText(c *RenderingContext, writer io.Writer) error {
	if t.getTable().IsMarginAreaRoot() {
		return t.BlockBox.ExportText(c, writer)
	} else {
		yPos := t.GetAbsY()
		if yPos >= c.GetPage().GetBottom() && t.IsInDocumentFlow() {
			if err := t.ExportPageBoxTextWithYPos(c, writer, yPos); err != nil {
				return err
			}
		}

		for _, box := range t.GetChildren() {
			cell := box.(*TableCellBox)
			buffer := &strings.Builder{}
			if err := cell.CollectText(c, buffer); err != nil {
				return err
			}
			// String.trim() removes leading and trailing characters up to U+0020.
			trimmed := strings.TrimFunc(buffer.String(), func(r rune) bool { return r <= ' ' })
			if _, err := io.WriteString(writer, trimmed); err != nil {
				return err
			}
			cSpan := cell.GetStyle().GetColSpan()
			for j := 0; j < cSpan; j++ {
				if _, err := io.WriteString(writer, "\t"); err != nil {
					return err
				}
			}
		}

		// System.lineSeparator()
		if _, err := io.WriteString(writer, "\n"); err != nil {
			return err
		}
		return nil
	}
}

// GetContentLimitContainer may return nil.
func (t *TableRowBox) GetContentLimitContainer() *ContentLimitContainer {
	return t.contentLimitContainer
}

func (t *TableRowBox) SetContentLimitContainer(contentLimitContainer *ContentLimitContainer) {
	t.contentLimitContainer = contentLimitContainer
}

func (t *TableRowBox) GetExtraSpaceTop() int {
	return t.extraSpaceTop
}

func (t *TableRowBox) SetExtraSpaceTop(extraSpaceTop int) {
	t.extraSpaceTop = extraSpaceTop
}

func (t *TableRowBox) GetExtraSpaceBottom() int {
	return t.extraSpaceBottom
}

func (t *TableRowBox) SetExtraSpaceBottom(extraSpaceBottom int) {
	t.extraSpaceBottom = extraSpaceBottom
}

func (t *TableRowBox) ForcePageBreakBefore(c *LayoutContext, pageBreakValue *IdentValue,
	pendingPageName bool) int {
	currentDelta := t.BlockBox.ForcePageBreakBefore(c, pageBreakValue, pendingPageName)

	// additional calculations for collapsed borders.
	if c.IsPrint() && t.GetStyle().IsCollapseBorders() {
		// get destination page for this row
		page := c.GetRootLayer().GetPage(c, t.GetAbsY()+currentDelta)
		if page != nil {

			// calculate max spill from the collapsed top borders of each child
			spill := 0
			for _, box := range t.GetChildren() {
				cell := box.(*TableCellBox)
				collapsed := cell.GetCollapsedPaintingBorder()
				if collapsed != nil {
					spill = max(spill, int(collapsed.Top())/2)
				}
			}

			// be sure that the current start of the row is >= the start of the page
			borderTop := t.GetAbsY() + currentDelta + int(t.GetMargin(c).Top()) - spill
			rowDelta := page.GetTop() - borderTop
			if rowDelta > 0 {
				t.SetY(t.GetY() + rowDelta)
				currentDelta += rowDelta
			}
		}
	}
	return currentDelta
}
