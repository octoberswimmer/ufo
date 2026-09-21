// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextOutputDevice.java

package pdf

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

type iTextOutputDeviceDrawType int

const (
	iTextOutputDeviceDrawTypeFill iTextOutputDeviceDrawType = iota
	iTextOutputDeviceDrawTypeStroke
	iTextOutputDeviceDrawTypeClip
)

var iTextOutputDeviceIdentity = geom.NewAffineTransform()

var iTextOutputDeviceStrokeOne = geom.NewBasicStrokeWithWidth(1)

var iTextOutputDeviceRoundRectDimensionsDown = ufo.ConfigurationIsTrue("xr.pdf.round.rect.dimensions.down", false)

var iTextOutputDeviceTaggableElements = map[string]writer.PdfName{
	"h1": writer.PdfNameH1,
	"h2": writer.PdfNameH2,
	"h3": writer.PdfNameH3,
	"h4": writer.PdfNameH4,
	"h5": writer.PdfNameH5,
	"h6": writer.PdfNameH6,
	"p":  writer.PdfNameP,
}

var iTextOutputDeviceListElements = map[string]writer.PdfName{
	"ul": writer.PdfNameL,
	"ol": writer.PdfNameL,
	"li": writer.PdfNameLi,
}

// No PdfName.COLSPAN/ROWSPAN constants exist in OpenPDF; these are the raw key names the PDF/UA
// Table attribute-owner dictionary expects (ISO 32000-2 §14.8.5.7).
var iTextOutputDeviceColspan = writer.NewPdfName("ColSpan")
var iTextOutputDeviceRowspan = writer.NewPdfName("RowSpan")

// No PdfName.ARTIFACT constant exists in OpenPDF either.
var iTextOutputDeviceArtifact = writer.NewPdfName("Artifact")

// ITextOutputDevice is largely based on OpenPDF's PdfGraphics2D. See
// http://sourceforge.net/projects/itext/ for license information.
//
// Tagged PDF: pdf/writer does not write a structure tree
// (writer.PdfWriter.IsTagged is always false), so the structure element
// methods below are never reached. They are ported against the writer's
// placeholder API and panic with the writer's *UnsupportedFeatureError as the
// cause if they are reached.
type ITextOutputDevice struct {
	ufo.AbstractOutputDevice

	currentPage *writer.PdfContentByte
	pageHeight  float32

	font *ITextFSFont

	transform      *geom.AffineTransform
	transformStack []*geom.AffineTransform

	color writer.Color

	fillColor   writer.Color
	strokeColor writer.Color

	stroke         geom.Stroke
	originalStroke geom.Stroke
	oldStroke      geom.Stroke

	opacity float32

	clip *geom.Area

	sharedContext *ufo.SharedContext
	dotsPerPoint  float32

	writer *writer.PdfWriter

	// readerCache is keyed by the string form of the URI.
	readerCache map[string]*writer.PdfReader

	defaultDestination *writer.PdfDestination

	bookmarks []*ITextOutputDeviceBookmark

	// metadata holds nil entries where SetMetadata removed a pair.
	metadata []*iTextOutputDeviceMetadata

	root ufo.BoxI

	startPageNo int

	nextFormFieldIndex int

	linkTargetAreas map[string]struct{}

	// structureElements is keyed by identity: a *dom.Element or a box.
	structureElements map[any]*writer.PdfStructureElement

	// A ListItem's own text is never marked directly against the LI structure element: OpenPDF requires a
	// structure element's /K entries to be either all marked-content references or all child structure
	// elements, never a mix — and an <li> can also host a nested <ul>/<ol> (a child L element). So the
	// item's text goes under its own LBody child instead, keeping LI's kids uniformly structural.
	listItemBodies map[*dom.Element]*writer.PdfStructureElement

	// Same constraint as above: a table cell can hold both its own text and a real child structure
	// element (an <img>'s Figure, added directly under the cell - see beginImageStructure). Marking text
	// directly against the cell's own structure element would make its /K entries a mix of marked-content
	// references and structure elements, which OpenPDF rejects. So cell text goes under a NonStruct child
	// instead, keeping the cell's own kids uniformly structural (this wrapper, plus any Figures).
	tableCellTextWrappers map[*ufo.TableCellBox]*writer.PdfStructureElement

	documentStructureElement *writer.PdfStructureElement

	// unsupportedWarned holds the features of pdf/writer's "Not supported"
	// list that a warning was already logged for in the current document.
	unsupportedWarned map[string]struct{}
}

func NewITextOutputDevice(dotsPerPoint float32) *ITextOutputDevice {
	d := &ITextOutputDevice{
		transform:             geom.NewAffineTransform(),
		color:                 writer.NewRGBColor(0, 0, 0),
		opacity:               1.0,
		dotsPerPoint:          dotsPerPoint,
		readerCache:           map[string]*writer.PdfReader{},
		linkTargetAreas:       map[string]struct{}{},
		structureElements:     map[any]*writer.PdfStructureElement{},
		listItemBodies:        map[*dom.Element]*writer.PdfStructureElement{},
		tableCellTextWrappers: map[*ufo.TableCellBox]*writer.PdfStructureElement{},
		unsupportedWarned:     map[string]struct{}{},
	}
	ufo.InitAbstractOutputDevice(&d.AbstractOutputDevice, d)
	return d
}

func (d *ITextOutputDevice) SetWriter(w *writer.PdfWriter) {
	if d.writer != w {
		// Cached structure elements belong to the previous writer's PDF; carrying them over would
		// reference objects from a different document when this device is reused across createPDF calls.
		clear(d.structureElements)
		d.documentStructureElement = nil
		clear(d.listItemBodies)
		clear(d.tableCellTextWrappers)
	}
	d.writer = w
}

// GetWriter returns nil before SetWriter is called.
func (d *ITextOutputDevice) GetWriter() *writer.PdfWriter {
	return d.writer
}

func (d *ITextOutputDevice) GetNextFormFieldIndex() int {
	d.nextFormFieldIndex++
	return d.nextFormFieldIndex
}

// warnUnsupportedOnce logs a warning that a feature of the Java
// implementation is skipped because pdf/writer does not support it. A feature
// is reported once per document: ITextReplacedElementFactory.Reset, which runs
// when a new document is loaded, empties the set of reported features.
func (d *ITextOutputDevice) warnUnsupportedOnce(err error, skipped string) {
	key := err.Error()
	if _, ok := d.unsupportedWarned[key]; ok {
		return
	}
	d.unsupportedWarned[key] = struct{}{}
	ufo.XRLogRender(ufo.LevelWarning, skipped+": "+err.Error())
}

func (d *ITextOutputDevice) resetUnsupportedWarnings() {
	clear(d.unsupportedWarned)
}

func (d *ITextOutputDevice) InitializePage(currentPage *writer.PdfContentByte, height float32) {
	d.currentPage = currentPage
	d.pageHeight = height

	d.currentPage.SaveState()

	d.transform = geom.NewAffineTransform()
	d.transform.Scale(1.0/float64(d.dotsPerPoint), 1.0/float64(d.dotsPerPoint))

	d.stroke = d.transformStroke(iTextOutputDeviceStrokeOne)
	d.originalStroke = d.stroke
	d.oldStroke = d.stroke

	d.setStrokeDiff(d.stroke, nil)

	if d.defaultDestination == nil {
		d.defaultDestination = writer.NewPdfDestinationWithParameter(writer.PdfDestinationFith, height)
		d.defaultDestination.AddPage(d.writer.GetPageReference(1))
	}

	clear(d.linkTargetAreas)
}

func (d *ITextOutputDevice) FinishPage() {
	d.currentPage.RestoreState()
}

func (d *ITextOutputDevice) PaintReplacedElement(c *ufo.RenderingContext, box ufo.BlockBoxI) {
	structElement := d.beginImageStructure(box)
	element := box.GetReplacedElement().(ITextReplacedElement)
	element.Paint(c, d, box)
	if structElement != nil {
		d.currentPage.EndMarkedContentSequence()
	}
}

func (d *ITextOutputDevice) DrawText(c *ufo.RenderingContext, inlineText *ufo.InlineText) {
	var structElement *writer.PdfStructureElement
	if d.isTagged() {
		structElement = d.beginTextStructure(inlineText)
	}
	d.AbstractOutputDevice.DrawText(c, inlineText)
	if structElement != nil {
		d.currentPage.EndMarkedContentSequence()
	}
}

