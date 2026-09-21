// Tests Breaker (flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/Breaker.java)
// against results of the Java implementation.

package ufo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo/geom"
)

// breakerTestCharWidth is the width of every character for
// breakerTestTextRenderer, in both the Java harness and this test.
const breakerTestCharWidth = 10

// breakerTestTextRenderer is a text renderer for which every character (code
// point) is breakerTestCharWidth wide.
type breakerTestTextRenderer struct{}

func (breakerTestTextRenderer) Setup(context FontContext) {}
func (breakerTestTextRenderer) DrawString(outputDevice OutputDevice, str string, x float32, y float32) {
}
func (breakerTestTextRenderer) DrawStringWithInfo(outputDevice OutputDevice, str string, x float32, y float32, info *JustificationInfo) {
}
func (breakerTestTextRenderer) DrawGlyphVector(outputDevice OutputDevice, vector FSGlyphVector, x float32, y float32) {
}
func (breakerTestTextRenderer) GetGlyphVector(outputDevice OutputDevice, font FSFont, str string) FSGlyphVector {
	return nil
}
func (breakerTestTextRenderer) GetGlyphPositions(outputDevice OutputDevice, font FSFont, fsGlyphVector FSGlyphVector) []float32 {
	return nil
}
func (breakerTestTextRenderer) GetGlyphBounds(outputDevice OutputDevice, font FSFont, fsGlyphVector FSGlyphVector, index int, x float32, y float32) *geom.Rectangle {
	return nil
}
func (breakerTestTextRenderer) GetFSFontMetrics(context FontContext, font FSFont, str string) FSFontMetrics {
	return nil
}
func (breakerTestTextRenderer) GetWidth(context FontContext, font FSFont, str string) int {
	return breakerTestCharWidth * utf8.RuneCountInString(str)
}
func (breakerTestTextRenderer) SetFontScale(scale float32)             {}
func (breakerTestTextRenderer) GetFontScale() float32                  { return 1 }
func (breakerTestTextRenderer) SetSmoothingThreshold(fontsize float32) {}

type breakerTestFont struct{}

func (breakerTestFont) GetSize2D() float32 { return 12 }

type breakerTestFontResolver struct{}

func (breakerTestFontResolver) ResolveFont(renderingContext *SharedContext, spec *FontSpecification) FSFont {
	return nil
}
func (breakerTestFontResolver) FlushCache() {}

// breakerTestStyle answers the five style questions that Breaker and
// LayoutTextUtilTextWidth ask. Every other method of CalculatedStyleI panics
// on the nil embedded interface.
type breakerTestStyle struct {
	CalculatedStyleI
	whitespace    *IdentValue
	wordWrap      *IdentValue
	wordBreak     *IdentValue
	letterSpacing float32
}

func (s *breakerTestStyle) GetWhitespace() *IdentValue             { return s.whitespace }
func (s *breakerTestStyle) GetWordWrap() *IdentValue               { return s.wordWrap }
func (s *breakerTestStyle) GetWordBreak() *IdentValue              { return s.wordBreak }
func (s *breakerTestStyle) GetFSFont(cssContext CssContext) FSFont { return breakerTestFont{} }
func (s *breakerTestStyle) LetterSpacing(ctx CssContext) float32   { return s.letterSpacing }

// breakerTestHyphen3Strategy offers a break point with the hyphen "-" after
// every third byte, and one at the end of the text. The Java harness does the
// same with UTF-16 offsets, so it is used with ASCII texts only.
type breakerTestHyphen3Strategy struct{}

func (breakerTestHyphen3Strategy) GetBreakPointsProvider(text string, lang string, style CalculatedStyleI) BreakPointsProvider {
	var l []*BreakPoint
	for p := 3; p < len(text); p += 3 {
		bp := NewBreakPoint(p)
		bp.SetHyphen("-")
		l = append(l, bp)
	}
	l = append(l, NewBreakPoint(len(text)))
	return NewListBreakPointsProvider(l)
}

func breakerTestUnescape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\\' {
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(ch)
		}
	}
	return b.String()
}

func breakerTestEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return strings.ReplaceAll(s, "\t", "\\t")
}

// breakerTestRun breaks text into lines the way the Java harness does:
// BreakerBreakText is called for the rest of the text until the text is
// used up, and each call is recorded as
// start:end:width:flags, the offsets in bytes of the master text.
func breakerTestRun(sc *SharedContext, text string, avail int, fullLineWidth int, style CalculatedStyleI) (steps string, master string) {
	c := sc.NewLayoutContextInstance(nil)
	ctx := NewLineBreakContext(text, nil)
	var res []string
	guard := 0
	for {
		ctx.Reset()
		ctx.SetEndsOnNL(false)
		if ctx.GetStartSubstring() == "" {
			res = append(res, "empty")
			break
		}
		failed := func() (failed bool) {
			defer func() {
				if r := recover(); r != nil {
					failed = true
				}
			}()
			BreakerBreakText(c, ctx, avail, fullLineWidth, style)
			return false
		}()
		if failed {
			res = append(res, "ERR")
			break
		}
		flags := []byte("---")
		if ctx.IsNeedsNewLine() {
			flags[0] = 'N'
		}
		if ctx.IsUnbreakable() {
			flags[1] = 'U'
		}
		if ctx.IsEndsOnNL() {
			flags[2] = 'L'
		}
		res = append(res, fmt.Sprintf("%d:%d:%d:%s", ctx.GetStart(), ctx.GetEnd(), ctx.GetWidth(), flags))
		ctx.SetStart(ctx.GetEnd())
		guard++
		if ctx.IsFinished() || guard >= 500 {
			break
		}
	}
	return strings.Join(res, " "), ctx.GetMaster()
}

