// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/ParserTest.java

package ufo

import (
	"strings"
	"testing"
	"time"
)

const parserTestTest = "div { background-image: url('something') }\n"

func parserTestErrorHandler(t *testing.T) CSSErrorHandler {
	return CSSErrorHandlerFunc(func(uri string, message string) {
		t.Log(message)
	})
}

func TestParserCssParsingPerformance(t *testing.T) {
	count := 10_000
	longTest := strings.Repeat(parserTestTest, count)
	if len(longTest) != len(parserTestTest)*count {
		t.Fatalf("Long enough input: len = %d, want %d", len(longTest), len(parserTestTest)*count)
	}
	errorHandler := parserTestErrorHandler(t)

	var total time.Duration
	for i := 0; i < 40; i++ {
		start := time.Now()
		p := NewCSSParser(errorHandler)
		stylesheet, err := p.ParseStylesheet("", StylesheetInfoOriginUserAgent, strings.NewReader(longTest))
		total += time.Since(start)
		if err != nil {
			t.Fatal(err)
		}

		if len(stylesheet.GetContents()) != count {
			t.Fatalf("len(contents) = %d, want %d", len(stylesheet.GetContents()), count)
		}
	}
	t.Logf("Average %d ms", (total / 40).Milliseconds())

	total = 0
	for i := 0; i < 10; i++ {
		start := time.Now()
		p := NewCSSParser(errorHandler)
		stylesheet, err := p.ParseStylesheet("", StylesheetInfoOriginUserAgent, strings.NewReader(longTest))
		total += time.Since(start)
		if err != nil {
			t.Fatal(err)
		}
		if len(stylesheet.GetContents()) != count {
			t.Fatalf("len(contents) = %d, want %d", len(stylesheet.GetContents()), count)
		}
	}
	t.Logf("Average %d ms", (total / 10).Milliseconds())

	p := NewCSSParser(errorHandler)

	total = 0
	for i := 0; i < 10; i++ {
		start := time.Now()
		for j := 0; j < 10000; j++ {
			stylesheet, err := p.ParseStylesheet("", StylesheetInfoOriginUserAgent, strings.NewReader(parserTestTest))
			if err != nil {
				t.Fatal(err)
			}
			if stylesheet.GetURI() != "" {
				t.Fatalf("URI = %q, want none", stylesheet.GetURI())
			}
			if stylesheet.GetOrigin() != StylesheetInfoOriginUserAgent {
				t.Fatalf("origin = %v, want USER_AGENT", stylesheet.GetOrigin())
			}
			if len(stylesheet.GetContents()) != 1 {
				t.Fatalf("len(contents) = %d, want 1", len(stylesheet.GetContents()))
			}
		}
		total += time.Since(start)
	}
	t.Logf("Average %d ms", (total / 10).Milliseconds())
}

func TestParserParseCss(t *testing.T) {
	p := NewCSSParser(parserTestErrorHandler(t))

	stylesheet, err := p.ParseStylesheet("", StylesheetInfoOriginUserAgent, strings.NewReader(parserTestTest))
	if err != nil {
		t.Fatal(err)
	}
	if len(stylesheet.GetContents()) != 1 {
		t.Fatalf("len(contents) = %d, want 1", len(stylesheet.GetContents()))
	}
	ruleset, ok := stylesheet.GetContents()[0].(*Ruleset)
	if !ok {
		t.Fatalf("first content is a %T, want *Ruleset", stylesheet.GetContents()[0])
	}
	if len(ruleset.GetFSSelectors()) != 1 {
		t.Fatalf("len(selectors) = %d, want 1", len(ruleset.GetFSSelectors()))
	}
	if ruleset.GetFSSelectors()[0] == nil {
		t.Errorf("selector is nil")
	}
	if len(ruleset.GetPropertyDeclarations()) != 1 {
		t.Fatalf("len(declarations) = %d, want 1", len(ruleset.GetPropertyDeclarations()))
	}
	propertyDeclaration := ruleset.GetPropertyDeclarations()[0]
	if propertyDeclaration.GetPropertyName() != "background-image" {
		t.Errorf("property name = %q, want background-image", propertyDeclaration.GetPropertyName())
	}
	if propertyDeclaration.GetCSSName().ToString() != "background-image" {
		t.Errorf("CSS name = %q, want background-image", propertyDeclaration.GetCSSName().ToString())
	}
	if propertyDeclaration.GetValue().GetCssText() != "url('something')" {
		t.Errorf("css text = %q, want url('something')", propertyDeclaration.GetValue().GetCssText())
	}
}
