// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/HTMLOutline.java

package pdf

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
)

var htmlOutlineHeading = regexp.MustCompile(`(?i)^h(\d+)$`)

// sectioning roots: https://www.w3.org/TR/html51/sections.html#sectioning-roots
var htmlOutlineRoot = regexp.MustCompile(`(?i)^(?:blockquote|details|fieldset|figure|td)$`)

// htmlOutlineWs is Java's \s+, which includes the vertical tab that Go's \s
// leaves out.
var htmlOutlineWs = regexp.MustCompile("[ \\t\\n\\x0B\\f\\r]+")

const htmlOutlineMaxNameLength = 200

type HTMLOutline struct {
	parent   *HTMLOutline
	level    int
	bookmark *ITextOutputDeviceBookmark
}

func newHTMLOutline() *HTMLOutline {
	return newHTMLOutlineWithLevelNameParent(0, "root", nil)
}

func newHTMLOutlineWithLevelNameParent(level int, name string, parent *HTMLOutline) *HTMLOutline {
	o := &HTMLOutline{
		level:    level,
		bookmark: NewITextOutputDeviceBookmark(name, ""),
		parent:   parent,
	}
	if parent != nil {
		parent.bookmark.AddChild(o.bookmark)
	}
	return o
}

// HTMLOutlineGenerate creates a bookmark list of the document outline
// generated for the given element context (usually the root document
// element).
//
// The current algorithm is more simple than the one suggested in the HTML5
// specification such as it is not affected by sectioning content
// (https://www.w3.org/TR/html51/dom.html#sectioning-content) but just the
// heading level. For example
// (https://www.w3.org/TR/html51/sections.html#example-d42b7aaf):
//
//	<body>
//	  <h1>Foo</h1>
//	  <h3>Bar</h3>
//	  <blockquote>
//	    <h5>Bla</h5>
//	  </blockquote>
//	  <p>Baz</p>
//	  <h2>Quux</h2>
//	  <section>
//	    <h3>Thud</h3>
//	  </section>
//	  <h4>Grunt</h4>
//	</body>
//
// Should generate outline as:
//
//	Foo
//	    Bar
//	    Quux
//	    Thud
//	    Grunt
//
// But it generates outline as:
//
//	Foo
//	    Bar
//	    Quux
//	        Thud
//	            Grunt
//
// # Example document customizations
//
// Include non-heading element as bookmark (level 4):
//
//	<strong data-pdf-bookmark="4">Foo bar</strong>
//
// Specify bookmark name:
//
//	<tr data-pdf-bookmark="5" data-pdf-bookmark-name="Bar baz">...</tr>
//
// Exclude individual heading from bookmarks:
//
//	<h3 data-pdf-bookmark="none">Baz qux</h3>
//
// Prevent automatic bookmarks for the whole of the document:
//
//	<html data-pdf-bookmark="exclude">...</html>
//
// context is the top element a sectioning outline would be generated for;
// box is the box hierarchy the outline bookmarks would get mapped into. The
// result is the bookmarks of the outline generated for the given element
// context. See https://www.w3.org/TR/html51/sections.html#creating-an-outline
func HTMLOutlineGenerate(context *dom.Element, box ufo.BoxI) []*ITextOutputDeviceBookmark {
	iterator := htmlOutlineNestedSectioningFilterIterator(context)

	if iterator == nil {
		return nil
	}

	root := newHTMLOutline()
	current := root
	bookmarks := map[*dom.Element]*ITextOutputDeviceBookmark{}

	for element := iterator.nextNode(); element != nil; element = iterator.nextNode() {
		// Integer.parseInt: a value outside the int range is invalid too.
		parsed, err := strconv.ParseInt(htmlOutlineGetOutlineLevel(element), 10, 32)
		if err != nil {
			continue // Invalid value
		}
		level := int(parsed)
		if level < 1 {
			continue // Illegal value
		}

		name := htmlOutlineGetBookmarkName(element)

		for current.level >= level {
			current = current.parent
		}
		current = newHTMLOutlineWithLevelNameParent(level, name, current)
		bookmarks[element] = current.bookmark
	}
	htmlOutlineInitBoxRefs(bookmarks, box)
	return root.bookmark.GetChildren()
}

func htmlOutlineInitBoxRefs(bookmarks map[*dom.Element]*ITextOutputDeviceBookmark, box ufo.BoxI) {
	if element := box.GetElement(); element != nil {
		if bookmark := bookmarks[element]; bookmark != nil {
			bookmark.SetBox(box)
		}
	}
	for i, length := 0, box.GetChildCount(); i < length; i++ {
		htmlOutlineInitBoxRefs(bookmarks, box.GetChild(i))
	}
}

