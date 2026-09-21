// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/PageRule.java

package ufo

// PageRule is an @page rule. name and pseudoPage are "" where Java has null:
// the parser never produces an empty page name or pseudo page.
type PageRule struct {
	name       string
	pseudoPage string
	ruleset    *Ruleset
	origin     StylesheetInfoOrigin

	marginBoxes map[*MarginBoxName][]*PropertyDeclaration

	pos int

	specificityF int
	specificityG int
	specificityH int
}

func NewPageRule(origin StylesheetInfoOrigin, name string, pseudoPage string, marginBoxes map[*MarginBoxName][]*PropertyDeclaration, ruleset *Ruleset) *PageRule {
	p := &PageRule{
		origin:      origin,
		name:        name,
		ruleset:     ruleset,
		pseudoPage:  pseudoPage,
		marginBoxes: make(map[*MarginBoxName][]*PropertyDeclaration, len(marginBoxes)),
	}
	if name != "" {
		p.specificityF = 1
	}
	if pseudoPage == "first" {
		p.specificityG = 1
		p.specificityH = 0
	} else {
		p.specificityG = 0
		p.specificityH = 1
	}
	for marginBoxName, declarations := range marginBoxes {
		p.marginBoxes[marginBoxName] = declarations
	}
	return p
}

func (p *PageRule) GetPseudoPage() string {
	return p.pseudoPage
}

func (p *PageRule) GetRuleset() *Ruleset {
	return p.ruleset
}

func (p *PageRule) AddContent(ruleset *Ruleset) {
	panic(NewXRRuntimeException("Ruleset has already been set"))
}

func (p *PageRule) GetOrigin() StylesheetInfoOrigin {
	return p.origin
}

func (p *PageRule) GetName() string {
	return p.name
}

func (p *PageRule) GetMarginBoxProperties(name *MarginBoxName) []*PropertyDeclaration {
	return p.marginBoxes[name]
}

func (p *PageRule) GetMarginBoxes() map[*MarginBoxName][]*PropertyDeclaration {
	return p.marginBoxes
}

func (p *PageRule) GetOrder() int64 {
	var result int64

	result |= int64(p.specificityF) << 32
	result |= int64(p.specificityG) << 24
	result |= int64(p.specificityH) << 16
	result |= int64(p.pos)

	return result
}

// Applies reports whether this rule applies to a page with the given name
// ("" for an unnamed page) and pseudo page ("" for none).
func (p *PageRule) Applies(pageName string, pseudoPage string) bool {
	if p.name == "" && p.pseudoPage == "" {
		return true
	} else if p.name == "" &&
		(p.pseudoPage == pseudoPage ||
			p.pseudoPage == "right" && pseudoPage == "first") { // assume first page is a right page
		return true
	} else if p.name != "" && p.name == pageName && p.pseudoPage == "" {
		return true
	} else {
		return p.name != "" && p.name == pageName && p.pseudoPage == pseudoPage
	}
}

func (p *PageRule) GetPos() int {
	return p.pos
}

func (p *PageRule) SetPos(pos int) {
	p.pos = pos
}
