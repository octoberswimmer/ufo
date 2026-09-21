// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/layout/TextUtilTest.java
//
// The tests after the JUnit ones compare against values produced by running
// the Java TextUtil (JDK 25) over the same inputs.

package ufo

import (
	"testing"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo/geom"
)

func TestTextUtilCapitalizesWords(t *testing.T) {
	if got := LayoutTextUtilTransformTextWithTransformFontVariant("hellO worlD!", IdentValueCapitalize, IdentValueShow); got != "HellO WorlD!" {
		t.Errorf("got %q", got)
	}
}

func TestTextUtilToLowerCase(t *testing.T) {
	if got := LayoutTextUtilTransformTextWithTransformFontVariant("hellO worlD!", IdentValueLowercase, IdentValueShow); got != "hello world!" {
		t.Errorf("got %q", got)
	}
}

func TestTextUtilToUpperCase(t *testing.T) {
	if got := LayoutTextUtilTransformTextWithTransformFontVariant("hellO worlD!", IdentValueUppercase, IdentValueShow); got != "HELLO WORLD!" {
		t.Errorf("got %q", got)
	}
}

func TestTextUtilSmallCaps(t *testing.T) {
	if got := LayoutTextUtilTransformTextWithTransformFontVariant("hellO worlD!", IdentValueShow, IdentValueSmallCaps); got != "HELLO WORLD!" {
		t.Errorf("got %q", got)
	}
}

func TestTextUtilReplace(t *testing.T) {
	tests := []struct{ source, target, replacement, want string }{
		{"", "a", "b", ""},
		{"Hello, World", "Hello", "Goodbye", "Goodbye, World"},
		{"abababab", "ab", "c", "cccc"},
		{"Foo Zoo", "o", "oo", "Foooo Zoooo"},
	}
	for _, tt := range tests {
		if got := UtilReplace(tt.source, tt.target, tt.replacement); got != tt.want {
			t.Errorf("UtilReplace(%q, %q, %q) = %q, want %q", tt.source, tt.target, tt.replacement, got, tt.want)
		}
	}
}

