package writer

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"fmt"
	"io"
	"sort"
	"time"
)

// PDF versions, as PdfWriter.VERSION_1_2 etc.
const (
	PdfWriterVersion12 = "1.2"
	PdfWriterVersion13 = "1.3"
	PdfWriterVersion14 = "1.4"
	PdfWriterVersion15 = "1.5"
	PdfWriterVersion16 = "1.6"
	PdfWriterVersion17 = "1.7"
	PdfWriterVersion20 = "2.0"
)

// Compression levels for SetCompressionLevel. With PdfStreamNoCompression
// streams are written without a filter, which is what tests that read the
// content streams as text want.
const (
	PdfStreamDefaultCompression = -1
	PdfStreamNoCompression      = 0
	PdfStreamBestSpeed          = 1
	PdfStreamBestCompression    = 9
)

// Viewer preferences accepted by SetViewerPreferences, with OpenPDF's values.
const (
	PdfWriterPageLayoutSinglePage = 1
	PdfWriterPageModeUseOutlines  = 128
	PdfWriterHideToolbar          = 1 << 12
	PdfWriterHideMenubar          = 1 << 13
	PdfWriterHideWindowUI         = 1 << 14
	PdfWriterFitWindow            = 1 << 15
	PdfWriterCenterWindow         = 1 << 16
	PdfWriterDisplayDocTitle      = 1 << 17
)

// Permissions and encryption types, as PdfWriter.ALLOW_PRINTING etc. They
// exist so PDFEncryption can be ported; SetEncryption reports that
// encryption is not supported.
const (
	PdfWriterAllowPrinting         = 4 + 2048
	PdfWriterAllowModifyContents   = 8
	PdfWriterAllowCopy             = 16
	PdfWriterAllowModifyAnnotation = 32
	PdfWriterAllowFillIn           = 256
	PdfWriterAllowScreenreaders    = 512
	PdfWriterAllowAssembly         = 1024
	PdfWriterAllowDegradedPrinting = 4
	PdfWriterStandardEncryption40  = 0
	PdfWriterStandardEncryption128 = 1
	PdfWriterEncryptionAes128      = 2
	PdfWriterEncryptionAes256V3    = 3
)

// PDF/X and PDF/A conformance levels, as PdfWriter.PDFXNONE etc.
const (
	PdfWriterPdfxnone   = 0
	PdfWriterPdfx1a2001 = 1
	PdfWriterPdfx32002  = 2
	PdfWriterPdfa1a     = 3
	PdfWriterPdfa1b     = 4
	PdfWriterPdfa3b     = 5
)

// PdfPageEvent receives the page events of a PdfWriter, the part of
// com.lowagie.text.pdf.PdfPageEvent that does not depend on OpenPDF's
// high-level layout.
type PdfPageEvent interface {
	OnOpenDocument(writer *PdfWriter, document *Document)
	OnStartPage(writer *PdfWriter, document *Document)
	OnEndPage(writer *PdfWriter, document *Document)
	OnCloseDocument(writer *PdfWriter, document *Document)
}

// resourceSet records the resources one content stream refers to.
type resourceSet struct {
	fonts    map[PdfName]*PdfIndirectReference
	xobjects map[PdfName]*PdfIndirectReference
	gstates  map[PdfName]*PdfIndirectReference
}

func newResourceSet() *resourceSet {
	return &resourceSet{
		fonts:    map[PdfName]*PdfIndirectReference{},
		xobjects: map[PdfName]*PdfIndirectReference{},
		gstates:  map[PdfName]*PdfIndirectReference{},
	}
}

func (r *resourceSet) merge(other *resourceSet) {
	for k, v := range other.fonts {
		r.fonts[k] = v
	}
	for k, v := range other.xobjects {
		r.xobjects[k] = v
	}
	for k, v := range other.gstates {
		r.gstates[k] = v
	}
}

func refDictionary(m map[PdfName]*PdfIndirectReference) *PdfDictionary {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, string(k))
	}
	sort.Strings(names)
	d := NewPdfDictionary()
	for _, n := range names {
		d.Put(PdfName(n), m[PdfName(n)])
	}
	return d
}

