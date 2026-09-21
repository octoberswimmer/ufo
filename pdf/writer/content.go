package writer

import (
	"bytes"
	"fmt"
	"math"
)

// Text rendering modes and line styles, as PdfContentByte.TEXT_RENDER_MODE_*
// and LINE_CAP_* / LINE_JOIN_*.
const (
	PdfContentByteTextRenderModeFill           = 0
	PdfContentByteTextRenderModeStroke         = 1
	PdfContentByteTextRenderModeFillStroke     = 2
	PdfContentByteTextRenderModeInvisible      = 3
	PdfContentByteTextRenderModeFillClip       = 4
	PdfContentByteTextRenderModeStrokeClip     = 5
	PdfContentByteTextRenderModeFillStrokeClip = 6
	PdfContentByteTextRenderModeClip           = 7
	PdfContentByteLineCapButt                  = 0
	PdfContentByteLineCapRound                 = 1
	PdfContentByteLineCapProjectingSquare      = 2
	PdfContentByteLineJoinMiter                = 0
	PdfContentByteLineJoinRound                = 1
	PdfContentByteLineJoinBevel                = 2
)

// Color is a color SetColorFill and SetColorStroke accept: *RGBColor (the
// counterpart of java.awt.Color), *CMYKColor or *GrayColor.
type Color interface {
	// GetAlpha returns the opacity from 0 to 255.
	GetAlpha() int
}

// RGBColor is a DeviceRGB color with components from 0 to 255.
type RGBColor struct {
	Red, Green, Blue, Alpha int
}

// NewRGBColor returns an opaque RGB color.
func NewRGBColor(red, green, blue int) *RGBColor {
	return &RGBColor{Red: red, Green: green, Blue: blue, Alpha: 255}
}

// NewRGBColorWithAlpha returns an RGB color with the given opacity.
func NewRGBColorWithAlpha(red, green, blue, alpha int) *RGBColor {
	return &RGBColor{Red: red, Green: green, Blue: blue, Alpha: alpha}
}

func (c *RGBColor) GetAlpha() int { return c.Alpha }

// CMYKColor is a DeviceCMYK color with components from 0 to 1.
type CMYKColor struct {
	Cyan, Magenta, Yellow, Black float32
	Alpha                        int
}

// NewCMYKColor returns an opaque CMYK color.
func NewCMYKColor(cyan, magenta, yellow, black float32) *CMYKColor {
	return &CMYKColor{Cyan: cyan, Magenta: magenta, Yellow: yellow, Black: black, Alpha: 255}
}

func (c *CMYKColor) GetAlpha() int { return c.Alpha }

// GrayColor is a DeviceGray color from 0 (black) to 1 (white).
type GrayColor struct {
	Gray  float32
	Alpha int
}

// NewGrayColor returns an opaque gray.
func NewGrayColor(gray float32) *GrayColor { return &GrayColor{Gray: gray, Alpha: 255} }

func (c *GrayColor) GetAlpha() int { return c.Alpha }

// graphicState is the part of the graphics state the content byte tracks so
// that RestoreState brings it back.
type graphicState struct {
	font        *fontEntry
	size        float32
	fillAlpha   int
	strokeAlpha int
}

// PdfContentByte builds a content stream with the operators of
// com.lowagie.text.pdf.PdfContentByte. Coordinates are PDF user space:
// origin at the bottom left, y up, in points.
//
// Misuse that OpenPDF reports with a runtime exception (RestoreState without
// SaveState, showing text before SetFontAndSize) panics with a
// *DocumentException.
type PdfContentByte struct {
	writer     *PdfWriter
	content    bytes.Buffer
	resources  *resourceSet
	state      graphicState
	stateStack []graphicState
	inText     bool
	mcDepth    int
}

func newContentByte(w *PdfWriter) *PdfContentByte {
	return &PdfContentByte{writer: w, resources: newResourceSet(), state: graphicState{fillAlpha: 255, strokeAlpha: 255}}
}

// GetPdfWriter returns the writer the content belongs to.
func (cb *PdfContentByte) GetPdfWriter() *PdfWriter { return cb.writer }

// Size returns the number of bytes written so far.
func (cb *PdfContentByte) Size() int { return cb.content.Len() }

// ToPdf returns the bytes written so far.
func (cb *PdfContentByte) ToPdf() []byte { return append([]byte(nil), cb.content.Bytes()...) }

