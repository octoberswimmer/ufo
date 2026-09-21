// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/TextUtil.java

package ufo

import (
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// TextUtilReadTextContentOrNull returns the text content of the element, and
// "" where Java returns null, which is when the element has no text.
func TextUtilReadTextContentOrNull(element *dom.Element) string {
	return TextUtilReadTextContent(element)
}

func TextUtilReadTextContent(element *dom.Element) string {
	var result strings.Builder
	current := element.GetFirstChild()
	for current != nil {
		nodeType := current.GetNodeType()
		if nodeType == dom.TextNode || nodeType == dom.CDATASectionNode {
			t := current.(*dom.Text)
			result.WriteString(t.GetData())
		}
		current = current.GetNextSibling()
	}
	return result.String()
}