func (r *resourceSet) dictionary() *PdfDictionary {
	d := NewPdfDictionary()
	d.Put("ProcSet", NewPdfArray(PdfName("PDF"), PdfName("Text"), PdfName("ImageB"), PdfName("ImageC"), PdfName("ImageI")))
	if len(r.fonts) > 0 {
		d.Put("Font", refDictionary(r.fonts))
	}
	if len(r.xobjects) > 0 {
		d.Put("XObject", refDictionary(r.xobjects))
	}
	if len(r.gstates) > 0 {
		d.Put("ExtGState", refDictionary(r.gstates))
	}
	return d
}

// fontEntry is a font as used by one writer (OpenPDF's FontDetails).
type fontEntry struct {
	font *BaseFont
	name PdfName
	ref  *PdfIndirectReference
	// used maps each glyph id shown to the first character it was shown for.
	used map[uint16]rune
}

type xobjectEntry struct {
	name PdfName
	ref  *PdfIndirectReference
}

// PdfWriter writes a Document to an io.Writer. Objects are written as soon as
// they are complete: each page when it ends, images when they are first
// used, and fonts, templates, outlines, the catalog and the information
// dictionary when the document is closed.
type PdfWriter struct {
	doc *Document
	out io.Writer
	pos int64
	err error

	// offsets[n] is the file position of object n; 0 for an object number
	// that was reserved and never written, which becomes a free entry.
	offsets []int64

	version          string
	compressionLevel int
	info             *PdfDictionary
	extraCatalog     *PdfDictionary
	viewerPrefs      int
	pageEvent        PdfPageEvent

	opened bool
	closed bool

	pagesRef  *PdfIndirectReference
	pageRefs  []*PdfIndirectReference
	pageCount int

	directContent      *PdfContentByte
	directContentUnder *PdfContentByte
	currentPageSize    *Rectangle
	currentAnnots      []*PdfIndirectReference

	fonts        []*fontEntry
	fontsByKey   map[any]*fontEntry
	images       map[*imageData]*xobjectEntry
	templates    []*PdfTemplate
	gstates      map[string]*xobjectEntry
	xobjectCount int

	rootOutline *PdfOutline
}

// PdfWriterGetInstance returns the writer that writes document to out, as
// PdfWriter.getInstance(document, os).
func PdfWriterGetInstance(document *Document, out io.Writer) (*PdfWriter, error) {
	if document.writer != nil {
		return nil, newDocumentException("the document already has a PdfWriter")
	}
	w := &PdfWriter{
		doc:              document,
		out:              out,
		offsets:          []int64{0},
		version:          PdfWriterVersion14,
		compressionLevel: PdfStreamDefaultCompression,
		info:             NewPdfDictionary(),
		extraCatalog:     NewPdfDictionary(),
		fontsByKey:       map[any]*fontEntry{},
		images:           map[*imageData]*xobjectEntry{},
		gstates:          map[string]*xobjectEntry{},
	}
	now := pdfDate(time.Now())
	w.info.Put(PdfNameProducer, NewPdfString("ufo pdf/writer"))
	w.info.Put(PdfNameCreationdate, NewPdfString(now))
	w.info.Put(PdfNameModdate, NewPdfString(now))
	w.directContent = newContentByte(w)
	w.directContentUnder = newContentByte(w)
	w.rootOutline = &PdfOutline{writer: w, open: true}
	document.writer = w
	return w, nil
}

func pdfDate(t time.Time) string {
	s := t.Format("D:20060102150405")
	_, offset := t.Zone()
	if offset == 0 {
		return s + "Z"
	}
	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}
	return fmt.Sprintf("%s%c%02d'%02d'", s, sign, offset/3600, offset%3600/60)
}

// SetPdfVersion selects the version written in the file header. It has to
// be called before the document is opened.
func (w *PdfWriter) SetPdfVersion(version string) {
	if w.opened {
		return
	}
	w.version = version
}

// GetInfo returns the information dictionary, which is written when the
// document is closed.
func (w *PdfWriter) GetInfo() *PdfDictionary { return w.info }

// GetExtraCatalog returns a dictionary whose entries are added to the
// document catalog.
func (w *PdfWriter) GetExtraCatalog() *PdfDictionary { return w.extraCatalog }

// SetCompressionLevel sets the zlib level of the streams written after the
// call; PdfStreamNoCompression writes them without a filter.
func (w *PdfWriter) SetCompressionLevel(level int) {
	if level < PdfStreamDefaultCompression || level > PdfStreamBestCompression {
		level = PdfStreamDefaultCompression
	}
	w.compressionLevel = level
}

