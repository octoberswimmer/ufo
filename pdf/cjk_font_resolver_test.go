// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/CJKFontResolverTest.java

package pdf

import (
	"errors"
	"testing"

	"github.com/octoberswimmer/ufo/pdf/writer"
)

// The CJK families are the built-in CMap fonts of iTextAsian, which the port
// does not have (CHECKLIST.md, "Not ported"). Java asserts that
// HYGoThic-Medium-V, HYSMyeongJoStd-Medium-H and MSung-Light-H are loaded
// with four descriptions each; the port asserts that none of the 22 CJK
// families is loaded and that each one is reported as an unsupported feature.
func TestCJKFontResolver_loadsChinaJapanKoreanFonts(t *testing.T) {
	resolver := NewCJKFontResolver()
	cjkFonts := resolver.LoadFonts()
	if n := len(cjkFonts); n < 10 || n > 1000 {
		t.Errorf("%d families, want between 10 and 1000", n)
	}
	for _, name := range []string{"HYGoThic-Medium-V", "HYSMyeongJoStd-Medium-H", "MSung-Light-H"} {
		if _, ok := cjkFonts[name]; ok {
			t.Errorf("family %s was loaded; the port has no CJK CMap fonts", name)
		}
	}
	errs := resolver.GetCJKFontErrors()
	if len(errs) != len(cjkFontResolverCjkFonts) {
		t.Fatalf("%d CJK font errors, want %d", len(errs), len(cjkFontResolverCjkFonts))
	}
	for _, err := range errs {
		var unsupported *writer.UnsupportedFeatureError
		if !errors.As(err, &unsupported) {
			t.Errorf("error %v is not a *writer.UnsupportedFeatureError", err)
		}
	}
}

func TestCJKFontResolver_loadsDefaultFontsAsWell(t *testing.T) {
	resolver := NewCJKFontResolver()
	cjkFonts := resolver.LoadFonts()
	if n := len(cjkFonts); n < 10 || n > 1000 {
		t.Errorf("%d families, want between 10 and 1000", n)
	}
	for _, name := range []string{"Courier", "Helvetica", "Serif", "TimesRoman"} {
		if _, ok := cjkFonts[name]; !ok {
			t.Errorf("family %s is missing", name)
		}
	}
	courier := cjkFonts["Courier"]
	if courier == nil {
		t.FailNow()
	}
	if courier.GetName() != "Courier" {
		t.Errorf("GetName() = %q", courier.GetName())
	}
	if n := len(courier.GetFontDescriptions()); n != 4 {
		t.Fatalf("%d font descriptions, want 4", n)
	}
	if s := courier.GetFontDescriptions()[0].ToString(); s != "Font Courier-Oblique:400" {
		t.Errorf("first description %q, want %q", s, "Font Courier-Oblique:400")
	}
}
