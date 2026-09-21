// Tests of CSSParser that Flying Saucer does not have: the at-rules, the
// selector grammar, the term grammar and the error recovery.
//
// Each case parses css and writes what the parser produced as text: the
// messages given to the CSSErrorHandler ("ERR uri | message") and to the CSS3
// feature listener, the import rules, the font face rules, and the rulesets,
// media rules and page rules with their declarations. A selector is written
// as the axis, specificity, pseudo element and pseudo classes of each
// selector of its chain, followed by the ids of the elements of
// cssParserDumpTestXML that it matches. The expected text of every case is
// the output of the same dump written in Java and run against Flying
// Saucer's CSSParser.

package ufo

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

const cssParserDumpTestXML = `<root id='root' xmlns:svg='http://www.w3.org/2000/svg'>
  <div id='d1' class='a b' lang='en-US' title='hello world'>
    <p id='p1' class='note'>x</p>
    <p id='p2'>x<span id='s1' class='a'>y</span></p>
    <p id='p3' class='note last' data-x='abc-def'>x</p>
    <a id='a1' href='http://example.com/x.pdf'>l</a>
  </div>
  <div id='d2' lang='fr'>
    <ul id='u1'><li id='l1'>1</li><li id='l2'>2</li><li id='l3'>3</li><li id='l4'>4</li><li id='l5'>5</li></ul>
    <svg:rect id='r1' width='10'/>
  </div>
</root>`

