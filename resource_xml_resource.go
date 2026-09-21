// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/resource/XMLResource.java

package ufo

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/octoberswimmer/ufo/dom"
)

// XMLResource is a parsed XML document.
//
// Java parses with a pooled SAX XMLReader feeding a pooled TrAX identity
// transformer, with FSEntityResolver supplying the XHTML DTDs so that the
// named entities resolve. The port builds the tree with dom.ParseXML, which
// accepts the HTML named entities and reads no external entity, so the
// parser-specific parts of the Java class have no counterpart here:
// newXMLReader, XMLReaderPool (and the xr.load.* parser feature settings it
// reads), WhitespacePreservingFilter, IdentityTransformerPool, ObjectPool, and
// load(javax.xml.transform.Source).
type XMLResource struct {
	AbstractResource
	document        *dom.Document
	elapsedLoadTime int64
}

var xmlResourceXMLResourceBuilder = &xmlResourceBuilder{}

func newXMLResource(source *InputSource, document *dom.Document, elapsedLoadTime int64) *XMLResource {
	r := &XMLResource{document: document, elapsedLoadTime: elapsedLoadTime}
	r.initAbstractResource(source)
	return r
}

// XMLResourceLoad ports XMLResource.load(Reader).
func XMLResourceLoad(reader io.Reader) *XMLResource {
	return xmlResourceXMLResourceBuilder.createXMLResource(NewInputSourceReader(reader))
}

// XMLResourceLoadURL ports XMLResource.load(URL).
func XMLResourceLoadURL(source *url.URL) *XMLResource {
	return XMLResourceLoadInputSource(InputSourcesFromURL(source))
}

// XMLResourceLoadInputStream ports XMLResource.load(InputStream).
func XMLResourceLoadInputStream(stream io.Reader) *XMLResource {
	return xmlResourceXMLResourceBuilder.createXMLResource(InputSourcesFromStream(stream))
}

// XMLResourceLoadInputSource ports XMLResource.load(InputSource).
func XMLResourceLoadInputSource(source *InputSource) *XMLResource {
	return xmlResourceXMLResourceBuilder.createXMLResource(source)
}

// XMLResourceLoadString ports XMLResource.load(String).
func XMLResourceLoadString(xml string) *XMLResource {
	return XMLResourceLoad(strings.NewReader(xml))
}

func (r *XMLResource) GetDocument() *dom.Document {
	return r.document
}

// GetElapsedLoadTime returns the time parsing took, in milliseconds.
func (r *XMLResource) GetElapsedLoadTime() int64 {
	return r.elapsedLoadTime
}

// xmlResourceBuilder ports the nested class XMLResource.XMLResourceBuilder.
type xmlResourceBuilder struct {
}

func (b *xmlResourceBuilder) createXMLResource(inputSource *InputSource) *XMLResource {
	start := time.Now().UnixMilli()
	document := b.parse(inputSource)
	elapsedLoadTime := time.Now().UnixMilli() - start

	XRLogLoad(fmt.Sprintf("Loaded document in %dms", elapsedLoadTime))

	return newXMLResource(inputSource, document, elapsedLoadTime)
}

// parse reads the document from the input source: its character stream when
// it has one, else its byte stream, else the content its system id names,
// which is the order a SAX parser uses. A failure panics with the
// XRRuntimeException the Java transform step throws.
func (b *xmlResourceBuilder) parse(inputSource *InputSource) *dom.Document {
	reader, closer, err := xmlResourceOpen(inputSource)
	if err == nil {
		if closer != nil {
			defer closer.Close()
		}
		var document *dom.Document
		document, err = dom.ParseXML(reader)
		if err == nil {
			return document
		}
	}
	panic(NewXRRuntimeExceptionWithCause("Can't load the XML resource (using TrAX transformer). "+err.Error(), err))
}

func xmlResourceOpen(inputSource *InputSource) (io.Reader, io.Closer, error) {
	if inputSource == nil {
		return nil, nil, fmt.Errorf("the input source is null")
	}
	if characterStream := inputSource.GetCharacterStream(); characterStream != nil {
		return characterStream, nil, nil
	}
	if byteStream := inputSource.GetByteStream(); byteStream != nil {
		return byteStream, nil, nil
	}
	systemId := inputSource.GetSystemId()
	if systemId == "" {
		return nil, nil, fmt.Errorf("the input source has no stream and no system id")
	}
	parsed, err := url.Parse(systemId)
	if err != nil {
		return nil, nil, err
	}
	switch parsed.Scheme {
	case "", "file":
		path := parsed.Path
		if parsed.Scheme == "" {
			path = systemId
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		return file, file, nil
	case "http", "https":
		response, err := http.Get(systemId)
		if err != nil {
			return nil, nil, err
		}
		if response.StatusCode < 200 || response.StatusCode > 299 {
			response.Body.Close()
			return nil, nil, fmt.Errorf("%s: %s", systemId, response.Status)
		}
		return response.Body, response.Body, nil
	}
	return nil, nil, fmt.Errorf("unknown protocol: %s", parsed.Scheme)
}
