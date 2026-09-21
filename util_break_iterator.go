// Stands for the JDK class java.text.BreakIterator as
// flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/ uses it:
// the line instance of the root locale.

package ufo

import (
	"sort"
	"unicode"
	"unicode/utf8"
)

// BreakIteratorDone is BreakIterator.DONE: the value that Next, Previous,
// Following and Preceding return when there is no further boundary.
const BreakIteratorDone = -1

// BreakIteratorI lists the methods of java.text.BreakIterator, which is an
// abstract class that Flying Saucer extends (UrlAwareLineBreakIterator).
type BreakIteratorI interface {
	First() int
	Last() int
	Next() int
	NextWithN(n int) int
	Previous() int
	Following(offset int) int
	Preceding(offset int) int
	IsBoundary(offset int) bool
	Current() int
	GetText() string
	SetText(newText string)
}

// BreakIterator finds the positions in a text at which a line may be broken.
// It reproduces BreakIterator.getLineInstance() of the JDK for the root
// locale.
//
// Positions are byte offsets into the Go string, where Java has offsets in
// UTF-16 code units: text[a:b] between two boundaries is what
// text.substring(a, b) is in Java. A boundary is never inside the UTF-8
// encoding of a character.
//
// # Rules
//
// The JDK does not implement the pair table of UAX #14. Its line iterator
// is a state machine compiled from a grammar over character categories
// (sun.text.resources.BreakIteratorRules, "LineBreakRules"), which predates
// UAX #14 and agrees with it for most Latin and CJK text. This type
// implements that grammar, because the positions at which Flying Saucer
// breaks lines, and the code of UrlAwareLineBreakIterator, depend on what
// the JDK returns. The character categories and the grammar were determined
// by running the JDK (OpenJDK 25) over every code point and over generated
// strings; testdata/util/break_iterator_jdk.txt holds JDK results that the
// tests compare against.
//
// A text is cut into segments, and there is a break opportunity after each
// segment. A segment is the longest prefix of the remaining text that
// matches
//
//	segment := unit ( NB+ SP? unit )* tail CR? BK?
//	unit    := word ( GL+ word )* SP*
//	word    := PR* core suffix
//	core    := ( CH | DG | MK )*  |  KJ  |  number
//	number  := ( PR | QU | GL | DA )* DG+ ( MN DG+ )*
//	suffix  := DA+  |  PO*
//	tail    := MK* ( "(" ")" PO* MK* )?
//
// Every part may be empty, but a segment has at least one character. CF
// characters are skipped wherever they occur and stay in the segment in
// which they occur. A character of the class XX is a segment by itself.
//
// The tail holds two things that the JDK attaches to a segment and after
// which nothing but the line terminator can follow, not even a space:
// combining marks that do not follow a word character, and the two
// characters "()" after a word, a suffix or spaces, so that "foo() bar" has
// the segments "foo()", " " and "bar", while "foo(x) bar" has "foo",
// "(x) " and "bar". Other pairs of brackets are not treated like that.
//
// Known differences from the JDK, all for a "$" or a "'" that its state
// machine takes for the prefix of a number: before a NB that follows such a
// character, with nothing but number prefix characters and spaces between
// them, the JDK sometimes ends the segment ("a$$" | NB) and sometimes does
// not ("a$" NB), and this type never does; and after such a character and a
// run of prefix characters that ends with two or more dashes, the JDK
// sometimes keeps a following PO in the segment ("$--)") and this type never
// does. util_break_iterator_test.go has the pattern of these texts.
//
// The categories, with the UAX #14 line breaking classes they correspond to,
// and the rules of UAX #14 that the grammar reproduces:
//
//   - BK: U+0003, tab, LF, FF, LS (U+2028), PS (U+2029). UAX #14 classes BK
//     and LF. A segment ends after BK: the mandatory break of LB4 and LB5.
//     The JDK also ends a segment after a tab. It treats VT (U+000B) and
//     NEL (U+0085, UAX #14 class NL) as spaces.
//   - CR: carriage return. UAX #14 class CR. CR BK is one terminator (LB5:
//     no break between CR and LF), and a CR alone ends the segment.
//   - SP: the categories Zs and Cc other than the characters above and NB.
//     UAX #14 classes SP and, for the fixed-width spaces U+2000..U+200A, BA.
//     Spaces end a unit, so the break opportunity is after the run of spaces
//     that follows a word, never before it or inside it (LB7, LB18).
//   - NB: U+00A0 NO-BREAK SPACE, U+2007 FIGURE SPACE, U+2011 NON-BREAKING
//     HYPHEN, U+202F NARROW NO-BREAK SPACE, U+0F0C, U+FEFF. UAX #14 class GL
//     (and WJ for U+FEFF). NB joins the units on both sides, including the
//     spaces before it and one space after it (LB12).
//   - CF: the category Cf except U+00AD and U+FEFF; it holds U+2060 WORD
//     JOINER (UAX #14 class WJ), U+200D ZERO WIDTH JOINER (ZWJ), the
//     bidirectional controls, and U+200B ZERO WIDTH SPACE (ZW). The JDK
//     skips these characters, so there is no break before or after WORD
//     JOINER (LB11). There is also no break after ZERO WIDTH SPACE, where
//     LB8 of UAX #14 has one: for the JDK it is a format character like the
//     others.
//   - DA: the category Pd other than U+2011, and U+00AD SOFT HYPHEN. UAX #14
//     classes HY (hyphen-minus), BA (hyphens, en dash, soft hyphen) and B2
//     (em dash). A run of dashes ends a word, which gives the break
//     opportunity after a hyphen between letters ("well-" | "known") and
//     after a soft hyphen (LB21: no break before HY and BA, break after).
//     A dash followed by a digit at the start of a word belongs to the
//     number ("-1", LB25: HY × NU); after a letter or a digit it ends the
//     word ("a-" | "1"), where LB25 would keep "a-1" together.
//   - PR: the categories Ps, Pi and Sc except "$" and U+00A2. UAX #14 classes
//     OP (opening punctuation), QU (initial quotes) and PR (currency
//     symbols). They start a word: no break after them (LB14, LB25: PR × NU,
//     PR × AL), and a break opportunity before them even without a space
//     ("abc" | "(def)"), which UAX #14 does not have after a letter (LB30).
//   - PO: the categories Pe and Pf, and ! % : ; ? ¢ ° ‰ ‱ ′ ″ ‴ ℃ ℅ ℉, the
//     ideographic comma and full stop, the iteration marks, the small kana,
//     the prolonged sound mark, and the fullwidth forms of ! % , . : ; ?.
//     UAX #14 classes CL, CP (closing punctuation), EX (! ?), IS (: ;),
//     PO (% ° and the like), and NS and CJ (small kana, iteration marks).
//     They end a word: no break before them (LB13, LB21 for NS), and a
//     break opportunity after the run of them even when a letter follows
//     ("a:" | "b", "a)" | "b"), where LB29 and LB30 of UAX #14 have none.
//   - MN: "." and ",". UAX #14 class IS. Inside a number, between two
//     digits, they do not end the word, so "1,000.50" is unbroken (LB25:
//     NU × IS, IS × NU). Anywhere else they are PO: "example." | "com".
//   - QU: the quotation mark ". UAX #14 class QU. It is PR and PO at once:
//     it starts a word after a space and ends a word after a letter.
//   - GL: "$" and "'". They join the words on both sides ("a.$b", "it's"),
//     but not across spaces, and they can be part of the prefix of a number.
//     UAX #14 has PR for "$" and QU for "'".
//   - DG: the categories Nd and No. UAX #14 class NU. Digits are word
//     characters like CH, and they also form numbers.
//   - KJ: U+4E00..U+9FA5 (CJK unified ideographs), U+F900..U+FA2D and
//     U+FA30..U+FA6A (compatibility ideographs), U+AC00..U+D7A3 (Hangul
//     syllables), and hiragana U+3041..U+3094 and katakana U+30A1..U+30FA
//     except the small kana. UAX #14 classes ID, H2 and H3. One such
//     character is a whole word core, so there is a break opportunity before
//     and after each ideograph (LB31 applied to ID), except that closing
//     punctuation stays with the ideograph before it and opening punctuation
//     with the one after it. Ideographs outside these ranges (extension A,
//     the supplementary planes) are CH in the JDK.
//   - MK: the categories Mn and Me. UAX #14 class CM. Where a word character
//     can follow, a mark is a word character, so it stays with its base
//     (LB9) and the word goes on after it. After a suffix, after spaces and
//     after an ideograph it is attached as well, as part of the tail.
//   - PM: U+3099 and U+309A, the combining kana voicing marks. They are PO
//     where PO can follow and MK elsewhere.
//   - XX: the dandas U+0964 and U+0965, U+FFFF, and the last code point of
//     each range of Cf characters in the supplementary planes (U+110BD,
//     U+110CD, U+1343F, U+1BCA3, U+1D17A, U+E0001, U+E007F), which the
//     category tables of the JDK leave out. No rule matches them.
//   - CH: every other character. UAX #14 class AL among others: letters,
//     symbols such as / @ & = + _ * # ~ | < >, U+2026 HORIZONTAL ELLIPSIS
//     and U+2212 MINUS SIGN, emoji, and the scripts that UAX #14 breaks with
//     a dictionary (class SA, e.g. Thai), which the root locale of the JDK
//     does not break.
//
// The categories come from the Unicode tables of the Go standard library,
// so a character assigned in a Unicode version that the Go tables do not
// have yet is CH.
type BreakIterator struct {
	text string
	// boundaries are the break positions in ascending order. The first is 0
	// and the last is len(text).
	boundaries []int
	// current is the index into boundaries of the current position.
	current int
}

