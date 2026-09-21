// Tests of the tokenizer ported from
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/Lexer.flex
//
// Flying Saucer has no JUnit test of the lexer. The expected token sequences
// below are what Lexer.flex specifies; each was also produced by the
// JFlex-generated Lexer.java for the same input. A token is written as its
// name, its text (Yytext) and the line on which it starts (YyLine).

package ufo

import (
	"errors"
	"strings"
	"testing"
)

type lexerTestToken struct {
	name string
	text string
	line int
}

var lexerTests = []struct {
	name   string
	input  string
	tokens []lexerTestToken
}{
	{
		name:  "identifiersWithEscapes",
		input: "te\\st \\41 bc -moz-x _a \\{x caf\u00e9 \\-a --b -1",
		tokens: []lexerTestToken{
			{"IDENT", "te\\st", 0}, {"S", " ", 0}, {"IDENT", "\\41 bc", 0}, {"S", " ", 0},
			{"IDENT", "-moz-x", 0}, {"S", " ", 0}, {"IDENT", "_a", 0}, {"S", " ", 0},
			{"IDENT", "\\{x", 0}, {"S", " ", 0}, {"IDENT", "caf\u00e9", 0}, {"S", " ", 0},
			{"IDENT", "\\-a", 0}, {"S", " ", 0}, {"MINUS", "-", 0}, {"IDENT", "-b", 0},
			{"S", " ", 0}, {"MINUS", "-", 0}, {"NUMBER", "1", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "stringsWithEscapedNewlines",
		input: "\"a\\\nb\" 'c\\\r\nd' \"e\\\ff\" 'it\\'s' \"\\22 x\" \"q'q\" ''",
		tokens: []lexerTestToken{
			{"STRING", "\"a\\\nb\"", 0}, {"S", " ", 1}, {"STRING", "'c\\\r\nd'", 1}, {"S", " ", 2},
			{"STRING", "\"e\\\ff\"", 2}, {"S", " ", 3}, {"STRING", "'it\\'s'", 3}, {"S", " ", 3},
			{"STRING", "\"\\22 x\"", 3}, {"S", " ", 3}, {"STRING", "\"q'q\"", 3}, {"S", " ", 3},
			{"STRING", "''", 3}, {"EOF", "", 3},
		},
	},
	{
		name:  "unclosedStrings",
		input: "\"abc\nx 'def\\",
		tokens: []lexerTestToken{
			{"INVALID", "\"abc", 0}, {"S", "\n", 0}, {"IDENT", "x", 1}, {"S", " ", 1},
			{"INVALID", "'def", 1}, {"OTHER", "\\", 1}, {"EOF", "", 1},
		},
	},
	{
		name:  "urlQuotedAndUnquoted",
		input: "url(\"a b.png\") url( 'x' ) url(foo/bar.png) url( a\\)b ) URL(x) url() url(\\41 b) url(a b) url('a",
		tokens: []lexerTestToken{
			{"URI", "url(\"a b.png\")", 0}, {"S", " ", 0}, {"URI", "url( 'x' )", 0}, {"S", " ", 0},
			{"URI", "url(foo/bar.png)", 0}, {"S", " ", 0}, {"URI", "url( a\\)b )", 0}, {"S", " ", 0},
			{"URI", "URL(x)", 0}, {"S", " ", 0}, {"URI", "url()", 0}, {"S", " ", 0},
			{"URI", "url(\\41 b)", 0}, {"S", " ", 0}, {"FUNCTION", "url(", 0}, {"IDENT", "a", 0},
			{"S", " ", 0}, {"IDENT", "b", 0}, {"RPAREN", ")", 0}, {"S", " ", 0},
			{"FUNCTION", "url(", 0}, {"INVALID", "'a", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "numbersWithEachUnit",
		input: "1em 1.5ex 10px 2cm 3mm 4in 5pt 6pc 90deg 1rad 100grad 5ms 2s 50hz 3khz 0.25turn 1emx 12 .5 1. 1.5.5",
		tokens: []lexerTestToken{
			{"EMS", "1em", 0}, {"S", " ", 0}, {"EXS", "1.5ex", 0}, {"S", " ", 0},
			{"PX", "10px", 0}, {"S", " ", 0}, {"CM", "2cm", 0}, {"S", " ", 0},
			{"MM", "3mm", 0}, {"S", " ", 0}, {"IN", "4in", 0}, {"S", " ", 0},
			{"PT", "5pt", 0}, {"S", " ", 0}, {"PC", "6pc", 0}, {"S", " ", 0},
			{"ANGLE", "90deg", 0}, {"S", " ", 0}, {"ANGLE", "1rad", 0}, {"S", " ", 0},
			{"ANGLE", "100grad", 0}, {"S", " ", 0}, {"TIME", "5ms", 0}, {"S", " ", 0},
			{"TIME", "2s", 0}, {"S", " ", 0}, {"FREQ", "50hz", 0}, {"S", " ", 0},
			{"FREQ", "3khz", 0}, {"S", " ", 0}, {"DIMENSION", "0.25turn", 0}, {"S", " ", 0},
			{"DIMENSION", "1emx", 0}, {"S", " ", 0}, {"NUMBER", "12", 0}, {"S", " ", 0},
			{"NUMBER", ".5", 0}, {"S", " ", 0}, {"NUMBER", "1", 0}, {"PERIOD", ".", 0},
			{"S", " ", 0}, {"NUMBER", "1.5", 0}, {"NUMBER", ".5", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "unitsInMixedCaseAndEscaped",
		input: "1EM;2Px;3e\\6d;3E\\4D;4\\70\\78;5\\p\\x;6\\000070 x;7e\\6dd;8\\g\\r\\a\\d;8\\g\\r\\61\\64;9\\00000070x;1e\\6d x;1\\65\r\n\\6d\tx",
		tokens: []lexerTestToken{
			{"EMS", "1EM", 0}, {"SEMICOLON", ";", 0}, {"PX", "2Px", 0}, {"SEMICOLON", ";", 0},
			{"EMS", "3e\\6d", 0}, {"SEMICOLON", ";", 0}, {"EMS", "3E\\4D", 0}, {"SEMICOLON", ";", 0},
			{"PX", "4\\70\\78", 0}, {"SEMICOLON", ";", 0}, {"PX", "5\\p\\x", 0}, {"SEMICOLON", ";", 0},
			{"PX", "6\\000070 x", 0}, {"SEMICOLON", ";", 0}, {"DIMENSION", "7e\\6dd", 0},
			{"SEMICOLON", ";", 0}, {"DIMENSION", "8\\g\\r\\a\\d", 0}, {"SEMICOLON", ";", 0},
			{"ANGLE", "8\\g\\r\\61\\64", 0}, {"SEMICOLON", ";", 0}, {"DIMENSION", "9\\00000070x", 0},
			{"SEMICOLON", ";", 0}, {"DIMENSION", "1e\\6d x", 0}, {"SEMICOLON", ";", 0}, {"DIMENSION", "1\\65\r\n\\6d\tx", 0},
			{"EOF", "", 1},
		},
	},
	{
		name:  "unitsWithUnicodeCaseFolding",
		input: "1\u212ahz 2m\u017f 3\u0131n 4\u0130n",
		tokens: []lexerTestToken{
			{"FREQ", "1\u212ahz", 0}, {"S", " ", 0}, {"TIME", "2m\u017f", 0}, {"S", " ", 0},
			{"IN", "3\u0131n", 0}, {"S", " ", 0}, {"IN", "4\u0130n", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "percentages",
		input: "50% .5% 1.25% 5 %",
		tokens: []lexerTestToken{
			{"PERCENTAGE", "50%", 0}, {"S", " ", 0}, {"PERCENTAGE", ".5%", 0}, {"S", " ", 0},
			{"PERCENTAGE", "1.25%", 0}, {"S", " ", 0}, {"NUMBER", "5", 0}, {"S", " ", 0},
			{"OTHER", "%", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "hash",
		input: "#fff #00FF00 #-a_b #\\31 x # x #",
		tokens: []lexerTestToken{
			{"HASH", "#fff", 0}, {"S", " ", 0}, {"HASH", "#00FF00", 0}, {"S", " ", 0},
			{"HASH", "#-a_b", 0}, {"S", " ", 0}, {"HASH", "#\\31 x", 0}, {"S", " ", 0},
			{"OTHER", "#", 0}, {"S", " ", 0}, {"IDENT", "x", 0}, {"S", " ", 0},
			{"OTHER", "#", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "important",
		input: "!important ! IMPORTANT !/* c */ /* d */important !\n\timportant !importan !/* x important",
		tokens: []lexerTestToken{
			{"IMPORTANT_SYM", "!important", 0}, {"S", " ", 0}, {"IMPORTANT_SYM", "! IMPORTANT", 0},
			{"S", " ", 0}, {"IMPORTANT_SYM", "!/* c */ /* d */important", 0}, {"S", " ", 0},
			{"IMPORTANT_SYM", "!\n\timportant", 0}, {"S", " ", 1}, {"OTHER", "!", 1}, {"IDENT", "importan", 1},
			{"S", " ", 1}, {"OTHER", "!", 1}, {"VIRGULE", "/", 1}, {"ASTERISK", "*", 1},
			{"S", " ", 1}, {"IDENT", "x", 1}, {"S", " ", 1}, {"IDENT", "important", 1},
			{"EOF", "", 1},
		},
	},
	{
		name:  "atKeywordsInMixedCase",
		input: "@IMPORT @Page @mEdIa @charset \"x\"; @CHARSET  @charset; @NameSpace @Font-Face @top-left @importx @\\69mport @ @1",
		tokens: []lexerTestToken{
			{"IMPORT_SYM", "@IMPORT", 0}, {"S", " ", 0}, {"PAGE_SYM", "@Page", 0}, {"S", " ", 0},
			{"MEDIA_SYM", "@mEdIa", 0}, {"S", " ", 0}, {"CHARSET_SYM", "@charset ", 0}, {"STRING", "\"x\"", 0},
			{"SEMICOLON", ";", 0}, {"S", " ", 0}, {"CHARSET_SYM", "@CHARSET ", 0}, {"S", " ", 0},
			{"AT_RULE", "@charset", 0}, {"SEMICOLON", ";", 0}, {"S", " ", 0}, {"NAMESPACE_SYM", "@NameSpace", 0},
			{"S", " ", 0}, {"FONT_FACE_SYM", "@Font-Face", 0}, {"S", " ", 0}, {"AT_RULE", "@top-left", 0},
			{"S", " ", 0}, {"AT_RULE", "@importx", 0}, {"S", " ", 0}, {"AT_RULE", "@\\69mport", 0},
			{"S", " ", 0}, {"OTHER", "@", 0}, {"S", " ", 0}, {"OTHER", "@", 0},
			{"NUMBER", "1", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "comments",
		input: "a/* c1 */b /**/ /* multi\nline **/c /***/ /* unterminated",
		tokens: []lexerTestToken{
			{"IDENT", "a", 0}, {"IDENT", "b", 0}, {"S", " ", 0}, {"S", " ", 0},
			{"IDENT", "c", 1}, {"S", " ", 1}, {"S", " ", 1}, {"VIRGULE", "/", 1},
			{"ASTERISK", "*", 1}, {"S", " ", 1}, {"IDENT", "unterminated", 1}, {"EOF", "", 1},
		},
	},
	{
		name:  "cdoAndCdc",
		input: "<!-- a --> <! -- <!- -->x",
		tokens: []lexerTestToken{
			{"CDO", "<!--", 0}, {"S", " ", 0}, {"IDENT", "a", 0}, {"S", " ", 0},
			{"CDC", "-->", 0}, {"S", " ", 0}, {"OTHER", "<", 0}, {"OTHER", "!", 0},
			{"S", " ", 0}, {"MINUS", "-", 0}, {"MINUS", "-", 0}, {"S", " ", 0},
			{"OTHER", "<", 0}, {"OTHER", "!", 0}, {"MINUS", "-", 0}, {"S", " ", 0},
			{"CDC", "-->", 0}, {"IDENT", "x", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "unicodeEscapes",
		input: "\\26 B \\000026B \\41\r\nx \\41\rx \\41\n\nx \\0000417 \\zz",
		tokens: []lexerTestToken{
			{"IDENT", "\\26 B", 0}, {"S", " ", 0}, {"IDENT", "\\000026B", 0}, {"S", " ", 0},
			{"IDENT", "\\41\r\nx", 0}, {"S", " ", 1}, {"IDENT", "\\41\rx", 1}, {"S", " ", 2},
			{"IDENT", "\\41\n", 2}, {"S", "\n", 3}, {"IDENT", "x", 4}, {"S", " ", 4},
			{"IDENT", "\\0000417", 4}, {"S", " ", 4}, {"IDENT", "\\zz", 4}, {"EOF", "", 4},
		},
	},
	{
		name:  "operatorsAndPunctuation",
		input: "a  {  b + c > d , e~=f|=g^=h$=i*=j } ; / : - ) [ ] . = * | ( & ~ ^ $ < \\",
		tokens: []lexerTestToken{
			{"IDENT", "a", 0}, {"LBRACE", "  {", 0}, {"S", "  ", 0}, {"IDENT", "b", 0},
			{"PLUS", " +", 0}, {"S", " ", 0}, {"IDENT", "c", 0}, {"GREATER", " >", 0},
			{"S", " ", 0}, {"IDENT", "d", 0}, {"COMMA", " ,", 0}, {"S", " ", 0},
			{"IDENT", "e", 0}, {"INCLUDES", "~=", 0}, {"IDENT", "f", 0}, {"DASHMATCH", "|=", 0},
			{"IDENT", "g", 0}, {"PREFIXMATCH", "^=", 0}, {"IDENT", "h", 0}, {"SUFFIXMATCH", "$=", 0},
			{"IDENT", "i", 0}, {"SUBSTRINGMATCH", "*=", 0}, {"IDENT", "j", 0}, {"S", " ", 0},
			{"RBRACE", "}", 0}, {"S", " ", 0}, {"SEMICOLON", ";", 0}, {"S", " ", 0},
			{"VIRGULE", "/", 0}, {"S", " ", 0}, {"COLON", ":", 0}, {"S", " ", 0},
			{"MINUS", "-", 0}, {"S", " ", 0}, {"RPAREN", ")", 0}, {"S", " ", 0},
			{"LBRACKET", "[", 0}, {"S", " ", 0}, {"RBRACKET", "]", 0}, {"S", " ", 0},
			{"PERIOD", ".", 0}, {"S", " ", 0}, {"EQUALS", "=", 0}, {"S", " ", 0},
			{"ASTERISK", "*", 0}, {"S", " ", 0}, {"VERTICAL_BAR", "|", 0}, {"S", " ", 0},
			{"OTHER", "(", 0}, {"S", " ", 0}, {"OTHER", "&", 0}, {"S", " ", 0},
			{"OTHER", "~", 0}, {"S", " ", 0}, {"OTHER", "^", 0}, {"S", " ", 0},
			{"OTHER", "$", 0}, {"S", " ", 0}, {"OTHER", "<", 0}, {"S", " ", 0},
			{"OTHER", "\\", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "functions",
		input: "rgb(1,2,3) counter( x ) -fs-f(\\66(1) a\\((x",
		tokens: []lexerTestToken{
			{"FUNCTION", "rgb(", 0}, {"NUMBER", "1", 0}, {"COMMA", ",", 0}, {"NUMBER", "2", 0},
			{"COMMA", ",", 0}, {"NUMBER", "3", 0}, {"RPAREN", ")", 0}, {"S", " ", 0},
			{"FUNCTION", "counter(", 0}, {"S", " ", 0}, {"IDENT", "x", 0}, {"S", " ", 0},
			{"RPAREN", ")", 0}, {"S", " ", 0}, {"FUNCTION", "-fs-f(", 0}, {"FUNCTION", "\\66(", 0},
			{"NUMBER", "1", 0}, {"RPAREN", ")", 0}, {"S", " ", 0}, {"FUNCTION", "a\\((", 0},
			{"IDENT", "x", 0}, {"EOF", "", 0},
		},
	},
	{
		name:  "whitespaceBeforeBraceAndCombinators",
		input: "a \n{b\t+c\r\n>d\f,e f",
		tokens: []lexerTestToken{
			{"IDENT", "a", 0}, {"LBRACE", " \n{", 0}, {"IDENT", "b", 1}, {"PLUS", "\t+", 1},
			{"IDENT", "c", 1}, {"GREATER", "\r\n>", 1}, {"IDENT", "d", 2}, {"COMMA", "\f,", 2},
			{"IDENT", "e", 3}, {"S", " ", 3}, {"IDENT", "f", 3}, {"EOF", "", 3},
		},
	},
	{
		name:  "ruleset",
		input: "div.c#i > p:first-child { color: #FFF !important; margin: -1px auto }",
		tokens: []lexerTestToken{
			{"IDENT", "div", 0}, {"PERIOD", ".", 0}, {"IDENT", "c", 0}, {"HASH", "#i", 0},
			{"GREATER", " >", 0}, {"S", " ", 0}, {"IDENT", "p", 0}, {"COLON", ":", 0},
			{"IDENT", "first-child", 0}, {"LBRACE", " {", 0}, {"S", " ", 0}, {"IDENT", "color", 0},
			{"COLON", ":", 0}, {"S", " ", 0}, {"HASH", "#FFF", 0}, {"S", " ", 0},
			{"IMPORTANT_SYM", "!important", 0}, {"SEMICOLON", ";", 0}, {"S", " ", 0}, {"IDENT", "margin", 0},
			{"COLON", ":", 0}, {"S", " ", 0}, {"MINUS", "-", 0}, {"PX", "1px", 0},
			{"S", " ", 0}, {"IDENT", "auto", 0}, {"S", " ", 0}, {"RBRACE", "}", 0},
			{"EOF", "", 0},
		},
	},
}

func TestLexerTokens(t *testing.T) {
	for _, tt := range lexerTests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			for i, want := range tt.tokens {
				token, err := lexer.Yylex()
				if err != nil {
					t.Fatalf("token %d: %v", i, err)
				}
				got := lexerTestToken{token.GetName(), lexer.Yytext(), lexer.YyLine()}
				if got != want {
					t.Fatalf("token %d: got %+q, want %+q", i, []any{got.name, got.text, got.line}, []any{want.name, want.text, want.line})
				}
				if lexer.Yylength() != len([]rune(want.text)) {
					t.Errorf("token %d: Yylength() = %d, want %d", i, lexer.Yylength(), len([]rune(want.text)))
				}
			}
			if !lexer.YyatEOF() {
				t.Errorf("YyatEOF() = false after the end of file token")
			}
		})
	}
}

func TestLexerReturnsTheTokenConstants(t *testing.T) {
	lexer := NewLexer(strings.NewReader("a ?"))
	want := []*Token{TokenTkIdent, TokenTkS, nil, TokenTkEof, TokenTkEof}
	for i, w := range want {
		token, err := lexer.Yylex()
		if err != nil {
			t.Fatal(err)
		}
		if w == nil {
			if token.GetType() != TokenTypeOther || token.GetExternalName() != "? (other)" {
				t.Errorf("token %d: got %v (%s), want an other token for ?", i, token, token.GetExternalName())
			}
		} else if token != w {
			t.Errorf("token %d: got %v, want %v", i, token, w)
		}
	}
}

func TestLexerYyreset(t *testing.T) {
	lexer := NewLexer(strings.NewReader("a\nb"))
	for i := 0; i < 3; i++ {
		if _, err := lexer.Yylex(); err != nil {
			t.Fatal(err)
		}
	}
	if lexer.Yytext() != "b" || lexer.YyLine() != 1 {
		t.Fatalf("got %q on line %d, want \"b\" on line 1", lexer.Yytext(), lexer.YyLine())
	}

	lexer.Yyreset(strings.NewReader("10px"))
	if lexer.YyLine() != 0 || lexer.YyatEOF() || lexer.Yystate() != LexerYyinitial {
		t.Errorf("Yyreset left line %d, atEOF %v, state %d", lexer.YyLine(), lexer.YyatEOF(), lexer.Yystate())
	}
	token, err := lexer.Yylex()
	if err != nil {
		t.Fatal(err)
	}
	if token != TokenTkPx || lexer.Yytext() != "10px" {
		t.Errorf("got %v %q, want PX \"10px\"", token, lexer.Yytext())
	}
}

func TestLexerSetYyLine(t *testing.T) {
	lexer := NewLexer(strings.NewReader("a\nb"))
	lexer.SetYyLine(10)
	for i := 0; i < 3; i++ {
		if _, err := lexer.Yylex(); err != nil {
			t.Fatal(err)
		}
	}
	if lexer.YyLine() != 11 {
		t.Errorf("YyLine() = %d, want 11", lexer.YyLine())
	}
}

func TestLexerYypushback(t *testing.T) {
	lexer := NewLexer(strings.NewReader("abc;"))
	if _, err := lexer.Yylex(); err != nil {
		t.Fatal(err)
	}
	if lexer.Yycharat(0) != 'a' || lexer.Yycharat(2) != 'c' {
		t.Errorf("Yycharat: got %c and %c, want a and c", lexer.Yycharat(0), lexer.Yycharat(2))
	}

	lexer.Yypushback(2)
	if lexer.Yytext() != "a" || lexer.Yylength() != 1 {
		t.Errorf("after Yypushback(2): Yytext() = %q, Yylength() = %d", lexer.Yytext(), lexer.Yylength())
	}
	token, err := lexer.Yylex()
	if err != nil {
		t.Fatal(err)
	}
	if token != TokenTkIdent || lexer.Yytext() != "bc" {
		t.Errorf("got %v %q, want IDENT \"bc\"", token, lexer.Yytext())
	}

	defer func() {
		if _, ok := recover().(*XRRuntimeException); !ok {
			t.Errorf("Yypushback(3) of a 2 character token did not panic with an XRRuntimeException")
		}
	}()
	lexer.Yypushback(3)
}

type lexerTestFailingReader struct{}

func (lexerTestFailingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestLexerReturnsReadError(t *testing.T) {
	lexer := NewLexer(lexerTestFailingReader{})
	if _, err := lexer.Yylex(); err == nil || err.Error() != "read failed" {
		t.Errorf("Yylex() error = %v, want read failed", err)
	}
}

func TestLexerYyclose(t *testing.T) {
	lexer := NewLexer(strings.NewReader("a b"))
	if _, err := lexer.Yylex(); err != nil {
		t.Fatal(err)
	}
	if err := lexer.Yyclose(); err != nil {
		t.Fatal(err)
	}
	token, err := lexer.Yylex()
	if err != nil {
		t.Fatal(err)
	}
	if token != TokenTkEof {
		t.Errorf("got %v after Yyclose, want EOF", token)
	}
}