// cssParserDumpTests: mode is "sheet" for ParseStylesheet, "sheet-cmyk" for
// the same with CMYK colors supported, "decl" for ParseDeclaration and
// "value:<property>" for ParsePropertyValue.
var cssParserDumpTests = []struct {
	name string
	mode string
	uri  string
	css  string
	want string
}{
	{
		name: "charsetImportsNamespace",
		mode: "sheet",
		uri:  "http://example.com/css/main.css",
		css:  "@charset \"utf-8\";\n<!-- @import \"a.css\";\n@import url(b.css) print, screen;\n@import url(\"/c.css\") TV;\n@import \"http://other.org/d.css\" all;\n-->\n@namespace svg \"http://www.w3.org/2000/svg\";\n@namespace url(http://www.w3.org/1999/xhtml);\nsvg|rect { color: red }\n|div { color: blue }\n*|p { color: green }\np { color: black }\n",
		want: `IMPORT http://example.com/css/a.css media=all
IMPORT http://example.com/css/b.css media=print,screen
IMPORT http://example.com/c.css media=tv
IMPORT http://other.org/d.css media=all
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH r1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#008000 ident= op= rgb=0,128,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#000000 ident= op= rgb=0,0,0,1000>
`,
	},
	{
		name: "importRelativeToRelativeSheet",
		mode: "sheet",
		uri:  "css/main.css",
		css:  "@import \"a.css\";\n@import url(b.css);\n",
		want: `IMPORT css/a.css media=all
IMPORT css/css/b.css media=all
`,
	},
	{
		name: "importIntoInlineSheet",
		mode: "sheet",
		uri:  "inline:1",
		css:  "@import \"a.css\";\n@import url(/b.css);\n",
		want: `IMPORT a.css media=all
IMPORT /b.css media=all
`,
	},
	{
		name: "importIntoFileSheet",
		mode: "sheet",
		uri:  "file:/x/y.css",
		css:  "@import \"a.css\";\n@import \"../b.css\";\n@import url(c.css);\n@import \"http://example.com/d.css\";\n",
		want: `IMPORT file:/x/a.css media=all
IMPORT file:/b.css media=all
IMPORT file:/x/c.css media=all
IMPORT http://example.com/d.css media=all
`,
	},
	{
		name: "importErrors",
		mode: "sheet",
		uri:  "http://example.com/main.css",
		css:  "@import foo;\n@import \"a.css\" print,;\n@import \"b.css\" screen\np { color: red }\n@import \"late.css\";\n@namespace x \"late\";\ndiv { color: blue }\n",
		want: `ERR http://example.com/main.css | Found an identifier where a string or a URI was expected at line 1. Skipping @import rule.
ERR http://example.com/main.css | Found ; where an identifier was expected at line 2. Skipping @import rule.
ERR http://example.com/main.css | Found an identifier where ; was expected at line 4. Skipping @import rule.
IMPORT http://example.com/late.css media=all
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
`,
	},
	{
		name: "charsetErrors",
		mode: "sheet",
		uri:  "",
		css:  "@charset foo;\np { color: red }\n",
		want: `ERR null | Found an identifier where a string was expected at line 1. Skipping @charset rule.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "namespaceErrors",
		mode: "sheet",
		uri:  "",
		css:  "@namespace 12;\n@namespace a \"u\"\n p { color: red }\nfoo|p { color: red }\nb { color: blue }\n",
		want: `ERR null | Found a number where a string or a URI was expected at line 1. Skipping @namespace rule.
ERR null | Found an identifier where ; was expected at line 3. Skipping @namespace rule.
ERR null | There is no namespace with prefix foo defined at line 4. Skipping ruleset.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
`,
	},
	{
		name: "mediaRules",
		mode: "sheet",
		uri:  "",
		css:  "@media print { p { color: red } div > p { margin: 0 } }\n@media screen, TV { a:link { color: blue } }\n@media { p { color: red } }\n@media print, { p { color: red } }\n@media all ; p { color: green }\n@media print { p { color: } b { color: red } }\n",
		want: `ERR null | Found a { where an identifier was expected at line 3. Skipping @media rule.
ERR null | Found a { where an identifier was expected at line 4. Skipping @media rule.
ERR null | Found ; where a { was expected at line 5. Skipping @media rule.
ERR null | Found } where one of a number, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 6. Skipping declaration.
MEDIA print=true screen=false tv=false
  RULESET
    SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
    DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
  RULESET
    SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] [CHILD_AXIS 0,0,2 pe= pc=] MATCH d1 d2
    DECL margin-top <VALUE_TYPE_LENGTH 1 f=0 s= css=0 ident= op=>
    DECL margin-right <VALUE_TYPE_LENGTH 1 f=0 s= css=0 ident= op=>
    DECL margin-bottom <VALUE_TYPE_LENGTH 1 f=0 s= css=0 ident= op=>
    DECL margin-left <VALUE_TYPE_LENGTH 1 f=0 s= css=0 ident= op=>
MEDIA print=false screen=true tv=true
  RULESET
    SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH a1
    DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#008000 ident= op= rgb=0,128,0,1000>
MEDIA print=true screen=false tv=false
  RULESET
    SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
    DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "pageRules",
		mode: "sheet",
		uri:  "",
		css:  "@page { size: A4 landscape; margin: 1in; @top-center { content: \"Title\"; color: red } @bottom-right { content: counter(page) \" of \" counter(pages) } }\n@page chapter:first { margin-top: 5cm }\n@page :left { margin-left: 3cm; }\n@page auto { margin: 0 }\n@page :middle { margin: 0 }\n@page x { @bogus-box { color: red } margin: 2px }\n@page y { @top-left color: red; } margin: 3px }\np { color: red }\n",
		want: `ERR null | page name may not be auto at line 4. Skipping @page rule.
ERR null | Pseudo page must be one of first, left, or right at line 5. Skipping @page rule.
ERR null | bogus-box is not a valid margin box name at line 6. Skipping at rule.
ERR null | Found an identifier where a { was expected at line 7. Skipping margin box.
ERR null | Found whitespace where an identifier or function was expected at line 7. Skipping ruleset.
PAGE name= pseudo=
  DECL -fs-page-orientation <VALUE_TYPE_IDENT 21 f=0 s=landscape css=landscape ident= op=>
  DECL -fs-page-width <VALUE_TYPE_LENGTH 7 f=210000 s= css=210mm ident= op=>
  DECL -fs-page-height <VALUE_TYPE_LENGTH 7 f=297000 s= css=297mm ident= op=>
  DECL margin-top <VALUE_TYPE_LENGTH 8 f=1000 s= css=1in ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 8 f=1000 s= css=1in ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 8 f=1000 s= css=1in ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 8 f=1000 s= css=1in ident= op=>
  MARGIN bottom-right
    DECL content <VALUE_TYPE_LIST 0 f=0 s= css=[counter(page,), " of ", counter(pages,)] ident= op= list=(<VALUE_TYPE_FUNCTION 0 f=0 s= css=counter(page,) ident= op= fn=counter(<VALUE_TYPE_IDENT 21 f=0 s=page css=page ident= op=>)><VALUE_TYPE_STRING 19 f=0 s= of  css=" of " ident= op=><VALUE_TYPE_FUNCTION 0 f=0 s= css=counter(pages,) ident= op= fn=counter(<VALUE_TYPE_IDENT 21 f=0 s=pages css=pages ident= op=>)>)>
  MARGIN top-center
    DECL content <VALUE_TYPE_LIST 0 f=0 s= css=["Title"] ident= op= list=(<VALUE_TYPE_STRING 19 f=0 s=Title css="Title" ident= op=>)>
    DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
PAGE name=chapter pseudo=first
  DECL margin-top <VALUE_TYPE_LENGTH 6 f=5000 s= css=5cm ident= op=>
PAGE name= pseudo=left
  DECL margin-left <VALUE_TYPE_LENGTH 6 f=3000 s= css=3cm ident= op=>
PAGE name=x pseudo=
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
PAGE name=y pseudo=
`,
	},
	{
		name: "fontFace",
		mode: "sheet",
		uri:  "http://example.com/css/main.css",
		css:  "@font-face { font-family: \"My Font\"; src: url(fonts/my.ttf); font-weight: bold; -fs-pdf-font-embed: embed; -fs-pdf-font-encoding: Identity-H }\n@font-face { src: url(x.ttf) }\n@font-face ; p { color: red }\n",
		want: `ERR http://example.com/css/main.css | Found ; where a { was expected at line 3. Skipping @font-face rule.
FONTFACE family=true weight=true style=false
FONTFACE family=false weight=false style=false
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "selectors",
		mode: "sheet",
		uri:  "",
		css:  "div p { color: red }\ndiv > p { color: red }\np + p { color: red }\ndiv p + p span { color: red }\n#d1 .note, #d2 li { color: red }\np.note.last { color: red }\n*.a { color: red }\n[lang] { color: red }\n[lang=fr] { color: red }\n[lang|=\"en\"] { color: red }\n[class~=b] { color: red }\n[title^=hello] { color: red }\n[href$=\".pdf\"] { color: red }\n[data-x*=\"c-d\"] { color: red }\n[ lang = 'fr' ] { color: red }\n",
		want: `RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] [DESCENDANT_AXIS 0,0,2 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] [CHILD_AXIS 0,0,2 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,2 pe= pc=] MATCH p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] [DESCENDANT_AXIS 0,0,3 pe= pc=] [DESCENDANT_AXIS 0,0,4 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 1,0,0 pe= pc=] [DESCENDANT_AXIS 1,1,0 pe= pc=] MATCH d1
  SEL [DESCENDANT_AXIS 1,0,0 pe= pc=] [DESCENDANT_AXIS 1,0,1 pe= pc=] MATCH d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,2,1 pe= pc=] MATCH p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH d1 s1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH d1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH d1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH d1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH a1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "pseudoSelectors",
		mode: "sheet",
		uri:  "",
		css:  "a:link { color: red }\na:visited { color: red }\na:hover, a:active, a:focus { color: red }\nli:first-child { color: red }\nli:last-child { color: red }\nli:even { color: red }\nli:odd { color: red }\nli:nth-child(2n+1) { color: red }\nli:nth-child( 3 ) { color: red }\nli:nth-child(odd) { color: red }\nli:nth-child(-n+2) { color: red }\ndiv:lang(en) { color: red }\np:first-line { color: red }\np::first-letter { color: red }\np:before { color: red }\np::after { color: red }\ndiv:has(> p.note) { color: red }\ndiv:has(ul li, span) { color: red }\np:has(+ p.last) { color: red }\n",
		want: `RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH a1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=2.] MATCH a1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=4.] MATCH a1
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=8.] MATCH a1
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=16.] MATCH a1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l5
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l1 l3 l5
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l2 l4
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l1 l3 l5
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l1 l3 l5
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH l1 l2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH d1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,2 pe=first-line pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,2 pe=first-letter pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,2 pe=before pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,2 pe=after pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,2 pe= pc=] MATCH d1
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,3 pe= pc=] MATCH d1 d2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,2 pe= pc=] MATCH p2
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "selectorErrors",
		mode: "sheet",
		uri:  "",
		css:  "p:bogus { color: red }\np::bogus { color: red }\np:foo(x) { color: red }\np:lang() { color: red }\np:nth-child(foo) { color: red }\np:first-line span { color: red }\ndiv:has() { color: red }\ndiv:has(,p) { color: red }\ndiv:has(p q { color: red }\np > { color: red }\n> p { color: red }\np..a { color: red }\np[] { color: red }\np[a=] { color: red }\np[a=b { color: red }\np[|] { color: red }\np, { color: red }\n123 { color: red }\nok { color: green }\n",
		want: `ERR null | bogus is not a recognized pseudo-class at line 1. Skipping ruleset.