// Reset empties the content and forgets the tracked state.
func (cb *PdfContentByte) Reset() {
	cb.content.Reset()
	cb.resources = newResourceSet()
	cb.state = graphicState{fillAlpha: 255, strokeAlpha: 255}
	cb.stateStack = nil
	cb.inText = false
	cb.mcDepth = 0
}

func (cb *PdfContentByte) num(v float32) *PdfContentByte {
	cb.content.WriteString(formatNumber(float64(v)))
	return cb
}

func (cb *PdfContentByte) op(operator string, operands ...float32) {
	for _, v := range operands {
		cb.content.WriteString(formatNumber(float64(v)))
		cb.content.WriteByte(' ')
	}
	cb.content.WriteString(operator)
	cb.content.WriteByte('\n')
}

// SaveState writes q.
func (cb *PdfContentByte) SaveState() {
	cb.content.WriteString("q\n")
	cb.stateStack = append(cb.stateStack, cb.state)
}

// RestoreState writes Q and brings back the font, size and color opacities
// tracked at the matching SaveState.
func (cb *PdfContentByte) RestoreState() {
	cb.content.WriteString("Q\n")
	n := len(cb.stateStack)
	if n == 0 {
		panic(newDocumentException("Unbalanced save/restore state operators."))
	}
	cb.state = cb.stateStack[n-1]
	cb.stateStack = cb.stateStack[:n-1]
}

// ConcatCTM writes a b c d e f cm.
func (cb *PdfContentByte) ConcatCTM(a, b, c, d, e, f float32) { cb.op("cm", a, b, c, d, e, f) }

// Transform is ConcatCTM for a matrix given as [a b c d e f].
func (cb *PdfContentByte) Transform(m [6]float64) {
	cb.ConcatCTM(float32(m[0]), float32(m[1]), float32(m[2]), float32(m[3]), float32(m[4]), float32(m[5]))
}

// MoveTo writes x y m.
func (cb *PdfContentByte) MoveTo(x, y float32) { cb.op("m", x, y) }

// LineTo writes x y l.
func (cb *PdfContentByte) LineTo(x, y float32) { cb.op("l", x, y) }

// CurveTo writes x1 y1 x2 y2 x3 y3 c.
func (cb *PdfContentByte) CurveTo(x1, y1, x2, y2, x3, y3 float32) {
	cb.op("c", x1, y1, x2, y2, x3, y3)
}

// CurveToWithX2Y2X3Y3 is PdfContentByte.curveTo(x2, y2, x3, y3): it writes
// x2 y2 x3 y3 v, a curve whose first control point is the current point.
func (cb *PdfContentByte) CurveToWithX2Y2X3Y3(x2, y2, x3, y3 float32) { cb.op("v", x2, y2, x3, y3) }

// CurveFromTo writes x1 y1 x3 y3 y, a curve whose second control point is
// the end point.
func (cb *PdfContentByte) CurveFromTo(x1, y1, x3, y3 float32) { cb.op("y", x1, y1, x3, y3) }

// Rectangle writes x y w h re.
func (cb *PdfContentByte) Rectangle(x, y, w, h float32) { cb.op("re", x, y, w, h) }

// ClosePath writes h.
func (cb *PdfContentByte) ClosePath() { cb.op("h") }

// NewPath writes n.
func (cb *PdfContentByte) NewPath() { cb.op("n") }

// Stroke writes S.
func (cb *PdfContentByte) Stroke() { cb.op("S") }

// ClosePathStroke writes s.
func (cb *PdfContentByte) ClosePathStroke() { cb.op("s") }

// Fill writes f.
func (cb *PdfContentByte) Fill() { cb.op("f") }

// EoFill writes f*.
func (cb *PdfContentByte) EoFill() { cb.op("f*") }

// FillStroke writes B.
func (cb *PdfContentByte) FillStroke() { cb.op("B") }

// EoFillStroke writes B*.
func (cb *PdfContentByte) EoFillStroke() { cb.op("B*") }

// ClosePathFillStroke writes b.
func (cb *PdfContentByte) ClosePathFillStroke() { cb.op("b") }

// Clip writes W.
func (cb *PdfContentByte) Clip() { cb.op("W") }

// EoClip writes W*.
func (cb *PdfContentByte) EoClip() { cb.op("W*") }

// SetLineWidth writes w.
func (cb *PdfContentByte) SetLineWidth(w float32) { cb.op("w", w) }

