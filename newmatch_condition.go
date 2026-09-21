// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/Condition.java

package ufo

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/octoberswimmer/ufo/dom"
)

// Condition is part of a Selector. Java declares it as an abstract class
// whose only instance method is matches; the static factory methods are the
// ConditionCreate... functions.
type Condition interface {
	Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool
}

// ConditionCreateAttributeExistsCondition creates the CSS condition
// [attribute]
func ConditionCreateAttributeExistsCondition(namespaceURI *string, name string) Condition {
	return newConditionAttributeExistsCondition(namespaceURI, name)
}

// ConditionCreateAttributePrefixCondition creates the CSS condition
// [attribute^=value]
func ConditionCreateAttributePrefixCondition(namespaceURI *string, name string, value string) Condition {
	return newConditionAttributePrefixCondition(namespaceURI, name, value)
}

// ConditionCreateAttributeSuffixCondition creates the CSS condition
// [attribute$=value]
func ConditionCreateAttributeSuffixCondition(namespaceURI *string, name string, value string) Condition {
	return newConditionAttributeSuffixCondition(namespaceURI, name, value)
}

// ConditionCreateAttributeSubstringCondition creates the CSS condition
// [attribute*=value]
func ConditionCreateAttributeSubstringCondition(namespaceURI *string, name string, value string) Condition {
	return newConditionAttributeSubstringCondition(namespaceURI, name, value)
}

// ConditionCreateAttributeEqualsCondition creates the CSS condition
// [attribute=value]
func ConditionCreateAttributeEqualsCondition(namespaceURI *string, name string, value string) Condition {
	return newConditionAttributeEqualsCondition(namespaceURI, name, value)
}

// ConditionCreateAttributeMatchesListCondition creates the CSS condition
// [attribute~=value]
func ConditionCreateAttributeMatchesListCondition(namespaceURI *string, name string, value string) Condition {
	return newConditionAttributeMatchesListCondition(namespaceURI, name, value)
}

// ConditionCreateAttributeMatchesFirstPartCondition creates the CSS condition
// [attribute|=value]
func ConditionCreateAttributeMatchesFirstPartCondition(namespaceURI *string, name string, value string) Condition {
	return newConditionAttributeMatchesFirstPartCondition(namespaceURI, name, value)
}

// ConditionCreateClassCondition creates the CSS condition .class
func ConditionCreateClassCondition(className string) Condition {
	return newConditionClassCondition(className)
}

// ConditionCreateIDCondition creates the CSS condition #ID
func ConditionCreateIDCondition(id string) Condition {
	return newConditionIDCondition(id)
}

// ConditionCreateLangCondition creates the CSS condition lang(Xx)
func ConditionCreateLangCondition(lang string) Condition {
	return newConditionLangCondition(lang)
}

// ConditionCreateFirstChildCondition creates the CSS condition that element
// has pseudo-class :first-child
func ConditionCreateFirstChildCondition() Condition {
	return &ConditionFirstChildCondition{}
}

// ConditionCreateLastChildCondition creates the CSS condition that element
// has pseudo-class :last-child
func ConditionCreateLastChildCondition() Condition {
	return &ConditionLastChildCondition{}
}

// ConditionCreateNthChildCondition creates the CSS condition that element has
// pseudo-class :nth-child(an+b)
func ConditionCreateNthChildCondition(number string) Condition {
	return ConditionNthChildConditionFromString(number)
}

// ConditionCreateEvenChildCondition creates the CSS condition that element
// has pseudo-class :even
func ConditionCreateEvenChildCondition() Condition {
	return &ConditionEvenChildCondition{}
}

// ConditionCreateOddChildCondition creates the CSS condition that element has
// pseudo-class :odd
func ConditionCreateOddChildCondition() Condition {
	return &ConditionOddChildCondition{}
}

// ConditionCreateLinkCondition creates the CSS condition that element has
// pseudo-class :link
func ConditionCreateLinkCondition() Condition {
	return &ConditionLinkCondition{}
}

// ConditionCreateUnsupportedCondition is for unsupported or invalid CSS
func ConditionCreateUnsupportedCondition() Condition {
	return &ConditionUnsupportedCondition{}
}

func ConditionCreateHasCondition(relativeSelectors []*SelectorHasRelativeSelector) Condition {
	return newConditionHasCondition(relativeSelectors)
}

// ConditionAttributeCompareCondition is the abstract base of the attribute
// conditions. compare is the Java abstract method, supplied by the
// constructor of each subclass.
type ConditionAttributeCompareCondition struct {
	namespaceURI *string
	name         string
	value        string
	compare      func(attrValue string, conditionValue string) bool
}