func (d *ITextOutputDevice) isTagged() bool {
	return d.writer != nil && d.writer.IsTagged()
}

// newStructureElement is new PdfStructureElement(parent, tag).
func (d *ITextOutputDevice) newStructureElement(parent *writer.PdfStructureElement, tag writer.PdfName) *writer.PdfStructureElement {
	result, err := writer.NewPdfStructureElement(parent, tag)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(err.Error(), err))
	}
	return result
}

// beginMarkedContentSequence is
// PdfContentByte.beginMarkedContentSequence(PdfStructureElement).
func (d *ITextOutputDevice) beginMarkedContentSequence(structElement *writer.PdfStructureElement) {
	if err := d.currentPage.BeginMarkedContentSequenceStruct(structElement); err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(err.Error(), err))
	}
}

// beginTextStructure returns nil when the text is inside no taggable element.
func (d *ITextOutputDevice) beginTextStructure(inlineText *ufo.InlineText) *writer.PdfStructureElement {
	var box ufo.BoxI
	if parent := inlineText.GetParent(); parent != nil {
		box = parent
	}

	// All text anywhere inside a table cell is associated with that cell (rather than tagging a
	// <p>/<h1> nested inside it as its own child structure element), so headings/paragraphs/lists lose
	// their own tag within a cell but always correctly nest under the required Table/TR/TD hierarchy
	// instead of leaking out to the flat Document node. The text itself goes under the cell's NonStruct
	// wrapper (see tableCellTextStructureElementFor), not the cell's own structure element directly,
	// to stay homogeneous with any Figure children the cell might also have (see beginImageStructure).
	cell := iTextOutputDeviceFindNearestTableCell(box)
	if cell != nil {
		structElement := d.tableCellTextStructureElementFor(cell)
		d.beginMarkedContentSequence(structElement)
		return structElement
	}

	for box != nil {
		element := box.GetElement()
		if element != nil {
			tagName := strings.ToLower(element.GetNodeName())
			flatTag, ok := iTextOutputDeviceTaggableElements[tagName]
			if ok {
				structElement := d.structureElementFor(element, flatTag, d.getDocumentStructureElement())
				d.beginMarkedContentSequence(structElement)
				return structElement
			}
			listTag, ok := iTextOutputDeviceListElements[tagName]
			if ok {
				var structElement *writer.PdfStructureElement
				if listTag == writer.PdfNameLi {
					structElement = d.listItemBodyStructureElementFor(element)
				} else {
					structElement = d.listStructureElementFor(element, listTag)
				}
				d.beginMarkedContentSequence(structElement)
				return structElement
			}
			if "a" == tagName && d.sharedContext.GetNamespaceHandler().GetLinkUri(element) != nil {
				structElement := d.structureElementFor(element, writer.PdfNameLink, d.getDocumentStructureElement())
				d.beginMarkedContentSequence(structElement)
				return structElement
			}
		}
		box = box.GetParent()
	}
	return nil
}

// iTextOutputDeviceFindNearestTableCell returns nil when box is in no table
// cell.
func iTextOutputDeviceFindNearestTableCell(box ufo.BoxI) *ufo.TableCellBox {
	for box != nil {
		if cell, ok := box.(*ufo.TableCellBox); ok {
			return cell
		}
		box = box.GetParent()
	}
	return nil
}

func (d *ITextOutputDevice) tableCellTextStructureElementFor(cell *ufo.TableCellBox) *writer.PdfStructureElement {
	result, ok := d.tableCellTextWrappers[cell]
	if !ok {
		result = d.newStructureElement(d.tableStructureElementFor(cell), writer.PdfNameNonstruct)
		d.tableCellTextWrappers[cell] = result
	}
	return result
}

func (d *ITextOutputDevice) listItemBodyStructureElementFor(liElement *dom.Element) *writer.PdfStructureElement {
	result, ok := d.listItemBodies[liElement]
	if !ok {
		result = d.newStructureElement(d.listStructureElementFor(liElement, writer.PdfNameLi), writer.PdfNameLbody)
		d.listItemBodies[liElement] = result
	}
	return result
}

func (d *ITextOutputDevice) listStructureElementFor(element *dom.Element, tag writer.PdfName) *writer.PdfStructureElement {
	cached := d.structureElements[element]
	if cached != nil {
		return cached
	}
	parentElement := iTextOutputDeviceNearestListAncestor(element.GetParentNode())
	var parent *writer.PdfStructureElement
	if parentElement != nil {
		parent = d.listStructureElementFor(parentElement, iTextOutputDeviceListElements[strings.ToLower(parentElement.GetNodeName())])
	} else {
		parent = d.getDocumentStructureElement()
	}
	return d.structureElementFor(element, tag, parent)
}

// iTextOutputDeviceNearestListAncestor returns nil when node has no list
// element among itself and its ancestors.
func iTextOutputDeviceNearestListAncestor(node dom.Node) *dom.Element {
	for {
		element, ok := node.(*dom.Element)
		if !ok || element == nil {
			return nil
		}
		if _, isList := iTextOutputDeviceListElements[strings.ToLower(element.GetNodeName())]; isList {
			return element
		}
		node = element.GetParentNode()
	}
}

// beginImageStructure returns nil when the document is not tagged or the box
// is not an image that is exposed as a Figure.
func (d *ITextOutputDevice) beginImageStructure(box ufo.BlockBoxI) *writer.PdfStructureElement {
	if !d.isTagged() {
		return nil
	}
	element := box.GetElement()
	if element == nil || !strings.EqualFold("img", element.GetNodeName()) {
		return nil
	}
	// alt="" marks the image as decorative; don't expose it to assistive technology as a Figure.
	if element.HasAttribute("alt") && element.GetAttribute("alt") == "" {
		return nil
	}
	structElement := d.structureElementFor(element, writer.PdfNameFigure, d.tableAncestorStructureElement(box))
	if element.HasAttribute("alt") {
		structElement.Put(writer.PdfNameAlt, writer.NewPdfString(element.GetAttribute("alt")))
	}
	d.beginMarkedContentSequence(structElement)
	return structElement
}

func (d *ITextOutputDevice) tableAncestorStructureElement(box ufo.BoxI) *writer.PdfStructureElement {
	ancestor := box.GetParent()
	for ancestor != nil {
		if cell, ok := ancestor.(*ufo.TableCellBox); ok {
			return d.tableStructureElementFor(cell)
		}
		ancestor = ancestor.GetParent()
	}
	return d.getDocumentStructureElement()
}

func (d *ITextOutputDevice) tableStructureElementFor(box ufo.BlockBoxI) *writer.PdfStructureElement {
	cached := d.structureElements[box]
	if cached != nil {
		return cached
	}
	parentBox := box.GetParent()
	var parent *writer.PdfStructureElement
	if iTextOutputDeviceIsTableBox(parentBox) {
		parent = d.tableStructureElementFor(parentBox.(ufo.BlockBoxI))
	} else {
		parent = d.getDocumentStructureElement()
	}
	structElement := d.structureElementFor(box, iTextOutputDeviceTableTag(box), parent)
	if cell, ok := box.(*ufo.TableCellBox); ok {
		iTextOutputDeviceAddCellSpanAttributes(structElement, cell)
	}
	return structElement
}

func iTextOutputDeviceAddCellSpanAttributes(structElement *writer.PdfStructureElement, cell *ufo.TableCellBox) {
	colSpan := cell.GetStyle().GetColSpan()
	rowSpan := cell.GetStyle().GetRowSpan()
	if colSpan <= 1 && rowSpan <= 1 {
		return
	}
	attributes := writer.NewPdfDictionary()
	attributes.Put(writer.PdfNameO, writer.PdfNameTable)
	if colSpan > 1 {
		attributes.Put(iTextOutputDeviceColspan, writer.NewPdfNumber(float64(colSpan)))
	}
	if rowSpan > 1 {
		attributes.Put(iTextOutputDeviceRowspan, writer.NewPdfNumber(float64(rowSpan)))
	}
	structElement.Put(writer.PdfNameA, attributes)
}

func iTextOutputDeviceIsTableBox(box ufo.BoxI) bool {
	switch box.(type) {
	case *ufo.TableBox, *ufo.TableSectionBox, *ufo.TableRowBox, *ufo.TableCellBox:
		return true
	}
	return false
}

