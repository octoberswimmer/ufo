package writer

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"

	"golang.org/x/image/font/sfnt"
)

// ttFont is a parsed TrueType font. The character map is read through
// golang.org/x/image/font/sfnt; the head, hhea, OS/2, post, name, hmtx and
// kern tables are read directly so the values (and the integer
// arithmetic on them) are the ones OpenPDF's TrueTypeFont uses.
type ttFont struct {
	// data is a standalone font file: the file given, or the member
	// extracted from a TrueType collection. It is what gets embedded, and
	// table offsets are relative to it.
	data   []byte
	tables map[string][2]int

	mu  sync.Mutex
	sf  *sfnt.Font
	buf sfnt.Buffer

	unitsPerEm             int
	xMin, yMin, xMax, yMax int
	macStyle               int

	hheaAscender, hheaDescender, hheaLineGap int
	advanceWidthMax                          int
	caretSlopeRise, caretSlopeRun            int
	numberOfHMetrics                         int

	usWeightClass                      int
	usWidthClass                       int
	fsType                             int
	ySubscriptYSize, ySubscriptYOffset int
	ySuperscriptYSize                  int
	ySuperscriptYOffset                int
	yStrikeoutSize, yStrikeoutPosition int
	fsSelection                        int
	sTypoAscender, sTypoDescender      int
	sTypoLineGap                       int
	usWinAscent, usWinDescent          int
	sCapHeight                         int

	italicAngle        float64
	underlinePosition  int
	underlineThickness int
	isFixedPitch       bool

	numGlyphs   int
	glyphWidths []int // 1000-unit glyph space, by glyph id
	symbolic    bool

	fontName   string
	fullName   [][]string
	familyName [][]string
	allNames   [][]string

	glyphCache map[rune]uint16
	kernPairs  map[uint32]int
}

// splitTTCName splits "path.ttc,2" into the path and the collection index.
// The index is -1 when the name has none.
func splitTTCName(name string) (string, int) {
	lower := strings.ToLower(name)
	i := strings.Index(lower, ".ttc,")
	if i < 0 {
		return name, -1
	}
	idx, err := strconv.Atoi(strings.TrimSpace(name[i+5:]))
	if err != nil {
		return name[:i+4], -1
	}
	return name[:i+4], idx
}

func isTTC(data []byte) bool {
	return len(data) >= 12 && string(data[:4]) == "ttcf"
}

// ttcOffsets returns the offset of each member font of a collection.
func ttcOffsets(data []byte) ([]int, error) {
	if !isTTC(data) {
		return nil, newDocumentException("not a valid TTC file")
	}
	n := int(binary.BigEndian.Uint32(data[8:12]))
	if n <= 0 || 12+4*n > len(data) {
		return nil, newDocumentException("not a valid TTC file")
	}
	offsets := make([]int, n)
	for i := range offsets {
		offsets[i] = int(binary.BigEndian.Uint32(data[12+4*i:]))
	}
	return offsets, nil
}

// extractTTCMember builds a standalone font file from the member of a
// collection whose table directory starts at dirOffset: the same tables,
// copied, behind a new table directory.
func extractTTCMember(data []byte, dirOffset int) ([]byte, error) {
	if dirOffset < 0 || dirOffset+12 > len(data) {
		return nil, newDocumentException("not a valid TTC file")
	}
	numTables := int(binary.BigEndian.Uint16(data[dirOffset+4:]))
	if dirOffset+12+16*numTables > len(data) {
		return nil, newDocumentException("not a valid TTC file")
	}
	headerLen := 12 + 16*numTables
	out := make([]byte, headerLen)
	copy(out, data[dirOffset:dirOffset+12])
	for i := 0; i < numTables; i++ {
		rec := data[dirOffset+12+16*i : dirOffset+12+16*(i+1)]
		off := int(binary.BigEndian.Uint32(rec[8:]))
		length := int(binary.BigEndian.Uint32(rec[12:]))
		if off < 0 || length < 0 || off+length > len(data) {
			return nil, newDocumentException("not a valid TTC file")
		}
		newOff := len(out)
		out = append(out, data[off:off+length]...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
		dst := out[12+16*i : 12+16*(i+1)]
		copy(dst, rec[:8])
		binary.BigEndian.PutUint32(dst[8:], uint32(newOff))
		binary.BigEndian.PutUint32(dst[12:], uint32(length))
	}
	return out, nil
}

// EnumerateTTCNames returns the full font name of each member of the
// TrueType collection file at path, as BaseFont.enumerateTTCNames.
func EnumerateTTCNames(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return EnumerateTTCNamesBytes(data)
}

// EnumerateTTCNamesBytes is EnumerateTTCNames for a collection in memory.
func EnumerateTTCNamesBytes(data []byte) ([]string, error) {
	offsets, err := ttcOffsets(data)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(offsets))
	for i, off := range offsets {
		member, err := extractTTCMember(data, off)
		if err != nil {
			return nil, err
		}
		tt := &ttFont{data: member}
		if err := tt.readDirectory(); err != nil {
			return nil, err
		}
		full, err := tt.getNames(4)
		if err != nil {
			return nil, err
		}
		if len(full) > 0 {
			names[i] = full[0][3]
		}
	}
	return names, nil
}

