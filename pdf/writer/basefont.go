package writer

import (
	"os"
	"strings"
)

// The names of the 14 standard Type 1 fonts, as BaseFont.COURIER etc.
const (
	BaseFontCourier              = "Courier"
	BaseFontCourierBold          = "Courier-Bold"
	BaseFontCourierOblique       = "Courier-Oblique"
	BaseFontCourierBoldoblique   = "Courier-BoldOblique"
	BaseFontHelvetica            = "Helvetica"
	BaseFontHelveticaBold        = "Helvetica-Bold"
	BaseFontHelveticaOblique     = "Helvetica-Oblique"
	BaseFontHelveticaBoldoblique = "Helvetica-BoldOblique"
	BaseFontSymbol               = "Symbol"
	BaseFontTimesRoman           = "Times-Roman"
	BaseFontTimesBold            = "Times-Bold"
	BaseFontTimesItalic          = "Times-Italic"
	BaseFontTimesBolditalic      = "Times-BoldItalic"
	BaseFontZapfdingbats         = "ZapfDingbats"
)

// Keys of GetFontDescriptor, with OpenPDF's values.
const (
	BaseFontAscent                 = 1
	BaseFontCapheight              = 2
	BaseFontDescent                = 3
	BaseFontItalicangle            = 4
	BaseFontBboxllx                = 5
	BaseFontBboxlly                = 6
	BaseFontBboxurx                = 7
	BaseFontBboxury                = 8
	BaseFontAwtAscent              = 9
	BaseFontAwtDescent             = 10
	BaseFontAwtLeading             = 11
	BaseFontAwtMaxadvance          = 12
	BaseFontUnderlinePosition      = 13
	BaseFontUnderlineThickness     = 14
	BaseFontStrikethroughPosition  = 15
	BaseFontStrikethroughThickness = 16
	BaseFontSubscriptSize          = 17
	BaseFontSubscriptOffset        = 18
	BaseFontSuperscriptSize        = 19
	BaseFontSuperscriptOffset      = 20
)

// Font types, as BaseFont.FONT_TYPE_T1 and BaseFont.FONT_TYPE_TTUNI.
const (
	BaseFontFontTypeT1    = 0
	BaseFontFontTypeTtuni = 3
)

// Encodings and the embedded flag, as BaseFont.CP1252 etc.
const (
	BaseFontCp1252      = "Cp1252"
	BaseFontWinansi     = "Cp1252"
	BaseFontIdentityH   = "Identity-H"
	BaseFontEmbedded    = true
	BaseFontNotEmbedded = false
	BaseFontCached      = true
	BaseFontNotCached   = false
)

var base14Fonts = map[string]bool{
	BaseFontCourier: true, BaseFontCourierBold: true, BaseFontCourierOblique: true, BaseFontCourierBoldoblique: true,
	BaseFontHelvetica: true, BaseFontHelveticaBold: true, BaseFontHelveticaOblique: true, BaseFontHelveticaBoldoblique: true,
	BaseFontSymbol: true, BaseFontTimesRoman: true, BaseFontTimesBold: true, BaseFontTimesItalic: true,
	BaseFontTimesBolditalic: true, BaseFontZapfdingbats: true,
}

// BaseFont is a font the writer can measure and show text with: one of the
// 14 standard Type 1 fonts with WinAnsi (Cp1252) encoding, or a TrueType
// font written as a Type 0 font with Identity-H encoding.
//
// A TrueType font is always embedded in full and always written with
// Identity-H, whatever encoding and embedded flag CreateFont is given; the
// arguments are kept for GetEncoding and IsEmbedded reports true.
type BaseFont struct {
	fontType int
	encoding string
	embedded bool

	afm *afmMetrics
	// widths holds the width of each WinAnsi code of a Type 1 font.
	widths [256]int

	tt *ttFont
}

// CreateFont returns the font with the given name: one of the 14 standard
// font names, or the path of a .ttf/.otf file, or "path.ttc,N" for member N
// of a TrueType collection.
// IsBase14Font reports whether name is one of the 14 standard PDF fonts
// ("Helvetica", "Times-Bold", ...), which need no font file.
func IsBase14Font(name string) bool {
	return base14Fonts[name]
}

