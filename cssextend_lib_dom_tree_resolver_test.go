// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/extend/lib/DOMTreeResolver.java.
// Flying Saucer has no JUnit test for it.

package ufo

import (
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

func TestDOMTreeResolver_parentElement(t *testing.T) {
	doc := newmatchTestParse(t)
	resolver := NewDOMTreeResolver()
	if parent := resolver.GetParentElement(newmatchTestElement(t, doc, "p1")); parent != dom.Node(newmatchTestElement(t, doc, "first")) {
		t.Errorf("the parent of p1 is %v", parent)
	}
	// The parent of the root element is the document, which is not an element.
	if parent := resolver.GetParentElement(doc.GetDocumentElement()); parent != nil {
		t.Errorf("the parent of the root element is %v", parent)
	}
}

func TestDOMTreeResolver_previousSiblingElementSkipsText(t *testing.T) {
	doc := newmatchTestParse(t)
	resolver := NewDOMTreeResolver()
	if sibling := resolver.GetPreviousSiblingElement(newmatchTestElement(t, doc, "p2")); sibling != dom.Node(newmatchTestElement(t, doc, "p1")) {
		t.Errorf("the previous sibling of p2 is %v", sibling)
	}
	if sibling := resolver.GetPreviousSiblingElement(newmatchTestElement(t, doc, "p1")); sibling != nil {
		t.Errorf("the previous sibling of p1 is %v", sibling)
	}
}

func TestDOMTreeResolver_firstAndLastChildElement(t *testing.T) {
	doc := newmatchTestParse(t)
	resolver := NewDOMTreeResolver()
	cases := []struct {
		id    string
		first bool
		last  bool
	}{
		{"p1", true, false},
		{"p2", false, false},
		{"link", false, true},
		{"em1", true, true},
	}
	for _, c := range cases {
		element := newmatchTestElement(t, doc, c.id)
		if actual := resolver.IsFirstChildElement(element); actual != c.first {
			t.Errorf("IsFirstChildElement(#%s) = %v", c.id, actual)
		}
		if actual := resolver.IsLastChildElement(element); actual != c.last {
			t.Errorf("IsLastChildElement(#%s) = %v", c.id, actual)
		}
	}
}

func TestDOMTreeResolver_positionOfElementCountsElementsOnly(t *testing.T) {
	doc := newmatchTestParse(t)
	resolver := NewDOMTreeResolver()
	for id, expected := range map[string]int{"p1": 0, "p2": 1, "p3": 2, "link": 3, "li5": 4} {
		if actual := resolver.GetPositionOfElement(newmatchTestElement(t, doc, id)); actual != expected {
			t.Errorf("GetPositionOfElement(#%s) = %d, expected %d", id, actual, expected)
		}
	}
}

func TestDOMTreeResolver_elementNameIsTheLocalName(t *testing.T) {
	doc, err := dom.ParseXMLString(`<r xmlns:b="urn:b"><b:x id="b">text</b:x></r>`)
	if err != nil {
		t.Fatal(err)
	}
	resolver := NewDOMTreeResolver()
	element := doc.GetElementByID("b")
	if name := resolver.GetElementName(element); name != "x" {
		t.Errorf("GetElementName() = %q", name)
	}
	if name := resolver.GetElementName(element.GetFirstChild()); name != "#text" {
		t.Errorf("GetElementName() of the text = %q", name)
	}
}

func TestDOMTreeResolver_matchesElement(t *testing.T) {
	doc, err := dom.ParseXMLString(`<r xmlns:b="urn:b"><x id="plain"/><b:x id="b"/></r>`)
	if err != nil {
		t.Fatal(err)
	}
	resolver := NewDOMTreeResolver()
	plain := doc.GetElementByID("plain")
	namespaced := doc.GetElementByID("b")
	namespaceB := "urn:b"
	namespaceC := "urn:c"
	noNamespace := TreeResolverNoNamespace
	cases := []struct {
		element      *dom.Element
		namespaceURI *string
		name         string
		expected     bool
	}{
		{plain, nil, "x", true},
		{plain, nil, "y", false},
		{namespaced, nil, "x", true},
		{namespaced, &namespaceB, "x", true},
		{namespaced, &namespaceB, "y", false},
		{namespaced, &namespaceC, "x", false},
		{plain, &namespaceB, "x", false},
		{namespaced, &noNamespace, "x", false},
		// As in Java, where "" is compared with a null namespace URI.
		{plain, &noNamespace, "x", false},
	}
	for i, c := range cases {
		if actual := resolver.MatchesElement(c.element, c.namespaceURI, c.name); actual != c.expected {
			t.Errorf("case %d: MatchesElement = %v, expected %v", i, actual, c.expected)
		}
	}
}
