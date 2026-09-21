// Tests for layout_whitespace_stripper.go. Flying Saucer has no JUnit test for
// WhitespaceStripper; the expected values were produced by running the Java
// WhitespaceStripper.stripInlineContent over the same inputs.

package ufo

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/octoberswimmer/ufo/dom"
)

// whitespaceStripperTestStyleable is a Styleable that is not an InlineBox: a
// floated block or an inline-block.
type whitespaceStripperTestStyleable struct {
	style CalculatedStyleI
}

func (s *whitespaceStripperTestStyleable) GetStyle() CalculatedStyleI      { return s.style }
func (s *whitespaceStripperTestStyleable) SetStyle(style CalculatedStyleI) { s.style = style }
func (s *whitespaceStripperTestStyleable) GetElement() *dom.Element        { return nil }
func (s *whitespaceStripperTestStyleable) SetElement(e *dom.Element)       {}
func (s *whitespaceStripperTestStyleable) GetPseudoElementOrClass() string { return "" }

// layoutTestUnescape decodes the \uXXXX escapes (UTF-16 code units) used in
// the tables of the layout tests.
func layoutTestUnescape(s string) string {
	var units []uint16
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+5 < len(s) && s[i+1] == 'u' {
			v, err := strconv.ParseUint(s[i+2:i+6], 16, 16)
			if err != nil {
				panic(err)
			}
			units = append(units, uint16(v))
			i += 5
		} else {
			units = append(units, uint16(s[i]))
		}
	}
	return string(utf16.Decode(units))
}

// layoutTestEscape is the inverse of layoutTestUnescape.
func layoutTestEscape(s string) string {
	var sb strings.Builder
	for _, unit := range utf16.Encode([]rune(s)) {
		if unit < 0x20 || unit > 0x7e || unit == '\\' || unit == '|' {
			sb.WriteString("\\u")
			hex := strconv.FormatUint(uint64(unit), 16)
			sb.WriteString(strings.Repeat("0", 4-len(hex)) + hex)
		} else {
			sb.WriteByte(byte(unit))
		}
	}
	return sb.String()
}

func whitespaceStripperTestStyle(display *IdentValue, whiteSpace *IdentValue, tabSize int, float *IdentValue) CalculatedStyleI {
	decls := []*PropertyDeclaration{
		CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, display),
	}
	if whiteSpace != nil {
		decls = append(decls, CascadedStyleCreateLayoutPropertyDeclaration(CSSNameWhiteSpace, whiteSpace))
	}
	if float != nil {
		decls = append(decls, CascadedStyleCreateLayoutPropertyDeclaration(CSSNameFloat, float))
	}
	if tabSize >= 0 {
		decls = append(decls, NewPropertyDeclaration(CSSNameTabSize,
			NewPropertyValueFloat(CSSPrimitiveValueCssNumber, float32(tabSize), strconv.Itoa(tabSize)), true, StylesheetInfoOriginUser))
	}
	return NewEmptyStyle().DeriveStyle(CascadedStyleCreateLayoutStyle(decls...))
}

