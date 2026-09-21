// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/Matcher.java
// and CascadedStyle.java. Flying Saucer has no JUnit test for them.

package ufo

import (
	"io"
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

// matcherTestStylesheetFactory resolves style attribute text through a map.
type matcherTestStylesheetFactory struct {
	declarations map[string]*Ruleset
}

func (f *matcherTestStylesheetFactory) Parse(reader io.Reader, info *StylesheetInfo) *Stylesheet {
	panic("not used")
}

func (f *matcherTestStylesheetFactory) ParseWithUriOrigin(reader io.Reader, uri string, origin StylesheetInfoOrigin) *Stylesheet {
	panic("not used")
}

func (f *matcherTestStylesheetFactory) ParseStyleDeclaration(origin StylesheetInfoOrigin, style string) *Ruleset {
	return f.declarations[style]
}

func (f *matcherTestStylesheetFactory) GetStylesheet(si *StylesheetInfo) *Stylesheet {
	panic("not used")
}

func matcherTestDisplay(display *IdentValue, important bool, origin StylesheetInfoOrigin) *PropertyDeclaration {
	return NewPropertyDeclaration(CSSNameDisplay, NewPropertyValueIdentValue(display), important, origin)
}

// matcherTestRuleset creates an author rule set "selectors { display: ... }".
// Each build function configures the first selector of one chain and returns
// the selector to add to the rule set.
func matcherTestRuleset(display *IdentValue, builds ...func(s *Selector) *Selector) *Ruleset {
	ruleset := NewRuleset(StylesheetInfoOriginAuthor)
	ruleset.AddProperty(matcherTestDisplay(display, false, StylesheetInfoOriginAuthor))
	for _, build := range builds {
		ruleset.AddFSSelector(build(NewSelector(ruleset)))
	}
	return ruleset
}

func matcherTestNamed(name string) func(s *Selector) *Selector {
	return func(s *Selector) *Selector {
		s.SetName(name)
		return s
	}
}

// matcherTestChain builds "ancestor descendant" or "ancestor > descendant" as
// CSSParser.mergeSimpleSelectors does.
func matcherTestChain(ancestor string, axis SelectorAxis, descendant string) func(s *Selector) *Selector {
	return func(first *Selector) *Selector {
		first.SetName(ancestor)
		second := NewSelector(first.GetRuleset())
		second.SetName(descendant)
		second.SetAxis(axis)
		second.SetSpecificityD(second.GetSpecificityD() + first.GetSpecificityD())
		first.SetChainedSelector(second)
		return first
	}
}

func matcherTestSheet(contents ...any) *Stylesheet {
	sheet := NewStylesheet("test.css", StylesheetInfoOriginAuthor)
	for _, content := range contents {
		switch c := content.(type) {
		case *Ruleset:
			sheet.AddContent(c)
		case *MediaRule:
			sheet.AddContentMediaRule(c)
		case *PageRule:
			sheet.AddContentPageRule(c)
		}
	}
	return sheet
}

func matcherTestNew(sheets ...*Stylesheet) *Matcher {
	return NewMatcher(NewDOMTreeResolver(), &newmatchTestAttributeResolver{}, nil, sheets, "print")
}

func matcherTestAssertDisplay(t *testing.T, matcher *Matcher, doc *dom.Document, id string, expected *IdentValue) {
	t.Helper()
	style := matcher.GetCascadedStyle(newmatchTestElement(t, doc, id), false)
	actual := style.GetIdent(CSSNameDisplay)
	if actual != expected {
		t.Errorf("display of #%s = %v, expected %v", id, actual, expected)
	}
}

func TestMatcher_higherSpecificityWinsWhateverTheSourceOrder(t *testing.T) {
	doc := newmatchTestParse(t)
	matcher := matcherTestNew(matcherTestSheet(
		matcherTestRuleset(IdentValueNone, func(s *Selector) *Selector { s.AddIDCondition("p1"); return s }),
		matcherTestRuleset(IdentValueInline, func(s *Selector) *Selector { s.AddClassCondition("intro"); return s }),
		matcherTestRuleset(IdentValueBlock, matcherTestNamed("p")),
	))
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueNone)
	matcherTestAssertDisplay(t, matcher, doc, "p2", IdentValueBlock)
	matcherTestAssertDisplay(t, matcher, doc, "link", nil)
}