func (c *ConditionAttributeCompareCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	if attRes == nil {
		return false
	}
	val := attRes.GetAttributeValueWithNamespaceURI(e, c.namespaceURI, c.name)
	if val == nil {
		return false
	}

	return c.compare(*val, c.value)
}

type ConditionAttributeExistsCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributeExistsCondition(namespaceURI *string, name string) *ConditionAttributeExistsCondition {
	c := &ConditionAttributeExistsCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.compare = func(attrValue string, conditionValue string) bool {
		return attrValue != ""
	}
	return c
}

type ConditionAttributeEqualsCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributeEqualsCondition(namespaceURI *string, name string, value string) *ConditionAttributeEqualsCondition {
	c := &ConditionAttributeEqualsCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.value = value
	c.compare = func(attrValue string, conditionValue string) bool {
		return attrValue == conditionValue
	}
	return c
}

type ConditionAttributePrefixCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributePrefixCondition(namespaceURI *string, name string, value string) *ConditionAttributePrefixCondition {
	c := &ConditionAttributePrefixCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.value = value
	c.compare = func(attrValue string, conditionValue string) bool {
		return strings.HasPrefix(attrValue, conditionValue)
	}
	return c
}

type ConditionAttributeSuffixCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributeSuffixCondition(namespaceURI *string, name string, value string) *ConditionAttributeSuffixCondition {
	c := &ConditionAttributeSuffixCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.value = value
	c.compare = func(attrValue string, conditionValue string) bool {
		return strings.HasSuffix(attrValue, conditionValue)
	}
	return c
}

type ConditionAttributeSubstringCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributeSubstringCondition(namespaceURI *string, name string, value string) *ConditionAttributeSubstringCondition {
	c := &ConditionAttributeSubstringCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.value = value
	c.compare = func(attrValue string, conditionValue string) bool {
		return strings.Contains(attrValue, conditionValue)
	}
	return c
}

type ConditionAttributeMatchesListCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributeMatchesListCondition(namespaceURI *string, name string, value string) *ConditionAttributeMatchesListCondition {
	c := &ConditionAttributeMatchesListCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.value = value
	c.compare = func(attrValue string, conditionValue string) bool {
		ca := conditionSplit(attrValue, ' ')
		matched := false
		for _, s := range ca {
			if conditionValue == s {
				matched = true
				break
			}
		}
		return matched
	}
	return c
}

type ConditionAttributeMatchesFirstPartCondition struct {
	ConditionAttributeCompareCondition
}

func newConditionAttributeMatchesFirstPartCondition(namespaceURI *string, name string, value string) *ConditionAttributeMatchesFirstPartCondition {
	c := &ConditionAttributeMatchesFirstPartCondition{}
	c.namespaceURI = namespaceURI
	c.name = name
	c.value = value
	c.compare = func(attrValue string, conditionValue string) bool {
		ca := conditionSplit(attrValue, '-')
		// An attribute value made only of '-' characters splits into no
		// parts, and Java throws ArrayIndexOutOfBoundsException here.
		if len(ca) == 0 {
			panic(NewXRRuntimeException("Index 0 out of bounds for length 0"))
		}
		return conditionValue == ca[0]
	}
	return c
}

// ConditionClassCondition compares on Unicode code points. Java compares on
// UTF-16 code units; the results are the same because every whitespace
// character is a single code unit.
type ConditionClassCondition struct {
	className       []rune
	classNameLength int
}

func newConditionClassCondition(className string) *ConditionClassCondition {
	runes := []rune(className)
	return &ConditionClassCondition{className: runes, classNameLength: len(runes)}
}

func (c *ConditionClassCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	if attRes == nil {
		return false
	}
	// GetClass returns "" for the Java null, and an empty class attribute
	// contains no class name.
	cl := attRes.GetClass(e)
	return cl != "" && c.ContainsClassName(cl)
}

func (c *ConditionClassCondition) ContainsClassName(classAttribute string) bool {
	return c.containsClassNameWithFromIndex([]rune(classAttribute), -1)
}

func (c *ConditionClassCondition) containsClassNameWithFromIndex(classAttribute []rune, fromIndex int) bool {
	// This is much faster than calling `split()` and comparing individual values in a loop.
	// NOTE: In jQuery, for example, the attribute value first has whitespace normalized to spaces. But
	// in an XML DOM, space normalization in attributes is supposed to have happened already.
	index := conditionIndexOf(classAttribute, c.className, fromIndex)
	if index == -1 {
		return false
	}

	return c.isWhitespace(classAttribute, index-1) &&
		c.isWhitespace(classAttribute, index+c.classNameLength) ||
		c.containsClassNameWithFromIndex(classAttribute, index+c.classNameLength)
}