// parseTrueType parses a font file (or the member ttcIndex of a collection;
// -1 selects member 0).
func parseTrueType(data []byte, ttcIndex int, label string) (*ttFont, error) {
	if isTTC(data) {
		offsets, err := ttcOffsets(data)
		if err != nil {
			return nil, err
		}
		if ttcIndex < 0 {
			ttcIndex = 0
		}
		if ttcIndex >= len(offsets) {
			return nil, newDocumentException("The font index for %s must be between 0 and %d. It was %d.", label, len(offsets)-1, ttcIndex)
		}
		member, err := extractTTCMember(data, offsets[ttcIndex])
		if err != nil {
			return nil, err
		}
		data = member
	}
	tt := &ttFont{data: data, glyphCache: map[rune]uint16{}}
	if err := tt.readDirectory(); err != nil {
		return nil, newDocumentException("%s is not a valid TTF or OTF file.", label)
	}
	if _, ok := tt.tables["CFF "]; ok {
		return nil, unsupported("OpenType font with CFF outlines (" + label + ")")
	}
	if _, ok := tt.tables["glyf"]; !ok {
		return nil, newDocumentException("Table 'glyf' does not exist in %s", label)
	}
	sf, err := sfnt.Parse(data)
	if err != nil {
		return nil, newDocumentException("%s is not a valid TTF file: %v", label, err)
	}
	tt.sf = sf
	if err := tt.fillTables(label); err != nil {
		return nil, err
	}
	return tt, nil
}

func (tt *ttFont) readDirectory() error {
	d := tt.data
	if len(d) < 12 {
		return errors.New("short font file")
	}
	version := binary.BigEndian.Uint32(d)
	if version != 0x00010000 && version != 0x4f54544f && version != 0x74727565 {
		return errors.New("bad sfnt version")
	}
	numTables := int(binary.BigEndian.Uint16(d[4:]))
	if 12+16*numTables > len(d) {
		return errors.New("short table directory")
	}
	tt.tables = make(map[string][2]int, numTables)
	for i := 0; i < numTables; i++ {
		rec := d[12+16*i:]
		off := int(binary.BigEndian.Uint32(rec[8:]))
		length := int(binary.BigEndian.Uint32(rec[12:]))
		if off+length > len(d) {
			return errors.New("table outside file")
		}
		tt.tables[string(rec[:4])] = [2]int{off, length}
	}
	return nil
}

// table returns the bytes of a table, or nil.
func (tt *ttFont) table(tag string) []byte {
	loc, ok := tt.tables[tag]
	if !ok {
		return nil
	}
	return tt.data[loc[0] : loc[0]+loc[1]]
}

func u16(b []byte, off int) int {
	if off+2 > len(b) {
		return 0
	}
	return int(binary.BigEndian.Uint16(b[off:]))
}

func s16(b []byte, off int) int {
	if off+2 > len(b) {
		return 0
	}
	return int(int16(binary.BigEndian.Uint16(b[off:])))
}

func u32(b []byte, off int) int {
	if off+4 > len(b) {
		return 0
	}
	return int(binary.BigEndian.Uint32(b[off:]))
}

