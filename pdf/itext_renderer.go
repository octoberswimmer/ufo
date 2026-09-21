// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextRenderer.java

package pdf

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// These two defaults combine to produce an effective resolution of 96 px to the inch
const (
	ITextRendererDefaultDotsPerPoint float32 = 20.0 * 4.0 / 3.0
	ITextRendererDefaultDotsPerPixel int     = 20
)

// ITextRenderer renders a document to PDF: SetDocument (or one of its
// variants), Layout, CreatePDF.
//
// The methods that load, lay out or write a document return an error where
// the Java methods throw: a panic of the core (an *ufo.XRRuntimeException for
// Java's runtime exceptions) is recovered and returned.
type ITextRenderer struct {
	pdfProducer        string
	pdfCreator         string
	compression        int
	compressionEnabled bool

	sharedContext *ufo.SharedContext
	outputDevice  *ITextOutputDevice

	// doc is nil until a document is set.
	doc *dom.Document
	// root is a nil interface until Layout has run.
	root ufo.BlockBoxI

	dotsPerPoint float32

	pdfDoc *writer.Document
	// writer is nil until CreatePDF has run.
	writer *writer.PdfWriter

	// pdfEncryption may be nil.
	pdfEncryption *PDFEncryption

	// note: not hard-coding a default version in the pdfVersion field as this
	// may change between writer releases
	// check for "" before calling writer.SetPdfVersion()
	// use one of the values in writer.PdfWriterVersion...
	pdfVersion string

	// pdfPageEvent may be nil.
	pdfPageEvent writer.PdfPageEvent

	// dim is nil until Layout has run.
	dim *geom.Dimension

	scaleToFit bool

	// validPdfVersions is a TreeSet in Java, for an exception message with
	// non-random order; the keys are sorted where the message is built.
	validPdfVersions map[string]struct{}

	// pdfXConformance is nil until SetPDFXConformance is called.
	pdfXConformance *int

	// pdfAConformance may be nil.
	pdfAConformance *PdfAConformance

	tagged bool

	// listener may be nil.
	listener PDFCreationListener
}

// iTextRendererRecover turns a panic into the error that the exported entry
// points of this package return (PORTING.md, Exceptions). A Go runtime error
// (the counterpart of a NullPointerException or an
// ArrayIndexOutOfBoundsException) is returned with the stack of the panic.
func iTextRendererRecover(err *error) {
	if r := recover(); r != nil {
		switch e := r.(type) {
		case runtime.Error:
			*err = fmt.Errorf("%w\n%s", e, debug.Stack())
		case error:
			*err = e
		default:
			*err = fmt.Errorf("%v", r)
		}
	}
}

// NewITextRendererWithFile ports ITextRenderer(File): it creates a renderer
// for the document in the file, with the directory of the file as base URL.
func NewITextRendererWithFile(file string) (r *ITextRenderer, err error) {
	defer iTextRendererRecover(&err)
	r = NewITextRenderer()
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	fileURL, err := naiveUserAgentFileToURL(abs)
	if err != nil {
		return nil, err
	}
	parentURL, err := naiveUserAgentFileToURL(filepath.Dir(abs))
	if err != nil {
		return nil, err
	}
	if err := r.SetDocumentWithUrl(
		r.loadDocument(fileURL),
		parentURL,
	); err != nil {
		return nil, err
	}
	return r, nil
}

func NewITextRenderer() *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixel(ITextRendererDefaultDotsPerPoint, ITextRendererDefaultDotsPerPixel)
}

func NewITextRendererWithFontResolver(fontResolver ufo.FontResolver) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelFontResolver(ITextRendererDefaultDotsPerPoint, ITextRendererDefaultDotsPerPixel, fontResolver)
}

func NewITextRendererWithDotsPerPointDotsPerPixel(dotsPerPoint float32, dotsPerPixel int) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDevice(dotsPerPoint, dotsPerPixel, NewITextOutputDevice(dotsPerPoint))
}

func NewITextRendererWithDotsPerPointDotsPerPixelFontResolver(dotsPerPoint float32, dotsPerPixel int, fontResolver ufo.FontResolver) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceFontResolver(dotsPerPoint, dotsPerPixel, NewITextOutputDevice(dotsPerPoint), fontResolver)
}

