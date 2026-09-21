// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/Selector.java

package ufo

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/octoberswimmer/ufo/dom"
)

// SelectorHasRelativeSelector is one relative selector of a :has() condition:
// the compound selectors from left to right and, for each, the axis that
// relates it to the selector before it (to the :has() scope for the first).
type SelectorHasRelativeSelector struct {
	axes      []SelectorAxis
	selectors []*Selector
}

func NewSelectorHasRelativeSelector(axes []SelectorAxis, selectors []*Selector) *SelectorHasRelativeSelector {
	if len(axes) == 0 || len(selectors) == 0 || len(axes) != len(selectors) {
		panic(NewXRRuntimeException("Axes/selectors must be non-empty and same size"))
	}
	return &SelectorHasRelativeSelector{axes: axes, selectors: selectors}
}

func (h *SelectorHasRelativeSelector) Axes() []SelectorAxis {
	return h.axes
}

func (h *SelectorHasRelativeSelector) Selectors() []*Selector {
	return h.selectors
}

// A Selector is really a chain of CSS selectors that all need to be valid for
// the selector to match.
type Selector struct {
	parent          *Ruleset
	chainedSelector *Selector
	siblingSelector *Selector
	axis            SelectorAxis
	// name is "" where Java has null: the selector matches an element of any
	// name.
	name string
	// text is nil until SetName is called with a name or AddClassCondition is
	// called.
	text *string
	// namespaceURI is nil for any namespace and points to
	// TreeResolverNoNamespace for no namespace.
	namespaceURI *string
	pc           int
	// pe is "" where Java has null: the selector has no pseudo-element.
	pe string

	//specificity - correct values are gotten from the last Selector in the chain
	specificityB int
	specificityC int
	specificityD int

	pos int //to distinguish between selectors of same specificity

	conditions []Condition

	// Give each a unique ID to be able to create a key to internalize Matcher.Mappers
	selectorID int
}

// SelectorAxis is the Java enum Selector.Axis.
type SelectorAxis int

const (
	SelectorAxisDescendantAxis SelectorAxis = iota
	SelectorAxisChildAxis
	SelectorAxisImmediateSiblingAxis
)

// String returns the name of the Java enum constant.
func (a SelectorAxis) String() string {
	switch a {
	case SelectorAxisDescendantAxis:
		return "DESCENDANT_AXIS"
	case SelectorAxisChildAxis:
		return "CHILD_AXIS"
	case SelectorAxisImmediateSiblingAxis:
		return "IMMEDIATE_SIBLING_AXIS"
	}
	panic(NewXRRuntimeException("unknown Selector.Axis"))
}

func (a SelectorAxis) ToString() string {
	return a.String()
}

const (
	SelectorVisitedPseudoclass = 2
	SelectorHoverPseudoclass   = 4
	SelectorActivePseudoclass  = 8
	SelectorFocusPseudoclass   = 16
)

// selectorCount is the Java static Selector.selectorCount. Java increments it
// without synchronization; the Go counter is atomic so that stylesheets can be
// parsed on several goroutines.
var selectorCount atomic.Int64

func NewSelector(ruleset *Ruleset) *Selector {
	return &Selector{
		parent:     ruleset,
		axis:       SelectorAxisDescendantAxis,
		selectorID: int(selectorCount.Add(1) - 1),
	}
}

// Matches checks if the given Element matches this selector. Note: the parser
// should give all class
func (s *Selector) Matches(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	if s.siblingSelector != nil {
		sib := s.siblingSelector.GetAppropriateSibling(e, treeRes)
		if sib == nil {
			return false
		}
		if !s.siblingSelector.Matches(sib, attRes, treeRes) {
			return false
		}
	}
	if s.name == "" || treeRes.MatchesElement(e, s.namespaceURI, s.name) {
		// all conditions need to be true
		for _, condition := range s.conditions {
			if !condition.Matches(e, attRes, treeRes) {
				return false
			}
		}
		return true
	}
	return false
}

// MatchesDynamic checks if the given Element matches this selector's dynamic
// properties. Note: the parser should give all class
func (s *Selector) MatchesDynamic(e dom.Node, attRes AttributeResolver, treeRes TreeResolver) bool {
	if s.siblingSelector != nil {
		sib := s.siblingSelector.GetAppropriateSibling(e, treeRes)
		if sib == nil {
			return false
		}
		if !s.siblingSelector.MatchesDynamic(sib, attRes, treeRes) {
			return false
		}
	}
	if s.IsPseudoClass(SelectorVisitedPseudoclass) {
		if attRes == nil || !attRes.IsVisited(e) {
			return false
		}
	}
	if s.IsPseudoClass(SelectorActivePseudoclass) {
		if attRes == nil || !attRes.IsActive(e) {
			return false
		}
	}
	if s.IsPseudoClass(SelectorHoverPseudoclass) {
		if attRes == nil || !attRes.IsHover(e) {
			return false
		}
	}
	if s.IsPseudoClass(SelectorFocusPseudoclass) {
		return attRes != nil && attRes.IsFocus(e)
	}
	return true
}

