package writer

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"
)

// PdfObject is a value that can be serialized into a PDF file: a name, a
// number, a string, an array, a dictionary, a stream or an indirect
// reference.
type PdfObject interface {
	// WritePdf appends the PDF syntax of the object to b.
	WritePdf(b *bytes.Buffer)
}

// DocumentException is the error type for misuse of the writer API. Methods
// that OpenPDF declares without a checked exception (the content stream
// operators) panic with a *DocumentException; methods that do I/O return it.
type DocumentException struct {
	Message string
}

func (e *DocumentException) Error() string { return e.Message }

func newDocumentException(format string, args ...any) *DocumentException {
	return &DocumentException{Message: fmt.Sprintf(format, args...)}
}

// UnsupportedFeatureError is returned by the methods that exist so ported
// code can call them, for features this writer does not implement.
type UnsupportedFeatureError struct {
	Feature string
}

func (e *UnsupportedFeatureError) Error() string {
	return "pdf/writer: " + e.Feature + " is not supported"
}

func unsupported(feature string) error {
	return &UnsupportedFeatureError{Feature: feature}
}

// formatNumber writes a number the way OpenPDF's ByteBuffer.formatDouble
// writes it into a content stream or object (OpenPDF 3.0.5, the version
// Flying Saucer 10 uses), so that ufo's output places everything exactly where
// Flying Saucer's does:
//
//   - a magnitude below 0.000015 is 0;
//   - below 1, it is rounded half up to five decimals and trailing zeros are
//     dropped ("0.5", "0.00002"), and a value that rounds to 1 is 1;
//   - up to 32767, it is rounded half up to two decimals and trailing zeros
//     are dropped ("43.54", "2.68", "10.1", "5");
//   - above 32767, OpenPDF's append path writes nothing at all, which leaves a
//     content stream with a missing operand; this writes the value rounded half
//     up to an integer, which is what OpenPDF's string path returns.
//
// The rounding adds 0.005 (or 0.000005) and truncates, in float64, as
// OpenPDF does; so 1.005 is written as 1, because 1.005 is slightly below
// 1.005 in binary. TestFormatNumberMatchesOpenPDF checks 20,000 values against
// OpenPDF 3.0.5.
func formatNumber(d float64) string {
	if math.IsNaN(d) {
		// Java's (long) NaN is 0.
		return "0"
	}
	if math.Abs(d) < 0.000015 {
		return "0"
	}
	var b strings.Builder
	if d < 0 {
		b.WriteByte('-')
		d = -d
	}
	switch {
	case d < 1.0:
		d += 0.000005
		if d >= 1 {
			b.WriteByte('1')
			return b.String()
		}
		v := int(d * 100000)
		b.WriteString("0.")
		b.WriteByte(byte('0' + v/10000))
		if v%10000 != 0 {
			b.WriteByte(byte('0' + (v/1000)%10))
			if v%1000 != 0 {
				b.WriteByte(byte('0' + (v/100)%10))
				if v%100 != 0 {
					b.WriteByte(byte('0' + (v/10)%10))
					if v%10 != 0 {
						b.WriteByte(byte('0' + v%10))
					}
				}
			}
		}
	case d <= 32767:
		d += 0.005
		v := int(d * 100)
		b.WriteString(strconv.Itoa(v / 100))
		if v%100 != 0 {
			b.WriteByte('.')
			b.WriteByte(byte('0' + (v/10)%10))
			if v%10 != 0 {
				b.WriteByte(byte('0' + v%10))
			}
		}
	default:
		if math.IsInf(d, 0) {
			// No PDF number stands for infinity; OpenPDF writes nothing.
			return "0"
		}
		b.WriteString(strconv.FormatInt(int64(d+0.5), 10))
	}
	return b.String()
}

// delimited is implemented by the objects whose syntax starts with a
// delimiter character: names, strings, arrays and dictionaries.
type delimited interface{ startsWithDelimiter() }