func htmlOutlineGetBookmarkName(element *dom.Element) string {
	name := htmlOutlineJavaTrim(element.GetAttribute("data-pdf-bookmark-name"))
	if name == "" {
		name = element.GetTextContent()
	}
	name = htmlOutlineWs.ReplaceAllString(htmlOutlineJavaTrim(name), " ")
	if utf8.RuneCountInString(name) > htmlOutlineMaxNameLength {
		// The limit counts characters, so that the cut never lands inside
		// one.
		name = string([]rune(name)[:htmlOutlineMaxNameLength])
	}
	return name
}

func htmlOutlineGetOutlineLevel(element *dom.Element) string {
	bookmark := htmlOutlineJavaTrim(element.GetAttribute("data-pdf-bookmark"))
	if bookmark == "" {
		return HTMLOutlineGetOutlineLevelFromTagName(element.GetTagName())
	}
	return bookmark
}

func HTMLOutlineGetOutlineLevelFromTagName(tagName string) string {
	heading := htmlOutlineHeading.FindStringSubmatch(tagName)
	if heading != nil {
		return heading[1]
	} else if htmlOutlineRoot.MatchString(tagName) {
		return "exclude"
	} else {
		return "none"
	}
}

// htmlOutlineJavaTrim is String.trim(): it removes leading and trailing
// characters up to U+0020.
func htmlOutlineJavaTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}

// NodeFilter results, with the values of org.w3c.dom.traversal.NodeFilter.
const (
	htmlOutlineFilterAccept int16 = 1
	htmlOutlineFilterReject int16 = 2
	htmlOutlineFilterSkip   int16 = 3
)

type htmlOutlineNestedSectioningFilter struct{}

var htmlOutlineNestedSectioningFilterInstance = &htmlOutlineNestedSectioningFilter{}

// htmlOutlineNestedSectioningFilterIterator returns an iterator over the
// elements of the subtree of root, root included, in document order. Java
// returns null when the owner document does not implement DocumentTraversal;
// a dom.Document can always be traversed, so the result is nil only for an
// element without an owner document.
func htmlOutlineNestedSectioningFilterIterator(root *dom.Element) *htmlOutlineNodeIterator {
	if root == nil || root.GetOwnerDocument() == nil {
		return nil
	}
	return newHTMLOutlineNodeIterator(root, htmlOutlineNestedSectioningFilterInstance)
}

func (f *htmlOutlineNestedSectioningFilter) acceptNode(n *dom.Element) int16 {
	outlineLevel := htmlOutlineGetOutlineLevel(n)
	if strings.EqualFold(outlineLevel, "none") {
		return htmlOutlineFilterSkip
	}
	if strings.EqualFold(outlineLevel, "exclude") {
		return htmlOutlineFilterReject
	}
	return htmlOutlineFilterAccept
}

// htmlOutlineNodeIterator stands in for the org.w3c.dom.traversal.NodeIterator
// that Document.createNodeIterator(root, SHOW_ELEMENT, filter, true) returns.
// A NodeIterator walks a flat list of the nodes in document order, so it
// treats FILTER_REJECT like FILTER_SKIP: the descendants of a rejected element
// are still visited. (Only a TreeWalker leaves out the subtree of a rejected
// node.) The JDK's iterator was run over a document with a heading inside a
// rejected blockquote to confirm that the heading is returned.
type htmlOutlineNodeIterator struct {
	elements []*dom.Element
	next     int
	filter   *htmlOutlineNestedSectioningFilter
}

func newHTMLOutlineNodeIterator(root *dom.Element, filter *htmlOutlineNestedSectioningFilter) *htmlOutlineNodeIterator {
	it := &htmlOutlineNodeIterator{filter: filter}
	it.collect(root)
	return it
}

func (it *htmlOutlineNodeIterator) collect(e *dom.Element) {
	it.elements = append(it.elements, e)
	for _, child := range e.GetChildNodes() {
		if childElement, ok := child.(*dom.Element); ok {
			it.collect(childElement)
		}
	}
}

// nextNode returns the next element the filter accepts, or nil at the end.
func (it *htmlOutlineNodeIterator) nextNode() *dom.Element {
	for it.next < len(it.elements) {
		e := it.elements[it.next]
		it.next++
		if it.filter.acceptNode(e) == htmlOutlineFilterAccept {
			return e
		}
	}
	return nil
}
