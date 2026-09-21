// Package pipe renders documents for another program over a pair of streams,
// so that program can use ufo without linking it: it starts `ufo pipe`,
// writes a render request to its standard input, answers the requests for
// resources the document refers to, and reads the PDF from its standard
// output. aer renders Visualforce pages with renderAs="pdf" this way.
//
// # Protocol
//
// Messages are JSON objects, one per line (UTF-8, terminated by "\n"), in both
// directions. Byte strings are base64 encoded. Protocol version 1:
//
// The client sends one request:
//
//	{"type":"render","protocol":1,"html":"<html>…","baseURL":"https://host/apex/Page"}
//
// html is parsed with the HTML5 parsing algorithm, or as XML when "xml":true.
// Relative URLs in the document resolve against baseURL. An optional
// "fontFamilies" object names CSS font families to render with one of the PDF
// base-14 fonts:
//
//	"fontFamilies":{"Arial Unicode MS":"Helvetica"}
//
// Each such family has that single face, which also serves bold text as it
// is, without the simulated bold ufo otherwise draws for a family with no bold
// face; this is how Salesforce's PDF renderer draws bold text in its
// single-face Arial Unicode MS.
//
// While rendering, ufo asks for each resource the document refers to
// (stylesheets, images, fonts), identified by its resolved URL:
//
//	{"type":"fetch","id":1,"url":"https://host/resource/Styles"}
//
// and the client answers each one before ufo continues:
//
//	{"type":"resource","id":1,"data":"<base64>"}
//	{"type":"resource","id":1,"error":"not found"}
//
// A resource that is not available is left out, as a browser leaves out one
// it cannot load. data: URIs are decoded by ufo and never fetched.
//
// ufo finishes with exactly one of:
//
//	{"type":"result","pdf":"<base64>"}
//	{"type":"error","message":"…"}
//
// and exits. A request with a protocol version ufo does not speak is answered
// with an error message.
package pipe

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// ProtocolVersion is the protocol this package speaks.
const ProtocolVersion = 1

// Message is one line of the protocol, in either direction.
type Message struct {
	Type     string `json:"type"`
	Protocol int    `json:"protocol,omitempty"`
	HTML     string `json:"html,omitempty"`
	XML      bool   `json:"xml,omitempty"`
	BaseURL  string `json:"baseURL,omitempty"`
	// FontFamilies maps CSS font family names to base-14 font names.
	FontFamilies map[string]string `json:"fontFamilies,omitempty"`
	ID           int               `json:"id,omitempty"`
	URL          string            `json:"url,omitempty"`
	Data         []byte            `json:"data,omitempty"`
	Error        string            `json:"error,omitempty"`
	PDF          []byte            `json:"pdf,omitempty"`
	Message      string            `json:"message,omitempty"`
}

// Serve reads one render request from in, renders it, and writes the result
// (or the error that stopped it) to out. It returns an error only when the
// streams themselves fail.
func Serve(in io.Reader, out io.Writer) error {
	conn := &conn{in: bufio.NewReader(in), out: json.NewEncoder(out)}
	var request Message
	if err := conn.read(&request); err != nil {
		return err
	}
	var content []byte
	var renderErr error
	switch {
	case request.Type != "render":
		renderErr = fmt.Errorf("expected a render request, got %q", request.Type)
	case request.Protocol != ProtocolVersion:
		renderErr = fmt.Errorf("protocol version %d is not supported; this ufo speaks version %d", request.Protocol, ProtocolVersion)
	default:
		content, renderErr = conn.render(request)
	}
	if conn.streamErr != nil {
		return conn.streamErr
	}
	if renderErr != nil {
		return conn.out.Encode(Message{Type: "error", Message: renderErr.Error()})
	}
	return conn.out.Encode(Message{Type: "result", PDF: content})
}

type conn struct {
	in        *bufio.Reader
	out       *json.Encoder
	nextID    int
	streamErr error
}

func (c *conn) read(m *Message) error {
	line, err := c.in.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		return err
	}
	return json.Unmarshal(line, m)
}

