// No Flying Saucer class is ported here. This file holds the counterparts of
// the JDK classes that DocumentSplitter and SAXEventRecorder are written
// against: org.xml.sax.ContentHandler, Attributes and Locator, a
// namespace-aware SAX parser (SAXParse over a stream, SAXParseDocument over a
// tree), and the TransformerHandler with a DOMResult that DocumentSplitter
// builds its documents with (saxDOMBuilder).

package pdf

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/octoberswimmer/ufo/dom"
)

const saxXmlnsNamespace = "http://www.w3.org/2000/xmlns/"

// SAXContentHandler is org.xml.sax.ContentHandler. A method that throws
// SAXException in Java returns an error. Character data arrives as a string
// instead of a slice of a char array that the parser reuses.
type SAXContentHandler interface {
	SetDocumentLocator(locator SAXLocator)
	StartDocument() error
	EndDocument() error
	StartPrefixMapping(prefix string, uri string) error
	EndPrefixMapping(prefix string) error
	StartElement(uri string, localName string, qName string, attributes SAXAttributes) error
	EndElement(uri string, localName string, qName string) error
	Characters(ch string) error
	IgnorableWhitespace(ch string) error
	ProcessingInstruction(target string, data string) error
	SkippedEntity(name string) error
}

// SAXLocator is org.xml.sax.Locator. The parsers of this file do not supply
// one.
type SAXLocator interface {
	GetPublicId() string
	GetSystemId() string
	GetLineNumber() int
	GetColumnNumber() int
}

// SAXAttribute is one entry of SAXAttributes.
type SAXAttribute struct {
	URI       string
	LocalName string
	QName     string
	Value     string
}

// SAXAttributes is org.xml.sax.Attributes; an empty or nil slice is
// new AttributesImpl(). Namespace declarations are not among the attributes,
// as with a SAX parser whose namespace-prefixes feature is off.
type SAXAttributes []SAXAttribute

// SAXParse reads a well-formed XML document and reports it to handler the
// way a namespace-aware, non-validating SAX parser does. Comments and the
// document type declaration are not reported, since ContentHandler has no
// method for them. The named HTML entities are accepted, as in dom.ParseXML.
func SAXParse(r io.Reader, handler SAXContentHandler) error {
	decoder := xml.NewDecoder(r)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = true

	if err := handler.StartDocument(); err != nil {
		return err
	}

	type frame struct {
		// namespaces maps a prefix to its URI for the open element.
		namespaces map[string]string
		// declared holds the prefixes the element itself declares.
		declared  []string
		uri       string
		localName string
		qName     string
	}
	frames := []frame{{namespaces: map[string]string{}}}
	depth := 0
	seenRoot := false

	for {
		token, err := decoder.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("pdf: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			parent := frames[len(frames)-1]
			f := frame{namespaces: map[string]string{}}
			for prefix, uri := range parent.namespaces {
				f.namespaces[prefix] = uri
			}
			var attributes SAXAttributes
			for _, attr := range t.Attr {
				switch {
				case attr.Name.Space == "" && attr.Name.Local == "xmlns":
					f.namespaces[""] = attr.Value
					f.declared = append(f.declared, "")
				case attr.Name.Space == "xmlns":
					f.namespaces[attr.Name.Local] = attr.Value
					f.declared = append(f.declared, attr.Name.Local)
				}
			}
			for _, attr := range t.Attr {
				if attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns") {
					continue
				}
				uri := ""
				switch {
				case attr.Name.Space == "xml":
					uri = "http://www.w3.org/XML/1998/namespace"
				case attr.Name.Space != "":
					uri = f.namespaces[attr.Name.Space]
				}
				attributes = append(attributes, SAXAttribute{URI: uri, LocalName: attr.Name.Local, QName: saxRawName(attr.Name), Value: attr.Value})
			}
			f.uri = f.namespaces[t.Name.Space]
			f.localName = t.Name.Local
			f.qName = saxRawName(t.Name)
			for _, prefix := range f.declared {
				if err := handler.StartPrefixMapping(prefix, f.namespaces[prefix]); err != nil {
					return err
				}
			}
			if err := handler.StartElement(f.uri, f.localName, f.qName, attributes); err != nil {
				return err
			}
			frames = append(frames, f)
			depth++
			seenRoot = true
		case xml.EndElement:
			if depth == 0 || frames[len(frames)-1].qName != saxRawName(t.Name) {
				return fmt.Errorf("pdf: unexpected end tag </%s>", saxRawName(t.Name))
			}
			f := frames[len(frames)-1]
			frames = frames[:len(frames)-1]
			depth--
			if err := handler.EndElement(f.uri, f.localName, f.qName); err != nil {
				return err
			}
			for _, prefix := range f.declared {
				if err := handler.EndPrefixMapping(prefix); err != nil {
					return err
				}
			}
		case xml.CharData:
			if depth == 0 {
				// Whitespace outside the root element.
				continue
			}
			if err := handler.Characters(string(t)); err != nil {
				return err
			}
		case xml.ProcInst:
			if t.Target == "xml" {
				continue
			}
			if err := handler.ProcessingInstruction(t.Target, string(t.Inst)); err != nil {
				return err
			}
		}
	}
	if depth != 0 {
		return fmt.Errorf("pdf: unexpected end of document inside <%s>", frames[len(frames)-1].qName)
	}
	if !seenRoot {
		return fmt.Errorf("pdf: the document has no root element")
	}
	return handler.EndDocument()
}