// AddUnsupportedCondition is for unsupported or invalid CSS
func (s *Selector) AddUnsupportedCondition() {
	s.addCondition(ConditionCreateUnsupportedCondition())
}

// AddLinkCondition adds the CSS condition that element has pseudo-class :link
func (s *Selector) AddLinkCondition() {
	s.specificityC++
	s.addCondition(ConditionCreateLinkCondition())
}

// AddFirstChildCondition adds the CSS condition that element has pseudo-class
// :first-child
func (s *Selector) AddFirstChildCondition() {
	s.specificityC++
	s.addCondition(ConditionCreateFirstChildCondition())
}

// AddLastChildCondition adds the CSS condition that element has pseudo-class
// :last-child
func (s *Selector) AddLastChildCondition() {
	s.specificityC++
	s.addCondition(ConditionCreateLastChildCondition())
}

// AddNthChildCondition adds the CSS condition that element has pseudo-class
// :nth-child(an+b)
func (s *Selector) AddNthChildCondition(number string) {
	s.specificityC++
	s.addCondition(ConditionCreateNthChildCondition(number))
}

// AddEvenChildCondition adds the CSS condition that element has pseudo-class
// :even
func (s *Selector) AddEvenChildCondition() {
	s.specificityC++
	s.addCondition(ConditionCreateEvenChildCondition())
}

// AddOddChildCondition adds the CSS condition that element has pseudo-class
// :odd
func (s *Selector) AddOddChildCondition() {
	s.specificityC++
	s.addCondition(ConditionCreateOddChildCondition())
}

// AddLangCondition adds the CSS condition :lang(Xx)
func (s *Selector) AddLangCondition(lang string) {
	s.specificityC++
	s.addCondition(ConditionCreateLangCondition(lang))
}

func (s *Selector) AddHasCondition(relativeSelectors []*SelectorHasRelativeSelector, specificityB int, specificityC int, specificityD int) {
	s.specificityB += specificityB
	s.specificityC += specificityC
	s.specificityD += specificityD
	s.addCondition(ConditionCreateHasCondition(relativeSelectors))
}

// AddIDCondition adds the CSS condition #ID
func (s *Selector) AddIDCondition(id string) {
	s.specificityB++
	s.addCondition(ConditionCreateIDCondition(id))
}

// AddClassCondition adds the CSS condition .class
func (s *Selector) AddClassCondition(className string) {
	s.specificityC++
	s.addCondition(ConditionCreateClassCondition(className))
	name := s.name
	if name == "" {
		// Java concatenates the null name.
		name = "null"
	}
	text := name + TokenTkPeriod.GetExternalName() + className
	s.text = &text
}

// AddAttributeExistsCondition adds the CSS condition [attribute]
func (s *Selector) AddAttributeExistsCondition(namespaceURI *string, name string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributeExistsCondition(namespaceURI, name))
}

// AddAttributeEqualsCondition adds the CSS condition [attribute=value]
func (s *Selector) AddAttributeEqualsCondition(namespaceURI *string, name string, value string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributeEqualsCondition(namespaceURI, name, value))
}

// AddAttributePrefixCondition adds the CSS condition [attribute^=value]
func (s *Selector) AddAttributePrefixCondition(namespaceURI *string, name string, value string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributePrefixCondition(namespaceURI, name, value))
}

// AddAttributeSuffixCondition adds the CSS condition [attribute$=value]
func (s *Selector) AddAttributeSuffixCondition(namespaceURI *string, name string, value string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributeSuffixCondition(namespaceURI, name, value))
}

// AddAttributeSubstringCondition adds the CSS condition [attribute*=value]
func (s *Selector) AddAttributeSubstringCondition(namespaceURI *string, name string, value string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributeSubstringCondition(namespaceURI, name, value))
}

// AddAttributeMatchesListCondition adds the CSS condition [attribute~=value]
func (s *Selector) AddAttributeMatchesListCondition(namespaceURI *string, name string, value string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributeMatchesListCondition(namespaceURI, name, value))
}

// AddAttributeMatchesFirstPartCondition adds the CSS condition
// [attribute|=value]
func (s *Selector) AddAttributeMatchesFirstPartCondition(namespaceURI *string, name string, value string) {
	s.specificityC++
	s.addCondition(ConditionCreateAttributeMatchesFirstPartCondition(namespaceURI, name, value))
}

// SetPseudoClass sets which pseudo-classes must apply for this selector.
//
// For pc the Selector...Pseudoclass values should be used. Once set they
// cannot be unset. Note that the pseudo-classes should be set one at a time,
// otherwise specificity of declaration becomes wrong.
func (s *Selector) SetPseudoClass(pc int) {
	if !s.IsPseudoClass(pc) {
		s.specificityC++
	}
	s.pc |= pc
}

