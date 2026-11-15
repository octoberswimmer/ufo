package writer

import (
	"bytes"
	"os"
	"testing"
)

const subsetTestFont = "../testdata/fonts/Jacquard24-Regular.ttf"

// subsetOf returns the subset of the font holding the glyphs of text, and
// the glyph ids of those characters.
func subsetOf(t *testing.T, bf *BaseFont, text string) (*ttFont, []byte, []uint16) {
	t.Helper()
	used := map[uint16]rune{}
	var glyphs []uint16
	for _, r := range text {
		g := bf.tt.glyphIndex(r)
		if g == 0 {
			t.Fatalf("the font has no glyph for %q", r)
		}
		used[g] = r
		glyphs = append(glyphs, g)
	}
	data, err := bf.tt.subset(used)
	if err != nil {
		t.Fatal(err)
	}
	sub := &ttFont{data: data}
	if err := sub.readDirectory(); err != nil {
		t.Fatalf("the subset is not a font file: %v", err)
	}
	return sub, data, glyphs
}

// glyphOutline returns the outline of a glyph of a font with a long or short
// loca table.
func glyphOutline(tt *ttFont, g int) []byte {
	loca, glyf := tt.table("loca"), tt.table("glyf")
	if u16(tt.table("head"), 50) == 0 {
		return glyf[u16(loca, 2*g)*2 : u16(loca, 2*g+2)*2]
	}
	return glyf[u32(loca, 4*g):u32(loca, 4*g+4)]
}

func TestSubsetKeepsOnlyTheGlyphsShown(t *testing.T) {
	bf, err := CreateFont(subsetTestFont, BaseFontIdentityH, true)
	if err != nil {
		t.Fatal(err)
	}
	sub, data, glyphs := subsetOf(t, bf, "ab")
	if len(data) >= len(bf.tt.data) {
		t.Errorf("the subset has %d bytes, the font %d", len(data), len(bf.tt.data))
	}
	for tag := range sub.tables {
		found := false
		for _, kept := range subsetTables {
			found = found || kept == tag
		}
		if !found {
			t.Errorf("the subset holds the table %q", tag)
		}
	}
	for _, tag := range []string{"glyf", "head", "hhea", "hmtx", "loca", "maxp"} {
		if sub.table(tag) == nil {
			t.Errorf("the subset lacks the table %q", tag)
		}
	}
	if !bytes.Equal(sub.table("hmtx"), bf.tt.table("hmtx")) {
		t.Errorf("the hmtx table changed")
	}
	if len(sub.table("loca")) != len(bf.tt.table("loca")) {
		t.Errorf("the loca table has %d bytes, want %d: glyph ids must not change", len(sub.table("loca")), len(bf.tt.table("loca")))
	}
	for _, g := range append([]uint16{0}, glyphs...) {
		if want := glyphOutline(bf.tt, int(g)); !bytes.Equal(glyphOutline(sub, int(g)), want) {
			t.Errorf("glyph %d: the outline changed", g)
		}
	}
	other := int(bf.tt.glyphIndex('z'))
	if len(glyphOutline(bf.tt, other)) == 0 {
		t.Fatalf("glyph %d has no outline in the font", other)
	}
	if n := len(glyphOutline(sub, other)); n != 0 {
		t.Errorf("glyph %d, which is not shown, keeps %d bytes of outline", other, n)
	}

	// The directory gives each table its checksum.
	for tag, location := range sub.tables {
		padded := padTable(sub.data[location[0] : location[0]+location[1]])
		for i := 0; i < int(u16(sub.data, 4)); i++ {
			record := sub.data[12+16*i:]
			if string(record[:4]) == tag && uint32(u32(record, 4)) != tableChecksum(padded) {
				t.Errorf("table %q: checksum %08x, want %08x", tag, u32(record, 4), tableChecksum(padded))
			}
		}
	}
}

func TestSubsetIsTheSameForTheSameGlyphs(t *testing.T) {
	bf, err := CreateFont(subsetTestFont, BaseFontIdentityH, true)
	if err != nil {
		t.Fatal(err)
	}
	_, first, _ := subsetOf(t, bf, "abc")
	_, second, _ := subsetOf(t, bf, "cba")
	if !bytes.Equal(first, second) {
		t.Errorf("two subsets of the same glyphs differ")
	}
	_, third, _ := subsetOf(t, bf, "abd")
	if bytes.Equal(first, third) {
		t.Errorf("subsets of different glyphs are equal")
	}
	if a, b := subsetTag([]int{1, 2, 3}), subsetTag([]int{1, 2, 4}); a == b || len(a) != 6 {
		t.Errorf("subsetTag = %q and %q", a, b)
	}
}

// A composite glyph is drawn from other glyphs, which the subset keeps.
func TestSubsetKeepsTheComponentsOfCompositeGlyphs(t *testing.T) {
	if _, err := os.Stat(testTrueTypeFont); err != nil {
		t.Skipf("TrueType font %s is not available", testTrueTypeFont)
	}
	bf, err := CreateFont(testTrueTypeFont, BaseFontIdentityH, true)
	if err != nil {
		t.Fatal(err)
	}
	sub, _, glyphs := subsetOf(t, bf, "é")
	outline := glyphOutline(bf.tt, int(glyphs[0]))
	if s16(outline, 0) >= 0 {
		t.Fatalf("é is not a composite glyph in %s", testTrueTypeFont)
	}
	for _, r := range "e" {
		g := int(bf.tt.glyphIndex(r))
		if want := glyphOutline(bf.tt, g); len(want) == 0 || !bytes.Equal(glyphOutline(sub, g), want) {
			t.Errorf("the component glyph %d (%q) is not in the subset", g, r)
		}
	}
	if n := len(glyphOutline(sub, int(bf.tt.glyphIndex('z')))); n != 0 {
		t.Errorf("z, which is not shown, keeps %d bytes of outline", n)
	}
}