// fillTables reads the tables in the order, and with the fallbacks, of
// TrueTypeFont.fillTables.
func (tt *ttFont) fillTables(label string) error {
	head := tt.table("head")
	if len(head) < 54 {
		return newDocumentException("Table 'head' does not exist in %s", label)
	}
	tt.unitsPerEm = u16(head, 18)
	if tt.unitsPerEm == 0 {
		return newDocumentException("%s has unitsPerEm 0", label)
	}
	tt.xMin, tt.yMin, tt.xMax, tt.yMax = s16(head, 36), s16(head, 38), s16(head, 40), s16(head, 42)
	tt.macStyle = u16(head, 44)

	hhea := tt.table("hhea")
	if len(hhea) < 36 {
		return newDocumentException("Table 'hhea' does not exist in %s", label)
	}
	tt.hheaAscender, tt.hheaDescender, tt.hheaLineGap = s16(hhea, 4), s16(hhea, 6), s16(hhea, 8)
	tt.advanceWidthMax = u16(hhea, 10)
	tt.caretSlopeRise, tt.caretSlopeRun = s16(hhea, 18), s16(hhea, 20)
	tt.numberOfHMetrics = u16(hhea, 34)

	os2 := tt.table("OS/2")
	if os2 == nil {
		return newDocumentException("Table 'OS/2' does not exist in %s", label)
	}
	version := u16(os2, 0)
	tt.usWeightClass = u16(os2, 4)
	tt.usWidthClass = u16(os2, 6)
	tt.fsType = s16(os2, 8)
	tt.ySubscriptYSize = s16(os2, 12)
	tt.ySubscriptYOffset = s16(os2, 16)
	tt.ySuperscriptYSize = s16(os2, 20)
	tt.ySuperscriptYOffset = s16(os2, 24)
	tt.yStrikeoutSize = s16(os2, 26)
	tt.yStrikeoutPosition = s16(os2, 28)
	tt.fsSelection = u16(os2, 62)
	tt.sTypoAscender = s16(os2, 68)
	tt.sTypoDescender = s16(os2, 70)
	if tt.sTypoDescender > 0 {
		tt.sTypoDescender = -tt.sTypoDescender
	}
	tt.sTypoLineGap = s16(os2, 72)
	tt.usWinAscent = u16(os2, 74)
	tt.usWinDescent = u16(os2, 76)
	if version > 1 {
		tt.sCapHeight = s16(os2, 88)
	} else {
		tt.sCapHeight = int(0.7 * float64(tt.unitsPerEm))
	}

	post := tt.table("post")
	if post == nil {
		tt.italicAngle = -math.Atan2(float64(tt.caretSlopeRun), float64(tt.caretSlopeRise)) * 180 / math.Pi
	} else {
		mantissa := s16(post, 4)
		fraction := u16(post, 6)
		tt.italicAngle = float64(mantissa) + float64(fraction)/16384.0
		tt.underlinePosition = s16(post, 8)
		tt.underlineThickness = s16(post, 10)
		tt.isFixedPitch = u32(post, 12) != 0
	}

	maxp := tt.table("maxp")
	tt.numGlyphs = u16(maxp, 4)
	if tt.numGlyphs == 0 {
		tt.numGlyphs = tt.sf.NumGlyphs()
	}
	hmtx := tt.table("hmtx")
	if hmtx == nil {
		return newDocumentException("Table 'hmtx' does not exist in %s", label)
	}
	tt.glyphWidths = make([]int, tt.numGlyphs)
	last := 0
	for g := 0; g < tt.numGlyphs; g++ {
		if g < tt.numberOfHMetrics {
			last = (u16(hmtx, 4*g) * 1000) / tt.unitsPerEm
		}
		tt.glyphWidths[g] = last
	}

	tt.symbolic = tt.hasSymbolCmapOnly()

	var err error
	if tt.familyName, err = tt.getNames(1); err != nil {
		return err
	}
	if tt.fullName, err = tt.getNames(4); err != nil {
		return err
	}
	if tt.allNames, err = tt.getAllNames(); err != nil {
		return err
	}
	tt.fontName = tt.getBaseFont()

	tt.readKerning()
	return nil
}

// readKerning reads the horizontal format 0 subtables of the kern table, as
// TrueTypeFont.readKerning. Values are stored in 1000-unit glyph space.
func (tt *ttFont) readKerning() {
	kern := tt.table("kern")
	if kern == nil {
		return
	}
	nTables := u16(kern, 2)
	checkpoint := 4
	for k := 0; k < nTables; k++ {
		length := u16(kern, checkpoint+2)
		coverage := u16(kern, checkpoint+4)
		if coverage&0xfff7 == 0x0001 {
			nPairs := u16(kern, checkpoint+6)
			pos := checkpoint + 14
			for j := 0; j < nPairs && pos+6 <= len(kern); j++ {
				pair := uint32(u16(kern, pos))<<16 | uint32(u16(kern, pos+2))
				value := s16(kern, pos+4) * 1000 / tt.unitsPerEm
				if tt.kernPairs == nil {
					tt.kernPairs = map[uint32]int{}
				}
				tt.kernPairs[pair] = value
				pos += 6
			}
		}
		if length == 0 {
			break
		}
		checkpoint += length
	}
}

// hasSymbolCmapOnly reports whether the cmap table has a (3,0) symbol
// subtable and no (3,1) or (3,10) Unicode subtable, which is what makes
// TrueTypeFont treat the font as font-specific.
func (tt *ttFont) hasSymbolCmapOnly() bool {
	cmap := tt.table("cmap")
	n := u16(cmap, 2)
	symbol, unicode := false, false
	for i := 0; i < n; i++ {
		platform, enc := u16(cmap, 4+8*i), u16(cmap, 6+8*i)
		if platform == 3 && enc == 0 {
			symbol = true
		}
		if platform == 3 && (enc == 1 || enc == 10) || platform == 0 {
			unicode = true
		}
	}
	return symbol && !unicode
}

