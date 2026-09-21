// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/simple/extend/XhtmlCssOnlyNamespaceHandler.java

package ufo

import (
	"embed"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/octoberswimmer/ufo/dom"
)

// XhtmlCssOnlyNamespaceHandler handles xhtml but only css styling is honored,
// no presentational html attributes (see css 2.1 spec, 6.4.4).
type XhtmlCssOnlyNamespaceHandler struct {
	NoNamespaceHandler
}

const xhtmlCssOnlyNamespaceHandlerNamespace = "http://www.w3.org/1999/xhtml"

// xhtmlCssOnlyNamespaceHandlerResources holds the files Java reads from the
// classpath under /resources/css/.
//
//go:embed resources/css
var xhtmlCssOnlyNamespaceHandlerResources embed.FS

// XhtmlCssOnlyNamespaceHandlerDefaultStylesheetScheme starts the URI of the
// default stylesheet. Java names the stylesheet by the jar: or file: URL of
// the classpath resource and the user agent opens it; an embedded file has no
// URL, so the StylesheetInfo carries the content of the file and this URI
// identifies it in the stylesheet cache.
const XhtmlCssOnlyNamespaceHandlerDefaultStylesheetScheme = "embed:"

var (
	xhtmlCssOnlyNamespaceHandlerDefaultStylesheet     *StylesheetInfo
	xhtmlCssOnlyNamespaceHandlerDefaultStylesheetLock sync.Mutex
	xhtmlCssOnlyNamespaceHandlerInlineCssCounter      atomic.Int64
)

func NewXhtmlCssOnlyNamespaceHandler() *XhtmlCssOnlyNamespaceHandler {
	return &XhtmlCssOnlyNamespaceHandler{}
}

// GetNamespace gets the namespace attribute of the XhtmlNamespaceHandler
// object.
func (h *XhtmlCssOnlyNamespaceHandler) GetNamespace() string {
	return xhtmlCssOnlyNamespaceHandlerNamespace
}

// GetClass gets the class attribute of the XhtmlNamespaceHandler object.
func (h *XhtmlCssOnlyNamespaceHandler) GetClass(e *dom.Element) string {
	return e.GetAttribute("class")
}

// GetID gets the iD attribute of the XhtmlNamespaceHandler object.
func (h *XhtmlCssOnlyNamespaceHandler) GetID(e *dom.Element) string {
	return h.GetAttribute(e, "id")
}

// GetAttribute returns the trimmed value of the attribute, and "" (null in
// Java) when the element has no such attribute or the trimmed value is empty.
func (h *XhtmlCssOnlyNamespaceHandler) GetAttribute(e *dom.Element, attrName string) string {
	return styleBuilderTrim(e.GetAttribute(attrName))
}

// GetElementStyling gets the elementStyling attribute of the
// XhtmlNamespaceHandler object.
func (h *XhtmlCssOnlyNamespaceHandler) GetElementStyling(e *dom.Element) string {
	style := NewStyleBuilder()
	switch e.GetNodeName() {
	case "td", "th":
		style.Append(e, "colspan", "-fs-table-cell-colspan: ")
		style.Append(e, "rowspan", "-fs-table-cell-rowspan: ")
	case "img", "svg":
		style.AppendWidth(e)
		style.AppendHeight(e)
	case "colgroup", "col":
		style.Append(e, "span", "-fs-table-cell-colspan: ")
		style.AppendWidth(e)
	}
	style.AppendRawStyle(e.GetAttribute("style"))
	return style.String()
}

// GetLinkUri gets the linkUri attribute of the XhtmlNamespaceHandler object.
func (h *XhtmlCssOnlyNamespaceHandler) GetLinkUri(e *dom.Element) *string {
	if strings.EqualFold(e.GetNodeName(), "a") && e.HasAttribute("href") {
		href := e.GetAttribute("href")
		return &href
	}
	return nil
}

func (h *XhtmlCssOnlyNamespaceHandler) GetAnchorName(e *dom.Element) string {
	if e != nil && strings.EqualFold("a", e.GetNodeName()) && e.HasAttribute("name") {
		return e.GetAttribute("name")
	}
	return ""
}