func NewITextRendererWithOutputDeviceUserAgent(outputDevice *ITextOutputDevice, userAgent *ITextUserAgent) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolver(outputDevice.GetDotsPerPoint(), userAgent.GetDotsPerPixel(), outputDevice, userAgent, NewITextFontResolver())
}

func NewITextRendererWithDotsPerPointDotsPerPixelOutputDevice(dotsPerPoint float32, dotsPerPixel int, outputDevice *ITextOutputDevice) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgent(dotsPerPoint, dotsPerPixel, outputDevice, NewITextUserAgent(outputDevice, dotsPerPixel))
}

func NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceFontResolver(dotsPerPoint float32, dotsPerPixel int, outputDevice *ITextOutputDevice, fontResolver ufo.FontResolver) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolver(dotsPerPoint, dotsPerPixel, outputDevice, NewITextUserAgent(outputDevice, dotsPerPixel), fontResolver)
}

func NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgent(dotsPerPoint float32, dotsPerPixel int, outputDevice *ITextOutputDevice, userAgent *ITextUserAgent) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolver(dotsPerPoint, dotsPerPixel, outputDevice, userAgent, NewITextFontResolver())
}

func NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolver(dotsPerPoint float32, dotsPerPixel int, outputDevice *ITextOutputDevice, userAgent *ITextUserAgent,
	fontResolver ufo.FontResolver) *ITextRenderer {
	return NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolverReplacedElementFactoryTextRenderer(dotsPerPoint, dotsPerPixel, outputDevice, userAgent, fontResolver,
		NewITextReplacedElementFactory(outputDevice), NewITextTextRenderer())
}

// NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolverReplacedElementFactoryTextRenderer
// takes the user agent as ufo.UserAgentCallback where Java takes an
// ITextUserAgent, so that a user agent that does not embed ITextUserAgent can
// be supplied; the renderer only passes it to the SharedContext.
func NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgentFontResolverReplacedElementFactoryTextRenderer(dotsPerPoint float32, dotsPerPixel int, outputDevice *ITextOutputDevice, userAgent ufo.UserAgentCallback,
	fontResolver ufo.FontResolver, replacedElementFactory ufo.ReplacedElementFactory,
	textRenderer ufo.TextRenderer) *ITextRenderer {
	r := &ITextRenderer{
		compression:        9,
		compressionEnabled: true,
		validPdfVersions: map[string]struct{}{
			writer.PdfWriterVersion12: {},
			writer.PdfWriterVersion13: {},
			writer.PdfWriterVersion14: {},
			writer.PdfWriterVersion15: {},
			writer.PdfWriterVersion16: {},
			writer.PdfWriterVersion17: {},
			writer.PdfWriterVersion20: {},
		},
	}
	r.pdfProducer = r.defaultPdfProducer()
	r.pdfCreator = r.pdfProducer
	r.dotsPerPoint = dotsPerPoint
	r.outputDevice = outputDevice
	r.sharedContext = ufo.NewSharedContextWithUserAgentFontResolverReplacedElementFactoryTextRendererDpiDotsPerPixel(userAgent, fontResolver, replacedElementFactory, textRenderer,
		72*r.dotsPerPoint, dotsPerPixel)

	r.outputDevice.SetSharedContext(r.sharedContext)
	return r
}

// GetDocument returns nil until a document is set.
func (r *ITextRenderer) GetDocument() *dom.Document {
	return r.doc
}

// GetFontResolver panics when the renderer was created with a font resolver
// that is not an ITextFontResolverI, as the Java cast does.
func (r *ITextRenderer) GetFontResolver() ITextFontResolverI {
	return r.sharedContext.GetFontResolver().(ITextFontResolverI)
}

func (r *ITextRenderer) loadDocument(uri string) *dom.Document {
	return r.sharedContext.GetUac().GetXMLResource(uri).GetDocument()
}

func ITextRendererFromUrl(uri string) (renderer *ITextRenderer, err error) {
	defer iTextRendererRecover(&err)
	renderer = NewITextRenderer()
	if err := renderer.SetDocumentWithUrl(renderer.loadDocument(uri), uri); err != nil {
		return nil, err
	}
	return renderer, nil
}

