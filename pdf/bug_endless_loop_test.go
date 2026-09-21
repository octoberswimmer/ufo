// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/bug/EndlessLoopTest.java

package pdf

import (
	"testing"
	"time"
)

// The Java test has a timeout of 10 seconds.
func TestEndlessLoop_wordwrap(t *testing.T) {
	done := make(chan []byte, 1)
	go func() {
		done <- testUtilsFromUrl(t, "org/xhtmlrenderer/pdf/bug/EndlessLoopTest_wordwrap.html")
	}()
	select {
	case pdf := <-done:
		testUtilsPrintFile(t, pdf, "EndlessLoopTest_wordwrap.pdf").containsText(t,
			"floated",
			"word wrapped",
		)
	case <-time.After(10 * time.Second):
		t.Fatal("rendering did not finish within 10 seconds")
	}
}
