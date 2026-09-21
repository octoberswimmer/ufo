// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/layout/BoxBuilderHtml5WarningTest.java

package ufo_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf"
)

func TestBoxBuilderHtml5Warning_warnsWhenHtml5SemanticElementEncountered(t *testing.T) {
	for _, tag := range []string{"header", "footer", "nav", "article", "section", "aside", "main"} {
		t.Run(tag, func(t *testing.T) {
			html := fmt.Sprintf(`<html>
    <body>
        <%s>content</%s>
        <summary>foo</summary>
        <video/>
    </body>
</html>
`, tag, tag)
			tags := boxBuilderHtml5WarningTestRender(t, html).GetUnsupportedTags()
			boxBuilderHtml5WarningTestAssertInAnyOrder(t, tags, tag, "summary", "video")
		})
	}
}

func TestBoxBuilderHtml5Warning_noWarningForHtml4Tags(t *testing.T) {
	html := `<html>
    <body>
        <p>Hello</p>
        <div>World</div>
        <table><tr><td>cell</td></tr></table>
    </body>
</html>
`
	tags := boxBuilderHtml5WarningTestRender(t, html).GetUnsupportedTags()
	boxBuilderHtml5WarningTestAssertInAnyOrder(t, tags)
}

func TestBoxBuilderHtml5Warning_warnsOnFlexboxDisplayValue(t *testing.T) {
	html := `<html>
    <head><style>.box { display: flex; }</style></head>
    <body><div class="box">Hello</div></body>
</html>
`
	unsupportedCssFeatures := boxBuilderHtml5WarningTestRender(t, html).GetCss().GetUnsupportedCssFeatures()

	if strings.Join(unsupportedCssFeatures, "|") != "display: flex" {
		t.Errorf("unsupported CSS features %q, want exactly [display: flex]", unsupportedCssFeatures)
	}
}

func TestBoxBuilderHtml5Warning_warnsOnGridDisplayValue(t *testing.T) {
	html := `<html>
    <head><style>.box { display: grid; }</style></head>
    <body><div class="box">Hello</div></body>
</html>
`
	unsupportedCssFeatures := boxBuilderHtml5WarningTestRender(t, html).GetCss().GetUnsupportedCssFeatures()

	if strings.Join(unsupportedCssFeatures, "|") != "display: grid" {
		t.Errorf("unsupported CSS features %q, want exactly [display: grid]", unsupportedCssFeatures)
	}
}

func TestBoxBuilderHtml5Warning_warnsOnCss3PropertyNames(t *testing.T) {
	for _, property := range []string{"transition", "animation", "backdrop-filter", "box-shadow", "filter"} {
		t.Run(property, func(t *testing.T) {
			html := fmt.Sprintf(`<html>
    <head><style>.box { %s: none; } body {resize: both; display: flex}</style></head>
    <body><div class="box">Hello</div></body>
</html>
`, property)
			unsupportedCssFeatures := boxBuilderHtml5WarningTestRender(t, html).GetCss().GetUnsupportedCssFeatures()

			boxBuilderHtml5WarningTestAssertInAnyOrder(t, unsupportedCssFeatures, property, "resize", "display: flex")
		})
	}
}

func TestBoxBuilderHtml5Warning_noWarningForCss2Properties(t *testing.T) {
	html := `<html>
    <head><style>.box { color: red; font-size: 14px; margin: 10px; }</style></head>
    <body><div class="box">Hello</div></body>
</html>
`
	unsupportedCssFeatures := boxBuilderHtml5WarningTestRender(t, html).GetCss().GetUnsupportedCssFeatures()

	boxBuilderHtml5WarningTestAssertInAnyOrder(t, unsupportedCssFeatures)
}

// boxBuilderHtml5WarningTestRender lays out the document and returns the
// shared context. The Java test renders it to an image with the Swing
// Java2DRenderer, which is not ported; ITextRenderer builds and lays out the
// box tree through the same BoxBuilder and SharedContext.
func boxBuilderHtml5WarningTestRender(t *testing.T, html string) *ufo.SharedContext {
	t.Helper()
	document, err := ufo.XMLUtilNewDocumentBuilder().Parse(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	renderer := pdf.NewITextRenderer()
	if err := renderer.SetDocumentWithUrl(document, ""); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Layout(); err != nil {
		t.Fatal(err)
	}
	if renderer.GetRootBox() == nil {
		t.Fatal("no root box")
	}
	return renderer.GetSharedContext()
}

// boxBuilderHtml5WarningTestAssertInAnyOrder is containsExactlyInAnyOrder.
func boxBuilderHtml5WarningTestAssertInAnyOrder(t *testing.T, actual []string, expected ...string) {
	t.Helper()
	a := append([]string{}, actual...)
	e := append([]string{}, expected...)
	sort.Strings(a)
	sort.Strings(e)
	if strings.Join(a, "|") != strings.Join(e, "|") {
		t.Errorf("got %q, want %q in any order", actual, expected)
	}
}
