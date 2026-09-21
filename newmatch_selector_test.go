// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/Selector.java
// and the conditions of Condition.java. Flying Saucer has no JUnit test for
// them.

package ufo

import (
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

// newmatchTestAttributeResolver reads class, id, lang and the other
// attributes from the dom elements. An element is a link when it has an href
// attribute; the dynamic states are the sets of element ids.
type newmatchTestAttributeResolver struct {
	visited map[string]bool
	hover   map[string]bool
	active  map[string]bool
	focus   map[string]bool
}

func (r *newmatchTestAttributeResolver) GetAttributeValue(e dom.Node, attrName string) *string {
	return r.GetAttributeValueWithNamespaceURI(e, nil, attrName)
}

func (r *newmatchTestAttributeResolver) GetAttributeValueWithNamespaceURI(e dom.Node, namespaceURI *string, attrName string) *string {
	element := e.(*dom.Element)
	for _, attr := range element.GetAttributes() {
		if attr.LocalName == attrName && (namespaceURI == nil || *namespaceURI == attr.NamespaceURI) {
			value := attr.Value
			return &value
		}
	}
	return nil
}

func (r *newmatchTestAttributeResolver) GetClass(e dom.Node) string {
	return e.(*dom.Element).GetAttribute("class")
}

func (r *newmatchTestAttributeResolver) GetID(e dom.Node) string {
	return e.(*dom.Element).GetAttribute("id")
}

func (r *newmatchTestAttributeResolver) GetNonCssStyling(e dom.Node) string {
	return e.(*dom.Element).GetAttribute("noncss")
}

func (r *newmatchTestAttributeResolver) GetElementStyling(e dom.Node) string {
	return e.(*dom.Element).GetAttribute("style")
}

func (r *newmatchTestAttributeResolver) GetLang(e dom.Node) string {
	return e.(*dom.Element).GetAttribute("lang")
}

func (r *newmatchTestAttributeResolver) IsLink(e dom.Node) bool {
	return e.(*dom.Element).HasAttribute("href")
}

func (r *newmatchTestAttributeResolver) IsVisited(e dom.Node) bool {
	return r.visited[r.GetID(e)]
}

func (r *newmatchTestAttributeResolver) IsHover(e dom.Node) bool {
	return r.hover[r.GetID(e)]
}

func (r *newmatchTestAttributeResolver) IsActive(e dom.Node) bool {
	return r.active[r.GetID(e)]
}

func (r *newmatchTestAttributeResolver) IsFocus(e dom.Node) bool {
	return r.focus[r.GetID(e)]
}

const newmatchTestDocument = `<html>
<body id="body" lang="en">
	<div id="first" class="box wide">
		<p id="p1" class="intro">one</p>
		text
		<p id="p2" title="hello world" lang="en-GB">two</p>
		<p id="p3" data-kind="alpha-beta">three</p>
		<a id="link" href="http://example.com/">four</a>
	</div>
	<div id="second">
		<span id="s1"><em id="em1">five</em></span>
	</div>
	<ul id="list"><li id="li1"/><li id="li2"/><li id="li3"/><li id="li4"/><li id="li5"/></ul>
</body>
</html>`

func newmatchTestParse(t *testing.T) *dom.Document {
	t.Helper()
	doc, err := dom.ParseXMLString(newmatchTestDocument)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func newmatchTestElement(t *testing.T, doc *dom.Document, id string) *dom.Element {
	t.Helper()
	element := doc.GetElementByID(id)
	if element == nil {
		t.Fatalf("no element with id %q", id)
	}
	return element
}

// newmatchTestMatchingIDs returns the ids of the elements, in document order,
// that the selector matches on its own (not along a chain).
func newmatchTestMatchingIDs(doc *dom.Document, selector *Selector, attRes AttributeResolver) []string {
	treeRes := NewDOMTreeResolver()
	var ids []string
	for _, element := range doc.GetElementsByTagName("*") {
		if selector.Matches(element, attRes, treeRes) && selector.MatchesDynamic(element, attRes, treeRes) {
			ids = append(ids, element.GetAttribute("id"))
		}
	}
	return ids
}

func newmatchTestAssertIDs(t *testing.T, actual []string, expected ...string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("matched %v, expected %v", actual, expected)
	}
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("matched %v, expected %v", actual, expected)
		}
	}
}

