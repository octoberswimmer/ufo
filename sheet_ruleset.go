// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/Ruleset.java

package ufo

import (
	"fmt"
	"strings"
)

type Ruleset struct {
	origin      StylesheetInfoOrigin
	props       []*PropertyDeclaration
	fsSelectors []*Selector
}

func NewRuleset(orig StylesheetInfoOrigin) *Ruleset {
	return &Ruleset{origin: orig}
}

// GetPropertyDeclarations returns the PropertyDeclarations of this rule set.
// Java returns an unmodifiable view; callers must not modify the slice.
func (r *Ruleset) GetPropertyDeclarations() []*PropertyDeclaration {
	return r.props
}

func (r *Ruleset) AddProperty(decl *PropertyDeclaration) {
	r.props = append(r.props, decl)
}

func (r *Ruleset) AddAllProperties(props []*PropertyDeclaration) {
	r.props = append(r.props, props...)
}

func (r *Ruleset) AddFSSelector(selector *Selector) {
	r.fsSelectors = append(r.fsSelectors, selector)
}

func (r *Ruleset) GetFSSelectors() []*Selector {
	return r.fsSelectors
}

func (r *Ruleset) GetOrigin() StylesheetInfoOrigin {
	return r.origin
}

func (r *Ruleset) String() string {
	props := make([]string, len(r.props))
	for i, prop := range r.props {
		props[i] = prop.String()
	}
	selectors := make([]string, len(r.fsSelectors))
	for i, selector := range r.fsSelectors {
		selectors[i] = selector.String()
	}
	// The lists print as java.util.List.toString does: "[a, b]".
	return fmt.Sprintf("%s{%s [%s] [%s]}", "Ruleset", r.origin, strings.Join(props, ", "), strings.Join(selectors, ", "))
}

func (r *Ruleset) ToString() string {
	return r.String()
}