func iTextOutputDeviceTableTag(box ufo.BlockBoxI) writer.PdfName {
	switch b := box.(type) {
	case *ufo.TableBox:
		return writer.PdfNameTable
	case *ufo.TableSectionBox:
		if b.IsHeader() {
			return writer.PdfNameThead
		} else if b.IsFooter() {
			return writer.PdfNameTfoot
		}
		return writer.PdfNameTbody
	case *ufo.TableRowBox:
		return writer.PdfNameTablerow
	case *ufo.TableCellBox:
		element := b.GetElement()
		if element != nil && strings.EqualFold("th", element.GetNodeName()) {
			return writer.PdfNameTh
		}
		return writer.PdfNameTd
	}
	panic(ufo.NewXRRuntimeException(fmt.Sprintf("Not a table box: %v", box)))
}

func (d *ITextOutputDevice) structureElementFor(key any, tag writer.PdfName, parent *writer.PdfStructureElement) *writer.PdfStructureElement {
	result, ok := d.structureElements[key]
	if !ok {
		result = d.newStructureElement(parent, tag)
		d.structureElements[key] = result
	}
	return result
}

// getDocumentStructureElement is the Java method documentStructureElement();
// the field has that name here.
func (d *ITextOutputDevice) getDocumentStructureElement() *writer.PdfStructureElement {
	if d.documentStructureElement == nil {
		root, err := d.writer.GetStructureTreeRoot()
		if err != nil {
			panic(ufo.NewXRRuntimeExceptionWithCause(err.Error(), err))
		}
		result, err := writer.NewPdfStructureElementWithRoot(root, writer.PdfNameDocument)
		if err != nil {
			panic(ufo.NewXRRuntimeExceptionWithCause(err.Error(), err))
		}
		d.documentStructureElement = result
	}
	return d.documentStructureElement
}

func (d *ITextOutputDevice) PaintBackground(c *ufo.RenderingContext, box ufo.BoxI) {
	d.paintAsArtifact(func() { d.AbstractOutputDevice.PaintBackground(c, box) })

	d.processLink(c, box)
}

func (d *ITextOutputDevice) PaintBackgroundWithStyleBoundsBgImageContainerBorder(
	c *ufo.RenderingContext, style ufo.CalculatedStyleI,
	bounds *geom.Rectangle, bgImageContainer *geom.Rectangle,
	border *ufo.BorderPropertySet) {
	d.paintAsArtifact(func() {
		d.AbstractOutputDevice.PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, style, bounds, bgImageContainer, border)
	})
}

func (d *ITextOutputDevice) PaintBorder(c *ufo.RenderingContext, box ufo.BoxI) {
	d.paintAsArtifact(func() { d.AbstractOutputDevice.PaintBorder(c, box) })
}

func (d *ITextOutputDevice) PaintBorderWithStyleEdgeSides(c *ufo.RenderingContext, style ufo.CalculatedStyleI, edge *geom.Rectangle, sides int) {
	d.paintAsArtifact(func() { d.AbstractOutputDevice.PaintBorderWithStyleEdgeSides(c, style, edge, sides) })
}

func (d *ITextOutputDevice) PaintCollapsedBorder(c *ufo.RenderingContext, border *ufo.BorderPropertySet, bounds *geom.Rectangle, side int) {
	d.paintAsArtifact(func() { d.AbstractOutputDevice.PaintCollapsedBorder(c, border, bounds, side) })
}

// Decorative chrome (backgrounds/borders) is marked as an Artifact rather than left untagged, so
// assistive technology explicitly skips it instead of treating it as unmarked/ambiguous content.
func (d *ITextOutputDevice) paintAsArtifact(painter func()) {
	if d.isTagged() {
		d.currentPage.BeginMarkedContentSequence(iTextOutputDeviceArtifact)
		painter()
		d.currentPage.EndMarkedContentSequence()
	} else {
		painter()
	}
}

func (d *ITextOutputDevice) calcTotalLinkArea(c *ufo.RenderingContext, box ufo.BoxI) *writer.Rectangle {
	current := box
	for {
		prev := current.GetPreviousSibling()
		if prev == nil || prev.GetElement() != box.GetElement() {
			break
		}

		current = prev
	}

	result := d.createLocalTargetAreaWithUseAggregateBounds(c, current, true)

	current = current.GetNextSibling()
	for current != nil && current.GetElement() == box.GetElement() {
		result = d.add(result, d.createLocalTargetAreaWithUseAggregateBounds(c, current, true))

		current = current.GetNextSibling()
	}

	return result
}

func (d *ITextOutputDevice) add(r1 *writer.Rectangle, r2 *writer.Rectangle) *writer.Rectangle {
	llx := min(r1.GetLeft(), r2.GetLeft())
	urx := max(r1.GetRight(), r2.GetRight())
	lly := min(r1.GetBottom(), r2.GetBottom())
	ury := max(r1.GetTop(), r2.GetTop())

	return writer.NewRectangle(llx, lly, urx, ury)
}

// createRectKey builds the key that tells link areas apart. Java joins the
// Float.toString forms of the four numbers; the key is only compared with
// other keys made here, so the shortest decimal form that identifies each
// float32 serves the same purpose.
func (d *ITextOutputDevice) createRectKey(rect *writer.Rectangle) string {
	return iTextOutputDeviceFloatToString(rect.GetLeft()) + ":" + iTextOutputDeviceFloatToString(rect.GetBottom()) + ":" +
		iTextOutputDeviceFloatToString(rect.GetRight()) + ":" + iTextOutputDeviceFloatToString(rect.GetTop())
}

func iTextOutputDeviceFloatToString(f float32) string {
	return strconv.FormatFloat(float64(f), 'g', -1, 32)
}

// checkLinkArea returns nil when a link annotation was already added for the
// same area on the current page.
func (d *ITextOutputDevice) checkLinkArea(c *ufo.RenderingContext, box ufo.BoxI) *writer.Rectangle {
	targetArea := d.calcTotalLinkArea(c, box)
	key := d.createRectKey(targetArea)
	if _, ok := d.linkTargetAreas[key]; ok {
		return nil
	}
	d.linkTargetAreas[key] = struct{}{}
	return targetArea
}

func (d *ITextOutputDevice) processLink(c *ufo.RenderingContext, box ufo.BoxI) {
	elem := box.GetElement()
	if elem != nil {
		handler := d.sharedContext.GetNamespaceHandler()
		uriValue := handler.GetLinkUri(elem)
		if uriValue != nil {
			uri := *uriValue
			if len(uri) > 1 && uri[0] == '#' {
				anchor := uri[1:]
				target := d.sharedContext.GetBoxById(anchor)
				if target != nil {
					dest := d.createDestination(c, target)

					if dest != nil {
						var action *writer.PdfAction
						// Java's NamespaceHandler returns "" for a missing attribute; the Go
						// interface returns nil for one.
						onclick := handler.GetAttributeValue(elem, "onclick")
						if onclick == nil || *onclick == "" {
							action = iTextOutputDeviceGotoDestination(dest)
						} else {
							action = writer.PdfActionJavaScript(*onclick, d.writer)
						}

						if targetArea := d.checkLinkArea(c, box); targetArea != nil {
							targetArea.SetBorder(0)
							targetArea.SetBorderWidth(0)

							d.addLinkAnnotation(action, targetArea)
						}
					}
				}
			} else {
				action := writer.NewPdfActionWithUrl(uri)
				if targetArea := d.checkLinkArea(c, box); targetArea != nil {
					d.addLinkAnnotation(action, targetArea)
				}
			}
		}
	}
}

func iTextOutputDeviceGotoDestination(dest *writer.PdfDestination) *writer.PdfAction {
	action := writer.NewPdfAction()
	action.Put(writer.PdfNameS, writer.PdfNameGoto)
	action.Put(writer.PdfNameD, dest)
	return action
}

