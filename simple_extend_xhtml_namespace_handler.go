// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/simple/extend/XhtmlNamespaceHandler.java

package ufo

import (
	"regexp"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// XhtmlNamespaceHandler handles xhtml documents, including presentational
// html attributes (see css 2.1 spec, 6.4.4). In this class ONLY handling (css
// equivalents) of presentational properties (according to css 2.1 spec,
// section 6.4.4) should be specified.
type XhtmlNamespaceHandler struct {
	XhtmlCssOnlyNamespaceHandler
}

// Java applies the pattern with Matcher.matches, hence the anchors.
var xhtmlNamespaceHandlerReMangledColor = regexp.MustCompile(`^[0-9a-f]{6}$`)

func NewXhtmlNamespaceHandler() *XhtmlNamespaceHandler {
	return &XhtmlNamespaceHandler{}
}

func (h *XhtmlNamespaceHandler) IsImageElement(e *dom.Element) bool {
	return strings.EqualFold(e.GetNodeName(), "img")
}

func (h *XhtmlNamespaceHandler) IsFormElement(e *dom.Element) bool {
	return strings.EqualFold(e.GetNodeName(), "form")
}

func (h *XhtmlNamespaceHandler) GetImageSourceURI(e *dom.Element) string {
	return e.GetAttribute("src")
}

func (h *XhtmlNamespaceHandler) GetNonCssStyling(e *dom.Element) string {
	switch e.GetNodeName() {
	case "table":
		return h.applyTableStyles(e)
	case "td", "th":
		return h.applyTableCellStyles(e)
	case "tr":
		return h.applyTableRowStyles(e)
	case "img":
		return h.applyImgStyles(e)
	case "p", "div":
		return h.applyBlockAlign(e)
	default:
		return ""
	}
}

func (h *XhtmlNamespaceHandler) applyBlockAlign(e *dom.Element) string {
	s := strings.ToLower(styleBuilderTrim(e.GetAttribute("align")))
	switch s {
	case "left",
		"right",
		"center",
		"justify":
		return "text-align: " + s + ";"
	default:
		return ""
	}
}

func (h *XhtmlNamespaceHandler) applyImgStyles(e *dom.Element) string {
	style := NewStyleBuilder()
	style.ApplyFloatingAlign(e)
	return style.String()
}

func (h *XhtmlNamespaceHandler) applyTableCellStyles(e *dom.Element) string {
	style := NewStyleBuilder()
	// check for cell padding
	table := h.FindTable(e)
	if table != nil {
		style.AppendLength(table, "cellpadding", "padding: ")
		s := h.GetAttribute(table, "border")
		if s != "" && s != "0" {
			style.AppendRawStyle("border: 1px outset black;")
		}
	}
	style.AppendWidth(e)
	style.AppendHeight(e)
	style.ApplyTableContentAlign(e)
	h.appendBackgroundColor(e, style)
	h.appendBackgroundImage(e, style)
	return style.String()
}

func (h *XhtmlNamespaceHandler) appendBackgroundColor(e *dom.Element, style *StyleBuilder) {
	s := styleBuilderTrim(e.GetAttribute("bgcolor"))
	if s != "" {
		color := s
		if h.LooksLikeAMangledColor(s) {
			color = "#" + s
		}
		style.AppendStyle("background-color: ", color)
	}
}

func (h *XhtmlNamespaceHandler) appendBackgroundImage(e *dom.Element, style *StyleBuilder) {
	style.AppendUrl(e, "background", "background-image: ")
}

func (h *XhtmlNamespaceHandler) applyTableStyles(e *dom.Element) string {
	style := NewStyleBuilder()
	style.AppendLength(e, "width", "width: ")
	style.AppendLengthWithSuffix(e, "border", "border: ", " inset black;")
	style.AppendLength(e, "cellspacing", "border-collapse: separate; border-spacing: ")
	h.appendBackgroundColor(e, style)
	h.appendBackgroundImage(e, style)
	style.ApplyFloatingAlign(e)
	return style.String()
}

func (h *XhtmlNamespaceHandler) applyTableRowStyles(e *dom.Element) string {
	style := NewStyleBuilder()
	style.ApplyTableContentAlign(e)
	return style.String()
}

func (h *XhtmlNamespaceHandler) LooksLikeAMangledColor(s string) bool {
	return xhtmlNamespaceHandlerReMangledColor.MatchString(s)
}

// FindTable returns the table element among the five nearest ancestors of the
// cell, or nil.
func (h *XhtmlNamespaceHandler) FindTable(cell dom.Node) *dom.Element {
	return h.Ancestor(cell, "table", 5)
}

// Ancestor returns the nearest ancestor element with the given tag name that
// is at most maxDepth levels above the element, or nil.
func (h *XhtmlNamespaceHandler) Ancestor(element dom.Node, tagName string, maxDepth int) *dom.Element {
	parent := element.GetParentNode()
	if parent == nil || maxDepth <= 0 {
		return nil
	}
	if parent.GetNodeType() == dom.ElementNode && parent.GetNodeName() == tagName {
		return parent.(*dom.Element)
	}
	return h.Ancestor(parent, tagName, maxDepth-1)
}

var _ NamespaceHandler = (*XhtmlNamespaceHandler)(nil)