// SetLineCap writes J.
func (cb *PdfContentByte) SetLineCap(style int) {
	if style >= 0 && style <= 2 {
		cb.op("J", float32(style))
	}
}

// SetLineJoin writes j.
func (cb *PdfContentByte) SetLineJoin(style int) {
	if style >= 0 && style <= 2 {
		cb.op("j", float32(style))
	}
}

// SetMiterLimit writes M. A limit of 1 or less is not written, as in
// OpenPDF.
func (cb *PdfContentByte) SetMiterLimit(miterLimit float32) {
	if miterLimit > 1 {
		cb.op("M", miterLimit)
	}
}

// SetFlatness writes i.
func (cb *PdfContentByte) SetFlatness(flatness float32) {
	if flatness >= 0 && flatness <= 100 {
		cb.op("i", flatness)
	}
}

// SetLineDash writes [array] phase d. An empty array gives a solid line.
func (cb *PdfContentByte) SetLineDash(array []float32, phase float32) {
	cb.content.WriteByte('[')
	for i, v := range array {
		if i > 0 {
			cb.content.WriteByte(' ')
		}
		cb.num(v)
	}
	cb.content.WriteString("] ")
	cb.num(phase)
	cb.content.WriteString(" d\n")
}

// SetLineDashPhase is setLineDash(phase): a solid line.
func (cb *PdfContentByte) SetLineDashPhase(phase float32) { cb.SetLineDash(nil, phase) }

// SetLineDashOn is setLineDash(unitsOn, phase).
func (cb *PdfContentByte) SetLineDashOn(unitsOn, phase float32) {
	cb.SetLineDash([]float32{unitsOn}, phase)
}

// SetLineDashOnOff is setLineDash(unitsOn, unitsOff, phase).
func (cb *PdfContentByte) SetLineDashOnOff(unitsOn, unitsOff, phase float32) {
	cb.SetLineDash([]float32{unitsOn, unitsOff}, phase)
}

// SetLiteral appends s to the content as given.
func (cb *PdfContentByte) SetLiteral(s string) { cb.content.WriteString(s) }

// SetLiteralChar appends one character.
func (cb *PdfContentByte) SetLiteralChar(c rune) { cb.content.WriteRune(c) }

// SetLiteralFloat appends a number in the writer's number format.
func (cb *PdfContentByte) SetLiteralFloat(n float32) { cb.num(n) }

