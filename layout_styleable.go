// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/Styleable.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// Styleable is implemented by all objects appearing in the layout tree. It
// can roughly be thought of as a styled element (although an InlineLayoutBox
// may be split across many lines) and some Styleable objects may not
// define an element at all (e.g. anonymous inline boxes) and some
// Styleable objects don't correspond to a real element
// (e.g. :before and :after pseudo-elements)
type Styleable interface {
	// GetStyle may return nil.
	GetStyle() CalculatedStyleI

	SetStyle(style CalculatedStyleI)

	// GetElement may return nil.
	GetElement() *dom.Element

	SetElement(e *dom.Element)

	// GetPseudoElementOrClass returns "" where Java returns null.
	GetPseudoElementOrClass() string
}
