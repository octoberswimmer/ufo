// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/HasPseudoClassTest.java

package pdf

import "testing"

func TestHasPseudoClass_hasPseudoClassAffectsRenderedContent(t *testing.T) {
	bytes := testUtilsFromUrl(t, "has-pseudo-class.html")
	pdf := testUtilsPrintFile(t, bytes, "has-pseudo-class.pdf")
	pdf.containsText(t, "CARD-1x [HAS]", "CARD-2 [PLAIN]")
}