// GetCompressionLevel returns the level set with SetCompressionLevel.
func (w *PdfWriter) GetCompressionLevel() int { return w.compressionLevel }

// SetFullCompression asks OpenPDF for object streams and a cross-reference
// stream. This writer always writes plain objects and a cross-reference
// table, which every reader accepts, so the call changes nothing.
func (w *PdfWriter) SetFullCompression() {}

// SetViewerPreferences sets the viewer preferences; the bits are or-ed with
// those of earlier calls. PdfWriterDisplayDocTitle and the other
// /ViewerPreferences flags and PdfWriterPageModeUseOutlines are written.
func (w *PdfWriter) SetViewerPreferences(preferences int) { w.viewerPrefs |= preferences }

// SetPageEvent sets the receiver of page events.
func (w *PdfWriter) SetPageEvent(event PdfPageEvent) { w.pageEvent = event }

// GetPageEvent returns the receiver of page events.
func (w *PdfWriter) GetPageEvent() PdfPageEvent { return w.pageEvent }

// GetDirectContent returns the content of the current page. The same object
// is returned for every page; it is emptied when a page ends.
func (w *PdfWriter) GetDirectContent() *PdfContentByte { return w.directContent }

// GetDirectContentUnder returns the content written before that of
// GetDirectContent on the current page.
func (w *PdfWriter) GetDirectContentUnder() *PdfContentByte { return w.directContentUnder }

// GetPageNumber returns the 1-based number of the current page (0 before
// the document is opened).
func (w *PdfWriter) GetPageNumber() int {
	if !w.opened {
		return 0
	}
	return w.pageCount + 1
}

// GetCurrentPageNumber is GetPageNumber.
func (w *PdfWriter) GetCurrentPageNumber() int { return w.GetPageNumber() }

// GetPageReference returns the reference of page number page (1-based). The
// page may not exist yet: its object number is reserved. A reference to a
// page that is never written points to a free object.
func (w *PdfWriter) GetPageReference(page int) *PdfIndirectReference {
	if page < 1 {
		panic(newDocumentException("The page numbers start at 1."))
	}
	for len(w.pageRefs) < page {
		w.pageRefs = append(w.pageRefs, w.newReference())
	}
	return w.pageRefs[page-1]
}

// GetRootOutline returns the parent of the top-level outlines.
func (w *PdfWriter) GetRootOutline() *PdfOutline { return w.rootOutline }

// GetPdfIndirectReference reserves an object number.
func (w *PdfWriter) GetPdfIndirectReference() *PdfIndirectReference { return w.newReference() }

func (w *PdfWriter) newReference() *PdfIndirectReference {
	w.offsets = append(w.offsets, 0)
	return &PdfIndirectReference{number: len(w.offsets) - 1}
}

// AddToBody writes an object to the file and returns it with its reference.
func (w *PdfWriter) AddToBody(object PdfObject) (*PdfIndirectObject, error) {
	return w.AddToBodyWithRef(object, w.newReference())
}

// AddToBodyWithRef writes an object under a reference reserved earlier.
func (w *PdfWriter) AddToBodyWithRef(object PdfObject, ref *PdfIndirectReference) (*PdfIndirectObject, error) {
	if !w.opened || w.closed {
		return nil, newDocumentException("the document is not open")
	}
	w.writeObject(ref, object)
	return &PdfIndirectObject{ref: ref, object: object}, w.err
}

func (w *PdfWriter) write(p []byte) {
	if w.err != nil {
		return
	}
	n, err := w.out.Write(p)
	w.pos += int64(n)
	if err != nil {
		w.err = err
	}
}

func (w *PdfWriter) writeObject(ref *PdfIndirectReference, object PdfObject) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%d 0 obj\n", ref.number)
	writeObject(&b, object)
	b.WriteString("\nendobj\n")
	w.offsets[ref.number] = w.pos
	w.write(b.Bytes())
}

// newStream returns a stream, deflated unless compression is off.
func (w *PdfWriter) newStream(data []byte) *PdfStream {
	s := NewPdfStream(data)
	if w.compressionLevel != PdfStreamNoCompression {
		level := w.compressionLevel
		if level == PdfStreamDefaultCompression {
			level = zlib.DefaultCompression
		}
		s.FlateCompress(level)
	}
	return s
}

