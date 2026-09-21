// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/ReplacedElement.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// ReplacedElement is an XML element in the document being rendered whose
// visual output is delegated. For example, an <img> element in HTML may be
// rendered using some form of image. The idea is that there are some XML
// elements which Flying Saucer knows how to position and size (that's in the
// CSS) but has no idea how to render on screen. Replaced elements serve that
// purpose.
type ReplacedElement interface {
	GetIntrinsicWidth() int

	GetIntrinsicHeight() int

	// GetLocation returns the current location where the element will be
	// rendered on the canvas.
	GetLocation() *geom.Point

	// SetLocation assigns the new locations where the element will be
	// rendered. x is the new horizontal position, y the new vertical
	// position.
	SetLocation(x int, y int)

	Detach(c *LayoutContext)

	IsRequiresInteractivePaint() bool

	HasBaseline() bool

	GetBaseline() int
}