// BreakIteratorGetLineInstance ports BreakIterator.getLineInstance(). The
// text is empty until SetText is called.
func BreakIteratorGetLineInstance() *BreakIterator {
	b := &BreakIterator{}
	b.SetText("")
	return b
}

// SetText sets a new text to scan and moves the current position to the
// start of it.
func (b *BreakIterator) SetText(newText string) {
	b.text = newText
	b.boundaries = b.boundaries[:0]
	b.boundaries = append(b.boundaries, 0)
	for pos := 0; pos < len(newText); {
		pos = breakIteratorSegmentEnd(newText, pos)
		b.boundaries = append(b.boundaries, pos)
	}
	b.current = 0
}

// GetText returns the text being scanned. Java returns a CharacterIterator
// over it.
func (b *BreakIterator) GetText() string {
	return b.text
}

// First moves to the first boundary, which is 0, and returns it.
func (b *BreakIterator) First() int {
	b.current = 0
	return b.boundaries[b.current]
}

// Last moves to the last boundary, which is the length of the text, and
// returns it.
func (b *BreakIterator) Last() int {
	b.current = len(b.boundaries) - 1
	return b.boundaries[b.current]
}

// Next moves to the boundary following the current one and returns it. At
// the last boundary it returns BreakIteratorDone and does not move.
func (b *BreakIterator) Next() int {
	if b.current == len(b.boundaries)-1 {
		return BreakIteratorDone
	}
	b.current++
	return b.boundaries[b.current]
}