func CreateFont(name, encoding string, embedded bool) (*BaseFont, error) {
	return CreateFontWithBytes(name, encoding, embedded, true, nil, nil)
}

// CreateFontWithBytes is BaseFont.createFont(name, encoding, embedded,
// cached, ttfAfm, pfb). When ttfAfm is not nil it holds the font file and
// name is used only to tell the kind of font and the collection index.
// Fonts are not cached, so cached has no effect.
func CreateFontWithBytes(name, encoding string, embedded, cached bool, ttfAfm, pfb []byte) (*BaseFont, error) {
	if base14Fonts[name] {
		return createBuiltinFont(name, encoding)
	}
	path, ttcIndex := splitTTCName(name)
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".ttf"), strings.HasSuffix(lower, ".otf"), strings.HasSuffix(lower, ".ttc"):
		data := ttfAfm
		if data == nil {
			var err error
			if data, err = os.ReadFile(path); err != nil {
				return nil, err
			}
		}
		tt, err := parseTrueType(data, ttcIndex, name)
		if err != nil {
			return nil, err
		}
		return &BaseFont{fontType: BaseFontFontTypeTtuni, encoding: encoding, embedded: true, tt: tt}, nil
	case strings.HasSuffix(lower, ".afm"), strings.HasSuffix(lower, ".pfm"):
		return nil, unsupported("Type 1 font from an AFM or PFM file (" + name + ")")
	}
	return nil, newDocumentException("Font '%s' with '%s' is not recognized.", name, encoding)
}

func createBuiltinFont(name, encoding string) (*BaseFont, error) {
	afm, err := loadAFM(name)
	if err != nil {
		return nil, err
	}
	if encoding != BaseFontCp1252 && encoding != "" {
		return nil, unsupported("encoding " + encoding + " for the standard font " + name)
	}
	bf := &BaseFont{fontType: BaseFontFontTypeT1, encoding: BaseFontCp1252, afm: afm}
	// BaseFont.createEncoding: a font-specific font takes the width of the
	// glyph at each code; any other font takes the width of the glyph named
	// after the character the encoding gives the code.
	for k := 0; k < 256; k++ {
		if afm.fontSpecific {
			bf.widths[k] = afm.widthByCode[k]
			continue
		}
		if name, ok := winAnsiGlyphNames[winAnsiDecode(byte(k))]; ok {
			bf.widths[k] = afm.widthByName[name]
		}
	}
	return bf, nil
}

// winAnsiHigh holds the characters of Cp1252 codes 0x80-0x9f; 0 marks an
// undefined code.
var winAnsiHigh = [32]rune{
	0x20ac, 0, 0x201a, 0x0192, 0x201e, 0x2026, 0x2020, 0x2021,
	0x02c6, 0x2030, 0x0160, 0x2039, 0x0152, 0, 0x017d, 0,
	0, 0x2018, 0x2019, 0x201c, 0x201d, 0x2022, 0x2013, 0x2014,
	0x02dc, 0x2122, 0x0161, 0x203a, 0x0153, 0, 0x017e, 0x0178,
}

var winAnsiEncodeMap = func() map[rune]byte {
	m := map[rune]byte{}
	for i, r := range winAnsiHigh {
		if r != 0 {
			m[r] = byte(0x80 + i)
		}
	}
	return m
}()

// winAnsiDecode returns the character of a Cp1252 code; U+FFFD for the five
// undefined codes.
func winAnsiDecode(c byte) rune {
	if c >= 0x80 && c <= 0x9f {
		if r := winAnsiHigh[c-0x80]; r != 0 {
			return r
		}
		return 0xfffd
	}
	return rune(c)
}

// winAnsiEncode returns the Cp1252 code of a character.
func winAnsiEncode(r rune) (byte, bool) {
	if r < 0x80 || (r >= 0xa0 && r <= 0xff) {
		return byte(r), true
	}
	c, ok := winAnsiEncodeMap[r]
	return c, ok
}

// GetFontType returns BaseFontFontTypeT1 or BaseFontFontTypeTtuni.
func (bf *BaseFont) GetFontType() int { return bf.fontType }

