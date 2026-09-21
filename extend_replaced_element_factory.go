// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/ReplacedElementFactory.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// ReplacedElementFactory creates the replaced elements of a document.
//
// setFormSubmissionListener(FormSubmissionListener) is not ported:
// FormSubmissionListener belongs to the Swing form support, and the PDF
// factory implements the method as a no-op.
type ReplacedElementFactory interface {
	// CreateReplacedElement returns the ReplacedElement or nil if no
	// ReplacedElement applies.
	//
	// NOTE: Only block equivalent elements can be replaced.
	//
	// cssWidth is the CSS width of the element in dots (or -1 if width is
	// auto). cssHeight is the CSS height of the element in dots (or -1 if the
	// height should be treated as auto).
	CreateReplacedElement(
		c *LayoutContext, box BlockBoxI,
		uac UserAgentCallback, cssWidth int, cssHeight int) ReplacedElement

	// Reset instructs the ReplacedElementFactory to discard any cached data
	// (typically because a new page is about to be loaded).
	Reset()

	// Remove removes any reference to Element e.
	Remove(e *dom.Element)
}