ERR null | bogus is not a recognized pseudo-element at line 2. Skipping ruleset.
ERR null | foo is not a valid function in this context at line 3. Skipping ruleset.
ERR null | Found ) where an identifier was expected at line 4. Skipping ruleset.
ERR null | Invalid nth-child selector: foo at line 5. Skipping ruleset.
ERR null | A simple selector with a pseudo element cannot be combined with another simple selector at line 6. Skipping ruleset.
ERR null | The :has() pseudo-class requires a non-empty selector argument at line 7. Skipping ruleset.
ERR null | The :has() pseudo-class does not allow empty selectors at line 8. Skipping ruleset.
ERR null | Found a { where one of a +, a >, whitespace, a comma, or ) was expected at line 9. Skipping ruleset.
ERR null | Found a { where one of an identifier, *, a hex color, ., [, or : was expected at line 10. Skipping ruleset.
ERR null | Found a > where one of a hex color, ., [, or : was expected at line 10. Skipping ruleset.
ERR null | Found . where an identifier was expected at line 12. Skipping ruleset.
ERR null | Found ] where an identifier or * was expected at line 13. Skipping ruleset.
ERR null | Found ] where an identifier or a string was expected at line 14. Skipping ruleset.
ERR null | Found a { where one of =, an attribute word match, an attribute hyphen match, an attribute prefix match, an attribute suffix match, an attribute substring match, or ] was expected at line 15. Skipping ruleset.
ERR null | Found ] where * or an identifier was expected at line 16. Skipping ruleset.
ERR null | Found a { where one of a hex color, ., [, or : was expected at line 17. Skipping ruleset.
ERR null | Found a number where one of a hex color, ., [, or : was expected at line 18. Skipping ruleset.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#008000 ident= op= rgb=0,128,0,1000>
`,
	},
	{
		name: "declarationValues",
		mode: "sheet",
		uri:  "http://example.com/css/main.css",
		css:  "p { margin: 1px 2em 3ex 4%; padding: 1cm 2mm 3in 4pt; text-indent: 5pc; line-height: 1.5; width: -10px; height: +20px; font-family: \"Times New Roman\", Arial, sans-serif; font: italic bold 12px/14px Helvetica, serif; color: #abc; background-color: #A1B2C3; border-color: rgb(10%, 20%, 30%); background-image: url( img/x.png ); content: \"a\\\"b\" attr(title) counter(c, upper-roman); z-index: 5; COLOR: RED !important; quotes: \"\\201C\" \"\\201D\" }\n",
		want: `ERR http://example.com/css/main.css | width may not be negative at line 1. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 3 f=2000 s= css=2em ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 4 f=3000 s= css=3ex ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 2 f=4000 s= css=4% ident= op=>
  DECL padding-top <VALUE_TYPE_LENGTH 6 f=1000 s= css=1cm ident= op=>
  DECL padding-right <VALUE_TYPE_LENGTH 7 f=2000 s= css=2mm ident= op=>
  DECL padding-bottom <VALUE_TYPE_LENGTH 8 f=3000 s= css=3in ident= op=>
  DECL padding-left <VALUE_TYPE_LENGTH 9 f=4000 s= css=4pt ident= op=>
  DECL text-indent <VALUE_TYPE_LENGTH 10 f=5000 s= css=5pc ident= op=>
  DECL line-height <VALUE_TYPE_NUMBER 1 f=1500 s= css=1.5 ident= op=>
  DECL height <VALUE_TYPE_LENGTH 5 f=20000 s= css=20px ident= op=>
  DECL font-family <VALUE_TYPE_STRING 19 f=0 s=Times New Roman,Arial,sans-serif css=Times New Roman,Arial,sans-serif ident= op=>
  DECL font-style <VALUE_TYPE_IDENT 21 f=0 s=italic css=italic ident=italic op=>
  DECL font-variant <VALUE_TYPE_IDENT 21 f=0 s=normal css=normal ident=normal op=>
  DECL font-weight <VALUE_TYPE_IDENT 21 f=0 s=bold css=bold ident=bold op=>
  DECL font-size <VALUE_TYPE_LENGTH 5 f=12000 s= css=12px ident= op=>
  DECL line-height <VALUE_TYPE_LENGTH 5 f=14000 s= css=14px ident= op=VIRGULE>
  DECL font-family <VALUE_TYPE_STRING 19 f=0 s=Helvetica,serif css=Helvetica,serif ident= op=>
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#aabbcc ident= op= rgb=170,187,204,1000>
  DECL background-color <VALUE_TYPE_COLOR 25 f=0 s= css=#a1b2c3 ident= op= rgb=161,178,195,1000>
  DECL border-top-color <VALUE_TYPE_COLOR 25 f=0 s= css=#19334c ident= op= rgb=25,51,76,1000>
  DECL border-right-color <VALUE_TYPE_COLOR 25 f=0 s= css=#19334c ident= op= rgb=25,51,76,1000>
  DECL border-bottom-color <VALUE_TYPE_COLOR 25 f=0 s= css=#19334c ident= op= rgb=25,51,76,1000>
  DECL border-left-color <VALUE_TYPE_COLOR 25 f=0 s= css=#19334c ident= op= rgb=25,51,76,1000>
  DECL background-image <VALUE_TYPE_STRING 20 f=0 s=http://example.com/css/ img/x.png  css=url( img/x.png ) ident= op=>
  DECL content <VALUE_TYPE_LIST 0 f=0 s= css=["a\"b", attr(title,), counter(c,upper-roman,)] ident= op= list=(<VALUE_TYPE_STRING 19 f=0 s=a"b css="a\"b" ident= op=><VALUE_TYPE_FUNCTION 0 f=0 s= css=attr(title,) ident= op= fn=attr(<VALUE_TYPE_IDENT 21 f=0 s=title css=title ident= op=>)><VALUE_TYPE_FUNCTION 0 f=0 s= css=counter(c,upper-roman,) ident= op= fn=counter(<VALUE_TYPE_IDENT 21 f=0 s=c css=c ident= op=><VALUE_TYPE_IDENT 21 f=0 s=upper-roman css=upper-roman ident= op=COMMA>)>)>
  DECL z-index <VALUE_TYPE_NUMBER 1 f=5000 s= css=5 ident= op=>
  DECL color !important <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
  DECL quotes <VALUE_TYPE_LIST 0 f=0 s= css=[“, ”] ident= op= list=({String}{String})>
`,
	},
	{
		name: "cmykUnsupported",
		mode: "sheet",
		uri:  "",
		css:  "p { color: cmyk(0, 0.5, 100%, 0); background-color: red }\n",
		want: `ERR null | The current output device does not support CMYK colors at line 1. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL background-color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "cmykSupported",
		mode: "sheet-cmyk",
		uri:  "",
		css:  "p { color: cmyk(0, 0.5, 100%, 0) }\nb { color: cmyk(0, 0.5, 1) }\ni { color: cmyk(2, 0, 0, 0) }\nu { color: cmyk(a, 0, 0, 0) }\n",
		want: `ERR null | The cmyk() function must have exactly four parameters at line 2. Skipping declaration.
ERR null | Parameter 1 to the cmyk() function must be between zero and one at line 3. Skipping declaration.
ERR null | Parameter 1 to the cmyk() function is not a number or a percentage at line 4. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=cmyk(0.0, 0.5, 1.0, 0.0) ident= op= cmyk=0,500,1000,0>
`,
	},
	{
		name: "colorErrors",
		mode: "sheet",
		uri:  "",
		css:  "p { color: #abcd }\nb { color: #ggg }\ni { color: rgb(1, 2) }\nu { color: rgb(1px, 2, 3) }\ns { color: rgba(1, 2, 3, 50%) }\nem { color: rgb(1 2 3) }\n",
		want: `ERR null | #abcd is not a valid color definition at line 1. Skipping declaration.
ERR null | #ggg is not a valid color definition at line 2. Skipping declaration.
ERR null | The rgb() function must have three or four parameters at line 3. Skipping declaration.
ERR null | Parameter 1 to the rgb() function is not a number or percentage at line 4. Skipping declaration.
ERR null | Parameter alpha to the rgba() function is not a number at line 5. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#010203 ident= op= rgb=1,2,3,1000>
`,
	},
	{
		name: "unitErrors",
		mode: "sheet",
		uri:  "",
		css:  "p { width: 10s }\nb { width: 10hz }\ni { width: 10foo }\nu { transform: rotate(1turn) }\ns { width: 1e\\6d }\n",
		want: `ERR null | Unsupported CSS unit s at line 1. Skipping declaration.
ERR null | Unsupported CSS unit hz at line 2. Skipping declaration.
ERR null | Unsupported CSS unit foo at line 3. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL transform <VALUE_TYPE_LIST 0 f=0 s= css=[rotate(1turn,)] ident= op= list=(<VALUE_TYPE_FUNCTION 0 f=0 s= css=rotate(1turn,) ident= op= fn=rotate(<VALUE_TYPE_LENGTH 11 f=360000 s= css=1turn ident= op=>)>)>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL width <VALUE_TYPE_LENGTH 3 f=1000 s= css=1e\6d  ident= op=>
`,
	},
	{
		name: "declarationErrors",
		mode: "sheet",
		uri:  "",
		css:  "p { color red; margin: 1px }\nb { : red; margin: 2px }\ni { color: red green blue; margin: 3px }\nu { bogus-prop: 1; margin: 4px }\ns { color: ; margin: 5px }\nem { color: red ! important; margin: 6px !important x }\ntt { margin: 1px,; padding: 2px }\ndd { margin: 1px / ; padding: 3px }\ndt { color: red; { nested: block } margin: 7px }\nli { color: \"unclosed\n; margin: 8px }\nol { flex-grow: 1; display: flex; display: grid; float: left }\n",
		want: `ERR null | Found an identifier where : was expected at line 1. Skipping declaration.
ERR null | Found : where an identifier was expected at line 2. Skipping declaration.
ERR null | Found 3 value(s) for color when 1 value(s) were expected at line 3. Skipping declaration.
ERR null | bogus-prop is an unrecognized CSS property at line 3. Ignoring declaration.
ERR null | Found ; where one of a number, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 5. Skipping declaration.
ERR null | Found an identifier where ; or } was expected at line 6. Skipping declaration.
ERR null | Found ; where one of a number, a +, -, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a dimension, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 7. Skipping declaration.
ERR null | Found ; where one of a number, a +, -, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a dimension, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 8. Skipping declaration.
ERR null | Found a { where an identifier was expected at line 9. Skipping declaration.
ERR null | Found an unclosed string where one of a number, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 10. Skipping declaration.
CSS3 flex-grow
ERR null | flex-grow is an unrecognized CSS property at line 11. Ignoring declaration.
CSS3 display: flex
ERR null | Value flex is not a recognized identifier at line 12. Skipping declaration.
CSS3 display: grid
ERR null | Value grid is not a recognized identifier at line 12. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=4000 s= css=4px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=4000 s= css=4px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=4000 s= css=4px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=4000 s= css=4px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=5000 s= css=5px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=5000 s= css=5px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=5000 s= css=5px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=5000 s= css=5px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color !important <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL padding-top <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL padding-right <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL padding-bottom <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
  DECL padding-left <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL padding-top <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
  DECL padding-right <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
  DECL padding-bottom <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
  DECL padding-left <VALUE_TYPE_LENGTH 5 f=3000 s= css=3px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=7000 s= css=7px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=7000 s= css=7px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=7000 s= css=7px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=7000 s= css=7px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH l1 l2 l3 l4 l5
  DECL margin-top <VALUE_TYPE_LENGTH 5 f=8000 s= css=8px ident= op=>
  DECL margin-right <VALUE_TYPE_LENGTH 5 f=8000 s= css=8px ident= op=>
  DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=8000 s= css=8px ident= op=>
  DECL margin-left <VALUE_TYPE_LENGTH 5 f=8000 s= css=8px ident= op=>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL float <VALUE_TYPE_IDENT 21 f=0 s=left css=left ident=left op=>
`,
	},
	{
		name: "eofRecovery",
		mode: "sheet",
		uri:  "",
		css:  "p { color: red; margin: 1px",
		want: `ERR null | Found end of file where an identifier was expected at line 1. Skipping declaration.
`,
	},
	{
		name: "eofInSelector",
		mode: "sheet",
		uri:  "",
		css:  "p { color: red }\ndiv p",
		want: `ERR null | Found end of file where a comma or a { was expected at line 2. Skipping ruleset.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "eofInValue",
		mode: "sheet",
		uri:  "",
		css:  "p { color: red }\nb { color:",
		want: `ERR null | Found end of file where one of a number, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 2. Skipping declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "eofInMedia",
		mode: "sheet",
		uri:  "",
		css:  "@media print { p { color: red }",
		want: `ERR null | Found end of file where one of a hex color, ., [, or : was expected at line 1. Skipping ruleset.
`,
	},
	{
		name: "eofInPage",
		mode: "sheet",
		uri:  "",
		css:  "@page { margin: 1in; @top-left { content: \"x\"",
		want: `ERR null | Found end of file where an identifier was expected at line 1. Skipping declaration.
`,
	},
	{
		name: "invalidAtRule",
		mode: "sheet",
		uri:  "",
		css:  "@foo bar;\np { color: red }\n@keyframes x { from { a: b } to { c: d } }\nb { color: blue }\n@bar { x: y } i { color: green }\n",
		want: `ERR null | Invalid at-rule at line 1. Skipping at-rule.
ERR null | Invalid at-rule at line 3. Skipping at-rule.
ERR null | Invalid at-rule at line 5. Skipping at-rule.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#008000 ident= op= rgb=0,128,0,1000>
`,
	},
	{
		name: "commentsAndCdo",
		mode: "sheet",
		uri:  "",
		css:  "/* c */ <!-- p /* x */ { /* y */ color /* z */ : /* w */ red /* v */ } --> b { color: blue } /* unterminated",
		want: `ERR null | Found a { where one of an identifier, *, a hex color, ., [, or : was expected at line 1. Skipping ruleset.
ERR null | Found / where one of a hex color, ., [, or : was expected at line 1. Skipping ruleset.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
`,
	},
	{
		name: "escapes",
		mode: "sheet",
		uri:  "",
		css:  "\\70  { col\\6fr: r\\65 d }\n.\\31 23 { color: red }\n#a\\:b { color: blue }\nP.Note { Color: Blue }\n",
		want: `RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,0 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
RULESET
  SEL [DESCENDANT_AXIS 1,0,0 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
RULESET
  SEL [DESCENDANT_AXIS 0,1,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#0000ff ident= op= rgb=0,0,255,1000>
`,
	},
	{
		name: "emptyAndWhitespace",
		mode: "sheet",
		uri:  "",
		css:  "  \n\t",
		want: ``,
	},
	{
		name: "emptyRuleset",
		mode: "sheet",
		uri:  "",
		css:  "p { }\nb { ; ; }\ni { color: red;; }\n",
		want: `RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "lineNumbers",
		mode: "sheet",
		uri:  "",
		css:  "p {\n  color: red;\n  bogus: 1;\n\n  margin: 1px 2px 3px 4px 5px;\n}\r\nb {\r\n  foo: bar;\r\n}\n",
		want: `ERR null | bogus is an unrecognized CSS property at line 2. Ignoring declaration.
ERR null | Found 5 values for margin when between 1 and 4 value(s) were expected at line 5. Skipping declaration.
ERR null | foo is an unrecognized CSS property at line 7. Ignoring declaration.
RULESET
  SEL [DESCENDANT_AXIS 0,0,1 pe= pc=] MATCH p1 p2 p3
  DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "styleAttribute",
		mode: "decl",
		uri:  "",
		css:  "color: red; margin: 1px 2px; background-image: url(a.png); bogus: x; width: 10px",
		want: `ERR style attribute | bogus is an unrecognized CSS property at line 0. Ignoring declaration.
DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
DECL margin-top <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
DECL margin-right <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
DECL margin-bottom <VALUE_TYPE_LENGTH 5 f=1000 s= css=1px ident= op=>
DECL margin-left <VALUE_TYPE_LENGTH 5 f=2000 s= css=2px ident= op=>
DECL background-image <VALUE_TYPE_STRING 20 f=0 s=a.png css=url(a.png) ident= op=>
DECL width <VALUE_TYPE_LENGTH 5 f=10000 s= css=10px ident= op=>
`,
	},
	{
		name: "styleAttributeTrailingSemicolon",
		mode: "decl",
		uri:  "",
		css:  " color: red ;",
		want: `DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "styleAttributeError",
		mode: "decl",
		uri:  "",
		css:  "color: red; }; width: 1px",
		want: `DECL color <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "styleAttributeEmpty",
		mode: "decl",
		uri:  "",
		css:  "",
		want: ``,
	},
	{
		name: "propertyValueColor",
		mode: "value:color",
		uri:  "",
		css:  "red",
		want: `VALUE <VALUE_TYPE_COLOR 25 f=0 s= css=#ff0000 ident= op= rgb=255,0,0,1000>
`,
	},
	{
		name: "propertyValueLength",
		mode: "value:width",
		uri:  "",
		css:  "10px",
		want: `VALUE <VALUE_TYPE_LENGTH 5 f=10000 s= css=10px ident= op=>
`,
	},
	{
		name: "propertyValueFontFamily",
		mode: "value:font-family",
		uri:  "",
		css:  "Arial",
		want: `VALUE <VALUE_TYPE_STRING 19 f=0 s=Arial css=Arial ident= op=>
`,
	},
	{
		name: "propertyValueError",
		mode: "value:width",
		uri:  "",
		css:  "bogus",
		want: `ERR width property value | Value bogus is not a recognized identifier at line 1. Skipping property value.
NULL
`,
	},
	{
		name: "propertyValueSyntaxError",
		mode: "value:width",
		uri:  "",
		css:  "}",
		want: `ERR width property value | Found } where one of a number, a percentage, a pixel value, an em value, an ex value, a pica value, a millimeter value, a centimeter value, an inch value, a point value, an angle value, a time value, a freq value, a string, an identifier, a URI, a hex color, or function was expected at line 1. Skipping property value.
NULL
`,
	},
}

// cssParserDumpTestAttributeResolver is the resolver of
// cssParserTestDomAttributeResolver, except that an element with an href
// attribute is a link.
type cssParserDumpTestAttributeResolver struct {
	cssParserTestDomAttributeResolver
}

func (r cssParserDumpTestAttributeResolver) IsLink(e dom.Node) bool {
	return r.GetAttributeValue(e, "href") != nil
}

type cssParserDumpTestDumper struct {
	out      strings.Builder
	elements []*dom.Element
}

func (d *cssParserDumpTestDumper) collect(e *dom.Element) {
	d.elements = append(d.elements, e)
	for n := e.GetFirstChild(); n != nil; n = n.GetNextSibling() {
		if child, ok := n.(*dom.Element); ok {
			d.collect(child)
		}
	}
}

func (d *cssParserDumpTestDumper) ruleset(indent string, rs *Ruleset) {
	d.out.WriteString(indent + "RULESET\n")
	attRes := cssParserDumpTestAttributeResolver{}
	treeRes := NewDOMTreeResolver()
	for _, s := range rs.GetFSSelectors() {
		d.out.WriteString(indent + "  SEL")
		for c := s; c != nil; c = c.GetChainedSelector() {
			fmt.Fprintf(&d.out, " [%s %d,%d,%d pe=%s pc=", c.GetAxis().ToString(), c.GetSpecificityB(), c.GetSpecificityC(), c.GetSpecificityD(), c.GetPseudoElement())
			for _, pc := range []int{SelectorVisitedPseudoclass, SelectorHoverPseudoclass, SelectorActivePseudoclass, SelectorFocusPseudoclass} {
				if c.IsPseudoClass(pc) {
					fmt.Fprintf(&d.out, "%d.", pc)
				}
			}
			d.out.WriteString("]")
		}
		d.out.WriteString(" MATCH")
		for _, e := range d.elements {
			if s.Matches(e, attRes, treeRes) {
				d.out.WriteString(" " + e.GetAttribute("id"))
			}
		}
		d.out.WriteString("\n")
	}
	d.decls(indent+"  ", rs.GetPropertyDeclarations())
}

func (d *cssParserDumpTestDumper) decls(indent string, decls []*PropertyDeclaration) {
	for _, decl := range decls {
		d.out.WriteString(indent + "DECL " + decl.GetCSSName().ToString())
		if decl.IsImportant() {
			d.out.WriteString(" !important ")
		} else {
			d.out.WriteString(" ")
		}
		d.value(decl.GetValue())
		d.out.WriteString("\n")
	}
}

func cssParserDumpTestRound(f float32) int64 {
	return int64(math.Floor(float64(f)*1000 + 0.5))
}

func (d *cssParserDumpTestDumper) value(v *PropertyValue) {
	fmt.Fprintf(&d.out, "<%s %d f=%d s=%s css=%s ident=", v.GetPropertyValueType().ToString(), v.GetPrimitiveType(),
		cssParserDumpTestRound(v.GetFloatValue()), v.GetStringValue(), v.GetCssText())
	if v.GetIdentValue() != nil {
		d.out.WriteString(v.GetIdentValue().ToString())
	}
	d.out.WriteString(" op=")
	if v.GetOperator() != nil {
		d.out.WriteString(v.GetOperator().GetName())
	}
	switch c := v.GetFSColor().(type) {
	case *FSRGBColor:
		fmt.Fprintf(&d.out, " rgb=%d,%d,%d,%d", c.GetRed(), c.GetGreen(), c.GetBlue(), cssParserDumpTestRound(c.GetAlpha()))
	case *FSCMYKColor:
		fmt.Fprintf(&d.out, " cmyk=%d,%d,%d,%d", cssParserDumpTestRound(c.GetCyan()), cssParserDumpTestRound(c.GetMagenta()),
			cssParserDumpTestRound(c.GetYellow()), cssParserDumpTestRound(c.GetBlack()))
	}
	if v.GetFunction() != nil {
		d.out.WriteString(" fn=" + v.GetFunction().GetName() + "(")
		for _, p := range v.GetFunction().GetParameters() {
			d.value(p)
		}
		d.out.WriteString(")")
	}
	if v.GetPropertyValueType() == PropertyValueTypeValueTypeList {
		d.out.WriteString(" list=(")
		for _, o := range v.GetValues() {
			switch o := o.(type) {
			case *PropertyValue:
				d.value(o)
			case string:
				// The Java dump writes the simple class name of the element.
				d.out.WriteString("{String}")
			default:
				fmt.Fprintf(&d.out, "{%T}", o)
			}
		}
		d.out.WriteString(")")
	}
	d.out.WriteString(">")
}

func TestCSSParserDump(t *testing.T) {
	document, err := dom.ParseXMLString(cssParserDumpTestXML)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range cssParserDumpTests {
		t.Run(tt.name, func(t *testing.T) {
			d := &cssParserDumpTestDumper{}
			d.collect(document.GetDocumentElement())

			parser := NewCSSParserWithCss3FeatureListener(CSSErrorHandlerFunc(func(uri string, message string) {
				if uri == "" {
					uri = "null"
				}
				d.out.WriteString("ERR " + uri + " | " + message + "\n")
			}), func(feature string) {
				d.out.WriteString("CSS3 " + feature + "\n")
			})
			if strings.Contains(tt.mode, "cmyk") {
				parser.SetSupportCMYKColors(true)
			}

			switch {
			case strings.HasPrefix(tt.mode, "decl"):
				rs := parser.ParseDeclaration(StylesheetInfoOriginAuthor, tt.css)
				d.decls("", rs.GetPropertyDeclarations())
			case strings.HasPrefix(tt.mode, "value:"):
				cssName := CSSNameGetByPropertyName(strings.TrimPrefix(tt.mode, "value:"))
				v := parser.ParsePropertyValue(cssName, StylesheetInfoOriginUser, tt.css)
				if v == nil {
					d.out.WriteString("NULL\n")
				} else {
					d.out.WriteString("VALUE ")
					d.value(v)
					d.out.WriteString("\n")
				}
			default:
				sheet, err := parser.ParseStylesheet(tt.uri, StylesheetInfoOriginAuthor, strings.NewReader(tt.css))
				if err != nil {
					t.Fatal(err)
				}
				for _, info := range sheet.GetImportRules() {
					d.out.WriteString("IMPORT " + info.GetUri() + " media=" + strings.Join(info.GetMedia(), ",") + "\n")
				}
				for _, r := range sheet.GetFontFaceRules() {
					fmt.Fprintf(&d.out, "FONTFACE family=%v weight=%v style=%v\n", r.HasFontFamily(), r.HasFontWeight(), r.HasFontStyle())
				}
				for _, o := range sheet.GetContents() {
					switch rule := o.(type) {
					case *Ruleset:
						d.ruleset("", rule)
					case *MediaRule:
						fmt.Fprintf(&d.out, "MEDIA print=%v screen=%v tv=%v\n", rule.Matches("print"), rule.Matches("screen"), rule.Matches("tv"))
						for _, rs := range rule.GetContents() {
							d.ruleset("  ", rs)
						}
					case *PageRule:
						d.out.WriteString("PAGE name=" + rule.GetName() + " pseudo=" + rule.GetPseudoPage() + "\n")
						d.decls("  ", rule.GetRuleset().GetPropertyDeclarations())
						byName := map[string]*MarginBoxName{}
						var names []string
						for n := range rule.GetMarginBoxes() {
							names = append(names, n.ToString())
							byName[n.ToString()] = n
						}
						sort.Strings(names)
						for _, n := range names {
							d.out.WriteString("  MARGIN " + n + "\n")
							d.decls("    ", rule.GetMarginBoxProperties(byName[n]))
						}
					}
				}
			}

			if got := d.out.String(); got != tt.want {
				t.Errorf("css:\n%s\ngot:\n%s\nwant:\n%s", tt.css, got, tt.want)
			}
		})
	}
}

func TestCSSParserReturnsReadError(t *testing.T) {
	parser := NewCSSParser(CSSErrorHandlerFunc(func(uri string, message string) {
		t.Errorf("unexpected error message: %s", message)
	}))
	stylesheet, err := parser.ParseStylesheet("", StylesheetInfoOriginAuthor, lexerTestFailingReader{})
	if err == nil || err.Error() != "read failed" {
		t.Errorf("error = %v, want read failed", err)
	}
	if stylesheet != nil {
		t.Errorf("stylesheet = %v, want nil", stylesheet)
	}
}
