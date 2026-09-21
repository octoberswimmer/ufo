package ufo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

// breakIteratorUnescape decodes the \uXXXX escapes of the JDK result files
// into UTF-16 code units.
func breakIteratorUnescape(t *testing.T, escaped string) []uint16 {
	t.Helper()
	var units []uint16
	for i := 0; i < len(escaped); i++ {
		if escaped[i] == '\\' && i+5 < len(escaped)+0 && escaped[i+1] == 'u' {
			value, err := strconv.ParseUint(escaped[i+2:i+6], 16, 16)
			if err != nil {
				t.Fatalf("bad escape in %q: %v", escaped, err)
			}
			units = append(units, uint16(value))
			i += 5
			continue
		}
		units = append(units, uint16(escaped[i]))
	}
	return units
}

// breakIteratorAllBoundaries returns the boundaries that First and Next
// yield, converted from byte offsets to offsets in UTF-16 code units, which
// is what the JDK reports.
func breakIteratorAllBoundaries(text string) []int {
	iterator := BreakIteratorGetLineInstance()
	iterator.SetText(text)
	var result []int
	for p := iterator.First(); p != BreakIteratorDone; p = iterator.Next() {
		result = append(result, len(utf16.Encode([]rune(text[:p]))))
	}
	return result
}

func breakIteratorFormat(positions []int) string {
	parts := make([]string, len(positions))
	for i, p := range positions {
		parts[i] = strconv.Itoa(p)
	}
	return strings.Join(parts, " ")
}

// breakIteratorKnownDeviation matches the texts for which the port is known
// to differ from the JDK. They all have a "$" or a "'" (category GL) that the
// state machine of the JDK treats as the prefix of a number:
//
//   - A GL followed by a NB character, with nothing between them but the
//     characters of a number prefix (GL, PR, QU, dashes), spaces and format
//     characters. The JDK sometimes ends the segment before the NB
//     ("a$$" | NB, NB "$" | NB) and sometimes does not ("a$" NB, "$$" NB).
//     The port never ends it there.
//   - A GL, a run of GL, PR, QU and dash characters that ends with two or
//     more dashes, and a PO character. The JDK sometimes keeps the PO in the
//     segment ("$--)") and sometimes does not ("a'--" | ","). The port never
//     keeps it.
//
// The pattern is over one letter per character: G for GL, N for NB, S for
// SP, F for CF, D for DA, P for PR, Q for QU, O for PO, MN and PM, and x for
// the rest.
var breakIteratorKnownDeviation = regexp.MustCompile(`G[GPQDSF]*N|G[GPQDF]*D[F]*D[DF]*[OQ]`)

func breakIteratorClassLetters(text string) string {
	var letters strings.Builder
	for _, r := range text {
		switch breakIteratorClassOf(r) {
		case breakIteratorGL:
			letters.WriteByte('G')
		case breakIteratorNB:
			letters.WriteByte('N')
		case breakIteratorSP:
			letters.WriteByte('S')
		case breakIteratorCF:
			letters.WriteByte('F')
		case breakIteratorDA:
			letters.WriteByte('D')
		case breakIteratorPR:
			letters.WriteByte('P')
		case breakIteratorQU:
			letters.WriteByte('Q')
		case breakIteratorPO, breakIteratorMN, breakIteratorPM:
			letters.WriteByte('O')
		default:
			letters.WriteByte('x')
		}
	}
	return letters.String()
}