func TestMatcher_laterRuleWinsAtEqualSpecificity(t *testing.T) {
	doc := newmatchTestParse(t)
	matcher := matcherTestNew(
		matcherTestSheet(matcherTestRuleset(IdentValueBlock, matcherTestNamed("p"))),
		matcherTestSheet(matcherTestRuleset(IdentValueInline, matcherTestNamed("p"))),
	)
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueInline)
}

func TestMatcher_descendantAndChildAxes(t *testing.T) {
	doc := newmatchTestParse(t)
	matcher := matcherTestNew(matcherTestSheet(
		matcherTestRuleset(IdentValueBlock, matcherTestChain("body", SelectorAxisDescendantAxis, "em")),
		matcherTestRuleset(IdentValueInline, matcherTestChain("body", SelectorAxisChildAxis, "p")),
		matcherTestRuleset(IdentValueListItem, matcherTestChain("ul", SelectorAxisChildAxis, "li")),
		matcherTestRuleset(IdentValueTableCell, matcherTestChain("div", SelectorAxisChildAxis, "em")),
	))
	// em1 is a descendant of body at depth 3, and not a child of a div.
	matcherTestAssertDisplay(t, matcher, doc, "em1", IdentValueBlock)
	// The p elements are grandchildren of body.
	matcherTestAssertDisplay(t, matcher, doc, "p1", nil)
	matcherTestAssertDisplay(t, matcher, doc, "li3", IdentValueListItem)
}

func TestMatcher_elementsMatchedByTheSameSelectorsShareAMapper(t *testing.T) {
	doc := newmatchTestParse(t)
	matcher := matcherTestNew(matcherTestSheet(
		matcherTestRuleset(IdentValueListItem, matcherTestNamed("li")),
	))
	matcher.GetCascadedStyle(newmatchTestElement(t, doc, "li1"), false)
	matcher.GetCascadedStyle(newmatchTestElement(t, doc, "li2"), false)
	li1 := matcher.matcherMap[newmatchTestElement(t, doc, "li1")]
	li2 := matcher.matcherMap[newmatchTestElement(t, doc, "li2")]
	list := matcher.matcherMap[newmatchTestElement(t, doc, "list")]
	if li1 == nil || li1 != li2 {
		t.Errorf("li1 and li2 have the mappers %p and %p", li1, li2)
	}
	if list == nil || list == li1 {
		t.Errorf("the list has the mapper %p", list)
	}
}

func TestMatcher_importanceAndOriginOutrankSpecificity(t *testing.T) {
	doc := newmatchTestParse(t)

	userAgent := NewRuleset(StylesheetInfoOriginUserAgent)
	userAgent.AddProperty(matcherTestDisplay(IdentValueBlock, true, StylesheetInfoOriginUserAgent))
	s := NewSelector(userAgent)
	s.AddIDCondition("p1")
	userAgent.AddFSSelector(s)

	userImportant := NewRuleset(StylesheetInfoOriginUser)
	userImportant.AddProperty(matcherTestDisplay(IdentValueTableCell, true, StylesheetInfoOriginUser))
	userImportant.AddFSSelector(matcherTestNamed("p")(NewSelector(userImportant)))

	authorImportant := NewRuleset(StylesheetInfoOriginAuthor)
	authorImportant.AddProperty(matcherTestDisplay(IdentValueInline, true, StylesheetInfoOriginAuthor))
	authorImportant.AddFSSelector(matcherTestNamed("p")(NewSelector(authorImportant)))

	author := matcherTestRuleset(IdentValueNone, func(s *Selector) *Selector { s.AddIDCondition("p1"); return s })

	matcher := matcherTestNew(matcherTestSheet(userImportant, authorImportant, author, userAgent))
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueTableCell)

	matcher = matcherTestNew(matcherTestSheet(authorImportant, author, userAgent))
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueInline)

	matcher = matcherTestNew(matcherTestSheet(author, userAgent))
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueNone)
}

