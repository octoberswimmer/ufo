// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/context/StandardAttributeResolver.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// StandardAttributeResolver is an instance which works together with a w3c
// DOM tree.
type StandardAttributeResolver struct {
	nsh NamespaceHandler
	uac UserAgentCallback
	ui  UserInterface
	// classAttributeCache is keyed by node identity (IdentityHashMap in
	// Java): dom nodes are pointers, so the interface value compares by
	// identity.
	classAttributeCache map[dom.Node]string
}

func NewStandardAttributeResolver(nsh NamespaceHandler, uac UserAgentCallback, ui UserInterface) *StandardAttributeResolver {
	return &StandardAttributeResolver{
		nsh:                 nsh,
		uac:                 uac,
		ui:                  ui,
		classAttributeCache: make(map[dom.Node]string),
	}
}

// GetAttributeValue gets the attributeValue attribute of the
// StandardAttributeResolver object.
func (r *StandardAttributeResolver) GetAttributeValue(e dom.Node, attrName string) *string {
	return r.nsh.GetAttributeValue(e.(*dom.Element), attrName)
}

func (r *StandardAttributeResolver) GetAttributeValueWithNamespaceURI(e dom.Node, namespaceURI *string, attrName string) *string {
	return r.nsh.GetAttributeValueWithNamespaceURI(e.(*dom.Element), namespaceURI, attrName)
}

// GetClass gets the class attribute of the StandardAttributeResolver object.
// Map.computeIfAbsent stores no mapping for a null result, so a class of ""
// (null in Java) is asked of the namespace handler again on the next call.
func (r *StandardAttributeResolver) GetClass(e dom.Node) string {
	if class, ok := r.classAttributeCache[e]; ok {
		return class
	}
	class := r.nsh.GetClass(e.(*dom.Element))
	if class != "" {
		r.classAttributeCache[e] = class
	}
	return class
}

// GetID gets the iD attribute of the StandardAttributeResolver object.
func (r *StandardAttributeResolver) GetID(e dom.Node) string {
	return r.nsh.GetID(e.(*dom.Element))
}

func (r *StandardAttributeResolver) GetNonCssStyling(e dom.Node) string {
	return r.nsh.GetNonCssStyling(e.(*dom.Element))
}

// GetElementStyling gets the elementStyling attribute of the
// StandardAttributeResolver object.
func (r *StandardAttributeResolver) GetElementStyling(e dom.Node) string {
	return r.nsh.GetElementStyling(e.(*dom.Element))
}

// GetLang gets the lang attribute of the StandardAttributeResolver object.
func (r *StandardAttributeResolver) GetLang(e dom.Node) string {
	return r.nsh.GetLang(e.(*dom.Element))
}

// IsLink gets the link attribute of the StandardAttributeResolver object.
func (r *StandardAttributeResolver) IsLink(e dom.Node) bool {
	return r.nsh.GetLinkUri(e.(*dom.Element)) != nil
}

// IsVisited gets the visited attribute of the StandardAttributeResolver
// object.
func (r *StandardAttributeResolver) IsVisited(e dom.Node) bool {
	return r.IsLink(e) && r.uac.IsVisited(*r.nsh.GetLinkUri(e.(*dom.Element)))
}

// IsHover gets the hover attribute of the StandardAttributeResolver object.
func (r *StandardAttributeResolver) IsHover(e dom.Node) bool {
	return r.ui.IsHover(e.(*dom.Element))
}

// IsActive gets the active attribute of the StandardAttributeResolver object.
func (r *StandardAttributeResolver) IsActive(e dom.Node) bool {
	return r.ui.IsActive(e.(*dom.Element))
}

// IsFocus gets the focus attribute of the StandardAttributeResolver object.
func (r *StandardAttributeResolver) IsFocus(e dom.Node) bool {
	return r.ui.IsFocus(e.(*dom.Element))
}

var _ AttributeResolver = (*StandardAttributeResolver)(nil)