func (w *PdfWriter) openDocument() {
	w.opened = true
	w.write([]byte("%PDF-" + w.version + "\n%\xe2\xe3\xcf\xd3\n"))
	w.pagesRef = w.newReference()
	if w.pageEvent != nil {
		w.pageEvent.OnOpenDocument(w, w.doc)
	}
	w.initPage()
}

func (w *PdfWriter) initPage() {
	w.currentPageSize = w.doc.pageSize.clone()
	w.currentAnnots = nil
	if w.pageEvent != nil {
		w.pageEvent.OnStartPage(w, w.doc)
	}
}

func (w *PdfWriter) isPageEmpty() bool {
	return w.directContent.Size() == 0 && w.directContentUnder.Size() == 0 && len(w.currentAnnots) == 0
}

func (w *PdfWriter) newPage() bool {
	if w.isPageEmpty() {
		w.currentPageSize = w.doc.pageSize.clone()
		return false
	}
	w.finishPage()
	w.initPage()
	return true
}

func (w *PdfWriter) finishPage() {
	if w.pageEvent != nil {
		w.pageEvent.OnEndPage(w, w.doc)
	}
	for _, cb := range []*PdfContentByte{w.directContentUnder, w.directContent} {
		if len(cb.stateStack) > 0 {
			panic(newDocumentException("Unbalanced save/restore state operators."))
		}
		if cb.inText {
			panic(newDocumentException("Unbalanced begin/end text operators."))
		}
	}
	var content bytes.Buffer
	content.Write(w.directContentUnder.content.Bytes())
	content.Write(w.directContent.content.Bytes())
	resources := newResourceSet()
	resources.merge(w.directContentUnder.resources)
	resources.merge(w.directContent.resources)

	contentRef := w.newReference()
	w.writeObject(contentRef, w.newStream(content.Bytes()))

	page := NewPdfDictionaryWithType("Page")
	page.Put("Parent", w.pagesRef)
	page.Put("MediaBox", w.currentPageSize.pdfArray())
	page.Put("Resources", resources.dictionary())
	page.Put("Contents", contentRef)
	if len(w.currentAnnots) > 0 {
		annots := NewPdfArray()
		for _, a := range w.currentAnnots {
			annots.Add(a)
		}
		page.Put("Annots", annots)
	}
	w.pageCount++
	w.writeObject(w.GetPageReference(w.pageCount), page)

	w.directContent.Reset()
	w.directContentUnder.Reset()
	w.currentAnnots = nil
}

// AddAnnotation writes an annotation and adds it to the current page.
func (w *PdfWriter) AddAnnotation(annot *PdfAnnotation) {
	if !w.opened || w.closed {
		panic(newDocumentException("the document is not open"))
	}
	ref := w.newReference()
	w.writeObject(ref, &annot.PdfDictionary)
	w.currentAnnots = append(w.currentAnnots, ref)
}

// addFont registers a font with the writer and returns its entry. The 14
// standard fonts are shared by name, TrueType fonts by BaseFont.
func (w *PdfWriter) addFont(bf *BaseFont) *fontEntry {
	var key any = bf
	if bf.tt == nil {
		key = bf.afm.fontName
	}
	if e, ok := w.fontsByKey[key]; ok {
		return e
	}
	e := &fontEntry{
		font: bf,
		name: PdfName(fmt.Sprintf("F%d", len(w.fonts)+1)),
		ref:  w.newReference(),
		used: map[uint16]rune{},
	}
	w.fonts = append(w.fonts, e)
	w.fontsByKey[key] = e
	return e
}

func (w *PdfWriter) addGState(gs *PdfGState) *xobjectEntry {
	key := gs.key()
	if e, ok := w.gstates[key]; ok {
		return e
	}
	e := &xobjectEntry{name: PdfName(fmt.Sprintf("GS%d", len(w.gstates)+1)), ref: w.newReference()}
	w.gstates[key] = e
	w.writeObject(e.ref, gs.dictionary())
	return e
}

func (w *PdfWriter) addTemplate(t *PdfTemplate) {
	if t.ref != nil {
		return
	}
	w.xobjectCount++
	t.name = PdfName(fmt.Sprintf("Xf%d", w.xobjectCount))
	t.ref = w.newReference()
	w.templates = append(w.templates, t)
}

