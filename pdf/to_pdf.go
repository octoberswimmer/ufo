// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ToPDF.java

package pdf

import (
	"fmt"
	"os"
	"strings"
)

// ToPDFMain ports ToPDF.main. Where Java prints the usage and exits with
// status 1, the usage is returned as the error.
func ToPDFMain(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Usage: ... [url] [pdf]")
	}
	url := args[0]
	if !strings.Contains(url, "://") {
		// maybe it's a file
		if _, err := os.Stat(url); err == nil {
			fileURL, err := naiveUserAgentFileToURL(url)
			if err != nil {
				return err
			}
			url = fileURL
		}
	}
	return toPDFCreatePDF(url, args[1])
}

func toPDFCreatePDF(url string, pdf string) (err error) {
	os, err := os.Create(pdf)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := os.Close(); err == nil {
			err = closeErr
		}
	}()
	renderer, err := ITextRendererFromUrl(url)
	if err != nil {
		return err
	}
	if err := renderer.Layout(); err != nil {
		return err
	}
	return renderer.CreatePDF(os)
}
