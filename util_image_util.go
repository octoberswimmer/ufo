// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/ImageUtil.java

package ufo

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

// Only the data URI functions of ImageUtil are ported. The rest of the Java
// class works on java.awt.image.BufferedImage for the Swing and Java2D
// renderers (clearImage, withGraphics, makeCompatible,
// createCompatibleBufferedImage, detectTransparency, getScaledInstance and
// its Scaler implementations, convertToBufferedImage,
// createTransparentImage) and is not ported.

// ImageUtilIsEmbeddedBase64Image detects if the given URI represents an
// embedded base 64 image.
func ImageUtilIsEmbeddedBase64Image(uri string) bool {
	return strings.HasPrefix(uri, "data:image")
}

// ImageUtilGetEmbeddedBase64Image returns the binary content of an embedded
// base 64 image, or nil when the URI is not base 64 encoded.
func ImageUtilGetEmbeddedBase64Image(imageDataUri string) []byte {
	b64Index := strings.Index(imageDataUri, "base64,")
	if b64Index != -1 {
		b64encoded := imageDataUri[b64Index+len("base64,"):]
		return utilDecodeBase64DataUriPayload(b64encoded)
	}
	XRLogLoadWithLevel(LevelSevere, "Embedded XHTML images must be encoded in base 64.")
	return nil
}

// ImageUtilLoadEmbeddedBase64Image returns the decoded image of an embedded
// base 64 image, as an image.Image where Java returns a BufferedImage. The
// formats are the ones the Go standard library decodes: PNG, JPEG and GIF.
// The result is nil when the URI is not base 64 encoded or the data is not
// an image in one of these formats.
func ImageUtilLoadEmbeddedBase64Image(imageDataUri string) image.Image {
	buffer := ImageUtilGetEmbeddedBase64Image(imageDataUri)
	if buffer != nil {
		img, _, err := image.Decode(bytes.NewReader(buffer))
		if errors.Is(err, image.ErrFormat) {
			// ImageIO.read returns null when no reader accepts the data.
			return nil
		}
		if err != nil {
			XRLogExceptionWithTh("Can't read XHTML embedded image", err)
			return nil
		}
		return img
	}
	return nil
}
