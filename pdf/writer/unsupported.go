package writer

// The methods in this file exist so that code ported from the Flying Saucer
// PDF module can call them and report the error. None of the features is
// implemented: encryption, PDF/A and PDF/X conformance (XMP metadata, output
// intents), tagged PDF (structure tree), AcroForm fields, imported PDF pages
// and SVG.

// SetEncryption is PdfWriter.setEncryption(userPassword, ownerPassword,
// permissions, encryptionType).
func (w *PdfWriter) SetEncryption(userPassword, ownerPassword []byte, permissions, encryptionType int) error {
	return unsupported("encryption")
}

// SetPDFXConformance is PdfWriter.setPDFXConformance. PdfWriterPdfxnone is
// accepted; any other level is not supported.
func (w *PdfWriter) SetPDFXConformance(pdfx int) error {
	if pdfx == PdfWriterPdfxnone {
		return nil
	}
	return unsupported("PDF/X and PDF/A conformance")
}

// GetPDFXConformance returns PdfWriterPdfxnone.
func (w *PdfWriter) GetPDFXConformance() int { return PdfWriterPdfxnone }

// SetOutputIntents is PdfWriter.setOutputIntents(outputConditionIdentifier,
// outputCondition, registryName, info, colorProfile).
func (w *PdfWriter) SetOutputIntents(outputConditionIdentifier, outputCondition, registryName, info string, iccProfile []byte) error {
	return unsupported("output intents")
}

// CreateXmpMetadata is PdfWriter.createXmpMetadata.
func (w *PdfWriter) CreateXmpMetadata() error { return unsupported("XMP metadata") }

// SetXmpMetadata is PdfWriter.setXmpMetadata.
func (w *PdfWriter) SetXmpMetadata(xmpMetadata []byte) error { return unsupported("XMP metadata") }

// SetPageXmpMetadata is PdfWriter.setPageXmpMetadata.
func (w *PdfWriter) SetPageXmpMetadata(xmpMetadata []byte) error {
	return unsupported("page XMP metadata")
}

// SetTagged is PdfWriter.setTagged.
func (w *PdfWriter) SetTagged() error { return unsupported("tagged PDF") }

// IsTagged reports false: SetTagged never succeeds.
func (w *PdfWriter) IsTagged() bool { return false }

// PdfStructureTreeRoot and PdfStructureElement are the types of the tagged
// PDF structure tree. No value of either can be obtained.
type PdfStructureTreeRoot struct{}

// PdfStructureElement is an element of the structure tree.
type PdfStructureElement struct {
	PdfDictionary
}

// GetStructureTreeRoot is PdfWriter.getStructureTreeRoot.
func (w *PdfWriter) GetStructureTreeRoot() (*PdfStructureTreeRoot, error) {
	return nil, unsupported("tagged PDF")
}

// NewPdfStructureElement is PdfStructureElement(parent, structureType).
func NewPdfStructureElement(parent *PdfStructureElement, structureType PdfName) (*PdfStructureElement, error) {
	return nil, unsupported("tagged PDF")
}

// NewPdfStructureElementWithRoot is PdfStructureElement(root, structureType).
func NewPdfStructureElementWithRoot(parent *PdfStructureTreeRoot, structureType PdfName) (*PdfStructureElement, error) {
	return nil, unsupported("tagged PDF")
}

// BeginMarkedContentSequenceStruct is
// PdfContentByte.beginMarkedContentSequence(PdfStructureElement).
func (cb *PdfContentByte) BeginMarkedContentSequenceStruct(element *PdfStructureElement) error {
	return unsupported("tagged PDF")
}

// PdfFormField stands for com.lowagie.text.pdf.PdfFormField; no value of it
// can be obtained.
type PdfFormField struct {
	PdfAnnotation
}

// PdfAcroForm stands for com.lowagie.text.pdf.PdfAcroForm.
type PdfAcroForm struct{}

// GetAcroForm is PdfWriter.getAcroForm.
func (w *PdfWriter) GetAcroForm() (*PdfAcroForm, error) { return nil, unsupported("AcroForm fields") }

// AddFormField stands for PdfWriter.addAnnotation(PdfFormField).
func (w *PdfWriter) AddFormField(field *PdfFormField) error { return unsupported("AcroForm fields") }

// CreateAppearance is PdfContentByte.createAppearance(width, height), which
// only form fields use.
func (cb *PdfContentByte) CreateAppearance(width, height float32) (*PdfTemplate, error) {
	return nil, unsupported("AcroForm fields")
}

// PdfReader stands for com.lowagie.text.pdf.PdfReader, which the Java code
// uses to draw a page of another PDF file as an image.
type PdfReader struct{}

// NewPdfReader is PdfReader(byte[]).
func NewPdfReader(data []byte) (*PdfReader, error) { return nil, unsupported("importing PDF pages") }

// GetImportedPage is PdfWriter.getImportedPage(reader, pageNumber).
func (w *PdfWriter) GetImportedPage(reader *PdfReader, pageNumber int) (*PdfTemplate, error) {
	return nil, unsupported("importing PDF pages")
}

// ImageGetInstanceFromSvg stands for the SVG rasterization of SvgImage.java.
func ImageGetInstanceFromSvg(svg []byte) (*Image, error) { return nil, unsupported("SVG") }