// TestWhitespaceStripperStripInlineContent: an input item is
// "I:<white-space>:<tab-size>:<a|e>:<text>" for an InlineBox (a = anonymous,
// e = has an element), "F" for a floated block and "B" for an inline-block;
// items are separated by "|". An output item is "I:<r|k>:<text>" (r = marked
// as removable whitespace), "F" or "B".
func TestWhitespaceStripperStripInlineContent(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"I:normal:8:a:  hello   world  ", "I:k: hello world "},
		{"I:normal:8:a:hello |I:normal:8:a: world", "I:k:hello |I:k:world"},
		{"I:normal:8:a:hello |I:normal:8:a:|I:normal:8:a: world", "I:k:hello |I:r:|I:k:world"},
		{"I:normal:8:a:a \\u000a b\\u0009c\\u000a\\u000ad \\u000b\\u000a e", "I:k:a b c d e"},
		{"I:nowrap:8:a:a \\u000a b\\u0009c\\u000a\\u000ad \\u000d\\u000a e ", "I:k:a b c d e "},
		{"I:pre:8:a:a \\u000a b\\u0009c \\u0009\\u000a\\u000ad  \\u000a e ", "I:k:a\\u000a b        c \\u000a\\u000ad \\u000a e "},
		{"I:pre:4:a:\\u0009x\\u0009y", "I:k:    x    y"},
		{"I:pre:0:a:\\u0009x\\u0009y", "I:k:xy"},
		{"I:pre-wrap:3:a:a \\u000a b\\u0009c\\u000a\\u000ad \\u000a e ", "I:k:a \\u000a b   c\\u000a\\u000ad \\u000a e "},
		{"I:pre-line:8:a:a \\u000a b\\u0009c\\u000a\\u000ad   \\u000a e  ", "I:k:a \\u000a b c\\u000a\\u000ad \\u000a e "},
		{"I:normal:8:a:   ", ""},
		{"I:normal:8:e:   ", "I:r:"},
		{"I:normal:8:a: \\u000a |I:normal:8:a:\\u0009", ""},
		{"I:normal:8:a: |I:normal:8:e: ", "I:r:|I:r:"},
		{"I:pre:8:a:   ", "I:k:   "},
		{"I:pre-wrap:8:a:   ", ""},
		{"I:pre-wrap:8:a: \\u000a ", "I:k: \\u000a "},
		{"I:pre-line:8:a:   ", ""},
		{"I:pre-line:8:a: \\u000a ", "I:k: \\u000a "},
		{"I:normal:8:a: |F|I:normal:8:a: ", "F"},
		{"I:normal:8:a: |B|I:normal:8:a: ", "I:r: |B|I:r: "},
		{"I:normal:8:a:x |F|I:normal:8:a: y", "I:k:x |F|I:k:y"},
		{"I:normal:8:a:x |B|I:normal:8:a: y", "I:k:x |B|I:k: y"},
		{"F", "F"},
		{"B", "B"},
		{"I:normal:8:a:", ""},
		{"I:normal:8:a:x |I:pre:8:a: y |I:normal:8:a: z", "I:k:x |I:k: y |I:k:z"},
		{"I:pre:8:a:x |I:normal:8:a: y", "I:k:x |I:k:y"},
		{"I:pre-wrap:8:a:x |I:normal:8:a: y", "I:k:x |I:k: y"},
		{"I:pre-line:8:a:x |I:normal:8:a: y", "I:k:x |I:k: y"},
		{"I:normal:8:a:\\u00a0 \\u00a0", "I:k:\\u00a0 \\u00a0"},
		{"I:normal:8:a:\\u3000 \\u2003 ", "I:k:\\u3000 \\u2003 "},
		{"I:normal:8:a:caf\\u00e9 \\u000a \\u4e2d\\u6587  \\u1f600  x", "I:k:caf\\u00e9 \\u4e2d\\u6587 \\u1f600 x"},
		{"I:normal:8:a:\\u000c\\u000a\\u000c|I:normal:8:a:\\u000d", ""},
		{"I:normal:8:a:\\u0001 ", ""},
		{"I:normal:8:a:a\\u000ab|I:normal:8:a:\\u000ac", "I:k:a b|I:k: c"},
		{"I:normal:8:a:\\u000a a", "I:k: a"},
		{"I:normal:8:a:a \\u000a", "I:k:a "},
		{"I:pre:8:e:\\u000c|I:pre-wrap:2:e:|F", "I:k:\\u000c|I:r:|F"},
		{"I:normal:8:e:\\u0009", "I:r:"},
		{"I:normal:8:a:a\\u000d", "I:k:a\\u000d"},
		{"I:pre-line:8:a:\\u000a\\u0009  |I:pre-wrap:8:a:\\u000b\\u00e9\\u000bba|I:normal:2:e: \\u0009", "I:k:\\u000a |I:k:\\u000b\\u00e9\\u000bba|I:r: "},
		{"F", "F"},
		{"B|I:pre:4:e:\\u000c|I:pre-wrap:2:e:  a\\u000b  a|I:pre-line:2:a:\\u0009\\u000a ", "B|I:k:\\u000c|I:k:  a\\u000b  a|I:k: \\u000a "},
		{"I:normal:4:a:\\u000d\\u000b \\u000a\\u000b\\u000d|I:pre-wrap:8:a:\\u0009 \\u000a", "I:r: |I:k:         \\u000a"},
		{"B|I:normal:4:a:\\u000aaa \\u000a\\u000d\\u000cb\\u00e9", "B|I:k: aa b\\u00e9"},
		{"I:pre-line:8:a:\\u000d\\u000d\\u000d \\u000b\\u000d", ""},
		{"I:pre:8:a:  ", "I:k:  "},
		{"I:pre-wrap:8:e: ", "I:r:"},
		{"I:pre:4:a:   |B|I:pre-line:8:e: ", "I:k:   |B|I:r: "},
		{"I:normal:8:e: a\\u000cb|I:pre-wrap:2:e:b\\u0009\\u00e9\\u0009\\u0009\\u000d\\u0009\\u0009", "I:k: a\\u000cb|I:k:b  \\u00e9    \\u000d    "},
		{"I:nowrap:4:e:\\u00e9b\\u000b", "I:k:\\u00e9b\\u000b"},
		{"I:normal:2:e:\\u0009\\u000b\\u00e9\\u00e9 ", "I:k: \\u000b\\u00e9\\u00e9 "},
		{"I:normal:4:a:\\u000b\\u000a\\u000d", ""},
		{"I:pre-line:8:a:\\u000a\\u000a|I:pre:2:e:\\u00e9\\u00e9|B|I:normal:8:e:", "I:k:\\u000a\\u000a|I:k:\\u00e9\\u00e9|B|I:r:"},
		{"I:normal:8:e:\\u0009|I:pre:4:e:", "I:r: |I:k:"},
		{"I:pre-wrap:2:a: |I:normal:2:a:  \\u0009|I:normal:2:a:  |I:normal:8:e:\\u000c\\u00e9\\u000c\\u0009a\\u000b\\u000c\\u000c\\u000b", "I:r: |I:r: |I:r:|I:k:\\u000c\\u00e9\\u000c a\\u000b\\u000c\\u000c\\u000b"},
		{"I:normal:8:e:\\u000bb \\u0009\\u000d |I:normal:8:e:\\u000aa\\u000a\\u000b\\u0009", "I:k:\\u000bb \\u000d |I:k:a \\u000b "},
		{"B|I:normal:8:e: \\u0009 |I:pre-line:4:e:\\u000c\\u000b\\u000b \\u000d|I:normal:8:e: ", "B|I:r: |I:r:\\u000c\\u000b\\u000b \\u000d|I:r: "},
		{"F|I:nowrap:2:e:a\\u000d\\u000a\\u000c\\u000c\\u00e9|F", "F|I:k:a \\u00e9|F"},
		{"I:normal:8:a:a", "I:k:a"},
		{"F|I:normal:2:a:|I:pre-wrap:8:a:\\u00e9\\u000a \\u000c", "F|I:r:|I:k:\\u00e9\\u000a \\u000c"},
		{"I:normal:4:e:\\u0009\\u000a|I:normal:4:a:", "I:r:|I:r:"},
		{"F", "F"},
		{"I:normal:2:e:\\u000d|I:nowrap:2:e:b\\u0009\\u000a", "I:r:\\u000d|I:k:b "},
		{"I:normal:8:a: \\u000a", ""},
		{"I:pre-line:4:e:\\u0009a \\u000b\\u000a\\u000aa\\u000b |I:pre-wrap:4:a:|I:normal:2:e:|I:normal:8:e:", "I:k: a \\u000b\\u000a\\u000aa\\u000b |I:r:|I:r:|I:r:"},
		{"B", "B"},
		{"I:nowrap:8:e: ", "I:r:"},
		{"I:pre:8:a:\\u000a  \\u000a|F|F", "I:k:\\u000a \\u000a|F|F"},
		{"I:pre:8:a:  \\u000c\\u0009\\u000ba \\u000b|I:normal:2:e:a a\\u0009\\u0009\\u0009\\u000b|F", "I:k:  \\u000c        \\u000ba \\u000b|I:k:a a \\u000b|F"},
		{"I:pre-wrap:4:e: \\u00e9\\u000a|I:pre-line:2:e:\\u000a   |I:normal:8:e:a\\u000ca\\u000b\\u000b\\u000b ", "I:k: \\u00e9\\u000a|I:k:\\u000a |I:k:a\\u000ca\\u000b\\u000b\\u000b "},
		{"I:normal:2:e:", "I:r:"},
		{"I:normal:8:e:|I:normal:8:a:a b\\u0009\\u000b\\u000b\\u000d |I:pre-line:4:a:\\u000a\\u000db\\u000d|I:normal:8:a:  ", "I:r:|I:k:a b \\u000b\\u000b\\u000d |I:k:\\u000a\\u000db\\u000d|I:r: "},
		{"I:normal:8:a: b\\u000da a  a|I:pre:2:a:b\\u000d |I:normal:8:a:\\u000b\\u00e9\\u000aa\\u000b ", "I:k: b\\u000da a a|I:k:b\\u000d |I:k:\\u000b\\u00e9 a\\u000b "},
		{"I:nowrap:4:e: \\u000a|I:pre:8:e: |I:pre:4:a:\\u000d\\u000a\\u000c\\u0009\\u0009 \\u000a|I:nowrap:2:a:\\u000a   ", "I:r: |I:k: |I:k:\\u000a\\u000c        \\u000a|I:r: "},
		{"I:normal:2:e:b\\u000a\\u000c\\u000c\\u0009 a\\u0009\\u000d|I:nowrap:8:a:|I:pre:8:a:  \\u000d\\u000c\\u000b\\u000b\\u0009|B", "I:k:b \\u000c\\u000c a \\u000d|I:r:|I:k:  \\u000d\\u000c\\u000b\\u000b        |B"},
		{"I:pre-line:8:e: \\u000c  \\u000a\\u0009\\u00e9", "I:k: \\u000c \\u000a \\u00e9"},
		{"I:pre-wrap:8:e: |I:pre-wrap:8:e:\\u000a ", "I:r: |I:k:\\u000a "},
		{"I:nowrap:8:a:\\u000b\\u000c\\u0009|I:pre-line:8:a:|I:pre-line:8:e:a|I:normal:4:a:\\u000db\\u000d\\u0009 ", "I:r:\\u000b\\u000c |I:r:|I:k:a|I:k:\\u000db\\u000d "},
		{"I:normal:2:a:\\u000a|I:nowrap:2:a: \\u000a\\u000a ", ""},
		{"I:normal:8:e:\\u000a   |I:pre-line:8:e: ", "I:r:|I:r:"},
		{"B|I:pre-line:8:a:\\u000dbb\\u000b", "B|I:k:\\u000dbb\\u000b"},
		{"F", "F"},
		{"I:normal:2:a:\\u000dba", "I:k:\\u000dba"},
		{"I:normal:2:a:\\u0009bb\\u000b \\u000d\\u0009", "I:k: bb\\u000b \\u000d "},
		{"F|F|I:nowrap:4:e:|I:pre-wrap:8:a:\\u0009\\u0009", "F|F|I:r:|I:r:"},
		{"I:normal:8:e:\\u000da\\u000d\\u000b\\u000a\\u000b\\u000a", "I:k:\\u000da "},
		{"I:nowrap:8:e:b\\u00e9 \\u000c\\u0009\\u000d\\u000a|F", "I:k:b\\u00e9 \\u000c \\u000d |F"},
		{"I:pre-wrap:8:a:  \\u0009", ""},
		{"I:pre-line:2:e:\\u0009\\u000a", "I:k: \\u000a"},
		{"I:pre-line:4:e:aa\\u00e9a|I:normal:8:e:\\u0009\\u0009", "I:k:aa\\u00e9a|I:r: "},
		{"I:pre:8:e:\\u000a|I:normal:4:a: ", "I:k:\\u000a|I:r: "},
		{"I:normal:2:e:\\u000a \\u0009\\u000a|I:pre-line:4:a:\\u00e9|F", "I:r: |I:k:\\u00e9|F"},
		{"B|I:nowrap:4:e: |I:normal:8:e:\\u0009", "B|I:r: |I:r:"},
		{"F|F|I:pre-wrap:2:e: \\u000a \\u0009|I:pre:8:e:\\u00e9b\\u000d\\u000d", "F|F|I:k: \\u000a   |I:k:\\u00e9b\\u000d\\u000d"},
		{"I:pre:2:a:|I:normal:8:e:\\u000b\\u000a\\u000a  ", "I:k:|I:r: "},
		{"I:nowrap:4:a:\\u000ab", "I:k: b"},
	}

	doc := dom.NewDocument()
	for _, tt := range tests {
		var content []Styleable
		for _, item := range strings.Split(tt.in, "|") {
			p := strings.SplitN(item, ":", 5)
			switch p[0] {
			case "I":
				tabSize, err := strconv.Atoi(p[2])
				if err != nil {
					t.Fatal(err)
				}
				style := whitespaceStripperTestStyle(IdentValueInline, IdentValueValueOf(p[1]), tabSize, nil)
				var element *dom.Element
				if p[3] == "e" {
					element = doc.CreateElement("span")
				}
				content = append(content, NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClassStyle(
					layoutTestUnescape(p[4]), nil, nil, nil, element, "", style))
			case "F":
				content = append(content, &whitespaceStripperTestStyleable{whitespaceStripperTestStyle(IdentValueBlock, nil, -1, IdentValueLeft)})
			default:
				content = append(content, &whitespaceStripperTestStyleable{whitespaceStripperTestStyle(IdentValueInlineBlock, nil, -1, nil)})
			}
		}

		content = WhitespaceStripperStripInlineContent(content)

		var got []string
		for _, s := range content {
			if iB, ok := s.(*InlineBox); ok {
				removable := "k"
				if iB.IsRemovableWhitespace() {
					removable = "r"
				}
				got = append(got, "I:"+removable+":"+layoutTestEscape(iB.GetText()))
			} else if s.GetStyle().IsFloated() {
				got = append(got, "F")
			} else {
				got = append(got, "B")
			}
		}
		if strings.Join(got, "|") != tt.want {
			t.Errorf("%q: got %q, want %q", tt.in, strings.Join(got, "|"), tt.want)
		}
	}
}
