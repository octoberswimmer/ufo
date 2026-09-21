// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/bug/PercentageMaxWidthTableCellTest.java

package pdf

import (
	"math"
	"testing"

	"github.com/octoberswimmer/ufo"
)

// Reproducible example for
// https://github.com/flyingsaucerproject/flyingsaucer/issues/697: a
// percentage max-width on a table cell used to resolve against a containing
// block (the table row) whose content width had not been established yet, so
// it was always clamped to zero.
func TestPercentageMaxWidthTableCell_percentageMaxWidthResolvesAgainstTheTableWidth(t *testing.T) {
	page := `<html><head><style>
  table { width: 600pt; }
</style></head>
<body>
<table><tr><td style="max-width: 60pt">fixed length</td></tr></table>
<table><tr><td style="max-width: 10%">percentage</td></tr></table>
</body></html>`

	renderer := NewITextRenderer()
	if err := renderer.SetDocumentFromString(page); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Layout(); err != nil {
		t.Fatal(err)
	}

	root := renderer.GetRootBox()
	isCell := func(b ufo.BoxI) bool { _, ok := b.(*ufo.TableCellBox); return ok }
	fixedLengthCell := percentageMaxWidthTableCellTestFindFirst(root, isCell).(*ufo.TableCellBox)
	percentageCell := percentageMaxWidthTableCellTestFindLast(root, isCell).(*ufo.TableCellBox)

	// Not byte-exact: the table layout algorithm allocates border-spacing per
	// column, so a percentage of the table's overall width is not
	// pixel-identical to an equivalent fixed length. It must, however, be in
	// the same ballpark, not resolve to (near) zero.
	fixed := float64(fixedLengthCell.GetContentWidth())
	percentage := float64(percentageCell.GetContentWidth())
	if math.Abs(percentage-fixed) > fixed*0.05 {
		t.Errorf("percentage cell content width %v is not within 5%% of %v", percentage, fixed)
	}
}

func percentageMaxWidthTableCellTestFindFirst(box ufo.BoxI, match func(ufo.BoxI) bool) ufo.BoxI {
	if match(box) {
		return box
	}
	for i := 0; i < box.GetChildCount(); i++ {
		if found := percentageMaxWidthTableCellTestFindFirst(box.GetChild(i), match); found != nil {
			return found
		}
	}
	return nil
}

func percentageMaxWidthTableCellTestFindLast(box ufo.BoxI, match func(ufo.BoxI) bool) ufo.BoxI {
	var last ufo.BoxI
	if match(box) {
		last = box
	}
	for i := 0; i < box.GetChildCount(); i++ {
		if found := percentageMaxWidthTableCellTestFindLast(box.GetChild(i), match); found != nil {
			last = found
		}
	}
	return last
}
