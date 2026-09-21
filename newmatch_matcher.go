// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/Matcher.java

package ufo

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/octoberswimmer/ufo/dom"
)

type Matcher struct {
	docMapper    *MatcherMapper
	attRes       AttributeResolver
	treeRes      TreeResolver
	styleFactory StylesheetFactory

	// lock is the Java monitor "lock". Java monitors are reentrant and the
	// private methods take the monitor again; sync.Mutex is not reentrant, so
	// only the exported entry points GetCascadedStyle and GetPECascadedStyle
	// lock it. Every private method that synchronizes in Java (matchElement,
	// getElementStyle, getNonCssStyle, MatcherMapper.GetCascadedStyle) is
	// reached only from those two.
	lock sync.Mutex
	// matcherMap is the Java field _map. It is read and written only while
	// lock is held.
	matcherMap map[dom.Node]*MatcherMapper

	//handle dynamic
	// hoverElements is a synchronizedSet in Java, and IsHoverStyled reads it
	// without holding lock, so it has a mutex of its own.
	hoverElementsLock sync.Mutex
	hoverElements     map[dom.Node]struct{}

	pageRules     []*PageRule
	fontFaceRules []*FontFaceRule
}

func NewMatcher(tr TreeResolver, ar AttributeResolver,
	factory StylesheetFactory, stylesheets []*Stylesheet, medium string) *Matcher {
	m := &Matcher{
		treeRes:       tr,
		attRes:        ar,
		styleFactory:  factory,
		matcherMap:    make(map[dom.Node]*MatcherMapper),
		hoverElements: make(map[dom.Node]struct{}),
	}
	m.docMapper = m.createDocumentMapper(stylesheets, medium)
	return m
}

func (m *Matcher) GetCascadedStyle(e *dom.Element, restyle bool) *CascadedStyle {
	m.lock.Lock()
	defer m.lock.Unlock()
	var em *MatcherMapper
	if restyle {
		em = m.matchElement(e)
	} else {
		em = m.getMapper(e)
	}
	return em.GetCascadedStyle(e)
}

// GetPECascadedStyle may return nil.
// We assume that restyle has already been done by a GetCascadedStyle if necessary.
func (m *Matcher) GetPECascadedStyle(e *dom.Element, pseudoElement string) *CascadedStyle {
	m.lock.Lock()
	defer m.lock.Unlock()
	em := m.getMapper(e)
	return em.GetPECascadedStyle(pseudoElement)
}

// GetPageCascadedStyle returns the style of a page. pageName is "" for a page
// without a name and pseudoPage is "" for no pseudo page.
func (m *Matcher) GetPageCascadedStyle(pageName string, pseudoPage string) *PageInfo {
	var props []*PropertyDeclaration
	marginBoxes := make(map[*MarginBoxName][]*PropertyDeclaration)

	for _, pageRule := range m.pageRules {
		if pageRule.Applies(pageName, pseudoPage) {
			props = append(props, pageRule.GetRuleset().GetPropertyDeclarations()...)
			// Map.putAll: each key is written once per rule, so the order
			// in which the rule's map is read does not change the result.
			for marginBoxName, declarations := range pageRule.GetMarginBoxes() {
				marginBoxes[marginBoxName] = declarations
			}
		}
	}

	var style *CascadedStyle
	if len(props) == 0 {
		style = CascadedStyleEmptyCascadedStyle
	} else {
		style = NewCascadedStyle(props)
	}

	return NewPageInfo(props, style, marginBoxes)
}

func (m *Matcher) GetFontFaceRules() []*FontFaceRule {
	return m.fontFaceRules
}

func (m *Matcher) IsHoverStyled(e dom.Node) bool {
	m.hoverElementsLock.Lock()
	defer m.hoverElementsLock.Unlock()
	_, ok := m.hoverElements[e]
	return ok
}

func (m *Matcher) matchElement(e dom.Node) *MatcherMapper {
	parent := m.treeRes.GetParentElement(e)
	if parent != nil {
		return m.getMapper(parent).MapChild(e)
	} else { // has to be a document or a fragment node
		return m.docMapper.MapChild(e)
	}
}