func TestSelector_matchesElementName(t *testing.T) {
	doc := newmatchTestParse(t)
	selector := NewSelector(nil)
	selector.SetName("p")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, &newmatchTestAttributeResolver{}), "p1", "p2", "p3")
}

func TestSelector_universalSelectorMatchesEveryElement(t *testing.T) {
	doc := newmatchTestParse(t)
	selector := NewSelector(nil)
	selector.SetName("")
	if actual := len(newmatchTestMatchingIDs(doc, selector, nil)); actual != len(doc.GetElementsByTagName("*")) {
		t.Errorf("matched %d elements", actual)
	}
	// The parser calls setName(null) for "*", and setName always counts.
	if selector.GetSpecificityD() != 1 {
		t.Errorf("specificity d = %d, expected 1", selector.GetSpecificityD())
	}
}

func TestSelector_namespaceMustEqualTheElementNamespace(t *testing.T) {
	doc, err := dom.ParseXMLString(`<r xmlns="urn:a" xmlns:b="urn:b"><x id="a"/><b:x id="b"/></r>`)
	if err != nil {
		t.Fatal(err)
	}
	namespaceB := "urn:b"
	selector := NewSelector(nil)
	selector.SetNamespaceURI(&namespaceB)
	selector.SetName("x")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil), "b")

	any := NewSelector(nil)
	any.SetName("x")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, any, nil), "a", "b")
}

func TestSelector_noNamespaceSelectorMatchesNothing(t *testing.T) {
	// DOMTreeResolver.matchesElement compares the "" of NO_NAMESPACE with the
	// null namespace of an element in no namespace, and they are not equal.
	doc := newmatchTestParse(t)
	noNamespace := TreeResolverNoNamespace
	selector := NewSelector(nil)
	selector.SetNamespaceURI(&noNamespace)
	selector.SetName("p")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil))
}

