// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/ITextFontResolverTest.java

package pdf

import "testing"

func iTextFontResolverTestAssert(t *testing.T, input string, expected string) {
	t.Helper()
	resolver := NewITextFontResolver()
	if actual := resolver.NormalizeFontFamily(input); actual != expected {
		t.Errorf("NormalizeFontFamily(%q) = %q, want %q", input, actual, expected)
	}
}

func TestITextFontResolver_normalizeFontFamily(t *testing.T) {
	iTextFontResolverTestAssert(t, "ArialUnicodeMS", "ArialUnicodeMS")
	iTextFontResolverTestAssert(t, "\"ArialUnicodeMS", "ArialUnicodeMS")
	iTextFontResolverTestAssert(t, "ArialUnicodeMS\"", "ArialUnicodeMS")
	iTextFontResolverTestAssert(t, "\"ArialUnicodeMS\"", "ArialUnicodeMS")
}

func TestITextFontResolver_normalizeFontFamily_serif(t *testing.T) {
	iTextFontResolverTestAssert(t, "serif", "Serif")
	iTextFontResolverTestAssert(t, "SERIF", "Serif")
	iTextFontResolverTestAssert(t, "sErIf", "Serif")
}

func TestITextFontResolver_normalizeFontFamily_sans_serif(t *testing.T) {
	iTextFontResolverTestAssert(t, "sans-serif", "SansSerif")
	iTextFontResolverTestAssert(t, "SANS-serif", "SansSerif")
	iTextFontResolverTestAssert(t, "sans-SERIF", "SansSerif")
	iTextFontResolverTestAssert(t, "\"sans-serif", "SansSerif")
	iTextFontResolverTestAssert(t, "sans-serif\"", "SansSerif")
	iTextFontResolverTestAssert(t, "\"sans-serif\"", "SansSerif")
}

func TestITextFontResolver_normalizeFontFamily_monospace(t *testing.T) {
	iTextFontResolverTestAssert(t, "monospace", "Monospaced")
	iTextFontResolverTestAssert(t, "MONOSPACE", "Monospaced")
	iTextFontResolverTestAssert(t, "\"monospace\"", "Monospaced")
}