func (m *Matcher) createDocumentMapper(stylesheets []*Stylesheet, medium string) *MatcherMapper {
	// sorter is a TreeMap<String, Selector> in Java, keyed by
	// Selector.getOrder(): a later selector with the same key replaces the
	// earlier one, and the values are read in ascending key order. The Go map
	// gives the replacement; sorting the keys gives the order. The keys hold
	// only ASCII characters, for which Go's byte order equals the UTF-16
	// order of String.compareTo.
	sorter := make(map[string]*Selector)
	m.addAllStylesheets(stylesheets, sorter, medium)
	XRLogMatch("Matcher created with " + strconv.Itoa(len(sorter)) + " selectors")
	keys := make([]string, 0, len(sorter))
	for key := range sorter {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	selectors := make([]*Selector, 0, len(keys))
	for _, key := range keys {
		selectors = append(selectors, sorter[key])
	}
	return newMatcherMapper(m, selectors)
}

func (m *Matcher) addAllStylesheets(stylesheets []*Stylesheet, sorter map[string]*Selector, medium string) {
	count := 0
	pCount := 0
	for _, stylesheet := range stylesheets {
		for _, obj := range stylesheet.GetContents() {
			switch content := obj.(type) {
			case *Ruleset:
				for _, selector := range content.GetFSSelectors() {
					count++
					selector.SetPos(count)
					sorter[selector.GetOrder()] = selector
				}
			case *PageRule:
				pCount++
				content.SetPos(pCount)
				m.pageRules = append(m.pageRules, content)
			case *MediaRule:
				if content.Matches(medium) {
					for _, ruleset := range content.GetContents() {
						for _, selector := range ruleset.GetFSSelectors() {
							count++
							selector.SetPos(count)
							sorter[selector.GetOrder()] = selector
						}
					}
				}
			}
		}

		m.fontFaceRules = append(m.fontFaceRules, stylesheet.GetFontFaceRules()...)
	}

	// List.sort is stable: rules with the same order keep their positions.
	sort.SliceStable(m.pageRules, func(i, j int) bool {
		return m.pageRules[i].GetOrder() < m.pageRules[j].GetOrder()
	})
}

func (m *Matcher) link(e dom.Node, mapper *MatcherMapper) {
	m.matcherMap[e] = mapper
}

func (m *Matcher) getMapper(e dom.Node) *MatcherMapper {
	if mapper := m.matcherMap[e]; mapper != nil {
		return mapper
	}
	return m.matchElement(e)
}

func (m *Matcher) getElementStyle(e dom.Node) *Ruleset {
	if m.attRes == nil || m.styleFactory == nil {
		return nil
	}

	style := m.attRes.GetElementStyling(e)
	if UtilIsNullOrEmpty(style) {
		return nil
	}

	return m.styleFactory.ParseStyleDeclaration(StylesheetInfoOriginAuthor, style)
}

func (m *Matcher) getNonCssStyle(e dom.Node) *Ruleset {
	if m.attRes == nil || m.styleFactory == nil {
		return nil
	}
	style := m.attRes.GetNonCssStyling(e)
	if UtilIsNullOrEmpty(style) {
		return nil
	}
	return m.styleFactory.ParseStyleDeclaration(StylesheetInfoOriginAuthor, style)
}

// MatcherMapper represents a local CSS for a Node that is used to match the
// Node's children. It is the Java inner class Matcher.Mapper; matcher is the
// enclosing instance.
type MatcherMapper struct {
	matcher         *Matcher
	axes            []*Selector
	pseudoSelectors map[string][]*Selector
	mappedSelectors []*Selector
	// children is keyed by the IDs of the selectors that matched the child,
	// in match order. Java uses the List<Integer> itself as the key; the Go
	// key is the IDs in decimal joined with ",", which is equal for two lists
	// exactly when the lists are equal.
	children map[string]*MatcherMapper
}

func newMatcherMapper(matcher *Matcher, selectors []*Selector) *MatcherMapper {
	return newMatcherMapperWithPseudoSelectors(matcher, append([]*Selector(nil), selectors...), nil, nil)
}

func newMatcherMapperWithPseudoSelectors(matcher *Matcher, childAxes []*Selector, pseudoSelectors map[string][]*Selector, mappedSelectors []*Selector) *MatcherMapper {
	return &MatcherMapper{
		matcher:         matcher,
		axes:            childAxes,
		pseudoSelectors: pseudoSelectors,
		mappedSelectors: mappedSelectors,
	}
}

// MapChild has a side effect: it creates and stores a Mapper for the element.
//
// The mapper it returns holds the selectors that matched, sorted according to
// specificity (more correct: preserves the sort order from Matcher creation)
func (mp *MatcherMapper) MapChild(e dom.Node) *MatcherMapper {
	m := mp.matcher
	childAxes := make([]*Selector, 0, len(mp.axes)+10)
	pseudoSelectors := make(map[string][]*Selector)
	var mappedSelectors []*Selector
	var key strings.Builder
	addToKey := func(selectorID int) {
		key.WriteString(strconv.Itoa(selectorID))
		key.WriteByte(',')
	}
	for _, axe := range mp.axes {
		switch axe.GetAxis() {
		case SelectorAxisDescendantAxis:
			childAxes = append(childAxes, axe) // carry it forward to other descendants
		case SelectorAxisImmediateSiblingAxis:
			panic(NewXRRuntimeException("Selector axis: " + SelectorAxisImmediateSiblingAxis.String()))
		case SelectorAxisChildAxis:
		}
		if !axe.Matches(e, m.attRes, m.treeRes) {
			continue
		}
		//Assumption: if it is a pseudo-element, it does not also have dynamic pseudo-class
		pseudoElement := axe.GetPseudoElement()
		if pseudoElement != "" {
			pseudoSelectors[pseudoElement] = append(pseudoSelectors[pseudoElement], axe)
			addToKey(axe.GetSelectorID())
			continue
		}
		if axe.IsPseudoClass(SelectorHoverPseudoclass) {
			m.hoverElementsLock.Lock()
			m.hoverElements[e] = struct{}{}
			m.hoverElementsLock.Unlock()
		}
		if !axe.MatchesDynamic(e, m.attRes, m.treeRes) {
			continue
		}
		addToKey(axe.GetSelectorID())
		chain := axe.GetChainedSelector()
		if chain == nil {
			mappedSelectors = append(mappedSelectors, axe)
		} else {
			switch chain.GetAxis() {
			case SelectorAxisImmediateSiblingAxis:
				panic(NewXRRuntimeException("Selector axis: " + SelectorAxisImmediateSiblingAxis.String()))
			case SelectorAxisChildAxis,
				SelectorAxisDescendantAxis:
				childAxes = append(childAxes, chain)
			}
		}
	}
	if mp.children == nil {
		mp.children = make(map[string]*MatcherMapper)
	}
	childMapper := mp.children[key.String()]
	if childMapper == nil {
		childMapper = newMatcherMapperWithPseudoSelectors(m, childAxes, pseudoSelectors, mappedSelectors)
		mp.children[key.String()] = childMapper
	}
	m.link(e, childMapper)
	return childMapper
}

func (mp *MatcherMapper) GetCascadedStyle(e dom.Node) *CascadedStyle {
	m := mp.matcher
	elementStyling := m.getElementStyle(e)
	nonCssStyling := m.getNonCssStyle(e)
	var propList []*PropertyDeclaration
	//specificity 0,0,0,0
	if nonCssStyling != nil {
		propList = append(propList, nonCssStyling.GetPropertyDeclarations()...)
	}
	//these should have been returned in order of specificity
	for _, selector := range mp.mappedSelectors {
		propList = append(propList, selector.GetRuleset().GetPropertyDeclarations()...)
	}
	//specificity 1,0,0,0
	if elementStyling != nil {
		propList = append(propList, elementStyling.GetPropertyDeclarations()...)
	}
	if len(propList) == 0 {
		return CascadedStyleEmptyCascadedStyle
	}
	return NewCascadedStyle(propList)
}

// GetPECascadedStyle may return nil.
// We assume that restyle has already been done by a GetCascadedStyle if necessary.
func (mp *MatcherMapper) GetPECascadedStyle(pseudoElement string) *CascadedStyle {
	if len(mp.pseudoSelectors) == 0 {
		return nil
	}

	pe, ok := mp.pseudoSelectors[pseudoElement]
	if !ok {
		return nil
	}

	var propList []*PropertyDeclaration
	for _, selector := range pe {
		propList = append(propList, selector.GetRuleset().GetPropertyDeclarations()...)
	}

	if len(propList) == 0 {
		return CascadedStyleEmptyCascadedStyle // already internalized
	}
	return NewCascadedStyle(propList)
}
