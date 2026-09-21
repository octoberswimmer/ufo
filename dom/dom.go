// Package dom is the document tree ufo renders: the part of org.w3c.dom that
// Flying Saucer reads, with the Java method names (GetNodeName, GetAttribute)
// so ported code maps onto it one to one.
package dom

import "strings"

// NodeType identifies the kind of a Node, with org.w3c.dom.Node's values.
type NodeType int

const (
	ElementNode               NodeType = 1
	AttributeNode             NodeType = 2
	TextNode                  NodeType = 3
	CDATASectionNode          NodeType = 4
	EntityReferenceNode       NodeType = 5
	ProcessingInstructionNode NodeType = 7
	CommentNode               NodeType = 8
	DocumentNode              NodeType = 9
	DocumentTypeNode          NodeType = 10
)

// Node is a node of the tree. The concrete types are *Document, *Element,
// *Text (text and CDATA sections), *Comment and *ProcessingInstruction.
//
// Methods returning a Node return a nil interface, never a typed nil pointer,
// when there is no such node, so `n.GetParentNode() == nil` is a valid test.
type Node interface {
	GetNodeType() NodeType
	GetNodeName() string
	// GetNodeValue is the character data of a text, CDATA, comment or
	// processing instruction node, and "" for the rest.
	GetNodeValue() string
	GetParentNode() Node
	GetChildNodes() []Node
	GetFirstChild() Node
	GetLastChild() Node
	GetPreviousSibling() Node
	GetNextSibling() Node
	GetOwnerDocument() *Document
	// GetTextContent is the concatenated character data of the node's text
	// descendants.
	GetTextContent() string
	HasChildNodes() bool
	AppendChild(child Node) Node
	InsertBefore(child, ref Node) Node
	RemoveChild(child Node) Node

	base() *node
}

type node struct {
	self     Node
	parent   Node
	children []Node
	document *Document
}

func (n *node) base() *node { return n }

func (n *node) GetParentNode() Node { return n.parent }

func (n *node) GetChildNodes() []Node { return n.children }

func (n *node) HasChildNodes() bool { return len(n.children) > 0 }

func (n *node) GetOwnerDocument() *Document { return n.document }

func (n *node) GetFirstChild() Node {
	if len(n.children) == 0 {
		return nil
	}
	return n.children[0]
}

func (n *node) GetLastChild() Node {
	if len(n.children) == 0 {
		return nil
	}
	return n.children[len(n.children)-1]
}

func (n *node) siblingAt(offset int) Node {
	if n.parent == nil {
		return nil
	}
	siblings := n.parent.base().children
	for i, sibling := range siblings {
		if sibling == n.self {
			if j := i + offset; j >= 0 && j < len(siblings) {
				return siblings[j]
			}
			return nil
		}
	}
	return nil
}

func (n *node) GetPreviousSibling() Node { return n.siblingAt(-1) }

func (n *node) GetNextSibling() Node { return n.siblingAt(1) }

func (n *node) GetTextContent() string {
	var b strings.Builder
	var walk func(children []Node)
	walk = func(children []Node) {
		for _, child := range children {
			switch child.GetNodeType() {
			case TextNode, CDATASectionNode:
				b.WriteString(child.GetNodeValue())
			case ElementNode, EntityReferenceNode:
				walk(child.GetChildNodes())
			}
		}
	}
	walk(n.children)
	return b.String()
}

// AppendChild adds child as the last child, removing it from its current
// parent first, and returns it.
func (n *node) AppendChild(child Node) Node {
	return n.InsertBefore(child, nil)
}

// InsertBefore inserts child before ref, or appends it when ref is nil.
func (n *node) InsertBefore(child, ref Node) Node {
	if parent := child.GetParentNode(); parent != nil {
		parent.RemoveChild(child)
	}
	child.base().parent = n.self
	if ref == nil {
		n.children = append(n.children, child)
		return child
	}
	for i, existing := range n.children {
		if existing == ref {
			n.children = append(n.children, nil)
			copy(n.children[i+1:], n.children[i:])
			n.children[i] = child
			return child
		}
	}
	panic("dom: InsertBefore reference node is not a child of this node")
}