// TestLayoutTextUtilTransformTextAgainstJava: text, want and wantFirstLetter
// use the escapes of layoutTestUnescape. want is transformText(text, style)
// and wantFirstLetter is transformFirstLetterText(text, style) for a style
// with the given text-transform and font-variant.
func TestLayoutTextUtilTransformTextAgainstJava(t *testing.T) {
	tests := []struct {
		text            string
		transform       string
		fontVariant     string
		want            string
		wantFirstLetter string
	}{
		{"hellO worlD!", "none", "normal", "hellO worlD!", "hellO worlD!"},
		{"hellO worlD!", "none", "small-caps", "HELLO WORLD!", "HellO worlD!"},
		{"hellO worlD!", "lowercase", "normal", "hello world!", "hellO worlD!"},
		{"hellO worlD!", "lowercase", "small-caps", "HELLO WORLD!", "hellO worlD!"},
		{"hellO worlD!", "uppercase", "normal", "HELLO WORLD!", "HellO worlD!"},
		{"hellO worlD!", "uppercase", "small-caps", "HELLO WORLD!", "HellO worlD!"},
		{"hellO worlD!", "capitalize", "normal", "HellO WorlD!", "HellO worlD!"},
		{"hellO worlD!", "capitalize", "small-caps", "HELLO WORLD!", "HellO worlD!"},
		{"", "none", "normal", "", ""},
		{"", "none", "small-caps", "", ""},
		{"", "lowercase", "normal", "", ""},
		{"", "lowercase", "small-caps", "", ""},
		{"", "uppercase", "normal", "", ""},
		{"", "uppercase", "small-caps", "", ""},
		{"", "capitalize", "normal", "", ""},
		{"", "capitalize", "small-caps", "", ""},
		{" ", "none", "normal", " ", " "},
		{" ", "none", "small-caps", " ", " "},
		{" ", "lowercase", "normal", " ", " "},
		{" ", "lowercase", "small-caps", " ", " "},
		{" ", "uppercase", "normal", " ", " "},
		{" ", "uppercase", "small-caps", " ", " "},
		{" ", "capitalize", "normal", " ", " "},
		{" ", "capitalize", "small-caps", " ", " "},
		{"a", "none", "normal", "a", "a"},
		{"a", "none", "small-caps", "A", "A"},
		{"a", "lowercase", "normal", "a", "a"},
		{"a", "lowercase", "small-caps", "A", "a"},
		{"a", "uppercase", "normal", "A", "A"},
		{"a", "uppercase", "small-caps", "A", "A"},
		{"a", "capitalize", "normal", "A", "A"},
		{"a", "capitalize", "small-caps", "A", "A"},
		{"hello  world", "none", "normal", "hello  world", "hello  world"},
		{"hello  world", "none", "small-caps", "HELLO  WORLD", "Hello  world"},
		{"hello  world", "lowercase", "normal", "hello  world", "hello  world"},
		{"hello  world", "lowercase", "small-caps", "HELLO  WORLD", "hello  world"},
		{"hello  world", "uppercase", "normal", "HELLO  WORLD", "Hello  world"},
		{"hello  world", "uppercase", "small-caps", "HELLO  WORLD", "Hello  world"},
		{"hello  world", "capitalize", "normal", "Hello  World", "Hello  world"},
		{"hello  world", "capitalize", "small-caps", "HELLO  WORLD", "Hello  world"},
		{" leading space", "none", "normal", " leading space", " leading space"},
		{" leading space", "none", "small-caps", " LEADING SPACE", " Leading space"},
		{" leading space", "lowercase", "normal", " leading space", " leading space"},
		{" leading space", "lowercase", "small-caps", " LEADING SPACE", " leading space"},
		{" leading space", "uppercase", "normal", " LEADING SPACE", " Leading space"},
		{" leading space", "uppercase", "small-caps", " LEADING SPACE", " Leading space"},
		{" leading space", "capitalize", "normal", " Leading Space", " Leading space"},
		{" leading space", "capitalize", "small-caps", " LEADING SPACE", " Leading space"},
		{"\"quoted\" text", "none", "normal", "\"quoted\" text", "\"quoted\" text"},
		{"\"quoted\" text", "none", "small-caps", "\"QUOTED\" TEXT", "\"Quoted\" text"},
		{"\"quoted\" text", "lowercase", "normal", "\"quoted\" text", "\"quoted\" text"},
		{"\"quoted\" text", "lowercase", "small-caps", "\"QUOTED\" TEXT", "\"quoted\" text"},
		{"\"quoted\" text", "uppercase", "normal", "\"QUOTED\" TEXT", "\"Quoted\" text"},
		{"\"quoted\" text", "uppercase", "small-caps", "\"QUOTED\" TEXT", "\"Quoted\" text"},
		{"\"quoted\" text", "capitalize", "normal", "\"quoted\" Text", "\"Quoted\" text"},
		{"\"quoted\" text", "capitalize", "small-caps", "\"QUOTED\" TEXT", "\"Quoted\" text"},
		{"\\u00abguillemets\\u00bb et cetera", "none", "normal", "\\u00abguillemets\\u00bb et cetera", "\\u00abguillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "none", "small-caps", "\\u00abGUILLEMETS\\u00bb ET CETERA", "\\u00abGuillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "lowercase", "normal", "\\u00abguillemets\\u00bb et cetera", "\\u00abguillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "lowercase", "small-caps", "\\u00abGUILLEMETS\\u00bb ET CETERA", "\\u00abguillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "uppercase", "normal", "\\u00abGUILLEMETS\\u00bb ET CETERA", "\\u00abGuillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "uppercase", "small-caps", "\\u00abGUILLEMETS\\u00bb ET CETERA", "\\u00abGuillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "capitalize", "normal", "\\u00abguillemets\\u00bb Et Cetera", "\\u00abGuillemets\\u00bb et cetera"},
		{"\\u00abguillemets\\u00bb et cetera", "capitalize", "small-caps", "\\u00abGUILLEMETS\\u00bb ET CETERA", "\\u00abGuillemets\\u00bb et cetera"},
		{"(parenthesised) word", "none", "normal", "(parenthesised) word", "(parenthesised) word"},
		{"(parenthesised) word", "none", "small-caps", "(PARENTHESISED) WORD", "(Parenthesised) word"},
		{"(parenthesised) word", "lowercase", "normal", "(parenthesised) word", "(parenthesised) word"},
		{"(parenthesised) word", "lowercase", "small-caps", "(PARENTHESISED) WORD", "(parenthesised) word"},
		{"(parenthesised) word", "uppercase", "normal", "(PARENTHESISED) WORD", "(Parenthesised) word"},
		{"(parenthesised) word", "uppercase", "small-caps", "(PARENTHESISED) WORD", "(Parenthesised) word"},
		{"(parenthesised) word", "capitalize", "normal", "(parenthesised) Word", "(Parenthesised) word"},
		{"(parenthesised) word", "capitalize", "small-caps", "(PARENTHESISED) WORD", "(Parenthesised) word"},
		{"\\u00a0nbsp first", "none", "normal", "\\u00a0nbsp first", "\\u00a0nbsp first"},
		{"\\u00a0nbsp first", "none", "small-caps", "\\u00a0NBSP FIRST", "\\u00a0Nbsp first"},
		{"\\u00a0nbsp first", "lowercase", "normal", "\\u00a0nbsp first", "\\u00a0nbsp first"},
		{"\\u00a0nbsp first", "lowercase", "small-caps", "\\u00a0NBSP FIRST", "\\u00a0nbsp first"},
		{"\\u00a0nbsp first", "uppercase", "normal", "\\u00a0NBSP FIRST", "\\u00a0Nbsp first"},
		{"\\u00a0nbsp first", "uppercase", "small-caps", "\\u00a0NBSP FIRST", "\\u00a0Nbsp first"},
		{"\\u00a0nbsp first", "capitalize", "normal", "\\u00a0nbsp First", "\\u00a0Nbsp first"},
		{"\\u00a0nbsp first", "capitalize", "small-caps", "\\u00a0NBSP FIRST", "\\u00a0Nbsp first"},
		{"\\u2003em space", "none", "normal", "\\u2003em space", "\\u2003em space"},
		{"\\u2003em space", "none", "small-caps", "\\u2003EM SPACE", "\\u2003Em space"},
		{"\\u2003em space", "lowercase", "normal", "\\u2003em space", "\\u2003em space"},
		{"\\u2003em space", "lowercase", "small-caps", "\\u2003EM SPACE", "\\u2003em space"},
		{"\\u2003em space", "uppercase", "normal", "\\u2003EM SPACE", "\\u2003Em space"},
		{"\\u2003em space", "uppercase", "small-caps", "\\u2003EM SPACE", "\\u2003Em space"},
		{"\\u2003em space", "capitalize", "normal", "\\u2003em Space", "\\u2003Em space"},
		{"\\u2003em space", "capitalize", "small-caps", "\\u2003EM SPACE", "\\u2003Em space"},
		{"stra\\u00dfe \\u00dfeta", "none", "normal", "stra\\u00dfe \\u00dfeta", "stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "none", "small-caps", "STRASSE SSETA", "Stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "lowercase", "normal", "stra\\u00dfe \\u00dfeta", "stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "lowercase", "small-caps", "STRASSE SSETA", "stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "uppercase", "normal", "STRASSE SSETA", "Stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "uppercase", "small-caps", "STRASSE SSETA", "Stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "capitalize", "normal", "Stra\\u00dfe SSeta", "Stra\\u00dfe \\u00dfeta"},
		{"stra\\u00dfe \\u00dfeta", "capitalize", "small-caps", "STRASSE SSETA", "Stra\\u00dfe \\u00dfeta"},
		{"\\u00e9cole \\u00e0 paris", "none", "normal", "\\u00e9cole \\u00e0 paris", "\\u00e9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "none", "small-caps", "\\u00c9COLE \\u00c0 PARIS", "\\u00c9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "lowercase", "normal", "\\u00e9cole \\u00e0 paris", "\\u00e9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "lowercase", "small-caps", "\\u00c9COLE \\u00c0 PARIS", "\\u00e9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "uppercase", "normal", "\\u00c9COLE \\u00c0 PARIS", "\\u00c9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "uppercase", "small-caps", "\\u00c9COLE \\u00c0 PARIS", "\\u00c9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "capitalize", "normal", "\\u00c9cole \\u00c0 Paris", "\\u00c9cole \\u00e0 paris"},
		{"\\u00e9cole \\u00e0 paris", "capitalize", "small-caps", "\\u00c9COLE \\u00c0 PARIS", "\\u00c9cole \\u00e0 paris"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "none", "normal", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "none", "small-caps", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "lowercase", "normal", "\\u03bf\\u03b4\\u03c5\\u03c3\\u03c3\\u03b5\\u03c5\\u03c2 \\u03c3", "\\u03bf\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "lowercase", "small-caps", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u03bf\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "uppercase", "normal", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "uppercase", "small-caps", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "capitalize", "normal", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "capitalize", "small-caps", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3", "\\u039f\\u0394\\u03a5\\u03a3\\u03a3\\u0395\\u03a5\\u03a3 \\u03a3"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "none", "normal", "\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "none", "small-caps", "\\u03a3\\u039f\\u03a6\\u038c\\u03a3 \\u03a3", "\\u03a3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "lowercase", "normal", "\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "lowercase", "small-caps", "\\u03a3\\u039f\\u03a6\\u038c\\u03a3 \\u03a3", "\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "uppercase", "normal", "\\u03a3\\u039f\\u03a6\\u038c\\u03a3 \\u03a3", "\\u03a3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "uppercase", "small-caps", "\\u03a3\\u039f\\u03a6\\u038c\\u03a3 \\u03a3", "\\u03a3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "capitalize", "normal", "\\u03a3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03a3", "\\u03a3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\u03c3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2", "capitalize", "small-caps", "\\u03a3\\u039f\\u03a6\\u038c\\u03a3 \\u03a3", "\\u03a3\\u03bf\\u03c6\\u03cc\\u03c2 \\u03c2"},
		{"\\ufb01ne \\ufb02our", "none", "normal", "\\ufb01ne \\ufb02our", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "none", "small-caps", "FINE FLOUR", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "lowercase", "normal", "\\ufb01ne \\ufb02our", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "lowercase", "small-caps", "FINE FLOUR", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "uppercase", "normal", "FINE FLOUR", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "uppercase", "small-caps", "FINE FLOUR", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "capitalize", "normal", "FIne FLour", "\\ufb01ne \\ufb02our"},
		{"\\ufb01ne \\ufb02our", "capitalize", "small-caps", "FINE FLOUR", "\\ufb01ne \\ufb02our"},
		{"\\u0149 \\u01f0", "none", "normal", "\\u0149 \\u01f0", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "none", "small-caps", "\\u02bcN J\\u030c", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "lowercase", "normal", "\\u0149 \\u01f0", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "lowercase", "small-caps", "\\u02bcN J\\u030c", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "uppercase", "normal", "\\u02bcN J\\u030c", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "uppercase", "small-caps", "\\u02bcN J\\u030c", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "capitalize", "normal", "\\u02bcN J\\u030c", "\\u0149 \\u01f0"},
		{"\\u0149 \\u01f0", "capitalize", "small-caps", "\\u02bcN J\\u030c", "\\u0149 \\u01f0"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "none", "normal", "i\\u0307stanbul \\u0130stanbul \\u0131", "i\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "none", "small-caps", "I\\u0307STANBUL \\u0130STANBUL I", "I\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "lowercase", "normal", "i\\u0307stanbul i\\u0307stanbul \\u0131", "i\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "lowercase", "small-caps", "I\\u0307STANBUL I\\u0307STANBUL I", "i\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "uppercase", "normal", "I\\u0307STANBUL \\u0130STANBUL I", "I\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "uppercase", "small-caps", "I\\u0307STANBUL \\u0130STANBUL I", "I\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "capitalize", "normal", "I\\u0307stanbul \\u0130stanbul I", "I\\u0307stanbul \\u0130stanbul \\u0131"},
		{"i\\u0307stanbul \\u0130stanbul \\u0131", "capitalize", "small-caps", "I\\u0307STANBUL \\u0130STANBUL I", "I\\u0307stanbul \\u0130stanbul \\u0131"},
		{"\\u01c6 \\u01c5 \\u01c4", "none", "normal", "\\u01c6 \\u01c5 \\u01c4", "\\u01c6 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "none", "small-caps", "\\u01c4 \\u01c4 \\u01c4", "\\u01c4 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "lowercase", "normal", "\\u01c6 \\u01c6 \\u01c6", "\\u01c6 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "lowercase", "small-caps", "\\u01c4 \\u01c4 \\u01c4", "\\u01c6 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "uppercase", "normal", "\\u01c4 \\u01c4 \\u01c4", "\\u01c4 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "uppercase", "small-caps", "\\u01c4 \\u01c4 \\u01c4", "\\u01c4 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "capitalize", "normal", "\\u01c4 \\u01c4 \\u01c4", "\\u01c4 \\u01c5 \\u01c4"},
		{"\\u01c6 \\u01c5 \\u01c4", "capitalize", "small-caps", "\\u01c4 \\u01c4 \\u01c4", "\\u01c4 \\u01c5 \\u01c4"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "none", "normal", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "none", "small-caps", "\\ud801\\udc00\\ud801\\udc01 \\ud801\\udc00X", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "lowercase", "normal", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "lowercase", "small-caps", "\\ud801\\udc00\\ud801\\udc01 \\ud801\\udc00X", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "uppercase", "normal", "\\ud801\\udc00\\ud801\\udc01 \\ud801\\udc00X", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "uppercase", "small-caps", "\\ud801\\udc00\\ud801\\udc01 \\ud801\\udc00X", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "capitalize", "normal", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x", "capitalize", "small-caps", "\\ud801\\udc00\\ud801\\udc01 \\ud801\\udc00X", "\\ud801\\udc28\\ud801\\udc29 \\ud801\\udc28x"},
		{"\\u4e2d\\u6587 text", "none", "normal", "\\u4e2d\\u6587 text", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "none", "small-caps", "\\u4e2d\\u6587 TEXT", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "lowercase", "normal", "\\u4e2d\\u6587 text", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "lowercase", "small-caps", "\\u4e2d\\u6587 TEXT", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "uppercase", "normal", "\\u4e2d\\u6587 TEXT", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "uppercase", "small-caps", "\\u4e2d\\u6587 TEXT", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "capitalize", "normal", "\\u4e2d\\u6587 Text", "\\u4e2d\\u6587 text"},
		{"\\u4e2d\\u6587 text", "capitalize", "small-caps", "\\u4e2d\\u6587 TEXT", "\\u4e2d\\u6587 text"},
		{"tab\\u0009sep\\u000anew line", "none", "normal", "tab\\u0009sep\\u000anew line", "tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "none", "small-caps", "TAB\\u0009SEP\\u000aNEW LINE", "Tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "lowercase", "normal", "tab\\u0009sep\\u000anew line", "tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "lowercase", "small-caps", "TAB\\u0009SEP\\u000aNEW LINE", "tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "uppercase", "normal", "TAB\\u0009SEP\\u000aNEW LINE", "Tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "uppercase", "small-caps", "TAB\\u0009SEP\\u000aNEW LINE", "Tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "capitalize", "normal", "Tab\\u0009sep\\u000anew Line", "Tab\\u0009sep\\u000anew line"},
		{"tab\\u0009sep\\u000anew line", "capitalize", "small-caps", "TAB\\u0009SEP\\u000aNEW LINE", "Tab\\u0009sep\\u000anew line"},
		{"---dash first", "none", "normal", "---dash first", "---dash first"},
		{"---dash first", "none", "small-caps", "---DASH FIRST", "---dash first"},
		{"---dash first", "lowercase", "normal", "---dash first", "---dash first"},
		{"---dash first", "lowercase", "small-caps", "---DASH FIRST", "---dash first"},
		{"---dash first", "uppercase", "normal", "---DASH FIRST", "---dash first"},
		{"---dash first", "uppercase", "small-caps", "---DASH FIRST", "---dash first"},
		{"---dash first", "capitalize", "normal", "---dash First", "---dash first"},
		{"---dash first", "capitalize", "small-caps", "---DASH FIRST", "---dash first"},
		{"\\u00bfque? \\u00a1hola!", "none", "normal", "\\u00bfque? \\u00a1hola!", "\\u00bfque? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "none", "small-caps", "\\u00bfQUE? \\u00a1HOLA!", "\\u00bfQue? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "lowercase", "normal", "\\u00bfque? \\u00a1hola!", "\\u00bfque? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "lowercase", "small-caps", "\\u00bfQUE? \\u00a1HOLA!", "\\u00bfque? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "uppercase", "normal", "\\u00bfQUE? \\u00a1HOLA!", "\\u00bfQue? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "uppercase", "small-caps", "\\u00bfQUE? \\u00a1HOLA!", "\\u00bfQue? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "capitalize", "normal", "\\u00bfque? \\u00a1hola!", "\\u00bfQue? \\u00a1hola!"},
		{"\\u00bfque? \\u00a1hola!", "capitalize", "small-caps", "\\u00bfQUE? \\u00a1HOLA!", "\\u00bfQue? \\u00a1hola!"},
		{"1st 2nd", "none", "normal", "1st 2nd", "1st 2nd"},
		{"1st 2nd", "none", "small-caps", "1ST 2ND", "1st 2nd"},
		{"1st 2nd", "lowercase", "normal", "1st 2nd", "1st 2nd"},
		{"1st 2nd", "lowercase", "small-caps", "1ST 2ND", "1st 2nd"},
		{"1st 2nd", "uppercase", "normal", "1ST 2ND", "1st 2nd"},
		{"1st 2nd", "uppercase", "small-caps", "1ST 2ND", "1st 2nd"},
		{"1st 2nd", "capitalize", "normal", "1st 2nd", "1st 2nd"},
		{"1st 2nd", "capitalize", "small-caps", "1ST 2ND", "1st 2nd"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "none", "normal", "\\u2018single\\u2019 \\u201cdouble\\u201d", "\\u2018single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "none", "small-caps", "\\u2018SINGLE\\u2019 \\u201cDOUBLE\\u201d", "\\u2018Single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "lowercase", "normal", "\\u2018single\\u2019 \\u201cdouble\\u201d", "\\u2018single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "lowercase", "small-caps", "\\u2018SINGLE\\u2019 \\u201cDOUBLE\\u201d", "\\u2018single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "uppercase", "normal", "\\u2018SINGLE\\u2019 \\u201cDOUBLE\\u201d", "\\u2018Single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "uppercase", "small-caps", "\\u2018SINGLE\\u2019 \\u201cDOUBLE\\u201d", "\\u2018Single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "capitalize", "normal", "\\u2018single\\u2019 \\u201cdouble\\u201d", "\\u2018Single\\u2019 \\u201cdouble\\u201d"},
		{"\\u2018single\\u2019 \\u201cdouble\\u201d", "capitalize", "small-caps", "\\u2018SINGLE\\u2019 \\u201cDOUBLE\\u201d", "\\u2018Single\\u2019 \\u201cdouble\\u201d"},
		{"\\u0345\\u03b1\\u0345", "none", "normal", "\\u0345\\u03b1\\u0345", "\\u0345\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "none", "small-caps", "\\u0399\\u0391\\u0399", "\\u0399\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "lowercase", "normal", "\\u0345\\u03b1\\u0345", "\\u0345\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "lowercase", "small-caps", "\\u0399\\u0391\\u0399", "\\u0345\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "uppercase", "normal", "\\u0399\\u0391\\u0399", "\\u0399\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "uppercase", "small-caps", "\\u0399\\u0391\\u0399", "\\u0399\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "capitalize", "normal", "\\u0399\\u03b1\\u0345", "\\u0399\\u03b1\\u0345"},
		{"\\u0345\\u03b1\\u0345", "capitalize", "small-caps", "\\u0399\\u0391\\u0399", "\\u0399\\u03b1\\u0345"},
		{"\\u1e9e \\u1e9eb", "none", "normal", "\\u1e9e \\u1e9eb", "\\u1e9e \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "none", "small-caps", "\\u1e9e \\u1e9eB", "\\u1e9e \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "lowercase", "normal", "\\u00df \\u00dfb", "\\u00df \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "lowercase", "small-caps", "SS SSB", "\\u00df \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "uppercase", "normal", "\\u1e9e \\u1e9eB", "\\u1e9e \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "uppercase", "small-caps", "\\u1e9e \\u1e9eB", "\\u1e9e \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "capitalize", "normal", "\\u1e9e \\u1e9eb", "\\u1e9e \\u1e9eb"},
		{"\\u1e9e \\u1e9eb", "capitalize", "small-caps", "\\u1e9e \\u1e9eB", "\\u1e9e \\u1e9eb"},
	}
	for _, tt := range tests {
		text := layoutTestUnescape(tt.text)
		style := NewEmptyStyle().DeriveStyle(CascadedStyleCreateLayoutStyle(
			CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueInline),
			CascadedStyleCreateLayoutPropertyDeclaration(CSSNameTextTransform, IdentValueValueOf(tt.transform)),
			CascadedStyleCreateLayoutPropertyDeclaration(CSSNameFontVariant, IdentValueValueOf(tt.fontVariant)),
		))
		if got := layoutTestEscape(LayoutTextUtilTransformText(text, style)); got != tt.want {
			t.Errorf("TransformText(%q, %s, %s) = %q, want %q", tt.text, tt.transform, tt.fontVariant, got, tt.want)
		}
		if got := layoutTestEscape(LayoutTextUtilTransformFirstLetterText(text, style)); got != tt.wantFirstLetter {
			t.Errorf("TransformFirstLetterText(%q, %s, %s) = %q, want %q", tt.text, tt.transform, tt.fontVariant, got, tt.wantFirstLetter)
		}
	}
}

