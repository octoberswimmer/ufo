// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/ListsTest.java

package pdf

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/octoberswimmer/ufo"
)

// The Java test renders each file and only checks that no exception is
// thrown; it writes target/<file>.pdf, which the port writes to a temporary
// directory.
func TestLists_pageWithLists(t *testing.T) {
	for _, fileName := range []string{"page-with-lists.html", "list-sample.xhtml"} {
		t.Run(fileName, func(t *testing.T) {
			htmlContent, err := os.ReadFile(filepath.Join("testdata", fileName))
			if err != nil {
				t.Fatal(err)
			}

			document, err := ufo.XMLUtilNewDocumentBuilder().Parse(bytes.NewReader(htmlContent))
			if err != nil {
				t.Fatal(err)
			}

			renderer := NewITextRenderer()

			userAgent := NewITextUserAgent(renderer.GetOutputDevice(), int(math.Floor(float64(renderer.GetOutputDevice().GetDotsPerPoint())+0.5)))
			renderer.GetSharedContext().SetUserAgentCallback(userAgent)

			if err := renderer.SetDocumentWithUrl(document, ""); err != nil {
				t.Fatal(err)
			}

			outputStream, err := os.Create(filepath.Join(t.TempDir(), fileName+".pdf"))
			if err != nil {
				t.Fatal(err)
			}
			defer outputStream.Close()
			if err := renderer.Layout(); err != nil {
				t.Fatal(err)
			}
			if err := renderer.CreatePDF(outputStream); err != nil {
				t.Fatal(err)
			}
		})
	}
}
