// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/FSImage.java

package ufo

// FSImage is an image the output device can draw. Java declares
// hasSize(width, height) as a default method; Go interfaces carry no default
// methods, so it is the function FSImageHasSize.
type FSImage interface {
	GetWidth() int
	GetHeight() int
	// Scale returns an image of the same concrete type with the given size
	// (Java: <T extends FSImage> T scale(int width, int height)).
	Scale(width int, height int) FSImage
}

// FSImageHasSize ports the default method FSImage.hasSize.
func FSImageHasSize(image FSImage, width int, height int) bool {
	return image.GetWidth() == width && image.GetHeight() == height
}