// NextWithN ports BreakIterator.next(int n): it moves n boundaries forward,
// or -n boundaries backward for a negative n, and returns the boundary
// reached, or BreakIteratorDone when the first or last boundary was reached
// before that. NextWithN(0) returns the current boundary.
func (b *BreakIterator) NextWithN(n int) int {
	result := b.Current()
	for n > 0 {
		result = b.Next()
		n--
	}
	for n < 0 {
		result = b.Previous()
		n++
	}
	return result
}

// Previous moves to the boundary preceding the current one and returns it.
// At the first boundary it returns BreakIteratorDone and does not move.
func (b *BreakIterator) Previous() int {
	if b.current == 0 {
		return BreakIteratorDone
	}
	b.current--
	return b.boundaries[b.current]
}

func (b *BreakIterator) checkOffset(offset int) {
	if offset < 0 || offset > len(b.text) {
		panic(NewXRRuntimeException("offset out of bounds"))
	}
}

// Following moves to the first boundary after offset and returns it. When
// there is none, because offset is the length of the text, it moves to the
// last boundary and returns BreakIteratorDone. An offset outside of the text
// is a panic, as it is an IllegalArgumentException in Java.
func (b *BreakIterator) Following(offset int) int {
	b.checkOffset(offset)
	i := sort.SearchInts(b.boundaries, offset+1)
	if i == len(b.boundaries) {
		b.current = len(b.boundaries) - 1
		return BreakIteratorDone
	}
	b.current = i
	return b.boundaries[b.current]
}

// Preceding moves to the last boundary before offset and returns it. When
// there is none, because offset is 0, it moves to the first boundary and
// returns BreakIteratorDone. An offset outside of the text is a panic, as it
// is an IllegalArgumentException in Java.
func (b *BreakIterator) Preceding(offset int) int {
	b.checkOffset(offset)
	i := sort.SearchInts(b.boundaries, offset) - 1
	if i < 0 {
		b.current = 0
		return BreakIteratorDone
	}
	b.current = i
	return b.boundaries[b.current]
}