func TestMatcher_elementAndNonCssStyling(t *testing.T) {
	doc := newmatchTestParse(t)
	newmatchTestElement(t, doc, "p1").SetAttribute("style", "display: table-cell")
	newmatchTestElement(t, doc, "p1").SetAttribute("noncss", "display: list-item")
	newmatchTestElement(t, doc, "link").SetAttribute("noncss", "display: list-item")
	newmatchTestElement(t, doc, "p2").SetAttribute("noncss", "display: list-item")

	styleAttribute := NewRuleset(StylesheetInfoOriginAuthor)
	styleAttribute.AddProperty(matcherTestDisplay(IdentValueTableCell, false, StylesheetInfoOriginAuthor))
	nonCss := NewRuleset(StylesheetInfoOriginAuthor)
	nonCss.AddProperty(matcherTestDisplay(IdentValueListItem, false, StylesheetInfoOriginAuthor))
	factory := &matcherTestStylesheetFactory{declarations: map[string]*Ruleset{
		"display: table-cell": styleAttribute,
		"display: list-item":  nonCss,
	}}

	sheet := matcherTestSheet(matcherTestRuleset(IdentValueBlock, matcherTestNamed("p")))
	matcher := NewMatcher(NewDOMTreeResolver(), &newmatchTestAttributeResolver{}, factory, []*Stylesheet{sheet}, "print")
	// The style attribute comes after every selector.
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueTableCell)
	// Non-CSS styling comes before every selector.
	matcherTestAssertDisplay(t, matcher, doc, "p2", IdentValueBlock)
	matcherTestAssertDisplay(t, matcher, doc, "link", IdentValueListItem)

	// Without a StylesheetFactory both kinds of styling are left out.
	matcher = matcherTestNew(sheet)
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueBlock)
	matcherTestAssertDisplay(t, matcher, doc, "link", nil)
}

func TestMatcher_elementWithoutMatchGetsTheEmptyStyle(t *testing.T) {
	doc := newmatchTestParse(t)
	matcher := matcherTestNew(matcherTestSheet(matcherTestRuleset(IdentValueBlock, matcherTestNamed("p"))))
	if style := matcher.GetCascadedStyle(newmatchTestElement(t, doc, "link"), false); style != CascadedStyleEmptyCascadedStyle {
		t.Errorf("expected CascadedStyleEmptyCascadedStyle")
	}
}

func TestMatcher_pseudoElements(t *testing.T) {
	doc := newmatchTestParse(t)
	matcher := matcherTestNew(matcherTestSheet(
		matcherTestRuleset(IdentValueBlock, func(s *Selector) *Selector {
			s.SetName("p")
			s.SetPseudoElement("before")
			return s
		}),
		matcherTestRuleset(IdentValueInline, matcherTestNamed("a")),
	))
	p1 := newmatchTestElement(t, doc, "p1")
	matcher.GetCascadedStyle(p1, false)
	before := matcher.GetPECascadedStyle(p1, "before")
	if before == nil || before.GetIdent(CSSNameDisplay) != IdentValueBlock {
		t.Errorf("p1::before has the style %v", before)
	}
	if after := matcher.GetPECascadedStyle(p1, "after"); after != nil {
		t.Errorf("p1::after has the style %v", after)
	}
	// The pseudo-element rule does not style the element itself.
	matcherTestAssertDisplay(t, matcher, doc, "p1", nil)
	if style := matcher.GetPECascadedStyle(newmatchTestElement(t, doc, "link"), "before"); style != nil {
		t.Errorf("a::before has the style %v", style)
	}
}

