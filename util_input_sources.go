// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/InputSources.java

package ufo

import (
	"bufio"
	"io"
	"net/url"
	"strings"
)

// InputSource stands for the JDK class org.xml.sax.InputSource: the input of
// an XML parser, given as a byte stream, as a character stream, or as a
// system id (a URL) to open. A nil reader and an empty string stand for the
// Java null.
type InputSource struct {
	publicId        string
	systemId        string
	byteStream      io.Reader
	characterStream io.Reader
	encoding        string
}

// NewInputSource ports InputSource().
func NewInputSource() *InputSource {
	return &InputSource{}
}

// NewInputSourceString ports InputSource(String systemId).
func NewInputSourceString(systemId string) *InputSource {
	return &InputSource{systemId: systemId}
}

// NewInputSourceInputStream ports InputSource(InputStream byteStream).
func NewInputSourceInputStream(byteStream io.Reader) *InputSource {
	return &InputSource{byteStream: byteStream}
}

// NewInputSourceReader ports InputSource(Reader characterStream). The reader
// yields the document as UTF-8 text.
func NewInputSourceReader(characterStream io.Reader) *InputSource {
	return &InputSource{characterStream: characterStream}
}

func (s *InputSource) SetPublicId(publicId string) { s.publicId = publicId }

func (s *InputSource) GetPublicId() string { return s.publicId }

func (s *InputSource) SetSystemId(systemId string) { s.systemId = systemId }

func (s *InputSource) GetSystemId() string { return s.systemId }

func (s *InputSource) SetByteStream(byteStream io.Reader) { s.byteStream = byteStream }

func (s *InputSource) GetByteStream() io.Reader { return s.byteStream }

func (s *InputSource) SetEncoding(encoding string) { s.encoding = encoding }

func (s *InputSource) GetEncoding() string { return s.encoding }

func (s *InputSource) SetCharacterStream(characterStream io.Reader) {
	s.characterStream = characterStream
}

func (s *InputSource) GetCharacterStream() io.Reader { return s.characterStream }

func InputSourcesFromStream(is io.Reader) *InputSource {
	if is == nil {
		return nil
	}
	return NewInputSourceInputStream(bufio.NewReader(is))
}

func InputSourcesFromURL(source *url.URL) *InputSource {
	return NewInputSourceString(source.String())
}

func InputSourcesFromString(source string) *InputSource {
	return NewInputSourceReader(strings.NewReader(source))
}