// IsBoundary reports whether offset is a boundary. As in Java, the current
// position becomes offset when it is a boundary, and the first boundary
// after offset otherwise.
func (b *BreakIterator) IsBoundary(offset int) bool {
	b.checkOffset(offset)
	if offset == 0 {
		b.First()
		return true
	}
	return b.Following(offset-1) == offset
}

// Current returns the current boundary.
func (b *BreakIterator) Current() int {
	return b.boundaries[b.current]
}

// breakIteratorClass is a character category of the grammar in the
// documentation of BreakIterator.
type breakIteratorClass int

const (
	breakIteratorCH breakIteratorClass = iota
	breakIteratorBK
	breakIteratorCR
	breakIteratorSP
	breakIteratorNB
	breakIteratorCF
	breakIteratorDA
	breakIteratorPR
	breakIteratorPO
	breakIteratorMN
	breakIteratorQU
	breakIteratorGL
	breakIteratorDG
	breakIteratorKJ
	breakIteratorMK
	breakIteratorPM
	breakIteratorXX
)

// breakIteratorPostWord holds the PO characters that are not in the
// categories Pe and Pf.
var breakIteratorPostWord = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x0021, Hi: 0x0021, Stride: 1}, // !
		{Lo: 0x0025, Hi: 0x0025, Stride: 1}, // %
		{Lo: 0x003a, Hi: 0x003b, Stride: 1}, // : ;
		{Lo: 0x003f, Hi: 0x003f, Stride: 1}, // ?
		{Lo: 0x00a2, Hi: 0x00a2, Stride: 1}, // cent sign
		{Lo: 0x00b0, Hi: 0x00b0, Stride: 1}, // degree sign
		{Lo: 0x066a, Hi: 0x066a, Stride: 1}, // Arabic percent sign
		{Lo: 0x2030, Hi: 0x2034, Stride: 1}, // per mille, per ten thousand, primes
		{Lo: 0x2103, Hi: 0x2103, Stride: 1}, // degree Celsius
		{Lo: 0x2105, Hi: 0x2105, Stride: 1}, // care of
		{Lo: 0x2109, Hi: 0x2109, Stride: 1}, // degree Fahrenheit
		{Lo: 0x3001, Hi: 0x3002, Stride: 1}, // ideographic comma and full stop
		{Lo: 0x3005, Hi: 0x3005, Stride: 1}, // ideographic iteration mark
		{Lo: 0x3041, Hi: 0x3049, Stride: 2}, // small hiragana a i u e o
		{Lo: 0x3063, Hi: 0x3063, Stride: 1}, // small hiragana tu
		{Lo: 0x3083, Hi: 0x3087, Stride: 2}, // small hiragana ya yu yo
		{Lo: 0x308e, Hi: 0x308e, Stride: 1}, // small hiragana wa
		{Lo: 0x309b, Hi: 0x309e, Stride: 1}, // voicing marks, hiragana iteration marks
		{Lo: 0x30a1, Hi: 0x30a9, Stride: 2}, // small katakana a i u e o
		{Lo: 0x30c3, Hi: 0x30c3, Stride: 1}, // small katakana tu
		{Lo: 0x30e3, Hi: 0x30e7, Stride: 2}, // small katakana ya yu yo
		{Lo: 0x30ee, Hi: 0x30ee, Stride: 1}, // small katakana wa
		{Lo: 0x30f5, Hi: 0x30f6, Stride: 1}, // small katakana ka ke
		{Lo: 0x30fc, Hi: 0x30fe, Stride: 1}, // prolonged sound mark, katakana iteration marks
		{Lo: 0xff01, Hi: 0xff01, Stride: 1}, // fullwidth !
		{Lo: 0xff05, Hi: 0xff05, Stride: 1}, // fullwidth %
		{Lo: 0xff0c, Hi: 0xff0c, Stride: 1}, // fullwidth ,
		{Lo: 0xff0e, Hi: 0xff0e, Stride: 1}, // fullwidth .
		{Lo: 0xff1a, Hi: 0xff1b, Stride: 1}, // fullwidth : ;
		{Lo: 0xff1f, Hi: 0xff1f, Stride: 1}, // fullwidth ?
	},
}

