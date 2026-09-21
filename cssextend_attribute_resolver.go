// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/extend/AttributeResolver.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// AttributeResolver: in XML, an application may or may not know how to find
// the ID and/or class and/or attribute defaults of an element.
//
// To enable matching of identity conditions, class conditions, language, and
// attribute defaults you need to provide an AttributeResolver to the StyleMap.
//
// NOTE: The application is required to look in a document's internal subset
// for default attribute values, but the application is not required to use its
// built-in knowledge of a namespace or look in the external subset.
//
// The matcher distinguishes a missing attribute (Java null) from an empty
// one, so the two GetAttributeValue methods return *string. The other string
// results are only tested for null together with a test that an empty string
// fails as well, so they are plain strings with "" for null.
type AttributeResolver interface {
	// GetAttributeValue is required to return nil if attribute does not
	// exist, and not nil if attribute exists.
	GetAttributeValue(e dom.Node, attrName string) *string

	// GetAttributeValueWithNamespaceURI is required to return nil if
	// attribute does not exist and not nil if attribute exists. A nil
	// namespaceURI matches an attribute in any namespace; a pointer to
	// TreeResolverNoNamespace matches an attribute in no namespace.
	GetAttributeValueWithNamespaceURI(e dom.Node, namespaceURI *string, attrName string) *string

	GetClass(e dom.Node) string

	GetID(e dom.Node) string

	// GetNonCssStyling returns the non css styling (specificity 0,0,0,0 on
	// author styles, according to css 2.1)
	GetNonCssStyling(e dom.Node) string

	// GetElementStyling returns the elementStyling value (corresponding to
	// xhtml style attribute, specificity 1,0,0,0 according to css 2.1)
	GetElementStyling(e dom.Node) string

	GetLang(e dom.Node) string

	// IsLink gets the link attribute of the AttributeResolver object
	IsLink(e dom.Node) bool

	// IsVisited gets the visited attribute of the AttributeResolver object
	IsVisited(e dom.Node) bool

	// IsHover gets the hover attribute of the AttributeResolver object
	IsHover(e dom.Node) bool

	// IsActive gets the active attribute of the AttributeResolver object
	IsActive(e dom.Node) bool

	// IsFocus gets the focus attribute of the AttributeResolver object
	IsFocus(e dom.Node) bool
}
