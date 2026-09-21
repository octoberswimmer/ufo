// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/ListStylesTest.java

package pdf

import "testing"

func TestListStyles_upperRoman(t *testing.T) {
	bytes := testUtilsFromUrl(t, "list-styles.upper-roman.html")
	pdf := testUtilsPrintFile(t, bytes, "list-styles.upper-roman.pdf")
	pdf.containsText(t,
		"I.", "II.", "III.", "IV.", "V.", "VI.", "VII.", "VIII.", "IX.", "X.",
		"i1", "i2", "i3", "i4", "i5", "i6", "i7", "i8", "i9", "i10",
	)
}

func TestListStyles_upperLatin(t *testing.T) {
	bytes := testUtilsFromUrl(t, "list-styles.upper-latin.html")
	pdf := testUtilsPrintFile(t, bytes, "list-styles.upper-latin.pdf")
	pdf.containsText(t,
		"A.", "B.", "C.", "D.", "E.", "F.", "G.", "H.", "I.", "J.",
		"i1", "i2", "i3", "i4", "i5", "i6", "i7", "i8", "i9", "i10",
	)
}
