// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/PDFCreationListener.java

package pdf

// PDFCreationListener is a callback listener for PDF creation. To use this,
// call ITextRenderer.SetListener. Note that with a handle on the
// ITextRenderer instance (provided in the callback arguments) you can access
// the writer.PdfWriter instance being used to create the document, using
// ITextRenderer.GetOutputDevice, then calling ITextOutputDevice.GetWriter.
type PDFCreationListener interface {
	// PreOpen is called immediately after the writer.Document instance is
	// created but before the call to writer.Document.Open is called. At this
	// point you may still modify certain properties of the PDF document
	// header via the writer.PdfWriter; once Open is called, you can't change,
	// e.g. the version.
	PreOpen(iTextRenderer *ITextRenderer)

	// PreWrite is called immediately before the pages of the PDF file are
	// about to be written out. This is an opportunity to modify any document
	// metadata that will be used to generate the PDF header fields (the
	// document information dictionary). Document metadata may be accessed
	// through the ITextOutputDevice that is returned by
	// ITextRenderer.GetOutputDevice.
	//
	// pageCount is the number of pages that will be written to the PDF
	// document.
	PreWrite(iTextRenderer *ITextRenderer, pageCount int)

	// OnClose is called immediately before the writer.Document instance is
	// closed, e.g. before writer.Document.Close is called.
	OnClose(renderer *ITextRenderer)
}
