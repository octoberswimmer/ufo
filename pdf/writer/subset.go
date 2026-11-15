package writer

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
)

// The subset of an embedded TrueType font is built as OpenPDF's
// TrueTypeFontSubSet builds it for a font written with Identity-H encoding:
// the font file keeps the tables a reader needs to draw glyphs by glyph id
// (subsetTables), and of the glyf table only the outlines of the glyphs
// shown, of glyph 0, and of the glyphs those are composed of. The loca table
// keeps one entry per glyph of the original font, so glyph ids do not change
// and the hmtx table is copied whole.

// subsetTables are the tables of a subset, in the order of their directory
// entries.
var subsetTables = []string{"cvt ", "fpgm", "glyf", "head", "hhea", "hmtx", "loca", "maxp", "prep"}

// Flags of a component of a composite glyph.
const (
	glyfArg1And2AreWords   = 1
	glyfWeHaveAScale       = 8
	glyfMoreComponents     = 32
	glyfWeHaveAnXAndYScale = 64
	glyfWeHaveATwoByTwo    = 128
)

// subset returns the font file reduced to the glyphs in used.
func (tt *ttFont) subset(used map[uint16]rune) ([]byte, error) {
	head := tt.table("head")
	if len(head) < 54 {
		return nil, newDocumentException("Table 'head' does not exist in %s", tt.fontName)
	}
	glyf := tt.table("glyf")
	if glyf == nil {
		return nil, newDocumentException("Table 'glyf' does not exist in %s", tt.fontName)
	}
	locaData := tt.table("loca")
	if locaData == nil {
		return nil, newDocumentException("Table 'loca' does not exist in %s", tt.fontName)
	}
	shortLoca := u16(head, 50) == 0
	var loca []int
	if shortLoca {
		loca = make([]int, len(locaData)/2)
		for i := range loca {
			loca[i] = u16(locaData, 2*i) * 2
		}
	} else {
		loca = make([]int, len(locaData)/4)
		for i := range loca {
			loca[i] = u32(locaData, 4*i)
		}
	}
	for i, offset := range loca {
		if offset > len(glyf) || (i > 0 && offset < loca[i-1]) {
			return nil, newDocumentException("Table 'loca' of %s is not valid", tt.fontName)
		}
	}

	// The glyphs kept: glyph 0, the glyphs shown, and the components of
	// composite glyphs among them, found transitively.
	kept := map[int]bool{0: true}
	pending := []int{0}
	for g := range used {
		if !kept[int(g)] {
			kept[int(g)] = true
			pending = append(pending, int(g))
		}
	}
	for len(pending) > 0 {
		g := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if g >= len(loca)-1 {
			continue
		}
		outline := glyf[loca[g]:loca[g+1]]
		if len(outline) < 10 || s16(outline, 0) >= 0 {
			continue
		}
		for pos := 10; pos+4 <= len(outline); {
			flags := u16(outline, pos)
			component := u16(outline, pos+2)
			if !kept[component] {
				kept[component] = true
				pending = append(pending, component)
			}
			if flags&glyfMoreComponents == 0 {
				break
			}
			pos += 4
			if flags&glyfArg1And2AreWords != 0 {
				pos += 4
			} else {
				pos += 2
			}
			switch {
			case flags&glyfWeHaveAScale != 0:
				pos += 2
			case flags&glyfWeHaveAnXAndYScale != 0:
				pos += 4
			case flags&glyfWeHaveATwoByTwo != 0:
				pos += 8
			}
		}
	}
	glyphs := make([]int, 0, len(kept))
	for g := range kept {
		if g < len(loca)-1 {
			glyphs = append(glyphs, g)
		}
	}
	sort.Ints(glyphs)

	newLoca := make([]int, len(loca))
	var newGlyf []byte
	next := 0
	for g := range newLoca {
		newLoca[g] = len(newGlyf)
		if next < len(glyphs) && glyphs[next] == g {
			next++
			newGlyf = append(newGlyf, glyf[loca[g]:loca[g+1]]...)
		}
	}
	glyfLength := len(newGlyf)
	newGlyf = padTable(newGlyf)

	var newLocaData []byte
	if shortLoca {
		newLocaData = make([]byte, len(newLoca)*2)
		for i, offset := range newLoca {
			binary.BigEndian.PutUint16(newLocaData[2*i:], uint16(offset/2))
		}
	} else {
		newLocaData = make([]byte, len(newLoca)*4)
		for i, offset := range newLoca {
			binary.BigEndian.PutUint32(newLocaData[4*i:], uint32(offset))
		}
	}
	locaLength := len(newLocaData)
	newLocaData = padTable(newLocaData)

	type entry struct {
		tag      string
		checksum uint32
		length   int
		data     []byte
	}
	var entries []entry
	for _, tag := range subsetTables {
		location, ok := tt.tables[tag]
		if !ok {
			continue
		}
		switch tag {
		case "glyf":
			entries = append(entries, entry{tag, tableChecksum(newGlyf), glyfLength, newGlyf})
		case "loca":
			entries = append(entries, entry{tag, tableChecksum(newLocaData), locaLength, newLocaData})
		default:
			data := padTable(tt.data[location[0] : location[0]+location[1]])
			entries = append(entries, entry{tag, tableChecksum(data), location[1], data})
		}
	}

	numTables := len(entries)
	selector := 0
	for 1<<(selector+1) <= numTables {
		selector++
	}
	size := 12 + 16*numTables
	for _, e := range entries {
		size += len(e.data)
	}
	out := make([]byte, 0, size)
	out = binary.BigEndian.AppendUint32(out, 0x00010000)
	out = binary.BigEndian.AppendUint16(out, uint16(numTables))
	out = binary.BigEndian.AppendUint16(out, uint16((1<<selector)*16))
	out = binary.BigEndian.AppendUint16(out, uint16(selector))
	out = binary.BigEndian.AppendUint16(out, uint16((numTables-(1<<selector))*16))
	offset := 12 + 16*numTables
	for _, e := range entries {
		out = append(out, e.tag...)
		out = binary.BigEndian.AppendUint32(out, e.checksum)
		out = binary.BigEndian.AppendUint32(out, uint32(offset))
		out = binary.BigEndian.AppendUint32(out, uint32(e.length))
		offset += len(e.data)
	}
	for _, e := range entries {
		out = append(out, e.data...)
	}
	return out, nil
}

// padTable returns the table extended with zero bytes to a multiple of four
// bytes, in a new slice when it has to grow.
func padTable(data []byte) []byte {
	if len(data)%4 == 0 {
		return data
	}
	padded := make([]byte, (len(data)+3)&^3)
	copy(padded, data)
	return padded
}

// tableChecksum is the sum of the 32-bit words of a table whose length is a
// multiple of four bytes.
func tableChecksum(data []byte) uint32 {
	var sum uint32
	for i := 0; i+4 <= len(data); i += 4 {
		sum += binary.BigEndian.Uint32(data[i:])
	}
	return sum
}

// subsetTag returns the six uppercase letters that prefix the name of a font
// subset. OpenPDF picks them at random; here they are derived from the glyph
// ids of the subset, so the same document is written as the same bytes.
func subsetTag(glyphs []int) string {
	h := sha256.New()
	var b [2]byte
	for _, g := range glyphs {
		binary.BigEndian.PutUint16(b[:], uint16(g))
		h.Write(b[:])
	}
	sum := h.Sum(nil)
	tag := make([]byte, 6)
	for i := range tag {
		tag[i] = 'A' + sum[i]%26
	}
	return string(tag)
}
