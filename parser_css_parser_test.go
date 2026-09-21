// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/CSSParserTest.java

package ufo

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

// cssParserTestParser returns a parser that collects error messages, and
// registers the check that the Java class makes in tearDown: the test must
// not have reported any error.
func cssParserTestParser(t *testing.T) *CSSParser {
	t.Helper()
	var errors []string
	parser := NewCSSParser(CSSErrorHandlerFunc(func(uri string, message string) {
		errors = append(errors, message)
	}))
	t.Cleanup(func() {
		if len(errors) != 0 {
			t.Errorf("errors = %q, want none", errors)
		}
	})
	return parser
}

func cssParserTestParse(t *testing.T, parser *CSSParser, uri string, css string) *Stylesheet {
	t.Helper()
	stylesheet, err := parser.ParseStylesheet(uri, StylesheetInfoOriginAuthor, strings.NewReader(css))
	if err != nil {
		t.Fatal(err)
	}
	return stylesheet
}

func cssParserTestFirstRuleset(t *testing.T, stylesheet *Stylesheet) *Ruleset {
	t.Helper()
	if len(stylesheet.GetContents()) == 0 {
		t.Fatalf("stylesheet has no contents")
	}
	ruleset, ok := stylesheet.GetContents()[0].(*Ruleset)
	if !ok {
		t.Fatalf("first content is a %T, want *Ruleset", stylesheet.GetContents()[0])
	}
	return ruleset
}

func TestCSSParserRgba(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", `p {
    color: rgb(255, 165, 11);
    background-color: rgba(233, 99, 71, 0.5);
    border-top-color: rgba(-20, 255, 300, 1);
}
`)
	if len(stylesheet.GetContents()) != 1 {
		t.Fatalf("len(contents) = %d, want 1", len(stylesheet.GetContents()))
	}
	ruleset := cssParserTestFirstRuleset(t, stylesheet)

	want := []*PropertyDeclaration{
		cssParserTestCss(CSSNameColor, NewFSRGBColor(255, 165, 11)),
		cssParserTestCss(CSSNameBackgroundColor, NewFSRGBColorWithAlpha(233, 99, 71, 0.5)),
		cssParserTestCss(CSSNameBorderTopColor, NewFSRGBColorWithAlpha(0, 255, 255, 1.0)),
	}
	if got := ruleset.GetPropertyDeclarations(); !reflect.DeepEqual(got, want) {
		t.Errorf("declarations = %v, want %v", got, want)
	}
}

func TestCSSParserImageUrl(t *testing.T) {
	tests := []struct {
		imageUrl             string
		cssUrl               string
		expectedFullImageUrl string
	}{
		{"background.png", "/css/v1/sample.css", "/css/v1/background.png"},
		{"background.png", "sample.css", "background.png"},
		{"https://some.com/background.png", "/css/v1/sample.css", "https://some.com/background.png"},
		{"/img/background.png", "https://some.com/css/v1/sample.css", "https://some.com/img/background.png"},
		{"/img/background.png", "/css/v1/sample.css", "/img/background.png"},
		// file: bases are hierarchical absolute; preserve the scheme colon (not file/...).
		{"/img/background.png", "file:/css/v1/sample.css", "file:/img/background.png"},
		{"/img/background.png", "file:///css/v1/sample.css", "file:/img/background.png"},
		{"/img/background.png", "file://localhost/css/v1/sample.css", "file://localhost/img/background.png"},
		{"data:image/svg+xml;charset=utf8,%3Csvg xmlns=", "/css/v1/sample.css", "data:image/svg+xml;charset=utf8,%3Csvg xmlns="},
		{"blob:https://site.com/258186e7-a0a1-40e5-bcf8-5ac35f965454", "/css/v1/sample.css", "blob:https://site.com/258186e7-a0a1-40e5-bcf8-5ac35f965454"},
		{"blob:", "/css/v1/sample.css", "blob:"},
		// Synthetic inline stylesheet keys are absolute (opaque) URIs but must not
		// rewrite server-relative resource paths (e.g. @font-face src in <style>).
		{"/assets/fonts/MyFont.ttf", "inline:1", "/assets/fonts/MyFont.ttf"},
		{"assets/fonts/MyFont.ttf", "inline:1", "assets/fonts/MyFont.ttf"},
		{"/assets/fonts/MyFont.ttf", "style attribute", "/assets/fonts/MyFont.ttf"},
	}
	for _, tt := range tests {
		t.Run(tt.imageUrl+"|"+tt.cssUrl, func(t *testing.T) {
			parser := cssParserTestParser(t)
			css := fmt.Sprintf("div { background-image: url('%s') }", tt.imageUrl)
			stylesheet := cssParserTestParse(t, parser, tt.cssUrl, css)
			if len(stylesheet.GetContents()) != 1 {
				t.Fatalf("len(contents) = %d, want 1", len(stylesheet.GetContents()))
			}
			ruleset := cssParserTestFirstRuleset(t, stylesheet)

			want := []*PropertyDeclaration{
				NewPropertyDeclaration(CSSNameBackgroundImage,
					NewPropertyValueString(CSSPrimitiveValueCssUri, tt.expectedFullImageUrl, fmt.Sprintf("url('%s')", tt.imageUrl)),
					false, StylesheetInfoOriginAuthor),
			}
			if got := ruleset.GetPropertyDeclarations(); !reflect.DeepEqual(got, want) {
				t.Errorf("declarations = %v, want %v", got, want)
			}
		})
	}
}

func TestCSSParserFontFaceSrcKeepsServerRelativeUriAgainstInlineStylesheet(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "inline:42", `@font-face {
    font-family: 'MyFont';
    src: url('/assets/fonts/MyFont.ttf');
}
`)

	if len(stylesheet.GetFontFaceRules()) != 1 {
		t.Fatalf("len(fontFaceRules) = %d, want 1", len(stylesheet.GetFontFaceRules()))
	}
	src := stylesheet.GetFontFaceRules()[0].
		GetCalculatedStyle().
		ValueByName(CSSNameSrc).
		AsString()
	if src != "/assets/fonts/MyFont.ttf" {
		t.Errorf("src = %q, want /assets/fonts/MyFont.ttf", src)
	}
}

func TestCSSParserHasPseudoClassParsesAndMatches(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", `section:has(> p.note) { color: red; }
article:has(+ article.highlight) { color: blue; }
`)

	if len(stylesheet.GetContents()) != 2 {
		t.Fatalf("len(contents) = %d, want 2", len(stylesheet.GetContents()))
	}
	first := stylesheet.GetContents()[0].(*Ruleset)
	second := stylesheet.GetContents()[1].(*Ruleset)
	firstSelector := first.GetFSSelectors()[0]
	secondSelector := second.GetFSSelectors()[0]

	document, err := dom.ParseXMLString(`<root>
  <section id='has-child'><p class='note'>x</p></section>
  <section id='no-child'><p>x</p></section>
  <article id='a1'/>
  <article id='a2' class='highlight'/>
