package dom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// XHTMLNamespace is the namespace of XHTML elements.
const XHTMLNamespace = "http://www.w3.org/1999/xhtml"

// ParseXML reads a well-formed XML document, which is what Flying Saucer
// requires of its input. Namespaces are resolved: an element's
// GetNamespaceURI is its namespace and GetLocalName its name without prefix,
// while GetTagName keeps the prefix as written. The named HTML entities
// (&nbsp;, &copy;) are accepted, as they are when Flying Saucer's entity
// resolver supplies the XHTML DTDs.
func ParseXML(r io.Reader) (*Document, error) {
	// The source is read whole so that a token's raw text can be inspected:
	// encoding/xml reports a CDATA section as ordinary character data, and a
	// Java parser keeps it as a node of its own, which the box builder then
	// lays out as a separate inline box.
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("dom: %w", err)
	}
	decoder := xml.NewDecoder(bytes.NewReader(source))
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = true

	doc := NewDocument()
	var current Node = doc
	// prefixes maps a namespace URI back to the prefix it was declared with,
	// one frame per open element, since encoding/xml reports the URI alone.
	prefixes := []map[string]string{{}}

	for {
		tokenStart := decoder.InputOffset()
		token, err := decoder.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("dom: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			frame := map[string]string{}
			for prefix, uri := range prefixes[len(prefixes)-1] {
				frame[prefix] = uri
			}
			for _, attr := range t.Attr {
				switch {
				case attr.Name.Space == "" && attr.Name.Local == "xmlns":
					frame[""] = attr.Value
				case attr.Name.Space == "xmlns":
					frame[attr.Name.Local] = attr.Value
				}
			}
			prefixes = append(prefixes, frame)

			if current == Node(doc) && doc.GetDocumentElement() != nil {
				return nil, fmt.Errorf("dom: a second root element <%s> after <%s>", rawName(t.Name), doc.GetDocumentElement().GetTagName())
			}
			element := doc.CreateElementNS(frame[t.Name.Space], rawName(t.Name))
			for _, attr := range t.Attr {
				namespaceURI := ""
				switch {
				case attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns"):
					namespaceURI = "http://www.w3.org/2000/xmlns/"
				case attr.Name.Space == "xml":
					namespaceURI = "http://www.w3.org/XML/1998/namespace"
				case attr.Name.Space != "":
					// An unprefixed attribute is in no namespace, whatever
					// the default namespace is.
					namespaceURI = frame[attr.Name.Space]
				}
				element.SetAttributeNS(namespaceURI, rawName(attr.Name), attr.Value)
			}
			current.AppendChild(element)
			current = element
		case xml.EndElement:
			if current.GetNodeType() != ElementNode || current.GetNodeName() != rawName(t.Name) {
				return nil, fmt.Errorf("dom: unexpected end tag </%s>", rawName(t.Name))
			}
			prefixes = prefixes[:len(prefixes)-1]
			current = current.GetParentNode()
		case xml.CharData:
			if current == Node(doc) {
				if len(bytes.TrimSpace(t)) > 0 {
					return nil, fmt.Errorf("dom: text outside the root element")
				}
				// Whitespace around the root element.
				continue
			}
			data := string(t)
			if bytes.HasPrefix(source[tokenStart:], []byte("<![CDATA[")) {
				current.AppendChild(doc.CreateCDATASection(data))
				continue
			}
			if last, ok := current.GetLastChild().(*Text); ok && !last.cdata {
				last.data += data
			} else {
				current.AppendChild(doc.CreateTextNode(data))
			}
		case xml.Comment:
			current.AppendChild(doc.CreateComment(string(t)))
		case xml.ProcInst:
			if t.Target == "xml" {
				continue
			}
			current.AppendChild(doc.CreateProcessingInstruction(t.Target, string(t.Inst)))
		case xml.Directive:
			// The document type declaration carries nothing the renderer reads.
		}
	}
	if current != Node(doc) {
		return nil, fmt.Errorf("dom: unexpected end of document inside <%s>", current.GetNodeName())
	}
	if doc.GetDocumentElement() == nil {
		return nil, fmt.Errorf("dom: the document has no root element")
	}
	return doc, nil
}

// rawName is a name as RawToken reports it: Space holds the prefix.
func rawName(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	return name.Space + ":" + name.Local
}

// ParseXMLString is ParseXML over a string.
func ParseXMLString(source string) (*Document, error) {
	return ParseXML(strings.NewReader(source))
}

// ParseHTML reads an HTML document with the HTML5 parsing algorithm, which
// accepts markup that is not well-formed XML (unclosed <br>, unquoted
// attributes) and builds the tree a browser would. Elements are in the XHTML
// namespace, so the XHTML namespace handler treats the result like a parsed
// XHTML document. Flying Saucer users get the same by parsing with jsoup and
// converting with W3CDom.
func ParseHTML(r io.Reader) (*Document, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("dom: %w", err)
	}
	doc := NewDocument()
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		convertHTMLNode(doc, doc, child)
	}
	if doc.GetDocumentElement() == nil {
		return nil, fmt.Errorf("dom: the document has no root element")
	}
	return doc, nil
}

// ParseHTMLString is ParseHTML over a string.
func ParseHTMLString(source string) (*Document, error) {
	return ParseHTML(strings.NewReader(source))
}

func convertHTMLNode(doc *Document, parent Node, n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		namespaceURI := XHTMLNamespace
		switch n.Namespace {
		case "svg":
			namespaceURI = "http://www.w3.org/2000/svg"
		case "math":
			namespaceURI = "http://www.w3.org/1998/Math/MathML"
		}
		element := doc.CreateElementNS(namespaceURI, n.Data)
		for _, attr := range n.Attr {
			name := attr.Key
			if attr.Namespace != "" {
				name = attr.Namespace + ":" + attr.Key
			}
			element.SetAttribute(name, attr.Val)
		}
		parent.AppendChild(element)
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			convertHTMLNode(doc, element, child)
		}
	case html.TextNode:
		if parent == Node(doc) {
			return
		}
		parent.AppendChild(doc.CreateTextNode(n.Data))
	case html.CommentNode:
		parent.AppendChild(doc.CreateComment(n.Data))
	}
}