func (d *ITextOutputDevice) addLinkAnnotation(action *writer.PdfAction, targetArea *writer.Rectangle) {
	annot := writer.NewPdfAnnotation(d.writer, targetArea.GetLeft(), targetArea.GetBottom(),
		targetArea.GetRight(), targetArea.GetTop(), action)
	annot.Put(writer.PdfNameSubtype, writer.PdfNameLink)
	annot.Put(writer.PdfNameF, writer.NewPdfNumber(writer.PdfAnnotationFlagsPrint))

	annot.SetBorderStyle(writer.NewPdfBorderDictionary(0.0, 0))
	annot.SetBorder(writer.NewPdfBorderArray(0.0, 0.0, 0))
	d.writer.AddAnnotation(annot)
}

func (d *ITextOutputDevice) CreateLocalTargetArea(c *ufo.RenderingContext, box ufo.BoxI) *writer.Rectangle {
	return d.createLocalTargetAreaWithUseAggregateBounds(c, box, false)
}

func (d *ITextOutputDevice) createLocalTargetAreaWithUseAggregateBounds(c *ufo.RenderingContext, box ufo.BoxI, useAggregateBounds bool) *writer.Rectangle {
	var bounds *geom.Rectangle
	if useAggregateBounds && box.GetPaintingInfo() != nil {
		bounds = box.GetPaintingInfo().GetAggregateBounds()
	} else {
		bounds = box.GetContentAreaEdge(box.GetAbsX(), box.GetAbsY(), c)
	}

	// Transform all four corners (not just one, plus untransformed width/height) so a rotated or
	// skewed transform still produces the correct axis-aligned bounding box in PDF space.
	docCorners := []*geom.Point2D{
		geom.NewPoint2DDouble(float64(bounds.X), float64(bounds.Y)),
		geom.NewPoint2DDouble(float64(bounds.X+bounds.Width), float64(bounds.Y)),
		geom.NewPoint2DDouble(float64(bounds.X), float64(bounds.Y+bounds.Height)),
		geom.NewPoint2DDouble(float64(bounds.X+bounds.Width), float64(bounds.Y+bounds.Height)),
	}

	var minX, maxX float32 = math.MaxFloat32, -math.MaxFloat32
	var minY, maxY float32 = math.MaxFloat32, -math.MaxFloat32
	for _, docCorner := range docCorners {
		pdfCorner := d.transform.Transform(docCorner, nil)
		x := float32(pdfCorner.GetX())
		y := d.normalizeY(float32(pdfCorner.GetY()))
		minX = min(minX, x)
		maxX = max(maxX, x)
		minY = min(minY, y)
		maxY = max(maxY, y)
	}

	return writer.NewRectangle(minX, minY, maxX, maxY)
}

func (d *ITextOutputDevice) CreateTargetArea(c *ufo.RenderingContext, box ufo.BoxI) *writer.Rectangle {
	current := c.GetPage()
	inCurrentPage := box.GetAbsY() > current.GetTop() && box.GetAbsY() < current.GetBottom()

	if inCurrentPage || box.IsContainedInMarginBox() {
		return d.CreateLocalTargetArea(c, box)
	} else {
		bounds := box.GetContentAreaEdge(box.GetAbsX(), box.GetAbsY(), c)
		page := d.root.GetLayer().GetPage(c, bounds.Y)

		bottom := d.GetDeviceLength(float32(page.GetBottom() - (bounds.Y + bounds.Height) +
			page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeBottom)))
		left := d.GetDeviceLength(float32(page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeLeft) + bounds.X))

		return writer.NewRectangle(left, bottom, left+d.GetDeviceLength(float32(bounds.Width)), bottom+
			d.GetDeviceLength(float32(bounds.Height)))
	}
}

func (d *ITextOutputDevice) GetDeviceLength(length float32) float32 {
	return length / d.dotsPerPoint
}

// createDestination returns nil when the box is on no page.
func (d *ITextOutputDevice) createDestination(c *ufo.RenderingContext, box ufo.BoxI) *writer.PdfDestination {
	var result *writer.PdfDestination

	page := d.root.GetLayer().GetPage(c, d.getPageRefY(box))
	if page != nil {
		distanceFromTop := page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeTop)
		distanceFromTop += int(float32(box.GetAbsY()) + box.GetMargin(c).Top() - float32(page.GetTop()))
		result = writer.NewPdfDestinationWithLeftTopZoom(writer.PdfDestinationXyz, 0, float32(page.GetHeight(c))/d.dotsPerPoint-float32(distanceFromTop)/d.dotsPerPoint, 0)
		result.AddPage(d.writer.GetPageReference(d.startPageNo + page.GetPageNo() + 1))
	}

	return result
}

func (d *ITextOutputDevice) DrawBorderLine(bounds geom.Shape, side int, lineWidth int, solid bool) {
	/*( float x = bounds.x;
	  float y = bounds.y;
	  float w = bounds.width;
	  float h = bounds.height;

	  float adj = solid ? (float) lineWidth / 2 : 0;
	  float adj2 = lineWidth % 2 != 0 ? 0.5f : 0f;

	  Line2D.Float line = null;

	  // FIXME: findbugs reports possible loss of precision, compare with
	  // width / (float)2
	  if (side == BorderPainter.TOP) {
	      line = new Line2D.Float(x + adj, y + lineWidth / 2 + adj2, x + w - adj, y + lineWidth / 2 + adj2);
	  } else if (side == BorderPainter.LEFT) {
	      line = new Line2D.Float(x + lineWidth / 2 + adj2, y + adj, x + lineWidth / 2 + adj2, y + h - adj);
	  } else if (side == BorderPainter.RIGHT) {
	      float offset = lineWidth / 2;
	      if (lineWidth % 2 != 0) {
	          offset += 1;
	      }
	      line = new Line2D.Float(x + w - offset + adj2, y + adj, x + w - offset + adj2, y + h - adj);
	  } else if (side == BorderPainter.BOTTOM) {
	      float offset = lineWidth / 2;
	      if (lineWidth % 2 != 0) {
	          offset += 1;
	      }
	      line = new Line2D.Float(x + adj, y + h - offset + adj2, x + w - adj, y + h - offset + adj2);
	  }*/

	d.Draw(bounds)
}

func (d *ITextOutputDevice) SetOpacity(opacity float32) {
	if opacity != d.opacity {
		gs := writer.NewPdfGState()

		gs.SetBlendMode(writer.PdfGStateBmNormal)
		gs.SetFillOpacity(opacity)

		d.currentPage.SetGState(gs)
		d.opacity = opacity
	}
}

func (d *ITextOutputDevice) SetColor(color ufo.FSColor) {
	switch c := color.(type) {
	case *ufo.FSRGBColor:
		d.color = writer.NewRGBColorWithAlpha(c.GetRed(), c.GetGreen(), c.GetBlue(), int(c.GetAlpha()*255))
	case *ufo.FSCMYKColor:
		d.color = writer.NewCMYKColor(c.GetCyan(), c.GetMagenta(), c.GetYellow(), c.GetBlack())
	default:
		panic(ufo.NewXRRuntimeException(fmt.Sprintf("internal error: unsupported color class %T", color)))
	}
}

func (d *ITextOutputDevice) Draw(s geom.Shape) {
	d.followPath(s, iTextOutputDeviceDrawTypeStroke)
}

// DrawLine is the protected drawLine that AbstractOutputDevice declares
// abstract.
func (d *ITextOutputDevice) DrawLine(x1 int, y1 int, x2 int, y2 int) {
	line := geom.NewLine2DDouble(float64(x1), float64(y1), float64(x2), float64(y2))
	d.Draw(line)
}

func (d *ITextOutputDevice) DrawRect(x int, y int, width int, height int) {
	d.Draw(geom.NewRectangle(x, y, width, height))
}

func (d *ITextOutputDevice) DrawOval(x int, y int, width int, height int) {
	oval := geom.NewEllipse2DFloat(float32(x), float32(y), float32(width), float32(height))
	d.Draw(oval)
}

func (d *ITextOutputDevice) Fill(s geom.Shape) {
	d.followPath(s, iTextOutputDeviceDrawTypeFill)
}

func (d *ITextOutputDevice) FillRect(x int, y int, width int, height int) {
	if iTextOutputDeviceRoundRectDimensionsDown {
		d.Fill(geom.NewRectangle(x, y, width-1, height-1))
	} else {
		d.Fill(geom.NewRectangle(x, y, width, height))
	}
}

func (d *ITextOutputDevice) FillOval(x int, y int, width int, height int) {
	oval := geom.NewEllipse2DFloat(float32(x), float32(y), float32(width), float32(height))
	d.Fill(oval)
}

