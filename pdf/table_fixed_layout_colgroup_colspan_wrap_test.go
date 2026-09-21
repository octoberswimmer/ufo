// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/TableFixedLayoutColgroupColspanWrapTest.java

package pdf

import "testing"

// Regression tests for table-layout:fixed + colgroup/col + colspan text
// wrapping.
//
// Original bug (PR #682): TableCellBox.setLayoutWidth() unconditionally
// called applyCSSMinMaxWidth(c), which honoured the <td>'s own
// max-width: 34% and shrank a colspan=4 cell from its col-allocated 67 %
// down to 34 %. Text then overflowed instead of wrapping within the correct
// cell width.
//
// The guard introduced by the follow-up PR used isBorderBox(), which causes
// two distinct regressions that this test class covers:
//
//  1. Content-box regression — in table-layout: auto tables, max-width is
//     silently ignored for content-box cells because isBorderBox() returns
//     false and the call is skipped.
//  2. Border-box regression — in table-layout: fixed tables, max-width still
//     overrides the col-allocated width for border-box cells because
//     isBorderBox() returns true and the call is still made.
//
// The correct guard is isFixedWidthAdvisoryOnly() (true only for
// table-layout: auto), which fixes both cases regardless of box-sizing.

// Three-test HTML: content-box (Test-1), no-max-width (Test-2), border-box (Test-3).
const tableFixedLayoutTestFixedHtml = "org/xhtmlrenderer/pdf/table-fixed-layout-colgroup-colspan-wrap.html"

// Single-cell auto-layout table with max-width on a content-box td.
const tableFixedLayoutTestAutoHtml = "org/xhtmlrenderer/pdf/table-auto-layout-content-box-max-width.html"

// 1. Smoke — must not throw

func TestTableFixedLayoutColgroupColspanWrap_renderDoesNotThrow(t *testing.T) {
	testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml)
}

// 2. All words present — no clipping / overflow loss

func TestTableFixedLayoutColgroupColspanWrap_allWordsInColspanCellArePresentInPdf(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml),
		"table-fixed-layout-colgroup-colspan-wrap.pdf")

	// Every Greek-letter name in the Test-1 VALUE cell must survive in the
	// PDF. If the bug is active the cell is shrunk to 34 % and overflow may
	// clip words.
	for _, word := range []string{"Alpha", "Epsilon", "Kappa", "Omicron", "Upsilon", "Omega", "Lorem", "aliqua"} {
		pdf.containsExactText(t, word)
	}
}

func TestTableFixedLayoutColgroupColspanWrap_keyColumnsArePresent(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml),
		"table-fixed-layout-colgroup-colspan-keys.pdf")

	pdf.containsExactText(t, "KEY")
	pdf.containsExactText(t, "KEY2")
	pdf.containsExactText(t, "KEY3")
}

// 3. Width detection — col-allocated 67 % vs buggy max-width 34 %

// Confirms that the Test-1 colspan cell uses the col-allocated 67 % width,
// not the max-width: 34% declared on the content-box <td>.
//
// The assertion is structural: "Eta" and "Theta" are consecutive words in the
// source text. At the correct 67 % cell width the first text line is wide
// enough to accommodate both words together. If max-width: 34% were
// incorrectly applied the narrower cell would wrap earlier, placing "Eta" at
// the start of a new line and separating the two words.
func TestTableFixedLayoutColgroupColspanWrap_etaAndThetaAreOnSameLineAtCorrectCellWidth(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml),
		"table-fixed-layout-colgroup-colspan-eta-theta.pdf")

	if !tableTestAnyLine(pdf.lines(), "Eta", "Theta") {
		t.Errorf("'Eta' and 'Theta' must appear on the same extracted-text line (Test-1, content-box cell):\n%s", pdf.text)
	}
}

