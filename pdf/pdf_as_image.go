// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/PDFAsImage.java

package pdf

import (
	"net/url"

	"github.com/octoberswimmer/ufo"
)

// PDFAsImage is a page of another PDF file used as an image. The class is
// ported in full; drawing one fails in ITextOutputDevice.DrawImage, because
// pdf/writer cannot import PDF pages (writer.NewPdfReader returns a
// *writer.UnsupportedFeatureError).
type PDFAsImage struct {
	source *url.URL

	width          float32
	height         float32
	unscaledWidth  float32
	unscaledHeight float32
}

func NewPDFAsImage(source *url.URL, width float32, height float32) *PDFAsImage {
	return &PDFAsImage{
		source:         source,
		width:          width,
		unscaledWidth:  width,
		height:         height,
		unscaledHeight: height,
	}
}

func newPDFAsImageWithUnscaledWidthUnscaledHeight(source *url.URL, unscaledWidth float32, unscaledHeight float32, width float32, height float32) *PDFAsImage {
	return &PDFAsImage{
		source:         source,
		width:          width,
		unscaledWidth:  unscaledWidth,
		height:         height,
		unscaledHeight: unscaledHeight,
	}
}

func (p *PDFAsImage) GetWidth() int {
	return int(p.width)
}

func (p *PDFAsImage) GetHeight() int {
	return int(p.height)
}

func (p *PDFAsImage) Scale(width int, height int) ufo.FSImage {
	targetWidth := float32(width)
	targetHeight := float32(height)

	if width == -1 {
		targetWidth = p.GetWidthAsFloat() * (targetHeight / float32(p.GetHeight()))
	}

	if height == -1 {
		targetHeight = p.GetHeightAsFloat() * (targetWidth / float32(p.GetWidth()))
	}

	return newPDFAsImageWithUnscaledWidthUnscaledHeight(p.source, p.width, p.height, targetWidth, targetHeight)
}

func (p *PDFAsImage) GetURI() *url.URL {
	return p.source
}

func (p *PDFAsImage) GetWidthAsFloat() float32 {
	return p.width
}

func (p *PDFAsImage) GetHeightAsFloat() float32 {
	return p.height
}

func (p *PDFAsImage) GetUnscaledHeight() float32 {
	return p.unscaledHeight
}

func (p *PDFAsImage) GetUnscaledWidth() float32 {
	return p.unscaledWidth
}

func (p *PDFAsImage) ScaleHeight() float32 {
	return p.height / p.unscaledHeight
}

func (p *PDFAsImage) ScaleWidth() float32 {
	return p.width / p.unscaledWidth
}
