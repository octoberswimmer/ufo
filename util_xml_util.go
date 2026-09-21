// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/XMLUtil.java

package ufo

import (
	"errors"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// DocumentBuilder stands for the JDK class
// javax.xml.parsers.DocumentBuilder as XMLUtil configures it. The parser is
// dom.ParseXML.
//
// The Java class hardens the JAXP parser against XXE: external entities are
// routed through FSEntityResolver, secure processing bounds entity
// expansion, and access to external DTDs and schemas is denied. dom.ParseXML
// has none of the features these settings restrict. It reads no document
// type definition, neither the internal subset nor an external one, opens no
// URL, and expands only the predefined XML entities and the named HTML
// entities (&nbsp;), which is what FSEntityResolver's local copies of the
// XHTML DTDs give the Java parser. A reference to an entity that the document
// declares itself (<!ENTITY xxe SYSTEM "...">, or the nested entities of a
// "billion laughs" document) is a parse error.
type DocumentBuilder struct{}

// Parse ports DocumentBuilder.parse(InputStream).
func (b *DocumentBuilder) Parse(is io.Reader) (*dom.Document, error) {
	if is == nil {
		return nil, errors.New("InputStream cannot be null")
	}
	return dom.ParseXML(is)
}

// ParseString ports DocumentBuilder.parse(String uri).
func (b *DocumentBuilder) ParseString(uri string) (*dom.Document, error) {
	if uri == "" {
		return nil, errors.New("URI cannot be null")
	}
	return b.ParseInputSource(NewInputSourceString(uri))
}

// ParseInputSource ports DocumentBuilder.parse(InputSource). The character
// stream is read when there is one, else the byte stream, else the system id
// is opened as a URL, or as a file path when it has no scheme.
func (b *DocumentBuilder) ParseInputSource(is *InputSource) (*dom.Document, error) {
	if is == nil {
		return nil, errors.New("InputSource cannot be null")
	}
	if characterStream := is.GetCharacterStream(); characterStream != nil {
		return dom.ParseXML(characterStream)
	}
	if byteStream := is.GetByteStream(); byteStream != nil {
		return dom.ParseXML(byteStream)
	}
	systemId := is.GetSystemId()
	if systemId == "" {
		return nil, errors.New("the input source has no stream and no system id")
	}
	var stream io.ReadCloser
	parsed, err := url.Parse(systemId)
	if err == nil && parsed.Scheme != "" {
		stream, err = ioUtilOpenURL(systemId, 0, 0, "")
	} else {
		stream, err = os.Open(systemId)
	}
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	return dom.ParseXML(stream)
}

// NewDocument ports DocumentBuilder.newDocument().
func (b *DocumentBuilder) NewDocument() *dom.Document {
	return dom.NewDocument()
}

func XMLUtilDocumentFromString(documentContents string) (*dom.Document, error) {
	return xmlUtilCreateDocumentBuilder().ParseInputSource(NewInputSourceReader(strings.NewReader(documentContents)))
}

func XMLUtilDocumentFromFile(filename string) (*dom.Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return xmlUtilCreateDocumentBuilder().Parse(file)
}

// XMLUtilNewDocumentBuilder returns a DocumentBuilder that is safe against
// XXE; see DocumentBuilder for how the settings of the Java method
// (FSEntityResolver, FEATURE_SECURE_PROCESSING, DOCTYPE declarations still
// allowed) are met. XMLUtil.newSecureDocumentBuilderFactory, which denies
// external DTD and schema access and XInclude at the JAXP level, has no
// counterpart: the parser has no such access to deny.
func XMLUtilNewDocumentBuilder() *DocumentBuilder {
	return &DocumentBuilder{}
}

func xmlUtilCreateDocumentBuilder() *DocumentBuilder {
	return XMLUtilNewDocumentBuilder()
}
