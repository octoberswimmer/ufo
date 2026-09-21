package writer

// Destination types, as PdfDestination.XYZ etc.
const (
	PdfDestinationXyz   = 0
	PdfDestinationFit   = 1
	PdfDestinationFith  = 2
	PdfDestinationFitv  = 3
	PdfDestinationFitr  = 4
	PdfDestinationFitb  = 5
	PdfDestinationFitbh = 6
	PdfDestinationFitbv = 7
)

// PdfDestination is an explicit destination array: [page /XYZ left top zoom],
// [page /FitH top] and so on. The page is added with AddPage.
type PdfDestination struct {
	PdfArray
	hasPage bool
}

func nullOrNumber(v float32) PdfObject {
	if v < 0 {
		return PdfNull{}
	}
	return PdfNumber(v)
}

// NewPdfDestination is PdfDestination(type): /Fit or /FitB.
func NewPdfDestination(destType int) *PdfDestination {
	d := &PdfDestination{}
	if destType == PdfDestinationFitb {
		d.Add(PdfName("FitB"))
	} else {
		d.Add(PdfName("Fit"))
	}
	return d
}

// NewPdfDestinationWithParameter is PdfDestination(type, parameter): /FitH
// top, /FitV left, /FitBH top or /FitBV left.
func NewPdfDestinationWithParameter(destType int, parameter float32) *PdfDestination {
	d := &PdfDestination{}
	switch destType {
	case PdfDestinationFitv:
		d.Add(PdfName("FitV"))
	case PdfDestinationFitbh:
		d.Add(PdfName("FitBH"))
	case PdfDestinationFitbv:
		d.Add(PdfName("FitBV"))
	default:
		d.Add(PdfName("FitH"))
	}
	d.Add(PdfNumber(parameter))
	return d
}

// NewPdfDestinationWithLeftTopZoom is PdfDestination(type, left, top, zoom):
// /XYZ left top zoom. A negative left or top is written as null (unchanged);
// a zoom of 0 keeps the zoom.
func NewPdfDestinationWithLeftTopZoom(destType int, left, top, zoom float32) *PdfDestination {
	d := &PdfDestination{}
	d.Add(PdfName("XYZ"))
	d.Add(nullOrNumber(left))
	d.Add(nullOrNumber(top))
	d.Add(PdfNumber(zoom))
	return d
}

// NewPdfDestinationWithLeftBottomRightTop is PdfDestination(type, left,
// bottom, right, top): /FitR.
func NewPdfDestinationWithLeftBottomRightTop(destType int, left, bottom, right, top float32) *PdfDestination {
	d := &PdfDestination{}
	d.Add(PdfName("FitR"))
	d.Add(PdfNumber(left))
	d.Add(PdfNumber(bottom))
	d.Add(PdfNumber(right))
	d.Add(PdfNumber(top))
	return d
}

// HasPage reports whether AddPage was called.
func (d *PdfDestination) HasPage() bool { return d.hasPage }

// AddPage puts the page reference at the front of the array. It reports
// false, and changes nothing, when the destination already has a page.
func (d *PdfDestination) AddPage(page *PdfIndirectReference) bool {
	if d.hasPage {
		return false
	}
	d.items = append([]PdfObject{page}, d.items...)
	d.hasPage = true
	return true
}

// PdfAction is an action dictionary.
type PdfAction struct {
	PdfDictionary
}

// NewPdfAction returns an empty action dictionary to fill with Put.
func NewPdfAction() *PdfAction { return &PdfAction{PdfDictionary: *NewPdfDictionary()} }

// NewPdfActionWithUrl is PdfAction(String url): a URI action.
func NewPdfActionWithUrl(url string) *PdfAction {
	a := NewPdfAction()
	a.Put(PdfNameS, PdfNameUri)
	a.Put(PdfNameUri, NewPdfString(url))
	return a
}

// PdfActionGotoLocalPage is PdfAction.gotoLocalPage(page, dest, writer): a
// GoTo action to a destination on page number page.
func PdfActionGotoLocalPage(page int, dest *PdfDestination, writer *PdfWriter) *PdfAction {
	dest.AddPage(writer.GetPageReference(page))
	a := NewPdfAction()
	a.Put(PdfNameS, PdfNameGoto)
	a.Put(PdfNameD, dest)
	return a
}

