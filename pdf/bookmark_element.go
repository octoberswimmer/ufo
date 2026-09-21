// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/BookmarkElement.java

package pdf

import (
	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/geom"
)

type BookmarkElement struct {
	location *geom.Point
	// anchorName is nil when the bookmark element has no name attribute.
	anchorName *string
}

func (e *BookmarkElement) GetIntrinsicWidth() int {
	return 0
}

func (e *BookmarkElement) GetIntrinsicHeight() int {
	return 0
}

func (e *BookmarkElement) GetLocation() *geom.Point {
	return e.location
}

func (e *BookmarkElement) SetLocation(x int, y int) {
	e.location = geom.NewPoint(x, y)
}

// Detach removes the box id registered under the anchor name. Java passes a
// null anchor name to a HashMap remove, which removes nothing; here a nil
// anchor name skips the call.
func (e *BookmarkElement) Detach(c *ufo.LayoutContext) {
	if e.anchorName == nil {
		return
	}
	c.RemoveBoxId(*e.anchorName)
}

// GetAnchorName returns nil when the element has no anchor name.
func (e *BookmarkElement) GetAnchorName() *string {
	return e.anchorName
}

func NewBookmarkElement(anchorName *string) *BookmarkElement {
	return &BookmarkElement{location: geom.NewPoint(0, 0), anchorName: anchorName}
}

func (e *BookmarkElement) IsRequiresInteractivePaint() bool {
	// N/A
	return false
}

func (e *BookmarkElement) Paint(c *ufo.RenderingContext, outputDevice *ITextOutputDevice, box ufo.BlockBoxI) {
}

func (e *BookmarkElement) GetBaseline() int {
	return 0
}

func (e *BookmarkElement) HasBaseline() bool {
	return false
}