func (PdfName) startsWithDelimiter()        {}
func (*PdfString) startsWithDelimiter()     {}
func (*PdfArray) startsWithDelimiter()      {}
func (*PdfDictionary) startsWithDelimiter() {}

// PdfName is a PDF name object, without the leading slash.
type PdfName string

// NewPdfName returns the name object for s (without a leading slash).
func NewPdfName(s string) PdfName { return PdfName(s) }

func (n PdfName) WritePdf(b *bytes.Buffer) {
	b.WriteByte('/')
	for i := 0; i < len(n); i++ {
		c := n[i]
		if c <= 32 || c >= 127 || strings.IndexByte("#()<>[]{}/%", c) >= 0 {
			fmt.Fprintf(b, "#%02x", c)
		} else {
			b.WriteByte(c)
		}
	}
}

// The names OpenPDF declares as PdfName constants and the Flying Saucer PDF
// module refers to.
const (
	PdfNameA            PdfName = "A"
	PdfNameAlt          PdfName = "Alt"
	PdfNameAnnot        PdfName = "Annot"
	PdfNameAuthor       PdfName = "Author"
	PdfNameCreationdate PdfName = "CreationDate"
	PdfNameCreator      PdfName = "Creator"
	PdfNameD            PdfName = "D"
	PdfNameDests        PdfName = "Dests"
	PdfNameDocument     PdfName = "Document"
	PdfNameF            PdfName = "F"
	PdfNameFigure       PdfName = "Figure"
	PdfNameGoto         PdfName = "GoTo"
	PdfNameH1           PdfName = "H1"
	PdfNameH2           PdfName = "H2"
	PdfNameH3           PdfName = "H3"
	PdfNameH4           PdfName = "H4"
	PdfNameH5           PdfName = "H5"
	PdfNameH6           PdfName = "H6"
	PdfNameJavascript   PdfName = "JavaScript"
	PdfNameJs           PdfName = "JS"
	PdfNameKeywords     PdfName = "Keywords"
	PdfNameL            PdfName = "L"
	PdfNameLang         PdfName = "Lang"
	PdfNameLbody        PdfName = "LBody"
	PdfNameLi           PdfName = "LI"
	PdfNameLink         PdfName = "Link"
	PdfNameModdate      PdfName = "ModDate"
	PdfNameNames        PdfName = "Names"
	PdfNameNonstruct    PdfName = "NonStruct"
	PdfNameO            PdfName = "O"
	PdfNameP            PdfName = "P"
	PdfNameProducer     PdfName = "Producer"
	PdfNameS            PdfName = "S"
	PdfNameSubject      PdfName = "Subject"
	PdfNameSubtype      PdfName = "Subtype"
	PdfNameTable        PdfName = "Table"
	PdfNameTablerow     PdfName = "TR"
	PdfNameTbody        PdfName = "TBody"
	PdfNameTd           PdfName = "TD"
	PdfNameTfoot        PdfName = "TFoot"
	PdfNameTh           PdfName = "TH"
	PdfNameThead        PdfName = "THead"
	PdfNameTitle        PdfName = "Title"
	PdfNameType         PdfName = "Type"
	PdfNameUri          PdfName = "URI"
)

// PdfNumber is a PDF numeric object.
type PdfNumber float64

// NewPdfNumber returns the numeric object for v.
func NewPdfNumber(v float64) PdfNumber { return PdfNumber(v) }

func (n PdfNumber) WritePdf(b *bytes.Buffer) { b.WriteString(formatNumber(float64(n))) }

// PdfBoolean is a PDF boolean object.
type PdfBoolean bool

