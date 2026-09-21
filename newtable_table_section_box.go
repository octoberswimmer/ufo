// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/TableSectionBox.java

package ufo

import (
	"github.com/octoberswimmer/ufo/dom"
)

type TableSectionBox struct {
	BlockBox

	grid []*RowData

	needCellWidthCalc bool
	needCellRecalc    bool

	footer bool
	header bool

	capturedOriginalAbsY bool
	originalAbsY         int
}

func NewTableSectionBox(element *dom.Element, style CalculatedStyleI, anonymous bool) *TableSectionBox {
	t := &TableSectionBox{}
	initBlockBox(&t.BlockBox, element, style, anonymous)
	t.SetSelf(t)
	return t
}

func (t *TableSectionBox) CopyOf() BlockBoxI {
	return NewTableSectionBox(t.GetElement(), t.GetStyle(), t.IsAnonymous())
}

// GetGrid returns the rows of the grid. No Java caller changes the returned
// list; the slice is replaced when rows are added or the grid is cleared, so a
// caller must not keep it across RecalcCells or Reset.
func (t *TableSectionBox) GetGrid() []*RowData {
	return t.grid
}

func (t *TableSectionBox) ExtendGridToColumnCount(columnCount int) {
	for _, row := range t.grid {
		row.ExtendToColumnCount(columnCount)
	}
}

func (t *TableSectionBox) SplitColumn(pos int) {
	for _, row := range t.grid {
		row.SplitColumn(pos)
	}
}

func (t *TableSectionBox) RecalcCells(c *LayoutContext) {
	cRow := 0
	t.grid = nil
	t.EnsureChildren(c)
	for _, child := range t.GetChildren() {
		row := child.(*TableRowBox)
		row.EnsureChildren(c)
		for _, box := range row.GetChildren() {
			cell := box.(*TableCellBox)
			t.addCell(cell, cRow)
		}
		cRow++
	}
}

func (t *TableSectionBox) CalcBorders(c *LayoutContext) {
	t.EnsureChildren(c)
	for _, box := range t.GetChildren() {
		row := box.(*TableRowBox)
		row.EnsureChildren(c)
		for _, value := range row.GetChildren() {
			cell := value.(*TableCellBox)
			cell.CalcCollapsedBorder(c)
		}
	}
}

// CellAt may return nil.
func (t *TableSectionBox) CellAt(row int, col int) *TableCellBox {
	if row >= len(t.grid) {
		return nil
	}
	rowData := t.grid[row]
	if col >= len(rowData.GetRow()) {
		return nil
	}
	return rowData.GetRow()[col]
}

func (t *TableSectionBox) setCellAt(row int, col int, cell *TableCellBox) {
	t.grid[row].GetRow()[col] = cell
}

func (t *TableSectionBox) ensureRows(numRows int) {
	nRows := len(t.grid)
	nCols := t.getTable().NumEffCols()

	for nRows < numRows {
		row := NewRowData()
		row.ExtendToColumnCount(nCols)
		t.grid = append(t.grid, row)
		nRows++
	}
}

func (t *TableSectionBox) getTable() *TableBox {
	return tableCellBoxCastTable(t.GetParent())
}

func (t *TableSectionBox) LayoutChildren(c *LayoutContext, contentStart int) {
	if t.isNeedCellRecalc() {
		t.RecalcCells(c)
		t.setNeedCellRecalc(false)
	}

	if t.IsNeedCellWidthCalc() {
		t.SetCellWidths(c)
		t.SetNeedCellWidthCalc(false)
	}

	t.BlockBox.LayoutChildren(c, contentStart)
}

func (t *TableSectionBox) addCell(cell *TableCellBox, cRow int) {
	rSpan := cell.GetStyle().GetRowSpan()
	cSpan := cell.GetStyle().GetColSpan()

	// Java holds the table's list of columns in a local variable and sees the
	// columns that appendColumn and splitColumn add to it. A Go slice taken
	// before those calls would not, so the list is read from the table again
	// at each use.
	nCols := len(t.getTable().GetColumns())
	cCol := 0

	t.ensureRows(cRow + rSpan)

	for cCol < nCols && t.CellAt(cRow, cCol) != nil {
		cCol++
	}

	col := cCol
	set := cell
	for cSpan > 0 {
		var currentSpan int
		for cCol >= len(t.getTable().GetColumns()) {
			t.getTable().AppendColumn(1)
		}
		cData := t.getTable().GetColumns()[cCol]
		if cSpan < cData.GetSpan() {
			t.getTable().SplitColumn(cCol, cSpan)
		}
		cData = t.getTable().GetColumns()[cCol]
		currentSpan = cData.GetSpan()

		r := 0
		for r < rSpan {
			if t.CellAt(cRow+r, cCol) == nil {
				t.setCellAt(cRow+r, cCol, set)
			}
			r++
		}
		cCol++
		cSpan -= currentSpan
		set = TableCellBoxSpanningCell
	}

	cell.SetRow(cRow)
	cell.SetCol(t.getTable().EffColToCol(col))
}