func (d *ITextOutputDevice) Translate(tx float64, ty float64) {
	d.transform.Translate(tx, ty)
}

func (d *ITextOutputDevice) PushTransform(c *ufo.RenderingContext, box ufo.BoxI) {
	d.transformStack = append(d.transformStack, d.transform.Clone())

	boxTransform := CssTransformToAffineTransform(c, box)
	if boxTransform != nil {
		d.transform.Concatenate(boxTransform)
	}
}

func (d *ITextOutputDevice) PopTransform() {
	last := len(d.transformStack) - 1
	d.transform = d.transformStack[last]
	d.transformStack = d.transformStack[:last]
}

// getRenderingHint and setRenderingHint are not ported: see ufo.OutputDevice.

// SetFont takes the *ITextFSFont of this module; the Java method is typed
// with it.
func (d *ITextOutputDevice) SetFont(font ufo.FSFont) {
	if font == nil {
		d.font = nil
		return
	}
	d.font = font.(*ITextFSFont)
}

func (d *ITextOutputDevice) normalizeMatrix(current *geom.AffineTransform) *geom.AffineTransform {
	mx := make([]float64, 6)
	result := geom.NewAffineTransform()
	result.GetMatrix(mx)
	mx[3] = -1
	mx[5] = float64(d.pageHeight)
	result = geom.NewAffineTransformWithFlatmatrix(mx)
	result.Concatenate(current)
	return result
}

// DrawString draws s with the current font and color. info is nil for text
// that is not justified.
func (d *ITextOutputDevice) DrawString(s string, x float32, y float32, info *ufo.JustificationInfo) {
	if ufo.ConfigurationIsTrue("xr.renderer.replace-missing-characters", false) {
		s = d.replaceMissingCharacters(s)
	}
	if s == "" {
		return
	}
	cb := d.currentPage
	d.ensureFillColor()
	at := d.GetTransform().Clone()
	at.Translate(float64(x), float64(y))
	inverse := d.normalizeMatrix(at)
	flipper := geom.AffineTransformGetScaleInstance(1, -1)
	inverse.Concatenate(flipper)
	inverse.Scale(float64(d.dotsPerPoint), float64(d.dotsPerPoint))
	mx := make([]float64, 6)
	inverse.GetMatrix(mx)
	cb.BeginText()
	// Check if bold or italic need to be emulated
	resetMode := false
	desc := d.font.GetFontDescription()
	fontSize := d.font.GetSize2D() / d.dotsPerPoint
	cb.SetFontAndSize(desc.GetFont(), fontSize)
	b := float32(mx[1])
	c := float32(mx[2])
	fontSpec := d.GetFontSpecification()
	if fontSpec != nil {
		need := ITextFontResolverConvertWeightToInt(fontSpec.FontWeight())
		have := desc.GetWeight()

		if need > have {
			cb.SetTextRenderingMode(writer.PdfContentByteTextRenderModeFillStroke)
			lineWidth := fontSize * 0.04 // 4% of font size
			cb.SetLineWidth(lineWidth)
			resetMode = true
			d.ensureStrokeColor()
		}
		if fontSpec.FontStyle() == ufo.IdentValueItalic && desc.GetStyle() != ufo.IdentValueItalic && desc.GetStyle() != ufo.IdentValueOblique {
			b = 0.0
			c = 0.21256
		}
	}
	cb.SetTextMatrix(float32(mx[0]), b, c, float32(mx[3]), float32(mx[4]), float32(mx[5]))
	if info == nil {
		cb.ShowText(s)
	} else {
		array := d.makeJustificationArray(s, info)
		cb.ShowTextArray(array)
	}
	if resetMode {
		cb.SetTextRenderingMode(writer.PdfContentByteTextRenderModeFill)
		cb.SetLineWidth(1)
	}
	cb.EndText()
}

func (d *ITextOutputDevice) replaceMissingCharacters(s string) string {
	charArr := []rune(s)
	replacementCharacter := ufo.ConfigurationValueAsChar("xr.renderer.missing-character-replacement", '#')

	// first check to see if the replacement character even exists in the
	// given font. If not, then do nothing.
	if !d.font.GetFontDescription().GetFont().CharExists(replacementCharacter) {
		ufo.XRLogRender(ufo.LevelInfo, "Missing replacement character ["+string(replacementCharacter)+":"+strconv.Itoa(int(replacementCharacter))+
			"]. No replacement will occur.")
		return s
	}

	// iterate through each character in the string and make an appropriate
	// replacement
	for i := range charArr {
		if !(charArr[i] == ' ' || charArr[i] == ' ' || charArr[i] == '　' || d.font.GetFontDescription().GetFont().
			CharExists(charArr[i])) {
			ufo.XRLogRender(ufo.LevelInfo, "Missing character ["+string(charArr[i])+":"+strconv.Itoa(int(charArr[i]))+"] in string ["+s+
				"]. Replacing with '"+string(replacementCharacter)+"'")
			charArr[i] = replacementCharacter
		}
	}

	return string(charArr)
}

func (d *ITextOutputDevice) makeJustificationArray(s string, info *ufo.JustificationInfo) *writer.PdfTextArray {
	array := writer.NewPdfTextArray()
	length := len(s)
	for i := 0; i < length; {
		c, size := utf8.DecodeRuneInString(s[i:])
		end := i + size
		array.AddString(s[i:end])
		if end != length {
			var offset float32
			if c == ' ' || c == ' ' || c == '　' {
				offset = info.SpaceAdjust()
			} else {
				offset = info.NonSpaceAdjust()
			}
			array.AddNumber(-offset / d.dotsPerPoint * 1000 / (d.font.GetSize2D() / d.dotsPerPoint))
		}
		i = end
	}
	return array
}

func (d *ITextOutputDevice) GetTransform() *geom.AffineTransform {
	return d.transform
}

// iTextOutputDeviceColorEquals is Color.equals as OpenPDF's colors define it.
// java.awt.Color.equals compares the packed RGB and alpha values of any two
// colors, so an RGB receiver equals a CMYKColor argument whose RGB
// approximation (the one ExtendedColor passes to the Color constructor) is
// the same. CMYKColor.equals compares the four components and accepts only
// another CMYKColor.
func iTextOutputDeviceColorEquals(receiver writer.Color, other writer.Color) bool {
	if receiver == nil || other == nil {
		return false
	}
	switch a := receiver.(type) {
	case *writer.CMYKColor:
		b, ok := other.(*writer.CMYKColor)
		return ok && a.Cyan == b.Cyan && a.Magenta == b.Magenta && a.Yellow == b.Yellow && a.Black == b.Black
	case *writer.RGBColor:
		switch b := other.(type) {
		case *writer.RGBColor:
			return *a == *b
		case *writer.CMYKColor:
			component := func(v float32) int {
				return int(min(max(v, 0), 1)*255 + 0.5)
			}
			return a.Alpha == 255 &&
				a.Red == component(1-b.Cyan-b.Black) &&
				a.Green == component(1-b.Magenta-b.Black) &&
				a.Blue == component(1-b.Yellow-b.Black)
		}
	}
	return false
}

func (d *ITextOutputDevice) ensureFillColor() {
	if !iTextOutputDeviceColorEquals(d.color, d.fillColor) {
		d.fillColor = d.color
		d.currentPage.SetColorFill(d.fillColor)

		if d.fillColor.GetAlpha() < 255 {
			d.SetOpacity(float32(d.fillColor.GetAlpha()) / 255.0)
		}
	}
}

func (d *ITextOutputDevice) ensureStrokeColor() {
	if !iTextOutputDeviceColorEquals(d.color, d.strokeColor) {
		d.strokeColor = d.color
		d.currentPage.SetColorStroke(d.strokeColor)
	}
}

func (d *ITextOutputDevice) GetCurrentPage() *writer.PdfContentByte {
	return d.currentPage
}