func TestLayoutTextUtilReplaceChar(t *testing.T) {
	tests := []struct {
		text    string
		newChar rune
		index   int
		want    string
	}{
		{"hello", 'X', 0, "Xello"},
		{"hello", 'X', 4, "hellX"},
		{"hello", 'X', 5, "hello"},
		{"hello", 'X', -1, "hello"},
		{"", 'X', 0, ""},
		// index is a byte offset: the two-byte character at offset 1 is replaced whole
		{"h\u00e9llo", 'E', 1, "hEllo"},
		{"h\u00e9llo", '\u00c9', 3, "h\u00e9\u00c9lo"},
	}
	for _, tt := range tests {
		if got := LayoutTextUtilReplaceChar(tt.text, tt.newChar, tt.index); got != tt.want {
			t.Errorf("ReplaceChar(%q, %q, %d) = %q, want %q", tt.text, tt.newChar, tt.index, got, tt.want)
		}
	}
}

// layoutTextUtilTestSeparatorRanges are the BMP characters for which the Java
// TextUtil.isFirstLetterSeparatorChar returns true, as inclusive ranges.
var layoutTextUtilTestSeparatorRanges = [][2]rune{
	{0x0020, 0x0023}, {0x0025, 0x002a}, {0x002c, 0x002c}, {0x002e, 0x002f}, {0x003a, 0x003b}, {0x003f, 0x0040},
	{0x005b, 0x005d}, {0x007b, 0x007b}, {0x007d, 0x007d}, {0x00a0, 0x00a1}, {0x00a7, 0x00a7}, {0x00ab, 0x00ab},
	{0x00b6, 0x00b7}, {0x00bb, 0x00bb}, {0x00bf, 0x00bf}, {0x037e, 0x037e}, {0x0387, 0x0387}, {0x055a, 0x055f},
	{0x0589, 0x0589}, {0x05c0, 0x05c0}, {0x05c3, 0x05c3}, {0x05c6, 0x05c6}, {0x05f3, 0x05f4}, {0x0609, 0x060a},
	{0x060c, 0x060d}, {0x061b, 0x061b}, {0x061d, 0x061f}, {0x066a, 0x066d}, {0x06d4, 0x06d4}, {0x0700, 0x070d},
	{0x07f7, 0x07f9}, {0x0830, 0x083e}, {0x085e, 0x085e}, {0x0964, 0x0965}, {0x0970, 0x0970}, {0x09fd, 0x09fd},
	{0x0a76, 0x0a76}, {0x0af0, 0x0af0}, {0x0c77, 0x0c77}, {0x0c84, 0x0c84}, {0x0df4, 0x0df4}, {0x0e4f, 0x0e4f},
	{0x0e5a, 0x0e5b}, {0x0f04, 0x0f12}, {0x0f14, 0x0f14}, {0x0f3a, 0x0f3d}, {0x0f85, 0x0f85}, {0x0fd0, 0x0fd4},
	{0x0fd9, 0x0fda}, {0x104a, 0x104f}, {0x10fb, 0x10fb}, {0x1360, 0x1368}, {0x166e, 0x166e}, {0x1680, 0x1680},
	{0x169b, 0x169c}, {0x16eb, 0x16ed}, {0x1735, 0x1736}, {0x17d4, 0x17d6}, {0x17d8, 0x17da}, {0x1800, 0x1805},
	{0x1807, 0x180a}, {0x1944, 0x1945}, {0x1a1e, 0x1a1f}, {0x1aa0, 0x1aa6}, {0x1aa8, 0x1aad}, {0x1b4e, 0x1b4f},
	{0x1b5a, 0x1b60}, {0x1b7d, 0x1b7f}, {0x1bfc, 0x1bff}, {0x1c3b, 0x1c3f}, {0x1c7e, 0x1c7f}, {0x1cc0, 0x1cc7},
	{0x1cd3, 0x1cd3}, {0x2000, 0x200a}, {0x2016, 0x2027}, {0x202f, 0x203e}, {0x2041, 0x2043}, {0x2045, 0x2051},
	{0x2053, 0x2053}, {0x2055, 0x205f}, {0x207d, 0x207e}, {0x208d, 0x208e}, {0x2308, 0x230b}, {0x2329, 0x232a},
	{0x2768, 0x2775}, {0x27c5, 0x27c6}, {0x27e6, 0x27ef}, {0x2983, 0x2998}, {0x29d8, 0x29db}, {0x29fc, 0x29fd},
	{0x2cf9, 0x2cfc}, {0x2cfe, 0x2cff}, {0x2d70, 0x2d70}, {0x2e00, 0x2e16}, {0x2e18, 0x2e19}, {0x2e1b, 0x2e2e},
	{0x2e30, 0x2e39}, {0x2e3c, 0x2e3f}, {0x2e41, 0x2e4f}, {0x2e52, 0x2e5c}, {0x3000, 0x3003}, {0x3008, 0x3011},
	{0x3014, 0x301b}, {0x301d, 0x301f}, {0x303d, 0x303d}, {0x30fb, 0x30fb}, {0xa4fe, 0xa4ff}, {0xa60d, 0xa60f},
	{0xa673, 0xa673}, {0xa67e, 0xa67e}, {0xa6f2, 0xa6f7}, {0xa874, 0xa877}, {0xa8ce, 0xa8cf}, {0xa8f8, 0xa8fa},
	{0xa8fc, 0xa8fc}, {0xa92e, 0xa92f}, {0xa95f, 0xa95f}, {0xa9c1, 0xa9cd}, {0xa9de, 0xa9df}, {0xaa5c, 0xaa5f},
	{0xaade, 0xaadf}, {0xaaf0, 0xaaf1}, {0xabeb, 0xabeb}, {0xfd3e, 0xfd3f}, {0xfe10, 0xfe19}, {0xfe30, 0xfe30},
	{0xfe35, 0xfe4c}, {0xfe50, 0xfe52}, {0xfe54, 0xfe57}, {0xfe59, 0xfe61}, {0xfe68, 0xfe68}, {0xfe6a, 0xfe6b},
	{0xff01, 0xff03}, {0xff05, 0xff0a}, {0xff0c, 0xff0c}, {0xff0e, 0xff0f}, {0xff1a, 0xff1b}, {0xff1f, 0xff20},
	{0xff3b, 0xff3d}, {0xff5b, 0xff5b}, {0xff5d, 0xff5d}, {0xff5f, 0xff65},
}