// Guards against silent overflow: the Greek-alphabet text is long enough that
// it must wrap at any realistic cell width, so it must occupy more than one
// line. This ensures the previous test is not vacuously satisfied by a
// non-wrapping overflow.
func TestTableFixedLayoutColgroupColspanWrap_valueTextWrapsAcrossMultipleLinesAtCorrectCellWidth(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml),
		"table-fixed-layout-colgroup-colspan-wrap-lines.pdf")

	// 24 Greek-letter words cannot fit on a single line at any realistic
	// width. Assert structurally that "Alpha" (first) and "Omega" (last) land
	// on different extracted-text lines.
	pdf.containsExactText(t, "Alpha") // Test-1-only: Test-3 starts with "Alfa"
	pdf.containsExactText(t, "Omega")
	if tableTestAnyLine(pdf.lines(), "Alpha", "Omega") {
		t.Errorf("'Alpha' and 'Omega' must be on different lines (Test-1 only):\n%s", pdf.text)
	}
}

// Same wrapping check for Test-2 (no max-width on td, plain colspan).
func TestTableFixedLayoutColgroupColspanWrap_colspanCellWithoutMaxWidthAlsoWraps(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml),
		"table-fixed-layout-colgroup-colspan-nowrap.pdf")

	pdf.containsExactText(t, "Lorem")
	pdf.containsExactText(t, "aliqua")
	if tableTestAnyLine(pdf.lines(), "Lorem", "aliqua") {
		t.Errorf("'Lorem' and 'aliqua' must be on different lines:\n%s", pdf.text)
	}
}

// 4. Regression: border-box cell in table-layout:fixed
//
// The isBorderBox() guard in the PR would call applyCSSMinMaxWidth for any
// border-box cell, including those in table-layout:fixed tables. The
// col-allocated width must still win for border-box cells in fixed layout —
// exactly the same requirement as for content-box cells.

// A border-box colspan <td> in a table-layout: fixed table must use the
// col-allocated width (67 %), not the max-width: 34% declared on the element.
//
// Test-3 in the HTML fixture has box-sizing: border-box on the same colspan=4
// cell. Under the buggy isBorderBox() guard applyCSSMinMaxWidth is called and
// shrinks the cell to the incorrect 34 %. Under the correct
// isFixedWidthAdvisoryOnly() guard the call is skipped for
// table-layout: fixed regardless of box-sizing, so "Eta" and "Theta" appear
// on the same line.
func TestTableFixedLayoutColgroupColspanWrap_borderBoxFixedLayoutColWidthTakesPrecedenceOverMaxWidth(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestFixedHtml),
		"table-fixed-layout-colgroup-colspan-borderbox.pdf")

	pdf.containsExactText(t, "KEY3")
	if !tableTestAnyLine(pdf.lines(), "Ita", "Thita") {
		t.Errorf("'Ita' and 'Thita' must appear on the same extracted-text line (Test-3 only):\n%s", pdf.text)
	}
}

// 5. Regression: content-box cell in table-layout:auto
//
// The isBorderBox() guard skips applyCSSMinMaxWidth for content-box cells
// everywhere, including table-layout:auto tables where max-width must be
// honoured. This worked correctly before PR #682.

// A content-box <td> in a table-layout: auto table must have its max-width
// honoured by applyCSSMinMaxWidth.
//
// Under the buggy isBorderBox() guard the call is skipped for content-box
// cells (the default), so max-width: 80pt is silently ignored and the cell
// expands to its natural content width, fitting all 24 words on a single
// wide line. Under the correct isFixedWidthAdvisoryOnly() guard the call IS
// made (auto-layout → widths are advisory), constraining the cell to 80 pt
// and forcing the words to wrap across many lines — so "Alpha" and "Omega"
// land on different lines.
func TestTableFixedLayoutColgroupColspanWrap_contentBoxAutoLayoutMaxWidthIsRespected(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableFixedLayoutTestAutoHtml),
		"table-auto-layout-content-box-max-width.pdf")

	// At max-width: 80pt the 24 Greek words cannot fit on one line.
	// "Alpha" (first word) and "Omega" (last word) must be on different lines.
	pdf.containsExactText(t, "Alpha")
	pdf.containsExactText(t, "Omega")
	if tableTestAnyLine(pdf.lines(), "Alpha", "Omega") {
		t.Errorf("'Alpha' and 'Omega' must be on different extracted-text lines:\n%s", pdf.text)
	}
}