func (d *ITextOutputDevice) followPath(s geom.Shape, drawType iTextOutputDeviceDrawType) {
	cb := d.currentPage

	if drawType == iTextOutputDeviceDrawTypeStroke {
		if _, ok := d.stroke.(*geom.BasicStroke); !ok {
			s = d.stroke.CreateStrokedShape(s)
			d.followPath(s, iTextOutputDeviceDrawTypeFill)
			return
		}
	}
	if drawType == iTextOutputDeviceDrawTypeStroke {
		d.setStrokeDiff(d.stroke, d.oldStroke)
		d.oldStroke = d.stroke
		d.ensureStrokeColor()
	} else if drawType == iTextOutputDeviceDrawTypeFill {
		d.ensureFillColor()
	}

	if drawType == iTextOutputDeviceDrawTypeClip {
		// A geom.Area is the intersection of its member shapes and has no
		// single outline. Successive clip operators intersect, so clipping to
		// each member in turn clips to the area.
		if area, ok := s.(*geom.Area); ok {
			for _, member := range area.Shapes() {
				d.followPath(member, iTextOutputDeviceDrawTypeClip)
			}
			return
		}
	}

	if area, ok := s.(*geom.Area); ok && drawType != iTextOutputDeviceDrawTypeClip && !area.IsRectangular() && len(area.Shapes()) > 1 {
		// Java's Area computes the outline of the intersection; geom.Area
		// keeps the member shapes (see its doc). The intersection is painted
		// by clipping to every member but the first, within a saved graphics
		// state, and painting the first. The color was set above, outside the
		// saved state, so it stays in effect after the restore.
		members := area.Shapes()
		cb.SaveState()
		for _, member := range members[1:] {
			d.followPath(d.transform.CreateTransformedShape(member), iTextOutputDeviceDrawTypeClip)
		}
		d.followPath(members[0], drawType)
		cb.RestoreState()
		return
	}

	var points geom.PathIterator
	if drawType == iTextOutputDeviceDrawTypeClip {
		points = s.GetPathIterator(iTextOutputDeviceIdentity)
	} else {
		points = s.GetPathIterator(d.transform)
	}
	coords := make([]float32, 6)
	traces := 0
	for !points.IsDone() {
		traces++
		segmentType := points.CurrentSegmentFloat(coords)
		d.normalizeYCoords(coords)
		switch segmentType {
		case geom.PathIteratorSegClose:
			cb.ClosePath()

		case geom.PathIteratorSegCubicto:
			cb.CurveTo(coords[0], coords[1], coords[2], coords[3], coords[4], coords[5])

		case geom.PathIteratorSegLineto:
			cb.LineTo(coords[0], coords[1])

		case geom.PathIteratorSegMoveto:
			cb.MoveTo(coords[0], coords[1])

		case geom.PathIteratorSegQuadto:
			cb.CurveToWithX2Y2X3Y3(coords[0], coords[1], coords[2], coords[3])
		}
		points.Next()
	}

	switch drawType {
	case iTextOutputDeviceDrawTypeFill:
		if traces > 0 {
			if points.GetWindingRule() == geom.PathIteratorWindEvenOdd {
				cb.EoFill()
			} else {
				cb.Fill()
			}
		}
	case iTextOutputDeviceDrawTypeStroke:
		if traces > 0 {
			cb.Stroke()
		}
	default: // drawType==CLIP
		if traces == 0 {
			cb.Rectangle(0, 0, 0, 0)
		}
		if points.GetWindingRule() == geom.PathIteratorWindEvenOdd {
			cb.EoClip()
		} else {
			cb.Clip()
		}
		cb.NewPath()
	}
}

func (d *ITextOutputDevice) normalizeY(y float32) float32 {
	return d.pageHeight - y
}

func (d *ITextOutputDevice) normalizeYCoords(coords []float32) {
	coords[1] = d.normalizeY(coords[1])
	coords[3] = d.normalizeY(coords[3])
	coords[5] = d.normalizeY(coords[5])
}

// setStrokeDiff writes the parts of newStroke that differ from oldStroke.
// oldStroke is nil when the stroke in effect is not known.
func (d *ITextOutputDevice) setStrokeDiff(newStroke geom.Stroke, oldStroke geom.Stroke) {
	cb := d.currentPage
	if newStroke == oldStroke {
		return
	}
	nStroke, ok := newStroke.(*geom.BasicStroke)
	if !ok {
		return
	}
	oStroke, oldOk := oldStroke.(*geom.BasicStroke)
	if !oldOk || nStroke.GetLineWidth() != oStroke.GetLineWidth() {
		cb.SetLineWidth(nStroke.GetLineWidth())
	}
	if !oldOk || nStroke.GetEndCap() != oStroke.GetEndCap() {
		switch nStroke.GetEndCap() {
		case geom.BasicStrokeCapButt:
			cb.SetLineCap(0)
		case geom.BasicStrokeCapSquare:
			cb.SetLineCap(2)
		default:
			cb.SetLineCap(1)
		}
	}
	if !oldOk || nStroke.GetLineJoin() != oStroke.GetLineJoin() {
		switch nStroke.GetLineJoin() {
		case geom.BasicStrokeJoinMiter:
			cb.SetLineJoin(0)
		case geom.BasicStrokeJoinBevel:
			cb.SetLineJoin(2)
		default:
			cb.SetLineJoin(1)
		}
	}
	if !oldOk || nStroke.GetMiterLimit() != oStroke.GetMiterLimit() {
		cb.SetMiterLimit(nStroke.GetMiterLimit())
	}
	makeDash := d.isMakeDash(oldOk, nStroke, oStroke)
	if makeDash {
		dash := nStroke.GetDashArray()
		if dash == nil {
			cb.SetLiteral("[]0 d\n")
		} else {
			cb.SetLiteralChar('[')
			for _, v := range dash {
				cb.SetLiteralFloat(v)
				cb.SetLiteralChar(' ')
			}
			cb.SetLiteralChar(']')
			cb.SetLiteralFloat(nStroke.GetDashPhase())
			cb.SetLiteral(" d\n")
		}
	}
}

func (d *ITextOutputDevice) isMakeDash(oldOk bool, nStroke *geom.BasicStroke, oStroke *geom.BasicStroke) bool {
	if oldOk {
		if nStroke.GetDashArray() != nil {
			return nStroke.GetDashPhase() != oStroke.GetDashPhase() ||
				!iTextOutputDeviceDashEquals(nStroke.GetDashArray(), oStroke.GetDashArray())
		} else {
			return oStroke.GetDashArray() != nil
		}
	} else {
		return true
	}
}

// iTextOutputDeviceDashEquals is Arrays.equals(float[], float[]): a null
// array equals only a null array.
func iTextOutputDeviceDashEquals(a []float32, b []float32) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	return slices.Equal(a, b)
}

func (d *ITextOutputDevice) SetStroke(s geom.Stroke) {
	d.originalStroke = s
	d.stroke = d.transformStroke(s)
}

func (d *ITextOutputDevice) transformStroke(stroke geom.Stroke) geom.Stroke {
	st, ok := stroke.(*geom.BasicStroke)
	if !ok {
		return stroke
	}
	scale := float32(math.Sqrt(math.Abs(d.transform.GetDeterminant())))
	dash := st.GetDashArray()
	if dash != nil {
		for k := range dash {
			dash[k] *= scale
		}
	}
	return geom.NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(st.GetLineWidth()*scale, st.GetEndCap(), st.GetLineJoin(), st.GetMiterLimit(), dash, st.GetDashPhase()*
		scale)
}

func (d *ITextOutputDevice) Clip(s geom.Shape) {
	if s != nil {
		s = d.transform.CreateTransformedShape(s)
		if d.clip == nil {
			d.clip = geom.NewArea(s)
		} else {
			d.clip.Intersect(geom.NewArea(s))
		}
		d.followPath(s, iTextOutputDeviceDrawTypeClip)
	} else {
		panic(ufo.NewXRRuntimeException("Shape is null, unexpected"))
	}
}

// GetClip returns nil when no clip is set or the transform has no inverse.
func (d *ITextOutputDevice) GetClip() geom.Shape {
	if d.clip == nil {
		return nil
	}
	inverse, err := d.transform.CreateInverse()
	if err != nil {
		return nil
	}
	return inverse.CreateTransformedShape(d.clip)
}

func (d *ITextOutputDevice) SetClip(s geom.Shape) {
	cb := d.currentPage
	cb.RestoreState()
	cb.SaveState()
	if s != nil {
		s = d.transform.CreateTransformedShape(s)
	}
	if s == nil {
		d.clip = nil
	} else {
		d.clip = geom.NewArea(s)
		d.followPath(s, iTextOutputDeviceDrawTypeClip)
	}
	d.fillColor = nil
	d.strokeColor = nil
	d.oldStroke = nil
}

