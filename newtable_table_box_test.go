// Tests of the port of flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/TableBox.java

package ufo

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// Flying Saucer has no JUnit test of TableBox in flying-saucer-core. The test
// in this file compares the column width calculation of the three table layout
// strategies with values recorded from the Java implementation.
//
// testdata/newtable/table_box_layout_cases.json was produced by laying out
// randomly generated HTML tables with the Java ITextRenderer and then, for
// every TableBox in the box tree, calling reset(), calcMinMaxWidth() and
// layout() of its TableLayout (and of a MarginTableLayout for tables with
// three effective columns) and recording the inputs the strategy reads (the
// effective columns, the column elements, the cell grid, the min/max width
// and style width of each cell, the table's widths and spacing) and its
// results (the table's min/max width, the column positions and the
// AutoTableLayout.Layout entries). Lengths are [type, value] with type 0 =
// variable, 1 = fixed, 2 = percent. In a grid, -1 is an empty position and -2
// is TableCellBox.SPANNING_CELL.

type tableBoxTestLength []int64

func (l tableBoxTestLength) length() *Length {
	return NewLength(l[1], LengthLengthType(l[0]))
}

type tableBoxTestCase struct {
	Layout    string
	Spans     []int
	StyleCols []struct {
		Span int
		W    tableBoxTestLength
		Pw   tableBoxTestLength
	}
	Sections []struct {
		Rows [][]int
		Grid [][]int
	}
	Cells []struct {
		ColSpan, Col, Min, Max int
		Osw, Osocw             tableBoxTestLength
	}
	Hspacing, BsTrue, BsFalse, Width, ContentWidth int
	Tw                                             tableBoxTestLength
	MinWidth, MaxWidth                             int
	ColumnPos                                      []int
	LayoutStruct                                   [][]json.RawMessage
}

// tableBoxTestStyle answers the style questions the table layout strategies
// ask. Every other method of CalculatedStyleI panics (nil embedded interface).
type tableBoxTestStyle struct {
	CalculatedStyleI

	colSpan         int
	width           *Length
	hspacing        int
	tableLayoutAuto bool
	marginLeft      float32
	marginRight     float32
}

func (s *tableBoxTestStyle) IsIdent(cssName *CSSName, val *IdentValue) bool {
	return cssName == CSSNameTableLayout && val == IdentValueAuto && s.tableLayoutAuto
}

func (s *tableBoxTestStyle) IsAutoWidth() bool                  { return false }
func (s *tableBoxTestStyle) IsCollapseBorders() bool            { return false }
func (s *tableBoxTestStyle) IsAutoLeftMargin() bool             { return false }
func (s *tableBoxTestStyle) IsAutoRightMargin() bool            { return true }
func (s *tableBoxTestStyle) IsBorderBox() bool                  { return true }
func (s *tableBoxTestStyle) GetColSpan() int                    { return s.colSpan }
func (s *tableBoxTestStyle) GetRowSpan() int                    { return 1 }
func (s *tableBoxTestStyle) GetBorderHSpacing(c CssContext) int { return s.hspacing }

// GetFloatPropertyProportionalWidth is called by TableBox.GetCSSWidth, whose
// result BlockBox.CalcDimensions ignores once the dimensions are calculated.
func (s *tableBoxTestStyle) GetFloatPropertyProportionalWidth(cssName *CSSName, parentWidth float32, ctx CssContext) float32 {
	return 0
}

func (s *tableBoxTestStyle) AsLength(c CssContext, cssName *CSSName) *Length {
	return s.width
}

func (s *tableBoxTestStyle) GetBorder(ctx CssContext) *BorderPropertySet {
	return BorderPropertySetEmptyBorder
}

func (s *tableBoxTestStyle) GetMarginRect(cbWidth float32, ctx CssContext) RectPropertySetI {
	return NewRectPropertySet(0, s.marginRight, 0, s.marginLeft)
}

func (s *tableBoxTestStyle) GetMarginRectWithUseCache(cbWidth float32, ctx CssContext, useCache bool) RectPropertySetI {
	return NewRectPropertySet(0, s.marginRight, 0, s.marginLeft)
}

func (s *tableBoxTestStyle) GetPaddingRect(cbWidth float32, ctx CssContext) RectPropertySetI {
	return NewRectPropertySet(0, 0, 0, 0)
}