// GetEncoding returns the encoding CreateFont was given.
func (bf *BaseFont) GetEncoding() string { return bf.encoding }

// IsEmbedded reports whether the font file is written into the document.
func (bf *BaseFont) IsEmbedded() bool { return bf.embedded }

// IsFontSpecific reports whether the font has its own encoding (Symbol,
// ZapfDingbats, a TrueType font with only a symbol character map).
func (bf *BaseFont) IsFontSpecific() bool {
	if bf.tt != nil {
		return bf.tt.symbolic
	}
	return bf.afm.fontSpecific
}

// GetWidth returns the width of a character in 1000-unit glyph space.
func (bf *BaseFont) GetWidth(char1 rune) int {
	if bf.tt != nil {
		g := bf.tt.glyphIndex(char1)
		if g == 0 {
			return 0
		}
		return bf.tt.glyphWidth(g)
	}
	c, ok := winAnsiEncode(char1)
	if !ok {
		return 0
	}
	return bf.widths[c]
}

// GetWidthString returns the width of text in 1000-unit glyph space, as
// BaseFont.getWidth(String).
func (bf *BaseFont) GetWidthString(text string) int {
	total := 0
	for _, r := range text {
		total += bf.GetWidth(r)
	}
	return total
}

// GetWidthPoint returns the width of text in points:
// getWidth(text) * 0.001f * fontSize in float arithmetic.
func (bf *BaseFont) GetWidthPoint(text string, fontSize float32) float32 {
	return float32(bf.GetWidthString(text)) * 0.001 * fontSize
}

// GetWidthPointChar is BaseFont.getWidthPoint(int, float).
func (bf *BaseFont) GetWidthPointChar(char1 rune, fontSize float32) float32 {
	return float32(bf.GetWidth(char1)) * 0.001 * fontSize
}

// GetWidthPointKerned returns the width of text in points with the kerning
// between each pair of adjacent characters added.
func (bf *BaseFont) GetWidthPointKerned(text string, fontSize float32) float32 {
	size := float32(bf.GetWidthString(text)) * 0.001 * fontSize
	if !bf.HasKernPairs() {
		return size
	}
	kern := 0
	runes := []rune(text)
	for k := 0; k+1 < len(runes); k++ {
		kern += bf.GetKerning(runes[k], runes[k+1])
	}
	return size + float32(kern)*0.001*fontSize
}

// HasKernPairs reports whether the font has kerning data.
func (bf *BaseFont) HasKernPairs() bool {
	if bf.tt != nil {
		return len(bf.tt.kernPairs) > 0
	}
	return len(bf.afm.kernPairs) > 0
}

// GetKerning returns the kerning between two characters in 1000-unit glyph
// space. For a Type 1 font the characters are converted to glyph names and
// the pair is looked up in the KPX entries of the AFM file; for a TrueType
// font the glyph pair is looked up in the kern table.
func (bf *BaseFont) GetKerning(char1, char2 rune) int {
	if bf.tt != nil {
		return bf.tt.kerning(char1, char2)
	}
	first, ok := winAnsiGlyphNames[char1]
	if !ok {
		return 0
	}
	second, ok := winAnsiGlyphNames[char2]
	if !ok {
		return 0
	}
	return bf.afm.kernPairs[first][second]
}

// CharExists reports whether the font can show the character.
func (bf *BaseFont) CharExists(c rune) bool {
	if bf.tt != nil {
		return bf.tt.glyphIndex(c) != 0
	}
	_, ok := winAnsiEncode(c)
	return ok
}

