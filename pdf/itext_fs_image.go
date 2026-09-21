// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextFSImage.java

package pdf

import (
	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

type ITextFSImage struct {
	image *writer.Image
}

func NewITextFSImage(image *writer.Image) *ITextFSImage {
	return &ITextFSImage{image: image}
}

func (i *ITextFSImage) GetWidth() int {
	return int(i.image.GetPlainWidth())
}

func (i *ITextFSImage) GetHeight() int {
	return int(i.image.GetPlainHeight())
}

func (i *ITextFSImage) Scale(width int, height int) ufo.FSImage {
	current := ufo.NewSize(i.GetWidth(), i.GetHeight())
	target := current.Scale(width, height)

	if !target.Equals(current) {
		scaledImage := writer.ImageGetInstanceFromImage(i.image)
		scaledImage.ScaleAbsolute(float32(target.Width()), float32(target.Height()))
		return NewITextFSImage(scaledImage)
	}
	return i
}

func (i *ITextFSImage) GetImage() *writer.Image {
	return i.image
}

func (i *ITextFSImage) Clone() *ITextFSImage {
	return NewITextFSImage(writer.ImageGetInstanceFromImage(i.image))
}
