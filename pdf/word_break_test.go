// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/WordBreakTest.java

package pdf

import "testing"

func TestWordBreak_breakAll(t *testing.T) {
	result := testUtilsFromClasspathResource(t, "org/xhtmlrenderer/pdf/break-all.html")
	pdf := testUtilsPrintFile(t, result, "break-all.pdf")
	pdf.containsExactText(t, "HelloWorld1\nHelloWorld2\nHelloWorld3\nHelloWorld4\nHelloWorld5\n")
}
