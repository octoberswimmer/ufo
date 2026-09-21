// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/ITextRendererTest.java

package pdf

import (
	"bytes"
	"errors"
	"testing"

	"github.com/octoberswimmer/ufo/pdf/writer"
)

// Java's null version is "" in the port.
func TestITextRenderer_versionIsNullByDefault(t *testing.T) {
	cut := NewITextRenderer()
	if v := cut.GetPDFVersion(); v != "" {
		t.Errorf("GetPDFVersion() = %q, want \"\"", v)
	}
}

func TestITextRenderer_pdfAConformanceIsNullByDefault(t *testing.T) {
	cut := NewITextRenderer()
	if c := cut.GetPdfAConformance(); c != nil {
		t.Errorf("GetPdfAConformance() = %v, want nil", c)
	}
}

func TestITextRenderer_getAndSetPdfAConformance(t *testing.T) {
	cut := NewITextRenderer()
	cut.SetPdfAConformance(PdfAConformancePdfA1b)
	if c := cut.GetPdfAConformance(); c != PdfAConformancePdfA1b {
		t.Errorf("GetPdfAConformance() = %v, want PDF_A_1B", c)
	}
}

func TestITextRenderer_canClearPdfAConformance(t *testing.T) {
	cut := NewITextRenderer()
	cut.SetPdfAConformance(PdfAConformancePdfA1b)
	cut.SetPdfAConformance(nil)
	if c := cut.GetPdfAConformance(); c != nil {
		t.Errorf("GetPdfAConformance() = %v, want nil", c)
	}
}

func TestITextRenderer_settingOrClearingPdfAConformanceDoesNotAffectPdfXConformance(t *testing.T) {
	cut := NewITextRenderer()
	cut.SetPDFXConformance(writer.PdfWriterPdfx1a2001)

	cut.SetPdfAConformance(PdfAConformancePdfA1b)
	if c := cut.GetPDFXConformance(); c != writer.PdfWriterPdfx1a2001 {
		t.Errorf("GetPDFXConformance() = %d, want %d", c, writer.PdfWriterPdfx1a2001)
	}

	cut.SetPdfAConformance(nil)
	if c := cut.GetPDFXConformance(); c != writer.PdfWriterPdfx1a2001 {
		t.Errorf("GetPDFXConformance() = %d, want %d", c, writer.PdfWriterPdfx1a2001)
	}
}

// PDF/A and encryption are not ported (CHECKLIST.md, "Not ported"). Java
// fails with IllegalStateException "PDF/A conformance and PDF encryption are
// mutually exclusive"; the port fails earlier, at the PDF/X conformance level
// that PDF/A sets on the writer, which it reports as an unsupported feature.
func TestITextRenderer_pdfAConformanceAndEncryptionAreMutuallyExclusive(t *testing.T) {
	cut := NewITextRenderer()
	cut.SetPdfAConformance(PdfAConformancePdfA1b)
	cut.SetPDFEncryption(NewPDFEncryption([]byte("user"), []byte("owner")))
	if err := cut.SetDocumentFromString("<html><body>Hello</body></html>"); err != nil {
		t.Fatal(err)
	}
	if err := cut.Layout(); err != nil {
		t.Fatal(err)
	}

	err := cut.CreatePDF(&bytes.Buffer{})
	var unsupported *writer.UnsupportedFeatureError
	if !errors.As(err, &unsupported) {
		t.Fatalf("CreatePDF error = %v, want a *writer.UnsupportedFeatureError", err)
	}
}

func TestITextRenderer_getAndSetPDFVersion(t *testing.T) {
	cut := NewITextRenderer()
	if err := cut.SetPDFVersion("2.0"); err != nil {
		t.Fatal(err)
	}
	if v := cut.GetPDFVersion(); v != "2.0" {
		t.Errorf("GetPDFVersion() = %q, want 2.0", v)
	}
}

func TestITextRenderer_canSetVersionNull(t *testing.T) {
	cut := NewITextRenderer()
	if err := cut.SetPDFVersion(""); err != nil {
		t.Fatal(err)
	}
	if v := cut.GetPDFVersion(); v != "" {
		t.Errorf("GetPDFVersion() = %q, want \"\"", v)
	}
}

func TestITextRenderer_cannotSetIllegalVersion(t *testing.T) {
	cut := NewITextRenderer()
	err := cut.SetPDFVersion("0.1")
	want := `Invalid PDF version character: "0.1"; use one of constants in [1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 2.0].`
	if err == nil || err.Error() != want {
		t.Errorf("SetPDFVersion(\"0.1\") error = %v, want %q", err, want)
	}
}