// tableBoxTestBuild builds the box tree of a recorded case: a table in a
// containing block, with the recorded effective columns and cell grid.
func tableBoxTestBuild(t *testing.T, tc *tableBoxTestCase) *TableBox {
	nEffCols := len(tc.Spans)
	spacing := (nEffCols + 1) * tc.Hspacing
	// marginsBordersPaddingAndSpacing(c, true) leaves out the auto right
	// margin; with false it includes both margins.
	style := &tableBoxTestStyle{
		hspacing:        tc.Hspacing,
		width:           tc.Tw.length(),
		tableLayoutAuto: tc.Layout != "fixed",
		marginLeft:      float32(tc.BsTrue - spacing),
		marginRight:     float32(tc.BsFalse - tc.BsTrue),
	}
	table := NewTableBox(nil, style, false)
	if tc.Layout == "margin" {
		table.SetMarginAreaRoot(true)
		table.SetStyle(style)
	}

	container := NewTableCellBox(nil, &tableBoxTestStyle{}, false)
	container.AddChild(table)

	for _, span := range tc.Spans {
		data := NewColumnData()
		data.SetSpan(span)
		table.columns = append(table.columns, data)
	}
	for _, sc := range tc.StyleCols {
		col := NewTableColumnWithElementStyle(nil, &tableBoxTestStyle{colSpan: sc.Span, width: sc.W.length()})
		if sc.Pw != nil {
			col.SetParent(NewTableColumnWithElementStyle(nil, &tableBoxTestStyle{width: sc.Pw.length()}))
		}
		table.AddStyleColumn(col)
	}

	cells := make([]*TableCellBox, len(tc.Cells))
	for i, c := range tc.Cells {
		cell := NewTableCellBox(nil, &tableBoxTestStyle{colSpan: c.ColSpan, width: c.Osw.length()}, false)
		cell.SetCol(c.Col)
		cell.SetMinWidth(c.Min)
		cell.SetMaxWidth(c.Max)
		cell.SetMinMaxCalculated(true)
		cells[i] = cell
	}
	for _, s := range tc.Sections {
		section := NewTableSectionBox(nil, &tableBoxTestStyle{}, false)
		table.AddChild(section)
		for _, r := range s.Rows {
			row := NewTableRowBox(nil, &tableBoxTestStyle{}, false)
			section.AddChild(row)
			for _, ci := range r {
				row.AddChild(cells[ci])
			}
		}
		for _, g := range s.Grid {
			rowData := NewRowData()
			rowData.ExtendToColumnCount(len(g))
			for col, ci := range g {
				switch ci {
				case -1:
				case -2:
					rowData.GetRow()[col] = TableCellBoxSpanningCell
				default:
					rowData.GetRow()[col] = cells[ci]
				}
			}
			section.grid = append(section.grid, rowData)
		}
	}

	// The state BlockBox.calcDimensions leaves behind.
	table.SetContentWidth(tc.ContentWidth)
	table.SetLeftMBP(int(style.marginLeft))
	table.SetRightMBP(int(style.marginRight))
	table.SetDimensionsCalculated(true)

	if got := table.GetWidth(); got != tc.Width {
		t.Fatalf("test setup: table width %d, want %d", got, tc.Width)
	}
	for i, c := range tc.Cells {
		if got := cells[i].GetOuterStyleOrColWidth(nil).String(); got != c.Osocw.length().String() {
			t.Fatalf("test setup: cell %d outer style or col width %s, want %s", i, got, c.Osocw.length())
		}
	}
	return table
}

func tableBoxTestLayoutStruct(layoutStruct []*TableBoxAutoTableLayoutLayout) []string {
	var result []string
	for _, l := range layoutStruct {
		result = append(result, fmt.Sprintf("[[%d,%d],[%d,%d],%d,%d,%d,%d,%d]",
			l.Width().Type(), l.Width().Value(), l.EffWidth().Type(), l.EffWidth().Value(),
			l.MinWidth(), l.MaxWidth(), l.EffMinWidth(), l.EffMaxWidth(), l.CalcWidth()))
	}
	return result
}