func (t *TableSectionBox) Reset(c *LayoutContext) {
	t.BlockBox.Reset(c)
	t.grid = nil
	t.SetNeedCellWidthCalc(true)
	t.setNeedCellRecalc(true)
	t.SetCapturedOriginalAbsY(false)
}

func (t *TableSectionBox) SetCellWidths(c *LayoutContext) {
	columnPos := t.getTable().GetColumnPos()

	for _, row := range t.grid {
		cols := row.GetRow()
		hspacing := t.getTable().GetStyle().GetBorderHSpacing(c)
		for j := 0; j < len(cols); j++ {
			cell := cols[j]

			if cell == nil || cell == TableCellBoxSpanningCell {
				continue
			}

			endCol := j
			cspan := cell.GetStyle().GetColSpan()
			for cspan > 0 && endCol < len(cols) {
				cspan -= t.getTable().SpanOfEffCol(endCol)
				endCol++
			}

			w := columnPos[endCol] - columnPos[j] - hspacing
			cell.SetLayoutWidth(c, w)
			cell.SetX(columnPos[j] + hspacing)
		}
	}
}

func (t *TableSectionBox) IsAutoHeight() bool {
	// FIXME Should properly handle absolute heights (%s resolve to auto)
	return true
}

func (t *TableSectionBox) NumRows() int {
	return len(t.grid)
}

func (t *TableSectionBox) IsSkipWhenCollapsingMargins() bool {
	return true
}

func (t *TableSectionBox) PaintBorder(c *RenderingContext) {
	// row groups never have borders
}

func (t *TableSectionBox) PaintBackground(c *RenderingContext) {
	// painted at the cell level
}

// GetLastRow may return nil.
func (t *TableSectionBox) GetLastRow() *TableRowBox {
	if t.GetChildCount() > 0 {
		return tableCellBoxCastRow(t.GetChild(t.GetChildCount() - 1))
	} else {
		return nil
	}
}

func (t *TableSectionBox) IsNeedCellWidthCalc() bool {
	return t.needCellWidthCalc
}

func (t *TableSectionBox) SetNeedCellWidthCalc(needCellWidthCalc bool) {
	t.needCellWidthCalc = needCellWidthCalc
}

func (t *TableSectionBox) isNeedCellRecalc() bool {
	return t.needCellRecalc
}

func (t *TableSectionBox) setNeedCellRecalc(needCellRecalc bool) {
	t.needCellRecalc = needCellRecalc
}

func (t *TableSectionBox) LayoutWithContentStart(c *LayoutContext, contentStart int) {
	running := c.IsPrint() && (t.IsHeader() || t.IsFooter()) && t.getTable().GetStyle().IsPaginateTable()

	if running {
		c.SetNoPageBreak(c.GetNoPageBreak() + 1)
	}

	t.BlockBox.LayoutWithContentStart(c, contentStart)

	if running {
		c.SetNoPageBreak(c.GetNoPageBreak() - 1)
	}
}

func (t *TableSectionBox) IsFooter() bool {
	return t.footer
}

func (t *TableSectionBox) SetFooter(footer bool) {
	t.footer = footer
}

func (t *TableSectionBox) IsHeader() bool {
	return t.header
}

func (t *TableSectionBox) SetHeader(header bool) {
	t.header = header
}

func (t *TableSectionBox) IsCapturedOriginalAbsY() bool {
	return t.capturedOriginalAbsY
}

func (t *TableSectionBox) SetCapturedOriginalAbsY(capturedOriginalAbsY bool) {
	t.capturedOriginalAbsY = capturedOriginalAbsY
}

func (t *TableSectionBox) GetOriginalAbsY() int {
	return t.originalAbsY
}

func (t *TableSectionBox) SetOriginalAbsY(originalAbsY int) {
	t.originalAbsY = originalAbsY
}