// breakIteratorKanji holds the KJ ranges. The small kana inside them are PO,
// which breakIteratorClassOf tests first.
var breakIteratorKanji = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x3041, Hi: 0x3094, Stride: 1},
		{Lo: 0x30a1, Hi: 0x30fa, Stride: 1},
		{Lo: 0x4e00, Hi: 0x9fa5, Stride: 1},
		{Lo: 0xac00, Hi: 0xd7a3, Stride: 1},
		{Lo: 0xf900, Hi: 0xfa2d, Stride: 1},
		{Lo: 0xfa30, Hi: 0xfa6a, Stride: 1},
	},
}

func breakIteratorClassOf(r rune) breakIteratorClass {
	switch r {
	case '\r':
		return breakIteratorCR
	case 0x0003, '\t', '\n', '\f', 0x2028, 0x2029:
		return breakIteratorBK
	case 0x00a0, 0x0f0c, 0x2007, 0x2011, 0x202f, 0xfeff:
		return breakIteratorNB
	case '.', ',':
		return breakIteratorMN
	case '"':
		return breakIteratorQU
	case '$', '\'':
		return breakIteratorGL
	case 0x0964, 0x0965, 0xffff, 0x110bd, 0x110cd, 0x1343f, 0x1bca3, 0x1d17a, 0xe0001, 0xe007f:
		return breakIteratorXX
	case 0x3099, 0x309a:
		return breakIteratorPM
	case 0x00ad:
		return breakIteratorDA
	}
	if r < 0x80 {
		// The ASCII characters that are not handled above.
		switch {
		case r <= ' ' || r == 0x7f:
			return breakIteratorSP
		case r >= '0' && r <= '9':
			return breakIteratorDG
		case r == '-':
			return breakIteratorDA
		case r == '(' || r == '[' || r == '{':
			return breakIteratorPR
		case r == ')' || r == ']' || r == '}' || r == '!' || r == '%' || r == ':' || r == ';' || r == '?':
			return breakIteratorPO
		}
		return breakIteratorCH
	}
	switch {
	case unicode.Is(unicode.Cf, r):
		return breakIteratorCF
	case unicode.In(r, unicode.Mn, unicode.Me):
		return breakIteratorMK
	case unicode.In(r, unicode.Zs, unicode.Cc):
		return breakIteratorSP
	case unicode.Is(unicode.Pd, r):
		return breakIteratorDA
	case unicode.In(r, unicode.Pe, unicode.Pf, breakIteratorPostWord):
		return breakIteratorPO
	case unicode.In(r, unicode.Sc, unicode.Ps, unicode.Pi):
		return breakIteratorPR
	case unicode.Is(breakIteratorKanji, r):
		return breakIteratorKJ
	case unicode.In(r, unicode.Nd, unicode.No):
		return breakIteratorDG
	}
	return breakIteratorCH
}

// breakIteratorStates is a set of positions in the grammar, one bit each.
// The grammar is matched as a nondeterministic automaton: the set holds
// every position that the characters read so far can have reached.
type breakIteratorStates uint32

const (
	// breakIteratorAtStart: nothing of the segment was read yet.
	breakIteratorAtStart breakIteratorStates = 1 << iota
	// breakIteratorInWordPrefix: after a PR of "PR*" at the start of a word.
	breakIteratorInWordPrefix
	// breakIteratorInNumberPrefix: after a PR or "-" of a number. A segment
	// cannot end here: the digits must follow.
	breakIteratorInNumberPrefix
	// breakIteratorInCore: after a character of "( CH | DG | MK )*".
	breakIteratorInCore
	// breakIteratorAfterKanji: after the KJ that is a whole word core.
	breakIteratorAfterKanji
	// breakIteratorInNumberDigits: after a digit of a number.
	breakIteratorInNumberDigits
	// breakIteratorAfterNumberSeparator: after a MN inside a number. A
	// segment cannot end here: a digit must follow.
	breakIteratorAfterNumberSeparator
	// breakIteratorInDashes: after a DA of the suffix "DA+".
	breakIteratorInDashes
	// breakIteratorInPostWord: after a PO of the suffix "PO*".
	breakIteratorInPostWord
	// breakIteratorInSpaces: after a SP that ends a unit.
	breakIteratorInSpaces
	// breakIteratorAfterGlue: after a GL; a word follows.
	breakIteratorAfterGlue
	// breakIteratorAfterNoBreak: after a NB; one SP and a unit follow.
	breakIteratorAfterNoBreak
	// breakIteratorAfterNoBreakSpace: after the SP that follows a NB; a
	// unit follows.
	breakIteratorAfterNoBreakSpace
	// breakIteratorInMarks: after a MK of the tail.
	breakIteratorInMarks
	// breakIteratorAfterOpenParen: after the "(" of a "()" that follows a
	// unit. A segment cannot end here: the ")" must follow.
	breakIteratorAfterOpenParen
	// breakIteratorAfterParenPair: after the ")" of such a "()", or a PO
	// that follows it.
	breakIteratorAfterParenPair
	// breakIteratorAfterCR: after the CR of the terminator.
	breakIteratorAfterCR
	// breakIteratorAtEnd: after the BK of the terminator.
	breakIteratorAtEnd
)

