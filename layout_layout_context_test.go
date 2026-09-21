package ufo

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

type layoutContextTestUserAgent struct {
	baseURL string
}

func (u *layoutContextTestUserAgent) GetCSSResource(uri string) *CSSResource     { return nil }
func (u *layoutContextTestUserAgent) GetImageResource(uri string) *ImageResource { return nil }
func (u *layoutContextTestUserAgent) GetXMLResource(uri string) *XMLResource     { return nil }
func (u *layoutContextTestUserAgent) GetBinaryResource(uri string) []byte        { return nil }
func (u *layoutContextTestUserAgent) IsVisited(uri string) bool                  { return false }
func (u *layoutContextTestUserAgent) SetBaseURL(url string)                      { u.baseURL = url }
func (u *layoutContextTestUserAgent) GetBaseURL() string                         { return u.baseURL }
func (u *layoutContextTestUserAgent) ResolveURI(uri string) string               { return uri }

type layoutContextTestUserInterface struct{}

func (layoutContextTestUserInterface) IsHover(e *dom.Element) bool  { return false }
func (layoutContextTestUserInterface) IsActive(e *dom.Element) bool { return false }
func (layoutContextTestUserInterface) IsFocus(e *dom.Element) bool  { return false }

// layoutContextTestSharedContext returns a SharedContext whose StyleReference
// matches styles for the document.
func layoutContextTestSharedContext(t *testing.T, source string) (*SharedContext, *dom.Document) {
	t.Helper()
	doc, err := dom.ParseXMLString(source)
	if err != nil {
		t.Fatal(err)
	}
	sc := NewSharedContextWithUac(&layoutContextTestUserAgent{})
	nsh := NewXhtmlNamespaceHandler()
	sc.SetNamespaceHandler(nsh)
	sc.GetCss().SetDocumentContext(sc, nsh, doc, layoutContextTestUserInterface{})
	return sc, doc
}

const layoutContextTestCounterDocument = `<html><head><style>
body { counter-reset: sec 3 fig; }
h1 { counter-increment: sec; }
h2 { counter-increment: sub 2 undeclared; counter-reset: fig 7; }
div.r { counter-reset: sec; }
p { counter-increment: list-item 5; }
span.li { display: list-item; }
</style></head><body>
<h1>a</h1>
<h2>b</h2>
<ol><li>one</li><li value="9">two<ol start="4"><li>n1</li><li>n2</li></ol></li><li>three</li></ol>
<div class="r"><h1>c</h1><h1>d</h1><div class="r"><h1>e</h1><h2>f</h2></div><h1>g</h1></div>
<h1>h</h1>
<p>para</p><p>para2</p>
<ul><li>u1</li><span class="li">s</span><li>u2</li></ul>
</body></html>`

// layoutContextTestCounterExpected is the output of the same walk run against
// the Java LayoutContext (flying-saucer-core 10.6.0-SNAPSHOT): for every
// element in document order, resolveCounters and then, for each counter name
// that is not skipped, getCurrentCounterValue followed by
// getCurrentCounterValues. Some names are skipped per element because asking
// for a counter that is not in scope creates it.
const layoutContextTestCounterExpected = `
html list-item=0[0] sec=0[0] sub=0[0] fig=0[0] undeclared=0[0]
 head list-item=0[0] sec=0[0] sub=0[0] fig=0[0] undeclared=0[0]
  style list-item=0[0] sec=0[0] sub=0[0] fig=0[0]
 body list-item=0[0] sec=3[0, 3] sub=0[0] fig=0[0, 0] undeclared=0[0]
  h1 list-item=0[0] sec=4[0, 4] sub=0[0] fig=0[0, 0]
  h2 list-item=0[0] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
  ol list-item=0[0] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
   li list-item=1[0, 1] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
   li list-item=9[0, 9] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
    ol list-item=8[0, 9, 8] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
     li list-item=4[0, 9, 8, 4] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
     li list-item=5[0, 9, 8, 5] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
   li list-item=10[0, 10] sec=4[0, 4] sub=2[2] fig=7[0, 0, 7]
  div undeclared=1[1]
   h1 list-item=0[0] sec=1[0, 4, 1] sub=2[2] fig=7[0, 0, 7]
   h1 list-item=0[0] sec=2[0, 4, 2] sub=2[2] fig=7[0, 0, 7]
   div undeclared=1[1]
    h1 list-item=0[0] sec=1[0, 4, 2, 1] sub=2[2] fig=7[0, 0, 7]
    h2 list-item=0[0] sec=1[0, 4, 2, 1] sub=4[4] fig=7[0, 0, 7, 7]
   h1 list-item=0[0] sec=2[0, 4, 2, 2] sub=4[4] fig=7[0, 0, 7]
  h1 list-item=0[0] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7]
  p list-item=5[0, 5] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7] undeclared=2[2]
  p list-item=10[0, 10] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7] undeclared=2[2]
  ul list-item=10[0, 10] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7]
   li list-item=1[0, 10, 1] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7]
   span list-item=2[0, 10, 2] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7] undeclared=2[2]
   li list-item=3[0, 10, 3] sec=3[0, 4, 3] sub=4[4] fig=7[0, 0, 7]
`

