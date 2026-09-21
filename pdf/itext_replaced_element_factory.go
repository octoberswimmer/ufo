// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextReplacedElementFactory.java

package pdf

import (
	"slices"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// ITextReplacedElementFactory creates the replaced elements of the PDF
// renderer.
//
// Differences from Java, all because pdf/writer does not support the feature
// (see "Not supported" in its package comment). Each logs a warning once per
// document and then does what Java does for an element it does not replace
// (CreateReplacedElement returns nil, so the element is laid out as ordinary
// content):
//   - an inline <svg> element is not rasterized (SvgImage, Batik), so
//     serializeToXml and ensureSvgNamespace have no counterpart here;
//   - <input> elements other than type="hidden" create no AcroForm field
//     (CheckboxFormField, RadioButtonFormField, TextFormField).
//
// setFormSubmissionListener is not ported: the core's ReplacedElementFactory
// does not declare it, and the Java method does nothing.
type ITextReplacedElementFactory struct {
	outputDevice *ITextOutputDevice

	radioButtonsByElem map[*dom.Element]*RadioButtonFormField
	radioButtonsByName map[string][]*RadioButtonFormField
}

func NewITextReplacedElementFactory(outputDevice *ITextOutputDevice) *ITextReplacedElementFactory {
	return &ITextReplacedElementFactory{
		outputDevice:       outputDevice,
		radioButtonsByElem: map[*dom.Element]*RadioButtonFormField{},
		radioButtonsByName: map[string][]*RadioButtonFormField{},
	}
}

// CreateReplacedElement returns nil when the element is not replaced.
func (f *ITextReplacedElementFactory) CreateReplacedElement(c *ufo.LayoutContext, box ufo.BlockBoxI,
	uac ufo.UserAgentCallback, cssWidth int, cssHeight int) ufo.ReplacedElement {
	e := box.GetElement()
	if e == nil {
		return nil
	}

	nodeName := e.GetNodeName()
	switch nodeName {
	case "img":
		srcAttr := e.GetAttribute("src")
		if srcAttr != "" {
			fsImage := uac.GetImageResource(srcAttr).GetImage()
			if fsImage != nil {
				if cssWidth != -1 || cssHeight != -1 {
					fsImage = fsImage.Scale(cssWidth, cssHeight)
				}
				return NewITextImageElement(fsImage)
			}
		}

	case "svg":
		f.outputDevice.warnUnsupportedOnce(&writer.UnsupportedFeatureError{Feature: "SVG"},
			"An inline <svg> element is not drawn")
	case "input":
		inputType := e.GetAttribute("type")
		switch inputType {
		case "hidden":
			return NewEmptyReplacedElement(1, 1)
		default:
			// "checkbox", "radio" and every other type: no form field is
			// created.
			f.outputDevice.warnUnsupportedOnce(&writer.UnsupportedFeatureError{Feature: "AcroForm fields"},
				"Form fields are left out of the PDF")
		}
	/*
	   } else if (nodeName.equals("select")) {//TODO Support select
	   return new SelectFormField(c, box, cssWidth, cssHeight);
	   } else if (isTextarea(e)) {//TODO Review if this is needed the textarea item prints fine currently
	   return new TextAreaFormField(c, box, cssWidth, cssHeight);
	*/
	case "bookmark":
		// HACK Add box as named anchor and return placeholder
		if e.HasAttribute("name") {
			name := e.GetAttribute("name")
			c.AddBoxId(name, box)
			return NewBookmarkElement(&name)
		}
		return NewBookmarkElement(nil)
	}

	return nil
}

// saveResult has no caller, because CreateReplacedElement creates no radio
// button fields.
func (f *ITextReplacedElementFactory) saveResult(e *dom.Element, result *RadioButtonFormField) {
	f.radioButtonsByElem[e] = result

	fieldName := result.getFieldName(f.outputDevice, e)
	f.radioButtonsByName[fieldName] = append(f.radioButtonsByName[fieldName], result)
}

func (f *ITextReplacedElementFactory) Reset() {
	clear(f.radioButtonsByElem)
	clear(f.radioButtonsByName)
	// A new document is about to be loaded: report its unsupported features
	// again.
	f.outputDevice.resetUnsupportedWarnings()
}

func (f *ITextReplacedElementFactory) Remove(e *dom.Element) {
	field, ok := f.radioButtonsByElem[e]
	delete(f.radioButtonsByElem, e)
	if ok && field != nil {
		fieldName := field.getFieldName(f.outputDevice, e)
		values, ok := f.radioButtonsByName[fieldName]
		if ok {
			if i := slices.Index(values, field); i >= 0 {
				values = slices.Delete(values, i, i+1)
				f.radioButtonsByName[fieldName] = values
			}
			if len(values) == 0 {
				delete(f.radioButtonsByName, fieldName)
			}
		}
	}
}

func (f *ITextReplacedElementFactory) RemoveString(fieldName string) {
	values, ok := f.radioButtonsByName[fieldName]
	if ok {
		for _, field := range values {
			delete(f.radioButtonsByElem, field.GetBox().GetElement())
		}
	}

	delete(f.radioButtonsByName, fieldName)
}

// GetRadioButtons returns nil when no radio button has the name.
func (f *ITextReplacedElementFactory) GetRadioButtons(name string) []*RadioButtonFormField {
	return f.radioButtonsByName[name]
}