// layoutTextUtilTestUnicode16Separators were added to the punctuation
// categories in Unicode 16, which JDK 25 implements. They are separators in
// Java and are not asserted here, because the answer depends on the Unicode
// version of the Go unicode package.
var layoutTextUtilTestUnicode16Separators = map[rune]bool{0x1b4e: true, 0x1b4f: true, 0x1b7f: true}

func TestLayoutTextUtilIsFirstLetterSeparatorCharAgainstJava(t *testing.T) {
	want := map[rune]bool{}
	for _, r := range layoutTextUtilTestSeparatorRanges {
		for c := r[0]; c <= r[1]; c++ {
			want[c] = true
		}
	}
	for c := rune(0); c <= 0xFFFF; c++ {
		if layoutTextUtilTestUnicode16Separators[c] {
			continue
		}
		if got := LayoutTextUtilIsFirstLetterSeparatorChar(c); got != want[c] {
			t.Errorf("IsFirstLetterSeparatorChar(U+%04X) = %v, want %v", c, got, want[c])
		}
	}
	// Java sees a character outside the BMP as two surrogates, which are not
	// separators: U+1F676 is in category Po.
	if LayoutTextUtilIsFirstLetterSeparatorChar(0x1F676) {
		t.Errorf("IsFirstLetterSeparatorChar(U+1F676) = true, want false")
	}
}

