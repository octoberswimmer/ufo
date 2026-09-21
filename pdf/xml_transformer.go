// No Flying Saucer class is ported here. This file stands for the identity
// javax.xml.transform.Transformer with OMIT_XML_DECLARATION that
// ITextRenderer.stringifyMetadata (and the DocumentSplitter test) use to
// write a DOM node as XML text.

package pdf

import (
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// xmlTransformerSerialize writes node and its descendants as XML without an
// XML declaration. Text is escaped as the JDK serializer escapes it (&amp;,
// &lt;, &gt;), an element without children is written as <name/>, and a
// namespace that the subtree uses but does not declare is declared on the
// element that first uses it.
func xmlTransformerSerialize(node dom.Node) string {
	var sb strings.Builder
	xmlTransformerWrite(&sb, node, map[string]string{})
	return sb.String()
}

func xmlTransformerWrite(sb *strings.Builder, n dom.Node, scope map[string]string) {
	switch node := n.(type) {
	case *dom.Document:
		for _, child := range node.GetChildNodes() {
			xmlTransformerWrite(sb, child, scope)
		}
	case *dom.Element:
		inner := map[string]string{}
		for prefix, uri := range scope {
			inner[prefix] = uri
		}
		for _, attr := range node.GetAttributes() {
			if attr.NamespaceURI == saxXmlnsNamespace {
				if attr.Name == "xmlns" {
					inner[""] = attr.Value
				} else {
					inner[attr.LocalName] = attr.Value
				}
			}
		}
		sb.WriteString("<")
		sb.WriteString(node.GetTagName())
		if inner[node.GetPrefix()] != node.GetNamespaceURI() {
			inner[node.GetPrefix()] = node.GetNamespaceURI()
			xmlTransformerWriteNamespace(sb, node.GetPrefix(), node.GetNamespaceURI())
		}
		for _, attr := range node.GetAttributes() {
			if attr.NamespaceURI != "" && attr.NamespaceURI != saxXmlnsNamespace {
				prefix, _, found := strings.Cut(attr.Name, ":")
				if found && prefix != "xml" && inner[prefix] != attr.NamespaceURI {
					inner[prefix] = attr.NamespaceURI
					xmlTransformerWriteNamespace(sb, prefix, attr.NamespaceURI)
				}
			}
			sb.WriteString(" ")
			sb.WriteString(attr.Name)
			sb.WriteString("=\"")
			sb.WriteString(xmlTransformerEscape(attr.Value, true))
			sb.WriteString("\"")
		}
		if !node.HasChildNodes() {
			sb.WriteString("/>")
			return
		}
		sb.WriteString(">")
		for _, child := range node.GetChildNodes() {
			xmlTransformerWrite(sb, child, inner)
		}
		sb.WriteString("</")
		sb.WriteString(node.GetTagName())
		sb.WriteString(">")
	case *dom.Text:
		if node.GetNodeType() == dom.CDATASectionNode {
			sb.WriteString("<![CDATA[")
			sb.WriteString(node.GetData())
			sb.WriteString("]]>")
		} else {
			sb.WriteString(xmlTransformerEscape(node.GetData(), false))
		}
	case *dom.Comment:
		sb.WriteString("<!--")
		sb.WriteString(node.GetData())
		sb.WriteString("-->")
	case *dom.ProcessingInstruction:
		sb.WriteString("<?")
		sb.WriteString(node.GetTarget())
		if node.GetData() != "" {
			sb.WriteString(" ")
			sb.WriteString(node.GetData())
		}
		sb.WriteString("?>")
	}
}

func xmlTransformerWriteNamespace(sb *strings.Builder, prefix string, uri string) {
	if prefix == "" {
		sb.WriteString(" xmlns=\"")
	} else {
		sb.WriteString(" xmlns:")
		sb.WriteString(prefix)
		sb.WriteString("=\"")
	}
	sb.WriteString(xmlTransformerEscape(uri, true))
	sb.WriteString("\"")
}

func xmlTransformerEscape(s string, attribute bool) string {
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r == '&':
			sb.WriteString("&amp;")
		case r == '<':
			sb.WriteString("&lt;")
		case r == '>' && !attribute:
			sb.WriteString("&gt;")
		case r == '"' && attribute:
			sb.WriteString("&quot;")
		case r == '\n' && attribute:
			sb.WriteString("&#10;")
		case r == '\r':
			sb.WriteString("&#13;")
		case r == '\t' && attribute:
			sb.WriteString("&#9;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