func (v PdfBoolean) WritePdf(b *bytes.Buffer) {
	if v {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
}

// PdfNull is the PDF null object.
type PdfNull struct{}

func (PdfNull) WritePdf(b *bytes.Buffer) { b.WriteString("null") }

// PdfLiteral is PDF syntax written as given.
type PdfLiteral string

func (l PdfLiteral) WritePdf(b *bytes.Buffer) { b.WriteString(string(l)) }

// Encodings of a PdfString, as OpenPDF's PdfObject.TEXT_PDFDOCENCODING and
// PdfObject.TEXT_UNICODE.
const (
	PdfObjectTextPdfdocencoding = "PDF"
	PdfObjectTextUnicode        = "UnicodeBig"
)

// PdfString is a PDF string object holding text.
//
// With either encoding, text that PDFDocEncoding can represent is written in
// PDFDocEncoding and any other text is written as UTF-16BE with a byte order
// mark. (OpenPDF does the same for TEXT_UNICODE; for TEXT_PDFDOCENCODING it
// drops the characters PDFDocEncoding lacks, which this writer does not do.)
type PdfString struct {
	value    string
	encoding string
	raw      []byte
	hex      bool
}

// NewPdfString returns a text string object in PDFDocEncoding.
func NewPdfString(value string) *PdfString {
	return &PdfString{value: value, encoding: PdfObjectTextPdfdocencoding}
}

// NewPdfStringWithEncoding returns a text string object with the given
// encoding (PdfObjectTextPdfdocencoding or PdfObjectTextUnicode).
func NewPdfStringWithEncoding(value, encoding string) *PdfString {
	return &PdfString{value: value, encoding: encoding}
}

// NewPdfStringBytes returns a string object holding raw bytes.
func NewPdfStringBytes(raw []byte) *PdfString {
	return &PdfString{raw: append([]byte(nil), raw...), encoding: "raw"}
}

// SetHexWriting selects the <...> form.
func (s *PdfString) SetHexWriting(hex bool) *PdfString {
	s.hex = hex
	return s
}

// String returns the text of the string object.
func (s *PdfString) String() string { return s.value }

// GetBytes returns the bytes written between the string delimiters.
func (s *PdfString) GetBytes() []byte {
	if s.encoding == "raw" {
		return s.raw
	}
	return encodeTextString(s.value)
}

func (s *PdfString) WritePdf(b *bytes.Buffer) {
	writeStringBytes(b, s.GetBytes(), s.hex)
}

func writeStringBytes(b *bytes.Buffer, data []byte, hex bool) {
	if hex {
		b.WriteByte('<')
		for _, c := range data {
			fmt.Fprintf(b, "%02x", c)
		}
		b.WriteByte('>')
		return
	}
	b.WriteByte('(')
	for _, c := range data {
		switch c {
		case '\r':
			b.WriteString("\\r")
		case '\n':
			b.WriteString("\\n")
		case '\t':
			b.WriteString("\\t")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		case '(', ')', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte(')')
}

// encodeTextString encodes a PDF text string: PDFDocEncoding when every
// character has a code in it, otherwise UTF-16BE with a byte order mark.
func encodeTextString(text string) []byte {
	out := make([]byte, 0, len(text))
	ok := true
	for _, r := range text {
		c, found := pdfDocEncode(r)
		if !found {
			ok = false
			break
		}
		out = append(out, c)
	}
	if ok {
		return out
	}
	units := utf16.Encode([]rune(text))
	out = make([]byte, 0, 2+2*len(units))
	out = append(out, 0xfe, 0xff)
	for _, u := range units {
		out = append(out, byte(u>>8), byte(u))
	}
	return out
}

// pdfDocEncodingHigh holds the characters of PDFDocEncoding codes 0x80-0xa0.
var pdfDocEncodingHigh = [...]rune{
	0x2022, 0x2020, 0x2021, 0x2026, 0x2014, 0x2013, 0x0192, 0x2044,
	0x2039, 0x203a, 0x2212, 0x2030, 0x201e, 0x201c, 0x201d, 0x2018,
	0x2019, 0x201a, 0x2122, 0xfb01, 0xfb02, 0x0141, 0x0152, 0x0160,
	0x0178, 0x017d, 0x0131, 0x0142, 0x0153, 0x0161, 0x017e, 0xfffd,
	0x20ac,
}

// pdfDocEncodingLow holds the characters of PDFDocEncoding codes 0x18-0x1f.
var pdfDocEncodingLow = [...]rune{0x02d8, 0x02c7, 0x02c6, 0x02d9, 0x02dd, 0x02db, 0x02da, 0x02dc}

func pdfDocEncode(r rune) (byte, bool) {
	switch {
	case r == '\t' || r == '\n' || r == '\r':
		return byte(r), true
	case r >= 0x20 && r < 0x7f:
		return byte(r), true
	case r >= 0xa1 && r <= 0xff && r != 0xad:
		return byte(r), true
	}
	for i, c := range pdfDocEncodingHigh {
		if c == r && c != 0xfffd {
			return byte(0x80 + i), true
		}
	}
	for i, c := range pdfDocEncodingLow {
		if c == r {
			return byte(0x18 + i), true
		}
	}
	return 0, false
}

func pdfDocDecode(c byte) rune {
	switch {
	case c >= 0x18 && c <= 0x1f:
		return pdfDocEncodingLow[c-0x18]
	case c >= 0x80 && c <= 0xa0:
		return pdfDocEncodingHigh[c-0x80]
	}
	return rune(c)
}

// decodeTextString decodes a PDF text string written by encodeTextString (or
// by another producer): UTF-16BE when it starts with the byte order mark,
// UTF-8 when it starts with the UTF-8 mark, otherwise PDFDocEncoding.
func decodeTextString(data []byte) string {
	if len(data) >= 2 && data[0] == 0xfe && data[1] == 0xff {
		units := make([]uint16, 0, len(data)/2)
		for i := 2; i+1 < len(data); i += 2 {
			units = append(units, uint16(data[i])<<8|uint16(data[i+1]))
		}
		return string(utf16.Decode(units))
	}
	if len(data) >= 3 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
		return string(data[3:])
	}
	runes := make([]rune, len(data))
	for i, c := range data {
		runes[i] = pdfDocDecode(c)
	}
	return string(runes)
}

// PdfArray is a PDF array object.
type PdfArray struct {
	items []PdfObject
}

// NewPdfArray returns an array holding items.
func NewPdfArray(items ...PdfObject) *PdfArray {
	return &PdfArray{items: append([]PdfObject(nil), items...)}
}

// NewPdfArrayFloats returns an array of numbers.
func NewPdfArrayFloats(values ...float32) *PdfArray {
	a := &PdfArray{}
	for _, v := range values {
		a.items = append(a.items, PdfNumber(v))
	}
	return a
}

// Add appends an object and reports true, as OpenPDF's PdfArray.add.
func (a *PdfArray) Add(o PdfObject) bool {
	a.items = append(a.items, o)
	return true
}

// Size returns the number of items.
func (a *PdfArray) Size() int { return len(a.items) }

// IsEmpty reports whether the array has no items.
func (a *PdfArray) IsEmpty() bool { return len(a.items) == 0 }

// GetPdfObject returns the item at index i.
func (a *PdfArray) GetPdfObject(i int) PdfObject { return a.items[i] }

func (a *PdfArray) WritePdf(b *bytes.Buffer) {
	b.WriteByte('[')
	for i, o := range a.items {
		if i > 0 {
			b.WriteByte(' ')
		}
		writeObject(b, o)
	}
	b.WriteByte(']')
}

func writeObject(b *bytes.Buffer, o PdfObject) {
	if o == nil {
		b.WriteString("null")
		return
	}
	o.WritePdf(b)
}

// PdfRectangle is the array form of a rectangle: [llx lly urx ury].
func NewPdfRectangle(llx, lly, urx, ury float32) *PdfArray {
	return NewPdfArrayFloats(llx, lly, urx, ury)
}

// PdfDictionary is a PDF dictionary object. Keys are written in the order
// they were first put.
type PdfDictionary struct {
	keys   []PdfName
	values map[PdfName]PdfObject
}

// NewPdfDictionary returns an empty dictionary.
func NewPdfDictionary() *PdfDictionary {
	return &PdfDictionary{values: map[PdfName]PdfObject{}}
}

// NewPdfDictionaryWithType returns a dictionary whose /Type is dicType.
func NewPdfDictionaryWithType(dicType PdfName) *PdfDictionary {
	d := NewPdfDictionary()
	d.Put(PdfNameType, dicType)
	return d
}

// Put sets key to value. A nil value removes the key, as in OpenPDF.
func (d *PdfDictionary) Put(key PdfName, value PdfObject) {
	if value == nil {
		d.Remove(key)
		return
	}
	if _, ok := d.values[key]; !ok {
		d.keys = append(d.keys, key)
	}
	d.values[key] = value
}

// Get returns the value of key, or nil.
func (d *PdfDictionary) Get(key PdfName) PdfObject { return d.values[key] }

// Contains reports whether key is present.
func (d *PdfDictionary) Contains(key PdfName) bool {
	_, ok := d.values[key]
	return ok
}

// Remove deletes key.
func (d *PdfDictionary) Remove(key PdfName) {
	if _, ok := d.values[key]; !ok {
		return
	}
	delete(d.values, key)
	for i, k := range d.keys {
		if k == key {
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			break
		}
	}
}

// GetKeys returns the keys in insertion order.
func (d *PdfDictionary) GetKeys() []PdfName { return append([]PdfName(nil), d.keys...) }

// Size returns the number of entries.
func (d *PdfDictionary) Size() int { return len(d.keys) }

// PutAll copies the entries of other into d.
func (d *PdfDictionary) PutAll(other *PdfDictionary) {
	for _, k := range other.keys {
		d.Put(k, other.values[k])
	}
}

func (d *PdfDictionary) WritePdf(b *bytes.Buffer) {
	b.WriteString("<<")
	for _, k := range d.keys {
		k.WritePdf(b)
		// A value that starts with a delimiter needs no space after the key.
		if _, ok := d.values[k].(delimited); !ok {
			b.WriteByte(' ')
		}
		writeObject(b, d.values[k])
	}
	b.WriteString(">>")
}

// PdfIndirectReference refers to an indirect object by number.
type PdfIndirectReference struct {
	number     int
	generation int
}

// GetNumber returns the object number.
func (r *PdfIndirectReference) GetNumber() int { return r.number }

// GetGeneration returns the generation number, which is always 0.
func (r *PdfIndirectReference) GetGeneration() int { return r.generation }

func (r *PdfIndirectReference) WritePdf(b *bytes.Buffer) {
	fmt.Fprintf(b, "%d %d R", r.number, r.generation)
}

// PdfIndirectObject is an object that was written to the body of the file.
type PdfIndirectObject struct {
	ref    *PdfIndirectReference
	object PdfObject
}

// GetIndirectReference returns the reference to the written object.
func (o *PdfIndirectObject) GetIndirectReference() *PdfIndirectReference { return o.ref }

// PdfStream is a PDF stream object: a dictionary followed by data.
type PdfStream struct {
	PdfDictionary
	data []byte
}

// NewPdfStream returns a stream holding data as given. The caller sets
// /Filter when data is already encoded.
func NewPdfStream(data []byte) *PdfStream {
	return &PdfStream{PdfDictionary: *NewPdfDictionary(), data: data}
}

// FlateCompress deflates the stream data at the given zlib level and sets
// /Filter /FlateDecode. It does nothing when the stream already has a filter.
func (s *PdfStream) FlateCompress(level int) {
	if s.Contains("Filter") {
		return
	}
	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, level)
	if err != nil {
		zw = zlib.NewWriter(&buf)
	}
	zw.Write(s.data)
	zw.Close()
	s.data = buf.Bytes()
	s.Put("Filter", PdfName("FlateDecode"))
}

func (s *PdfStream) WritePdf(b *bytes.Buffer) {
	s.Put("Length", PdfNumber(len(s.data)))
	s.PdfDictionary.WritePdf(b)
	b.WriteString("\nstream\n")
	b.Write(s.data)
	b.WriteString("\nendstream")
}
