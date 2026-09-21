// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/DocumentSplitter.java

package pdf

import (
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

const documentSplitterHeadElementName = "head"

// DocumentSplitter is a SAX content handler that splits a document into one
// document per child of the root element other than the head; the events of
// the head are recorded and replayed into each of them. Feed it with
// SAXParse (from a stream, so the large source document is never a tree) or
// SAXParseDocument (from a tree).
//
// SAXContentHandler (sax.go) stands for org.xml.sax.ContentHandler, and the
// handler that builds each document stands for Java's TransformerHandler
// with a DOMResult.
type DocumentSplitter struct {
	processingInstructions []documentSplitterProcessingInstruction
	head                   *SAXEventRecorder
	inHead                 bool

	depth int

	needNewNSScope bool
	currentNSScope *documentSplitterNamespaceScope

	needNSScopePop bool

	locator SAXLocator

	handler    SAXContentHandler
	inDocument bool

	documents []*dom.Document

	replayedHead bool
}

var _ SAXContentHandler = (*DocumentSplitter)(nil)

func NewDocumentSplitter() *DocumentSplitter {
	return &DocumentSplitter{
		head:           NewSAXEventRecorder(),
		currentNSScope: newDocumentSplitterNamespaceScope(),
	}
}

func (s *DocumentSplitter) Characters(ch string) error {
	if s.inHead {
		return s.head.Characters(ch)
	} else if s.inDocument {
		return s.handler.Characters(ch)
	}
	return nil
}

func (s *DocumentSplitter) EndDocument() error {
	return nil
}

func (s *DocumentSplitter) EndPrefixMapping(prefix string) error {
	if s.inHead {
		return s.head.EndPrefixMapping(prefix)
	} else if s.inDocument {
		return s.handler.EndPrefixMapping(prefix)
	} else {
		s.needNSScopePop = true
	}
	return nil
}

func (s *DocumentSplitter) IgnorableWhitespace(ch string) error {
	if s.inHead {
		return s.head.IgnorableWhitespace(ch)
	} else if s.inDocument {
		return s.handler.IgnorableWhitespace(ch)
	}
	return nil
}

func (s *DocumentSplitter) ProcessingInstruction(target string, data string) error {
	s.processingInstructions = append(s.processingInstructions, documentSplitterProcessingInstruction{target: target, data: data})
	return nil
}

func (s *DocumentSplitter) SetDocumentLocator(locator SAXLocator) {
	s.locator = locator
}

func (s *DocumentSplitter) SkippedEntity(name string) error {
	if s.inHead {
		return s.head.SkippedEntity(name)
	} else if s.inDocument {
		return s.handler.SkippedEntity(name)
	}
	return nil
}

func (s *DocumentSplitter) StartDocument() error {
	return nil
}

func (s *DocumentSplitter) StartElement(uri string, localName string, qName string, attributes SAXAttributes) error {
	if s.inHead {
		if err := s.head.StartElement(uri, localName, qName, attributes); err != nil {
			return err
		}
	} else if s.inDocument {
		if s.depth == 2 && !s.replayedHead {
			if strings.EqualFold(documentSplitterHeadElementName, qName) {
				if err := s.handler.StartElement(uri, localName, qName, attributes); err != nil {
					return err
				}
				if err := s.head.Replay(s.handler); err != nil {
					return err
				}
			} else {
				if err := s.handler.StartElement("", documentSplitterHeadElementName, documentSplitterHeadElementName, SAXAttributes{}); err != nil {
					return err
				}
				if err := s.head.Replay(s.handler); err != nil {
					return err
				}
				if err := s.handler.EndElement("", documentSplitterHeadElementName, documentSplitterHeadElementName); err != nil {
					return err
				}

				if err := s.handler.StartElement(uri, localName, qName, attributes); err != nil {
					return err
				}
			}

			s.replayedHead = true
		} else {
			if err := s.handler.StartElement(uri, localName, qName, attributes); err != nil {
				return err
			}
		}
	} else {
		if s.needNewNSScope {
			s.needNewNSScope = false
			s.currentNSScope = newDocumentSplitterNamespaceScopeWithParent(s.currentNSScope)
		}

		if s.depth == 1 {
			if strings.EqualFold(documentSplitterHeadElementName, qName) {
				s.inHead = true
				if err := s.currentNSScope.replay(s.head, true); err != nil {
					return err
				}
			} else {
				s.inDocument = true
				s.replayedHead = false

				doc := dom.NewDocument()
				s.documents = append(s.documents, doc)
				s.handler = newSAXDOMBuilder(doc)

				if err := s.handler.StartDocument(); err != nil {
					return err
				}
				s.handler.SetDocumentLocator(s.locator)
				for _, pI := range s.processingInstructions {
					if err := s.handler.ProcessingInstruction(pI.target, pI.data); err != nil {
						return err
					}
				}

				if err := s.currentNSScope.replay(s.handler, true); err != nil {
					return err
				}
				if err := s.handler.StartElement(uri, localName, qName, attributes); err != nil {
					return err
				}
			}
		}
	}

	s.depth++
	return nil
}

func (s *DocumentSplitter) EndElement(uri string, localName string, qName string) error {
	s.depth--

	if s.needNSScopePop {
		s.needNSScopePop = false
		s.currentNSScope = s.currentNSScope.getParent()
	}

	if s.inHead {
		if s.depth == 1 {
			if err := s.currentNSScope.replay(s.head, false); err != nil {
				return err
			}
			s.inHead = false
		} else {
			return s.head.EndElement(uri, localName, qName)
		}
	} else if s.inDocument {
		if s.depth == 1 {
			if err := s.currentNSScope.replay(s.handler, false); err != nil {
				return err
			}
			if err := s.handler.EndElement(uri, localName, qName); err != nil {
				return err
			}
			if err := s.handler.EndDocument(); err != nil {
				return err
			}
			s.inDocument = false
		} else {
			return s.handler.EndElement(uri, localName, qName)
		}
	}
	return nil
}

func (s *DocumentSplitter) StartPrefixMapping(prefix string, uri string) error {
	if s.inHead {
		return s.head.StartPrefixMapping(prefix, uri)
	} else if s.inDocument {
		return s.handler.StartPrefixMapping(prefix, uri)
	} else {
		s.needNewNSScope = true
		s.currentNSScope.addNamespace(&documentSplitterNamespace{prefix: prefix, uri: uri})
	}
	return nil
}

func (s *DocumentSplitter) GetDocuments() []*dom.Document {
	return s.documents
}

// documentSplitterNamespace ports the private class DocumentSplitter.Namespace.
type documentSplitterNamespace struct {
	prefix string
	uri    string
}

func (n *documentSplitterNamespace) getPrefix() string {
	return n.prefix
}

func (n *documentSplitterNamespace) getUri() string {
	return n.uri
}

// documentSplitterNamespaceScope ports the private class
// DocumentSplitter.NamespaceScope.
type documentSplitterNamespaceScope struct {
	// parent is nil for the outermost scope.
	parent     *documentSplitterNamespaceScope
	namespaces []*documentSplitterNamespace
}

func newDocumentSplitterNamespaceScope() *documentSplitterNamespaceScope {
	return &documentSplitterNamespaceScope{}
}

func newDocumentSplitterNamespaceScopeWithParent(parent *documentSplitterNamespaceScope) *documentSplitterNamespaceScope {
	return &documentSplitterNamespaceScope{parent: parent}
}

func (n *documentSplitterNamespaceScope) addNamespace(namespace *documentSplitterNamespace) {
	n.namespaces = append(n.namespaces, namespace)
}

func (n *documentSplitterNamespaceScope) replay(contentHandler SAXContentHandler, start bool) error {
	return n.replayWithSeen(contentHandler, map[string]struct{}{}, start)
}

func (n *documentSplitterNamespaceScope) replayWithSeen(contentHandler SAXContentHandler, seen map[string]struct{}, start bool) error {
	for _, ns := range n.namespaces {
		if _, ok := seen[ns.getPrefix()]; !ok {
			seen[ns.getPrefix()] = struct{}{}
			if start {
				if err := contentHandler.StartPrefixMapping(ns.getPrefix(), ns.getUri()); err != nil {
					return err
				}
			} else {
				if err := contentHandler.EndPrefixMapping(ns.getPrefix()); err != nil {
					return err
				}
			}
		}
	}

	if n.parent != nil {
		return n.parent.replayWithSeen(contentHandler, seen, start)
	}
	return nil
}

// getParent returns nil for the outermost scope.
func (n *documentSplitterNamespaceScope) getParent() *documentSplitterNamespaceScope {
	return n.parent
}

// documentSplitterProcessingInstruction ports the private record
// DocumentSplitter.ProcessingInstruction.
type documentSplitterProcessingInstruction struct {
	target string
	data   string
}
