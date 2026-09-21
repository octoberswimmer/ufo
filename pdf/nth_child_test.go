// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/NthChildTest.java

package pdf

import "testing"

func TestNthChild_samplePdf(t *testing.T) {
	bytes := testUtilsFromUrl(t, "nth-child.html")
	pdf := testUtilsPrintFile(t, bytes, "nth-child.pdf")
	pdf.containsText(t, "o1. ODD", "o2.", "o3. ODD", "o4.")
	pdf.containsText(t, "e1.", "e2. EVEN", "e3.", "e4. EVEN")
	pdf.containsText(t, "t1.", "t2.", "t3. THIRD", "t4.")
	pdf.containsText(t, "af1. ", "af2. ", "af3. ", "af4. AFTER-FOURTH", "af5. AFTER-FOURTH", "af6. AFTER-FOURTH")
	pdf.containsText(t, "d1. d'Artagnan", "d2.", "d3.", "d4. d'Artagnan", "d5.", "d6.", "d7. d'Artagnan")
}
