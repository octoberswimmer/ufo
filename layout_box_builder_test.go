// Tests for layout_box_builder.go. The expected box trees were produced by
// running the Java BoxBuilder (flying-saucer-core 10.6.0-SNAPSHOT) over the
// same documents: BoxBuilder.createRootBox, then ensureChildren on the root,
// which builds the whole tree because createChildren calls ensureChildren on
// every block box it creates.

package ufo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

type boxBuilderTestUserAgent struct {
	baseURL string
}

func (u *boxBuilderTestUserAgent) GetCSSResource(uri string) *CSSResource     { return nil }
func (u *boxBuilderTestUserAgent) GetImageResource(uri string) *ImageResource { return nil }
func (u *boxBuilderTestUserAgent) GetXMLResource(uri string) *XMLResource     { return nil }
func (u *boxBuilderTestUserAgent) GetBinaryResource(uri string) []byte        { return nil }
func (u *boxBuilderTestUserAgent) IsVisited(uri string) bool                  { return false }
func (u *boxBuilderTestUserAgent) SetBaseURL(url string)                      { u.baseURL = url }
func (u *boxBuilderTestUserAgent) GetBaseURL() string                         { return u.baseURL }
func (u *boxBuilderTestUserAgent) ResolveURI(uri string) string               { return uri }

type boxBuilderTestUserInterface struct{}

func (boxBuilderTestUserInterface) IsHover(e *dom.Element) bool  { return false }
func (boxBuilderTestUserInterface) IsActive(e *dom.Element) bool { return false }
func (boxBuilderTestUserInterface) IsFocus(e *dom.Element) bool  { return false }

// boxBuilderTestBuild parses source as XML and builds the box tree of the
// document.
func boxBuilderTestBuild(t *testing.T, source string) (*SharedContext, BlockBoxI) {
	t.Helper()
	doc, err := dom.ParseXMLString(source)
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSharedContextWithUac(&boxBuilderTestUserAgent{})
	nsh := NewXhtmlNamespaceHandler()
	sc.SetNamespaceHandler(nsh)
	sc.GetCss().SetDocumentContext(sc, nsh, doc, boxBuilderTestUserInterface{})
	c := sc.NewLayoutContextInstance(nil)
	root := BoxBuilderCreateRootBox(c, doc)
	root.EnsureChildren(c)
	return sc, root
}

func boxBuilderTestElementName(e *dom.Element) string {
	if e == nil {
		return "-"
	}
	return e.GetNodeName()
}

func boxBuilderTestDumpInlineBox(sb *strings.Builder, iB *InlineBox, indent string) {
	sb.WriteString(indent + "InlineBox el=" + boxBuilderTestElementName(iB.GetElement()))
	if iB.IsStartsHere() {
		sb.WriteString(" S")
	} else {
		sb.WriteString(" s")
	}
	if iB.IsEndsHere() {
		sb.WriteString("E")
	} else {
		sb.WriteString("e")
	}
	if iB.IsRemovableWhitespace() {
		sb.WriteString(" ws")
	}
	if iB.GetPseudoElementOrClass() != "" {
		sb.WriteString(" pe=" + iB.GetPseudoElementOrClass())
	}
	if iB.GetContentFunction() != nil {
		sb.WriteString(" fn")
	}
	sb.WriteString(" text=" + layoutTestEscape(iB.GetText()) + "\n")
}