func TestSelector_conditions(t *testing.T) {
	cases := []struct {
		name     string
		build    func(s *Selector)
		expected []string
	}{
		{"class", func(s *Selector) { s.AddClassCondition("wide") }, []string{"first"}},
		{"class is a whole word", func(s *Selector) { s.AddClassCondition("wid") }, nil},
		{"id", func(s *Selector) { s.AddIDCondition("p2") }, []string{"p2"}},
		{"attribute exists", func(s *Selector) { s.AddAttributeExistsCondition(nil, "title") }, []string{"p2"}},
		{"attribute equals", func(s *Selector) { s.AddAttributeEqualsCondition(nil, "title", "hello world") }, []string{"p2"}},
		{"attribute equals needs the whole value", func(s *Selector) { s.AddAttributeEqualsCondition(nil, "title", "hello") }, nil},
		{"attribute prefix", func(s *Selector) { s.AddAttributePrefixCondition(nil, "title", "hell") }, []string{"p2"}},
		{"attribute suffix", func(s *Selector) { s.AddAttributeSuffixCondition(nil, "title", "orld") }, []string{"p2"}},
		{"attribute suffix mismatch", func(s *Selector) { s.AddAttributeSuffixCondition(nil, "title", "hello") }, nil},
		{"attribute substring", func(s *Selector) { s.AddAttributeSubstringCondition(nil, "title", "lo wo") }, []string{"p2"}},
		{"attribute list", func(s *Selector) { s.AddAttributeMatchesListCondition(nil, "title", "world") }, []string{"p2"}},
		{"attribute list needs a whole word", func(s *Selector) { s.AddAttributeMatchesListCondition(nil, "title", "wor") }, nil},
		{"attribute first part", func(s *Selector) { s.AddAttributeMatchesFirstPartCondition(nil, "data-kind", "alpha") }, []string{"p3"}},
		{"attribute first part needs the whole part", func(s *Selector) { s.AddAttributeMatchesFirstPartCondition(nil, "data-kind", "alp") }, nil},
		{"lang", func(s *Selector) { s.AddLangCondition("en") }, []string{"body", "p2"}},
		{"link", func(s *Selector) { s.AddLinkCondition() }, []string{"link"}},
		{"first-child", func(s *Selector) { s.SetName("p"); s.AddFirstChildCondition() }, []string{"p1"}},
		{"last-child", func(s *Selector) { s.SetName("li"); s.AddLastChildCondition() }, []string{"li5"}},
		{"nth-child", func(s *Selector) { s.SetName("li"); s.AddNthChildCondition("2n+1") }, []string{"li1", "li3", "li5"}},
		{"nth-child counts elements only", func(s *Selector) { s.SetName("p"); s.AddNthChildCondition("2") }, []string{"p2"}},
		// :even and :odd use the 0-based position.
		{"even", func(s *Selector) { s.SetName("li"); s.AddEvenChildCondition() }, []string{"li1", "li3", "li5"}},
		{"odd", func(s *Selector) { s.SetName("li"); s.AddOddChildCondition() }, []string{"li2", "li4"}},
		{"unsupported", func(s *Selector) { s.SetName("li"); s.AddUnsupportedCondition() }, nil},
		{"all conditions must hold", func(s *Selector) { s.SetName("p"); s.AddClassCondition("intro"); s.AddIDCondition("p2") }, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := newmatchTestParse(t)
			selector := NewSelector(nil)
			c.build(selector)
			newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, &newmatchTestAttributeResolver{}), c.expected...)
		})
	}
}

func TestSelector_attributeConditionsNeedAnAttributeResolver(t *testing.T) {
	doc := newmatchTestParse(t)
	for _, build := range []func(s *Selector){
		func(s *Selector) { s.AddClassCondition("wide") },
		func(s *Selector) { s.AddIDCondition("p2") },
		func(s *Selector) { s.AddAttributeExistsCondition(nil, "title") },
		func(s *Selector) { s.AddLangCondition("en") },
	} {
		selector := NewSelector(nil)
		build(selector)
		newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil))
	}
}

func TestSelector_dynamicPseudoClasses(t *testing.T) {
	doc := newmatchTestParse(t)
	attRes := &newmatchTestAttributeResolver{
		visited: map[string]bool{"link": true},
		hover:   map[string]bool{"p1": true, "link": true},
		active:  map[string]bool{"p2": true},
		focus:   map[string]bool{"p3": true},
	}
	cases := []struct {
		pc       []int
		expected []string
	}{
		{[]int{SelectorVisitedPseudoclass}, []string{"link"}},
		{[]int{SelectorHoverPseudoclass}, []string{"p1", "link"}},
		{[]int{SelectorActivePseudoclass}, []string{"p2"}},
		{[]int{SelectorFocusPseudoclass}, []string{"p3"}},
		{[]int{SelectorHoverPseudoclass, SelectorVisitedPseudoclass}, []string{"link"}},
		{[]int{SelectorHoverPseudoclass, SelectorFocusPseudoclass}, nil},
	}
	for _, c := range cases {
		selector := NewSelector(nil)
		for _, pc := range c.pc {
			selector.SetPseudoClass(pc)
		}
		newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, attRes), c.expected...)
		// Without an AttributeResolver no dynamic pseudo-class holds.
		newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil))
		if selector.GetSpecificityC() != len(c.pc) {
			t.Errorf("specificity c = %d, expected %d", selector.GetSpecificityC(), len(c.pc))
		}
	}
}