func (r *ITextRenderer) SetDocument(doc *dom.Document) error {
	return r.SetDocumentWithUrl(doc, "")
}

// SetDocumentWithUrl ports setDocument(Document, String); url is "" where
// Java passes null.
func (r *ITextRenderer) SetDocumentWithUrl(doc *dom.Document, url string) error {
	return r.SetDocumentWithUrlNsh(doc, url, ufo.NewXhtmlNamespaceHandler())
}

func ITextRendererFromString(content string) (*ITextRenderer, error) {
	return ITextRendererFromStringWithBaseUrl(content, "")
}

// ITextRendererFromStringWithBaseUrl ports fromString(String, String);
// baseUrl is "" where Java passes null.
func ITextRendererFromStringWithBaseUrl(content string, baseUrl string) (*ITextRenderer, error) {
	renderer := NewITextRenderer()
	if err := renderer.SetDocumentFromStringWithBaseUrl(content, baseUrl); err != nil {
		return nil, err
	}
	return renderer, nil
}

// SetDocumentFromString sets the document from XML (XHTML) text.
func (r *ITextRenderer) SetDocumentFromString(content string) (err error) {
	defer iTextRendererRecover(&err)
	return r.SetDocumentWithUrl(r.parse(content), "")
}

// SetDocumentFromStringWithBaseUrl ports setDocumentFromString(String,
// String); baseUrl is "" where Java passes null.
func (r *ITextRenderer) SetDocumentFromStringWithBaseUrl(content string, baseUrl string) (err error) {
	defer iTextRendererRecover(&err)
	return r.SetDocumentWithUrl(r.parse(content), baseUrl)
}

// SetDocumentFromHTMLString sets the document from HTML text that need not
// be well-formed XML, parsed with dom.ParseHTMLString (the HTML5 parsing
// algorithm). baseURL may be "".
//
// This method is not part of the Java class: it is the one addition to its
// API. Java users parse HTML with jsoup and convert the result with jsoup's
// W3CDom before calling setDocument.
func (r *ITextRenderer) SetDocumentFromHTMLString(html string, baseURL string) (err error) {
	defer iTextRendererRecover(&err)
	doc, err := dom.ParseHTMLString(html)
	if err != nil {
		return err
	}
	return r.SetDocumentWithUrl(doc, baseURL)
}

func (r *ITextRenderer) parse(content string) *dom.Document {
	is := strings.NewReader(content)
	return ufo.XMLResourceLoadInputSource(ufo.NewInputSourceReader(is)).GetDocument()
}

// SetDocumentWithUrlNsh ports setDocument(Document, String,
// NamespaceHandler); url is "" where Java passes null.
func (r *ITextRenderer) SetDocumentWithUrlNsh(doc *dom.Document, url string, nsh ufo.NamespaceHandler) (err error) {
	defer iTextRendererRecover(&err)
	r.doc = doc

	r.GetFontResolver().FlushFontFaceFonts()

	r.sharedContext.Reset()
	r.sharedContext.SetBaseURL(url)
	r.sharedContext.SetNamespaceHandler(nsh)
	r.sharedContext.GetCss().SetDocumentContext(r.sharedContext, r.sharedContext.GetNamespaceHandler(), doc, &iTextRendererNullUserInterface{})
	r.GetFontResolver().ImportFontFaces(r.sharedContext.GetCss().GetFontFaceRules(), r.sharedContext.GetUac())
	return nil
}

// GetPDFEncryption may return nil.
func (r *ITextRenderer) GetPDFEncryption() *PDFEncryption {
	return r.pdfEncryption
}

// SetPDFEncryption requests encryption. The writer does not implement it:
// CreatePDF then returns a *writer.UnsupportedFeatureError.
func (r *ITextRenderer) SetPDFEncryption(pdfEncryption *PDFEncryption) {
	r.pdfEncryption = pdfEncryption
}