// fetch asks the client for a resource and waits for the answer.
func (c *conn) fetch(url string) ([]byte, error) {
	if c.streamErr != nil {
		return nil, c.streamErr
	}
	c.nextID++
	id := c.nextID
	if err := c.out.Encode(Message{Type: "fetch", ID: id, URL: url}); err != nil {
		c.streamErr = err
		return nil, err
	}
	var answer Message
	if err := c.read(&answer); err != nil {
		c.streamErr = err
		return nil, err
	}
	if answer.Type != "resource" || answer.ID != id {
		c.streamErr = fmt.Errorf("expected the resource answer for request %d, got %q for %d", id, answer.Type, answer.ID)
		return nil, c.streamErr
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("%s", answer.Error)
	}
	return answer.Data, nil
}

func (c *conn) render(request Message) (content []byte, err error) {
	defer func() {
		// The renderer's exported entry points return errors; this guards
		// against a panic outside them, which would otherwise leave the
		// client without an answer.
		if r := recover(); r != nil {
			err = fmt.Errorf("ufo failed: %v\n%s", r, debug.Stack())
		}
	}()
	outputDevice := pdf.NewITextOutputDevice(pdf.ITextRendererDefaultDotsPerPoint)
	userAgent := newUserAgent(outputDevice, c.fetch)
	renderer := pdf.NewITextRendererWithDotsPerPointDotsPerPixelOutputDeviceUserAgent(
		pdf.ITextRendererDefaultDotsPerPoint, pdf.ITextRendererDefaultDotsPerPixel, outputDevice, userAgent.ITextUserAgent)

	if err := addFontFamilies(renderer, request.FontFamilies); err != nil {
		return nil, err
	}

	if request.XML {
		doc, err := dom.ParseXMLString(request.HTML)
		if err != nil {
			return nil, fmt.Errorf("the document is not well-formed XML: %w", err)
		}
		if err := renderer.SetDocumentWithUrl(doc, request.BaseURL); err != nil {
			return nil, err
		}
	} else if err := renderer.SetDocumentFromHTMLString(request.HTML, request.BaseURL); err != nil {
		return nil, fmt.Errorf("the document could not be read: %w", err)
	}
	if err := renderer.Layout(); err != nil {
		return nil, fmt.Errorf("the document could not be laid out: %w", err)
	}
	var out bytes.Buffer
	if err := renderer.CreatePDF(&out); err != nil {
		return nil, fmt.Errorf("the PDF could not be written: %w", err)
	}
	return out.Bytes(), nil
}

// addFontFamilies registers each named family as the single face of a base-14
// font.
func addFontFamilies(renderer *pdf.ITextRenderer, families map[string]string) error {
	if len(families) == 0 {
		return nil
	}
	fonts := renderer.GetFontResolver().GetFonts()
	for name, fontName := range families {
		// Only the base-14 fonts: any other name is a font file path, which
		// a client must not be able to make ufo read.
		if !writer.IsBase14Font(fontName) {
			return fmt.Errorf("font family %q: %s is not one of the PDF base-14 fonts", name, fontName)
		}
		font, err := writer.CreateFont(fontName, writer.BaseFontCp1252, false)
		if err != nil {
			return fmt.Errorf("font family %q: %w", name, err)
		}
		family := pdf.NewFontFamily(name)
		family.AddFontDescription(pdf.NewFontDescriptionWithStyleWeight(font, ufo.IdentValueNormal, 400))
		family.AddFontDescription(pdf.NewFontDescriptionWithStyleWeight(font, ufo.IdentValueNormal, 700))
		fonts[name] = family
	}
	return nil
}

// userAgent is Flying Saucer's ITextUserAgent with its stream opening
// replaced, the way a Java caller subclasses it: every resource is asked of
// the client.
type userAgent struct {
	*pdf.ITextUserAgent
	fetch func(url string) ([]byte, error)
}

func newUserAgent(outputDevice *pdf.ITextOutputDevice, fetch func(string) ([]byte, error)) *userAgent {
	a := &userAgent{
		ITextUserAgent: pdf.NewITextUserAgent(outputDevice, pdf.ITextRendererDefaultDotsPerPixel),
		fetch:          fetch,
	}
	a.SetSelf(a)
	return a
}

// OpenStream asks the client for the resource.
func (a *userAgent) OpenStream(uri string) (io.ReadCloser, error) {
	content, err := a.fetch(uri)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}
