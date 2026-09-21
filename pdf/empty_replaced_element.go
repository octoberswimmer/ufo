// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/EmptyReplacedElement.java

package pdf

import (
	"unicode/utf8"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/geom"
)

const emptyReplacedElementFieldType = "Hidden"

// EmptyReplacedElement is the replaced element of a hidden input.
//
// User: beck
// Date: 11/4/11
type EmptyReplacedElement struct {
	AbstractFormField

	width  int
	height int

	location *geom.Point
}

func NewEmptyReplacedElement(width int, height int) *EmptyReplacedElement {
	e := &EmptyReplacedElement{
		width:    width,
		height:   height,
		location: geom.NewPoint(0, 0),
	}
	e.fieldType = e.getFieldType()
	return e
}

// Paint adds a hidden field to the AcroForm in Java. pdf/writer does not
// write AcroForm fields: GetAcroForm returns a *writer.UnsupportedFeatureError,
// which is logged as a warning once per document, and the field is left out.
func (e *EmptyReplacedElement) Paint(c *ufo.RenderingContext, outputDevice *ITextOutputDevice, box ufo.BlockBoxI) {
	w := outputDevice.GetWriter()

	acroForm, err := w.GetAcroForm()
	if err != nil {
		outputDevice.warnUnsupportedOnce(err, "Form fields are left out of the PDF")
		return
	}
	elem := box.GetElement()
	name := e.getFieldName(outputDevice, elem)
	value := e.getValue(elem)
	/*ISO-32000-1 defines the limit for a name in a PDF file to be at maximum 127 bytes.
	 *Source(http://www.adobe.com/content/dam/Adobe/en/devnet/acrobat/pdfs/PDF32000_2008.pdf)
	 *  see Annex C § 2 Architectural limits "Table C.1" pages 649 and 650.
	 *iText stores the hidden field value as a PDFName
	 */
	if utf8.RuneCountInString(value) > 127 {
		value = string([]rune(value)[:127])
	}
	// writer.PdfAcroForm has no addHiddenField, and no *writer.PdfAcroForm can
	// be obtained, so this point is not reached.
	_, _, _ = acroForm, name, value
	panic(ufo.NewXRRuntimeException("not ported: PdfAcroForm.addHiddenField"))
}

func (e *EmptyReplacedElement) GetIntrinsicWidth() int {
	return e.width
}

func (e *EmptyReplacedElement) GetIntrinsicHeight() int {
	return e.height
}

func (e *EmptyReplacedElement) GetLocation() *geom.Point {
	return e.location
}

func (e *EmptyReplacedElement) SetLocation(x int, y int) {
	e.location = geom.NewPoint(0, 0)
}

func (e *EmptyReplacedElement) getFieldType() string {
	return emptyReplacedElementFieldType
}

func (e *EmptyReplacedElement) HasBaseline() bool {
	return false
}

func (e *EmptyReplacedElement) GetBaseline() int {
	return 0
}