</root>
`)
	if err != nil {
		t.Fatal(err)
	}

	hasChild := cssParserTestElementById(t, document, "has-child")
	noChild := cssParserTestElementById(t, document, "no-child")
	a1 := cssParserTestElementById(t, document, "a1")
	a2 := cssParserTestElementById(t, document, "a2")

	attributes := cssParserTestDomAttributeResolver{}
	treeResolver := NewDOMTreeResolver()

	if !firstSelector.Matches(hasChild, attributes, treeResolver) {
		t.Errorf("section:has(> p.note) does not match #has-child")
	}
	if firstSelector.Matches(noChild, attributes, treeResolver) {
		t.Errorf("section:has(> p.note) matches #no-child")
	}
	if !secondSelector.Matches(a1, attributes, treeResolver) {
		t.Errorf("article:has(+ article.highlight) does not match #a1")
	}
	if secondSelector.Matches(a2, attributes, treeResolver) {
		t.Errorf("article:has(+ article.highlight) matches #a2")
	}
}

// cssParserTestDeclaration is the expected name, primitive type and float
// value of a declaration.
type cssParserTestDeclaration struct {
	cssName       *CSSName
	primitiveType int16
	floatValue    float32
}

func cssParserTestCheckDeclaration(t *testing.T, declaration *PropertyDeclaration, want cssParserTestDeclaration) {
	t.Helper()
	if declaration.GetCSSName() != want.cssName {
		t.Errorf("CSS name = %v, want %v", declaration.GetCSSName(), want.cssName)
	}
	if declaration.GetValue().GetPrimitiveType() != want.primitiveType {
		t.Errorf("%v: primitive type = %d, want %d", want.cssName, declaration.GetValue().GetPrimitiveType(), want.primitiveType)
	}
	if got := declaration.GetValue().GetFloatValueWithUnitType(want.primitiveType); got != want.floatValue {
		t.Errorf("%v: float value = %v, want %v", want.cssName, got, want.floatValue)
	}
}

func TestCSSParserParsesMultiColumnProperties(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", `div {
    column-count: 3;
    column-gap: 24px;
    column-width: 180px;
}
`)
	ruleset := cssParserTestFirstRuleset(t, stylesheet)
	declarations := ruleset.GetPropertyDeclarations()

	if len(declarations) != 3 {
		t.Fatalf("len(declarations) = %d, want 3", len(declarations))
	}
	cssParserTestCheckDeclaration(t, declarations[0], cssParserTestDeclaration{CSSNameColumnCount, CSSPrimitiveValueCssNumber, 3})
	cssParserTestCheckDeclaration(t, declarations[1], cssParserTestDeclaration{CSSNameColumnGap, CSSPrimitiveValueCssPx, 24})
	cssParserTestCheckDeclaration(t, declarations[2], cssParserTestDeclaration{CSSNameColumnWidth, CSSPrimitiveValueCssPx, 180})
}

func TestCSSParserExpandsColumnsShorthand(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", "div { columns: 12em 4; }\n")
	ruleset := cssParserTestFirstRuleset(t, stylesheet)
	declarations := ruleset.GetPropertyDeclarations()

	if len(declarations) != 2 {
		t.Fatalf("len(declarations) = %d, want 2", len(declarations))
	}
	if declarations[0].GetCSSName() != CSSNameColumnWidth {
		t.Errorf("CSS name = %v, want column-width", declarations[0].GetCSSName())
	}
	cssParserTestCheckDeclaration(t, declarations[1], cssParserTestDeclaration{CSSNameColumnCount, CSSPrimitiveValueCssNumber, 4})
}

func TestCSSParserExpandsColumnsShorthandWithSingleValue(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", "div { columns: 3; }\n")
	ruleset := cssParserTestFirstRuleset(t, stylesheet)
	declarations := ruleset.GetPropertyDeclarations()

	if len(declarations) != 2 {
		t.Fatalf("len(declarations) = %d, want 2", len(declarations))
	}
	if declarations[0].GetCSSName() != CSSNameColumnWidth {
		t.Errorf("CSS name = %v, want column-width", declarations[0].GetCSSName())
	}
	if declarations[0].GetValue().GetPrimitiveType() != CSSPrimitiveValueCssIdent {
		t.Errorf("primitive type = %d, want CSS_IDENT", declarations[0].GetValue().GetPrimitiveType())
	}
	if declarations[0].GetValue().GetCssText() != "auto" {
		t.Errorf("css text = %q, want auto", declarations[0].GetValue().GetCssText())
	}
	cssParserTestCheckDeclaration(t, declarations[1], cssParserTestDeclaration{CSSNameColumnCount, CSSPrimitiveValueCssNumber, 3})
}

func TestCSSParserParsesTransformFunctionsAndAngleUnits(t *testing.T) {
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", `div {
    transform: translate(10px, 20%) rotate(0.25turn) skewX(50grad);
    transform-origin: right bottom;
}
`)
	ruleset := cssParserTestFirstRuleset(t, stylesheet)
	declarations := ruleset.GetPropertyDeclarations()

	if len(declarations) != 2 {
		t.Fatalf("len(declarations) = %d, want 2", len(declarations))
	}
	if declarations[0].GetCSSName() != CSSNameTransform {
		t.Errorf("CSS name = %v, want transform", declarations[0].GetCSSName())
	}
	if declarations[1].GetCSSName() != CSSNameTransformOrigin {
		t.Errorf("CSS name = %v, want transform-origin", declarations[1].GetCSSName())
	}

	functions := declarations[0].GetValue().GetValues()
	if len(functions) != 3 {
		t.Fatalf("len(functions) = %d, want 3", len(functions))
	}
	// Function names are lower-cased by the tokenizer (CSS identifiers are case-insensitive).
	var names []string
	for _, v := range functions {
		names = append(names, v.(*PropertyValue).GetFunction().GetName())
	}
	if want := []string{"translate", "rotate", "skewx"}; !reflect.DeepEqual(names, want) {
		t.Errorf("function names = %q, want %q", names, want)
	}

	// "turn" has no dedicated DOM unit, so it's normalized to degrees while parsing: 0.25turn == 90deg
	rotateAngle := functions[1].(*PropertyValue).GetFunction().GetParameters()[0]
	if rotateAngle.GetPrimitiveType() != CSSPrimitiveValueCssDeg {
		t.Errorf("rotate: primitive type = %d, want CSS_DEG", rotateAngle.GetPrimitiveType())
	}
	if rotateAngle.GetFloatValue() != 90 {
		t.Errorf("rotate: float value = %v, want 90", rotateAngle.GetFloatValue())
	}

	skewAngle := functions[2].(*PropertyValue).GetFunction().GetParameters()[0]
	if skewAngle.GetPrimitiveType() != CSSPrimitiveValueCssGrad {
		t.Errorf("skewX: primitive type = %d, want CSS_GRAD", skewAngle.GetPrimitiveType())
	}
	if skewAngle.GetFloatValue() != 50 {
		t.Errorf("skewX: float value = %v, want 50", skewAngle.GetFloatValue())
	}

	origin := declarations[1].GetValue().GetValues()
	if got := origin[0].(*PropertyValue).GetFloatValue(); got != 100 { // right
		t.Errorf("origin[0] = %v, want 100", got)
	}
	if got := origin[1].(*PropertyValue).GetFloatValue(); got != 100 { // bottom
		t.Errorf("origin[1] = %v, want 100", got)
	}
}

func TestCSSParserParsesTurnAsANonFirstFunctionArgument(t *testing.T) {
	// A DIMENSION token (like "turn") appearing after a comma, rather than as the first
	// argument, is a separate code path in expr() from the one exercised above.
	parser := cssParserTestParser(t)
	stylesheet := cssParserTestParse(t, parser, "", "div { transform: skew(50grad, 0.25turn); }\n")
	ruleset := cssParserTestFirstRuleset(t, stylesheet)
	declarations := ruleset.GetPropertyDeclarations()

	if len(declarations) != 1 {
		t.Fatalf("len(declarations) = %d, want 1", len(declarations))
	}
	functions := declarations[0].GetValue().GetValues()
	args := functions[0].(*PropertyValue).GetFunction().GetParameters()

	if args[0].GetPrimitiveType() != CSSPrimitiveValueCssGrad {
		t.Errorf("args[0]: primitive type = %d, want CSS_GRAD", args[0].GetPrimitiveType())
	}
	if args[1].GetPrimitiveType() != CSSPrimitiveValueCssDeg {
		t.Errorf("args[1]: primitive type = %d, want CSS_DEG", args[1].GetPrimitiveType())
	}
	if args[1].GetFloatValue() != 90 {
		t.Errorf("args[1]: float value = %v, want 90", args[1].GetFloatValue())
	}
}

func cssParserTestCss(property *CSSName, color *FSRGBColor) *PropertyDeclaration {
	return NewPropertyDeclaration(property, NewPropertyValueFSColor(color), false, StylesheetInfoOriginAuthor)
}

func cssParserTestElementById(t *testing.T, document *dom.Document, id string) *dom.Element {
	t.Helper()
	node := document.GetDocumentElement().GetFirstChild()
	for node != nil {
		if node.GetNodeType() == dom.ElementNode {
			element := node.(*dom.Element)
			if id == element.GetAttribute("id") {
				return element
			}
		}
		node = node.GetNextSibling()
	}
	t.Fatalf("Element with id '%s' not found", id)
	return nil
}

// cssParserTestDomAttributeResolver is the anonymous AttributeResolver of the
// Java test's domAttributeResolver().
type cssParserTestDomAttributeResolver struct{}

func (r cssParserTestDomAttributeResolver) GetAttributeValue(e dom.Node, attrName string) *string {
	if element, ok := e.(*dom.Element); ok && element.HasAttribute(attrName) {
		value := element.GetAttribute(attrName)
		return &value
	}
	return nil
}

func (r cssParserTestDomAttributeResolver) GetAttributeValueWithNamespaceURI(e dom.Node, namespaceURI *string, attrName string) *string {
	return r.GetAttributeValue(e, attrName)
}

func (r cssParserTestDomAttributeResolver) attribute(e dom.Node, attrName string) string {
	if value := r.GetAttributeValue(e, attrName); value != nil {
		return *value
	}
	return ""
}

func (r cssParserTestDomAttributeResolver) GetClass(e dom.Node) string {
	return r.attribute(e, "class")
}

func (r cssParserTestDomAttributeResolver) GetID(e dom.Node) string {
	return r.attribute(e, "id")
}

func (r cssParserTestDomAttributeResolver) GetNonCssStyling(e dom.Node) string {
	return ""
}

func (r cssParserTestDomAttributeResolver) GetElementStyling(e dom.Node) string {
	return ""
}

func (r cssParserTestDomAttributeResolver) GetLang(e dom.Node) string {
	return r.attribute(e, "lang")
}

func (r cssParserTestDomAttributeResolver) IsLink(e dom.Node) bool {
	return false
}

func (r cssParserTestDomAttributeResolver) IsVisited(e dom.Node) bool {
	return false
}

func (r cssParserTestDomAttributeResolver) IsHover(e dom.Node) bool {
	return false
}

func (r cssParserTestDomAttributeResolver) IsActive(e dom.Node) bool {
	return false
}

func (r cssParserTestDomAttributeResolver) IsFocus(e dom.Node) bool {
	return false
}
