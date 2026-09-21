package writer

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// This file is a minimal PDF reader for tests: it reads back what this
// package writes (and other simple files with a classic cross-reference
// table and no encryption) and reports the page count, page sizes, text,
// information dictionary, outlines, link annotations and images.

// Values of parsed PDF objects: nil, bool, float64, rdName, rdString,
// []any, rdDict, *rdStream and rdRef.
type (
	rdName   string
	rdString []byte
	rdDict   map[string]any
	rdRef    struct{ num, gen int }
	rdStream struct {
		dict rdDict
		raw  []byte
	}
)

// ReadDocument is a PDF file read by ReadPDF. Page numbers are 0-based.
type ReadDocument struct {
	data    []byte
	version string
	xref    map[int]int
	trailer rdDict
	cache   map[int]any
	pages   []rdDict
	pageNum map[int]int // object number of a page -> index
	fonts   map[int]*rdFont
}

// ReadDestination is a destination of an outline item or a link.
type ReadDestination struct {
	// Page is the 0-based index of the target page, -1 when the destination
	// does not refer to a page of the document.
	Page int
	// Type is the destination type: "XYZ", "Fit", "FitH", "FitV", "FitR",
	// "FitB", "FitBH" or "FitBV".
	Type string
	// Left, Bottom, Right, Top and Zoom hold the parameters the type has;
	// a parameter that is absent or null is 0.
	Left, Bottom, Right, Top, Zoom float64
	// Name is the name of a named destination, when the item used one.
	Name string
}

// ReadOutline is an item of the document outline.
type ReadOutline struct {
	Title    string
	Dest     *ReadDestination
	URI      string
	Open     bool
	Children []*ReadOutline
}

// ReadTextRun is a string shown on a page (one Tj, or one string of a TJ
// array), decoded, with the point where it starts in the space of the page.
type ReadTextRun struct {
	Text string
	X, Y float64
}

// ReadLink is a link annotation.
type ReadLink struct {
	// Rect is [llx lly urx ury].
	Rect [4]float64
	// URI is set for a URI action.
	URI string
	// Dest is set for a GoTo action or a /Dest entry.
	Dest *ReadDestination
	// JavaScript is set for a JavaScript action.
	JavaScript string
}

// ReadImage describes an image XObject a page refers to.
type ReadImage struct {
	Name             string
	Width, Height    int
	BitsPerComponent int
	ColorSpace       string
	Filter           string
	HasSMask         bool
	// Data is the stream data with FlateDecode removed; for a DCTDecode
	// image it is the JPEG file.
	Data []byte
}

// ReadPDF parses the cross-reference table and the page tree of a PDF file.
func ReadPDF(data []byte) (*ReadDocument, error) {
	d := &ReadDocument{data: data, xref: map[int]int{}, cache: map[int]any{}, pageNum: map[int]int{}, fonts: map[int]*rdFont{}}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, fmt.Errorf("pdf reader: the file does not start with %%PDF-")
	}
	end := bytes.IndexAny(data[:min(len(data), 32)], "\r\n")
	if end < 0 {
		return nil, fmt.Errorf("pdf reader: the header line has no end")
	}
	d.version = string(data[5:end])

	idx := bytes.LastIndex(data, []byte("startxref"))
	if idx < 0 {
		return nil, fmt.Errorf("pdf reader: startxref not found")
	}
	p := &rdParser{data: data, pos: idx + len("startxref")}
	off, ok := p.parseObject().(float64)
	if !ok {
		return nil, fmt.Errorf("pdf reader: startxref has no offset")
	}
	seen := map[int]bool{}
	for xrefPos := int(off); ; {
		if seen[xrefPos] || xrefPos < 0 || xrefPos >= len(data) {
			return nil, fmt.Errorf("pdf reader: bad cross-reference offset %d", xrefPos)
		}
		seen[xrefPos] = true
		trailer, err := d.readXrefSection(xrefPos)
		if err != nil {
			return nil, err
		}
		if d.trailer == nil {
			d.trailer = trailer
		}
		prev, ok := trailer["Prev"].(float64)
		if !ok {
			break
		}
		xrefPos = int(prev)
	}
	if _, ok := d.trailer["Encrypt"]; ok {
		return nil, fmt.Errorf("pdf reader: encrypted files are not supported")
	}
	root, ok := d.resolve(d.trailer["Root"]).(rdDict)
	if !ok {
		return nil, fmt.Errorf("pdf reader: the trailer has no /Root")
	}
	pagesRef := root["Pages"]
	if err := d.walkPages(pagesRef, rdDict{}, 0); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *ReadDocument) readXrefSection(pos int) (rdDict, error) {
	p := &rdParser{data: d.data, pos: pos}
	p.skipSpace()
	if !bytes.HasPrefix(d.data[p.pos:], []byte("xref")) {
		return nil, fmt.Errorf("pdf reader: no cross-reference table at offset %d (cross-reference streams are not supported)", pos)
	}
	p.pos += 4
	for {
		p.skipSpace()
		if bytes.HasPrefix(d.data[p.pos:], []byte("trailer")) {
			p.pos += len("trailer")
			break
		}
		start, ok1 := p.parseObject().(float64)
		count, ok2 := p.parseObject().(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("pdf reader: bad cross-reference subsection header at offset %d", p.pos)
		}
		for i := 0; i < int(count); i++ {
			offset, ok1 := p.parseObject().(float64)
			_, ok2 := p.parseObject().(float64)
			p.skipSpace()
			if !ok1 || !ok2 || p.pos >= len(d.data) {
				return nil, fmt.Errorf("pdf reader: bad cross-reference entry for object %d", int(start)+i)
			}
			kind := d.data[p.pos]
			p.pos++
			num := int(start) + i
			if kind == 'n' {
				// The newest section is read first and wins.
				if _, ok := d.xref[num]; !ok {
					d.xref[num] = int(offset)
				}
			} else if kind != 'f' {
				return nil, fmt.Errorf("pdf reader: bad cross-reference entry type %q", kind)
			}
		}
	}
	trailer, ok := p.parseObject().(rdDict)
	if !ok {
		return nil, fmt.Errorf("pdf reader: the trailer dictionary is missing")
	}
	return trailer, nil
}

