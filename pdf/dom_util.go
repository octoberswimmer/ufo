// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/DOMUtil.java

package pdf

import (
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// DOMUtilGetChild returns nil when parent has no child element with the given
// tag name.
func DOMUtilGetChild(parent *dom.Element, name string) *dom.Element {
	children := parent.GetChildNodes()
	for i := 0; i < len(children); i++ {
		n := children[i]
		if n.GetNodeType() == dom.ElementNode {
			elem := n.(*dom.Element)
			if elem.GetTagName() == name {
				return elem
			}
		}
	}
	return nil
}

func DOMUtilGetChildren(parent *dom.Element, name string) []*dom.Element {
	children := parent.GetChildNodes()
	result := make([]*dom.Element, 0, len(children))
	for i := 0; i < len(children); i++ {
		n := children[i]
		if n.GetNodeType() == dom.ElementNode {
			elem := n.(*dom.Element)
			if elem.GetTagName() == name {
				result = append(result, elem)
			}
		}
	}
	return result
}

// DOMUtilGetText loads all the text content in all offspring of an element.
// Ignores all attributes, comments and processing instructions.
//
// It returns a string with the text content of an element (maybe an empty
// string).
func DOMUtilGetText(parent *dom.Element) string {
	var sb strings.Builder
	DOMUtilGetTextWithSb(parent, &sb)
	return sb.String()
}

// DOMUtilGetTextWithSb appends all text content in all offspring of an
// element to a strings.Builder. Ignores all attributes, comments and
// processing instructions.
func DOMUtilGetTextWithSb(parent *dom.Element, sb *strings.Builder) {
	children := parent.GetChildNodes()
	for i := 0; i < len(children); i++ {
		n := children[i]
		if n.GetNodeType() == dom.ElementNode {
			DOMUtilGetTextWithSb(n.(*dom.Element), sb)
		} else if n.GetNodeType() == dom.TextNode {
			sb.WriteString(n.GetNodeValue())
		}
	}
}