// xhtmlCssOnlyNamespaceHandlerIsWhitespace ports Character.isWhitespace: a
// space, line or paragraph separator other than the non-breaking spaces, or
// one of the control characters U+0009 to U+000D and U+001C to U+001F.
func xhtmlCssOnlyNamespaceHandlerIsWhitespace(c rune) bool {
	switch {
	case c >= 0x09 && c <= 0x0D, c >= 0x1C && c <= 0x1F:
		return true
	case c == 0x00A0, c == 0x2007, c == 0x202F:
		return false
	}
	return unicode.In(c, unicode.Zs, unicode.Zl, unicode.Zp)
}

func XhtmlCssOnlyNamespaceHandlerCollapseWhiteSpace(text string) string {
	var last rune = '?'
	var result strings.Builder
	result.Grow(len(text))
	for _, c := range text {
		if xhtmlCssOnlyNamespaceHandlerIsWhitespace(c) {
			c = ' '
		}
		if c != ' ' || last != ' ' {
			result.WriteRune(c)
			last = c
		}
	}
	return result.String()
}

// GetDocumentTitle returns the title of the document as located in the
// contents of /html/head/title, or "" if none could be found.
func (h *XhtmlCssOnlyNamespaceHandler) GetDocumentTitle(doc *dom.Document) string {
	title := ""

	html := doc.GetDocumentElement()
	head := h.findFirstChild(html, "head")
	if head != nil {
		titleElem := h.findFirstChild(head, "title")
		if titleElem != nil {
			title = XhtmlCssOnlyNamespaceHandlerCollapseWhiteSpace(styleBuilderTrim(TextUtilReadTextContent(titleElem)))
		}
	}
	return title
}

func (h *XhtmlCssOnlyNamespaceHandler) findFirstChild(parent *dom.Element, targetName string) *dom.Element {
	for _, n := range parent.GetChildNodes() {
		if n.GetNodeType() == dom.ElementNode && n.GetNodeName() == targetName {
			return n.(*dom.Element)
		}
	}

	return nil
}

// ReadStyleElement returns nil when the style element is empty.
func (h *XhtmlCssOnlyNamespaceHandler) ReadStyleElement(style *dom.Element) *StylesheetInfo {
	css := xhtmlCssOnlyNamespaceHandlerExtractContent(style)
	if css == "" {
		return nil
	}

	media := style.GetAttribute("media")
	uri := "inline:" + strconv.FormatInt(xhtmlCssOnlyNamespaceHandlerInlineCssCounter.Add(1), 10) // just some unique value to cache by
	return NewStylesheetInfo(StylesheetInfoOriginAuthor, uri, StylesheetInfoMediaTypes(media), &css)
}

// xhtmlCssOnlyNamespaceHandlerExtractContent concatenates the data of the
// CharacterData children of the style element: text, CDATA sections and
// comments.
func xhtmlCssOnlyNamespaceHandlerExtractContent(style *dom.Element) string {
	var buf strings.Builder
	current := style.GetFirstChild()
	for current != nil {
		switch characterData := current.(type) {
		case *dom.Text:
			buf.WriteString(characterData.GetData())
		case *dom.Comment:
			buf.WriteString(characterData.GetData())
		}
		current = current.GetNextSibling()
	}
	return styleBuilderTrim(buf.String())
}

// ReadLinkElement returns nil when the link element is not a stylesheet link.
func (h *XhtmlCssOnlyNamespaceHandler) ReadLinkElement(link *dom.Element) *StylesheetInfo {
	rel := strings.ToLower(link.GetAttribute("rel"))
	if strings.Contains(rel, "alternate") {
		return nil
	} //DON'T get alternate stylesheets
	if !strings.Contains(rel, "stylesheet") {
		return nil
	}

	uri := link.GetAttribute("href")
	return NewStylesheetInfo(StylesheetInfoOriginAuthor, uri, StylesheetInfoMediaTypes(link.GetAttribute("media")), nil)
}