// nameRecords returns {platformID, encodingID, languageID, nameID, text} for
// every record of the name table.
func (tt *ttFont) nameRecords() ([][5]string, error) {
	name := tt.table("name")
	if name == nil {
		return nil, newDocumentException("Table 'name' does not exist")
	}
	numRecords := u16(name, 2)
	startOfStorage := u16(name, 4)
	var out [][5]string
	for k := 0; k < numRecords; k++ {
		rec := 6 + 12*k
		if rec+12 > len(name) {
			break
		}
		platformID, encodingID, languageID := u16(name, rec), u16(name, rec+2), u16(name, rec+4)
		nameID, length, offset := u16(name, rec+6), u16(name, rec+8), u16(name, rec+10)
		start := startOfStorage + offset
		if start+length > len(name) {
			continue
		}
		raw := name[start : start+length]
		var text string
		if platformID == 0 || platformID == 3 || (platformID == 2 && encodingID == 1) {
			units := make([]uint16, 0, len(raw)/2)
			for i := 0; i+1 < len(raw); i += 2 {
				units = append(units, uint16(raw[i])<<8|uint16(raw[i+1]))
			}
			text = string(utf16.Decode(units))
		} else {
			// TrueTypeFont.readStandardString decodes with Cp1252.
			runes := make([]rune, len(raw))
			for i, c := range raw {
				runes[i] = winAnsiDecode(c)
			}
			text = string(runes)
		}
		out = append(out, [5]string{strconv.Itoa(platformID), strconv.Itoa(encodingID), strconv.Itoa(languageID), strconv.Itoa(nameID), text})
	}
	return out, nil
}

// getNames returns {platformID, encodingID, languageID, name} for each name
// record with the given name id, as TrueTypeFont.getNames.
func (tt *ttFont) getNames(id int) ([][]string, error) {
	recs, err := tt.nameRecords()
	if err != nil {
		return nil, err
	}
	want := strconv.Itoa(id)
	out := [][]string{}
	for _, r := range recs {
		if r[3] == want {
			out = append(out, []string{r[0], r[1], r[2], r[4]})
		}
	}
	return out, nil
}

// getAllNames returns {nameID, platformID, encodingID, languageID, name} for
// every name record, as TrueTypeFont.getAllNames.
func (tt *ttFont) getAllNames() ([][]string, error) {
	recs, err := tt.nameRecords()
	if err != nil {
		return nil, err
	}
	out := [][]string{}
	for _, r := range recs {
		out = append(out, []string{r[3], r[0], r[1], r[2], r[4]})
	}
	return out, nil
}

// getBaseFont returns the PostScript name: name id 6, or the file-independent
// fallback of the first full name with spaces removed.
func (tt *ttFont) getBaseFont() string {
	names, _ := tt.getNames(6)
	if len(names) > 0 && names[0][3] != "" {
		return names[0][3]
	}
	if len(tt.fullName) > 0 {
		return strings.ReplaceAll(tt.fullName[0][3], " ", "-")
	}
	return "Unknown"
}

// glyphIndex returns the glyph id the character map gives r, or 0.
func (tt *ttFont) glyphIndex(r rune) uint16 {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	if g, ok := tt.glyphCache[r]; ok {
		return g
	}
	g, err := tt.sf.GlyphIndex(&tt.buf, r)
	if err != nil {
		g = 0
	}
	if g == 0 && tt.symbolic && r < 0x100 {
		// A (3,0) symbol character map places the glyphs at U+F000-U+F0FF.
		if g2, err := tt.sf.GlyphIndex(&tt.buf, 0xf000|r); err == nil {
			g = g2
		}
	}
	if int(g) >= tt.numGlyphs {
		g = 0
	}
	tt.glyphCache[r] = uint16(g)
	return uint16(g)
}

// glyphWidth returns the advance of a glyph in 1000-unit glyph space.
func (tt *ttFont) glyphWidth(g uint16) int {
	if int(g) < len(tt.glyphWidths) {
		return tt.glyphWidths[g]
	}
	return 0
}

// kerning returns the kerning between two characters in 1000-unit glyph
// space, 0 when the kern table has no entry for the pair.
func (tt *ttFont) kerning(c1, c2 rune) int {
	if len(tt.kernPairs) == 0 {
		return 0
	}
	g1, g2 := tt.glyphIndex(c1), tt.glyphIndex(c2)
	if g1 == 0 || g2 == 0 {
		return 0
	}
	return tt.kernPairs[uint32(g1)<<16|uint32(g2)]
}