// PdfActionJavaScript is PdfAction.javaScript(code, writer).
func PdfActionJavaScript(code string, writer *PdfWriter) *PdfAction {
	a := NewPdfAction()
	a.Put(PdfNameS, PdfNameJavascript)
	a.Put(PdfNameJs, NewPdfStringWithEncoding(code, PdfObjectTextUnicode))
	return a
}

// Annotation flags, as PdfAnnotation.FLAGS_PRINT etc.
const (
	PdfAnnotationFlagsInvisible = 1
	PdfAnnotationFlagsHidden    = 2
	PdfAnnotationFlagsPrint     = 4
	PdfAnnotationFlagsNozoom    = 8
	PdfAnnotationFlagsNorotate  = 16
	PdfAnnotationFlagsNoview    = 32
	PdfAnnotationFlagsReadonly  = 64
)

// Border styles, as PdfBorderDictionary.STYLE_SOLID etc.
const (
	PdfBorderDictionaryStyleSolid     = 0
	PdfBorderDictionaryStyleDashed    = 1
	PdfBorderDictionaryStyleBeveled   = 2
	PdfBorderDictionaryStyleInset     = 3
	PdfBorderDictionaryStyleUnderline = 4
)

// PdfAnnotation is an annotation dictionary. PdfWriter.AddAnnotation writes
// it and adds it to the current page.
type PdfAnnotation struct {
	PdfDictionary
}

// NewPdfAnnotation is PdfAnnotation(writer, llx, lly, urx, ury, action): a
// link annotation over the rectangle that performs the action.
func NewPdfAnnotation(writer *PdfWriter, llx, lly, urx, ury float32, action *PdfAction) *PdfAnnotation {
	a := &PdfAnnotation{PdfDictionary: *NewPdfDictionaryWithType(PdfNameAnnot)}
	a.Put(PdfNameSubtype, PdfNameLink)
	a.Put("Rect", NewPdfRectangle(llx, lly, urx, ury))
	if action != nil {
		a.Put(PdfNameA, &action.PdfDictionary)
	}
	return a
}

// NewPdfAnnotationWithRectangle returns an annotation with only /Type and
// /Rect, for a caller that sets /Subtype and the rest with Put.
func NewPdfAnnotationWithRectangle(writer *PdfWriter, rect *Rectangle) *PdfAnnotation {
	a := &PdfAnnotation{PdfDictionary: *NewPdfDictionaryWithType(PdfNameAnnot)}
	a.Put("Rect", rect.pdfArray())
	return a
}

// SetFlags sets /F.
func (a *PdfAnnotation) SetFlags(flags int) { a.Put(PdfNameF, PdfNumber(flags)) }

// SetBorderStyle sets /BS.
func (a *PdfAnnotation) SetBorderStyle(border *PdfDictionary) { a.Put("BS", border) }

// SetBorder sets /Border.
func (a *PdfAnnotation) SetBorder(border *PdfArray) { a.Put("Border", border) }

// NewPdfBorderDictionary is PdfBorderDictionary(borderWidth, borderStyle).
func NewPdfBorderDictionary(borderWidth float32, borderStyle int) *PdfDictionary {
	d := NewPdfDictionary()
	d.Put("W", PdfNumber(borderWidth))
	style := PdfName("S")
	switch borderStyle {
	case PdfBorderDictionaryStyleDashed:
		style = "D"
	case PdfBorderDictionaryStyleBeveled:
		style = "B"
	case PdfBorderDictionaryStyleInset:
		style = "I"
	case PdfBorderDictionaryStyleUnderline:
		style = "U"
	}
	d.Put(PdfNameS, style)
	return d
}

// NewPdfBorderArray is PdfBorderArray(hRadius, vRadius, width).
func NewPdfBorderArray(hRadius, vRadius, width float32) *PdfArray {
	return NewPdfArrayFloats(hRadius, vRadius, width)
}

// PdfOutline is an item of the document outline (a bookmark). Items are
// written when the document is closed.
type PdfOutline struct {
	writer      *PdfWriter
	parent      *PdfOutline
	kids        []*PdfOutline
	title       string
	destination *PdfDestination
	action      *PdfAction
	open        bool
	count       int
	ref         *PdfIndirectReference
}

