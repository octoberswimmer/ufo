// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/extend/lib/DOMTreeResolver.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// DOMTreeResolver works for a dom tree.
type DOMTreeResolver struct{}

func NewDOMTreeResolver() *DOMTreeResolver {
	return &DOMTreeResolver{}
}

func (r *DOMTreeResolver) GetParentElement(element dom.Node) dom.Node {
	parent := element.GetParentNode()
	if parent.GetNodeType() != dom.ElementNode {
		parent = nil
	}
	return parent
}

func (r *DOMTreeResolver) GetPreviousSiblingElement(element dom.Node) dom.Node {
	sibling := element.GetPreviousSibling()
	for sibling != nil && sibling.GetNodeType() != dom.ElementNode {
		sibling = sibling.GetPreviousSibling()
	}
	if sibling == nil || sibling.GetNodeType() != dom.ElementNode {
		return nil
	}
	return sibling
}

// domTreeResolverLocalName is org.w3c.dom.Node.getLocalName: the local name
// of an element, and no value (false) for every other kind of node.
func domTreeResolverLocalName(node dom.Node) (string, bool) {
	if element, ok := node.(*dom.Element); ok {
		return element.GetLocalName(), true
	}
	return "", false
}

// domTreeResolverNamespaceURI is org.w3c.dom.Node.getNamespaceURI: the
// namespace of an element that has one, and no value (false) otherwise.
func domTreeResolverNamespaceURI(node dom.Node) (string, bool) {
	if element, ok := node.(*dom.Element); ok && element.GetNamespaceURI() != "" {
		return element.GetNamespaceURI(), true
	}
	return "", false
}

func (r *DOMTreeResolver) GetElementName(element dom.Node) string {
	name, ok := domTreeResolverLocalName(element)
	if !ok {
		name = element.GetNodeName()
	}
	return name
}

func (r *DOMTreeResolver) IsFirstChildElement(element dom.Node) bool {
	parent := element.GetParentNode()
	currentChild := parent.GetFirstChild()
	for currentChild != nil && currentChild.GetNodeType() != dom.ElementNode {
		currentChild = currentChild.GetNextSibling()
	}
	return currentChild == element
}

func (r *DOMTreeResolver) IsLastChildElement(element dom.Node) bool {
	parent := element.GetParentNode()
	currentChild := parent.GetLastChild()
	for currentChild != nil && currentChild.GetNodeType() != dom.ElementNode {
		currentChild = currentChild.GetPreviousSibling()
	}
	return currentChild == element
}

func (r *DOMTreeResolver) MatchesElement(element dom.Node, namespaceURI *string, name string) bool {
	localName, hasLocalName := domTreeResolverLocalName(element)
	eName := localName
	if !hasLocalName {
		eName = element.GetNodeName()
	}
	if namespaceURI != nil {
		// An element in no namespace has a null namespace URI in Java, which
		// equals no string. A selector for TreeResolverNoNamespace ("|name")
		// therefore matches no element: the Java branch that was written for
		// it comes after this one and tests a non-null value for being null,
		// so it never runs.
		elementNamespaceURI, hasNamespaceURI := domTreeResolverNamespaceURI(element)
		return hasLocalName && name == localName && hasNamespaceURI && *namespaceURI == elementNamespaceURI
	} else /* if (namespaceURI == null) */ {
		return name == eName
	}
}

func (r *DOMTreeResolver) GetPositionOfElement(element dom.Node) int {
	parent := element.GetParentNode()
	nl := parent.GetChildNodes()

	eltCount := 0
	i := 0
	for i < len(nl) {
		if nl[i].GetNodeType() == dom.ElementNode {
			if nl[i] == element {
				return eltCount
			} else {
				eltCount++
			}
		}
		i++
	}

	//should not happen
	return -1
}