func TestSelector_setPseudoClassCountsEachClassOnce(t *testing.T) {
	selector := NewSelector(nil)
	selector.SetPseudoClass(SelectorHoverPseudoclass)
	selector.SetPseudoClass(SelectorHoverPseudoclass)
	if selector.GetSpecificityC() != 1 {
		t.Errorf("specificity c = %d, expected 1", selector.GetSpecificityC())
	}
	if !selector.IsPseudoClass(SelectorHoverPseudoclass) || selector.IsPseudoClass(SelectorFocusPseudoclass) {
		t.Errorf("IsPseudoClass gives the wrong classes")
	}
}

func TestSelector_immediateSibling(t *testing.T) {
	doc := newmatchTestParse(t)
	// p.intro + p
	sibling := NewSelector(nil)
	sibling.SetName("p")
	sibling.AddClassCondition("intro")
	sibling.SetAxis(SelectorAxisImmediateSiblingAxis)
	selector := NewSelector(nil)
	selector.SetName("p")
	selector.SetSiblingSelector(sibling)
	// The text between p1 and p2 is skipped.
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, &newmatchTestAttributeResolver{}), "p2")
}

func TestSelector_immediateSiblingWithDynamicPseudoClass(t *testing.T) {
	doc := newmatchTestParse(t)
	// p:hover + p
	sibling := NewSelector(nil)
	sibling.SetName("p")
	sibling.SetPseudoClass(SelectorHoverPseudoclass)
	sibling.SetAxis(SelectorAxisImmediateSiblingAxis)
	selector := NewSelector(nil)
	selector.SetName("p")
	selector.SetSiblingSelector(sibling)
	attRes := &newmatchTestAttributeResolver{hover: map[string]bool{"p2": true}}
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, attRes), "p3")
}

func TestSelector_siblingSelectorOnAnotherAxisMatchesNothing(t *testing.T) {
	doc := newmatchTestParse(t)
	sibling := NewSelector(nil)
	sibling.SetName("p")
	selector := NewSelector(nil)
	selector.SetName("p")
	selector.SetSiblingSelector(sibling)
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil))
}

func TestSelector_hasCondition(t *testing.T) {
	named := func(name string) *Selector {
		s := NewSelector(nil)
		s.SetName(name)
		return s
	}
	cases := []struct {
		name      string
		axes      []SelectorAxis
		selectors []*Selector
		expected  []string
	}{
		{":has(em)", []SelectorAxis{SelectorAxisDescendantAxis}, []*Selector{named("em")}, []string{"", "body", "second", "s1"}},
		{":has(> em)", []SelectorAxis{SelectorAxisChildAxis}, []*Selector{named("em")}, []string{"s1"}},
		{":has(+ div)", []SelectorAxis{SelectorAxisImmediateSiblingAxis}, []*Selector{named("div")}, []string{"first"}},
		// The candidates come from the first axis (the children of the scope)
		// and are matched against the last compound selector, so a relative
		// selector that starts with ">" and has a second ">" step matches
		// nothing, as in Java.
		{":has(> span > em)", []SelectorAxis{SelectorAxisChildAxis, SelectorAxisChildAxis}, []*Selector{named("span"), named("em")}, nil},
		{":has(span em)", []SelectorAxis{SelectorAxisDescendantAxis, SelectorAxisDescendantAxis}, []*Selector{named("span"), named("em")}, []string{"", "body", "second"}},
		{":has(> p + a)", []SelectorAxis{SelectorAxisChildAxis, SelectorAxisImmediateSiblingAxis}, []*Selector{named("p"), named("a")}, []string{"first"}},
		{":has(> em + a)", []SelectorAxis{SelectorAxisChildAxis, SelectorAxisImmediateSiblingAxis}, []*Selector{named("em"), named("a")}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := newmatchTestParse(t)
			selector := NewSelector(nil)
			selector.AddHasCondition([]*SelectorHasRelativeSelector{NewSelectorHasRelativeSelector(c.axes, c.selectors)}, 1, 2, 3)
			newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, &newmatchTestAttributeResolver{}), c.expected...)
			if selector.GetSpecificityB() != 1 || selector.GetSpecificityC() != 2 || selector.GetSpecificityD() != 3 {
				t.Errorf("specificity = %d,%d,%d", selector.GetSpecificityB(), selector.GetSpecificityC(), selector.GetSpecificityD())
			}
		})
	}
}