func boxBuilderTestDumpBox(sb *strings.Builder, b BoxI, indent string) {
	var className string
	switch b.(type) {
	case *AnonymousBlockBox:
		className = "AnonymousBlockBox"
	case *TableBox:
		className = "TableBox"
	case *TableSectionBox:
		className = "TableSectionBox"
	case *TableRowBox:
		className = "TableRowBox"
	case *TableCellBox:
		className = "TableCellBox"
	case *BlockBox:
		className = "BlockBox"
	default:
		className = fmt.Sprintf("%T", b)
	}
	sb.WriteString(indent + className + " el=" + boxBuilderTestElementName(b.GetElement()) +
		" display=" + b.GetStyle().GetIdent(CSSNameDisplay).String())
	if b.IsAnonymous() {
		sb.WriteString(" anon")
	}
	if b.GetPseudoElementOrClass() != "" {
		sb.WriteString(" pe=" + b.GetPseudoElementOrClass())
	}
	bb, isBlock := b.(BlockBoxI)
	if isBlock {
		sb.WriteString(" content=" + bb.GetChildrenContentType().String())
		if bb.GetFloatedBoxData() != nil {
			sb.WriteString(" floated")
		}
		if bb.IsFromCaptionedTable() {
			sb.WriteString(" captioned")
		}
		if bb.GetListCounter() != 0 {
			sb.WriteString(fmt.Sprintf(" counter=%d", bb.GetListCounter()))
		}
		if bb.GetFirstLineStyle() != nil {
			sb.WriteString(" first-line")
		}
		if bb.GetFirstLetterStyle() != nil {
			sb.WriteString(" first-letter")
		}
	}
	if s, ok := b.(*TableSectionBox); ok {
		if s.IsHeader() {
			sb.WriteString(" header")
		}
		if s.IsFooter() {
			sb.WriteString(" footer")
		}
	}
	if table, ok := b.(*TableBox); ok {
		sb.WriteString(" cols=")
		for _, col := range table.GetStyleColumns() {
			sb.WriteString(boxBuilderTestElementName(col.GetElement()))
			if col.GetParent() != nil {
				sb.WriteString("^" + boxBuilderTestElementName(col.GetParent().GetElement()))
			}
			sb.WriteString(",")
		}
	}
	if a, ok := b.(*AnonymousBlockBox); ok {
		sb.WriteString(" open=")
		if a.GetOpenInlineBoxes() == nil {
			sb.WriteString("null")
		} else {
			for _, iB := range a.GetOpenInlineBoxes() {
				sb.WriteString(boxBuilderTestElementName(iB.GetElement()) + ",")
			}
		}
	}
	sb.WriteString("\n")
	if isBlock {
		for _, s := range bb.GetInlineContent() {
			if iB, ok := s.(*InlineBox); ok {
				boxBuilderTestDumpInlineBox(sb, iB, indent+"  ")
			} else {
				boxBuilderTestDumpBox(sb, s.(BoxI), indent+"  ")
			}
		}
	}
	for _, child := range b.GetChildren() {
		boxBuilderTestDumpBox(sb, child, indent+"  ")
	}
}