// SetPDFVersion takes "" where Java takes null. It returns an error (Java:
// IllegalArgumentException) for a version that is not one of the
// writer.PdfWriterVersion... constants.
func (r *ITextRenderer) SetPDFVersion(v string) error {
	if _, ok := r.validPdfVersions[v]; v != "" && !ok {
		versions := make([]string, 0, len(r.validPdfVersions))
		for version := range r.validPdfVersions {
			versions = append(versions, version)
		}
		sort.Strings(versions)
		return ufo.NewXRRuntimeException(fmt.Sprintf(
			"Invalid PDF version character: \"%s\"; use one of constants in [%s].", v, strings.Join(versions, ", ")))
	}
	r.pdfVersion = v
	return nil
}

// GetPDFVersion returns "" when no version was set.
func (r *ITextRenderer) GetPDFVersion() string {
	return r.pdfVersion
}

// SetPDFXConformance requests a PDF/X conformance level. The writer
// implements none but writer.PdfWriterPdfxnone: CreatePDF returns a
// *writer.UnsupportedFeatureError for any other.
func (r *ITextRenderer) SetPDFXConformance(pdfXConformance int) {
	r.pdfXConformance = &pdfXConformance
}

func (r *ITextRenderer) GetPDFXConformance() int {
	if r.pdfXConformance == nil {
		return '0'
	}
	return *r.pdfXConformance
}

// SetPdfAConformance requests PDF/A conformance for the generated document:
// registers an sRGB ICC output intent and document-level XMP metadata in
// addition to setting the underlying PDF/X conformance flag. See
// PdfAConformance for the font-embedding caveat that this setting cannot
// enforce on its own.
//
// PDF/A forbids encryption, so combining this with SetPDFEncryption will fail
// at CreatePDF time.
//
// The writer does not implement PDF/A: CreatePDF returns a
// *writer.UnsupportedFeatureError when a conformance level is set.
// pdfAConformance may be nil.
func (r *ITextRenderer) SetPdfAConformance(pdfAConformance *PdfAConformance) {
	r.pdfAConformance = pdfAConformance
}

// GetPdfAConformance may return nil.
func (r *ITextRenderer) GetPdfAConformance() *PdfAConformance {
	return r.pdfAConformance
}

// SetTagged requests a tagged PDF: a structure tree describing headings,
// paragraphs and images (with their alt text) so that screen readers and
// other assistive technology can navigate the document. See
// ITextOutputDevice for which HTML elements are currently tagged.
//
// The writer does not implement tagged PDF: CreatePDF returns a
// *writer.UnsupportedFeatureError when tagged is true.
func (r *ITextRenderer) SetTagged(tagged bool) {
	r.tagged = tagged
}

func (r *ITextRenderer) IsTagged() bool {
	return r.tagged
}

func (r *ITextRenderer) Layout() (err error) {
	defer iTextRendererRecover(&err)
	c := r.newLayoutContext()
	root := ufo.BoxBuilderCreateRootBox(c, r.doc)
	root.SetContainingBlock(ufo.NewViewportBox(r.getInitialExtents(c)))
	root.Layout(c)
	c.GetSharedContext().LogUnsupportedFeatures()
	r.dim = root.GetLayer().GetPaintingDimension(c)
	root.GetLayer().TrimEmptyPages(r.dim.Height)
	root.GetLayer().LayoutPages(c)
	r.root = root
	return nil
}

func (r *ITextRenderer) getInitialExtents(c *ufo.LayoutContext) *geom.Rectangle {
	first := ufo.LayerCreatePageBox(c, "first")

	return geom.NewRectangle(0, 0, first.GetContentWidth(c), first.GetContentHeight(c))
}

func (r *ITextRenderer) newRenderingContext(initialPageNo int) *ufo.RenderingContext {
	fontContext := NewITextFontContext()
	r.sharedContext.GetTextRenderer().Setup(fontContext)
	return r.sharedContext.NewRenderingContextInstanceWithRootLayerInitialPageNo(r.outputDevice, fontContext, r.root.GetLayer(), initialPageNo)
}

func (r *ITextRenderer) newLayoutContext() *ufo.LayoutContext {
	fontContext := NewITextFontContext()
	result := r.sharedContext.NewLayoutContextInstance(fontContext)
	r.sharedContext.GetTextRenderer().Setup(fontContext)

	return result
}

