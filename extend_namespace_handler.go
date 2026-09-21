// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/NamespaceHandler.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// NamespaceHandler provides knowledge specific to a certain document type,
// like resolving style-sheets.
//
// Methods that Java marks @Nullable return "" for null, except where a caller
// distinguishes null from the empty string: GetAttributeValue and
// GetAttributeValueWithNamespaceURI (an attribute that exists with an empty
// value) and GetLinkUri (an element with an empty href is a link).
type NamespaceHandler interface {
	GetNamespace() string

	// GetDefaultStylesheet returns nil when there is no default stylesheet
	// (Optional.empty in Java).
	GetDefaultStylesheet() *StylesheetInfo

	GetDocumentTitle(doc *dom.Document) string

	// GetStylesheets returns all links to CSS stylesheets (type="text/css")
	// in this document.
	GetStylesheets(doc *dom.Document) []*StylesheetInfo

	// GetAttributeValue may return nil. Required to return nil if attribute
	// does not exist and not nil if attribute exists.
	GetAttributeValue(e *dom.Element, attrName string) *string

	// GetAttributeValueWithNamespaceURI takes a nil namespaceURI to signify
	// any namespace, and TreeResolverNoNamespace to signify no namespace.
	GetAttributeValueWithNamespaceURI(e *dom.Element, namespaceURI *string, attrName string) *string

	GetClass(e *dom.Element) string

	GetID(e *dom.Element) string

	GetElementStyling(e *dom.Element) string

	// GetNonCssStyling returns the corresponding css properties for styling
	// that is obtained in other ways.
	GetNonCssStyling(e *dom.Element) string

	GetLang(e *dom.Element) string

	// GetLinkUri should return nil if element is not a link.
	GetLinkUri(e *dom.Element) *string

	// GetAnchorName accepts a nil element.
	GetAnchorName(e *dom.Element) string

	// IsImageElement returns true if the Element represents an image.
	IsImageElement(e *dom.Element) bool

	// IsFormElement determines whether the specified Element represents a
	// <form>.
	IsFormElement(e *dom.Element) bool

	// GetImageSourceURI, for an element where IsImageElement returns true,
	// retrieves the URI associated with that Image, as reported by the
	// element; makes no guarantee that the URI is correct, complete or points
	// to anything in particular. For elements where IsImageElement returns
	// false, this method may return "", and may also return "" if the Element
	// is not correctly formed and contains no URI; check the return value
	// carefully.
	GetImageSourceURI(e *dom.Element) string
}
