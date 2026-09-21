// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/AbstractFormField.java
// (the members EmptyReplacedElement uses; see the comment below).

package pdf

import (
	"strconv"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
)

// The AcroForm field classes of flying-saucer-pdf, which are otherwise not
// ported because pdf/writer does not write AcroForm fields
// (writer.PdfWriter.GetAcroForm returns a *writer.UnsupportedFeatureError).
// They hold only the members that the ported classes use.

// AbstractFormField stands for
// flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/AbstractFormField.java.
// It carries the field name logic and the ReplacedElement methods that
// EmptyReplacedElement inherits. The abstract getFieldType() is the fieldType
// value set by the embedding class.
type AbstractFormField struct {
	fieldType string

	// fieldName is nil until getFieldName computes it.
	fieldName *string
}

const abstractFormFieldDefaultCheckedState = "Yes"

func (f *AbstractFormField) getFieldName(outputDevice *ITextOutputDevice, e *dom.Element) string {
	if f.fieldName == nil {
		result := e.GetAttribute("name")

		var fieldName string
		if ufo.UtilIsNullOrEmpty(result) {
			fieldName = f.fieldType + strconv.Itoa(outputDevice.GetNextFormFieldIndex())
		} else {
			fieldName = result
		}
		f.fieldName = &fieldName
	}

	return *f.fieldName
}

func (f *AbstractFormField) getValue(e *dom.Element) string {
	result := e.GetAttribute("value")

	if ufo.UtilIsNullOrEmpty(result) {
		return abstractFormFieldDefaultCheckedState
	} else {
		return result
	}
}

func (f *AbstractFormField) Detach(c *ufo.LayoutContext) {
}

func (f *AbstractFormField) IsRequiresInteractivePaint() bool {
	// N/A
	return false
}

// RadioButtonFormField stands for
// flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/RadioButtonFormField.java.
// It has no constructor: ITextReplacedElementFactory creates no radio button
// fields, so its radio button maps stay empty.
type RadioButtonFormField struct {
	AbstractFormField
	box ufo.BoxI
}

func (f *RadioButtonFormField) GetBox() ufo.BoxI {
	return f.box
}