// layoutTextUtilTestTextRenderer measures every character as 100 dots.
type layoutTextUtilTestTextRenderer struct{}

func (layoutTextUtilTestTextRenderer) Setup(context FontContext) {}
func (layoutTextUtilTestTextRenderer) DrawString(outputDevice OutputDevice, str string, x float32, y float32) {
}
func (layoutTextUtilTestTextRenderer) DrawStringWithInfo(outputDevice OutputDevice, str string, x float32, y float32, info *JustificationInfo) {
}
func (layoutTextUtilTestTextRenderer) DrawGlyphVector(outputDevice OutputDevice, vector FSGlyphVector, x float32, y float32) {
}
func (layoutTextUtilTestTextRenderer) GetGlyphVector(outputDevice OutputDevice, font FSFont, str string) FSGlyphVector {
	return nil
}
func (layoutTextUtilTestTextRenderer) GetGlyphPositions(outputDevice OutputDevice, font FSFont, fsGlyphVector FSGlyphVector) []float32 {
	return nil
}
func (layoutTextUtilTestTextRenderer) GetGlyphBounds(outputDevice OutputDevice, font FSFont, fsGlyphVector FSGlyphVector, index int, x float32, y float32) *geom.Rectangle {
	return nil
}
func (layoutTextUtilTestTextRenderer) GetFSFontMetrics(context FontContext, font FSFont, str string) FSFontMetrics {
	return nil
}
func (layoutTextUtilTestTextRenderer) GetWidth(context FontContext, font FSFont, str string) int {
	return 100 * utf8.RuneCountInString(str)
}
func (layoutTextUtilTestTextRenderer) SetFontScale(scale float32)         {}
func (layoutTextUtilTestTextRenderer) GetFontScale() float32              { return 1 }
func (layoutTextUtilTestTextRenderer) SetSmoothingThreshold(size float32) {}