// CreatePDFDocument ports byte[] createPDF(Document): it sets the document,
// lays it out and returns the PDF.
func (r *ITextRenderer) CreatePDFDocument(source *dom.Document) ([]byte, error) {
	if err := r.SetDocumentWithUrl(source, source.DocumentURI); err != nil {
		return nil, err
	}
	if err := r.Layout(); err != nil {
		return nil, err
	}

	var bos bytes.Buffer
	if err := r.CreatePDF(&bos); err != nil {
		return nil, err
	}
	if err := r.FinishPDF(); err != nil {
		return nil, err
	}
	return bos.Bytes(), nil
}

// CreatePDFDocumentWithOs ports createPDF(Document, OutputStream).
func (r *ITextRenderer) CreatePDFDocumentWithOs(source *dom.Document, os io.Writer) error {
	if err := r.SetDocumentWithUrl(source, source.DocumentURI); err != nil {
		return err
	}
	if err := r.Layout(); err != nil {
		return err
	}
	if err := r.CreatePDF(os); err != nil {
		return err
	}
	return r.FinishPDF()
}

// CreatePDF ports createPDF(OutputStream): it writes the document that
// Layout laid out and closes the PDF.
func (r *ITextRenderer) CreatePDF(os io.Writer) error {
	return r.CreatePDFWithFinishInitialPageNo(os, true, 0)
}

func (r *ITextRenderer) WriteNextDocument() error {
	return r.WriteNextDocumentWithInitialPageNo(0)
}

func (r *ITextRenderer) WriteNextDocumentWithInitialPageNo(initialPageNo int) (err error) {
	defer iTextRendererRecover(&err)
	pages := r.root.GetLayer().GetPages()

	c := r.newRenderingContext(initialPageNo)
	firstPage := pages[0]
	firstPageSize :=
		writer.NewRectangle(0, 0, float32(firstPage.GetWidth(c))/r.dotsPerPoint,
			float32(firstPage.GetHeight(c))/r.dotsPerPoint)

	r.outputDevice.SetStartPageNo(r.writer.GetPageNumber())

	r.pdfDoc.SetPageSize(firstPageSize)
	r.pdfDoc.NewPage()

	return r.writePDF(pages, c, firstPageSize, r.pdfDoc, r.writer)
}

func (r *ITextRenderer) FinishPDF() (err error) {
	defer iTextRendererRecover(&err)
	if r.pdfDoc != nil {
		return r.closeDocument(r.pdfDoc, r.writer)
	}
	return nil
}

func (r *ITextRenderer) CreatePDFWithFinish(os io.Writer, finish bool) error {
	return r.CreatePDFWithFinishInitialPageNo(os, finish, 0)
}

