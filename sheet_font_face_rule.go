// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/FontFaceRule.java

package ufo

import "fmt"

type FontFaceRule struct {
	origin          StylesheetInfoOrigin
	ruleset         *Ruleset
	calculatedStyle *LazyEvaluated[CalculatedStyleI]
}

func NewFontFaceRule(origin StylesheetInfoOrigin) *FontFaceRule {
	f := &FontFaceRule{origin: origin}
	f.calculatedStyle = LazyEvaluatedLazy(func() CalculatedStyleI {
		return NewEmptyStyle().DeriveStyle(
			CascadedStyleCreateLayoutStyle(f.ruleset.GetPropertyDeclarations()...))
	})
	return f
}

func (f *FontFaceRule) AddContent(ruleset *Ruleset) {
	if f.ruleset != nil {
		panic(NewXRRuntimeException("Ruleset can only be set once"))
	}
	f.ruleset = ruleset
}

func (f *FontFaceRule) GetOrigin() StylesheetInfoOrigin {
	return f.origin
}

func (f *FontFaceRule) GetCalculatedStyle() CalculatedStyleI {
	return f.calculatedStyle.Get()
}

func (f *FontFaceRule) HasFontFamily() bool {
	return f.has("font-family")
}

func (f *FontFaceRule) HasFontWeight() bool {
	return f.has("font-weight")
}

func (f *FontFaceRule) HasFontStyle() bool {
	return f.has("font-style")
}

func (f *FontFaceRule) has(property string) bool {
	for _, declaration := range f.ruleset.GetPropertyDeclarations() {
		if property == declaration.GetPropertyName() {
			return true
		}
	}
	return false
}

func (f *FontFaceRule) String() string {
	ruleset := "null"
	if f.ruleset != nil {
		ruleset = f.ruleset.String()
	}
	return fmt.Sprintf("%s{origin: %s, %s}", "FontFaceRule", f.origin, ruleset)
}

func (f *FontFaceRule) ToString() string {
	return f.String()
}