func TestLayoutContext_counterContextMatchesJava(t *testing.T) {
	sc, doc := layoutContextTestSharedContext(t, layoutContextTestCounterDocument)
	c := sc.NewLayoutContextInstance(nil)

	var out strings.Builder
	var walk func(e *dom.Element, indent string)
	walk = func(e *dom.Element, indent string) {
		style := sc.GetStyle(e)
		var start *int
		if e.GetNodeName() == "ol" && e.HasAttribute("start") {
			v, _ := strconv.Atoi(e.GetAttribute("start"))
			v--
			start = &v
		}
		if e.GetNodeName() == "li" && e.HasAttribute("value") {
			v, _ := strconv.Atoi(e.GetAttribute("value"))
			v--
			start = &v
		}
		if start != nil {
			c.ResolveCountersWithStartIndex(style, start)
		} else {
			c.ResolveCounters(style)
		}
		cc := c.GetCounterContext(style)
		out.WriteString(indent + e.GetNodeName())
		for _, name := range []string{"list-item", "sec", "sub", "fig", "undeclared"} {
			if (len(e.GetNodeName())+len(name))%3 == 0 {
				continue
			}
			value := cc.GetCurrentCounterValue(name)
			values := cc.GetCurrentCounterValues(name)
			parts := make([]string, len(values))
			for i, v := range values {
				parts[i] = strconv.Itoa(v)
			}
			fmt.Fprintf(&out, " %s=%d[%s]", name, value, strings.Join(parts, ", "))
		}
		out.WriteString("\n")
		for _, child := range e.GetChildNodes() {
			if ch, ok := child.(*dom.Element); ok {
				walk(ch, indent+" ")
			}
		}
	}
	walk(doc.GetDocumentElement(), "")

	expected := strings.TrimPrefix(layoutContextTestCounterExpected, "\n")
	if out.String() != expected {
		t.Errorf("counter values differ from Java\ngot:\n%s\nwant:\n%s", out.String(), expected)
	}
}

func TestLayoutContext_getCounterContextOfUnresolvedStyle(t *testing.T) {
	sc, doc := layoutContextTestSharedContext(t, layoutContextTestCounterDocument)
	c := sc.NewLayoutContextInstance(nil)
	if cc := c.GetCounterContext(sc.GetStyle(doc.GetDocumentElement())); cc != nil {
		t.Errorf("expected no counter context before ResolveCounters, got %v", cc)
	}
}

func TestLayoutContext_pageState(t *testing.T) {
	sc, _ := layoutContextTestSharedContext(t, layoutContextTestCounterDocument)
	sc.SetPrint(true)
	c := sc.NewLayoutContextInstance(nil)
	if !c.IsPrint() || !c.IsPageBreaksAllowed() || !c.IsMayCheckKeepTogether() {
		t.Fatalf("unexpected defaults: print=%v pageBreaksAllowed=%v mayCheckKeepTogether=%v",
			c.IsPrint(), c.IsPageBreaksAllowed(), c.IsMayCheckKeepTogether())
	}
	c.SetNoPageBreak(2)
	if c.IsPageBreaksAllowed() {
		t.Error("page breaks must not be allowed while noPageBreak is 2")
	}
	c.SetPageName("chapter")
	c.SetExtraSpaceTop(11)
	c.SetExtraSpaceBottom(22)

	state := c.CaptureLayoutState()
	c.SetNoPageBreak(0)
	c.SetPageName("other")
	c.SetExtraSpaceTop(1)
	c.SetExtraSpaceBottom(2)
	c.RestoreLayoutState(state)
	if c.GetNoPageBreak() != 2 || c.GetPageName() != "chapter" {
		t.Errorf("noPageBreak=%d pageName=%q", c.GetNoPageBreak(), c.GetPageName())
	}
	// Java's captureLayoutState passes the bottom space as the top space and
	// the reverse, so a capture followed by a restore exchanges the two.
	if c.GetExtraSpaceTop() != 22 || c.GetExtraSpaceBottom() != 11 {
		t.Errorf("extraSpaceTop=%d extraSpaceBottom=%d, want 22 and 11", c.GetExtraSpaceTop(), c.GetExtraSpaceBottom())
	}

	relayout := c.CopyStateForRelayout()
	if relayout.GetPageName() != "chapter" || relayout.GetNoPageBreak() != 0 ||
		relayout.GetExtraSpaceTop() != 0 || relayout.GetExtraSpaceBottom() != 0 || len(relayout.GetBFCs()) != 0 {
		t.Errorf("unexpected relayout state: %+v", relayout)
	}
	if relayout.GetFirstLines() == c.GetFirstLinesTracker() || relayout.GetFirstLetters() == c.GetFirstLettersTracker() {
		t.Error("CopyStateForRelayout must copy the style trackers")
	}
}