func (d *ITextOutputDevice) GetStroke() geom.Stroke {
	return d.originalStroke
}

func (d *ITextOutputDevice) DrawImage(fsImage ufo.FSImage, x int, y int) {
	if pdfAsImage, ok := fsImage.(*PDFAsImage); ok {
		d.drawPDFAsImage(pdfAsImage, x, y)
	} else if iTextImage, ok := fsImage.(*ITextFSImage); ok {
		image := iTextImage.GetImage()

		if fsImage.GetHeight() <= 0 || fsImage.GetWidth() <= 0 {
			return
		}

		at := geom.AffineTransformGetTranslateInstance(float64(x), float64(y))
		at.Translate(0, float64(fsImage.GetHeight()))
		at.Scale(float64(fsImage.GetWidth()), float64(fsImage.GetHeight()))

		inverse := d.normalizeMatrix(d.transform)
		flipper := geom.AffineTransformGetScaleInstance(1, -1)
		inverse.Concatenate(at)
		inverse.Concatenate(flipper)

		mx := make([]float64, 6)
		inverse.GetMatrix(mx)

		err := d.currentPage.AddImage(image, float32(mx[0]), float32(mx[1]), float32(mx[2]), float32(mx[3]), float32(mx[4]), float32(mx[5]))
		if err != nil {
			panic(ufo.NewXRRuntimeExceptionWithCause(err.Error(), err))
		}
	} else {
		panic(ufo.NewXRRuntimeException(fmt.Sprintf("Unsupported image type: %T", fsImage)))
	}
}

func (d *ITextOutputDevice) DrawLinearGradient(gradient *ufo.FSLinearGradient, x int, y int, width int, height int) {
}

// drawPDFAsImage draws the first page of another PDF file. pdf/writer cannot
// import PDF pages: GetReader fails with a *writer.UnsupportedFeatureError,
// which this method reports as the cause of the XRRuntimeException it panics
// with, as Java does for a PDF file that cannot be loaded.
func (d *ITextOutputDevice) drawPDFAsImage(image *PDFAsImage, x int, y int) {
	uri := image.GetURI()

	reader, err := d.GetReader(uri)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause("Could not load "+uri.String()+": "+err.Error(), err))
	}

	page, err := d.GetWriter().GetImportedPage(reader, 1)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause("Could not load "+uri.String()+": "+err.Error(), err))
	}

	at := geom.AffineTransformGetTranslateInstance(float64(x), float64(y))
	at.Translate(0, float64(image.GetHeightAsFloat()))
	at.Scale(float64(image.GetWidthAsFloat()), float64(image.GetHeightAsFloat()))

	inverse := d.normalizeMatrix(d.transform)
	flipper := geom.AffineTransformGetScaleInstance(1, -1)
	inverse.Concatenate(at)
	inverse.Concatenate(flipper)

	mx := make([]float64, 6)
	inverse.GetMatrix(mx)

	mx[0] = float64(image.ScaleWidth())
	mx[3] = float64(image.ScaleHeight())

	d.currentPage.RestoreState()
	d.currentPage.AddTemplate(page, float32(mx[0]), float32(mx[1]), float32(mx[2]), float32(mx[3]), float32(mx[4]), float32(mx[5]))
	d.currentPage.SaveState()
}

// GetReader returns the reader of the PDF file at uri, reading the file the
// first time. With pdf/writer it always returns a
// *writer.UnsupportedFeatureError.
func (d *ITextOutputDevice) GetReader(uri *url.URL) (*writer.PdfReader, error) {
	result := d.readerCache[uri.String()]
	if result == nil {
		var err error
		result, err = writer.NewPdfReader(d.GetSharedContext().GetUserAgentCallback().GetBinaryResource(uri.String()))
		if err != nil {
			return nil, err
		}
		d.readerCache[uri.String()] = result
	}
	return result, nil
}

func (d *ITextOutputDevice) GetDotsPerPoint() float32 {
	return d.dotsPerPoint
}

func (d *ITextOutputDevice) Start(doc *dom.Document) {
	d.loadBookmarks(doc)
	d.loadMetadata(doc)
}

func (d *ITextOutputDevice) Finish(c *ufo.RenderingContext, root ufo.BoxI) {
	d.writeOutline(c, root)
	d.writeNamedDestinations(c)
	d.bookmarks = nil
}

func (d *ITextOutputDevice) writeOutline(c *ufo.RenderingContext, root ufo.BoxI) {
	if len(d.bookmarks) == 0 {
		d.bookmarks = HTMLOutlineGenerate(root.GetElement(), root)
	}
	if len(d.bookmarks) != 0 {
		d.writeBookmarks(c, root, d.writer.GetRootOutline(), d.bookmarks)
	}
}

func (d *ITextOutputDevice) writeBookmarks(c *ufo.RenderingContext, root ufo.BoxI, parent *writer.PdfOutline, bookmarks []*ITextOutputDeviceBookmark) {
	for _, bookmark := range bookmarks {
		d.writeBookmark(c, root, parent, bookmark)
	}
}

