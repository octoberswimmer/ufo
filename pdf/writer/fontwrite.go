package writer

import (
	"bytes"
	"fmt"
	"sort"
)

// writeFont writes the font dictionary of a font and the objects it needs.
func (w *PdfWriter) writeFont(e *fontEntry) {
	if e.font.tt == nil {
		w.writeType1Font(e)
		return
	}
	w.writeType0Font(e)
}

// writeType1Font writes a standard font: no widths and no descriptor, which
// a reader supplies itself for the 14 standard fonts. Symbol and
// ZapfDingbats keep their built-in encoding.
func (w *PdfWriter) writeType1Font(e *fontEntry) {
	d := NewPdfDictionaryWithType("Font")
	d.Put("Subtype", PdfName("Type1"))
	d.Put("BaseFont", PdfName(e.font.afm.fontName))
	if !e.font.afm.fontSpecific {
		d.Put("Encoding", PdfName("WinAnsiEncoding"))
	}
	w.writeObject(e.ref, d)
}

// writeType0Font writes a TrueType font as a Type 0 font with Identity-H
// encoding over a CIDFontType2 descendant: the whole font file as FontFile2,
// a /W array and a ToUnicode CMap for the glyphs that were shown.
func (w *PdfWriter) writeType0Font(e *fontEntry) {
	tt := e.font.tt
	upem := tt.unitsPerEm

	fileStream := w.newStream(tt.data)
	fileStream.Put("Length1", PdfNumber(len(tt.data)))
	fileRef := w.newReference()
	w.writeObject(fileRef, fileStream)

	desc := NewPdfDictionaryWithType("FontDescriptor")
	desc.Put("FontName", PdfName(tt.fontName))
	flags := 0
	if tt.isFixedPitch {
		flags |= 1
	}
	if tt.symbolic {
		flags |= 4
	} else {
		flags |= 32
	}
	if tt.macStyle&2 != 0 {
		flags |= 64
	}
	if tt.macStyle&1 != 0 {
		flags |= 262144
	}
	desc.Put("Flags", PdfNumber(flags))
	desc.Put("FontBBox", NewPdfArray(
		PdfNumber(tt.xMin*1000/upem), PdfNumber(tt.yMin*1000/upem),
		PdfNumber(tt.xMax*1000/upem), PdfNumber(tt.yMax*1000/upem)))
	desc.Put("ItalicAngle", PdfNumber(tt.italicAngle))
	desc.Put("Ascent", PdfNumber(tt.sTypoAscender*1000/upem))
	desc.Put("Descent", PdfNumber(tt.sTypoDescender*1000/upem))
	desc.Put("CapHeight", PdfNumber(tt.sCapHeight*1000/upem))
	desc.Put("StemV", PdfNumber(80))
	desc.Put("FontFile2", fileRef)
	descRef := w.newReference()
	w.writeObject(descRef, desc)

	glyphs := make([]int, 0, len(e.used))
	for g := range e.used {
		glyphs = append(glyphs, int(g))
	}
	sort.Ints(glyphs)

	// /W as runs of consecutive glyph ids: first [w w w ...].
	widths := NewPdfArray()
	for i := 0; i < len(glyphs); {
		j := i
		run := NewPdfArray()
		for j < len(glyphs) && glyphs[j] == glyphs[i]+(j-i) {
			run.Add(PdfNumber(tt.glyphWidth(uint16(glyphs[j]))))
			j++
		}
		widths.Add(PdfNumber(glyphs[i]))
		widths.Add(run)
		i = j
	}

	cid := NewPdfDictionaryWithType("Font")
	cid.Put("Subtype", PdfName("CIDFontType2"))
	cid.Put("BaseFont", PdfName(tt.fontName))
	sysInfo := NewPdfDictionary()
	sysInfo.Put("Registry", NewPdfString("Adobe"))
	sysInfo.Put("Ordering", NewPdfString("Identity"))
	sysInfo.Put("Supplement", PdfNumber(0))
	cid.Put("CIDSystemInfo", sysInfo)
	cid.Put("FontDescriptor", descRef)
	cid.Put("DW", PdfNumber(1000))
	if widths.Size() > 0 {
		cid.Put("W", widths)
	}
	cid.Put("CIDToGIDMap", PdfName("Identity"))
	cidRef := w.newReference()
	w.writeObject(cidRef, cid)

	toUnicodeRef := w.newReference()
	w.writeObject(toUnicodeRef, w.newStream(toUnicodeCMap(glyphs, e.used)))

	d := NewPdfDictionaryWithType("Font")
	d.Put("Subtype", PdfName("Type0"))
	d.Put("BaseFont", PdfName(tt.fontName))
	d.Put("Encoding", PdfName("Identity-H"))
	d.Put("DescendantFonts", NewPdfArray(cidRef))
	d.Put("ToUnicode", toUnicodeRef)
	w.writeObject(e.ref, d)
}

// toUnicodeCMap returns a CMap that maps each two-byte glyph id to the
// character it was shown for, as UTF-16BE, in bfchar blocks of at most 100
// entries.
func toUnicodeCMap(glyphs []int, used map[uint16]rune) []byte {
	var b bytes.Buffer
	b.WriteString("/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n")
	b.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	b.WriteString("/CMapName /Adobe-Identity-UCS def\n/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")
	for i := 0; i < len(glyphs); i += 100 {
		end := i + 100
		if end > len(glyphs) {
			end = len(glyphs)
		}
		fmt.Fprintf(&b, "%d beginbfchar\n", end-i)
		for _, g := range glyphs[i:end] {
			r := used[uint16(g)]
			fmt.Fprintf(&b, "<%04X> <", g)
			if r >= 0x10000 {
				r -= 0x10000
				fmt.Fprintf(&b, "%04X%04X", 0xd800+(r>>10), 0xdc00+(r&0x3ff))
			} else {
				fmt.Fprintf(&b, "%04X", r)
			}
			b.WriteString(">\n")
		}
		b.WriteString("endbfchar\n")
	}
	b.WriteString("endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	return b.Bytes()
}
