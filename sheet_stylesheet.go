// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/Stylesheet.java

package ufo

import "fmt"

// Stylesheet is a representation of a CSS style sheet. A Stylesheet has the
// sheet's rules in Ruleset, and has an origin--either user agent, user, or
// author. After instantiation, you can query the origin and the Ruleset.
type Stylesheet struct {
	// The info for this stylesheet
	uri    string
	origin StylesheetInfoOrigin

	fontFaceRules []*FontFaceRule
	importRules   []*StylesheetInfo
	// contents holds *Ruleset, *MediaRule and *PageRule values in the order
	// they were added (List<Object> in Java).
	contents []any
}

// NewStylesheet creates a new instance of Stylesheet.
func NewStylesheet(uri string, origin StylesheetInfoOrigin) *Stylesheet {
	return &Stylesheet{uri: uri, origin: origin}
}

// GetOrigin gets the origin attribute of the Stylesheet object.
func (s *Stylesheet) GetOrigin() StylesheetInfoOrigin {
	return s.origin
}

// GetURI gets the URI of the Stylesheet object.
func (s *Stylesheet) GetURI() string {
	return s.uri
}

// AddContent is addContent(Ruleset), the overload RulesetContainer declares.
func (s *Stylesheet) AddContent(ruleset *Ruleset) {
	s.contents = append(s.contents, ruleset)
}

// AddContentMediaRule is addContent(MediaRule).
func (s *Stylesheet) AddContentMediaRule(rule *MediaRule) {
	s.contents = append(s.contents, rule)
}

// AddContentPageRule is addContent(PageRule).
func (s *Stylesheet) AddContentPageRule(rule *PageRule) {
	s.contents = append(s.contents, rule)
}

// GetContents returns the *Ruleset, *MediaRule and *PageRule values of the
// sheet in the order they were added.
func (s *Stylesheet) GetContents() []any {
	return s.contents
}

func (s *Stylesheet) AddImportRule(info *StylesheetInfo) {
	s.importRules = append(s.importRules, info)
}

func (s *Stylesheet) GetImportRules() []*StylesheetInfo {
	return s.importRules
}

func (s *Stylesheet) AddFontFaceRule(rule *FontFaceRule) {
	s.fontFaceRules = append(s.fontFaceRules, rule)
}

func (s *Stylesheet) GetFontFaceRules() []*FontFaceRule {
	return s.fontFaceRules
}

func (s *Stylesheet) String() string {
	return fmt.Sprintf("%s{uri:%s, origin: %s}", "Stylesheet", s.uri, s.origin)
}

func (s *Stylesheet) ToString() string {
	return s.String()
}
