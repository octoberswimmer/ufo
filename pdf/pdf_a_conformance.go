// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/PdfAConformance.java

package pdf

import "github.com/octoberswimmer/ufo/pdf/writer"

// PdfAConformance lists the PDF/A conformance levels accepted by
// ITextRenderer.SetPdfAConformance.
//
// Note that PDF/A conformance also requires every font used in the document
// to be embedded. The built-in base-14 fonts (Helvetica, Times, Courier,
// Symbol, ZapfDingbats) can never be embedded, so a document that falls back
// to one of them cannot be PDF/A conformant regardless of this setting.
//
// The writer does not implement PDF/A: ITextRenderer.CreatePDF passes the
// level to writer.PdfWriter.SetPDFXConformance and returns its
// *writer.UnsupportedFeatureError.
type PdfAConformance struct {
	name            string
	pdfXConformance int
}

var (
	PdfAConformancePdfA1b = &PdfAConformance{name: "PDF_A_1B", pdfXConformance: writer.PdfWriterPdfa1b}
	PdfAConformancePdfA3b = &PdfAConformance{name: "PDF_A_3B", pdfXConformance: writer.PdfWriterPdfa3b}
)

func PdfAConformanceValues() []*PdfAConformance {
	return []*PdfAConformance{PdfAConformancePdfA1b, PdfAConformancePdfA3b}
}

func (c *PdfAConformance) Name() string {
	return c.name
}

func (c *PdfAConformance) String() string {
	return c.name
}

func (c *PdfAConformance) ToString() string {
	return c.name
}

// PdfXConformance is package-private in Java.
func (c *PdfAConformance) PdfXConformance() int {
	return c.pdfXConformance
}
