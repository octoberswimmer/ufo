// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/TableAutoLayoutColspanMaxWidthTest.java

package pdf

import (
	"strings"
	"testing"
)

// Regression tests for table-layout:auto + colspan > 1 + max-width.
//
// Browsers do NOT enforce max-width as a hard layout width on colspan > 1
// table cells. The allocated width is the authoritative sum of the spanned
// column widths and must not be overridden by applyCSSMinMaxWidth after
// allocation.
//
// This gap was NOT covered by PR #702, which correctly fixed
// table-layout:fixed cells via the isFixedWidthAdvisoryOnly() guard but left
// table-layout:auto + colspan > 1 cells still subject to incorrect max-width
// enforcement.
//
// The correct guard is:
//
//	if (isFixedWidthAdvisoryOnly() && getStyle().getColSpan() <= 1) {
//	    applyCSSMinMaxWidth(c);
//	}

const tableAutoLayoutColspanMaxWidthTestAutoColspanHtml = "org/xhtmlrenderer/pdf/table-auto-layout-colspan-max-width.html"

func TestTableAutoLayoutColspanMaxWidth_renderDoesNotThrow(t *testing.T) {
	testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableAutoLayoutColspanMaxWidthTestAutoColspanHtml),
		"table-auto-layout-colspan-max-width.pdf")
}

// "Alpha" and "Beta" must be on the SAME line.
//
// The colspan="2" cell's allocated width is the sum of both columns (~400pt),
// which is far wider than the max-width: 80pt. If applyCSSMinMaxWidth is
// incorrectly applied, the cell shrinks to 80pt and "Beta" is forced onto a
// second line. If max-width is correctly ignored (browser behavior), both
// words fit comfortably on the first line.
func TestTableAutoLayoutColspanMaxWidth_alphaAndBetaAreOnSameLineWhenMaxWidthNotEnforced(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableAutoLayoutColspanMaxWidthTestAutoColspanHtml),
		"table-auto-layout-colspan-max-width.pdf")

	if !tableTestAnyLine(pdf.lines(), "Alpha", "Beta") {
		t.Errorf("no line contains both Alpha and Beta:\n%s", pdf.text)
	}
}

// "Alpha" and "Kappa" must be on DIFFERENT lines.
//
// Verifies the cell still wraps at the full allocated column width (~400pt)
// rather than becoming infinitely wide. Content must wrap, just not
// prematurely at max-width: 80pt.
func TestTableAutoLayoutColspanMaxWidth_textStillWrapsAtAllocatedColumnWidthNotAtMaxWidth(t *testing.T) {
	pdf := testUtilsPrintFile(t, testUtilsFromClasspathResource(t, tableAutoLayoutColspanMaxWidthTestAutoColspanHtml),
		"table-auto-layout-colspan-max-width.pdf")

	if tableTestAnyLine(pdf.lines(), "Alpha", "Omega") {
		t.Errorf("a line contains both Alpha and Omega:\n%s", pdf.text)
	}
}

// tableTestAnyLine reports whether a line contains both a and b.
func tableTestAnyLine(lines []string, a string, b string) bool {
	for _, line := range lines {
		if strings.Contains(line, a) && strings.Contains(line, b) {
			return true
		}
	}
	return false
}
