package pipe

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"regexp"
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

// A family with a fallback font draws the characters its own font cannot
// show with the bundled font, embedded as a subset, and keeps its own font
// for the rest.
func TestServe_draws_other_scripts_with_the_fallback_font(t *testing.T) {
	text := "Total 税込 ¥1,000 Привет 한국어 end"
	request := Message{
		Type: "render", Protocol: ProtocolVersion,
		HTML:          `<body style="font-family: Arial Unicode MS"><p>` + text + `</p></body>`,
		FontFamilies:  map[string]string{"Arial Unicode MS": "Helvetica"},
		FallbackFonts: map[string]string{"Arial Unicode MS": "WenQuanYi Micro Hei"},
	}
	_, result := client(t, request, nil)
	if result.Type != "result" {
		t.Fatalf("result = %+v", result)
	}
	doc, err := writer.ReadPDF(result.PDF)
	if err != nil {
		t.Fatal(err)
	}
	fonts := doc.PageFonts(0)
	if len(fonts) != 2 || fonts[0] != "Helvetica" || !regexp.MustCompile(`^[A-Z]{6}\+WenQuanYiMicroHei$`).MatchString(fonts[1]) {
		t.Errorf("fonts = %q, want Helvetica and a subset of WenQuanYi Micro Hei", fonts)
	}
	got, err := doc.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(got) != text {
		t.Errorf("text = %q, want %q", got, text)
	}
	// The whole font file has 5 MB; the subset holds the glyphs drawn.
	if len(result.PDF) > 200_000 {
		t.Errorf("the PDF has %d bytes: the font was not embedded as a subset", len(result.PDF))
	}

	_, again := client(t, request, nil)
	if !bytes.Equal(embeddedFontFile(t, result.PDF), embeddedFontFile(t, again.PDF)) {
		t.Errorf("two renderings of one document embed different font files")
	}
}

// embeddedFontFile returns the stream of the one FontFile2 of a PDF, as
// written.
func embeddedFontFile(t *testing.T, pdf []byte) []byte {
	t.Helper()
	found := regexp.MustCompile(`(?s)/Length1 \d+.*?stream\r?\n(.*?)endstream`).FindAllSubmatch(pdf, -1)
	if len(found) != 1 {
		t.Fatalf("the PDF has %d embedded font files, want 1", len(found))
	}
	return found[0][1]
}

// Text the family's own font can show is drawn as it is without a fallback
// font, and the bundled font is not embedded.
func TestServe_fallback_font_leaves_windows_1252_text_alone(t *testing.T) {
	request := Message{
		Type: "render", Protocol: ProtocolVersion,
		HTML:         `<body style="font-family: Arial Unicode MS"><p>Total ¥1,000 — café</p></body>`,
		FontFamilies: map[string]string{"Arial Unicode MS": "Helvetica"},
	}
	_, plain := client(t, request, nil)
	request.FallbackFonts = map[string]string{"Arial Unicode MS": "WenQuanYi Micro Hei"}
	_, withFallback := client(t, request, nil)
	if plain.Type != "result" || withFallback.Type != "result" {
		t.Fatalf("results = %+v, %+v", plain, withFallback)
	}
	doc, err := writer.ReadPDF(withFallback.PDF)
	if err != nil {
		t.Fatal(err)
	}
	if fonts := strings.Join(doc.PageFonts(0), " "); fonts != "Helvetica" {
		t.Errorf("fonts = %q, want only Helvetica", fonts)
	}
	plainContent, err := pageContent(plain.PDF)
	if err != nil {
		t.Fatal(err)
	}
	content, err := doc.PageContent(0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plainContent, content) {
		t.Errorf("the page content differs from the one without a fallback font:\n%s\nwant\n%s", content, plainContent)
	}
}

func pageContent(pdf []byte) ([]byte, error) {
	doc, err := writer.ReadPDF(pdf)
	if err != nil {
		return nil, err
	}
	return doc.PageContent(0)
}

func TestServe_rejects_unknown_fallback_fonts_and_families(t *testing.T) {
	_, result := client(t, Message{
		Type: "render", Protocol: ProtocolVersion, HTML: `<p>x</p>`,
		FontFamilies:  map[string]string{"Mine": "Helvetica"},
		FallbackFonts: map[string]string{"Mine": "/etc/fonts/secret.ttf"},
	}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "is not a font bundled with ufo (WenQuanYi Micro Hei)") {
		t.Errorf("result = %+v, want the bundled font error", result)
	}
	_, result = client(t, Message{
		Type: "render", Protocol: ProtocolVersion, HTML: `<p>x</p>`,
		FallbackFonts: map[string]string{"Nowhere": "WenQuanYi Micro Hei"},
	}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "there is no font family of that name") {
		t.Errorf("result = %+v, want the font family error", result)
	}
}

// A client that speaks version 1 is still served.
func TestServe_accepts_protocol_version_1(t *testing.T) {
	_, result := client(t, Message{Type: "render", Protocol: 1, HTML: "<p>x</p>"}, nil)
	if result.Type != "result" {
		t.Errorf("result = %+v", result)
	}
	_, result = client(t, Message{Type: "render", Protocol: 0, HTML: "<p>x</p>"}, nil)
	if result.Type != "error" || !strings.Contains(result.Message, "protocol version 0 is not supported") {
		t.Errorf("result = %+v", result)
	}
}

// Justified text spreads the space of a line over the characters of both
// fonts, so each line of a paragraph ends at the same place.
func TestServe_justifies_text_drawn_with_both_fonts(t *testing.T) {
	_, result := client(t, Message{
		Type: "render", Protocol: ProtocolVersion,
		HTML: `<body style="font-family: Arial Unicode MS"><p style="text-align: justify; width: 200px">` +
			`Justified text with 日本語 テキスト mixed into a paragraph that wraps over several lines.</p></body>`,
		FontFamilies:  map[string]string{"Arial Unicode MS": "Helvetica"},
		FallbackFonts: map[string]string{"Arial Unicode MS": "WenQuanYi Micro Hei"},
	}, nil)
	if result.Type != "result" {
		t.Fatalf("result = %+v", result)
	}
	doc, err := writer.ReadPDF(result.PDF)
	if err != nil {
		t.Fatal(err)
	}
	got, err := doc.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	// A line may end between two Japanese characters, so the comparison
	// leaves out all white space.
	if text := strings.Join(strings.Fields(got), ""); text != "Justifiedtextwith日本語テキストmixedintoaparagraphthatwrapsoverseverallines." {
		t.Errorf("text = %q", text)
	}
	content, err := doc.PageContent(0)
	if err != nil {
		t.Fatal(err)
	}
	// In a justified line, the last character a font draws before the other
	// font takes over is followed by its adjustment like any other.
	if !regexp.MustCompile(`>-?[\d.]+\]TJ\s*/F\d+ [\d.]+ Tf`).Match(content) {
		t.Errorf("no piece of a justified line ends with an adjustment before the font changes:\n%s", content)
	}
}
