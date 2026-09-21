package ufo

import (
	"os"
	"strings"
	"testing"
)

// TestTableCellBoxCompareBorders compares
// TableCellBoxCompareBordersWithReturnNullOnEqual with the results that
// Java's TableCellBox.compareBorders gave for the same arguments. The expected
// file holds one character per call, in the order of the loops below: '1' when
// the first border was returned, '2' for the second, 'n' for null.
func TestTableCellBoxCompareBorders(t *testing.T) {
	data, err := os.ReadFile("testdata/newtable/table_cell_box_compare_borders_jdk.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(string(data))

	styles := []*IdentValue{nil, IdentValueNone, IdentValueHidden, IdentValueDouble, IdentValueSolid,
		IdentValueDashed, IdentValueDotted, IdentValueRidge, IdentValueOutset, IdentValueGroove, IdentValueInset}
	widths := []int{1, 2}
	precedences := []int{6, 9, 10}

	i := 0
	for _, returnNullOnEqual := range []bool{false, true} {
		for _, s1 := range styles {
			for _, w1 := range widths {
				for _, p1 := range precedences {
					for _, s2 := range styles {
						for _, w2 := range widths {
							for _, p2 := range precedences {
								b1 := NewCollapsedBorderValue(s1, w1, nil, p1)
								b2 := NewCollapsedBorderValue(s2, w2, nil, p2)
								r := TableCellBoxCompareBordersWithReturnNullOnEqual(b1, b2, returnNullOnEqual)
								var got byte
								switch r {
								case nil:
									got = 'n'
								case b1:
									got = '1'
								case b2:
									got = '2'
								default:
									got = '?'
								}
								if i >= len(expected) {
									t.Fatalf("expected file has %d results, need more", len(expected))
								}
								if got != expected[i] {
									t.Errorf("call %d (returnNullOnEqual=%v, %v/%d/%d, %v/%d/%d): got %c, want %c",
										i, returnNullOnEqual, s1, w1, p1, s2, w2, p2, got, expected[i])
								}
								i++
							}
						}
					}
				}
			}
		}
	}
	if i != len(expected) {
		t.Errorf("made %d calls, expected file has %d results", i, len(expected))
	}
}