// CreatePDFWithFinishInitialPageNo ports createPDF(OutputStream, boolean,
// int).
//
// NOTE: Caller is responsible for cleaning up the io.Writer if something goes
// wrong.
func (r *ITextRenderer) CreatePDFWithFinishInitialPageNo(os io.Writer, finish bool, initialPageNo int) (err error) {
	defer iTextRendererRecover(&err)
	pages := r.root.GetLayer().GetPages()

	c := r.newRenderingContext(initialPageNo)

	firstPage := pages[0]

	pageWidth := r.calculateWidth(c, firstPage)

	firstPageSize :=
		writer.NewRectangle(0, 0, float32(pageWidth)/r.dotsPerPoint,
			float32(firstPage.GetHeight(c))/r.dotsPerPoint)

	doc := writer.NewDocument(firstPageSize, 0, 0, 0, 0)
	w, err := writer.PdfWriterGetInstance(doc, os)
	if err != nil {
		return err
	}
	if r.pdfVersion != "" {
		w.SetPdfVersion(r.pdfVersion)
	}

	info := w.GetInfo()
	info.Put(writer.PdfNameProducer, writer.NewPdfString(r.pdfProducer))
	info.Put(writer.PdfNameCreator, writer.NewPdfString(r.pdfCreator))

	if r.compressionEnabled {
		w.SetCompressionLevel(r.compression)
		// Object/cross-reference streams are a PDF 1.5+ feature; PDF/A-1 is pinned to PDF 1.4 and forbids them.
		if r.pdfAConformance != PdfAConformancePdfA1b {
			w.SetFullCompression()
		}
	}

	effectiveXConformance := r.pdfXConformance
	if r.pdfAConformance != nil {
		pdfXConformance := r.pdfAConformance.PdfXConformance()
		effectiveXConformance = &pdfXConformance
	}
	if effectiveXConformance != nil {
		if err := w.SetPDFXConformance(*effectiveXConformance); err != nil {
			return err
		}
	}

	if r.tagged {
		if err := w.SetTagged(); err != nil {
			return err
		}
		w.SetViewerPreferences(writer.PdfWriterDisplayDocTitle)
		lang := r.doc.GetDocumentElement().GetAttribute("lang")
		if lang != "" {
			w.GetExtraCatalog().Put(writer.PdfNameLang, writer.NewPdfString(lang))
		}
	}

	if r.pdfAConformance != nil {
		if r.pdfEncryption != nil {
			return ufo.NewXRRuntimeException("PDF/A conformance and PDF encryption are mutually exclusive")
		}
		nonEmbeddedFonts := r.GetFontResolver().GetNonEmbeddedFontFaceFamilies()
		if len(nonEmbeddedFonts) != 0 {
			return ufo.NewXRRuntimeException(
				"PDF/A conformance requires all fonts to be embedded; not embedded: [" + strings.Join(nonEmbeddedFonts, ", ") + "]")
		}
	}

	if r.pdfPageEvent != nil {
		w.SetPageEvent(r.pdfPageEvent)
	}

	if r.pdfEncryption != nil {
		if err := w.SetEncryption(r.pdfEncryption.GetUserPassword(), r.pdfEncryption.GetOwnerPassword(),
			r.pdfEncryption.GetAllowedPrivileges(), r.pdfEncryption.GetEncryptionType()); err != nil {
			return err
		}
	}
	r.pdfDoc = doc
	r.writer = w

	r.firePreOpen()
	doc.Open()

	if r.pdfAConformance != nil {
		if err := iTextRendererSetOutputIntent(w); err != nil {
			return err
		}
	}

	if err := r.writePDF(pages, c, firstPageSize, doc, w); err != nil {
		return err
	}

	if finish {
		return r.closeDocument(doc, w)
	}
	return nil
}

// iTextRendererSetOutputIntent registers the sRGB output intent. Java passes
// the JDK's sRGB ICC profile; the writer does not implement output intents
// and returns a *writer.UnsupportedFeatureError before looking at the
// profile, so none is supplied.
func iTextRendererSetOutputIntent(w *writer.PdfWriter) error {
	return w.SetOutputIntents("", "sRGB IEC61966-2.1", "http://www.color.org", "sRGB IEC61966-2.1", nil)
}

// closeDocument takes a writer that may be nil.
func (r *ITextRenderer) closeDocument(doc *writer.Document, w *writer.PdfWriter) error {
	if r.pdfAConformance != nil && w != nil {
		if err := w.CreateXmpMetadata(); err != nil {
			return err
		}
	}
	r.fireOnClose()
	return doc.Close()
}

func (r *ITextRenderer) firePreOpen() {
	if r.listener != nil {
		r.listener.PreOpen(r)
	}
}

func (r *ITextRenderer) firePreWrite(pageCount int) {
	if r.listener != nil {
		r.listener.PreWrite(r, pageCount)
	}
}

func (r *ITextRenderer) fireOnClose() {
	if r.listener != nil {
		r.listener.OnClose(r)
	}
}