// GetFontDescriptor returns a font metric scaled to fontSize.
//
// For a standard Type 1 font, from the AFM file (Type1Font):
// ASCENT and AWT_ASCENT = Ascender, DESCENT and AWT_DESCENT = Descender,
// CAPHEIGHT = CapHeight, BBOX* = FontBBox, AWT_LEADING = 0,
// AWT_MAXADVANCE = urx - llx, UNDERLINE_POSITION = UnderlinePosition,
// UNDERLINE_THICKNESS = UnderlineThickness, each times fontSize / 1000;
// ITALICANGLE is the angle itself; the strikethrough keys are 0.
//
// For a TrueType font (TrueTypeFont): ASCENT = OS/2 sTypoAscender,
// DESCENT = OS/2 sTypoDescender (made negative), CAPHEIGHT = OS/2
// sCapHeight (0.7 em when the OS/2 version is below 2), BBOX* = head
// xMin/yMin/xMax/yMax, AWT_ASCENT/AWT_DESCENT/AWT_LEADING = hhea
// Ascender/Descender/LineGap, AWT_MAXADVANCE = hhea advanceWidthMax,
// UNDERLINE_POSITION = post underlinePosition - underlineThickness / 2
// (integer division), UNDERLINE_THICKNESS = post underlineThickness,
// STRIKETHROUGH_POSITION/THICKNESS = OS/2 yStrikeoutPosition/yStrikeoutSize,
// SUBSCRIPT/SUPERSCRIPT size and offset = OS/2 ySubscriptYSize,
// -ySubscriptYOffset, ySuperscriptYSize, ySuperscriptYOffset, each times
// fontSize / unitsPerEm; ITALICANGLE is the post table angle.
func (bf *BaseFont) GetFontDescriptor(key int, fontSize float32) float32 {
	if bf.tt != nil {
		t := bf.tt
		upem := float32(t.unitsPerEm)
		switch key {
		case BaseFontAscent:
			return float32(t.sTypoAscender) * fontSize / upem
		case BaseFontCapheight:
			return float32(t.sCapHeight) * fontSize / upem
		case BaseFontDescent:
			return float32(t.sTypoDescender) * fontSize / upem
		case BaseFontItalicangle:
			return float32(t.italicAngle)
		case BaseFontBboxllx:
			return fontSize * float32(t.xMin) / upem
		case BaseFontBboxlly:
			return fontSize * float32(t.yMin) / upem
		case BaseFontBboxurx:
			return fontSize * float32(t.xMax) / upem
		case BaseFontBboxury:
			return fontSize * float32(t.yMax) / upem
		case BaseFontAwtAscent:
			return fontSize * float32(t.hheaAscender) / upem
		case BaseFontAwtDescent:
			return fontSize * float32(t.hheaDescender) / upem
		case BaseFontAwtLeading:
			return fontSize * float32(t.hheaLineGap) / upem
		case BaseFontAwtMaxadvance:
			return fontSize * float32(t.advanceWidthMax) / upem
		case BaseFontUnderlinePosition:
			return float32(t.underlinePosition-t.underlineThickness/2) * fontSize / upem
		case BaseFontUnderlineThickness:
			return float32(t.underlineThickness) * fontSize / upem
		case BaseFontStrikethroughPosition:
			return float32(t.yStrikeoutPosition) * fontSize / upem
		case BaseFontStrikethroughThickness:
			return float32(t.yStrikeoutSize) * fontSize / upem
		case BaseFontSubscriptSize:
			return float32(t.ySubscriptYSize) * fontSize / upem
		case BaseFontSubscriptOffset:
			return float32(-t.ySubscriptYOffset) * fontSize / upem
		case BaseFontSuperscriptSize:
			return float32(t.ySuperscriptYSize) * fontSize / upem
		case BaseFontSuperscriptOffset:
			return float32(t.ySuperscriptYOffset) * fontSize / upem
		}
		return 0
	}
	a := bf.afm
	switch key {
	case BaseFontAwtAscent, BaseFontAscent:
		return float32(a.ascender) * fontSize / 1000
	case BaseFontCapheight:
		return float32(a.capHeight) * fontSize / 1000
	case BaseFontAwtDescent, BaseFontDescent:
		return float32(a.descender) * fontSize / 1000
	case BaseFontItalicangle:
		return a.italicAngle
	case BaseFontBboxllx:
		return float32(a.llx) * fontSize / 1000
	case BaseFontBboxlly:
		return float32(a.lly) * fontSize / 1000
	case BaseFontBboxurx:
		return float32(a.urx) * fontSize / 1000
	case BaseFontBboxury:
		return float32(a.ury) * fontSize / 1000
	case BaseFontAwtLeading:
		return 0
	case BaseFontAwtMaxadvance:
		return float32(a.urx-a.llx) * fontSize / 1000
	case BaseFontUnderlinePosition:
		return float32(a.underlinePosition) * fontSize / 1000
	case BaseFontUnderlineThickness:
		return float32(a.underlineThickness) * fontSize / 1000
	}
	return 0
}

