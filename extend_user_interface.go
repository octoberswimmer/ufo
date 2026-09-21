// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/UserInterface.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

type UserInterface interface {
	// IsHover gets the hover attribute of the UserInterface object.
	IsHover(e *dom.Element) bool

	// IsActive gets the active attribute of the UserInterface object.
	IsActive(e *dom.Element) bool

	// IsFocus gets the focus attribute of the UserInterface object.
	IsFocus(e *dom.Element) bool
}
