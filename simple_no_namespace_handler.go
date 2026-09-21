// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/simple/NoNamespaceHandler.java

package ufo

import (
	"regexp"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// NoNamespaceHandler handles a general XML document.
type NoNamespaceHandler struct {
}

const noNamespaceHandlerNamespace = "http://www.w3.org/XML/1998/namespace"

func NewNoNamespaceHandler() *NoNamespaceHandler {
	return &NoNamespaceHandler{}
}

func (h *NoNamespaceHandler) GetNamespace() string {
	return noNamespaceHandlerNamespace
}

func (h *NoNamespaceHandler) GetAttributeValue(e *dom.Element, attrName string) *string {
	value := e.GetAttribute(attrName)
	return &value
}

func (h *NoNamespaceHandler) GetAttributeValueWithNamespaceURI(e *dom.Element, namespaceURI *string, attrName string) *string {
	var value string
	if namespaceURI != nil && *namespaceURI == TreeResolverNoNamespace {
		value = e.GetAttribute(attrName)
	} else if namespaceURI == nil {
		if e.GetLocalName() == "" { // No namespaces
			value = e.GetAttribute(attrName)
		} else {
			value = ""
			for _, attr := range e.GetAttributes() {
				if attrName == attr.LocalName {
					value = attr.Value
					break
				}
			}
		}
	} else {
		value = e.GetAttributeNS(*namespaceURI, attrName)
	}
	return &value
}

func (h *NoNamespaceHandler) GetClass(e *dom.Element) string {
	return ""
}

func (h *NoNamespaceHandler) GetID(e *dom.Element) string {
	return ""
}

func (h *NoNamespaceHandler) GetLang(e *dom.Element) string {
	return e.GetAttribute("lang")
}

func (h *NoNamespaceHandler) GetElementStyling(e *dom.Element) string {
	return ""
}

func (h *NoNamespaceHandler) GetNonCssStyling(e *dom.Element) string {
	return ""
}

func (h *NoNamespaceHandler) GetLinkUri(e *dom.Element) *string {
	return nil
}

func (h *NoNamespaceHandler) GetDocumentTitle(doc *dom.Document) string {
	return ""
}

func (h *NoNamespaceHandler) GetAnchorName(e *dom.Element) string {
	return ""
}

func (h *NoNamespaceHandler) IsImageElement(e *dom.Element) bool {
	return false
}

func (h *NoNamespaceHandler) GetImageSourceURI(e *dom.Element) string {
	return ""
}

func (h *NoNamespaceHandler) IsFormElement(e *dom.Element) bool {
	return false
}

// noNamespaceHandlerWhitespace is what \s matches in java.util.regex; \s in
// Go's regexp leaves out U+000B.
const noNamespaceHandlerWhitespace = `[ \t\n\x0B\f\r]`

var (
	noNamespaceHandlerTypePattern = regexp.MustCompile(`type` + noNamespaceHandlerWhitespace + `?=` + noNamespaceHandlerWhitespace + `?`)
	noNamespaceHandlerHrefPattern = regexp.MustCompile(`href` + noNamespaceHandlerWhitespace + `?=` + noNamespaceHandlerWhitespace + `?`)
	// Java applies this pattern with Matcher.matches, which succeeds only
	// when the whole processing instruction is the pattern.
	noNamespaceHandlerAlternatePattern = regexp.MustCompile(`^(?:alternate` + noNamespaceHandlerWhitespace + `?=` + noNamespaceHandlerWhitespace + `?)$`)
	noNamespaceHandlerMediaPattern     = regexp.MustCompile(`media` + noNamespaceHandlerWhitespace + `?=` + noNamespaceHandlerWhitespace + `?`)
)

func (h *NoNamespaceHandler) GetStylesheets(doc *dom.Document) []*StylesheetInfo {
	var list []*StylesheetInfo
	//get the processing-instructions (actually for XmlDocuments)
	//type and href are required to be set
	for _, node := range doc.GetChildNodes() {
		if node.GetNodeType() != dom.ProcessingInstructionNode {
			continue
		}
		piNode := node.(*dom.ProcessingInstruction)
		if piNode.GetTarget() != "xml-stylesheet" {
			continue
		}
		pi := piNode.GetData()
		if m := noNamespaceHandlerAlternatePattern.FindStringIndex(pi); m != nil {
			alternate := noNamespaceHandlerQuotedValue(pi, m[1])
			//TODO: handle alternate stylesheets
			if alternate == "yes" {
				continue //DON'T get alternate stylesheets for now
			}
		}

		typ := h.detectType(pi)
		//TODO: handle other stylesheet types
		if typ != "text/css" {
			continue // for now
		}

		info := NewStylesheetInfo(StylesheetInfoOriginAuthor, h.detectUri(pi), StylesheetInfoMediaTypes(h.detectMediaTypes(pi)), nil)
		list = append(list, info)
	}

	return list
}

// noNamespaceHandlerQuotedValue ports the expression
// pi.substring(start + 1, pi.indexOf(pi.charAt(start), start + 1)): the text
// between the quote character at start and its next occurrence. Like the Java
// expression it fails (with an index out of range panic) when there is no
// character at start or no closing quote.
func noNamespaceHandlerQuotedValue(pi string, start int) string {
	end := strings.IndexByte(pi[start+1:], pi[start])
	if end != -1 {
		end += start + 1
	}
	return pi[start+1 : end]
}

// detectType returns "" when the processing instruction has no type.
func (h *NoNamespaceHandler) detectType(pi string) string {
	if m := noNamespaceHandlerTypePattern.FindStringIndex(pi); m != nil {
		return noNamespaceHandlerQuotedValue(pi, m[1])
	}
	return ""
}

// detectUri returns "" when the processing instruction has no href.
func (h *NoNamespaceHandler) detectUri(pi string) string {
	if m := noNamespaceHandlerHrefPattern.FindStringIndex(pi); m != nil {
		return noNamespaceHandlerQuotedValue(pi, m[1])
	}
	return ""
}

func (h *NoNamespaceHandler) detectMediaTypes(pi string) string {
	if m := noNamespaceHandlerMediaPattern.FindStringIndex(pi); m != nil {
		return noNamespaceHandlerQuotedValue(pi, m[1])
	}
	return "screen"
}

func (h *NoNamespaceHandler) GetDefaultStylesheet() *StylesheetInfo {
	return nil
}

var _ NamespaceHandler = (*NoNamespaceHandler)(nil)