// TestBreakerBreakTextAgainstJava compares BreakerBreakText with the Java
// Breaker.breakText. testdata/layout/breaker/breaker_java.tsv was written by
// a Java harness that ran Flying Saucer's Breaker with a text renderer for
// which every code point is 10 units wide. The columns are: line breaking
// strategy ("default" or "hyphen3"), available width, full line width,
// white-space, word-wrap, word-break, letter-spacing, the text, the recorded
// calls (see breakerTestRun; Java's UTF-16 offsets were converted to UTF-8 byte
// offsets), and the master text after the last call ("=" when it is the
// text).
func TestBreakerBreakTextAgainstJava(t *testing.T) {
	data, err := os.ReadFile("testdata/layout/breaker/breaker_java.tsv")
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSharedContextWithUacFrRefTrDpi(nil, breakerTestFontResolver{}, nil, breakerTestTextRenderer{}, 96)
	def := sc.GetLineBreakingStrategy()
	if _, ok := def.(*DefaultLineBreakingStrategy); !ok {
		t.Fatalf("the line breaking strategy of a new SharedContext is %T, want *DefaultLineBreakingStrategy", def)
	}
	failures := 0
	cases := 0
	for lineNo, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\t", 10)
		if len(f) != 10 {
			t.Fatalf("line %d: %d columns", lineNo+1, len(f))
		}
		cases++
		if f[0] == "hyphen3" {
			sc.SetLineBreakingStrategy(breakerTestHyphen3Strategy{})
		} else {
			sc.SetLineBreakingStrategy(def)
		}
		avail, _ := strconv.Atoi(f[1])
		full, _ := strconv.Atoi(f[2])
		letterSpacing, _ := strconv.ParseFloat(f[6], 32)
		style := &breakerTestStyle{
			whitespace:    IdentValueValueOf(f[3]),
			wordWrap:      IdentValueValueOf(f[4]),
			wordBreak:     IdentValueValueOf(f[5]),
			letterSpacing: float32(letterSpacing),
		}
		text := breakerTestUnescape(f[7])
		steps, master := breakerTestRun(sc, text, avail, full, style)
		wantMaster := text
		if f[9] != "=" {
			wantMaster = breakerTestUnescape(f[9])
		}
		if steps != f[8] || master != wantMaster {
			failures++
			if failures <= 20 {
				t.Errorf("line %d: %s avail=%s full=%s white-space=%s word-wrap=%s word-break=%s letter-spacing=%s text=%q\n got  %s master %q\n want %s master %q",
					lineNo+1, f[0], f[1], f[2], f[3], f[4], f[5], f[6], text, steps, breakerTestEscape(master), f[8], breakerTestEscape(wantMaster))
			}
		}
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
	if failures > 0 {
		t.Errorf("%d of %d cases differ from Java", failures, cases)
	}
}

// TestBreaker_wordWrapBreakWordSingleCharOverflow covers what the Java test
// WordWrapBreakWordSingleCharOverflowTest checks through the Swing renderer
// (Graphics2DRenderer, not ported): with word-wrap: break-word and a single
// character wider than the full line width, doBreakText consumes one character
// per call and never marks the text unbreakable, so the layout makes progress
// and produces one line per character.
func TestBreaker_wordWrapBreakWordSingleCharOverflow(t *testing.T) {
	sc := NewSharedContextWithUacFrRefTrDpi(nil, breakerTestFontResolver{}, nil, breakerTestTextRenderer{}, 96)
	style := &breakerTestStyle{
		whitespace: IdentValueNormal,
		wordWrap:   IdentValueBreakWord,
		wordBreak:  IdentValueNormal,
	}
	// every character is 10 wide, the line is 5 wide
	steps, _ := breakerTestRun(sc, "WWW", 5, 5, style)
	want := "0:1:10:N-- 1:2:10:N-- 2:3:10:N--"
	if steps != want {
		t.Errorf("got %s, want %s", steps, want)
	}
}

func TestBreakerBreakFirstLetter(t *testing.T) {
	sc := NewSharedContextWithUacFrRefTrDpi(nil, breakerTestFontResolver{}, nil, breakerTestTextRenderer{}, 96)
	c := sc.NewLayoutContextInstance(nil)
	style := &breakerTestStyle{whitespace: IdentValueNormal, wordWrap: IdentValueNormal, wordBreak: IdentValueNormal}
	tests := []struct {
		text        string
		start       int
		avail       int
		end         int
		width       int
		unbreakable bool
	}{
		{"Hello", 0, 100, 1, 10, false},
		{"\"Hello", 0, 100, 2, 20, false},
		{"“Été", 0, 100, len("“É"), 20, false},
		{"A", 0, 100, 1, 10, false},
		{"Hello", 0, 5, 1, 10, true},
		{"  Hello", 2, 100, 3, 10, false},
	}
	for _, test := range tests {
		ctx := NewLineBreakContext(test.text, nil)
		ctx.SetStart(test.start)
		BreakerBreakFirstLetter(c, ctx, test.avail, style)
		if ctx.GetEnd() != test.end || ctx.GetWidth() != test.width ||
			ctx.IsUnbreakable() != test.unbreakable || ctx.IsNeedsNewLine() != test.unbreakable {
			t.Errorf("%q: end %d width %d unbreakable %v needsNewLine %v, want end %d width %d unbreakable %v",
				test.text, ctx.GetEnd(), ctx.GetWidth(), ctx.IsUnbreakable(), ctx.IsNeedsNewLine(),
				test.end, test.width, test.unbreakable)
		}
	}
}