// GetStylesheets gets the stylesheetLinks attribute of the
// XhtmlNamespaceHandler object.
func (h *XhtmlCssOnlyNamespaceHandler) GetStylesheets(doc *dom.Document) []*StylesheetInfo {
	//get the processing-instructions (actually for XmlDocuments)
	result := append([]*StylesheetInfo{}, h.NoNamespaceHandler.GetStylesheets(doc)...)

	//get the link elements
	html := doc.GetDocumentElement()
	head := h.findFirstChild(html, "head")
	if head != nil {
		current := head.GetFirstChild()
		for current != nil {
			if current.GetNodeType() == dom.ElementNode {
				elem := current.(*dom.Element)
				elemName := elem.GetLocalName()
				if elemName == "" {
					elemName = elem.GetTagName()
				}
				var info *StylesheetInfo
				switch elemName {
				case "link":
					info = h.ReadLinkElement(elem)
				case "style":
					info = h.ReadStyleElement(elem)
				}
				if info != nil {
					result = append(result, info)
				}
			}
			current = current.GetNextSibling()
		}
	}

	return result
}

func (h *XhtmlCssOnlyNamespaceHandler) GetDefaultStylesheet() *StylesheetInfo {
	xhtmlCssOnlyNamespaceHandlerDefaultStylesheetLock.Lock()
	defer xhtmlCssOnlyNamespaceHandlerDefaultStylesheetLock.Unlock()
	if xhtmlCssOnlyNamespaceHandlerDefaultStylesheet == nil {
		uri, content := h.getDefaultStylesheetUrl()
		xhtmlCssOnlyNamespaceHandlerDefaultStylesheet = NewStylesheetInfo(StylesheetInfoOriginUserAgent, uri, StylesheetInfoMediaTypes(""), &content)
	}
	return xhtmlCssOnlyNamespaceHandlerDefaultStylesheet
}

// getDefaultStylesheetUrl returns the URI of the default stylesheet and the
// content of the embedded file the URI names.
func (h *XhtmlCssOnlyNamespaceHandler) getDefaultStylesheetUrl() (string, string) {
	defaultStyleSheet := ConfigurationValueFor("xr.css.user-agent-default-css") + "XhtmlNamespaceHandler.css"
	content, err := xhtmlCssOnlyNamespaceHandlerResources.ReadFile(strings.TrimPrefix(defaultStyleSheet, "/"))
	if err != nil {
		panic(NewXRRuntimeExceptionWithCause("Can't load default CSS from "+defaultStyleSheet+"."+
			"This file must be on your CLASSPATH. Please check before continuing.", err))
	}
	return XhtmlCssOnlyNamespaceHandlerDefaultStylesheetScheme + defaultStyleSheet, string(content)
}

func (h *XhtmlCssOnlyNamespaceHandler) getMetaInfo(doc *dom.Document) map[string]string {
	metadata := make(map[string]string)

	html := doc.GetDocumentElement()
	head := h.findFirstChild(html, "head")
	if head != nil {
		current := head.GetFirstChild()
		for current != nil {
			if current.GetNodeType() == dom.ElementNode {
				elem := current.(*dom.Element)
				elemName := elem.GetLocalName()
				if elemName == "" {
					elemName = elem.GetTagName()
				}
				if elemName == "meta" {
					http_equiv := elem.GetAttribute("http-equiv")
					content := elem.GetAttribute("content")

					if http_equiv != "" && content != "" {
						metadata[http_equiv] = content
					}
				}
			}
			current = current.GetNextSibling()
		}
	}

	return metadata
}

// GetLang accepts a nil element.
func (h *XhtmlCssOnlyNamespaceHandler) GetLang(e *dom.Element) string {
	if e == nil {
		return ""
	}
	lang := e.GetAttribute("lang")
	if lang == "" {
		lang = h.getMetaInfo(e.GetOwnerDocument())["Content-Language"]
	}
	return lang
}

var _ NamespaceHandler = (*XhtmlCssOnlyNamespaceHandler)(nil)
