// Package fonts holds the fonts that ship inside ufo, for text that the 14
// standard PDF fonts cannot show: those are limited to the characters of
// Windows-1252. A bundled font is given to a FontDescription as its fallback
// font (pdf.FontDescription.SetFallbackFont). NOTICE.md in this directory
// gives the origin and the license of each font.
package fonts

import (
	_ "embed"
	"sync"

	"github.com/octoberswimmer/ufo/pdf/writer"
)

// WenQuanYiMicroHei is the name of the bundled sans-serif font with glyphs
// for Latin, Greek, Cyrillic, Japanese, Chinese and Korean text.
const WenQuanYiMicroHei = "WenQuanYi Micro Hei"

//go:embed wqy-microhei.ttc
var wqyMicroHei []byte

// Font is a bundled font. It is parsed when it is first asked for, so a
// document that does not use it does not pay for it.
type Font struct {
	name string
	// file is the name the font file is created under, which tells the
	// writer its kind and the member of a collection.
	file string
	data []byte

	once sync.Once
	font *writer.BaseFont
	err  error
}

// Name returns the name the font is looked up by.
func (f *Font) Name() string { return f.name }

// BaseFont returns the font, written to a PDF as an embedded subset with
// Identity-H encoding.
func (f *Font) BaseFont() (*writer.BaseFont, error) {
	f.once.Do(func() {
		f.font, f.err = writer.CreateFontWithBytes(f.file, writer.BaseFontIdentityH, true, false, f.data, nil)
	})
	return f.font, f.err
}

var bundled = map[string]*Font{
	WenQuanYiMicroHei: {name: WenQuanYiMicroHei, file: "wqy-microhei.ttc,0", data: wqyMicroHei},
}

// Lookup returns the bundled font of that name. Names are matched exactly.
func Lookup(name string) (*Font, bool) {
	f, ok := bundled[name]
	return f, ok
}

// Names returns the names of the bundled fonts.
func Names() []string {
	return []string{WenQuanYiMicroHei}
}
