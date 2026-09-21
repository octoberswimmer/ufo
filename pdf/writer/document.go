package writer

// Rectangle is com.lowagie.text.Rectangle: lower-left and upper-right
// corners in points.
type Rectangle struct {
	llx, lly, urx, ury float32
	border             int
	borderWidth        float32
}

// NewRectangle returns the rectangle with the given corners.
func NewRectangle(llx, lly, urx, ury float32) *Rectangle {
	return &Rectangle{llx: llx, lly: lly, urx: urx, ury: ury, border: -1, borderWidth: -1}
}

// NewRectangleWithUrxUry returns the rectangle from the origin to (urx, ury).
func NewRectangleWithUrxUry(urx, ury float32) *Rectangle {
	return NewRectangle(0, 0, urx, ury)
}

func (r *Rectangle) GetLeft() float32   { return r.llx }
func (r *Rectangle) GetBottom() float32 { return r.lly }
func (r *Rectangle) GetRight() float32  { return r.urx }
func (r *Rectangle) GetTop() float32    { return r.ury }
func (r *Rectangle) GetWidth() float32  { return r.urx - r.llx }
func (r *Rectangle) GetHeight() float32 { return r.ury - r.lly }

// SetBorder and SetBorderWidth record the values; the writer draws no
// rectangle borders.
func (r *Rectangle) SetBorder(border int)         { r.border = border }
func (r *Rectangle) SetBorderWidth(width float32) { r.borderWidth = width }
func (r *Rectangle) GetBorder() int               { return r.border }
func (r *Rectangle) GetBorderWidth() float32      { return r.borderWidth }

func (r *Rectangle) clone() *Rectangle {
	c := *r
	return &c
}

func (r *Rectangle) pdfArray() *PdfArray { return NewPdfRectangle(r.llx, r.lly, r.urx, r.ury) }

// Document is com.lowagie.text.Document reduced to what a caller that
// positions everything itself uses: the page size, page breaks and the
// document information. The margins are recorded and have no effect.
type Document struct {
	pageSize                                         *Rectangle
	marginLeft, marginRight, marginTop, marginBottom float32
	writer                                           *PdfWriter
	open                                             bool
	closed                                           bool
}

// NewDocument returns a document whose first page has the given size.
func NewDocument(pageSize *Rectangle, marginLeft, marginRight, marginTop, marginBottom float32) *Document {
	return &Document{
		pageSize:   pageSize.clone(),
		marginLeft: marginLeft, marginRight: marginRight, marginTop: marginTop, marginBottom: marginBottom,
	}
}

// Open starts the document: the file header is written and the first page
// begins with the current page size. It needs a PdfWriter obtained with
// PdfWriterGetInstance.
func (d *Document) Open() {
	if d.open || d.closed {
		return
	}
	if d.writer == nil {
		panic(newDocumentException("the document has no PdfWriter"))
	}
	d.open = true
	d.writer.openDocument()
}

// IsOpen reports whether Open was called and Close was not.
func (d *Document) IsOpen() bool { return d.open && !d.closed }

// SetPageSize sets the size of the pages started after this call.
func (d *Document) SetPageSize(pageSize *Rectangle) bool {
	d.pageSize = pageSize.clone()
	return true
}

// GetPageSize returns the size set for the next page.
func (d *Document) GetPageSize() *Rectangle { return d.pageSize }

// NewPage ends the current page and starts a new one with the size last
// given to SetPageSize. When nothing was written to the current page, no
// page is added: the current page takes the new size and NewPage reports
// false, as in OpenPDF.
func (d *Document) NewPage() bool {
	if !d.IsOpen() {
		return false
	}
	return d.writer.newPage()
}

// GetPageNumber returns the 1-based number of the current page (0 before
// the document is opened).
func (d *Document) GetPageNumber() int {
	if d.writer == nil {
		return 0
	}
	return d.writer.GetPageNumber()
}

// Close ends the last page and writes the fonts, outlines, catalog,
// information dictionary, cross-reference table and trailer. It returns the
// first error of the underlying io.Writer, or a *DocumentException when the
// document has no pages.
func (d *Document) Close() error {
	if d.closed {
		return nil
	}
	if !d.open {
		d.closed = true
		return nil
	}
	d.closed = true
	return d.writer.closeDocument()
}

func (d *Document) addInfo(key PdfName, value string) bool {
	if d.writer == nil || d.closed {
		return false
	}
	d.writer.info.Put(key, NewPdfStringWithEncoding(value, PdfObjectTextUnicode))
	return true
}

// AddTitle sets /Title of the information dictionary.
func (d *Document) AddTitle(title string) bool { return d.addInfo(PdfNameTitle, title) }

// AddAuthor sets /Author.
func (d *Document) AddAuthor(author string) bool { return d.addInfo(PdfNameAuthor, author) }

// AddSubject sets /Subject.
func (d *Document) AddSubject(subject string) bool { return d.addInfo(PdfNameSubject, subject) }

// AddKeywords sets /Keywords.
func (d *Document) AddKeywords(keywords string) bool { return d.addInfo(PdfNameKeywords, keywords) }

// AddCreator sets /Creator.
func (d *Document) AddCreator(creator string) bool { return d.addInfo(PdfNameCreator, creator) }

// AddProducer sets /Producer.
func (d *Document) AddProducer(producer string) bool { return d.addInfo(PdfNameProducer, producer) }

// AddHeader sets an arbitrary key of the information dictionary.
func (d *Document) AddHeader(name, content string) bool { return d.addInfo(PdfName(name), content) }