func TestMatcher_hoverIsTrackedAndAppliedOnRestyle(t *testing.T) {
	doc := newmatchTestParse(t)
	attRes := &newmatchTestAttributeResolver{hover: map[string]bool{}}
	sheet := matcherTestSheet(
		matcherTestRuleset(IdentValueBlock, matcherTestNamed("a")),
		matcherTestRuleset(IdentValueInline, func(s *Selector) *Selector {
			s.SetName("a")
			s.SetPseudoClass(SelectorHoverPseudoclass)
			return s
		}),
	)
	matcher := NewMatcher(NewDOMTreeResolver(), attRes, nil, []*Stylesheet{sheet}, "print")
	link := newmatchTestElement(t, doc, "link")
	matcherTestAssertDisplay(t, matcher, doc, "link", IdentValueBlock)
	if !matcher.IsHoverStyled(link) {
		t.Errorf("the link is not recorded as hover styled")
	}
	if matcher.IsHoverStyled(newmatchTestElement(t, doc, "p1")) {
		t.Errorf("p1 is recorded as hover styled")
	}

	attRes.hover["link"] = true
	// Without restyle the stored mapper is used.
	matcherTestAssertDisplay(t, matcher, doc, "link", IdentValueBlock)
	if actual := matcher.GetCascadedStyle(link, true).GetIdent(CSSNameDisplay); actual != IdentValueInline {
		t.Errorf("display after restyle = %v", actual)
	}
}

func TestMatcher_mediaRules(t *testing.T) {
	doc := newmatchTestParse(t)
	screen := NewMediaRule(StylesheetInfoOriginAuthor)
	screen.AddMedium("screen")
	screen.AddContent(matcherTestRuleset(IdentValueNone, matcherTestNamed("p")))
	printRule := NewMediaRule(StylesheetInfoOriginAuthor)
	printRule.AddMedium("print")
	printRule.AddContent(matcherTestRuleset(IdentValueInline, matcherTestNamed("a")))

	matcher := matcherTestNew(matcherTestSheet(
		matcherTestRuleset(IdentValueBlock, matcherTestNamed("p")),
		screen,
		printRule,
	))
	matcherTestAssertDisplay(t, matcher, doc, "p1", IdentValueBlock)
	matcherTestAssertDisplay(t, matcher, doc, "link", IdentValueInline)
}

func TestMatcher_selectorsWithTheSameOrderKeyReplaceEachOther(t *testing.T) {
	// The position keeps five digits in Selector.getOrder, so selector
	// 100001 has the key of selector 1 and replaces it in the TreeMap.
	doc := newmatchTestParse(t)
	var contents []any
	contents = append(contents, matcherTestRuleset(IdentValueBlock, matcherTestNamed("p")))
	for i := 0; i < 99999; i++ {
		contents = append(contents, matcherTestRuleset(IdentValueNone, matcherTestNamed("x")))
	}
	contents = append(contents, matcherTestRuleset(IdentValueInline, matcherTestNamed("a")))
	matcher := matcherTestNew(matcherTestSheet(contents...))
	matcherTestAssertDisplay(t, matcher, doc, "p1", nil)
	matcherTestAssertDisplay(t, matcher, doc, "link", IdentValueInline)
}

func matcherTestPageRule(name string, pseudoPage string, display *IdentValue, marginBoxes map[*MarginBoxName][]*PropertyDeclaration) *PageRule {
	ruleset := NewRuleset(StylesheetInfoOriginAuthor)
	ruleset.AddProperty(matcherTestDisplay(display, false, StylesheetInfoOriginAuthor))
	return NewPageRule(StylesheetInfoOriginAuthor, name, pseudoPage, marginBoxes, ruleset)
}