func (s *Selector) SetPseudoElement(pseudoElement string) {
	if s.pe != "" {
		s.AddUnsupportedCondition()
		XRLogMatchWithLevel(LevelWarning, "Trying to set more than one pseudo-element")
	} else {
		s.specificityD++
		s.pe = pseudoElement
	}
}

// IsPseudoClass queries if a pseudo-class must apply for this selector.
//
// For pc the Selector...Pseudoclass values should be used.
func (s *Selector) IsPseudoClass(pc int) bool {
	return (s.pc & pc) != 0
}

// GetPseudoElement gets the pseudoElement attribute of the Selector object,
// "" when the selector has none.
func (s *Selector) GetPseudoElement() string {
	return s.pe
}

// GetChainedSelector gets the next selector in the chain, for matching
// against elements along the appropriate axis
func (s *Selector) GetChainedSelector() *Selector {
	return s.chainedSelector
}

// GetRuleset gets the Ruleset that this Selector is part of
func (s *Selector) GetRuleset() *Ruleset {
	return s.parent
}

// GetAxis gets the axis that this selector should be evaluated on
func (s *Selector) GetAxis() SelectorAxis {
	return s.axis
}

// GetSpecificityB is the correct specificity value for this selector and its
// sibling-axis selectors
func (s *Selector) GetSpecificityB() int {
	return s.specificityB
}

// GetSpecificityD is the correct specificity value for this selector and its
// sibling-axis selectors
func (s *Selector) GetSpecificityD() int {
	return s.specificityD
}

// GetSpecificityC is the correct specificity value for this selector and its
// sibling-axis selectors
func (s *Selector) GetSpecificityC() int {
	return s.specificityC
}

// GetOrder returns "a number in a large base" with specificity and
// specification order of selector
func (s *Selector) GetOrder() string {
	if s.chainedSelector != nil {
		return s.chainedSelector.GetOrder()
	} //only "deepest" value is correct
	b := "000" + strconv.Itoa(s.GetSpecificityB())
	c := "000" + strconv.Itoa(s.GetSpecificityC())
	d := "000" + strconv.Itoa(s.GetSpecificityD())
	p := "00000" + strconv.Itoa(s.pos)
	return "0" + b[len(b)-3:] + c[len(c)-3:] + d[len(d)-3:] + p[len(p)-5:]
}

// GetAppropriateSibling gets the appropriateSibling attribute of the Selector
// object
func (s *Selector) GetAppropriateSibling(e dom.Node, treeRes TreeResolver) dom.Node {
	switch s.axis {
	case SelectorAxisImmediateSiblingAxis:
		return treeRes.GetPreviousSiblingElement(e)
	case SelectorAxisDescendantAxis, SelectorAxisChildAxis:
		XRLogException("Bad sibling axis")
		return nil
	}
	panic(NewXRRuntimeException("unknown Selector.Axis"))
}

// addCondition adds a feature to the Condition attribute of the Selector
// object
func (s *Selector) addCondition(c Condition) {
	if s.pe != "" {
		s.conditions = append(s.conditions, ConditionCreateUnsupportedCondition())
		XRLogMatchWithLevel(LevelWarning, "Trying to append conditions to pseudoElement "+s.pe)
	}
	s.conditions = append(s.conditions, c)
}

func (s *Selector) GetSelectorID() int {
	return s.selectorID
}

// SetName sets the element name the selector matches; "" is the Java null,
// which the parser passes for the universal selector. Either way the call
// counts towards specificity d.
func (s *Selector) SetName(name string) {
	s.name = name
	if name == "" {
		s.text = nil
	} else {
		s.text = &name
	}
	s.specificityD++
}

func (s *Selector) GetSelectorText() string {
	text := "null"
	if s.text != nil {
		text = *s.text
	}
	if s.chainedSelector != nil {
		return text + " " + s.chainedSelector.GetSelectorText()
	}
	return text
}

func (s *Selector) SetPos(pos int) {
	s.pos = pos
	if s.siblingSelector != nil {
		s.siblingSelector.SetPos(pos)
	}
	if s.chainedSelector != nil {
		s.chainedSelector.SetPos(pos)
	}
}

func (s *Selector) SetAxis(axis SelectorAxis) {
	s.axis = axis
}

func (s *Selector) SetSpecificityB(b int) {
	s.specificityB = b
}

func (s *Selector) SetSpecificityC(c int) {
	s.specificityC = c
}

func (s *Selector) SetSpecificityD(d int) {
	s.specificityD = d
}

func (s *Selector) SetChainedSelector(selector *Selector) {
	s.chainedSelector = selector
}

func (s *Selector) SetSiblingSelector(selector *Selector) {
	s.siblingSelector = selector
}

func (s *Selector) SetNamespaceURI(namespaceURI *string) {
	s.namespaceURI = namespaceURI
}

func (s *Selector) String() string {
	name := s.name
	if name == "" {
		name = "null"
	}
	return fmt.Sprintf("%s{%s}", "Selector", name)
}

func (s *Selector) ToString() string {
	return s.String()
}