// checkBreakIteratorAgainstJdkFile compares the iterator with a file of JDK
// results. Each line is a text with \uXXXX escapes for UTF-16 code units, a
// tab, and the boundaries that BreakIterator.getLineInstance(Locale.ROOT)
// of OpenJDK 25 returns for it, as offsets in UTF-16 code units. A text may
// differ from the JDK only if breakIteratorKnownDeviation matches it.
func checkBreakIteratorAgainstJdkFile(t *testing.T, path string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	failures := 0
	deviations := 0
	lines := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		escaped, want, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatalf("no tab in line %q", line)
		}
		lines++
		text := string(utf16.Decode(breakIteratorUnescape(t, escaped)))
		got := breakIteratorFormat(breakIteratorAllBoundaries(text))
		if got != want {
			if breakIteratorKnownDeviation.MatchString(breakIteratorClassLetters(text)) {
				deviations++
				continue
			}
			failures++
			if failures <= 40 {
				t.Errorf("%s: boundaries %s, JDK has %s", escaped, got, want)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if failures > 0 {
		t.Errorf("%d of %d texts differ from the JDK", failures, lines)
	}
	t.Logf("%d texts, %d known deviations", lines, deviations)
}

func TestBreakIteratorJdkResults(t *testing.T) {
	checkBreakIteratorAgainstJdkFile(t, "testdata/util/break_iterator_jdk.txt")
}

// TestBreakIteratorJdkResultsExtra runs over a larger file of JDK results
// named by the environment variable UFO_BREAK_ITERATOR_JDK_FILE, when set.
func TestBreakIteratorJdkResultsExtra(t *testing.T) {
	path := os.Getenv("UFO_BREAK_ITERATOR_JDK_FILE")
	if path == "" {
		t.Skip("UFO_BREAK_ITERATOR_JDK_FILE is not set")
	}
	checkBreakIteratorAgainstJdkFile(t, path)
}

// TestBreakIteratorLineBreakOpportunities lists the break opportunities of
// texts for which the result of the JDK is well known. "|" marks a boundary
// inside the text; the start and the end of the text are boundaries as well.
func TestBreakIteratorLineBreakOpportunities(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"break after spaces", "Hello |world |foo"},
		{"break after the whole run of spaces", "a   |b"},
		{"leading space is a segment", " |a"},
		{"break after a hyphen between letters", "well-|known"},
		{"break after a hyphen before a digit that follows a letter", "a-|1"},
		{"no break between a leading hyphen and a digit", "a |-1"},
		{"break after each run of dashes", "a--|b"},
		{"break after en dash and em dash", "a\u2013|b\u2014|c"},
		{"break after soft hyphen", "hy\u00ad|phen"},
		{"no break around non-breaking hyphen", "a\u2011b"},
		{"no break around NBSP", "a\u00a0b |c"},
		{"NBSP keeps the spaces around it", "a \u00a0 b"},
		{"no break around WORD JOINER", "a\u2060b"},
		{"no break around ZERO WIDTH JOINER", "a\u200db"},
		{"the JDK does not break after ZERO WIDTH SPACE", "a\u200bb"},
		{"no break inside a number", "1,000.50 |is"},
		{"currency sign stays with the number", "$1,000.50"},
		{"percent sign stays with the number", "50% |abc"},
		{"period ends a word", "www.|example.|com"},
		{"comma and colon end a word", "a,|b:|c"},
		{"time is broken after the colon", "12:|30"},
		{"no break before closing punctuation", "(abc) |def"},
		{"no break after opening punctuation", "abc |(def)"},
		{"break before opening punctuation after a letter", "abc|(def)"},
		{"break after closing punctuation before a letter", "a)|b"},
		{"empty parentheses stay with the word", "foo()| |bar"},
		{"quotation mark opens after a space and closes after a letter", "a |\"b\" |c"},
		{"apostrophe does not break", "it's |a"},
		{"slash does not break", "a/b"},
		{"break before and after ideographs", "\u4e2d|\u6587|\u5b57"},
		{"ideographic full stop stays with the ideograph before it", "\u4e2d\u3002|\u6587"},
		{"corner brackets stay with the ideographs", "\u300c\u4e2d\u300d|\u6587"},
		{"break between letters and ideographs", "abc|\u4e2d|def"},
		{"small kana stays with the kana before it", "\u30a2\u30c3|\u30d7"},
		{"Hangul syllables break", "\uac00|\uac01"},
		{"mandatory break after a newline", "a\n|b"},
		{"CR LF is one break", "a\r\n|b"},
		{"CR alone breaks", "a\r|b"},
		{"each newline is a break", "a\n|\n|b"},
		{"tab ends a segment", "a\t|b"},
		{"line separator breaks", "a\u2028|b"},
		{"combining mark stays with its base", "e\u0301a |b"},
		{"emoji do not break", "\U0001F600\U0001F600"},
		{"Thai is not broken without a dictionary", "\u0e01\u0e02\u0e03"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text := strings.ReplaceAll(test.text, "|", "")
			want := []int{0}
			offset := 0
			for _, part := range strings.Split(test.text, "|") {
				offset += len(part)
				want = append(want, offset)
			}
			iterator := BreakIteratorGetLineInstance()
			iterator.SetText(text)
			var got []int
			for p := iterator.First(); p != BreakIteratorDone; p = iterator.Next() {
				got = append(got, p)
			}
			if breakIteratorFormat(got) != breakIteratorFormat(want) {
				t.Errorf("%q: boundaries %v, want %v", text, got, want)
			}
		})
	}
}