func (w *PdfWriter) closeDocument() error {
	if !w.isPageEmpty() {
		w.finishPage()
	}
	if w.pageEvent != nil {
		w.pageEvent.OnCloseDocument(w, w.doc)
	}
	if w.pageCount == 0 {
		w.closed = true
		return newDocumentException("The document has no pages.")
	}

	// A template may add further templates while it is written.
	for i := 0; i < len(w.templates); i++ {
		w.writeTemplate(w.templates[i])
	}
	for _, f := range w.fonts {
		w.writeFont(f)
	}

	kids := NewPdfArray()
	for i := 0; i < w.pageCount; i++ {
		kids.Add(w.pageRefs[i])
	}
	pages := NewPdfDictionaryWithType("Pages")
	pages.Put("Kids", kids)
	pages.Put("Count", PdfNumber(w.pageCount))
	w.writeObject(w.pagesRef, pages)

	catalog := NewPdfDictionaryWithType("Catalog")
	catalog.Put("Pages", w.pagesRef)
	if len(w.rootOutline.kids) > 0 {
		catalog.Put("PageMode", PdfName("UseOutlines"))
		catalog.Put("Outlines", w.writeOutlines())
	}
	w.putViewerPreferences(catalog)
	catalog.PutAll(w.extraCatalog)
	catalogRef := w.newReference()
	w.writeObject(catalogRef, catalog)

	infoRef := w.newReference()
	w.writeObject(infoRef, w.info)

	var infoBytes bytes.Buffer
	w.info.WritePdf(&infoBytes)
	sum := md5.Sum(append(infoBytes.Bytes(), []byte(fmt.Sprintf("%d %d", w.pos, time.Now().UnixNano()))...))
	id := NewPdfStringBytes(sum[:]).SetHexWriting(true)

	xrefPos := w.pos
	var b bytes.Buffer
	fmt.Fprintf(&b, "xref\n0 %d\n", len(w.offsets))
	b.WriteString("0000000000 65535 f \n")
	for n := 1; n < len(w.offsets); n++ {
		if w.offsets[n] == 0 {
			b.WriteString("0000000000 00001 f \n")
		} else {
			fmt.Fprintf(&b, "%010d 00000 n \n", w.offsets[n])
		}
	}
	trailer := NewPdfDictionary()
	trailer.Put("Size", PdfNumber(len(w.offsets)))
	trailer.Put("Root", catalogRef)
	trailer.Put("Info", infoRef)
	trailer.Put("ID", NewPdfArray(id, id))
	b.WriteString("trailer\n")
	trailer.WritePdf(&b)
	fmt.Fprintf(&b, "\nstartxref\n%d\n%%%%EOF\n", xrefPos)
	w.write(b.Bytes())
	w.closed = true
	return w.err
}

func (w *PdfWriter) putViewerPreferences(catalog *PdfDictionary) {
	if w.viewerPrefs&PdfWriterPageModeUseOutlines != 0 {
		catalog.Put("PageMode", PdfName("UseOutlines"))
	}
	if w.viewerPrefs&PdfWriterPageLayoutSinglePage != 0 {
		catalog.Put("PageLayout", PdfName("SinglePage"))
	}
	prefs := NewPdfDictionary()
	flags := []struct {
		bit  int
		name PdfName
	}{
		{PdfWriterHideToolbar, "HideToolbar"},
		{PdfWriterHideMenubar, "HideMenubar"},
		{PdfWriterHideWindowUI, "HideWindowUI"},
		{PdfWriterFitWindow, "FitWindow"},
		{PdfWriterCenterWindow, "CenterWindow"},
		{PdfWriterDisplayDocTitle, "DisplayDocTitle"},
	}
	for _, f := range flags {
		if w.viewerPrefs&f.bit != 0 {
			prefs.Put(f.name, PdfBoolean(true))
		}
	}
	if prefs.Size() > 0 {
		catalog.Put("ViewerPreferences", prefs)
	}
}

func (w *PdfWriter) writeTemplate(t *PdfTemplate) {
	if len(t.stateStack) > 0 {
		panic(newDocumentException("Unbalanced save/restore state operators."))
	}
	s := w.newStream(t.content.Bytes())
	s.Put("Type", PdfName("XObject"))
	s.Put("Subtype", PdfName("Form"))
	s.Put("FormType", PdfNumber(1))
	s.Put("BBox", t.bbox.pdfArray())
	if t.matrix != nil {
		s.Put("Matrix", NewPdfArrayFloats(t.matrix[:]...))
	}
	s.Put("Resources", t.resources.dictionary())
	w.writeObject(t.ref, s)
}
