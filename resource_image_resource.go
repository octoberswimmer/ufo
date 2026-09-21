// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/resource/ImageResource.java

package ufo

type ImageResource struct {
	AbstractResource
	imageUri string
	img      FSImage
}

// NewImageResource creates an ImageResource. uri may be "" and img may be
// nil.
//
// HACK: at least for now, till we know what we want to do here
func NewImageResource(uri string, img FSImage) *ImageResource {
	r := &ImageResource{imageUri: uri, img: img}
	r.initAbstractResource(nil)
	return r
}

// GetImage returns nil when the image could not be loaded.
func (r *ImageResource) GetImage() FSImage {
	return r.img
}

// IsLoaded reports whether the image has been loaded. Java returns false only
// for a swing MutableFSImage that is still loading; that class is not ported,
// so every image is loaded.
func (r *ImageResource) IsLoaded() bool {
	return true
}

func (r *ImageResource) GetImageUri() string {
	return r.imageUri
}

// HasDimensions reports whether the image is loaded and has the given size.
// Java requires the image to be a swing AWTFSImage; that class is not ported,
// so any non-nil FSImage is compared.
func (r *ImageResource) HasDimensions(width int, height int) bool {
	return r.IsLoaded() &&
		r.img != nil &&
		FSImageHasSize(r.img, width, height)
}