func TestSelector_hasRelativeSelectorRejectsUnequalLists(t *testing.T) {
	defer func() {
		if _, ok := recover().(*XRRuntimeException); !ok {
			t.Errorf("expected an *XRRuntimeException panic")
		}
	}()
	NewSelectorHasRelativeSelector([]SelectorAxis{SelectorAxisChildAxis}, nil)
}

func TestSelector_specificityAndOrder(t *testing.T) {
	selector := NewSelector(nil)
	selector.SetName("p")                              // d
	selector.AddIDCondition("x")                       // b
	selector.AddClassCondition("y")                    // c
	selector.AddAttributeExistsCondition(nil, "title") // c
	selector.AddFirstChildCondition()                  // c
	selector.SetPseudoClass(SelectorHoverPseudoclass)  // c
	selector.SetPseudoElement("before")                // d
	selector.SetPos(42)
	if actual := selector.GetOrder(); actual != "000100400200042" {
		t.Errorf("GetOrder() = %q", actual)
	}
}

func TestSelector_orderComesFromTheLastSelectorOfTheChain(t *testing.T) {
	first := NewSelector(nil)
	first.SetName("div")
	first.AddIDCondition("x")
	last := NewSelector(nil)
	last.SetName("p")
	last.SetSpecificityB(12)
	last.SetSpecificityC(1234)
	last.SetSpecificityD(5)
	first.SetChainedSelector(last)
	first.SetPos(123456)
	// Each specificity keeps its last three digits and the position its last
	// five.
	if actual := first.GetOrder(); actual != "001223400523456" {
		t.Errorf("GetOrder() = %q", actual)
	}
	if first.GetChainedSelector() != last {
		t.Errorf("GetChainedSelector() gives another selector")
	}
}

func TestSelector_selectorText(t *testing.T) {
	first := NewSelector(nil)
	first.SetName("div")
	first.AddClassCondition("box")
	last := NewSelector(nil)
	last.SetName("p")
	first.SetChainedSelector(last)
	if actual := first.GetSelectorText(); actual != "div.box p" {
		t.Errorf("GetSelectorText() = %q", actual)
	}

	// Java concatenates the null name of a selector without an element name.
	classOnly := NewSelector(nil)
	classOnly.AddClassCondition("box")
	if actual := classOnly.GetSelectorText(); actual != "null.box" {
		t.Errorf("GetSelectorText() = %q", actual)
	}
}

func TestSelector_secondPseudoElementMakesTheSelectorUnsupported(t *testing.T) {
	doc := newmatchTestParse(t)
	selector := NewSelector(nil)
	selector.SetName("p")
	selector.SetPseudoElement("before")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil), "p1", "p2", "p3")
	selector.SetPseudoElement("after")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, nil))
	if selector.GetPseudoElement() != "before" {
		t.Errorf("GetPseudoElement() = %q", selector.GetPseudoElement())
	}
	if selector.GetSpecificityD() != 2 {
		t.Errorf("specificity d = %d, expected 2", selector.GetSpecificityD())
	}
}

func TestSelector_conditionAfterPseudoElementMakesTheSelectorUnsupported(t *testing.T) {
	doc := newmatchTestParse(t)
	selector := NewSelector(nil)
	selector.SetName("p")
	selector.SetPseudoElement("before")
	selector.AddClassCondition("intro")
	newmatchTestAssertIDs(t, newmatchTestMatchingIDs(doc, selector, &newmatchTestAttributeResolver{}))
}

func TestSelector_selectorIDsIncrease(t *testing.T) {
	first := NewSelector(nil)
	second := NewSelector(nil)
	if second.GetSelectorID() != first.GetSelectorID()+1 {
		t.Errorf("ids %d and %d", first.GetSelectorID(), second.GetSelectorID())
	}
}