// GetPostscriptFontName returns the PostScript name of the font.
func (bf *BaseFont) GetPostscriptFontName() string {
	if bf.tt != nil {
		return bf.tt.fontName
	}
	return bf.afm.fontName
}

// GetFamilyFontName returns the family names. Each entry is
// {platformID, encodingID, languageID, name}; a Type 1 font has one entry
// whose first three strings are empty.
func (bf *BaseFont) GetFamilyFontName() [][]string {
	if bf.tt != nil {
		return bf.tt.familyName
	}
	return [][]string{{"", "", "", bf.afm.familyName}}
}

// GetFullFontName returns the full names, in the form of GetFamilyFontName.
func (bf *BaseFont) GetFullFontName() [][]string {
	if bf.tt != nil {
		return bf.tt.fullName
	}
	return [][]string{{"", "", "", bf.afm.fullName}}
}

// GetAllNameEntries returns every name record as
// {nameID, platformID, encodingID, languageID, name}.
func (bf *BaseFont) GetAllNameEntries() [][]string {
	if bf.tt != nil {
		return bf.tt.allNames
	}
	return [][]string{
		{"4", "", "", "", bf.afm.fullName},
		{"1", "", "", "", bf.afm.familyName},
		{"6", "", "", "", bf.afm.fontName},
	}
}

// GetXHeight returns XHeight of the AFM file in 1000-unit glyph space; 0 for
// a TrueType font.
func (bf *BaseFont) GetXHeight() int {
	if bf.afm != nil {
		return bf.afm.xHeight
	}
	return 0
}

// GetStdVW returns StdVW of the AFM file; 80 for a TrueType font, which is
// the StemV OpenPDF writes for every TrueType font.
func (bf *BaseFont) GetStdVW() int {
	if bf.afm != nil {
		return bf.afm.stdVW
	}
	return 80
}

// GetWeight returns the Weight entry of the AFM file ("Medium", "Bold");
// empty for a TrueType font, whose weight is GetOS2WeightClass.
func (bf *BaseFont) GetWeight() string {
	if bf.afm != nil {
		return bf.afm.weight
	}
	return ""
}

// GetCharBBox returns {llx, lly, urx, ury} of the glyph of a character in
// 1000-unit glyph space for a Type 1 font, and false when the character has
// no glyph. It returns false for a TrueType font.
func (bf *BaseFont) GetCharBBox(c rune) ([4]int, bool) {
	if bf.afm == nil {
		return [4]int{}, false
	}
	code, ok := winAnsiEncode(c)
	if !ok {
		return [4]int{}, false
	}
	if bf.afm.fontSpecific {
		box, ok := bf.afm.bboxByCode[int(code)]
		return box, ok
	}
	box, ok := bf.afm.bboxByName[winAnsiGlyphNames[c]]
	return box, ok
}

// IsFixedPitch reports the IsFixedPitch entry of the AFM file or the
// isFixedPitch field of the post table.
func (bf *BaseFont) IsFixedPitch() bool {
	if bf.tt != nil {
		return bf.tt.isFixedPitch
	}
	return bf.afm.isFixedPitch
}

// IsItalic reports, for a TrueType font, whether OS/2 fsSelection has the
// ITALIC or OBLIQUE bit or head macStyle has the italic bit; for a Type 1
// font, whether ItalicAngle is not 0.
func (bf *BaseFont) IsItalic() bool {
	if bf.tt != nil {
		return bf.tt.fsSelection&0x0201 != 0 || bf.tt.macStyle&2 != 0
	}
	return bf.afm.italicAngle != 0
}

// IsBold reports, for a TrueType font, whether OS/2 fsSelection has the BOLD
// bit or head macStyle has the bold bit; for a Type 1 font, whether Weight
// is "Bold".
func (bf *BaseFont) IsBold() bool {
	if bf.tt != nil {
		return bf.tt.fsSelection&0x20 != 0 || bf.tt.macStyle&1 != 0
	}
	return bf.afm.weight == "Bold"
}