func (n *node) RemoveChild(child Node) Node {
	for i, existing := range n.children {
		if existing == child {
			n.children = append(n.children[:i:i], n.children[i+1:]...)
			child.base().parent = nil
			return child
		}
	}
	panic("dom: RemoveChild node is not a child of this node")
}

// Document is the root of a tree.
type Document struct {
	node
	// DocumentURI is where the document was loaded from, "" when unknown.
	DocumentURI string
}

// NewDocument creates an empty document.
func NewDocument() *Document {
	d := &Document{}
	d.self = d
	d.document = d
	return d
}

func (d *Document) GetNodeType() NodeType { return DocumentNode }
func (d *Document) GetNodeName() string   { return "#document" }
func (d *Document) GetNodeValue() string  { return "" }
func (d *Document) GetTextContent() string {
	return ""
}

// GetDocumentElement is the document's root element, nil for an empty
// document.
func (d *Document) GetDocumentElement() *Element {
	for _, child := range d.children {
		if element, ok := child.(*Element); ok {
			return element
		}
	}
	return nil
}

// CreateElement creates an element with no namespace.
func (d *Document) CreateElement(tagName string) *Element {
	return d.CreateElementNS("", tagName)
}

// CreateElementNS creates an element; qualifiedName may carry a prefix.
func (d *Document) CreateElementNS(namespaceURI, qualifiedName string) *Element {
	e := &Element{namespaceURI: namespaceURI, tagName: qualifiedName}
	e.prefix, e.localName = splitQualifiedName(qualifiedName)
	e.self = e
	e.document = d
	return e
}

func (d *Document) CreateTextNode(data string) *Text {
	t := &Text{data: data}
	t.self = t
	t.document = d
	return t
}

func (d *Document) CreateCDATASection(data string) *Text {
	t := d.CreateTextNode(data)
	t.cdata = true
	return t
}

func (d *Document) CreateComment(data string) *Comment {
	c := &Comment{data: data}
	c.self = c
	c.document = d
	return c
}

func (d *Document) CreateProcessingInstruction(target, data string) *ProcessingInstruction {
	p := &ProcessingInstruction{target: target, data: data}
	p.self = p
	p.document = d
	return p
}

// GetElementByID returns the first element, in document order, whose id
// attribute has the given value.
func (d *Document) GetElementByID(id string) *Element {
	var found *Element
	var walk func(children []Node) bool
	walk = func(children []Node) bool {
		for _, child := range children {
			if element, ok := child.(*Element); ok {
				if element.GetAttribute("id") == id {
					found = element
					return true
				}
				if walk(element.children) {
					return true
				}
			}
		}
		return false
	}
	walk(d.children)
	return found
}

// GetElementsByTagName returns the elements with the given tag name in
// document order; "*" matches every element.
func (d *Document) GetElementsByTagName(name string) []*Element {
	return elementsByTagName(d.children, name)
}

func elementsByTagName(children []Node, name string) []*Element {
	var found []*Element
	var walk func(children []Node)
	walk = func(children []Node) {
		for _, child := range children {
			if element, ok := child.(*Element); ok {
				if name == "*" || element.tagName == name {
					found = append(found, element)
				}
				walk(element.children)
			}
		}
	}
	walk(children)
	return found
}

// Attr is one attribute of an element.
type Attr struct {
	NamespaceURI string
	// Name is the qualified name as written, prefix included.
	Name      string
	LocalName string
	Value     string
}

// Element is an element node.
type Element struct {
	node
	namespaceURI string
	tagName      string
	prefix       string
	localName    string
	attributes   []Attr
}

func (e *Element) GetNodeType() NodeType { return ElementNode }
func (e *Element) GetNodeName() string   { return e.tagName }
func (e *Element) GetNodeValue() string  { return "" }

// GetTagName is the qualified name as written, prefix included.
func (e *Element) GetTagName() string { return e.tagName }

// GetLocalName is the name without its prefix.
func (e *Element) GetLocalName() string { return e.localName }

