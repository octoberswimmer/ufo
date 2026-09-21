package ufo_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf"
)

// TestBlockBoxLayoutMatchesJava lays out randomly generated documents of
// nested block boxes (widths, heights, margins, padding, borders, box-sizing,
// min/max sizes, floats, clears, absolute and relative positioning, forced
// page breaks, @page sizes) with ITextRenderer and compares every box's
// position and dimensions, the floats and child layers of every layer, the
// pages and the RENDER dump with the output of the same walk over the box
// tree Flying Saucer's ITextRenderer.layout produced (flying-saucer
// 10.6.0-SNAPSHOT with OpenPDF 3.0.5). The archive holds, per case, the
// document followed by the Java output.
//
// The layout runs through the real renderer, with its base-14 fonts, because
// line heights and text widths come from the font metrics the Java side used.
func TestBlockBoxLayoutMatchesJava(t *testing.T) {
	f, err := os.Open("testdata/render/block_box_layout_java.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	cases := strings.Split(string(data), "\n=== case ")
	for _, tc := range cases[1:] {
		name, rest, _ := strings.Cut(tc, "\n")
		source, want, _ := strings.Cut(rest, "\n--- java\n")
		t.Run(name, func(t *testing.T) {
			got := layoutDump(t, source)
			if got != want {
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(want, "\n")
				for i := range wantLines {
					if i >= len(gotLines) || gotLines[i] != wantLines[i] {
						g := "<end of output>"
						if i < len(gotLines) {
							g = gotLines[i]
						}
						t.Fatalf("line %d differs from Java:\n got  %s\n want %s", i+1, g, wantLines[i])
					}
				}
				t.Fatalf("%d lines of output, Java has %d", len(gotLines), len(wantLines))
			}
		})
	}
}

func layoutDump(t *testing.T, source string) (result string) {
	t.Helper()
	doc, err := dom.ParseXMLString(source)
	if err != nil {
		t.Fatal(err)
	}
	renderer := pdf.NewITextRenderer()
	if err := renderer.SetDocumentWithUrl(doc, ""); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Layout(); err != nil {
		return fmt.Sprintf("EXCEPTION %v\n", err)
	}
	root := renderer.GetRootBox()
	c := renderer.GetSharedContext().NewLayoutContextInstance(nil)
	var sb strings.Builder
	layoutDumpWalk(root, "", &sb)
	layoutDumpLayers(root.GetLayer(), "", &sb)
	for _, p := range root.GetLayer().GetPages() {
		fmt.Fprintf(&sb, "page %d top=%d bottom=%d\n", p.GetPageNo(), p.GetTop(), p.GetBottom())
	}
	sb.WriteString(root.Dump(c, "", ufo.BoxDumpRender))
	sb.WriteByte('\n')
	return sb.String()
}

func layoutDumpWalk(b ufo.BoxI, indent string, sb *strings.Builder) {
	id := ""
	if b.GetElement() != nil {
		id = b.GetElement().GetAttribute("id")
	}
	anon := ""
	if b.IsAnonymous() {
		anon = " anon"
	}
	fmt.Fprintf(sb, "%s%s id=%s%s x=%d y=%d ax=%d ay=%d w=%d h=%d cw=%d l=%d r=%d tx=%d ty=%d",
		indent, ufo.BoxSimpleClassNameForTest(b), id, anon, b.GetX(), b.GetY(), b.GetAbsX(), b.GetAbsY(),
		b.GetWidth(), b.GetHeight(), b.GetContentWidth(), b.GetLeftMBP(), b.GetRightMBP(), b.GetTx(), b.GetTy())
	if bb, ok := b.(ufo.BlockBoxI); ok {
		fmt.Fprintf(sb, " ch=%d ct=%s fl=%t layer=%t", bb.GetChildrenHeight(), bb.GetChildrenContentType(),
			bb.IsFloated(), bb.GetLayer() != nil)
	}
	fmt.Fprintf(sb, " | %s\n", b.ToString())
	for _, child := range b.GetChildren() {
		layoutDumpWalk(child, indent+" ", sb)
	}
}

func layoutDumpLayers(layer *ufo.Layer, indent string, sb *strings.Builder) {
	id := ""
	if layer.GetMaster().GetElement() != nil {
		id = layer.GetMaster().GetElement().GetAttribute("id")
	}
	fmt.Fprintf(sb, "%slayer master=%s\n", indent, id)
	for _, f := range ufo.LayerFloatsForTest(layer) {
		fmt.Fprintf(sb, "%s float:\n", indent)
		layoutDumpWalk(f, indent+"  ", sb)
	}
	for _, child := range layer.GetChildren() {
		fmt.Fprintf(sb, "%s child:\n", indent)
		layoutDumpWalk(child.GetMaster(), indent+"  ", sb)
		layoutDumpLayers(child, indent+" ", sb)
	}
}