func TestMatcher_pageRulesAreSortedBySpecificityThenPosition(t *testing.T) {
	first := matcherTestPageRule("", "first", IdentValueNone, nil)
	named := matcherTestPageRule("cover", "", IdentValueTableCell, nil)
	plainA := matcherTestPageRule("", "", IdentValueBlock, nil)
	right := matcherTestPageRule("", "right", IdentValueInline, nil)
	plainB := matcherTestPageRule("", "", IdentValueListItem, nil)
	matcher := matcherTestNew(matcherTestSheet(named, first, plainA, right, plainB))

	expected := []*PageRule{plainA, right, plainB, first, named}
	for i, rule := range matcher.pageRules {
		if rule != expected[i] {
			t.Fatalf("page rule %d is not the expected one", i)
		}
	}

	cases := []struct {
		pageName   string
		pseudoPage string
		expected   *IdentValue
		properties int
	}{
		{"", "left", IdentValueListItem, 2},
		{"", "right", IdentValueListItem, 3},
		// The first page counts as a right page.
		{"", "first", IdentValueNone, 4},
		{"cover", "first", IdentValueTableCell, 5},
		{"other", "left", IdentValueListItem, 2},
	}
	for _, c := range cases {
		info := matcher.GetPageCascadedStyle(c.pageName, c.pseudoPage)
		if actual := info.GetPageStyle().GetIdent(CSSNameDisplay); actual != c.expected {
			t.Errorf("page %q:%q has display %v, expected %v", c.pageName, c.pseudoPage, actual, c.expected)
		}
		if len(info.GetProperties()) != c.properties {
			t.Errorf("page %q:%q has %d properties, expected %d", c.pageName, c.pseudoPage, len(info.GetProperties()), c.properties)
		}
	}
}

func TestMatcher_pageMarginBoxes(t *testing.T) {
	topCenter := []*PropertyDeclaration{matcherTestDisplay(IdentValueBlock, false, StylesheetInfoOriginAuthor)}
	topCenterFirst := []*PropertyDeclaration{matcherTestDisplay(IdentValueInline, false, StylesheetInfoOriginAuthor)}
	xmp := []*PropertyDeclaration{matcherTestDisplay(IdentValueNone, false, StylesheetInfoOriginAuthor)}
	matcher := matcherTestNew(matcherTestSheet(
		matcherTestPageRule("", "first", IdentValueBlock, map[*MarginBoxName][]*PropertyDeclaration{
			MarginBoxNameTopCenter: topCenterFirst,
		}),
		matcherTestPageRule("", "", IdentValueBlock, map[*MarginBoxName][]*PropertyDeclaration{
			MarginBoxNameTopCenter:        topCenter,
			MarginBoxNameFsPdfXmpMetadata: xmp,
		}),
	))

	info := matcher.GetPageCascadedStyle("", "first")
	if boxes := info.GetMarginBoxes(); len(boxes) != 1 || &boxes[MarginBoxNameTopCenter][0] != &topCenterFirst[0] {
		t.Errorf("the margin boxes of the first page are %v", boxes)
	}
	if list := info.GetXMPPropertyList(); len(list) != 1 || list[0] != xmp[0] {
		t.Errorf("the XMP property list is %v", list)
	}
	if !info.HasAny([]*MarginBoxName{MarginBoxNameTopLeft, MarginBoxNameTopCenter}) {
		t.Errorf("HasAny does not find top-center")
	}
	if info.HasAny([]*MarginBoxName{MarginBoxNameTopLeft, MarginBoxNameFsPdfXmpMetadata}) {
		t.Errorf("HasAny finds a box the page does not have")
	}

	style := info.CreateMarginBoxStyle(MarginBoxNameTopCenter, false)
	// The declared display is replaced by the important table-cell.
	if style.GetIdent(CSSNameDisplay) != IdentValueTableCell ||
		style.GetIdent(CSSNameVerticalAlign) != IdentValueMiddle ||
		style.GetIdent(CSSNameTextAlign) != IdentValueCenter {
		t.Errorf("the top-center style is %v", style.GetCascadedPropertyDeclarations())
	}
	if info.CreateMarginBoxStyle(MarginBoxNameTopLeft, false) != nil {
		t.Errorf("a style was created for a box the page does not have")
	}
	if created := info.CreateMarginBoxStyle(MarginBoxNameTopLeft, true); created == nil || created.CountAssigned() != 3 {
		t.Errorf("alwaysCreate gave %v", created)
	}

	// The rule's own map still has the XMP entry.
	if again := matcher.GetPageCascadedStyle("", "left"); len(again.GetXMPPropertyList()) != 1 {
		t.Errorf("the second PageInfo has no XMP property list")
	}
}