// layoutTextUtilTestContext is the 20 dots per pixel CssContext of the
// CalculatedStyle tests with a text renderer.
type layoutTextUtilTestContext struct {
	calculatedStyleTestContext
}

func (c *layoutTextUtilTestContext) GetTextRenderer() TextRenderer {
	return layoutTextUtilTestTextRenderer{}
}

func TestLayoutTextUtilTextWidth(t *testing.T) {
	tests := []struct {
		letterSpacing string
		value         float32
		text          string
		want          int
	}{
		{"", 0, "héllo", 500},
		// 0.25px is 5 dots after each of the 5 characters
		{"0.25px", 0.25, "héllo", 525},
		// LengthValue rounds 5.2 dots to 5 dots
		{"0.26px", 0.26, "abc", 315},
		{"0.26px", 0.26, "", 0},
		// a character outside the BMP counts once
		{"0.25px", 0.25, "😀", 105},
	}
	for _, tt := range tests {
		decls := []*PropertyDeclaration{
			CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueInline),
		}
		if tt.letterSpacing != "" {
			decls = append(decls, NewPropertyDeclaration(CSSNameLetterSpacing,
				NewPropertyValueFloat(CSSPrimitiveValueCssPx, tt.value, tt.letterSpacing), true, StylesheetInfoOriginUser))
		}
		style := NewEmptyStyle().DeriveStyle(CascadedStyleCreateLayoutStyle(decls...))
		c := &layoutTextUtilTestContext{}
		if got := LayoutTextUtilTextWidth(c, style, nil, tt.text); got != tt.want {
			t.Errorf("TextWidth(%q, letter-spacing %q) = %d, want %d", tt.text, tt.letterSpacing, got, tt.want)
		}
	}
}