// TestBreakIteratorNavigation compares Following, Preceding, IsBoundary,
// NextWithN, Last and Previous with the results of the JDK for an ASCII
// text, in which byte offsets and UTF-16 offsets are the same.
func TestBreakIteratorNavigation(t *testing.T) {
	data, err := os.ReadFile("testdata/util/break_iterator_jdk_api.txt")
	if err != nil {
		t.Fatal(err)
	}
	var iterator BreakIteratorI = BreakIteratorGetLineInstance()
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		switch {
		case strings.HasPrefix(line, "# ") && iterator.GetText() == "":
			iterator.SetText(strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "# next(n) from first"):
			iterator.First()
		case strings.HasPrefix(line, "#"):
		case fields[0] == "n":
			n, _ := strconv.Atoi(fields[1])
			got := fmt.Sprintf("n %d %d %d", n, iterator.NextWithN(n), iterator.Current())
			if got != line {
				t.Errorf("NextWithN: got %q, JDK has %q", got, line)
			}
		case fields[0] == "l":
			got := fmt.Sprintf("l %d %d %d %d %d %d %d", iterator.Last(), iterator.Next(), iterator.Current(),
				iterator.Previous(), iterator.First(), iterator.Previous(), iterator.Current())
			if got != line {
				t.Errorf("Last and Previous: got %q, JDK has %q", got, line)
			}
		case fields[0] == "e":
			iterator.SetText("")
			got := fmt.Sprintf("e %d %d %d %d %d %d %d %t", iterator.First(), iterator.Next(), iterator.Last(),
				iterator.Previous(), iterator.Current(), iterator.Following(0), iterator.Preceding(0), iterator.IsBoundary(0))
			if got != line {
				t.Errorf("empty text: got %q, JDK has %q", got, line)
			}
		default:
			offset, _ := strconv.Atoi(fields[0])
			following := iterator.Following(offset)
			followingCurrent := iterator.Current()
			preceding := iterator.Preceding(offset)
			precedingCurrent := iterator.Current()
			isBoundary := iterator.IsBoundary(offset)
			got := fmt.Sprintf("%d %d %d %d %d %t %d", offset, following, followingCurrent, preceding, precedingCurrent,
				isBoundary, iterator.Current())
			if got != line {
				t.Errorf("offset %d: got %q, JDK has %q", offset, got, line)
			}
		}
	}
}

func TestBreakIteratorOffsetOutOfBounds(t *testing.T) {
	iterator := BreakIteratorGetLineInstance()
	iterator.SetText("abc")
	for _, offset := range []int{-1, 4} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Following(%d) did not panic", offset)
				}
			}()
			iterator.Following(offset)
		}()
	}
}

// TestBreakIteratorByteOffsets checks that boundaries are byte offsets into
// the Go string.
func TestBreakIteratorByteOffsets(t *testing.T) {
	text := "caf\u00e9 \u4e2d\u6587 \U0001F600 x"
	iterator := BreakIteratorGetLineInstance()
	iterator.SetText(text)
	var segments []string
	start := iterator.First()
	for end := iterator.Next(); end != BreakIteratorDone; end = iterator.Next() {
		segments = append(segments, text[start:end])
		start = end
	}
	want := []string{"caf\u00e9 ", "\u4e2d", "\u6587 ", "\U0001F600 ", "x"}
	if strings.Join(segments, "|") != strings.Join(want, "|") {
		t.Errorf("segments %q, want %q", segments, want)
	}
}
