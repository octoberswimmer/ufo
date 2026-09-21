// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/CollapsedBorderSide.java

package ufo

// CollapsedBorderSide contains a single border side of a collapsed cell.
// Collapsed border sides are painted in order of priority (so for example,
// wider borders always paint over narrower borders regardless of the relative
// tree order of the cells in question).
//
// Java's hashCode() is not ported: it is derived from the cell's identity
// hash, and nothing stores a CollapsedBorderSide in a hashed collection.
type CollapsedBorderSide struct {
	cell *TableCellBox
	side int
}

func NewCollapsedBorderSide(cell *TableCellBox, side int) *CollapsedBorderSide {
	return &CollapsedBorderSide{side: side, cell: cell}
}

func (s *CollapsedBorderSide) GetCell() *TableCellBox {
	return s.cell
}

func (s *CollapsedBorderSide) GetSide() int {
	return s.side
}

func (s *CollapsedBorderSide) CompareTo(that *CollapsedBorderSide) int {
	v1 := collapsedBorderSideGetCollapsedBorder(s)
	v2 := collapsedBorderSideGetCollapsedBorder(that)
	result := TableCellBoxCompareBordersWithReturnNullOnEqual(v1, v2, true)

	if result == nil {
		return 0
	} else if result == v1 {
		return 1
	} else {
		return -1
	}
}

// collapsedBorderSideGetCollapsedBorder may return nil.
func collapsedBorderSideGetCollapsedBorder(c1 *CollapsedBorderSide) *CollapsedBorderValue {
	switch c1.side {
	case BorderPainterTop:
		return c1.cell.GetCollapsedBorderTop()
	case BorderPainterRight:
		return c1.cell.GetCollapsedBorderRight()
	case BorderPainterBottom:
		return c1.cell.GetCollapsedBorderBottom()
	case BorderPainterLeft:
		return c1.cell.GetCollapsedBorderLeft()
	default:
		return nil
	}
}

// Equals compares the side and the identity of the cell (TableCellBox does
// not override equals).
func (s *CollapsedBorderSide) Equals(o any) bool {
	that, ok := o.(*CollapsedBorderSide)
	if !ok || that == nil {
		return false
	}
	if s == that {
		return true
	}

	return s.side == that.side && s.cell == that.cell
}
