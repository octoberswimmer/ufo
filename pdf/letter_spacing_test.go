// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/LetterSpacingTest.java

package pdf

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The letter-spacing property should both widen the layout (line breaking)
// and space out the glyphs actually painted to the page.
//
// See https://github.com/flyingsaucerproject/flyingsaucer/issues/356.

func TestLetterSpacing_positiveLetterSpacingSpacesOutGlyphs(t *testing.T) {
	spaced := testUtilsFromClasspathResource(t, "letter-spacing.html")
	// 0.25em of extra spacing is 250/1000 of the em square, drawn as a
	// TJ kerning array: [(W) -250 (i) -250 (d) -250 (e)] TJ
	letterSpacingTestAssertCloseTo(t, letterSpacingTestKernBetween(t, testUtilsPageContent(t, spaced), 'W', 'i'), -250.0, 0.5)
}

func TestLetterSpacing_negativeLetterSpacingTightensGlyphs(t *testing.T) {
	spaced := testUtilsFromClasspathResource(t, "letter-spacing.html")
	// -1px at font-size 12px is -83.3/1000 of the em square; positive
	// TJ adjustments move the following glyph closer
	letterSpacingTestAssertCloseTo(t, letterSpacingTestKernBetween(t, testUtilsPageContent(t, spaced), 'T', 'i'), 250.0/3, 0.5)
}

func TestLetterSpacing_normalLetterSpacingDrawsPlainStrings(t *testing.T) {
	spaced := testUtilsFromClasspathResource(t, "letter-spacing.html")
	testUtilsPrintFile(t, spaced, "letter-spacing.pdf").containsText(t, "Hello")
	if content := testUtilsPageContent(t, spaced); !strings.Contains(content, "(Hello)Tj") {
		t.Errorf("page content does not contain (Hello)Tj:\n%s", content)
	}
}

func TestLetterSpacing_letterSpacingIsIncludedInLineBreaking(t *testing.T) {
	wrapped := testUtilsFromClasspathResource(t, "letter-spacing-wrap.html")
	content := testUtilsPageContent(t, wrapped)

	// the plain paragraph keeps "aaa aaa" on one line, drawn as a single string
	if !strings.Contains(content, "(aaa aaa)Tj") {
		t.Errorf("page content does not contain (aaa aaa)Tj:\n%s", content)
	}

	// the letter-spaced paragraph no longer fits 100px and wraps into two
	// kerned runs ("aaa" / "aaa"), so the page has three distinct baselines
	if n := len(strings.Split(content, "]TJ")); n != 3 {
		t.Errorf("content split at ]TJ has %d parts, want 3:\n%s", n, content)
	}
	if n := len(letterSpacingTestBaselines(content)); n != 3 {
		t.Errorf("%d baselines, want 3:\n%s", n, content)
	}
}

func letterSpacingTestAssertCloseTo(t *testing.T, actual float64, expected float64, within float64) {
	t.Helper()
	if math.Abs(actual-expected) > within {
		t.Errorf("%v is not within %v of %v", actual, within, expected)
	}
}

func letterSpacingTestKernBetween(t *testing.T, content string, first rune, second rune) float64 {
	t.Helper()
	matcher := regexp.MustCompile(`\(` + regexp.QuoteMeta(string(first)) + `\)\s*(-?[0-9.]+)\s*\(` + regexp.QuoteMeta(string(second)) + `\)`)
	match := matcher.FindStringSubmatch(content)
	if match == nil {
		t.Fatalf("expected a TJ kerning adjustment between (%c) and (%c) in:\n%s", first, second, content)
	}
	f, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func letterSpacingTestBaselines(content string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, match := range regexp.MustCompile(`1 0 0 1 [0-9.]+ ([0-9.]+) Tm`).FindAllStringSubmatch(content, -1) {
		result[match[1]] = struct{}{}
	}
	return result
}