func clamp01(v float32) float32 {
	if v < 0 || math.IsNaN(float64(v)) {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (cb *PdfContentByte) setFillAlpha(alpha int) {
	if cb.state.fillAlpha != alpha {
		gs := NewPdfGState()
		gs.SetFillOpacity(float32(alpha) / 255)
		cb.SetGState(gs)
	}
}

func (cb *PdfContentByte) setStrokeAlpha(alpha int) {
	if cb.state.strokeAlpha != alpha {
		gs := NewPdfGState()
		gs.SetStrokeOpacity(float32(alpha) / 255)
		cb.SetGState(gs)
	}
}

// SetColorFill sets the fill color: rg for *RGBColor, k for *CMYKColor, g
// for *GrayColor. When the opacity of the color differs from the fill
// opacity in effect, an ExtGState with the new /ca is set first, as OpenPDF
// does. Unlike OpenPDF, the opacity in effect is part of the state
// SaveState and RestoreState track, so it follows q and Q.
func (cb *PdfContentByte) SetColorFill(color Color) {
	cb.setFillAlpha(color.GetAlpha())
	switch c := color.(type) {
	case *RGBColor:
		cb.op("rg", clamp01(float32(c.Red&0xff)/255), clamp01(float32(c.Green&0xff)/255), clamp01(float32(c.Blue&0xff)/255))
	case *CMYKColor:
		cb.op("k", clamp01(c.Cyan), clamp01(c.Magenta), clamp01(c.Yellow), clamp01(c.Black))
	case *GrayColor:
		cb.op("g", clamp01(c.Gray))
	default:
		panic(newDocumentException("unsupported color type %T", color))
	}
}

// SetColorStroke sets the stroke color: RG, K or G, with /CA handled as
// SetColorFill handles /ca.
func (cb *PdfContentByte) SetColorStroke(color Color) {
	cb.setStrokeAlpha(color.GetAlpha())
	switch c := color.(type) {
	case *RGBColor:
		cb.op("RG", clamp01(float32(c.Red&0xff)/255), clamp01(float32(c.Green&0xff)/255), clamp01(float32(c.Blue&0xff)/255))
	case *CMYKColor:
		cb.op("K", clamp01(c.Cyan), clamp01(c.Magenta), clamp01(c.Yellow), clamp01(c.Black))
	case *GrayColor:
		cb.op("G", clamp01(c.Gray))
	default:
		panic(newDocumentException("unsupported color type %T", color))
	}
}

// SetRGBColorFill sets an opaque RGB fill color from components 0 to 255.
func (cb *PdfContentByte) SetRGBColorFill(red, green, blue int) {
	cb.SetColorFill(NewRGBColor(red, green, blue))
}

// SetRGBColorStroke sets an opaque RGB stroke color from components 0 to 255.
func (cb *PdfContentByte) SetRGBColorStroke(red, green, blue int) {
	cb.SetColorStroke(NewRGBColor(red, green, blue))
}

// SetRGBColorFillF sets an RGB fill color from components 0 to 1.
func (cb *PdfContentByte) SetRGBColorFillF(red, green, blue float32) {
	cb.setFillAlpha(255)
	cb.op("rg", clamp01(red), clamp01(green), clamp01(blue))
}

// SetRGBColorStrokeF sets an RGB stroke color from components 0 to 1.
func (cb *PdfContentByte) SetRGBColorStrokeF(red, green, blue float32) {
	cb.setStrokeAlpha(255)
	cb.op("RG", clamp01(red), clamp01(green), clamp01(blue))
}

// SetCMYKColorFillF sets a CMYK fill color from components 0 to 1.
func (cb *PdfContentByte) SetCMYKColorFillF(cyan, magenta, yellow, black float32) {
	cb.SetColorFill(NewCMYKColor(cyan, magenta, yellow, black))
}

// SetCMYKColorStrokeF sets a CMYK stroke color from components 0 to 1.
func (cb *PdfContentByte) SetCMYKColorStrokeF(cyan, magenta, yellow, black float32) {
	cb.SetColorStroke(NewCMYKColor(cyan, magenta, yellow, black))
}

// SetGrayFill sets a gray fill color.
func (cb *PdfContentByte) SetGrayFill(gray float32) { cb.SetColorFill(NewGrayColor(gray)) }

// SetGrayStroke sets a gray stroke color.
func (cb *PdfContentByte) SetGrayStroke(gray float32) { cb.SetColorStroke(NewGrayColor(gray)) }

// SetGState writes /GSn gs for an extended graphics state.
func (cb *PdfContentByte) SetGState(gstate *PdfGState) {
	e := cb.writer.addGState(gstate)
	cb.resources.gstates[e.name] = e.ref
	e.name.WritePdf(&cb.content)
	cb.content.WriteString(" gs\n")
	if gstate.fillOpacity != nil {
		cb.state.fillAlpha = int(math.Round(float64(*gstate.fillOpacity) * 255))
	}
	if gstate.strokeOpacity != nil {
		cb.state.strokeAlpha = int(math.Round(float64(*gstate.strokeOpacity) * 255))
	}
}

// BeginText writes BT.
func (cb *PdfContentByte) BeginText() {
	if cb.inText {
		panic(newDocumentException("Unbalanced begin/end text operators."))
	}
	cb.inText = true
	cb.content.WriteString("BT\n")
}

// EndText writes ET.
func (cb *PdfContentByte) EndText() {
	if !cb.inText {
		panic(newDocumentException("Unbalanced begin/end text operators."))
	}
	cb.inText = false
	cb.content.WriteString("ET\n")
}

// SetFontAndSize writes /Fn size Tf.
func (cb *PdfContentByte) SetFontAndSize(bf *BaseFont, size float32) {
	if bf == nil {
		panic(newDocumentException("Font cannot be null."))
	}
	if size < 0.0001 && size > -0.0001 {
		panic(newDocumentException("Font size too small: %v", size))
	}
	e := cb.writer.addFont(bf)
	cb.resources.fonts[e.name] = e.ref
	cb.state.font = e
	cb.state.size = size
	e.name.WritePdf(&cb.content)
	cb.content.WriteByte(' ')
	cb.num(size)
	cb.content.WriteString(" Tf\n")
}

// SetTextMatrix writes a b c d x y Tm.
func (cb *PdfContentByte) SetTextMatrix(a, b, c, d, x, y float32) { cb.op("Tm", a, b, c, d, x, y) }

// SetTextMatrixXY is setTextMatrix(x, y): 1 0 0 1 x y Tm.
func (cb *PdfContentByte) SetTextMatrixXY(x, y float32) { cb.SetTextMatrix(1, 0, 0, 1, x, y) }

// MoveText writes x y Td.
func (cb *PdfContentByte) MoveText(x, y float32) { cb.op("Td", x, y) }

// SetLeading writes TL.
func (cb *PdfContentByte) SetLeading(leading float32) { cb.op("TL", leading) }

// NewlineText writes T*.
func (cb *PdfContentByte) NewlineText() { cb.op("T*") }

// SetTextRenderingMode writes Tr.
func (cb *PdfContentByte) SetTextRenderingMode(rendering int) { cb.op("Tr", float32(rendering)) }

// SetCharacterSpacing writes Tc.
func (cb *PdfContentByte) SetCharacterSpacing(charSpace float32) { cb.op("Tc", charSpace) }

// SetWordSpacing writes Tw.
func (cb *PdfContentByte) SetWordSpacing(wordSpace float32) { cb.op("Tw", wordSpace) }

// SetHorizontalScaling writes Tz.
func (cb *PdfContentByte) SetHorizontalScaling(scale float32) { cb.op("Tz", scale) }

// SetTextRise writes Ts.
func (cb *PdfContentByte) SetTextRise(rise float32) { cb.op("Ts", rise) }

func (cb *PdfContentByte) currentFont() *fontEntry {
	if cb.state.font == nil {
		panic(newDocumentException("Font and size must be set before writing any text"))
	}
	return cb.state.font
}

func (cb *PdfContentByte) showTextBytes(text string) {
	e := cb.currentFont()
	data := e.font.convertToBytes(text, e.used)
	// Glyph ids are written in the <...> form, WinAnsi codes as a literal.
	writeStringBytes(&cb.content, data, e.font.tt != nil)
}

// ShowText writes (text) Tj in the encoding of the current font. Characters
// the font cannot show are left out.
func (cb *PdfContentByte) ShowText(text string) {
	cb.showTextBytes(text)
	cb.content.WriteString("Tj\n")
}

// ShowTextKerned writes text with a TJ array that applies the kerning of the
// current font between adjacent characters.
func (cb *PdfContentByte) ShowTextKerned(text string) {
	bf := cb.currentFont().font
	if !bf.HasKernPairs() {
		cb.ShowText(text)
		return
	}
	array := NewPdfTextArray()
	runes := []rune(text)
	start := 0
	for k := 0; k+1 < len(runes); k++ {
		if kern := bf.GetKerning(runes[k], runes[k+1]); kern != 0 {
			array.AddString(string(runes[start : k+1]))
			array.AddNumber(float32(-kern))
			start = k + 1
		}
	}
	array.AddString(string(runes[start:]))
	cb.ShowTextArray(array)
}

// ShowTextArray is PdfContentByte.showText(PdfTextArray): it writes
// [(string) number ...] TJ. A number moves the next string left by that many
// thousandths of the font size. As in OpenPDF, a space separates two
// adjacent numbers only.
func (cb *PdfContentByte) ShowTextArray(text *PdfTextArray) {
	cb.currentFont()
	cb.content.WriteByte('[')
	lastWasNumber := false
	for _, item := range text.items {
		if item.isNumber {
			if lastWasNumber {
				cb.content.WriteByte(' ')
			} else {
				lastWasNumber = true
			}
			cb.num(item.number)
		} else {
			cb.showTextBytes(item.text)
			lastWasNumber = false
		}
	}
	cb.content.WriteString("]TJ\n")
}

// PdfTextArray is the operand of TJ: strings and position adjustments.
type PdfTextArray struct {
	items []textArrayItem

	// lastStr and lastNum are the last string or number added, or nil
	// after a value of the other kind, as in OpenPDF.
	lastStr *string
	lastNum *float32
}

type textArrayItem struct {
	text     string
	number   float32
	isNumber bool
}

// NewPdfTextArray returns an empty text array.
func NewPdfTextArray() *PdfTextArray { return &PdfTextArray{} }

func (a *PdfTextArray) replaceLast(item textArrayItem) {
	a.items[len(a.items)-1] = item
}

// AddString is PdfTextArray.add(String). Adjacent strings are joined.
func (a *PdfTextArray) AddString(s string) {
	if s != "" {
		if a.lastStr != nil {
			joined := *a.lastStr + s
			a.lastStr = &joined
			a.replaceLast(textArrayItem{text: joined})
		} else {
			a.lastStr = &s
			a.items = append(a.items, textArrayItem{text: s})
		}
		a.lastNum = nil
	}
	// adding an empty string doesn't modify the TextArray at all
}

// AddNumber is PdfTextArray.add(float). Adjacent numbers are added up, and
// a sum of zero removes the number.
func (a *PdfTextArray) AddNumber(number float32) {
	if number != 0 {
		if a.lastNum != nil {
			sum := number + *a.lastNum
			a.lastNum = &sum
			if sum != 0 {
				a.replaceLast(textArrayItem{number: sum, isNumber: true})
			} else {
				a.items = a.items[:len(a.items)-1]
			}
		} else {
			a.lastNum = &number
			a.items = append(a.items, textArrayItem{number: number, isNumber: true})
		}

		a.lastStr = nil
	}
	// adding zero doesn't modify the TextArray at all
}

// AddImage writes q a b c d e f cm /ImN Do Q: the image fills the unit
// square transformed by the matrix.
func (cb *PdfContentByte) AddImage(image *Image, a, b, c, d, e, f float32) error {
	if image == nil {
		return newDocumentException("image is nil")
	}
	entry, err := cb.writer.addImage(image)
	if err != nil {
		return err
	}
	cb.resources.xobjects[entry.name] = entry.ref
	cb.content.WriteString("q ")
	cb.op("cm", a, b, c, d, e, f)
	entry.name.WritePdf(&cb.content)
	cb.content.WriteString(" Do Q\n")
	return nil
}

// AddImageAbsolute is PdfContentByte.addImage(image): the image is placed at
// its absolute position with its scaled width and height.
func (cb *PdfContentByte) AddImageAbsolute(image *Image) error {
	if !image.HasAbsolutePosition() {
		return newDocumentException("The image must have absolute positioning.")
	}
	return cb.AddImage(image, image.GetScaledWidth(), 0, 0, image.GetScaledHeight(), image.GetAbsoluteX(), image.GetAbsoluteY())
}

// CreateTemplate returns a form XObject with the bounding box
// [0 0 width height]. Its content may be written before or after it is
// added; it is written to the file when the document is closed.
func (cb *PdfContentByte) CreateTemplate(width, height float32) *PdfTemplate {
	t := &PdfTemplate{bbox: NewRectangle(0, 0, width, height)}
	t.PdfContentByte = *newContentByte(cb.writer)
	return t
}

// AddTemplate writes q a b c d e f cm /XfN Do Q.
func (cb *PdfContentByte) AddTemplate(template *PdfTemplate, a, b, c, d, e, f float32) {
	if template == nil {
		panic(newDocumentException("template is nil"))
	}
	if template.writer != cb.writer {
		panic(newDocumentException("the template belongs to another PdfWriter"))
	}
	cb.writer.addTemplate(template)
	cb.resources.xobjects[template.name] = template.ref
	cb.content.WriteString("q ")
	cb.op("cm", a, b, c, d, e, f)
	template.name.WritePdf(&cb.content)
	cb.content.WriteString(" Do Q\n")
}

// AddTemplateAt is addTemplate(template, x, y).
func (cb *PdfContentByte) AddTemplateAt(template *PdfTemplate, x, y float32) {
	cb.AddTemplate(template, 1, 0, 0, 1, x, y)
}

// BeginMarkedContentSequence writes /Tag BMC.
func (cb *PdfContentByte) BeginMarkedContentSequence(tag PdfName) {
	tag.WritePdf(&cb.content)
	cb.content.WriteString(" BMC\n")
	cb.mcDepth++
}

// BeginMarkedContentSequenceWithProperties writes /Tag <<...>> BDC with the
// property list inline.
func (cb *PdfContentByte) BeginMarkedContentSequenceWithProperties(tag PdfName, properties *PdfDictionary) {
	tag.WritePdf(&cb.content)
	cb.content.WriteByte(' ')
	properties.WritePdf(&cb.content)
	cb.content.WriteString(" BDC\n")
	cb.mcDepth++
}

// EndMarkedContentSequence writes EMC.
func (cb *PdfContentByte) EndMarkedContentSequence() {
	if cb.mcDepth == 0 {
		panic(newDocumentException("Unbalanced marked content operators."))
	}
	cb.mcDepth--
	cb.content.WriteString("EMC\n")
}

// PdfTemplate is a form XObject. It has every operator of PdfContentByte.
type PdfTemplate struct {
	PdfContentByte
	bbox   *Rectangle
	matrix *[6]float32
	name   PdfName
	ref    *PdfIndirectReference
}

// GetWidth returns the width of the bounding box.
func (t *PdfTemplate) GetWidth() float32 { return t.bbox.GetWidth() }

// GetHeight returns the height of the bounding box.
func (t *PdfTemplate) GetHeight() float32 { return t.bbox.GetHeight() }

// SetWidth sets the width of the bounding box.
func (t *PdfTemplate) SetWidth(width float32) { t.bbox.urx = t.bbox.llx + width }

// SetHeight sets the height of the bounding box.
func (t *PdfTemplate) SetHeight(height float32) { t.bbox.ury = t.bbox.lly + height }

// GetBoundingBox returns the bounding box.
func (t *PdfTemplate) GetBoundingBox() *Rectangle { return t.bbox }

// SetBoundingBox sets the bounding box.
func (t *PdfTemplate) SetBoundingBox(bbox *Rectangle) { t.bbox = bbox.clone() }

// SetMatrix sets /Matrix of the form XObject.
func (t *PdfTemplate) SetMatrix(a, b, c, d, e, f float32) { t.matrix = &[6]float32{a, b, c, d, e, f} }

// GetIndirectReference returns the reference of the form XObject, reserving
// it when the template was not added anywhere yet.
func (t *PdfTemplate) GetIndirectReference() *PdfIndirectReference {
	t.writer.addTemplate(t)
	return t.ref
}

// PdfGState is an extended graphics state dictionary.
type PdfGState struct {
	fillOpacity   *float32
	strokeOpacity *float32
	blendMode     PdfName
	overPrintFill *bool
	overPrintStrk *bool
}

// Blend modes, as PdfGState.BM_NORMAL etc.
const (
	PdfGStateBmNormal     PdfName = "Normal"
	PdfGStateBmCompatible PdfName = "Compatible"
	PdfGStateBmMultiply   PdfName = "Multiply"
	PdfGStateBmScreen     PdfName = "Screen"
	PdfGStateBmOverlay    PdfName = "Overlay"
	PdfGStateBmDarken     PdfName = "Darken"
	PdfGStateBmLighten    PdfName = "Lighten"
)

// NewPdfGState returns an empty extended graphics state.
func NewPdfGState() *PdfGState { return &PdfGState{} }

// SetFillOpacity sets /ca, from 0 to 1.
func (g *PdfGState) SetFillOpacity(n float32) { g.fillOpacity = &n }

// SetStrokeOpacity sets /CA, from 0 to 1.
func (g *PdfGState) SetStrokeOpacity(n float32) { g.strokeOpacity = &n }

// SetBlendMode sets /BM.
func (g *PdfGState) SetBlendMode(bm PdfName) { g.blendMode = bm }

// SetOverPrintStroking sets /OP.
func (g *PdfGState) SetOverPrintStroking(ov bool) { g.overPrintStrk = &ov }

// SetOverPrintNonStroking sets /op.
func (g *PdfGState) SetOverPrintNonStroking(ov bool) { g.overPrintFill = &ov }

func (g *PdfGState) dictionary() *PdfDictionary {
	d := NewPdfDictionaryWithType("ExtGState")
	if g.blendMode != "" {
		d.Put("BM", g.blendMode)
	}
	if g.fillOpacity != nil {
		d.Put("ca", PdfNumber(*g.fillOpacity))
	}
	if g.strokeOpacity != nil {
		d.Put("CA", PdfNumber(*g.strokeOpacity))
	}
	if g.overPrintStrk != nil {
		d.Put("OP", PdfBoolean(*g.overPrintStrk))
	}
	if g.overPrintFill != nil {
		d.Put("op", PdfBoolean(*g.overPrintFill))
	}
	return d
}

// key identifies states with the same entries so the writer writes each
// distinct state once.
func (g *PdfGState) key() string {
	var b bytes.Buffer
	g.dictionary().WritePdf(&b)
	return fmt.Sprint(b.String())
}