func (r *ITextRenderer) writePDF(pages []*ufo.PageBox, c *ufo.RenderingContext, firstPageSize *writer.Rectangle,
	doc *writer.Document,
	w *writer.PdfWriter) error {
	r.outputDevice.SetRoot(r.root)

	r.outputDevice.Start(r.doc)
	r.outputDevice.SetWriter(w)
	r.outputDevice.InitializePage(w.GetDirectContent(), firstPageSize.GetHeight())

	r.root.GetLayer().AssignPagePaintingPositions(c, ufo.LayerPagedModePagedModePrint)

	pageCount := len(r.root.GetLayer().GetPages())
	c.SetPageCount(pageCount)
	r.firePreWrite(pageCount) // opportunity to adjust meta data
	r.setDidValues(doc)       // set PDF header fields from meta data
	for i := 0; i < pageCount; i++ {

		// Java stops here with "Timeout occurred" when the thread was
		// interrupted; a goroutine has no interrupted flag.

		currentPage := pages[i]
		c.SetPage(i, currentPage)
		if err := r.paintPage(c, w, currentPage); err != nil {
			return err
		}
		r.outputDevice.FinishPage()
		if i != pageCount-1 {
			nextPage := pages[i+1]
			pageWidth := r.calculateWidth(c, nextPage)
			nextPageSize :=
				writer.NewRectangle(0, 0, float32(pageWidth)/r.dotsPerPoint,
					float32(nextPage.GetHeight(c))/r.dotsPerPoint)
			doc.SetPageSize(nextPageSize)
			doc.NewPage()
			r.outputDevice.InitializePage(w.GetDirectContent(), nextPageSize.GetHeight())
		}
	}

	r.outputDevice.Finish(c, r.root)
	return nil
}

// setDidValues sets the document information dictionary values from html
// metadata
func (r *ITextRenderer) setDidValues(doc *writer.Document) {
	v := r.outputDevice.GetMetadataByName("title")
	if v != nil {
		doc.AddTitle(*v)
	}
	v = r.outputDevice.GetMetadataByName("author")
	if v != nil {
		doc.AddAuthor(*v)
	}
	v = r.outputDevice.GetMetadataByName("subject")
	if v != nil {
		doc.AddSubject(*v)
	}
	v = r.outputDevice.GetMetadataByName("keywords")
	if v != nil {
		doc.AddKeywords(*v)
	}
}

func (r *ITextRenderer) paintPage(c *ufo.RenderingContext, w *writer.PdfWriter, page *ufo.PageBox) error {
	if err := r.provideMetadataToPage(w, page); err != nil {
		return err
	}

	page.PaintBackground(c, 0, ufo.LayerPagedModePagedModePrint)
	page.PaintMarginAreas(c, 0, ufo.LayerPagedModePagedModePrint)
	page.PaintBorder(c, 0, ufo.LayerPagedModePagedModePrint)

	working := r.outputDevice.GetClip()

	content := page.GetPrintClippingBounds(c)
	if r.IsScaleToFit() {
		pageWidth := r.calculateWidth(c, page)
		content.SetSize(pageWidth, int(content.GetSize().GetHeight())) //RTD - to change
	}
	r.outputDevice.Clip(content)

	top := -page.GetPaintingTop() + page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeTop)

	left := page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeLeft)

	r.outputDevice.Translate(float64(left), float64(top))
	r.root.GetLayer().Paint(c)
	r.outputDevice.Translate(float64(-left), float64(-top))

	r.outputDevice.SetClip(working)
	return nil
}

// provideMetadataToPage passes the XMP metadata of the page to the writer,
// which does not implement page metadata and returns a
// *writer.UnsupportedFeatureError.
func (r *ITextRenderer) provideMetadataToPage(w *writer.PdfWriter, page *ufo.PageBox) error {
	var metadata []byte
	if page.GetMetadata() != nil {
		metadataBody := r.stringifyMetadata(page.GetMetadata())
		if metadataBody != nil {
			metadata = []byte(r.createXPacket(*r.stringifyMetadata(page.GetMetadata())))
		}
	}

	if metadata != nil {
		return w.SetPageXmpMetadata(metadata)
	}
	return nil
}

// stringifyMetadata returns nil when the element has no child element.
func (r *ITextRenderer) stringifyMetadata(element *dom.Element) *string {
	target := iTextRendererGetFirstChildElement(element)
	if target == nil {
		return nil
	}

	// Java writes the element with an identity Transformer whose
	// OMIT_XML_DECLARATION property is "yes".
	output := xmlTransformerSerialize(target)
	return &output
}

// iTextRendererGetFirstChildElement returns nil when element has no child
// element.
func iTextRendererGetFirstChildElement(element *dom.Element) *dom.Element {
	n := element.GetFirstChild()
	for n != nil {
		if n.GetNodeType() == dom.ElementNode {
			return n.(*dom.Element)
		}
		n = n.GetNextSibling()
	}
	return nil
}