// breakIteratorNonAccepting are the positions at which a segment cannot end.
const breakIteratorNonAccepting = breakIteratorInNumberPrefix | breakIteratorAfterNumberSeparator | breakIteratorAfterOpenParen

// breakIteratorWordStart returns the positions reached by reading a
// character where a word can start: at the start of the segment, after the
// word prefix, after GL and after NB. Since every part of a word may be
// empty, the character may also be the suffix of an empty word, the space
// after it, or the terminator of the segment.
func breakIteratorWordStart(class breakIteratorClass, r rune) breakIteratorStates {
	switch class {
	case breakIteratorPR:
		return breakIteratorInWordPrefix | breakIteratorInNumberPrefix
	case breakIteratorQU:
		return breakIteratorInWordPrefix | breakIteratorInNumberPrefix | breakIteratorInPostWord
	case breakIteratorCH:
		return breakIteratorInCore
	case breakIteratorDG:
		return breakIteratorInCore | breakIteratorInNumberDigits
	case breakIteratorMK:
		return breakIteratorInCore | breakIteratorInMarks
	case breakIteratorKJ:
		return breakIteratorAfterKanji
	case breakIteratorDA:
		return breakIteratorInNumberPrefix | breakIteratorInDashes
	case breakIteratorGL:
		return breakIteratorInNumberPrefix | breakIteratorAfterGlue
	}
	return breakIteratorAfterWord(class)
}

// breakIteratorAfterWord returns the positions reached by reading a
// character after a complete word core: the suffix, GL and another word, the
// spaces that end the unit, NB and another unit, or the end of the segment.
func breakIteratorAfterWord(class breakIteratorClass) breakIteratorStates {
	switch class {
	case breakIteratorDA:
		return breakIteratorInDashes
	case breakIteratorPO, breakIteratorMN, breakIteratorQU, breakIteratorPM:
		return breakIteratorInPostWord
	}
	return breakIteratorAfterSuffix(class)
}

// breakIteratorAfterSuffix returns the positions reached by reading a
// character after the suffix of a word.
func breakIteratorAfterSuffix(class breakIteratorClass) breakIteratorStates {
	switch class {
	case breakIteratorGL:
		return breakIteratorAfterGlue
	}
	return breakIteratorAfterUnit(class)
}

// breakIteratorAfterUnit returns the positions reached by reading a
// character after a unit, which is also where the spaces of the unit are.
func breakIteratorAfterUnit(class breakIteratorClass) breakIteratorStates {
	switch class {
	case breakIteratorSP:
		return breakIteratorInSpaces
	case breakIteratorNB:
		return breakIteratorAfterNoBreak
	}
	return breakIteratorTerminator(class)
}

// breakIteratorTerminator returns the positions reached by reading a
// character of "MK* CR? BK?". The "(" of the tail is read in
// breakIteratorStep.
func breakIteratorTerminator(class breakIteratorClass) breakIteratorStates {
	switch class {
	case breakIteratorMK, breakIteratorPM:
		return breakIteratorInMarks
	case breakIteratorCR:
		return breakIteratorAfterCR
	case breakIteratorBK:
		return breakIteratorAtEnd
	}
	return 0
}

// breakIteratorStep returns the positions reached from one position by
// reading a character.
func breakIteratorStep(state breakIteratorStates, class breakIteratorClass, r rune) breakIteratorStates {
	next := breakIteratorStepInGrammar(state, class, r)
	if r == '(' && state&breakIteratorBeforeParenPair != 0 {
		next |= breakIteratorAfterOpenParen
	}
	return next
}