// NewPdfOutline is PdfOutline(parent, destination, title): an open item
// under parent, which is another item or PdfWriter.GetRootOutline.
func NewPdfOutline(parent *PdfOutline, destination *PdfDestination, title string) *PdfOutline {
	return NewPdfOutlineWithOpen(parent, destination, title, true)
}

// NewPdfOutlineWithOpen is PdfOutline(parent, destination, title, open).
func NewPdfOutlineWithOpen(parent *PdfOutline, destination *PdfDestination, title string, open bool) *PdfOutline {
	o := &PdfOutline{writer: parent.writer, parent: parent, title: title, destination: destination, open: open}
	parent.kids = append(parent.kids, o)
	return o
}

// NewPdfOutlineWithAction is PdfOutline(parent, action, title).
func NewPdfOutlineWithAction(parent *PdfOutline, action *PdfAction, title string) *PdfOutline {
	o := &PdfOutline{writer: parent.writer, parent: parent, title: title, action: action, open: true}
	parent.kids = append(parent.kids, o)
	return o
}

// GetTitle returns the title.
func (o *PdfOutline) GetTitle() string { return o.title }

// SetTitle sets the title.
func (o *PdfOutline) SetTitle(title string) { o.title = title }

// GetKids returns the child items.
func (o *PdfOutline) GetKids() []*PdfOutline { return o.kids }

// IsOpen reports whether the children are shown.
func (o *PdfOutline) IsOpen() bool { return o.open }

// SetOpen sets whether the children are shown.
func (o *PdfOutline) SetOpen(open bool) { o.open = open }

// Parent returns the parent item.
func (o *PdfOutline) Parent() *PdfOutline { return o.parent }

// GetPdfDestination returns the destination.
func (o *PdfOutline) GetPdfDestination() *PdfDestination { return o.destination }

// Level returns the depth of the item; the root is 0.
func (o *PdfOutline) Level() int {
	if o.parent == nil {
		return 0
	}
	return o.parent.Level() + 1
}

// countVisible sets count to the number of descendants shown when the item
// is open: each child, plus the visible descendants of each open child.
func (o *PdfOutline) countVisible() int {
	n := 0
	for _, k := range o.kids {
		n++
		sub := k.countVisible()
		if k.open {
			n += sub
		}
	}
	o.count = n
	return n
}

func (o *PdfOutline) assignRefs() {
	o.ref = o.writer.newReference()
	for _, k := range o.kids {
		k.assignRefs()
	}
}

// writeOutlines writes the outline tree and returns the reference of its
// root dictionary.
func (w *PdfWriter) writeOutlines() *PdfIndirectReference {
	root := w.rootOutline
	root.countVisible()
	root.assignRefs()
	var write func(o *PdfOutline)
	write = func(o *PdfOutline) {
		d := NewPdfDictionary()
		if o.parent == nil {
			d.Put(PdfNameType, PdfName("Outlines"))
		} else {
			d.Put(PdfNameTitle, NewPdfStringWithEncoding(o.title, PdfObjectTextUnicode))
			d.Put("Parent", o.parent.ref)
			siblings := o.parent.kids
			for i, s := range siblings {
				if s != o {
					continue
				}
				if i > 0 {
					d.Put("Prev", siblings[i-1].ref)
				}
				if i+1 < len(siblings) {
					d.Put("Next", siblings[i+1].ref)
				}
			}
			if o.destination != nil {
				d.Put("Dest", &o.destination.PdfArray)
			}
			if o.action != nil {
				d.Put(PdfNameA, &o.action.PdfDictionary)
			}
		}
		if len(o.kids) > 0 {
			d.Put("First", o.kids[0].ref)
			d.Put("Last", o.kids[len(o.kids)-1].ref)
			if o.open || o.parent == nil {
				d.Put("Count", PdfNumber(o.count))
			} else {
				d.Put("Count", PdfNumber(-o.count))
			}
		}
		w.writeObject(o.ref, d)
		for _, k := range o.kids {
			write(k)
		}
	}
	write(root)
	return root.ref
}
