// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/borderradius/BorderRadiusNonRegressionTest.java

package pdf

import "testing"

// This used to throw a ClassCastException (before this fix).
func TestBorderRadiusNonRegression_borderRadiusWithBorderWidthZero(t *testing.T) {
	pdf := testUtilsFromUrl(t, "org/xhtmlrenderer/pdf/borderradius/borderRadiusWithBorderWidthZero.html")
	testUtilsPrintFile(t, pdf, "borderRadiusWithBorderWidthZero.pdf").containsText(t, "Some content")
}
