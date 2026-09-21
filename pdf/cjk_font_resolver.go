// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/CJKFontResolver.java

package pdf

import (
	"fmt"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// CJKFontResolver is the class to use if you need to load iTextAsian fonts
// in addition to default fonts loaded by ITextRenderer.
//
// The iTextAsian fonts are fonts that a PDF viewer supplies, described by
// the CMap and metrics resources that ship with iText. The writer has
// neither, so none of the families can be created: constructing the resolver
// succeeds, LoadFonts returns the default families only and logs one error
// per CJK family, as Java does when the iTextAsian resources are missing,
// and a request for one of the families falls back like any unknown family.
// The errors, each a *writer.UnsupportedFeatureError naming the font and its
// CMap encoding, are available from GetCJKFontErrors.
type CJKFontResolver struct {
	ITextFontResolver

	cjkFontErrors []error
}

var _ ITextFontResolverI = (*CJKFontResolver)(nil)

func NewCJKFontResolver() *CJKFontResolver {
	r := &CJKFontResolver{}
	initITextFontResolver(&r.ITextFontResolver)
	r.SetSelf(r)
	return r
}

func (r *CJKFontResolver) LoadFonts() map[string]*FontFamily {
	result := r.ITextFontResolver.LoadFonts()
	for name, family := range r.loadCJKFonts() {
		result[name] = family
	}
	return result
}

// fontFamilyName, fontName, encoding
var cjkFontResolverCjkFonts = [][]string{
	{"STSong-Light-H", "STSong-Light", "UniGB-UCS2-H"},
	{"STSong-Light-V", "STSong-Light", "UniGB-UCS2-V"},
	{"STSongStd-Light-H", "STSongStd-Light", "UniGB-UCS2-H"},
	{"STSongStd-Light-V", "STSongStd-Light", "UniGB-UCS2-V"},
	{"MHei-Medium-H", "MHei-Medium", "UniCNS-UCS2-H"},
	{"MHei-Medium-V", "MHei-Medium", "UniCNS-UCS2-V"},
	{"MSung-Light-H", "MSung-Light", "UniCNS-UCS2-H"},
	{"MSung-Light-V", "MSung-Light", "UniCNS-UCS2-V"},
	{"MSungStd-Light-H", "MSungStd-Light", "UniCNS-UCS2-H"},
	{"MSungStd-Light-V", "MSungStd-Light", "UniCNS-UCS2-V"},
	{"HeiseiMin-W3-H", "HeiseiMin-W3", "UniJIS-UCS2-H"},
	{"HeiseiMin-W3-V", "HeiseiMin-W3", "UniJIS-UCS2-V"},
	{"HeiseiKakuGo-W5-H", "HeiseiKakuGo-W5", "UniJIS-UCS2-H"},
	{"HeiseiKakuGo-W5-V", "HeiseiKakuGo-W5", "UniJIS-UCS2-V"},
	{"KozMinPro-Regular-H", "KozMinPro-Regular", "UniJIS-UCS2-HW-H"},
	{"KozMinPro-Regular-V", "KozMinPro-Regular", "UniJIS-UCS2-HW-V"},
	{"HYGoThic-Medium-H", "HYGoThic-Medium", "UniKS-UCS2-H"},
	{"HYGoThic-Medium-V", "HYGoThic-Medium", "UniKS-UCS2-V"},
	{"HYSMyeongJo-Medium-H", "HYSMyeongJo-Medium", "UniKS-UCS2-H"},
	{"HYSMyeongJo-Medium-V", "HYSMyeongJo-Medium", "UniKS-UCS2-V"},
	{"HYSMyeongJoStd-Medium-H", "HYSMyeongJoStd-Medium", "UniKS-UCS2-H"},
	{"HYSMyeongJoStd-Medium-V", "HYSMyeongJoStd-Medium", "UniKS-UCS2-V"},
}

// loadCJKFonts tries to load the iTextAsian fonts.
func (r *CJKFontResolver) loadCJKFonts() map[string]*FontFamily {
	fontFamilyMap := map[string]*FontFamily{}
	r.cjkFontErrors = nil

	for _, cjkFont := range cjkFontResolverCjkFonts {
		fontFamilyName := cjkFont[0]
		fontName := cjkFont[1]
		encoding := cjkFont[2]

		fontFamily, err := r.addCJKFont(fontFamilyName, fontName, encoding)
		if err != nil {
			r.cjkFontErrors = append(r.cjkFontErrors, err)
			ufo.XRLogExceptionWithTh(fmt.Sprintf("Failed to load font %s %s %s", fontFamilyName, fontName, encoding), err)
			continue
		}
		fontFamilyMap[fontFamilyName] = fontFamily
	}
	return fontFamilyMap
}

// GetCJKFontErrors returns the errors of the last LoadFonts call, one per
// CJK family that could not be created. It is not part of the Java class,
// which only logs them.
func (r *CJKFontResolver) GetCJKFontErrors() []error {
	return r.cjkFontErrors
}

func (r *CJKFontResolver) addCJKFont(fontFamilyName string, fontName string, encoding string) (*FontFamily, error) {
	fontFamily := NewFontFamily(fontFamilyName)

	boldItalic, err := cjkFontResolverCreateFont(fontName+",BoldItalic", encoding)
	if err != nil {
		return nil, err
	}
	fontFamily.AddFontDescription(NewFontDescriptionWithStyleWeight(boldItalic, ufo.IdentValueOblique, 700))
	italic, err := cjkFontResolverCreateFont(fontName+",Italic", encoding)
	if err != nil {
		return nil, err
	}
	fontFamily.AddFontDescription(NewFontDescriptionWithStyleWeight(italic, ufo.IdentValueOblique, 400))
	bold, err := cjkFontResolverCreateFont(fontName+",Bold", encoding)
	if err != nil {
		return nil, err
	}
	fontFamily.AddFontDescription(NewFontDescriptionWithStyleWeight(bold, ufo.IdentValueNormal, 700))
	regular, err := cjkFontResolverCreateFont(fontName, encoding)
	if err != nil {
		return nil, err
	}
	fontFamily.AddFontDescription(NewFontDescriptionWithStyleWeight(regular, ufo.IdentValueNormal, 400))

	return fontFamily, nil
}

// cjkFontResolverCreateFont is BaseFont.createFont(name, encoding, false) for
// a built-in CJK font. The writer does not recognize these names; its error
// is replaced by one that names the missing feature.
func cjkFontResolverCreateFont(name string, encoding string) (*writer.BaseFont, error) {
	font, err := writer.CreateFont(name, encoding, false)
	if err != nil {
		return nil, &writer.UnsupportedFeatureError{
			Feature: fmt.Sprintf("built-in CJK font %s with CMap encoding %s (iTextAsian CMap and metrics resources)", name, encoding),
		}
	}
	return font, nil
}
