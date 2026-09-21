// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/ConcurrentPdfGenerationTest.java

package pdf

import (
	"sync"
	"testing"
)

func concurrentPdfGenerationTestVerifyPdf(t testing.TB, pdf *testUtilsPDF) {
	t.Helper()
	pdf.containsText(t, "Bill To:", "John doe", "john.do@mail.com")
	pdf.containsText(t, "Invoice #:", "INV-666")
	pdf.containsText(t, "Invoice Date:", "Sep 27, 2023")
	pdf.doesNotContainText(t, "Password", "Secret")
}

func TestConcurrentPdfGeneration_samplePdf(t *testing.T) {
	bytes := testUtilsFromUrl(t, "sample.html")
	pdf := testUtilsPrintFile(t, bytes, "sample.pdf")
	concurrentPdfGenerationTestVerifyPdf(t, pdf)
}

// The Java test renders the document 20 times on a pool of 4 threads while
// another thread calls System.gc every 15 ms; the port renders it on 4
// goroutines (run it with -race to check for data races).
func TestConcurrentPdfGeneration_concurrentPdfGeneration(t *testing.T) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				pdf := testUtilsFromUrl(t, "sample.html")
				concurrentPdfGenerationTestVerifyPdf(t, testUtilsPrintFile(t, pdf, "sample.pdf"))
				t.Logf("Check #%d ok", i)
			}
		}()
	}
	for j := range 20 {
		jobs <- j
	}
	close(jobs)
	wg.Wait()
}
