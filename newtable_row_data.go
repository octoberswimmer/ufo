// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/RowData.java

package ufo

// RowData is a row in the table grid.  The list of cells it maintains is always large
// enough to set a table cell at every position in the row.  If there are no
// colspans, rowspans, or missing cells, the grid row will exactly correspond
// to the row in the original markup.  On the other hand, colspans may force
// spanning cells to be inserted, rowspans will mean cells appear in more than
// one grid row, and positions may be nil if no cell occupies that
// position in the grid.
type RowData struct {
	row []*TableCellBox
}

func NewRowData() *RowData {
	return &RowData{}
}

// GetRow returns the cells of the row. Java returns the mutable list, which
// TableSectionBox changes with set(); the returned slice shares its backing
// array with the row, so an assignment to an existing position is seen by the
// row. No Java caller changes the length of the list.
func (d *RowData) GetRow() []*TableCellBox {
	return d.row
}

func (d *RowData) ExtendToColumnCount(columnCount int) {
	for len(d.row) < columnCount {
		d.row = append(d.row, nil)
	}
}

func (d *RowData) SplitColumn(pos int) {
	current := d.row[pos]
	var inserted *TableCellBox
	if current != nil {
		inserted = TableCellBoxSpanningCell
	}
	d.row = append(d.row, nil)
	copy(d.row[pos+2:], d.row[pos+1:])
	d.row[pos+1] = inserted
}