// object returns indirect object num, nil when the file does not have it.
func (d *ReadDocument) object(num int) any {
	if v, ok := d.cache[num]; ok {
		return v
	}
	off, ok := d.xref[num]
	if !ok || off >= len(d.data) {
		return nil
	}
	p := &rdParser{data: d.data, pos: off}
	n, ok1 := p.parseObject().(float64)
	_, ok2 := p.parseObject().(float64)
	p.skipSpace()
	if !ok1 || !ok2 || int(n) != num || !bytes.HasPrefix(d.data[p.pos:], []byte("obj")) {
		return nil
	}
	p.pos += 3
	v := p.parseObject()
	if dict, ok := v.(rdDict); ok {
		p.skipSpace()
		if bytes.HasPrefix(d.data[p.pos:], []byte("stream")) {
			p.pos += len("stream")
			if p.pos < len(d.data) && d.data[p.pos] == '\r' {
				p.pos++
			}
			if p.pos < len(d.data) && d.data[p.pos] == '\n' {
				p.pos++
			}
			d.cache[num] = nil // guards against a /Length that refers to itself
			length, ok := d.resolve(dict["Length"]).(float64)
			start := p.pos
			endPos := start + int(length)
			if !ok || endPos > len(d.data) || !bytes.HasPrefix(bytes.TrimLeft(d.data[endPos:min(len(d.data), endPos+12)], "\r\n "), []byte("endstream")) {
				i := bytes.Index(d.data[start:], []byte("endstream"))
				if i < 0 {
					return nil
				}
				endPos = start + i
			}
			v = &rdStream{dict: dict, raw: d.data[start:endPos]}
		}
	}
	d.cache[num] = v
	return v
}

func (d *ReadDocument) resolve(v any) any {
	for i := 0; i < 32; i++ {
		ref, ok := v.(rdRef)
		if !ok {
			return v
		}
		v = d.object(ref.num)
	}
	return nil
}

func (d *ReadDocument) dict(v any) rdDict {
	switch t := d.resolve(v).(type) {
	case rdDict:
		return t
	case *rdStream:
		return t.dict
	}
	return nil
}

func (d *ReadDocument) array(v any) []any {
	a, _ := d.resolve(v).([]any)
	return a
}

func (d *ReadDocument) number(v any) float64 {
	f, _ := d.resolve(v).(float64)
	return f
}

func (d *ReadDocument) name(v any) string {
	n, _ := d.resolve(v).(rdName)
	return string(n)
}

func (d *ReadDocument) text(v any) string {
	s, _ := d.resolve(v).(rdString)
	return decodeTextString(s)
}

// streamData returns the decoded data of a stream. FlateDecode is removed;
// DCTDecode data is returned as it is.
func (d *ReadDocument) streamData(s *rdStream) ([]byte, error) {
	var filters []string
	switch f := d.resolve(s.dict["Filter"]).(type) {
	case rdName:
		filters = []string{string(f)}
	case []any:
		for _, x := range f {
			filters = append(filters, d.name(x))
		}
	}
	data := s.raw
	for _, f := range filters {
		switch f {
		case "FlateDecode", "Fl":
			zr, err := zlib.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("pdf reader: FlateDecode: %w", err)
			}
			out, err := io.ReadAll(zr)
			if err != nil {
				return nil, fmt.Errorf("pdf reader: FlateDecode: %w", err)
			}
			if parms := d.dict(s.dict["DecodeParms"]); parms != nil && d.number(parms["Predictor"]) > 1 {
				return nil, fmt.Errorf("pdf reader: FlateDecode predictors are not supported")
			}
			data = out
		case "DCTDecode", "DCT":
		default:
			return nil, fmt.Errorf("pdf reader: filter %s is not supported", f)
		}
	}
	return data, nil
}

var inheritablePageKeys = []string{"Resources", "MediaBox", "CropBox", "Rotate"}