// breakIteratorBeforeParenPair are the positions after which the JDK takes
// the two characters "()" into the segment: after a word core, a suffix, the
// spaces of a unit, and the marks before the terminator.
const breakIteratorBeforeParenPair = breakIteratorInCore | breakIteratorAfterKanji | breakIteratorInNumberDigits |
	breakIteratorInDashes | breakIteratorInPostWord | breakIteratorInSpaces |
	breakIteratorInMarks

func breakIteratorStepInGrammar(state breakIteratorStates, class breakIteratorClass, r rune) breakIteratorStates {
	switch state {
	case breakIteratorAtStart, breakIteratorInWordPrefix, breakIteratorAfterGlue, breakIteratorAfterNoBreakSpace:
		return breakIteratorWordStart(class, r)
	case breakIteratorAfterNoBreak:
		if class == breakIteratorSP {
			return breakIteratorAfterNoBreakSpace
		}
		return breakIteratorWordStart(class, r)
	case breakIteratorInNumberPrefix:
		switch class {
		case breakIteratorPR, breakIteratorQU, breakIteratorDA, breakIteratorGL:
			return breakIteratorInNumberPrefix
		case breakIteratorDG:
			return breakIteratorInNumberDigits
		}
		return 0
	case breakIteratorInCore:
		switch class {
		case breakIteratorCH, breakIteratorDG:
			return breakIteratorInCore
		case breakIteratorMK:
			return breakIteratorInCore | breakIteratorInMarks
		}
		return breakIteratorAfterWord(class)
	case breakIteratorAfterKanji:
		return breakIteratorAfterWord(class)
	case breakIteratorInNumberDigits:
		switch class {
		case breakIteratorDG:
			return breakIteratorInNumberDigits
		case breakIteratorMN:
			return breakIteratorAfterNumberSeparator | breakIteratorInPostWord
		}
		return breakIteratorAfterWord(class)
	case breakIteratorAfterNumberSeparator:
		if class == breakIteratorDG {
			return breakIteratorInNumberDigits
		}
		return 0
	case breakIteratorInDashes:
		if class == breakIteratorDA {
			return breakIteratorInDashes
		}
		return breakIteratorAfterSuffix(class)
	case breakIteratorInPostWord:
		switch class {
		case breakIteratorPO, breakIteratorMN, breakIteratorQU, breakIteratorPM:
			return breakIteratorInPostWord
		}
		return breakIteratorAfterSuffix(class)
	case breakIteratorInSpaces:
		return breakIteratorAfterUnit(class)
	case breakIteratorInMarks:
		return breakIteratorTerminator(class)
	case breakIteratorAfterOpenParen:
		if r == ')' {
			return breakIteratorAfterParenPair
		}
		return 0
	case breakIteratorAfterParenPair:
		switch class {
		case breakIteratorPO, breakIteratorMN, breakIteratorQU, breakIteratorPM:
			return breakIteratorAfterParenPair
		}
		return breakIteratorTerminator(class)
	case breakIteratorAfterCR:
		if class == breakIteratorBK {
			return breakIteratorAtEnd
		}
		return 0
	}
	return 0
}

// breakIteratorSegmentEnd returns the end of the segment that starts at
// start, which is less than len(text).
func breakIteratorSegmentEnd(text string, start int) int {
	states := breakIteratorAtStart
	pos := start
	end := start
	for pos < len(text) {
		r, width := utf8.DecodeRuneInString(text[pos:])
		class := breakIteratorClassOf(r)
		if class == breakIteratorCF {
			// A format character is skipped: the positions stay what they
			// are, and the character belongs to the segment if the segment
			// can end here.
			pos += width
			if states&^breakIteratorNonAccepting != 0 {
				end = pos
			}
			continue
		}
		var next breakIteratorStates
		for remaining := states; remaining != 0; remaining &= remaining - 1 {
			next |= breakIteratorStep(remaining&-remaining, class, r)
		}
		if next == 0 {
			break
		}
		states = next
		pos += width
		if states&^breakIteratorNonAccepting != 0 {
			end = pos
		}
	}
	if end == start {
		// No rule matches the first character: it is a segment by itself.
		_, width := utf8.DecodeRuneInString(text[start:])
		end = start + width
	}
	return end
}
