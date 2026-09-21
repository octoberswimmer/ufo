// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/layout/breaker/UrlAwareLineBreakIteratorTest.java

package ufo

import "testing"

func TestUrlAwareLineBreakIterator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		segments []string
	}{
		{"breakAtSpace", "Hello World! World foo",
			[]string{"Hello ", "World! ", "World ", "foo"}},
		{"breakAtPunctuation", "The.quick,brown:fox;jumps!over?the(lazy)[dog]",
			[]string{"The.", "quick,", "brown:", "fox;", "jumps!", "over?", "the", "(lazy)", "[dog]"}},
		{"breakAtHyphen", "Pseudo-element",
			[]string{"Pseudo-", "element"}},
		{"breakAtSlash", "Justice/Law",
			[]string{"Justice", "/Law"}},
		{"wordBeginsWithSlash", "Justice /Law",
			[]string{"Justice ", "/Law"}},
		{"wordEndsWithSlash", "Justice/ Law",
			[]string{"Justice/ ", "Law"}},
		{"wordEndsWithSlashMultipleWhitespace", "Justice/    Law",
			[]string{"Justice/    ", "Law"}},
		{"slashSeparatedSequence", "/this/is/a/long/path/name/",
			[]string{"/this", "/is", "/a", "/long", "/path", "/name/"}},
		{"urlInside", "Sentence with url https://github.com/flyingsaucerproject/flyingsaucer?test=true&param2=false inside.",
			[]string{"Sentence ", "with ", "url ", "https://github.", "com", "/flyingsaucerproject", "/flyingsaucer?",
				"test=true&param2=false ", "inside."}},
		{"multipleSlashesInWord", "word/////word",
			[]string{"word", "/////word"}},
		{"multipleSlashesBeforeWord", "hello /////world",
			[]string{"hello ", "/////world"}},
		{"multipleSlashesAfterWord", "hello world/////",
			[]string{"hello ", "world/////"}},
		{"multipleSlashesAroundWord", "hello /////world/////",
			[]string{"hello ", "/////world/////"}},
		{"whitespaceAfterTrailingSlashes", "hello world///    ",
			[]string{"hello ", "world///    "}},
		{"shortUrl", "http://localhost",
			[]string{"http://localhost"}},
		{"incompleteUrl", "http://",
			[]string{"http://"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlAwareLineBreakIteratorTestAssertBreaksCorrectly(t, test.input, test.segments)
		})
	}
}

func urlAwareLineBreakIteratorTestAssertBreaksCorrectly(t *testing.T, input string, segments []string) {
	t.Helper()
	var iterator BreakIteratorI = NewUrlAwareLineBreakIterator(input)

	segmentIndex := 0
	lastBreakPoint := 0
	for breakpoint := iterator.Next(); breakpoint != BreakIteratorDone; breakpoint = iterator.Next() {
		if segmentIndex < len(segments) {
			segment := segments[segmentIndex]
			segmentIndex++
			if got := input[lastBreakPoint:breakpoint]; got != segment {
				t.Errorf("Segment #%d does not match: got %q, want %q", segmentIndex, got, segment)
			}
			lastBreakPoint = breakpoint
		} else {
			t.Fatal("Too few segments.")
		}
	}
	if lastBreakPoint != len(input) {
		t.Errorf("Last breakpoint is wrong: got %d, want %d", lastBreakPoint, len(input))
	}
	if segmentIndex != len(segments) {
		t.Error("Too many segments.")
	}
}