func (d *ReadDocument) walkPages(node any, inherited rdDict, depth int) error {
	if depth > 64 {
		return fmt.Errorf("pdf reader: the page tree is too deep")
	}
	dict := d.dict(node)
	if dict == nil {
		return fmt.Errorf("pdf reader: a page tree node is missing")
	}
	attrs := rdDict{}
	for k, v := range inherited {
		attrs[k] = v
	}
	for _, k := range inheritablePageKeys {
		if v, ok := dict[k]; ok {
			attrs[k] = v
		}
	}
	if d.name(dict["Type"]) == "Pages" || dict["Kids"] != nil {
		for _, kid := range d.array(dict["Kids"]) {
			if err := d.walkPages(kid, attrs, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	page := rdDict{}
	for k, v := range dict {
		page[k] = v
	}
	for k, v := range attrs {
		if _, ok := page[k]; !ok {
			page[k] = v
		}
	}
	if ref, ok := node.(rdRef); ok {
		d.pageNum[ref.num] = len(d.pages)
	}
	d.pages = append(d.pages, page)
	return nil
}

// Version returns the version of the file header, e.g. "1.4".
func (d *ReadDocument) Version() string { return d.version }

// NumPages returns the number of pages.
func (d *ReadDocument) NumPages() int { return len(d.pages) }

// PageSize returns the width and height of the MediaBox of page i.
func (d *ReadDocument) PageSize(i int) (width, height float64) {
	if i < 0 || i >= len(d.pages) {
		return 0, 0
	}
	box := d.array(d.pages[i]["MediaBox"])
	if len(box) < 4 {
		return 0, 0
	}
	return math.Abs(d.number(box[2]) - d.number(box[0])), math.Abs(d.number(box[3]) - d.number(box[1]))
}

// Info returns the text entries of the information dictionary.
func (d *ReadDocument) Info() map[string]string {
	out := map[string]string{}
	for k, v := range d.dict(d.trailer["Info"]) {
		if s, ok := d.resolve(v).(rdString); ok {
			out[k] = decodeTextString(s)
		}
	}
	return out
}

// Catalog returns the entries of the document catalog whose values are
// names, strings, numbers or booleans, as strings.
func (d *ReadDocument) Catalog() map[string]string {
	out := map[string]string{}
	for k, v := range d.dict(d.trailer["Root"]) {
		switch t := d.resolve(v).(type) {
		case rdName:
			out[k] = string(t)
		case rdString:
			out[k] = decodeTextString(t)
		case float64:
			out[k] = formatNumber(t)
		case bool:
			out[k] = strconv.FormatBool(t)
		}
	}
	return out
}

// PageContent returns the decoded content stream of page i; several content
// streams are joined with a newline.
func (d *ReadDocument) PageContent(i int) ([]byte, error) {
	if i < 0 || i >= len(d.pages) {
		return nil, fmt.Errorf("pdf reader: no page %d", i)
	}
	return d.contentOf(d.pages[i]["Contents"])
}

func (d *ReadDocument) contentOf(v any) ([]byte, error) {
	switch t := d.resolve(v).(type) {
	case *rdStream:
		return d.streamData(t)
	case []any:
		var out []byte
		for _, item := range t {
			part, err := d.contentOf(item)
			if err != nil {
				return nil, err
			}
			out = append(out, part...)
			out = append(out, '\n')
		}
		return out, nil
	}
	return nil, nil
}

// PageFonts returns the /BaseFont names of the fonts in the resources of
// page i, sorted.
func (d *ReadDocument) PageFonts(i int) []string {
	if i < 0 || i >= len(d.pages) {
		return nil
	}
	var out []string
	for _, f := range d.dict(d.dict(d.pages[i]["Resources"])["Font"]) {
		out = append(out, d.name(d.dict(f)["BaseFont"]))
	}
	sort.Strings(out)
	return out
}

// PageImages returns the image XObjects in the resources of page i, sorted
// by resource name.
func (d *ReadDocument) PageImages(i int) ([]*ReadImage, error) {
	if i < 0 || i >= len(d.pages) {
		return nil, fmt.Errorf("pdf reader: no page %d", i)
	}
	var out []*ReadImage
	xobjects := d.dict(d.dict(d.pages[i]["Resources"])["XObject"])
	names := make([]string, 0, len(xobjects))
	for k := range xobjects {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		s, ok := d.resolve(xobjects[k]).(*rdStream)
		if !ok || d.name(s.dict["Subtype"]) != "Image" {
			continue
		}
		data, err := d.streamData(s)
		if err != nil {
			return nil, err
		}
		filter := d.name(s.dict["Filter"])
		out = append(out, &ReadImage{
			Name:             k,
			Width:            int(d.number(s.dict["Width"])),
			Height:           int(d.number(s.dict["Height"])),
			BitsPerComponent: int(d.number(s.dict["BitsPerComponent"])),
			ColorSpace:       d.name(s.dict["ColorSpace"]),
			Filter:           filter,
			HasSMask:         s.dict["SMask"] != nil,
			Data:             data,
		})
	}
	return out, nil
}

func (d *ReadDocument) destination(v any) *ReadDestination {
	v = d.resolve(v)
	name := ""
	switch t := v.(type) {
	case rdName:
		name = string(t)
		v = d.resolve(d.dict(d.dict(d.trailer["Root"])["Dests"])[name])
	case rdString:
		name = decodeTextString(t)
		v = d.lookupNameTree(d.dict(d.dict(d.trailer["Root"])["Names"])["Dests"], name, 0)
	}
	if dict, ok := v.(rdDict); ok {
		v = d.resolve(dict["D"])
	}
	arr, ok := v.([]any)
	if !ok || len(arr) < 2 {
		if name != "" {
			return &ReadDestination{Page: -1, Name: name}
		}
		return nil
	}
	dest := &ReadDestination{Page: -1, Name: name, Type: d.name(arr[1])}
	if ref, ok := arr[0].(rdRef); ok {
		if idx, ok := d.pageNum[ref.num]; ok {
			dest.Page = idx
		}
	} else if n, ok := arr[0].(float64); ok {
		dest.Page = int(n)
	}
	param := func(i int) float64 {
		if 2+i < len(arr) {
			return d.number(arr[2+i])
		}
		return 0
	}
	switch dest.Type {
	case "XYZ":
		dest.Left, dest.Top, dest.Zoom = param(0), param(1), param(2)
	case "FitH", "FitBH":
		dest.Top = param(0)
	case "FitV", "FitBV":
		dest.Left = param(0)
	case "FitR":
		dest.Left, dest.Bottom, dest.Right, dest.Top = param(0), param(1), param(2), param(3)
	}
	return dest
}

func (d *ReadDocument) lookupNameTree(node any, name string, depth int) any {
	dict := d.dict(node)
	if dict == nil || depth > 32 {
		return nil
	}
	names := d.array(dict["Names"])
	for i := 0; i+1 < len(names); i += 2 {
		if d.text(names[i]) == name {
			return d.resolve(names[i+1])
		}
	}
	for _, kid := range d.array(dict["Kids"]) {
		if v := d.lookupNameTree(kid, name, depth+1); v != nil {
			return v
		}
	}
	return nil
}

// NamedDestinations returns the destinations of the /Dests name tree of the
// catalog's /Names dictionary, by name.
func (d *ReadDocument) NamedDestinations() map[string]*ReadDestination {
	out := map[string]*ReadDestination{}
	var walk func(node any, depth int)
	walk = func(node any, depth int) {
		dict := d.dict(node)
		if dict == nil || depth > 32 {
			return
		}
		names := d.array(dict["Names"])
		for i := 0; i+1 < len(names); i += 2 {
			name := d.text(names[i])
			if dest := d.destination(names[i+1]); dest != nil {
				dest.Name = name
				out[name] = dest
			}
		}
		for _, kid := range d.array(dict["Kids"]) {
			walk(kid, depth+1)
		}
	}
	walk(d.dict(d.dict(d.trailer["Root"])["Names"])["Dests"], 0)
	return out
}

// Outlines returns the top-level outline items; each item holds its
// children.
func (d *ReadDocument) Outlines() []*ReadOutline {
	root := d.dict(d.dict(d.trailer["Root"])["Outlines"])
	if root == nil {
		return nil
	}
	return d.outlineItems(root["First"], 0)
}

func (d *ReadDocument) outlineItems(first any, depth int) []*ReadOutline {
	var out []*ReadOutline
	seen := map[int]bool{}
	for cur := first; cur != nil && depth < 64; {
		if ref, ok := cur.(rdRef); ok {
			if seen[ref.num] {
				break
			}
			seen[ref.num] = true
		}
		dict := d.dict(cur)
		if dict == nil {
			break
		}
		item := &ReadOutline{Title: d.text(dict["Title"]), Open: d.number(dict["Count"]) >= 0}
		if dict["Dest"] != nil {
			item.Dest = d.destination(dict["Dest"])
		} else if action := d.dict(dict["A"]); action != nil {
			switch d.name(action["S"]) {
			case "GoTo":
				item.Dest = d.destination(action["D"])
			case "URI":
				item.URI = d.text(action["URI"])
			}
		}
		item.Children = d.outlineItems(dict["First"], depth+1)
		out = append(out, item)
		cur = dict["Next"]
	}
	return out
}

// LinkAnnotations returns the link annotations of page i in the order of
// the /Annots array.
func (d *ReadDocument) LinkAnnotations(page int) []*ReadLink {
	if page < 0 || page >= len(d.pages) {
		return nil
	}
	var out []*ReadLink
	for _, a := range d.array(d.pages[page]["Annots"]) {
		dict := d.dict(a)
		if dict == nil || d.name(dict["Subtype"]) != "Link" {
			continue
		}
		link := &ReadLink{}
		if rect := d.array(dict["Rect"]); len(rect) >= 4 {
			for i := 0; i < 4; i++ {
				link.Rect[i] = d.number(rect[i])
			}
		}
		if dict["Dest"] != nil {
			link.Dest = d.destination(dict["Dest"])
		}
		if action := d.dict(dict["A"]); action != nil {
			switch d.name(action["S"]) {
			case "URI":
				link.URI = d.text(action["URI"])
			case "GoTo":
				link.Dest = d.destination(action["D"])
			case "JavaScript":
				link.JavaScript = d.text(action["JS"])
			}
		}
		out = append(out, link)
	}
	return out
}

// Text returns the text of every page, each page followed by a newline.
func (d *ReadDocument) Text() (string, error) {
	var b strings.Builder
	for i := range d.pages {
		t, err := d.PageText(i)
		if err != nil {
			return "", err
		}
		b.WriteString(t)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

// PageText returns the text shown on page i, in the order of the content
// stream (form XObjects are read where they are drawn).
//
// Strings are decoded through the /ToUnicode CMap of the font when it has
// one, otherwise through its simple encoding (WinAnsiEncoding, with
// /Differences). The position of each string is followed through the text
// matrix, the transformation matrix and the glyph widths (/Widths, /W and
// /DW, or the AFM metrics of a standard font), and a separator is put
// between two strings by this rule, where x and y are the start of the new
// string in the space of the page, endX and endY the point where the
// previous string ended, and size the font size scaled by the matrices:
//
//   - |y - endY| > 0.5 * size: a newline;
//   - otherwise x - endX > 0.2 * size, or x - endX < -size: a space;
//   - otherwise nothing.
//
// No separator is added next to existing white space. TJ adjustments move
// the position like any other displacement, so a large adjustment gives a
// space by the same rule.
func (d *ReadDocument) PageText(i int) (string, error) {
	if i < 0 || i >= len(d.pages) {
		return "", fmt.Errorf("pdf reader: no page %d", i)
	}
	content, err := d.PageContent(i)
	if err != nil {
		return "", err
	}
	ex := &textExtractor{doc: d, ctm: rdMatrix{1, 0, 0, 1, 0, 0}, hScale: 1}
	if err := ex.run(content, d.dict(d.pages[i]["Resources"]), 0); err != nil {
		return "", err
	}
	return strings.TrimRight(ex.out.String(), " \n"), nil
}

// PageTextRuns returns the strings shown on page i in the order of the
// content stream, each with its starting point, as PageText reads them.
func (d *ReadDocument) PageTextRuns(i int) ([]ReadTextRun, error) {
	if i < 0 || i >= len(d.pages) {
		return nil, fmt.Errorf("pdf reader: no page %d", i)
	}
	content, err := d.PageContent(i)
	if err != nil {
		return nil, err
	}
	ex := &textExtractor{doc: d, ctm: rdMatrix{1, 0, 0, 1, 0, 0}, hScale: 1}
	if err := ex.run(content, d.dict(d.pages[i]["Resources"]), 0); err != nil {
		return nil, err
	}
	return ex.runs, nil
}

// rdMatrix is [a b c d e f].
type rdMatrix [6]float64

// mul returns m × n: m applied first.
func (m rdMatrix) mul(n rdMatrix) rdMatrix {
	return rdMatrix{
		m[0]*n[0] + m[1]*n[2], m[0]*n[1] + m[1]*n[3],
		m[2]*n[0] + m[3]*n[2], m[2]*n[1] + m[3]*n[3],
		m[4]*n[0] + m[5]*n[2] + n[4], m[4]*n[1] + m[5]*n[3] + n[5],
	}
}

func (m rdMatrix) apply(x, y float64) (float64, float64) {
	return x*m[0] + y*m[2] + m[4], x*m[1] + y*m[3] + m[5]
}

type textExtractor struct {
	doc *ReadDocument
	out strings.Builder

	ctm      rdMatrix
	ctmStack []rdMatrix

	tm, tlm  rdMatrix
	font     *rdFont
	size     float64
	charSp   float64
	wordSp   float64
	hScale   float64
	leading  float64
	rise     float64
	havePrev bool
	endX     float64
	endY     float64

	runs []ReadTextRun
}

func (ex *textExtractor) run(content []byte, resources rdDict, depth int) error {
	if depth > 16 {
		return fmt.Errorf("pdf reader: form XObjects are nested too deep")
	}
	p := &rdParser{data: content}
	var operands []any
	num := func(i int) float64 {
		if i < len(operands) {
			f, _ := operands[i].(float64)
			return f
		}
		return 0
	}
	for {
		p.skipSpace()
		if p.pos >= len(content) {
			return nil
		}
		c := content[p.pos]
		if c == '/' || c == '(' || c == '<' || c == '[' || c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9') {
			operands = append(operands, p.parseObject())
			continue
		}
		op := p.readKeyword()
		if op == "" {
			p.pos++
			continue
		}
		switch op {
		case "true", "false", "null":
			operands = append(operands, nil)
			continue
		case "q":
			ex.ctmStack = append(ex.ctmStack, ex.ctm)
		case "Q":
			if n := len(ex.ctmStack); n > 0 {
				ex.ctm = ex.ctmStack[n-1]
				ex.ctmStack = ex.ctmStack[:n-1]
			}
		case "cm":
			ex.ctm = rdMatrix{num(0), num(1), num(2), num(3), num(4), num(5)}.mul(ex.ctm)
		case "BT":
			ex.tm = rdMatrix{1, 0, 0, 1, 0, 0}
			ex.tlm = ex.tm
		case "Tf":
			if len(operands) >= 2 {
				name, _ := operands[0].(rdName)
				ex.font = ex.doc.font(ex.doc.dict(resources["Font"])[string(name)])
				ex.size = num(1)
			}
		case "Tm":
			ex.tm = rdMatrix{num(0), num(1), num(2), num(3), num(4), num(5)}
			ex.tlm = ex.tm
		case "Td":
			ex.tlm = rdMatrix{1, 0, 0, 1, num(0), num(1)}.mul(ex.tlm)
			ex.tm = ex.tlm
		case "TD":
			ex.leading = -num(1)
			ex.tlm = rdMatrix{1, 0, 0, 1, num(0), num(1)}.mul(ex.tlm)
			ex.tm = ex.tlm
		case "T*":
			ex.nextLine()
		case "TL":
			ex.leading = num(0)
		case "Tc":
			ex.charSp = num(0)
		case "Tw":
			ex.wordSp = num(0)
		case "Tz":
			ex.hScale = num(0) / 100
		case "Ts":
			ex.rise = num(0)
		case "Tj":
			if len(operands) >= 1 {
				if s, ok := operands[0].(rdString); ok {
					ex.show(s)
				}
			}
		case "'":
			ex.nextLine()
			if len(operands) >= 1 {
				if s, ok := operands[0].(rdString); ok {
					ex.show(s)
				}
			}
		case "\"":
			ex.wordSp, ex.charSp = num(0), num(1)
			ex.nextLine()
			if len(operands) >= 3 {
				if s, ok := operands[2].(rdString); ok {
					ex.show(s)
				}
			}
		case "TJ":
			if len(operands) >= 1 {
				arr, _ := operands[0].([]any)
				for _, item := range arr {
					switch t := item.(type) {
					case rdString:
						ex.show(t)
					case float64:
						ex.tm = rdMatrix{1, 0, 0, 1, -t / 1000 * ex.size * ex.hScale, 0}.mul(ex.tm)
					}
				}
			}
		case "Do":
			if len(operands) >= 1 {
				name, _ := operands[0].(rdName)
				if err := ex.doXObject(ex.doc.dict(resources["XObject"])[string(name)], resources, depth); err != nil {
					return err
				}
			}
		case "BI":
			// An inline image: skip to the EI that follows white space.
			i := bytes.Index(content[p.pos:], []byte("ID"))
			if i < 0 {
				return nil
			}
			p.pos += i + 2
			for {
				j := bytes.Index(content[p.pos:], []byte("EI"))
				if j < 0 {
					return nil
				}
				p.pos += j + 2
				if j > 0 && isPdfSpace(content[p.pos-3]) && (p.pos >= len(content) || isPdfSpace(content[p.pos])) {
					break
				}
			}
		}
		operands = operands[:0]
	}
}

func (ex *textExtractor) nextLine() {
	ex.tlm = rdMatrix{1, 0, 0, 1, 0, -ex.leading}.mul(ex.tlm)
	ex.tm = ex.tlm
}

func (ex *textExtractor) doXObject(v any, parentResources rdDict, depth int) error {
	s, ok := ex.doc.resolve(v).(*rdStream)
	if !ok || ex.doc.name(s.dict["Subtype"]) != "Form" {
		return nil
	}
	data, err := ex.doc.streamData(s)
	if err != nil {
		return err
	}
	resources := ex.doc.dict(s.dict["Resources"])
	if resources == nil {
		resources = parentResources
	}
	saved := ex.ctm
	if m := ex.doc.array(s.dict["Matrix"]); len(m) >= 6 {
		var fm rdMatrix
		for i := range fm {
			fm[i] = ex.doc.number(m[i])
		}
		ex.ctm = fm.mul(ex.ctm)
	}
	stack := ex.ctmStack
	ex.ctmStack = nil
	err = ex.run(data, resources, depth+1)
	ex.ctm, ex.ctmStack = saved, stack
	return err
}

func (ex *textExtractor) show(s rdString) {
	if ex.font == nil {
		return
	}
	text, advance := ex.font.decode(s, ex.size, ex.charSp, ex.wordSp)
	advance *= ex.hScale
	m := ex.tm.mul(ex.ctm)
	x, y := m.apply(0, ex.rise)
	size := ex.size * math.Sqrt(math.Abs(m[0]*m[3]-m[1]*m[2]))
	if text != "" {
		if ex.havePrev {
			sep := ""
			dx := x - ex.endX
			switch {
			case math.Abs(y-ex.endY) > 0.5*size:
				sep = "\n"
			case dx > 0.2*size || dx < -size:
				sep = " "
			}
			if sep != "" {
				cur := ex.out.String()
				last, _ := lastRune(cur)
				first := []rune(text)[0]
				if sep == "\n" {
					if last != '\n' {
						ex.out.WriteString(sep)
					}
				} else if !unicode.IsSpace(last) && !unicode.IsSpace(first) {
					ex.out.WriteString(sep)
				}
			}
		}
		ex.out.WriteString(text)
		ex.runs = append(ex.runs, ReadTextRun{Text: text, X: x, Y: y})
	}
	ex.tm = rdMatrix{1, 0, 0, 1, advance, 0}.mul(ex.tm)
	if text != "" {
		m = ex.tm.mul(ex.ctm)
		ex.endX, ex.endY = m.apply(0, ex.rise)
		ex.havePrev = true
	}
}

func lastRune(s string) (rune, bool) {
	if s == "" {
		return 0, false
	}
	r := []rune(s[max(0, len(s)-4):])
	return r[len(r)-1], true
}

// rdFont decodes the strings shown with one font.
type rdFont struct {
	twoByte      bool
	toUnicode    map[int]string
	encoding     [256]rune
	widths       map[int]float64
	defaultWidth float64
	base         *BaseFont
}

// font returns the decoder for a font dictionary.
func (d *ReadDocument) font(v any) *rdFont {
	key := -1
	if ref, ok := v.(rdRef); ok {
		key = ref.num
		if f, ok := d.fonts[key]; ok {
			return f
		}
	}
	dict := d.dict(v)
	if dict == nil {
		return nil
	}
	f := &rdFont{widths: map[int]float64{}}
	if s, ok := d.resolve(dict["ToUnicode"]).(*rdStream); ok {
		if data, err := d.streamData(s); err == nil {
			f.toUnicode = parseToUnicode(data)
		}
	}
	if d.name(dict["Subtype"]) == "Type0" {
		f.twoByte = true
		f.defaultWidth = 1000
		if desc := d.array(dict["DescendantFonts"]); len(desc) > 0 {
			cid := d.dict(desc[0])
			if dw, ok := d.resolve(cid["DW"]).(float64); ok {
				f.defaultWidth = dw
			}
			w := d.array(cid["W"])
			for i := 0; i < len(w); {
				first := int(d.number(w[i]))
				if i+1 >= len(w) {
					break
				}
				if list, ok := d.resolve(w[i+1]).([]any); ok {
					for k, x := range list {
						f.widths[first+k] = d.number(x)
					}
					i += 2
				} else if i+2 < len(w) {
					last := int(d.number(w[i+1]))
					for c := first; c <= last; c++ {
						f.widths[c] = d.number(w[i+2])
					}
					i += 3
				} else {
					break
				}
			}
		}
	} else {
		for c := 0; c < 256; c++ {
			f.encoding[c] = winAnsiDecode(byte(c))
		}
		if enc := d.dict(dict["Encoding"]); enc != nil {
			code := 0
			for _, item := range d.array(enc["Differences"]) {
				switch t := d.resolve(item).(type) {
				case float64:
					code = int(t)
				case rdName:
					if r, ok := glyphNameToRune(string(t)); ok && code < 256 {
						f.encoding[code] = r
					}
					code++
				}
			}
		}
		if ws := d.array(dict["Widths"]); len(ws) > 0 {
			first := int(d.number(dict["FirstChar"]))
			for k, x := range ws {
				f.widths[first+k] = d.number(x)
			}
		} else if name := d.name(dict["BaseFont"]); base14Fonts[name] {
			f.base, _ = CreateFont(name, BaseFontCp1252, false)
		}
	}
	if key >= 0 {
		d.fonts[key] = f
	}
	return f
}

var glyphNameRunes map[string]rune

func glyphNameToRune(name string) (rune, bool) {
	if glyphNameRunes == nil {
		m := make(map[string]rune, len(winAnsiGlyphNames))
		for r, n := range winAnsiGlyphNames {
			m[n] = r
		}
		glyphNameRunes = m
	}
	if r, ok := glyphNameRunes[name]; ok {
		return r, true
	}
	if strings.HasPrefix(name, "uni") && len(name) == 7 {
		if v, err := strconv.ParseUint(name[3:], 16, 32); err == nil {
			return rune(v), true
		}
	}
	return 0, false
}

// decode returns the text of a shown string and the distance the text
// position advances, in unscaled text space units.
func (f *rdFont) decode(s rdString, size, charSp, wordSp float64) (string, float64) {
	var b strings.Builder
	advance := 0.0
	step := 1
	if f.twoByte {
		step = 2
	}
	for i := 0; i+step <= len(s); i += step {
		code := int(s[i])
		if f.twoByte {
			code = code<<8 | int(s[i+1])
		}
		text, mapped := f.toUnicode[code]
		if !mapped && !f.twoByte {
			text = string(f.encoding[code])
		}
		b.WriteString(text)

		width, ok := f.widths[code]
		if !ok {
			switch {
			case f.base != nil:
				width = float64(f.base.widths[code])
			default:
				width = f.defaultWidth
			}
		}
		advance += width/1000*size + charSp
		if !f.twoByte && code == 32 {
			advance += wordSp
		}
	}
	return b.String(), advance
}

// parseToUnicode reads the bfchar and bfrange sections of a ToUnicode CMap.
func parseToUnicode(data []byte) map[int]string {
	out := map[int]string{}
	p := &rdParser{data: data}
	hexCode := func(v any) (int, bool) {
		s, ok := v.(rdString)
		if !ok {
			return 0, false
		}
		code := 0
		for _, c := range s {
			code = code<<8 | int(c)
		}
		return code, true
	}
	utf16Text := func(s rdString) string {
		units := make([]uint16, 0, len(s)/2)
		for i := 0; i+1 < len(s); i += 2 {
			units = append(units, uint16(s[i])<<8|uint16(s[i+1]))
		}
		return string(utf16.Decode(units))
	}
	for {
		p.skipSpace()
		if p.pos >= len(data) {
			return out
		}
		c := data[p.pos]
		if c == '/' || c == '(' || c == '<' || c == '[' || c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9') {
			p.parseObject()
			continue
		}
		switch kw := p.readKeyword(); kw {
		case "":
			p.pos++
		case "beginbfchar":
			for {
				p.skipSpace()
				if p.pos >= len(data) || data[p.pos] != '<' {
					break
				}
				src, ok1 := hexCode(p.parseObject())
				dst, ok2 := p.parseObject().(rdString)
				if ok1 && ok2 {
					out[src] = utf16Text(dst)
				}
			}
		case "beginbfrange":
			for {
				p.skipSpace()
				if p.pos >= len(data) || data[p.pos] != '<' {
					break
				}
				lo, ok1 := hexCode(p.parseObject())
				hi, ok2 := hexCode(p.parseObject())
				dst := p.parseObject()
				if !ok1 || !ok2 || hi-lo > 0xffff {
					continue
				}
				switch t := dst.(type) {
				case rdString:
					runes := []rune(utf16Text(t))
					if len(runes) == 0 {
						continue
					}
					for code := lo; code <= hi; code++ {
						r := append([]rune(nil), runes...)
						r[len(r)-1] += rune(code - lo)
						out[code] = string(r)
					}
				case []any:
					for k, item := range t {
						if s, ok := item.(rdString); ok && lo+k <= hi {
							out[lo+k] = utf16Text(s)
						}
					}
				}
			}
		}
	}
}

// rdParser reads PDF objects from a byte slice.
type rdParser struct {
	data []byte
	pos  int
}

func isPdfSpace(c byte) bool {
	return c == 0 || c == '\t' || c == '\n' || c == '\f' || c == '\r' || c == ' '
}

func isPdfDelimiter(c byte) bool {
	return strings.IndexByte("()<>[]{}/%", c) >= 0
}

func (p *rdParser) skipSpace() {
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if isPdfSpace(c) {
			p.pos++
		} else if c == '%' {
			for p.pos < len(p.data) && p.data[p.pos] != '\n' && p.data[p.pos] != '\r' {
				p.pos++
			}
		} else {
			return
		}
	}
}

// readKeyword reads a run of regular characters.
func (p *rdParser) readKeyword() string {
	start := p.pos
	for p.pos < len(p.data) && !isPdfSpace(p.data[p.pos]) && !isPdfDelimiter(p.data[p.pos]) {
		p.pos++
	}
	return string(p.data[start:p.pos])
}

// parseObject reads one object. A keyword that is not an object (an
// operator, endobj) is returned as nil with the position after it.
func (p *rdParser) parseObject() any {
	p.skipSpace()
	if p.pos >= len(p.data) {
		return nil
	}
	c := p.data[p.pos]
	switch {
	case c == '/':
		p.pos++
		return p.parseName()
	case c == '(':
		return p.parseLiteralString()
	case c == '<':
		if p.pos+1 < len(p.data) && p.data[p.pos+1] == '<' {
			return p.parseDict()
		}
		return p.parseHexString()
	case c == '[':
		p.pos++
		var arr []any
		for {
			p.skipSpace()
			if p.pos >= len(p.data) {
				return arr
			}
			if p.data[p.pos] == ']' {
				p.pos++
				if arr == nil {
					arr = []any{}
				}
				return arr
			}
			before := p.pos
			arr = append(arr, p.parseObject())
			if p.pos == before {
				p.pos++
			}
		}
	case c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9'):
		return p.parseNumberOrRef()
	}
	switch kw := p.readKeyword(); kw {
	case "true":
		return true
	case "false":
		return false
	}
	return nil
}

func (p *rdParser) parseName() rdName {
	var b []byte
	for p.pos < len(p.data) && !isPdfSpace(p.data[p.pos]) && !isPdfDelimiter(p.data[p.pos]) {
		c := p.data[p.pos]
		if c == '#' && p.pos+2 < len(p.data) {
			if v, err := strconv.ParseUint(string(p.data[p.pos+1:p.pos+3]), 16, 8); err == nil {
				b = append(b, byte(v))
				p.pos += 3
				continue
			}
		}
		b = append(b, c)
		p.pos++
	}
	return rdName(b)
}

func (p *rdParser) parseNumberOrRef() any {
	start := p.pos
	tok := p.readKeyword()
	f, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		if p.pos == start {
			p.pos++
		}
		return nil
	}
	// "n g R" is a reference when both numbers are non-negative integers.
	if f >= 0 && f == math.Trunc(f) && !strings.ContainsAny(tok, ".+-") {
		save := p.pos
		p.skipSpace()
		genStart := p.pos
		gen := p.readKeyword()
		if g, err := strconv.Atoi(gen); err == nil && g >= 0 && genStart != p.pos && !strings.ContainsAny(gen, ".+-") {
			p.skipSpace()
			if p.pos < len(p.data) && p.data[p.pos] == 'R' && (p.pos+1 >= len(p.data) || isPdfSpace(p.data[p.pos+1]) || isPdfDelimiter(p.data[p.pos+1])) {
				p.pos++
				return rdRef{num: int(f), gen: g}
			}
		}
		p.pos = save
	}
	return f
}

func (p *rdParser) parseLiteralString() rdString {
	p.pos++
	out := rdString{}
	depth := 1
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		p.pos++
		switch c {
		case '(':
			depth++
			out = append(out, c)
		case ')':
			depth--
			if depth == 0 {
				return out
			}
			out = append(out, c)
		case '\\':
			if p.pos >= len(p.data) {
				return out
			}
			e := p.data[p.pos]
			p.pos++
			switch e {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case 'b':
				out = append(out, '\b')
			case 'f':
				out = append(out, '\f')
			case '\r':
				if p.pos < len(p.data) && p.data[p.pos] == '\n' {
					p.pos++
				}
			case '\n':
			default:
				if e >= '0' && e <= '7' {
					v := int(e - '0')
					for k := 0; k < 2 && p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '7'; k++ {
						v = v*8 + int(p.data[p.pos]-'0')
						p.pos++
					}
					out = append(out, byte(v))
				} else {
					out = append(out, e)
				}
			}
		default:
			out = append(out, c)
		}
	}
	return out
}

func (p *rdParser) parseHexString() rdString {
	p.pos++
	out := rdString{}
	var digits []byte
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		p.pos++
		if c == '>' {
			break
		}
		if !isPdfSpace(c) {
			digits = append(digits, c)
		}
	}
	if len(digits)%2 == 1 {
		digits = append(digits, '0')
	}
	for i := 0; i+1 < len(digits); i += 2 {
		v, err := strconv.ParseUint(string(digits[i:i+2]), 16, 8)
		if err != nil {
			v = 0
		}
		out = append(out, byte(v))
	}
	return out
}

func (p *rdParser) parseDict() rdDict {
	p.pos += 2
	dict := rdDict{}
	for {
		p.skipSpace()
		if p.pos >= len(p.data) {
			return dict
		}
		if p.data[p.pos] == '>' {
			p.pos++
			if p.pos < len(p.data) && p.data[p.pos] == '>' {
				p.pos++
			}
			return dict
		}
		if p.data[p.pos] != '/' {
			p.pos++
			continue
		}
		p.pos++
		key := p.parseName()
		dict[string(key)] = p.parseObject()
	}
}