func (r *ITextRenderer) createXPacket(metadata string) string {
	return "<?xpacket begin='\ufeff' id='W5M0MpCehiHzreSzNTczkc9d'?>\n" +
		metadata +
		"\n<?xpacket end='r'?>"
}

func (r *ITextRenderer) GetOutputDevice() *ITextOutputDevice {
	return r.outputDevice
}

func (r *ITextRenderer) GetSharedContext() *ufo.SharedContext {
	return r.sharedContext
}

func (r *ITextRenderer) ExportText(w io.Writer) (err error) {
	defer iTextRendererRecover(&err)
	c := r.newRenderingContext(0)
	c.SetPageCount(len(r.root.GetLayer().GetPages()))
	return r.root.ExportText(c, w)
}

// GetRootBox returns a nil interface until Layout has run.
func (r *ITextRenderer) GetRootBox() ufo.BlockBoxI {
	return r.root
}

func (r *ITextRenderer) GetDotsPerPoint() float32 {
	return r.dotsPerPoint
}

func (r *ITextRenderer) FindPagePositionsByID(pattern *regexp.Regexp) (positions []*PagePosition, err error) {
	defer iTextRendererRecover(&err)
	return r.outputDevice.FindPagePositionsByID(r.newLayoutContext(), pattern), nil
}

// iTextRendererNullUserInterface ports the private class
// ITextRenderer.NullUserInterface.
type iTextRendererNullUserInterface struct {
}

var _ ufo.UserInterface = (*iTextRendererNullUserInterface)(nil)

func (u *iTextRendererNullUserInterface) IsHover(e *dom.Element) bool {
	return false
}

func (u *iTextRendererNullUserInterface) IsActive(e *dom.Element) bool {
	return false
}

func (u *iTextRendererNullUserInterface) IsFocus(e *dom.Element) bool {
	return false
}

func (r *ITextRenderer) calculateWidth(c *ufo.RenderingContext, firstPage *ufo.PageBox) int {
	if r.IsScaleToFit() {
		pageWidth := firstPage.GetWidth(c)
		pageRec := firstPage.GetPrintClippingBounds(c)
		if r.dim.GetWidth() > pageRec.GetWidth() {
			margin := firstPage.GetMargin(c)
			pageWidth = int(float32(r.dim.GetWidth()) + margin.Left() + margin.Right())
		}
		return pageWidth
	} else {
		return firstPage.GetWidth(c)
	}
}

// GetListener may return nil.
func (r *ITextRenderer) GetListener() PDFCreationListener {
	return r.listener
}

func (r *ITextRenderer) SetListener(listener PDFCreationListener) {
	r.listener = listener
}

// GetWriter returns nil until CreatePDF has run.
func (r *ITextRenderer) GetWriter() *writer.PdfWriter {
	return r.writer
}

// GetPdfPageEvent may return nil.
func (r *ITextRenderer) GetPdfPageEvent() writer.PdfPageEvent {
	return r.pdfPageEvent
}

// SetPdfPageEvent takes an event that may be nil.
func (r *ITextRenderer) SetPdfPageEvent(pdfPageEvent writer.PdfPageEvent) {
	r.pdfPageEvent = pdfPageEvent
}

func (r *ITextRenderer) SetScaleToFit(scaleToFit bool) {
	r.scaleToFit = scaleToFit
}

func (r *ITextRenderer) IsScaleToFit() bool {
	return r.scaleToFit
}

func (r *ITextRenderer) SetPDFProducer(pdfProducer string) {
	r.pdfProducer = pdfProducer
}

func (r *ITextRenderer) SetPDFCreator(pdfCreator string) {
	r.pdfCreator = pdfCreator
}

func (r *ITextRenderer) SetCompression(compression int) {
	r.compression = compression
}

func (r *ITextRenderer) SetCompressionEnabled(enabled bool) {
	r.compressionEnabled = enabled
}

// defaultPdfProducer is "Flying Saucer <version> with <OpenPDF version>" in
// Java, from the manifest of the jar and OpenPDF's Document.getVersion. The
// port has neither; it names the module and its PDF writer.
func (r *ITextRenderer) defaultPdfProducer() string {
	return "ufo (Flying Saucer port) with ufo pdf/writer"
}
