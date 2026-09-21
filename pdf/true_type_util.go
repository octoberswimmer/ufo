// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/TrueTypeUtil.java

package pdf

import (
	"fmt"
	"strings"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// TrueTypeUtil uses code from iText's DefaultFontMapper and TrueTypeFont
// classes. See http://sourceforge.net/projects/itext/ for license
// information.
//
// Java reads the private "tables" field of OpenPDF's TrueTypeFont through
// reflection and then reads the font file a second time through a
// RandomAccessFileOrArray. Here the table directory comes from
// writer.BaseFont.GetTables and the bytes from writer.BaseFont.GetFontData,
// so the font file is not opened again and the constant
// AVOID_MEMORY_MAPPED_FILES has no counterpart.

func trueTypeUtilGuessStyle(font *writer.BaseFont) *ufo.IdentValue {
	names := font.GetFullFontName()

	for _, name := range names {
		lower := strings.ToLower(name[3])
		if strings.Contains(lower, "italic") {
			return ufo.IdentValueItalic
		} else if strings.Contains(lower, "oblique") {
			return ufo.IdentValueOblique
		}
	}

	return ufo.IdentValueNormal
}

func TrueTypeUtilGetFamilyNames(font *writer.BaseFont) []string {
	names := font.GetFamilyFontName()

	if len(names) == 1 {
		return []string{names[0][3]}
	}

	var result []string
	for _, name := range names {
		if name[0] == "1" && name[1] == "0" || name[2] == "1033" {
			result = append(result, name[3])
		}
	}

	return result
}

// trueTypeUtilExtractTables returns the table directory of a TrueType font,
// {offset, length} by tag. Java reaches it through reflection ("HACK No
// accessor"); the writer has an accessor.
func trueTypeUtilExtractTables(font *writer.BaseFont) (map[string][2]int, error) {
	tables := font.GetTables()
	if tables == nil {
		return nil, fmt.Errorf("Could not find tables field")
	}
	return tables, nil
}

// trueTypeUtilGetTTCName is not called in the port: Java uses it to open the
// collection file again, and the port reads the bytes the font already holds.
func trueTypeUtilGetTTCName(name string) string {
	index := strings.Index(strings.ToLower(name), ".ttc,")
	if index < 0 {
		return name
	}
	return name[0 : index+4]
}

// TrueTypeUtilExtractDescription ports extractDescription(String, BaseFont,
// IdentValue). fontWeightOverride may be nil.
func TrueTypeUtilExtractDescription(path string, font *writer.BaseFont, fontWeightOverride *ufo.IdentValue) *FontDescription {
	decorations, err := trueTypeUtilReadFontDecorations(path, font, fontWeightOverride)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(fmt.Sprintf("Failed to read font description from %s", path), err))
	}
	return NewFontDescriptionWithIsFromFontFaceStyleDecorations(font, false, trueTypeUtilGuessStyle(font), decorations)
}

// TrueTypeUtilExtractDescriptionWithContents ports extractDescription(String,
// byte[], BaseFont, boolean, IdentValue, IdentValue). fontWeightOverride and
// fontStyleOverride may be nil. contents is not read: the font holds the
// same bytes.
func TrueTypeUtilExtractDescriptionWithContents(path string, contents []byte,
	font *writer.BaseFont, isFromFontFace bool,
	fontWeightOverride *ufo.IdentValue,
	fontStyleOverride *ufo.IdentValue) *FontDescription {
	style := fontStyleOverride
	if style == nil {
		style = trueTypeUtilGuessStyle(font)
	}
	decorations, err := trueTypeUtilReadFontDecorations(path, font, fontWeightOverride)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(fmt.Sprintf("Failed to read font description from %s", path), err))
	}
	return NewFontDescriptionWithIsFromFontFaceStyleDecorations(font, isFromFontFace, style, decorations)
}

// trueTypeUtilReader reads big-endian values at a position, the part of
// RandomAccessFileOrArray that readFontDecorations uses.
type trueTypeUtilReader struct {
	data []byte
	pos  int
}

func (r *trueTypeUtilReader) seek(pos int) {
	r.pos = pos
}

// skip returns the number of bytes skipped, which is less than n at the end
// of the data.
func (r *trueTypeUtilReader) skip(n int) int {
	remaining := len(r.data) - r.pos
	if remaining < 0 {
		remaining = 0
	}
	if n > remaining {
		n = remaining
	}
	r.pos += n
	return n
}

func (r *trueTypeUtilReader) readUnsignedShort() (int, error) {
	if r.pos < 0 || r.pos+2 > len(r.data) {
		return 0, fmt.Errorf("unexpected end of font data")
	}
	v := int(r.data[r.pos])<<8 | int(r.data[r.pos+1])
	r.pos += 2
	return v, nil
}

func (r *trueTypeUtilReader) readShort() (int, error) {
	v, err := r.readUnsignedShort()
	if err != nil {
		return 0, err
	}
	return int(int16(uint16(v))), nil
}

func trueTypeUtilReadFontDecorations(path string, font *writer.BaseFont, fontWeightOverride *ufo.IdentValue) (*FontDescriptionDecorations, error) {
	tables, err := trueTypeUtilExtractTables(font)
	if err != nil {
		return nil, err
	}
	rf := &trueTypeUtilReader{data: font.GetFontData()}

	location, ok := tables["OS/2"]
	if !ok {
		return nil, fmt.Errorf("Table 'OS/2' does not exist in %s", path)
	}

	rf.seek(location[0])
	want := 4
	got := rf.skip(want)
	if got < want {
		return nil, fmt.Errorf("Skip TT font weight, expect read %d bytes, but only got %d", want, got)
	}

	fontWeight, err := rf.readUnsignedShort()
	if err != nil {
		return nil, err
	}
	weight := fontWeight
	if fontWeightOverride != nil {
		weight = ITextFontResolverConvertWeightToInt(fontWeightOverride)
	}

	want = 20
	got = rf.skip(want)
	if got < want {
		return nil, fmt.Errorf("Skip TT font strikeout, expect read %d bytes, but only got %d", want, got)
	}

	yStrikeoutSize, err := rf.readShort()
	if err != nil {
		return nil, err
	}
	yStrikeoutPosition, err := rf.readShort()
	if err != nil {
		return nil, err
	}
	underlinePosition := 0
	underlineThickness := 0

	location, ok = tables["post"]

	if ok {
		rf.seek(location[0])
		want = 8
		got = rf.skip(want)
		if got < want {
			return nil, fmt.Errorf("Skip TT font underline, expect read %d bytes, but only got %d", want, got)
		}
		if underlinePosition, err = rf.readShort(); err != nil {
			return nil, err
		}
		if underlineThickness, err = rf.readShort(); err != nil {
			return nil, err
		}
	}

	return NewFontDescriptionDecorations(weight, float32(yStrikeoutSize), float32(yStrikeoutPosition),
		float32(underlinePosition), float32(underlineThickness)), nil
}