// saxRawName is a name as xml.Decoder.RawToken reports it: Space holds the
// prefix.
func saxRawName(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	return name.Space + ":" + name.Local
}

// SAXParseDocument reports a document that is already a tree to handler, the
// way SAXParse reports the text the tree was read from.
func SAXParseDocument(document *dom.Document, handler SAXContentHandler) error {
	if err := handler.StartDocument(); err != nil {
		return err
	}
	for _, child := range document.GetChildNodes() {
		if err := saxReportNode(child, handler); err != nil {
			return err
		}
	}
	return handler.EndDocument()
}

func saxReportNode(n dom.Node, handler SAXContentHandler) error {
	switch node := n.(type) {
	case *dom.Element:
		var declared []string
		var attributes SAXAttributes
		for _, attr := range node.GetAttributes() {
			if attr.NamespaceURI == saxXmlnsNamespace {
				prefix := attr.LocalName
				if attr.Name == "xmlns" {
					prefix = ""
				}
				declared = append(declared, prefix)
				if err := handler.StartPrefixMapping(prefix, attr.Value); err != nil {
					return err
				}
				continue
			}
			attributes = append(attributes, SAXAttribute{URI: attr.NamespaceURI, LocalName: attr.LocalName, QName: attr.Name, Value: attr.Value})
		}
		if err := handler.StartElement(node.GetNamespaceURI(), node.GetLocalName(), node.GetTagName(), attributes); err != nil {
			return err
		}
		for _, child := range node.GetChildNodes() {
			if err := saxReportNode(child, handler); err != nil {
				return err
			}
		}
		if err := handler.EndElement(node.GetNamespaceURI(), node.GetLocalName(), node.GetTagName()); err != nil {
			return err
		}
		for _, prefix := range declared {
			if err := handler.EndPrefixMapping(prefix); err != nil {
				return err
			}
		}
	case *dom.Text:
		return handler.Characters(node.GetData())
	case *dom.ProcessingInstruction:
		return handler.ProcessingInstruction(node.GetTarget(), node.GetData())
	}
	return nil
}

// saxDOMBuilder builds a document from SAX events: the TransformerHandler
// with a DOMResult of the Java class. A prefix mapping becomes an xmlns
// attribute of the element that starts next.
type saxDOMBuilder struct {
	document *dom.Document
	current  dom.Node
	pending  []SAXAttribute
}

var _ SAXContentHandler = (*saxDOMBuilder)(nil)

func newSAXDOMBuilder(document *dom.Document) *saxDOMBuilder {
	return &saxDOMBuilder{document: document, current: document}
}

func (b *saxDOMBuilder) SetDocumentLocator(locator SAXLocator) {}

func (b *saxDOMBuilder) StartDocument() error { return nil }

func (b *saxDOMBuilder) EndDocument() error { return nil }

func (b *saxDOMBuilder) StartPrefixMapping(prefix string, uri string) error {
	name := "xmlns"
	if prefix != "" {
		name = "xmlns:" + prefix
	}
	b.pending = append(b.pending, SAXAttribute{URI: saxXmlnsNamespace, QName: name, Value: uri})
	return nil
}

func (b *saxDOMBuilder) EndPrefixMapping(prefix string) error { return nil }

func (b *saxDOMBuilder) StartElement(uri string, localName string, qName string, attributes SAXAttributes) error {
	element := b.document.CreateElementNS(uri, qName)
	for _, attr := range b.pending {
		element.SetAttributeNS(attr.URI, attr.QName, attr.Value)
	}
	b.pending = nil
	for _, attr := range attributes {
		element.SetAttributeNS(attr.URI, attr.QName, attr.Value)
	}
	b.current.AppendChild(element)
	b.current = element
	return nil
}

func (b *saxDOMBuilder) EndElement(uri string, localName string, qName string) error {
	if b.current == dom.Node(b.document) {
		return fmt.Errorf("pdf: end of element %s without a start", qName)
	}
	b.current = b.current.GetParentNode()
	return nil
}

func (b *saxDOMBuilder) Characters(ch string) error {
	if b.current == dom.Node(b.document) {
		return nil
	}
	if last, ok := b.current.GetLastChild().(*dom.Text); ok && last.GetNodeType() == dom.TextNode {
		last.SetData(last.GetData() + ch)
		return nil
	}
	b.current.AppendChild(b.document.CreateTextNode(ch))
	return nil
}

func (b *saxDOMBuilder) IgnorableWhitespace(ch string) error {
	return b.Characters(ch)
}

func (b *saxDOMBuilder) ProcessingInstruction(target string, data string) error {
	b.current.AppendChild(b.document.CreateProcessingInstruction(target, data))
	return nil
}

func (b *saxDOMBuilder) SkippedEntity(name string) error { return nil }
