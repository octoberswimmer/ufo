// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/Html2Pdf.java

package pdf

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/octoberswimmer/ufo"
)

// Html2PdfFromClasspathResource renders a file of the classpath. The port has
// no class loader: the files embedded from the core's resources directory
// stand for the classpath (ufo.GeneralUtilGetURLFromClasspath).
func Html2PdfFromClasspathResource(fileName string) ([]byte, error) {
	htmlUrl := ufo.GeneralUtilGetURLFromClasspath(nil, fileName)
	if htmlUrl == nil {
		return nil, ufo.NewXRRuntimeException("Resource not found in classpath: " + fileName)
	}
	return Html2PdfFromUrl(htmlUrl)
}

func Html2PdfFromUrl(html *url.URL) (result []byte, err error) {
	defer iTextRendererRecover(&err)
	renderer := NewITextRenderer()
	renderer.GetSharedContext().SetMedia("pdf")
	renderer.GetSharedContext().SetInteractive(false)
	renderer.GetSharedContext().GetTextRenderer().SetSmoothingThreshold(0)

	builder := ufo.XMLUtilNewDocumentBuilder()
	doc, err := builder.ParseString(html.String())
	if err != nil {
		return nil, ufo.NewXRRuntimeExceptionWithCause(fmt.Sprintf("Failed to parse XML from %s", html), err)
	}
	// The JDK's DocumentBuilder.parse(String uri) sets the document URI,
	// which CreatePDFDocument uses as the base URL.
	if doc.DocumentURI == "" {
		doc.DocumentURI = html.String()
	}
	result, err = renderer.CreatePDFDocument(doc)
	if err != nil {
		return nil, ufo.NewXRRuntimeExceptionWithCause(fmt.Sprintf("Failed to parse XML from %s", html), err)
	}
	return result, nil
}

// Html2PdfMain renders args[0], a URL or the path of an XHTML file, to the
// PDF file args[1] with Html2PdfFromUrl. The Java class has no main method;
// this function is the command-line form of the class.
func Html2PdfMain(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Usage: ... [url] [pdf]")
	}
	source := args[0]
	if !strings.Contains(source, "://") {
		// maybe it's a file
		if _, err := os.Stat(source); err == nil {
			fileURL, err := naiveUserAgentFileToURL(source)
			if err != nil {
				return err
			}
			source = fileURL
		}
	}
	html, err := url.Parse(source)
	if err != nil {
		return err
	}
	data, err := Html2PdfFromUrl(html)
	if err != nil {
		return err
	}
	return os.WriteFile(args[1], data, 0o644)
}