func TestMatcher_fontFaceRulesAreCollected(t *testing.T) {
	sheetA := NewStylesheet("a.css", StylesheetInfoOriginAuthor)
	ruleA := NewFontFaceRule(StylesheetInfoOriginAuthor)
	sheetA.AddFontFaceRule(ruleA)
	sheetB := NewStylesheet("b.css", StylesheetInfoOriginAuthor)
	ruleB := NewFontFaceRule(StylesheetInfoOriginAuthor)
	sheetB.AddFontFaceRule(ruleB)
	matcher := matcherTestNew(sheetA, sheetB)
	rules := matcher.GetFontFaceRules()
	if len(rules) != 2 || rules[0] != ruleA || rules[1] != ruleB {
		t.Errorf("GetFontFaceRules() = %v", rules)
	}
}

func TestCascadedStyle_declarationsAreReadInCSSNameOrder(t *testing.T) {
	names := []*CSSName{CSSNameWidth, CSSNameColor, CSSNameDisplay}
	var declarations []*PropertyDeclaration
	for _, name := range names {
		declarations = append(declarations, CascadedStyleCreateLayoutPropertyDeclaration(name, IdentValueNone))
	}
	style := CascadedStyleCreateLayoutStyle(declarations...)
	list := style.GetCascadedPropertyDeclarations()
	if len(list) != 3 || style.CountAssigned() != 3 {
		t.Fatalf("the style has %d declarations", len(list))
	}
	fingerprint := ""
	for i, declaration := range list {
		if i > 0 && list[i-1].GetCSSName().FS_ID >= declaration.GetCSSName().FS_ID {
			t.Errorf("declaration %d is out of order", i)
		}
		fingerprint += declaration.GetFingerprint()
	}
	if style.GetFingerprint() != fingerprint {
		t.Errorf("GetFingerprint() = %q, expected %q", style.GetFingerprint(), fingerprint)
	}
}

func TestCascadedStyle_createLayoutStyleWithStartingPoint(t *testing.T) {
	start := CascadedStyleCreateLayoutStyle(
		CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueBlock),
		CascadedStyleCreateLayoutPropertyDeclaration(CSSNameWidth, IdentValueNone))
	style := CascadedStyleCreateLayoutStyleWithStartingPoint(start, []*PropertyDeclaration{
		CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueInline)})
	if style.GetIdent(CSSNameDisplay) != IdentValueInline || style.GetIdent(CSSNameWidth) != IdentValueNone {
		t.Errorf("the derived style is %v", style.GetCascadedPropertyDeclarations())
	}
	if start.GetIdent(CSSNameDisplay) != IdentValueBlock {
		t.Errorf("the starting point was modified")
	}
	if !style.HasProperty(CSSNameWidth) || style.HasProperty(CSSNameColor) || style.PropertyByName(CSSNameColor) != nil {
		t.Errorf("HasProperty or PropertyByName gives the wrong answer")
	}
}

func TestCascadedStyle_createAnonymousStyle(t *testing.T) {
	style := CascadedStyleCreateAnonymousStyle(IdentValueBlock)
	declaration := style.PropertyByName(CSSNameDisplay)
	if declaration == nil || !declaration.IsImportant() || declaration.GetOrigin() != StylesheetInfoOriginUser ||
		declaration.AsIdentValue() != IdentValueBlock {
		t.Errorf("the anonymous style has the display declaration %v", declaration)
	}
}
