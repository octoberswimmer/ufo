// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/HeaderTest.java

package pdf

import "testing"

func TestHeader_pageWithHeader(t *testing.T) {
	bytes := testUtilsFromClasspathResource(t, "page-with-header.html")

	pdf := testUtilsPrintFile(t, bytes, "page-with-header.pdf")
	pdf.containsText(t, "Header", "Body", "Footer")
}