// TestBoxBuilderBoxTreeAgainstJava: a line of the expected tree is the class
// of the box, its element, its display value and its flags. An InlineBox line
// has S/s for startsHere, E/e for endsHere, ws for removable whitespace, the
// pseudo-element, and the text with the escapes of layoutTestUnescape.
func TestBoxBuilderBoxTreeAgainstJava(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"anonymous_blocks", `<html><head><style>em { text-transform: uppercase; } p.first:first-line { color: red; } p.first:first-letter { color: blue; }</style></head><body>
<div>text before <span>span start <b>bold</b> <div>inner block</div> after block</span> tail <p class="first">para</p>   </div>
<div>   <p>one</p>   <p>two</p>   </div>
<div>only <em>inline</em> content<br/>here</div>
<div></div>
<div>   </div>
<div><span></span></div>
<div><span><i></i></span><p>x</p></div>
</body></html>
`, `BlockBox el=html display=block content=BLOCK
  BlockBox el=body display=block content=BLOCK
    BlockBox el=div display=block content=BLOCK
      AnonymousBlockBox el=div display=block anon content=INLINE open=null
        InlineBox el=- Se text=text before 
        InlineBox el=span Se text=span start 
        InlineBox el=b SE text=bold
        InlineBox el=span se ws text= 
      BlockBox el=div display=block content=INLINE
        InlineBox el=- SE text=inner block
      AnonymousBlockBox el=div display=block anon content=INLINE open=-,span,
        InlineBox el=span sE text= after block
        InlineBox el=- se text= tail 
      BlockBox el=p display=block content=INLINE first-line first-letter
        InlineBox el=- SE text=para
    BlockBox el=div display=block content=BLOCK
      BlockBox el=p display=block content=INLINE
        InlineBox el=- SE text=one
      BlockBox el=p display=block content=INLINE
        InlineBox el=- SE text=two
    BlockBox el=div display=block content=INLINE
      InlineBox el=- Se text=only 
      InlineBox el=em SE text=INLINE
      InlineBox el=- se text= content
      InlineBox el=br SE pe=before text=\u000a
      InlineBox el=br SE ws text=
      InlineBox el=- sE text=here
    BlockBox el=div display=block content=EMPTY
    BlockBox el=div display=block content=EMPTY
    BlockBox el=div display=block content=INLINE
      InlineBox el=span SE ws text=
    BlockBox el=div display=block content=BLOCK
      AnonymousBlockBox el=div display=block anon content=INLINE open=null
        InlineBox el=span Se ws text=
        InlineBox el=i SE ws text=
        InlineBox el=span sE ws text=
      BlockBox el=p display=block content=INLINE
        InlineBox el=- SE text=x
unsupported=[]
`},
		{"tables", `<html><head><style>.cell { display: table-cell; } .row { display: table-row; } .itab { display: inline-table; } .ftab { float: left; }</style></head><body>
<table><caption>Top</caption><colgroup><col/><col/></colgroup><col/><colgroup span="2"></colgroup><tfoot><tr><td>f</td></tr></tfoot><tbody><tr><td>b1</td><td> b2 <span>s</span></td></tr></tbody><thead><tr><th>h</th></tr></thead><caption style="caption-side: bottom">Bottom</caption><tbody><tr><td>second body</td></tr></tbody></table>
<div><span class="cell">c1</span><span class="cell">c2</span> loose text <span class="row"><span class="cell">c3</span></span></div>
<table><tr><td>no tbody</td></tr>stray text<td>stray cell</td></table>
<table><tbody>text in tbody<tr>text in row<td>cell</td></tr></tbody></table>
<p>inline <span>before <span class="cell">cell in span</span> after</span> tail</p>
<span class="itab"><span class="row"><span class="cell">it</span></span></span>
<table class="ftab"><caption>floated</caption><tr><td>x</td></tr></table>
<table class="ftab"><tr><td>floated no caption</td></tr></table>
<table><thead><tr><td>h1</td></tr></thead><thead><tr><td>h2</td></tr></thead><tfoot><tr><td>f1</td></tr></tfoot><tfoot><tr><td>f2</td></tr></tfoot></table>
</body></html>
`, `BlockBox el=html display=block content=BLOCK
  BlockBox el=body display=block content=BLOCK
    BlockBox el=table display=block anon content=BLOCK captioned
      BlockBox el=caption display=table-caption content=INLINE
        InlineBox el=- SE text=Top
      TableBox el=table display=table content=BLOCK cols=col^colgroup,col^colgroup,col,colgroup,
        TableSectionBox el=thead display=table-header-group content=BLOCK header
          TableRowBox el=tr display=table-row content=BLOCK
            TableCellBox el=th display=table-cell content=INLINE
              InlineBox el=- SE text=h
        TableSectionBox el=tbody display=table-row-group content=BLOCK
          TableRowBox el=tr display=table-row content=BLOCK
            TableCellBox el=td display=table-cell content=INLINE
              InlineBox el=- SE text=b1
            TableCellBox el=td display=table-cell content=INLINE
              InlineBox el=- SE text= b2 
              InlineBox el=span SE text=s
        TableSectionBox el=tbody display=table-row-group content=BLOCK
          TableRowBox el=tr display=table-row content=BLOCK
            TableCellBox el=td display=table-cell content=INLINE
              InlineBox el=- SE text=second body
        TableSectionBox el=tfoot display=table-footer-group content=BLOCK footer
          TableRowBox el=tr display=table-row content=BLOCK
            TableCellBox el=td display=table-cell content=INLINE
              InlineBox el=- SE text=f
      BlockBox el=caption display=table-caption content=INLINE
        InlineBox el=- SE text=Bottom
    BlockBox el=div display=block content=BLOCK
      TableBox el=span display=table anon content=BLOCK cols=
        TableSectionBox el=span display=table-row-group anon content=BLOCK
          TableRowBox el=span display=table-row anon content=BLOCK
            TableCellBox el=span display=table-cell content=INLINE
              InlineBox el=- SE text=c1
            TableCellBox el=span display=table-cell content=INLINE
              InlineBox el=- SE text=c2
      AnonymousBlockBox el=div display=block anon content=INLINE open=null
        InlineBox el=- SE text= loose text 
      TableBox el=span display=table anon content=BLOCK cols=
        TableSectionBox el=span display=table-row-group anon content=BLOCK
          TableRowBox el=span display=table-row content=BLOCK
            TableCellBox el=span display=table-cell content=INLINE
              InlineBox el=- SE text=c3
    TableBox el=table display=table content=BLOCK cols=
      TableSectionBox el=table display=table-row-group anon content=BLOCK
        TableRowBox el=tr display=table-row content=BLOCK
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=no tbody
        TableRowBox el=table display=table-row anon content=BLOCK
          TableCellBox el=table display=table-cell anon content=INLINE
            InlineBox el=- SE text=stray text
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=stray cell
    TableBox el=table display=table content=BLOCK cols=
      TableSectionBox el=tbody display=table-row-group content=BLOCK
        TableRowBox el=tbody display=table-row anon content=BLOCK
          TableCellBox el=tbody display=table-cell anon content=INLINE
            InlineBox el=- SE text=text in tbody
        TableRowBox el=tr display=table-row content=BLOCK
          TableCellBox el=tr display=table-cell anon content=INLINE
            InlineBox el=- SE text=text in row
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=cell
    BlockBox el=p display=block content=BLOCK
      AnonymousBlockBox el=p display=block anon content=INLINE open=null
        InlineBox el=- Se text=inline 
        InlineBox el=span Se text=before 
      TableBox el=span display=table anon content=BLOCK cols=
        TableSectionBox el=span display=table-row-group anon content=BLOCK
          TableRowBox el=span display=table-row anon content=BLOCK
            TableCellBox el=span display=table-cell content=INLINE
              InlineBox el=- SE text=cell in span
      AnonymousBlockBox el=p display=block anon content=INLINE open=-,span,
        InlineBox el=span sE text= after
        InlineBox el=- sE text= tail
    AnonymousBlockBox el=body display=block anon content=INLINE open=-,
      InlineBox el=- se ws text= 
      TableBox el=span display=inline-table content=BLOCK cols=
        TableSectionBox el=span display=table-row-group anon content=BLOCK
          TableRowBox el=span display=table-row content=BLOCK
            TableCellBox el=span display=table-cell content=INLINE
              InlineBox el=- SE text=it
      InlineBox el=- se ws text= 
      BlockBox el=table display=block anon content=BLOCK floated captioned
        BlockBox el=caption display=table-caption content=INLINE
          InlineBox el=- SE text=floated
        TableBox el=table display=table content=BLOCK cols=
          TableSectionBox el=table display=table-row-group anon content=BLOCK
            TableRowBox el=tr display=table-row content=BLOCK
              TableCellBox el=td display=table-cell content=INLINE
                InlineBox el=- SE text=x
      InlineBox el=- se ws text=
      TableBox el=table display=table content=BLOCK floated cols=
        TableSectionBox el=table display=table-row-group anon content=BLOCK
          TableRowBox el=tr display=table-row content=BLOCK
            TableCellBox el=td display=table-cell content=INLINE
              InlineBox el=- SE text=floated no caption
      InlineBox el=- se ws text=
    TableBox el=table display=table content=BLOCK cols=
      TableSectionBox el=thead display=table-header-group content=BLOCK header
        TableRowBox el=tr display=table-row content=BLOCK
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=h1
      TableSectionBox el=thead display=table-header-group content=BLOCK
        TableRowBox el=tr display=table-row content=BLOCK
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=h2
      TableSectionBox el=tfoot display=table-footer-group content=BLOCK
        TableRowBox el=tr display=table-row content=BLOCK
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=f2
      TableSectionBox el=tfoot display=table-footer-group content=BLOCK footer
        TableRowBox el=tr display=table-row content=BLOCK
          TableCellBox el=td display=table-cell content=INLINE
            InlineBox el=- SE text=f1
unsupported=[]
`},
		{"generated_content", `<html><head><style>
body { counter-reset: chapter; }
h1 { counter-increment: chapter; counter-reset: section; }
h1:before { content: "Chapter " counter(chapter, upper-roman) ". "; }
h2 { counter-increment: section; }
h2:before { content: counter(chapter) "." counter(section, lower-alpha) " "; }
h2:after { content: " [" attr(title) "]" attr(missing); }
q { quotes: "&lt;&lt;" "&gt;&gt;"; }
q:before { content: open-quote; }
q:after { content: close-quote; }
q.none { quotes: none; }
p.block:before { content: "block before"; display: block; }
p.tab:before { content: "table before"; display: table; }
p.cell:before { content: "cell before"; display: table-cell; }
p.none:before { content: none; }
p.hidden:before { content: "hidden"; display: none; }
p.normal:before { content: normal; }
p.upper:before { content: "shout "; text-transform: uppercase; }
p.float:before { content: "floated"; float: left; }
ol.nested { counter-reset: item; }
ol.nested li { counter-increment: item; }
ol.nested li:before { content: counters(item, ".") " " counters(item, "-", lower-roman) " "; }
p.pages:after { content: counter(page) " of " counter(pages); }
p.bad:before { content: counter(chapter, bogus-style) counter(a, b, c) counters(x) element(foo); }
p.reset:before { counter-reset: chapter 40; }
</style></head><body>
<h1>One</h1><h2 title="t1">First</h2><h2 title="t2">Second</h2>
<h1>Two</h1><h2>Third</h2>
<p>He said <q>hello <q>nested</q></q> and <q class="none">none</q></p>
<p class="block">b</p><p class="tab">t</p><p class="cell">c</p><p class="none">n</p><p class="hidden">h</p><p class="normal">n</p><p class="upper">u</p><p class="float">f</p>
<ol class="nested"><li>a<ol class="nested"><li>a.a</li><li>a.b</li></ol></li><li>b</li></ol>
<p class="pages">pages</p>
<p class="bad">bad</p>
<p class="reset">reset</p><h1>After reset</h1>
</body></html>
`, `BlockBox el=html display=block content=BLOCK
  BlockBox el=body display=block content=BLOCK
    BlockBox el=h1 display=block content=INLINE
      InlineBox el=h1 SE pe=before text=Chapter 
      InlineBox el=h1 SE pe=before text=I
      InlineBox el=h1 SE pe=before text=. 
      InlineBox el=- SE text=One
    BlockBox el=h2 display=block content=INLINE
      InlineBox el=h2 SE pe=before text=1
      InlineBox el=h2 SE pe=before text=.
      InlineBox el=h2 SE pe=before text=a
      InlineBox el=h2 SE ws pe=before text= 
      InlineBox el=- SE text=First
      InlineBox el=h2 SE pe=after text= [
      InlineBox el=h2 SE pe=after text=t1
      InlineBox el=h2 SE pe=after text=]
      InlineBox el=h2 SE ws pe=after text=
    BlockBox el=h2 display=block content=INLINE
      InlineBox el=h2 SE pe=before text=1
      InlineBox el=h2 SE pe=before text=.
      InlineBox el=h2 SE pe=before text=b
      InlineBox el=h2 SE ws pe=before text= 
      InlineBox el=- SE text=Second
      InlineBox el=h2 SE pe=after text= [
      InlineBox el=h2 SE pe=after text=t2
      InlineBox el=h2 SE pe=after text=]
      InlineBox el=h2 SE ws pe=after text=
    BlockBox el=h1 display=block content=INLINE
      InlineBox el=h1 SE pe=before text=Chapter 
      InlineBox el=h1 SE pe=before text=II
      InlineBox el=h1 SE pe=before text=. 
      InlineBox el=- SE text=Two
    BlockBox el=h2 display=block content=INLINE
      InlineBox el=h2 SE pe=before text=2
      InlineBox el=h2 SE pe=before text=.
      InlineBox el=h2 SE pe=before text=a
      InlineBox el=h2 SE ws pe=before text= 
      InlineBox el=- SE text=Third
      InlineBox el=h2 SE pe=after text= [
      InlineBox el=h2 SE ws pe=after text=
      InlineBox el=h2 SE pe=after text=]
      InlineBox el=h2 SE ws pe=after text=
    BlockBox el=p display=block content=INLINE
      InlineBox el=- Se text=He said 
      InlineBox el=q SE pe=before text=<<
      InlineBox el=q Se text=hello 
      InlineBox el=q SE pe=before text=<<
      InlineBox el=q SE text=nested
      InlineBox el=q SE pe=after text=>>
      InlineBox el=q sE ws text=
      InlineBox el=q SE pe=after text=>>
      InlineBox el=- sE text= and 
      InlineBox el=q SE text=none
    BlockBox el=p display=block content=BLOCK
      BlockBox el=p display=block pe=before content=INLINE
        InlineBox el=- SE pe=before text=block before
      AnonymousBlockBox el=p display=block anon content=INLINE open=null
        InlineBox el=- SE text=b
    BlockBox el=p display=block content=BLOCK
      BlockBox el=p display=block pe=before content=INLINE
        InlineBox el=- SE pe=before text=table before
      AnonymousBlockBox el=p display=block anon content=INLINE open=null
        InlineBox el=- SE text=t
    BlockBox el=p display=block content=BLOCK
      TableBox el=p display=table anon content=BLOCK cols=
        TableSectionBox el=p display=table-row-group anon content=BLOCK
          TableRowBox el=p display=table-row anon content=BLOCK
            TableCellBox el=p display=table-cell pe=before content=INLINE
              InlineBox el=- SE pe=before text=cell before
      AnonymousBlockBox el=p display=block anon content=INLINE open=null
        InlineBox el=- SE text=c
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text=n
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text=h
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text=n
    BlockBox el=p display=block content=INLINE
      InlineBox el=p SE pe=before text=SHOUT 
      InlineBox el=- SE text=u
    BlockBox el=p display=block content=INLINE
      BlockBox el=p display=inline pe=before content=INLINE floated
        InlineBox el=- SE pe=before text=floated
      InlineBox el=- SE text=f
    BlockBox el=ol display=block content=BLOCK
      BlockBox el=li display=list-item content=BLOCK counter=1
        AnonymousBlockBox el=li display=block anon content=INLINE open=null
          InlineBox el=li SE pe=before text=1
          InlineBox el=li SE ws pe=before text= 
          InlineBox el=li SE pe=before text=i
          InlineBox el=li SE ws pe=before text= 
          InlineBox el=- SE text=a
        BlockBox el=ol display=block content=BLOCK
          BlockBox el=li display=list-item content=INLINE counter=1
            InlineBox el=li SE pe=before text=1.1
            InlineBox el=li SE ws pe=before text= 
            InlineBox el=li SE pe=before text=i-i
            InlineBox el=li SE ws pe=before text= 
            InlineBox el=- SE text=a.a
          BlockBox el=li display=list-item content=INLINE counter=2
            InlineBox el=li SE pe=before text=1.2
            InlineBox el=li SE ws pe=before text= 
            InlineBox el=li SE pe=before text=i-ii
            InlineBox el=li SE ws pe=before text= 
            InlineBox el=- SE text=a.b
      BlockBox el=li display=list-item content=INLINE counter=2
        InlineBox el=li SE pe=before text=2
        InlineBox el=li SE ws pe=before text= 
        InlineBox el=li SE pe=before text=ii
        InlineBox el=li SE ws pe=before text= 
        InlineBox el=- SE text=b
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text=pages
      InlineBox el=p SE pe=after text= of 
    BlockBox el=p display=block content=INLINE
      InlineBox el=p SE pe=before text=2
      InlineBox el=- SE text=bad
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text=reset
    BlockBox el=h1 display=block content=INLINE
      InlineBox el=h1 SE pe=before text=Chapter 
      InlineBox el=h1 SE pe=before text=III
      InlineBox el=h1 SE pe=before text=. 
      InlineBox el=- SE text=After reset
unsupported=[]
`},
		{"lists_and_misc", `<html><head><style>
pre.p { white-space: pre; } .pw { white-space: pre-wrap; } .pl { white-space: pre-line; } .nw { white-space: nowrap; }
.sc { font-variant: small-caps; } .cap { text-transform: capitalize; } .low { text-transform: lowercase; }
.gone { display: none; } .abs { position: absolute; } .fl { float: right; } .ib { display: inline-block; }
.run { position: running(header); }
</style></head><body>
<ol start="5"><li>five</li><li value="10">ten</li><li value="x">bad</li><li value="+3">plus</li><li value="-2">minus</li><li value="99999999999">big</li></ol>
<ul><li>bullet</li></ul>
<ol start=""><li>empty start</li></ol>
<pre class="p">  keep	tab 
 lines  </pre>
<p class="pw">  pre	wrap 
 text  </p>
<p class="pl">  pre	line 
 text  </p>
<p class="nw">  no 
 wrap  </p>
<p><span class="sc">small caps</span> <span class="cap">capitalize these words</span> <span class="low">LOWER</span></p>
<p>a <span class="gone">hidden</span> b <span class="abs">absolute</span> c <span class="fl">float</span> d <span class="ib">inline block</span> e</p>
<p>  <span class="fl">f</span>   </p>
<p>  <span class="ib">ib</span>   </p>
<div class="run">running header</div>
<p>x<![CDATA[ cdata <text> ]]>y</p>
<header>h</header><nav>n</nav><MAIN>m</MAIN><section><article>a<video/></article></section><summary>s</summary><custom>c</custom>
</body></html>
`, `BlockBox el=html display=block content=BLOCK
  BlockBox el=body display=block content=BLOCK
    BlockBox el=ol display=block content=BLOCK
      BlockBox el=li display=list-item content=INLINE counter=5
        InlineBox el=- SE text=five
      BlockBox el=li display=list-item content=INLINE counter=10
        InlineBox el=- SE text=ten
      BlockBox el=li display=list-item content=INLINE counter=11
        InlineBox el=- SE text=bad
      BlockBox el=li display=list-item content=INLINE counter=3
        InlineBox el=- SE text=plus
      BlockBox el=li display=list-item content=INLINE counter=-2
        InlineBox el=- SE text=minus
      BlockBox el=li display=list-item content=INLINE counter=-1
        InlineBox el=- SE text=big
    BlockBox el=ul display=block content=BLOCK
      BlockBox el=li display=list-item content=INLINE counter=1
        InlineBox el=- SE text=bullet
    BlockBox el=ol display=block content=BLOCK
      BlockBox el=li display=list-item content=INLINE counter=1
        InlineBox el=- SE text=empty start
    BlockBox el=pre display=block content=INLINE
      InlineBox el=- SE text=  keep        tab\u000a lines  
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text=  pre        wrap \u000a text  
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text= pre line \u000a text 
    BlockBox el=p display=block content=INLINE
      InlineBox el=- SE text= no wrap 
    BlockBox el=p display=block content=INLINE
      InlineBox el=span SE text=SMALL CAPS
      InlineBox el=- Se ws text= 
      InlineBox el=span SE text=Capitalize These Words
      InlineBox el=- sE ws text= 
      InlineBox el=span SE text=lower
    BlockBox el=p display=block content=INLINE
      InlineBox el=- Se text=a 
      InlineBox el=- se text=b 
      BlockBox el=span display=inline content=INLINE
        InlineBox el=- SE text=absolute
      InlineBox el=- se text=c 
      BlockBox el=span display=inline content=INLINE floated
        InlineBox el=- SE text=float
      InlineBox el=- se text=d 
      BlockBox el=span display=inline-block content=INLINE
        InlineBox el=- SE text=inline block
      InlineBox el=- sE text= e
    BlockBox el=p display=block content=INLINE
      BlockBox el=span display=inline content=INLINE floated
        InlineBox el=- SE text=f
    BlockBox el=p display=block content=INLINE
      InlineBox el=- Se ws text= 
      BlockBox el=span display=inline-block content=INLINE
        InlineBox el=- SE text=ib
      InlineBox el=- sE ws text= 
    AnonymousBlockBox el=body display=block anon content=INLINE open=-,
      BlockBox el=div display=block content=INLINE
        InlineBox el=- SE text=running header
    BlockBox el=p display=block content=INLINE
      InlineBox el=- Se text=x
      InlineBox el=- se text= cdata <text> 
      InlineBox el=- sE text=y
    AnonymousBlockBox el=body display=block anon content=INLINE open=-,
      InlineBox el=- se ws text= 
      InlineBox el=header SE text=h
      InlineBox el=nav SE text=n
      InlineBox el=MAIN SE text=m
      InlineBox el=section Se ws text=
      InlineBox el=article Se text=a
      InlineBox el=video SE ws text=
      InlineBox el=article sE ws text=
      InlineBox el=section sE ws text=
      InlineBox el=summary SE text=s
      InlineBox el=custom SE text=c
      InlineBox el=- sE ws text= 
unsupported=[header, nav, MAIN, section, article, video, summary]
`},
		{"root_table", `<html style="display: table"><body style="display: table-cell">cell body <p>p</p></body></html>
`, `TableBox el=html display=table content=BLOCK cols=
  TableSectionBox el=html display=table-row-group anon content=BLOCK
    TableRowBox el=html display=table-row anon content=BLOCK
      TableCellBox el=body display=table-cell content=BLOCK
        AnonymousBlockBox el=body display=block anon content=INLINE open=null
          InlineBox el=- SE text=cell body 
        BlockBox el=p display=block content=INLINE
          InlineBox el=- SE text=p
unsupported=[]
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc, root := boxBuilderTestBuild(t, tt.source)
			var sb strings.Builder
			boxBuilderTestDumpBox(&sb, root, "")
			sb.WriteString("unsupported=[" + strings.Join(sc.GetUnsupportedTags(), ", ") + "]\n")
			if got := sb.String(); got != tt.want {
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(tt.want, "\n")
				for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
					var g, w string
					if i < len(gotLines) {
						g = gotLines[i]
					}
					if i < len(wantLines) {
						w = wantLines[i]
					}
					if g != w {
						t.Fatalf("line %d:\n got %q\nwant %q\nfull tree:\n%s", i+1, g, w, got)
					}
				}
			}
		})
	}
}