// GetOS2WeightClass returns OS/2 usWeightClass; for a Type 1 font 700 when
// Weight is "Bold" and 400 otherwise.
func (bf *BaseFont) GetOS2WeightClass() int {
	if bf.tt != nil {
		return bf.tt.usWeightClass
	}
	if bf.afm.weight == "Bold" {
		return 700
	}
	return 400
}

// GetOS2FsSelection returns OS/2 fsSelection; 0 for a Type 1 font.
func (bf *BaseFont) GetOS2FsSelection() int {
	if bf.tt != nil {
		return bf.tt.fsSelection
	}
	return 0
}

// GetHeadMacStyle returns head macStyle; 0 for a Type 1 font.
func (bf *BaseFont) GetHeadMacStyle() int {
	if bf.tt != nil {
		return bf.tt.macStyle
	}
	return 0
}

// GetUnitsPerEm returns head unitsPerEm; 1000 for a Type 1 font.
func (bf *BaseFont) GetUnitsPerEm() int {
	if bf.tt != nil {
		return bf.tt.unitsPerEm
	}
	return 1000
}

// GetOS2StrikeoutSize and the three accessors after it return the raw table
// values, in font units, that TrueTypeUtil.readFontDecorations reads from
// the OS/2 and post tables. They are 0 for a Type 1 font, except the
// underline values, which come from the AFM file.
func (bf *BaseFont) GetOS2StrikeoutSize() int {
	if bf.tt != nil {
		return bf.tt.yStrikeoutSize
	}
	return 0
}

// GetOS2StrikeoutPosition returns OS/2 yStrikeoutPosition in font units.
func (bf *BaseFont) GetOS2StrikeoutPosition() int {
	if bf.tt != nil {
		return bf.tt.yStrikeoutPosition
	}
	return 0
}

// GetPostUnderlinePosition returns post underlinePosition in font units.
func (bf *BaseFont) GetPostUnderlinePosition() int {
	if bf.tt != nil {
		return bf.tt.underlinePosition
	}
	return bf.afm.underlinePosition
}

// GetPostUnderlineThickness returns post underlineThickness in font units.
func (bf *BaseFont) GetPostUnderlineThickness() int {
	if bf.tt != nil {
		return bf.tt.underlineThickness
	}
	return bf.afm.underlineThickness
}

// GetTables returns, for a TrueType font, {offset, length} of each table
// within GetFontData, keyed by tag: the counterpart of the private
// TrueTypeFont.tables field TrueTypeUtil reads. It is nil for a Type 1 font.
func (bf *BaseFont) GetTables() map[string][2]int {
	if bf.tt == nil {
		return nil
	}
	out := make(map[string][2]int, len(bf.tt.tables))
	for k, v := range bf.tt.tables {
		out[k] = v
	}
	return out
}

// GetFontData returns the font file that is embedded: the file CreateFont
// read, or the standalone font extracted from a collection. It is nil for a
// Type 1 font. The slice must not be modified.
func (bf *BaseFont) GetFontData() []byte {
	if bf.tt == nil {
		return nil
	}
	return bf.tt.data
}

// convertToBytes returns the bytes that show text with this font: WinAnsi
// codes for a Type 1 font, two-byte glyph ids for a TrueType font. A
// character the font cannot show is left out, as in OpenPDF. used, when not
// nil, records the character each glyph id was used for.
func (bf *BaseFont) convertToBytes(text string, used map[uint16]rune) []byte {
	out := make([]byte, 0, len(text))
	if bf.tt != nil {
		for _, r := range text {
			g := bf.tt.glyphIndex(r)
			if g == 0 {
				continue
			}
			if used != nil {
				if _, ok := used[g]; !ok {
					used[g] = r
				}
			}
			out = append(out, byte(g>>8), byte(g))
		}
		return out
	}
	for _, r := range text {
		if c, ok := winAnsiEncode(r); ok {
			out = append(out, c)
		}
	}
	return out
}
