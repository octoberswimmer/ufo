// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/bug/OutlineGenerationIssueTest.java

package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/pdf/writer"
)

const outlineGenerationIssueTestHtml1 = `<!DOCTYPE html>
<html>
<head><style> h2 { page-break-before: always;}</style></head><body>
<h1>Decision #1</h1><p>Decision 1 ...</p>
<h2>Attachment A</h2><p>Attachment A#1 ...</p>
<h2>Attachment B</h2><p>Attachment B#1 ...</p>
</body>
</html>
`

const outlineGenerationIssueTestHtml2 = `<!DOCTYPE html>
<html>
<head><style> h2 { page-break-before: always;}</style></head><body>
<h1>Decision #2</h1><p>Decision 2 ...</p>
<h2>Attachment B</h2><p>Attachment B#2...</p>
</body>
</html>
`

// The Java test writes target/test-outline1.pdf and target/test-outline2.pdf
// and passes when no exception is thrown. The port also checks the pages and
// the outline of each file against the files the Java implementation writes
// (Flying Saucer 10 with OpenPDF 3.0.5): each document keeps its own outline
// tree, pointing at its own pages.
func TestOutlineGenerationIssue_outline(t *testing.T) {
	outline1 := outlineGenerationIssueTestWritePDF(t, NewITextOutputDevice(26.666666), outlineGenerationIssueTestHtml2, outlineGenerationIssueTestHtml1)
	outlineGenerationIssueTestAssert(t, outline1,
		[]string{"Decision #2\nDecision 2 ...", "Attachment B\nAttachment B#2...",
			"Decision #1\nDecision 1 ...", "Attachment A\nAttachment A#1 ...", "Attachment B\nAttachment B#1 ..."},
		"[Decision #2 p0 750 [Attachment B p1 756]][Decision #1 p2 750 [Attachment A p3 756][Attachment B p4 756]]")
	outline2 := outlineGenerationIssueTestWritePDF(t, NewITextOutputDevice(26.666666), outlineGenerationIssueTestHtml1, outlineGenerationIssueTestHtml2)
	outlineGenerationIssueTestAssert(t, outline2,
		[]string{"Decision #1\nDecision 1 ...", "Attachment A\nAttachment A#1 ...", "Attachment B\nAttachment B#1 ...",
			"Decision #2\nDecision 2 ...", "Attachment B\nAttachment B#2..."},
		"[Decision #1 p0 750 [Attachment A p1 756][Attachment B p2 756]][Decision #2 p3 750 [Attachment B p4 756]]")
}

func outlineGenerationIssueTestAssert(t *testing.T, pdf []byte, pages []string, outline string) {
	t.Helper()
	doc, err := writer.ReadPDF(pdf)
	if err != nil {
		t.Fatal(err)
	}
	if doc.NumPages() != len(pages) {
		t.Fatalf("%d pages, want %d", doc.NumPages(), len(pages))
	}
	for i, want := range pages {
		if got, _ := doc.PageText(i); got != want {
			t.Errorf("page %d text %q, want %q", i, got, want)
		}
	}
	if got := outlineGenerationIssueTestFormat(doc.Outlines()); got != outline {
		t.Errorf("outline\n%s\nwant\n%s", got, outline)
	}
}

// outlineGenerationIssueTestFormat writes each item as [title pN top
// children...].
func outlineGenerationIssueTestFormat(items []*writer.ReadOutline) string {
	var sb strings.Builder
	for _, item := range items {
		fmt.Fprintf(&sb, "[%s", item.Title)
		if item.Dest != nil {
			fmt.Fprintf(&sb, " p%d %v", item.Dest.Page, item.Dest.Top)
		}
		if len(item.Children) > 0 {
			sb.WriteString(" " + outlineGenerationIssueTestFormat(item.Children))
		}
		sb.WriteString("]")
	}
	return sb.String()
}

func outlineGenerationIssueTestWritePDF(t *testing.T, outputDevice *ITextOutputDevice, html1 string, html2 string) []byte {
	t.Helper()
	var stream bytes.Buffer
	renderer := NewITextRendererWithDotsPerPointDotsPerPixelOutputDevice(outputDevice.GetDotsPerPoint(), 20, outputDevice)

	if err := renderer.SetDocumentFromString(html1); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Layout(); err != nil {
		t.Fatal(err)
	}
	if err := renderer.CreatePDFWithFinish(&stream, false); err != nil {
		t.Fatal(err)
	}

	if err := renderer.SetDocumentFromString(html2); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Layout(); err != nil {
		t.Fatal(err)
	}
	if err := renderer.WriteNextDocument(); err != nil {
		t.Fatal(err)
	}

	if err := renderer.FinishPDF(); err != nil {
		t.Fatal(err)
	}
	return stream.Bytes()
}