func (e *Element) GetPrefix() string { return e.prefix }

// GetNamespaceURI is "" for an element in no namespace.
func (e *Element) GetNamespaceURI() string { return e.namespaceURI }

// GetAttributes returns the attributes in document order.
func (e *Element) GetAttributes() []Attr { return e.attributes }

// GetAttribute returns the value of the attribute with the given qualified
// name, and "" when the element has none, as org.w3c.dom does.
func (e *Element) GetAttribute(name string) string {
	for _, attr := range e.attributes {
		if attr.Name == name {
			return attr.Value
		}
	}
	return ""
}

func (e *Element) HasAttribute(name string) bool {
	for _, attr := range e.attributes {
		if attr.Name == name {
			return true
		}
	}
	return false
}

func (e *Element) GetAttributeNS(namespaceURI, localName string) string {
	for _, attr := range e.attributes {
		if attr.NamespaceURI == namespaceURI && attr.LocalName == localName {
			return attr.Value
		}
	}
	return ""
}

func (e *Element) HasAttributeNS(namespaceURI, localName string) bool {
	for _, attr := range e.attributes {
		if attr.NamespaceURI == namespaceURI && attr.LocalName == localName {
			return true
		}
	}
	return false
}

func (e *Element) SetAttribute(name, value string) {
	e.SetAttributeNS("", name, value)
}

func (e *Element) SetAttributeNS(namespaceURI, qualifiedName, value string) {
	_, localName := splitQualifiedName(qualifiedName)
	for i := range e.attributes {
		if e.attributes[i].NamespaceURI == namespaceURI && e.attributes[i].LocalName == localName {
			e.attributes[i].Name = qualifiedName
			e.attributes[i].Value = value
			return
		}
	}
	e.attributes = append(e.attributes, Attr{NamespaceURI: namespaceURI, Name: qualifiedName, LocalName: localName, Value: value})
}

func (e *Element) RemoveAttribute(name string) {
	for i, attr := range e.attributes {
		if attr.Name == name {
			e.attributes = append(e.attributes[:i:i], e.attributes[i+1:]...)
			return
		}
	}
}

func (e *Element) GetElementsByTagName(name string) []*Element {
	return elementsByTagName(e.children, name)
}

// Text is a text node or a CDATA section.
type Text struct {
	node
	data  string
	cdata bool
}

func (t *Text) GetNodeType() NodeType {
	if t.cdata {
		return CDATASectionNode
	}
	return TextNode
}

func (t *Text) GetNodeName() string {
	if t.cdata {
		return "#cdata-section"
	}
	return "#text"
}

func (t *Text) GetNodeValue() string   { return t.data }
func (t *Text) GetData() string        { return t.data }
func (t *Text) SetData(data string)    { t.data = data }
func (t *Text) GetTextContent() string { return t.data }

// Comment is a comment node.
type Comment struct {
	node
	data string
}

func (c *Comment) GetNodeType() NodeType  { return CommentNode }
func (c *Comment) GetNodeName() string    { return "#comment" }
func (c *Comment) GetNodeValue() string   { return c.data }
func (c *Comment) GetData() string        { return c.data }
func (c *Comment) GetTextContent() string { return c.data }

// ProcessingInstruction is a processing instruction such as
// <?xml-stylesheet href="a.css"?>.
type ProcessingInstruction struct {
	node
	target string
	data   string
}

func (p *ProcessingInstruction) GetNodeType() NodeType  { return ProcessingInstructionNode }
func (p *ProcessingInstruction) GetNodeName() string    { return p.target }
func (p *ProcessingInstruction) GetNodeValue() string   { return p.data }
func (p *ProcessingInstruction) GetTarget() string      { return p.target }
func (p *ProcessingInstruction) GetData() string        { return p.data }
func (p *ProcessingInstruction) GetTextContent() string { return p.data }

func splitQualifiedName(qualifiedName string) (prefix, localName string) {
	if i := strings.IndexByte(qualifiedName, ':'); i >= 0 {
		return qualifiedName[:i], qualifiedName[i+1:]
	}
	return "", qualifiedName
}