func TestTableBoxLayoutStrategiesMatchJava(t *testing.T) {
	data, err := os.ReadFile("testdata/newtable/table_box_layout_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []*tableBoxTestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases")
	}
	c := &LayoutContext{}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.Layout), func(t *testing.T) {
			table := tableBoxTestBuild(t, tc)

			var layoutStruct func() []*TableBoxAutoTableLayoutLayout
			switch tl := table.tableLayout.(type) {
			case *TableBoxAutoTableLayout:
				if tc.Layout != "auto" {
					t.Fatalf("SetStyle chose AutoTableLayout, want %s", tc.Layout)
				}
				layoutStruct = tl.GetLayoutStruct
			case *TableBoxFixedTableLayout:
				if tc.Layout != "fixed" {
					t.Fatalf("SetStyle chose FixedTableLayout, want %s", tc.Layout)
				}
			case *TableBoxMarginTableLayout:
				if tc.Layout != "margin" {
					t.Fatalf("SetStyle chose MarginTableLayout, want %s", tc.Layout)
				}
				layoutStruct = tl.GetLayoutStruct
			}

			table.tableLayout.Reset()
			table.tableLayout.CalcMinMaxWidth(c)
			if table.GetMinWidth() != tc.MinWidth || table.GetMaxWidth() != tc.MaxWidth {
				t.Errorf("min/max width %d/%d, want %d/%d", table.GetMinWidth(), table.GetMaxWidth(), tc.MinWidth, tc.MaxWidth)
			}

			table.tableLayout.Layout(c)
			if got := table.GetColumnPos(); !reflect.DeepEqual(got, tc.ColumnPos) {
				t.Errorf("column positions %v, want %v", got, tc.ColumnPos)
			}

			if layoutStruct != nil {
				var want []string
				for _, l := range tc.LayoutStruct {
					b, err := json.Marshal(l)
					if err != nil {
						t.Fatal(err)
					}
					want = append(want, string(b))
				}
				if got := tableBoxTestLayoutStruct(layoutStruct()); !reflect.DeepEqual(got, want) {
					t.Errorf("layout struct\n got %v\nwant %v", got, want)
				}
			}
		})
	}
}

func TestTableBoxEffectiveColumns(t *testing.T) {
	table := NewTableBox(nil, &tableBoxTestStyle{tableLayoutAuto: true}, false)
	table.AppendColumn(1)
	table.AppendColumn(3)
	table.AppendColumn(2)

	if got := table.NumEffCols(); got != 3 {
		t.Errorf("NumEffCols() = %d, want 3", got)
	}
	for col, want := range []int{0, 1, 2, 2, 2, 3, 3} {
		if got := table.ColToEffCol(col); got != want {
			t.Errorf("ColToEffCol(%d) = %d, want %d", col, got, want)
		}
	}
	for effCol, want := range []int{0, 1, 4, 6} {
		if got := table.EffColToCol(effCol); got != want {
			t.Errorf("EffColToCol(%d) = %d, want %d", effCol, got, want)
		}
	}

	table.SplitColumn(1, 2)
	var spans []int
	for _, col := range table.GetColumns() {
		spans = append(spans, col.GetSpan())
	}
	if want := []int{1, 2, 1, 2}; !reflect.DeepEqual(spans, want) {
		t.Errorf("spans after SplitColumn(1, 2) = %v, want %v", spans, want)
	}
}

func TestTableBoxColElement(t *testing.T) {
	table := NewTableBox(nil, &tableBoxTestStyle{tableLayoutAuto: true}, false)
	if got := table.ColElement(0); got != nil {
		t.Errorf("ColElement(0) without column elements = %v, want nil", got)
	}
	first := NewTableColumnWithElementStyle(nil, &tableBoxTestStyle{colSpan: 2})
	second := NewTableColumnWithElementStyle(nil, &tableBoxTestStyle{colSpan: 1})
	table.AddStyleColumn(first)
	table.AddStyleColumn(second)
	for col, want := range []*TableColumn{first, first, second, nil} {
		if got := table.ColElement(col); got != want {
			t.Errorf("ColElement(%d) = %p, want %p", col, got, want)
		}
	}
}

func TestRowDataSplitColumn(t *testing.T) {
	cell := NewTableCellBox(nil, &tableBoxTestStyle{}, false)
	row := NewRowData()
	row.ExtendToColumnCount(3)
	row.GetRow()[0] = cell
	row.SplitColumn(0)
	row.SplitColumn(2)
	want := []*TableCellBox{cell, TableCellBoxSpanningCell, nil, nil, nil}
	got := row.GetRow()
	if len(got) != len(want) {
		t.Fatalf("row length after splits = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row[%d] after splits = %p, want %p", i, got[i], want[i])
		}
	}
}
