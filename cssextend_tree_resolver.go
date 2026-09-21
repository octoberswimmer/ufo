// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/extend/TreeResolver.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// XXX Where should this go (used by parser, TreeResolver, and AttributeResolver
const TreeResolverNoNamespace = ""

// TreeResolver gives the css matcher access to the information it needs about
// the tree structure.
//
// Elements are the "things" in the tree structure that can be matched by the
// matcher.
type TreeResolver interface {
	// GetParentElement returns the parent element of an element, or nil if
	// this was the root element
	GetParentElement(element dom.Node) dom.Node

	// GetElementName returns the name of the element so that it may match
	// against the selectors
	GetElementName(element dom.Node) string

	// GetPreviousSiblingElement returns the previous sibling element, or nil
	// if none exists
	GetPreviousSiblingElement(node dom.Node) dom.Node

	// IsFirstChildElement returns true if this element is the first child
	// element of its parent
	IsFirstChildElement(element dom.Node) bool

	// IsLastChildElement returns true if this element is the last child
	// element of its parent
	IsLastChildElement(element dom.Node) bool

	// GetPositionOfElement returns the index of the position of the submitted
	// element among its element node siblings. It returns -1 in case of
	// error, 0 indexed position otherwise.
	GetPositionOfElement(element dom.Node) int

	// MatchesElement returns true if element has the local name name and
	// namespace URI namespaceURI.
	//
	// namespaceURI is the namespace to match, may be nil to signify any
	// namespace. Use a pointer to TreeResolverNoNamespace to signify that name
	// should only match when there is no namespace defined on element. name
	// is the name to match, may not be empty.
	MatchesElement(element dom.Node, namespaceURI *string, name string) bool
}
