// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/SAXEventRecorder.java

package pdf

// SAXEventRecorder records the SAX events it receives and replays them to
// another handler. SAXContentHandler (sax.go) stands for
// org.xml.sax.ContentHandler.
type SAXEventRecorder struct {
	events []saxEventRecorderEvent
}

var _ SAXContentHandler = (*SAXEventRecorder)(nil)

// saxEventRecorderEvent ports the private interface SAXEventRecorder.Event.
type saxEventRecorderEvent func(handler SAXContentHandler) error

func NewSAXEventRecorder() *SAXEventRecorder {
	return &SAXEventRecorder{}
}

func (r *SAXEventRecorder) Characters(ch string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.Characters(ch) })
	return nil
}

func (r *SAXEventRecorder) EndDocument() error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.EndDocument() })
	return nil
}

func (r *SAXEventRecorder) EndElement(uri string, localName string, qName string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.EndElement(uri, localName, qName) })
	return nil
}

func (r *SAXEventRecorder) EndPrefixMapping(prefix string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.EndPrefixMapping(prefix) })
	return nil
}

func (r *SAXEventRecorder) IgnorableWhitespace(ch string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.IgnorableWhitespace(ch) })
	return nil
}

func (r *SAXEventRecorder) ProcessingInstruction(target string, data string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.ProcessingInstruction(target, data) })
	return nil
}

func (r *SAXEventRecorder) SetDocumentLocator(locator SAXLocator) {
}

func (r *SAXEventRecorder) SkippedEntity(name string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.SkippedEntity(name) })
	return nil
}

func (r *SAXEventRecorder) StartDocument() error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.StartDocument() })
	return nil
}

func (r *SAXEventRecorder) StartElement(uri string, localName string, qName string, attributes SAXAttributes) error {
	r.events = append(r.events, func(handler SAXContentHandler) error {
		return handler.StartElement(uri, localName, qName, attributes)
	})
	return nil
}

func (r *SAXEventRecorder) StartPrefixMapping(prefix string, uri string) error {
	r.events = append(r.events, func(handler SAXContentHandler) error { return handler.StartPrefixMapping(prefix, uri) })
	return nil
}

func (r *SAXEventRecorder) Replay(handler SAXContentHandler) error {
	for _, e := range r.events {
		if err := e(handler); err != nil {
			return err
		}
	}
	return nil
}
