package pipe

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/pdf/writer"
)

// client drives Serve the way another program drives `ufo pipe`: it sends the
// request, answers each fetch from resources (an absent URL is answered with
// an error), and returns the fetched URLs and the final message.
func client(t *testing.T, request Message, resources map[string]string) ([]string, Message) {
	t.Helper()
	toServer, clientWrites := io.Pipe()
	clientReads, fromServer := io.Pipe()
	done := make(chan error, 1)
	go func() {
		err := Serve(toServer, fromServer)
		fromServer.Close()
		done <- err
	}()
	encoder := json.NewEncoder(clientWrites)
	if err := encoder.Encode(request); err != nil {
		t.Fatal(err)
	}
	var fetched []string
	scanner := bufio.NewScanner(clientReads)
	scanner.Buffer(make([]byte, 1<<20), 64<<20)
	for scanner.Scan() {
		var m Message
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			t.Fatal(err)
		}
		if m.Type != "fetch" {
			clientWrites.Close()
			if err := <-done; err != nil {
				t.Fatalf("Serve: %v", err)
			}
			return fetched, m
		}
		fetched = append(fetched, m.URL)
		answer := Message{Type: "resource", ID: m.ID}
		if content, ok := resources[m.URL]; ok {
			answer.Data = []byte(content)
		} else {
			answer.Error = "not found"
		}
		if err := encoder.Encode(answer); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatalf("the stream ended without a result: %v", scanner.Err())
	return nil, Message{}
}

func TestServe_renders_with_resources_from_the_client(t *testing.T) {
	fetched, result := client(t, Message{
		Type: "render", Protocol: ProtocolVersion,
		HTML: `<link rel="stylesheet" href="/resource/Print"/><link rel="stylesheet" href="missing.css"/>` +
			`<h1>Invoice</h1><p>Total &amp; tax</p>`,
		BaseURL: "https://example.test/apex/Invoice?id=1",
	}, map[string]string{
		"https://example.test/resource/Print": `@page { size: letter landscape; } h1 { font-family: Courier; }`,
	})
	if result.Type != "result" {
		t.Fatalf("result = %+v", result)
	}
	want := []string{"https://example.test/resource/Print", "https://example.test/apex/missing.css"}
	if strings.Join(fetched, " ") != strings.Join(want, " ") {
		t.Errorf("fetched %v, want %v", fetched, want)
	}
	doc, err := writer.ReadPDF(result.PDF)
	if err != nil {
		t.Fatalf("the result is not a readable PDF: %v", err)
	}
	if w, h := doc.PageSize(0); w != 792 || h != 612 {
		t.Errorf("page size = %vx%v, want letter landscape from the fetched stylesheet", w, h)
	}
	if text, _ := doc.Text(); !strings.Contains(text, "Invoice") || !strings.Contains(text, "Total & tax") {
		t.Errorf("text = %q", text)
	}
	if fonts := strings.Join(doc.PageFonts(0), " "); !strings.Contains(fonts, "Courier") {
		t.Errorf("fonts = %s", fonts)
	}
}

func TestServe_parses_xml_when_asked(t *testing.T) {
	_, result := client(t, Message{
		Type: "render", Protocol: ProtocolVersion, XML: true,
		HTML: `<html xmlns="http://www.w3.org/1999/xhtml"><body><p>XML</p></body></html>`,
	}, nil)
	if result.Type != "result" {
		t.Fatalf("result = %+v", result)
	}
	_, result = client(t, Message{Type: "render", Protocol: ProtocolVersion, XML: true, HTML: `<p>not closed`}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "not well-formed XML") {
		t.Errorf("result = %+v, want the XML error", result)
	}
}

func TestServe_rejects_other_protocols_and_requests(t *testing.T) {
	_, result := client(t, Message{Type: "render", Protocol: 99, HTML: "<p>x</p>"}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "protocol version 99 is not supported") {
		t.Errorf("result = %+v", result)
	}
	_, result = client(t, Message{Type: "fetch", Protocol: ProtocolVersion}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "expected a render request") {
		t.Errorf("result = %+v", result)
	}
}

// A client that goes away mid-render ends Serve with the stream's error
// rather than leaving it waiting.
func TestServe_returns_when_the_client_closes_the_stream(t *testing.T) {
	toServer, clientWrites := io.Pipe()
	clientReads, fromServer := io.Pipe()
	done := make(chan error, 1)
	go func() { done <- Serve(toServer, fromServer) }()
	json.NewEncoder(clientWrites).Encode(Message{Type: "render", Protocol: ProtocolVersion,
		HTML: `<link rel="stylesheet" href="a.css"/><p>x</p>`, BaseURL: "https://example.test/"})
	line, err := bufio.NewReader(clientReads).ReadBytes('\n')
	if err != nil || !strings.Contains(string(line), `"fetch"`) {
		t.Fatalf("first message = %q, %v", line, err)
	}
	clientWrites.Close()
	if err := <-done; err == nil {
		t.Error("Serve returned no error after the client closed its stream")
	}
}

// A family mapped to a base-14 font draws with that one face, bold included
// and without simulated bold (text render mode 2, fill and stroke);
// a name that is not a base-14 font is refused rather than read as a file.
func TestServe_maps_font_families_to_base14_fonts(t *testing.T) {
	_, result := client(t, Message{
		Type: "render", Protocol: ProtocolVersion,
		HTML:         `<body style="font-family: Arial Unicode MS"><h3>Heading</h3><p>Text</p></body>`,
		FontFamilies: map[string]string{"Arial Unicode MS": "Helvetica"},
	}, nil)
	if result.Type != "result" {
		t.Fatalf("result = %+v", result)
	}
	doc, err := writer.ReadPDF(result.PDF)
	if err != nil {
		t.Fatal(err)
	}
	if fonts := strings.Join(doc.PageFonts(0), " "); fonts != "Helvetica" {
		t.Errorf("fonts = %q, want only Helvetica: the bold heading uses the family's one face", fonts)
	}
	content, err := doc.PageContent(0)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "2 Tr") {
		t.Errorf("the bold heading was drawn with simulated bold")
	}

	_, result = client(t, Message{
		Type: "render", Protocol: ProtocolVersion, HTML: `<p>x</p>`,
		FontFamilies: map[string]string{"Mine": "/etc/fonts/secret.ttf"},
	}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "not one of the PDF base-14 fonts") {
		t.Errorf("result = %+v, want the base-14 error", result)
	}
}
