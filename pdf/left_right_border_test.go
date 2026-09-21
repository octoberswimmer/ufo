// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/LeftRightBorderTest.java

package pdf

import "testing"

func TestLeftRightBorder_deepHierarchyInCss(t *testing.T) {
	bytes := testUtilsFromUrl(t, "left-right-border.html")
	pdf := testUtilsPrintFile(t, bytes, "left-right-border.pdf")
	pdf.containsText(t,
		"Border top left radius",
		"Border top right radius",
		"Border bottom left radius",
		"Border bottom right radius",
	)
}
