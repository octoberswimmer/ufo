// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/DefaultPDFCreationListener.java

package pdf

// DefaultPDFCreationListener is a no-op implementation of a
// PDFCreationListener. Embed it and override methods as needed.
type DefaultPDFCreationListener struct {
}

var _ PDFCreationListener = (*DefaultPDFCreationListener)(nil)

func NewDefaultPDFCreationListener() *DefaultPDFCreationListener {
	return &DefaultPDFCreationListener{}
}

func (l *DefaultPDFCreationListener) PreOpen(iTextRenderer *ITextRenderer) {}

func (l *DefaultPDFCreationListener) PreWrite(iTextRenderer *ITextRenderer, pageCount int) {}

func (l *DefaultPDFCreationListener) OnClose(renderer *ITextRenderer) {}