// iTextOutputDeviceSortedIds returns the keys of the id map in ascending
// order. Java iterates a HashMap, whose order is unspecified; the sorted order
// makes the output reproducible, and a PDF name tree requires its names
// sorted.
func iTextOutputDeviceSortedIds(idMap map[string]ufo.BoxI) []string {
	ids := make([]string, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (d *ITextOutputDevice) writeNamedDestinations(c *ufo.RenderingContext) {
	idMap := d.GetSharedContext().GetIdMap()
	if len(idMap) != 0 {
		destinations := writer.NewPdfArray()
		addToBody := func(object writer.PdfObject) *writer.PdfIndirectReference {
			indirect, err := d.writer.AddToBody(object)
			if err != nil {
				panic(ufo.NewXRRuntimeExceptionWithCause(err.Error(), err))
			}
			return indirect.GetIndirectReference()
		}
		for _, anchorName := range iTextOutputDeviceSortedIds(idMap) {
			targetBox := idMap[anchorName]

			if targetBox.GetStyle().IsIdent(ufo.CSSNameFsNamedDestination, ufo.IdentValueCreate) {
				destinations.Add(writer.NewPdfStringWithEncoding(anchorName, writer.PdfObjectTextUnicode))

				dest := d.createDestination(c, targetBox)
				if dest != nil {
					ref := addToBody(dest)
					destinations.Add(ref)
				}
			}
		}

		if !destinations.IsEmpty() {
			nameTree := writer.NewPdfDictionary()
			nameTree.Put(writer.PdfNameNames, destinations)
			nameTreeRef := addToBody(nameTree)

			names := writer.NewPdfDictionary()
			names.Put(writer.PdfNameDests, nameTreeRef)
			destinationsRef := addToBody(names)

			d.writer.GetExtraCatalog().Put(writer.PdfNameNames, destinationsRef)
		}
	}
}

func (d *ITextOutputDevice) getPageRefY(box ufo.BoxI) int {
	if iB, ok := box.(*ufo.InlineLayoutBox); ok {
		return iB.GetAbsY() + iB.GetBaseline()
	} else {
		return box.GetAbsY()
	}
}

func (d *ITextOutputDevice) writeBookmark(c *ufo.RenderingContext, root ufo.BoxI, parent *writer.PdfOutline, bookmark *ITextOutputDeviceBookmark) {
	href := bookmark.GetHRef()
	var target *writer.PdfDestination
	box := bookmark.GetBox()
	if href != "" && href[0] == '#' {
		box = d.sharedContext.GetBoxById(href[1:])
	}
	if box != nil {
		page := root.GetLayer().GetPage(c, d.getPageRefY(box))
		distanceFromTop := page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeTop)
		distanceFromTop += box.GetAbsY() - page.GetTop()
		target = writer.NewPdfDestinationWithLeftTopZoom(writer.PdfDestinationXyz, 0, d.normalizeY(float32(distanceFromTop)/d.dotsPerPoint), 0)
		target.AddPage(d.writer.GetPageReference(d.startPageNo + page.GetPageNo() + 1))
	}
	if target == nil {
		target = d.defaultDestination
	}
	outline := writer.NewPdfOutline(parent, target, bookmark.GetName())
	d.writeBookmarks(c, root, outline, bookmark.GetChildren())
}

func (d *ITextOutputDevice) loadBookmarks(doc *dom.Document) {
	head := DOMUtilGetChild(doc.GetDocumentElement(), "head")
	if head != nil {
		bookmarks := DOMUtilGetChild(head, "bookmarks")
		if bookmarks != nil {
			for _, e := range DOMUtilGetChildren(bookmarks, "bookmark") {
				d.loadBookmark(nil, e)
			}
		}
	}
}

func (d *ITextOutputDevice) loadBookmark(parent *ITextOutputDeviceBookmark, bookmark *dom.Element) {
	us := NewITextOutputDeviceBookmark(bookmark.GetAttribute("name"), bookmark.GetAttribute("href"))
	if parent == nil {
		d.bookmarks = append(d.bookmarks, us)
	} else {
		parent.AddChild(us)
	}
	for _, e := range DOMUtilGetChildren(bookmark, "bookmark") {
		d.loadBookmark(us, e)
	}
}

type ITextOutputDeviceBookmark struct {
	name string
	href string
	box  ufo.BoxI

	children []*ITextOutputDeviceBookmark
}

func NewITextOutputDeviceBookmark(name string, href string) *ITextOutputDeviceBookmark {
	return &ITextOutputDeviceBookmark{name: name, href: href}
}

// GetBox returns nil for a bookmark that is mapped to no box.
func (b *ITextOutputDeviceBookmark) GetBox() ufo.BoxI {
	return b.box
}

func (b *ITextOutputDeviceBookmark) SetBox(box ufo.BoxI) {
	b.box = box
}

func (b *ITextOutputDeviceBookmark) GetHRef() string {
	return b.href
}

func (b *ITextOutputDeviceBookmark) GetName() string {
	return b.name
}

func (b *ITextOutputDeviceBookmark) AddChild(child *ITextOutputDeviceBookmark) {
	b.children = append(b.children, child)
}

func (b *ITextOutputDeviceBookmark) GetChildren() []*ITextOutputDeviceBookmark {
	return b.children
}

// Metadata methods

// Methods to load and search a document's metadata

// AddMetadata appends a name/content metadata pair to this output device.
// Java ignores a null name or content value; a Go string is never null.
func (d *ITextOutputDevice) AddMetadata(name string, value string) {
	m := newITextOutputDeviceMetadata(name, value)
	d.metadata = append(d.metadata, m)
}

// GetMetadataByName searches the metadata name/content pairs of the current
// document and returns the content value from the first pair with a matching
// name. The search is case-insensitive. The result is nil when no pair has
// the name.
func (d *ITextOutputDevice) GetMetadataByName(name string) *string {
	for _, m := range d.metadata {
		if m != nil && strings.EqualFold(m.getName(), name) {
			content := m.getContent()
			return &content
		}
	}
	return nil
}

// GetMetadataListByName searches the metadata name/content pairs of the
// current document and returns any content values with a matching name. The
// search is case-insensitive. The result is empty when no pair has the name.
func (d *ITextOutputDevice) GetMetadataListByName(name string) []string {
	result := []string{}
	for _, m := range d.metadata {
		if m != nil && strings.EqualFold(m.getName(), name) {
			result = append(result, m.getContent())
		}
	}
	return result
}

// loadMetadata locates and stores all metadata values in the document head
// that contain name/content pairs. If there is no pair with a name of
// "title", any content in the title element is saved as a "title" metadata
// item. doc is the Document level node of the parsed xhtml file.
func (d *ITextOutputDevice) loadMetadata(doc *dom.Document) {
	head := DOMUtilGetChild(doc.GetDocumentElement(), "head")
	if head != nil {
		for _, e := range DOMUtilGetChildren(head, "meta") {
			name := e.GetAttribute("name")
			content := e.GetAttribute("content")
			m := newITextOutputDeviceMetadata(name, content)
			d.metadata = append(d.metadata, m)
		}
		// If there is no title metadata attribute, use the document title.
		if d.GetMetadataByName("title") == nil {
			t := DOMUtilGetChild(head, "title")
			if t != nil {
				title := htmlOutlineJavaTrim(DOMUtilGetText(t))
				m := newITextOutputDeviceMetadata("title", title)
				d.metadata = append(d.metadata, m)
			}
		}
	}
}

// SetMetadata replaces all copies of the named metadata with a single value.
// A nil value results in the removal of all copies of the named metadata.
// Use AddMetadata to append additional values with the same name.
func (d *ITextOutputDevice) SetMetadata(name string, value *string) {
	remove := value == nil // removing all instances of name?
	free := -1             // first open slot in array
	for i, m := range d.metadata {
		if m != nil {
			if strings.EqualFold(m.getName(), name) {
				if !remove {
					remove = true // remove all other instances
					m.setContent(*value)
				} else {
					d.metadata[i] = nil
				}
			}
		} else if free == -1 {
			free = i
		}
	}
	if !remove { // not found?
		m := newITextOutputDeviceMetadata(name, *value)
		if free == -1 { // no open slots?
			d.metadata = append(d.metadata, m)
		} else {
			d.metadata[free] = m
		}
	}
}

// iTextOutputDeviceMetadata stores metadata element name/content pairs from
// the head section of an xhtml document.
type iTextOutputDeviceMetadata struct {
	name    string
	content string
}

func newITextOutputDeviceMetadata(name string, content string) *iTextOutputDeviceMetadata {
	return &iTextOutputDeviceMetadata{name: name, content: content}
}

func (m *iTextOutputDeviceMetadata) getContent() string {
	return m.content
}

func (m *iTextOutputDeviceMetadata) setContent(content string) {
	m.content = content
}

func (m *iTextOutputDeviceMetadata) getName() string {
	return m.name
}

// Metadata end

func (d *ITextOutputDevice) GetSharedContext() *ufo.SharedContext {
	return d.sharedContext
}

func (d *ITextOutputDevice) SetSharedContext(sharedContext *ufo.SharedContext) {
	d.sharedContext = sharedContext
	sharedContext.GetCss().SetSupportCMYKColors(true)
}

func (d *ITextOutputDevice) SetRoot(root ufo.BoxI) {
	d.root = root
}

func (d *ITextOutputDevice) GetStartPageNo() int {
	return d.startPageNo
}

func (d *ITextOutputDevice) SetStartPageNo(startPageNo int) {
	d.startPageNo = startPageNo
}

func (d *ITextOutputDevice) DrawSelection(c *ufo.RenderingContext, inlineText *ufo.InlineText) {
	panic(ufo.NewXRRuntimeException("Unsupported operation: drawSelection"))
}

func (d *ITextOutputDevice) IsSupportsSelection() bool {
	return false
}

func (d *ITextOutputDevice) IsSupportsCMYKColors() bool {
	return true
}

// FindPagePositionsByID returns the positions of the boxes whose id the
// pattern finds a match in, ordered by page number.
func (d *ITextOutputDevice) FindPagePositionsByID(c ufo.CssContext, pattern *regexp.Regexp) []*PagePosition {
	idMap := d.sharedContext.GetIdMap()
	if idMap == nil {
		return nil
	}

	result := []*PagePosition{}
	for _, id := range iTextOutputDeviceSortedIds(idMap) {
		if pattern.MatchString(id) {
			box := idMap[id]
			pos := d.calcPDFPagePosition(c, id, box)
			if pos != nil {
				result = append(result, pos)
			}
		}
	}

	sort.SliceStable(result, func(i, j int) bool { return result[i].GetPageNo() < result[j].GetPageNo() })
	return result
}

// calcPDFPagePosition returns nil when the box is on no page.
func (d *ITextOutputDevice) calcPDFPagePosition(c ufo.CssContext, id string, box ufo.BoxI) *PagePosition {
	page := d.root.GetLayer().GetLastPageWithCBox(c, box)
	if page == nil {
		return nil
	}

	x := float32(box.GetAbsX() + page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeLeft))
	y := float32(page.GetBottom() - (box.GetAbsY() + box.GetHeight()) + page.GetMarginBorderPadding(c, ufo.CalculatedStyleEdgeBottom))

	return NewPagePosition(id, page.GetPageNo(),
		x/d.dotsPerPoint, float32(box.GetEffectiveWidth())/d.dotsPerPoint,
		y/d.dotsPerPoint, float32(box.GetHeight())/d.dotsPerPoint,
	)
}
