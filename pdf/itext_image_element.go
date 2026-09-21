// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextImageElement.java

package pdf

import (
	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/geom"
)

type ITextImageElement struct {
	image ufo.FSImage

	location *geom.Point
}

func NewITextImageElement(image ufo.FSImage) *ITextImageElement {
	return &ITextImageElement{image: image, location: geom.NewPoint(0, 0)}
}

func (e *ITextImageElement) GetIntrinsicWidth() int {
	return e.image.GetWidth()
}

func (e *ITextImageElement) GetIntrinsicHeight() int {
	return e.image.GetHeight()
}

func (e *ITextImageElement) GetLocation() *geom.Point {
	return e.location
}

func (e *ITextImageElement) SetLocation(x int, y int) {
	e.location = geom.NewPoint(x, y)
}

func (e *ITextImageElement) GetImage() ufo.FSImage {
	return e.image
}

func (e *ITextImageElement) Detach(c *ufo.LayoutContext) {
}

func (e *ITextImageElement) IsRequiresInteractivePaint() bool {
	// N/A
	return false
}

func (e *ITextImageElement) Paint(c *ufo.RenderingContext, outputDevice *ITextOutputDevice, box ufo.BlockBoxI) {
	contentBounds := box.GetContentAreaEdge(box.GetAbsX(), box.GetAbsY(), c)
	element := box.GetReplacedElement()
	outputDevice.DrawImage(
		element.(*ITextImageElement).GetImage(),
		contentBounds.X, contentBounds.Y)
}

func (e *ITextImageElement) GetBaseline() int {
	return 0
}

func (e *ITextImageElement) HasBaseline() bool {
	return false
}