func (c *ConditionClassCondition) isWhitespace(classAttribute []rune, index int) bool {
	return index < 0 || index >= len(classAttribute) || conditionIsWhitespace(classAttribute[index])
}

// conditionIndexOf is String.indexOf(String, int): the index of the first
// occurrence of needle in haystack at or after fromIndex, or -1. fromIndex is
// limited to the range from 0 to the length of haystack.
func conditionIndexOf(haystack []rune, needle []rune, fromIndex int) int {
	if fromIndex < 0 {
		fromIndex = 0
	} else if fromIndex > len(haystack) {
		fromIndex = len(haystack)
	}
	for i := fromIndex; i+len(needle) <= len(haystack); i++ {
		found := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

// conditionIsWhitespace is Character.isWhitespace: a space, line or paragraph
// separator other than the non-breaking spaces, or one of U+0009 to U+000D
// and U+001C to U+001F.
func conditionIsWhitespace(r rune) bool {
	switch {
	case r >= 0x0009 && r <= 0x000D, r >= 0x001C && r <= 0x001F:
		return true
	case r == 0x00A0, r == 0x2007, r == 0x202F:
		return false
	}
	return unicode.In(r, unicode.Zs, unicode.Zl, unicode.Zp)
}

type ConditionIDCondition struct {
	id string
}

func newConditionIDCondition(id string) *ConditionIDCondition {
	return &ConditionIDCondition{id: id}
}

func (c *ConditionIDCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	if attRes == nil {
		return false
	}
	return c.id == attRes.GetID(e)
}

type ConditionLangCondition struct {
	lang string
}

func newConditionLangCondition(lang string) *ConditionLangCondition {
	return &ConditionLangCondition{lang: lang}
}

func (c *ConditionLangCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	if attRes == nil {
		return false
	}
	// GetLang returns "" for the Java null. No :lang() argument is empty, so
	// an empty lang attribute matches nothing either way.
	langAttribute := attRes.GetLang(e)
	return langAttribute != "" && c.MatchesString(langAttribute)
}

// MatchesString is the Java overload matches(String langAttribute).
func (c *ConditionLangCondition) MatchesString(langAttribute string) bool {
	if conditionEqualsIgnoreCase(c.lang, langAttribute) {
		return true
	}
	i := strings.IndexByte(langAttribute, '-')
	if i == -1 {
		return false
	}
	prefix := langAttribute[:i]
	// Java compares the UTF-16 index of '-' with the UTF-16 length of the
	// language.
	return len(utf16.Encode([]rune(prefix))) == len(utf16.Encode([]rune(c.lang))) &&
		conditionEqualsIgnoreCase(prefix, c.lang)
}

// conditionEqualsIgnoreCase is String.equalsIgnoreCase: the strings have the
// same number of UTF-16 code units and each pair of units is equal, or equal
// after Character.toUpperCase, or equal after Character.toLowerCase of that.
func conditionEqualsIgnoreCase(a string, b string) bool {
	ua := utf16.Encode([]rune(a))
	ub := utf16.Encode([]rune(b))
	if len(ua) != len(ub) {
		return false
	}
	for i := range ua {
		c1 := rune(ua[i])
		c2 := rune(ub[i])
		if c1 == c2 {
			continue
		}
		u1 := unicode.ToUpper(c1)
		u2 := unicode.ToUpper(c2)
		if u1 == u2 {
			continue
		}
		if unicode.ToLower(u1) == unicode.ToLower(u2) {
			continue
		}
		return false
	}
	return true
}

type ConditionFirstChildCondition struct{}

func (c *ConditionFirstChildCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	return treeRes.IsFirstChildElement(e)
}

type ConditionLastChildCondition struct{}

func (c *ConditionLastChildCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	return treeRes.IsLastChildElement(e)
}

// ConditionNthChildCondition is <An+B> from
// https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-child
// Represents elements whose numeric position in a series of siblings matches
// the pattern An+B, for every non-negative integer n. The index of the first
// element is 1. The values A and B must both be integers.
type ConditionNthChildCondition struct {
	a int
	b int
}

// conditionNthChildConditionPattern is matched against the whole string, as
// java.util.regex.Matcher.matches does. The character class is what \s means
// in Java; Go's \s leaves out U+000B.
var conditionNthChildConditionPattern = regexp.MustCompile(`^(?:([-+]?)(\d*)n([ \t\n\x0B\f\r]*([-+])[ \t\n\x0B\f\r]*(\d+))?)$`)

func NewConditionNthChildCondition(a int, b int) *ConditionNthChildCondition {
	return &ConditionNthChildCondition{a: a, b: b}
}

func (c *ConditionNthChildCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	// getPositionOfElement() starts at 0, CSS spec starts at 1
	position := treeRes.GetPositionOfElement(e) + 1
	return c.MatchesInt(position)
}

// MatchesInt is the Java overload matches(int position).
func (c *ConditionNthChildCondition) MatchesInt(position int) bool {
	switch c.a {
	case 0:
		return position == c.b
	default:
		an := position - c.b
		return an%c.a == 0 &&
			(an == 0 || (an > 0) == (c.a > 0)) // effectively same as "an / a >= 0"
	}
}

// conditionParseInt is Integer.parseInt for the inputs that reach it here:
// an optional sign and decimal digits, within the range of a Java int.
func conditionParseInt(s string) (int, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	return int(n), err
}

func ConditionNthChildConditionFromString(number string) *ConditionNthChildCondition {
	// String.trim removes the characters up to U+0020 from both ends.
	number = strings.ToLower(strings.TrimFunc(number, func(r rune) bool { return r <= ' ' }))

	switch number {
	case "even":
		return NewConditionNthChildCondition(2, 0)
	case "odd":
		return NewConditionNthChildCondition(2, 1)
	default:
		n, err := conditionParseInt(number)
		if err == nil {
			return NewConditionNthChildCondition(0, n)
		}
		m := conditionNthChildConditionPattern.FindStringSubmatch(number)

		if m == nil {
			panic(NewCSSParseExceptionWithCause("Invalid nth-child selector: "+number, -1, err))
		}
		a := 1
		if m[2] != "" {
			a = conditionMustParseInt(m[2])
		}
		// Group 5 is one or more digits, so it is empty only when the
		// optional group did not take part in the match (null in Java).
		b := 0
		if m[5] != "" {
			b = conditionMustParseInt(m[5])
		}
		if m[1] == "-" {
			a *= -1
		}
		if m[4] == "-" {
			b *= -1
		}

		return NewConditionNthChildCondition(a, b)
	}
}

// conditionMustParseInt parses digits that the pattern matched. A number
// outside the int range is a NumberFormatException in Java, which nothing
// catches as a parse error.
func conditionMustParseInt(s string) int {
	n, err := conditionParseInt(s)
	if err != nil {
		panic(NewXRRuntimeException("For input string: \"" + s + "\""))
	}
	return n
}

type ConditionEvenChildCondition struct{}

func (c *ConditionEvenChildCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	position := treeRes.GetPositionOfElement(e)
	return position >= 0 && position%2 == 0
}

type ConditionOddChildCondition struct{}

func (c *ConditionOddChildCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	position := treeRes.GetPositionOfElement(e)
	return position%2 == 1
}

type ConditionLinkCondition struct{}

func (c *ConditionLinkCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	return attRes.IsLink(e)
}

// ConditionUnsupportedCondition represents unsupported (or invalid) css,
// never matches
type ConditionUnsupportedCondition struct{}

func (c *ConditionUnsupportedCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	return false
}

type ConditionHasCondition struct {
	relativeSelectors []*SelectorHasRelativeSelector
}

func newConditionHasCondition(relativeSelectors []*SelectorHasRelativeSelector) *ConditionHasCondition {
	return &ConditionHasCondition{relativeSelectors: relativeSelectors}
}

func (c *ConditionHasCondition) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	for _, relativeSelector := range c.relativeSelectors {
		if c.matchesRelativeSelector(e, attRes, treeRes, relativeSelector) {
			return true
		}
	}
	return false
}

func (c *ConditionHasCondition) matchesRelativeSelector(scope dom.Node, attRes AttributeResolver, treeRes TreeResolver, relativeSelector *SelectorHasRelativeSelector) bool {
	axes := relativeSelector.Axes()
	selectors := relativeSelector.Selectors()
	firstAxis := axes[0]
	for _, candidate := range c.findCandidates(scope, firstAxis) {
		if c.matchesRelativeSelectorAt(scope, candidate, len(selectors)-1, selectors, axes, attRes, treeRes) {
			return true
		}
	}
	return false
}

func (c *ConditionHasCondition) matchesRelativeSelectorAt(
	scope dom.Node,
	candidate dom.Node,
	selectorIndex int,
	selectors []*Selector,
	axes []SelectorAxis,
	attRes AttributeResolver,
	treeRes TreeResolver) bool {
	selector := selectors[selectorIndex]
	if !selector.Matches(candidate, attRes, treeRes) || !selector.MatchesDynamic(candidate, attRes, treeRes) {
		return false
	}

	relation := axes[selectorIndex]
	if selectorIndex == 0 {
		return c.matchesScopeRelation(scope, candidate, relation, treeRes)
	}

	switch relation {
	case SelectorAxisChildAxis:
		parent := treeRes.GetParentElement(candidate)
		return parent != nil && c.matchesRelativeSelectorAt(scope, parent, selectorIndex-1, selectors, axes, attRes, treeRes)
	case SelectorAxisImmediateSiblingAxis:
		previous := treeRes.GetPreviousSiblingElement(candidate)
		return previous != nil && c.matchesRelativeSelectorAt(scope, previous, selectorIndex-1, selectors, axes, attRes, treeRes)
	case SelectorAxisDescendantAxis:
		ancestor := treeRes.GetParentElement(candidate)
		matched := false
		for ancestor != nil {
			if c.matchesRelativeSelectorAt(scope, ancestor, selectorIndex-1, selectors, axes, attRes, treeRes) {
				matched = true
				break
			}
			ancestor = treeRes.GetParentElement(ancestor)
		}
		return matched
	}
	panic(NewXRRuntimeException("unknown Selector.Axis"))
}

func (c *ConditionHasCondition) matchesScopeRelation(scope dom.Node, candidate dom.Node, relation SelectorAxis, treeRes TreeResolver) bool {
	switch relation {
	case SelectorAxisChildAxis:
		return scope == treeRes.GetParentElement(candidate)
	case SelectorAxisImmediateSiblingAxis:
		return scope == treeRes.GetPreviousSiblingElement(candidate)
	case SelectorAxisDescendantAxis:
		return c.isDescendantOf(candidate, scope, treeRes)
	}
	panic(NewXRRuntimeException("unknown Selector.Axis"))
}

func (c *ConditionHasCondition) isDescendantOf(candidate dom.Node, scope dom.Node, treeRes TreeResolver) bool {
	parent := treeRes.GetParentElement(candidate)
	for parent != nil {
		if parent == scope {
			return true
		}
		parent = treeRes.GetParentElement(parent)
	}
	return false
}

func (c *ConditionHasCondition) findCandidates(scope dom.Node, firstAxis SelectorAxis) []dom.Node {
	switch firstAxis {
	case SelectorAxisChildAxis:
		return c.findChildElements(scope)
	case SelectorAxisDescendantAxis:
		return c.findDescendantElements(scope)
	case SelectorAxisImmediateSiblingAxis:
		return c.findNextSiblingElement(scope)
	}
	panic(NewXRRuntimeException("unknown Selector.Axis"))
}

func (c *ConditionHasCondition) findChildElements(scope dom.Node) []dom.Node {
	var result []dom.Node
	child := scope.GetFirstChild()
	for child != nil {
		if child.GetNodeType() == dom.ElementNode {
			result = append(result, child)
		}
		child = child.GetNextSibling()
	}
	return result
}

func (c *ConditionHasCondition) findDescendantElements(scope dom.Node) []dom.Node {
	var result []dom.Node
	c.collectDescendantElements(scope, &result)
	return result
}

func (c *ConditionHasCondition) collectDescendantElements(node dom.Node, acc *[]dom.Node) {
	child := node.GetFirstChild()
	for child != nil {
		if child.GetNodeType() == dom.ElementNode {
			*acc = append(*acc, child)
			c.collectDescendantElements(child, acc)
		}
		child = child.GetNextSibling()
	}
}

func (c *ConditionHasCondition) findNextSiblingElement(scope dom.Node) []dom.Node {
	sibling := scope.GetNextSibling()
	for sibling != nil && sibling.GetNodeType() != dom.ElementNode {
		sibling = sibling.GetNextSibling()
	}
	if sibling == nil {
		return nil
	}
	return []dom.Node{sibling}
}

// conditionSplit splits s on ch and leaves out the empty parts.
func conditionSplit(s string, ch byte) []string {
	if strings.IndexByte(s, ch) == -1 {
		return []string{s}
	} else {
		var result []string

		last := 0

		for {
			next := strings.IndexByte(s[last:], ch)
			if next == -1 {
				break
			}
			next += last
			if next != last {
				result = append(result, s[last:next])
			}
			last = next + 1
		}

		if last != len(s) {
			result = append(result, s[last:])
		}

		return result
	}
}
