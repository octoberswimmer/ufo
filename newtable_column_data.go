// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/ColumnData.java

package ufo

// ColumnData instances are effective columns in the table grid. Table
// columns are managed in terms of effective columns which ensures that only the
// minimum number of columns necessary to manage the grid are created. For
// example, a table cell with colspan="1000" will only create a single effective
// column unless there are other table cells in other rows which force the
// column to be split.
type ColumnData struct {
	span int
}

func NewColumnData() *ColumnData {
	return &ColumnData{span: 1}
}

func (d *ColumnData) GetSpan() int {
	return d.span
}

func (d *ColumnData) SetSpan(span int) {
	d.span = span
}
