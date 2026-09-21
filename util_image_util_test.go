// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/util/ImageUtilTest.java

package ufo

import (
	"os"
	"strings"
	"testing"
)

func TestImageUtilEmbeddedImageUrl(t *testing.T) {
	tests := []struct {
		uri  string
		want bool
	}{
		// The first Java assertion passes null, which is "" here.
		{"", false},
		{"https://selenide.org/images/selenide-logo-big.png", false},
		{"data:image/png;base64,iVBORw0KG...", true},
		{"data:image/png;base64, iVBORw0KGgo...", true},
		{"data:image/; base64,iVBORw0KG...", true},
		{"data:image/;base64,iVBORw0KG...", true},
		{"data:image;base64,iVBORw0KG...", true},
		{"data:image/svg+xml;utf8,<svg xmlns='....", true},
		{"data:image/svg;utf8,<svg xmlns='....", true},
	}
	for _, test := range tests {
		if got := ImageUtilIsEmbeddedBase64Image(test.uri); got != test.want {
			t.Errorf("ImageUtilIsEmbeddedBase64Image(%q) = %t, want %t", test.uri, got, test.want)
		}
	}
}

// TestImageUtilEmbeddedBase64Image decodes the two data URIs of the Java
// test, which are in testdata/util because of their size. The second one has
// its "+" characters written as %2B.
func TestImageUtilEmbeddedBase64Image(t *testing.T) {
	for _, name := range []string{
		"testdata/util/image_util_embedded_base64_image_1.txt",
		"testdata/util/image_util_embedded_base64_image_2.txt",
	} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		image := ImageUtilLoadEmbeddedBase64Image(strings.TrimSpace(string(data)))
		if image == nil {
			t.Fatalf("%s: no image", name)
		}
		if width := image.Bounds().Dx(); width != 352 {
			t.Errorf("%s: width %d, want 352", name, width)
		}
		if height := image.Bounds().Dy(); height != 186 {
			t.Errorf("%s: height %d, want 186", name, height)
		}
	}
}

// The Java test detectsImageType covers ImageUtil.detectTransparency, which
// works on the image types of java.awt.image.BufferedImage and is not ported.

func TestImageUtilGetEmbeddedBase64ImageWithoutBase64(t *testing.T) {
	if data := ImageUtilGetEmbeddedBase64Image("data:image/svg+xml;utf8,<svg/>"); data != nil {
		t.Errorf("got %d bytes for a URI that is not base 64 encoded", len(data))
	}
	if image := ImageUtilLoadEmbeddedBase64Image("data:image/svg+xml;utf8,<svg/>"); image != nil {
		t.Errorf("got an image for a URI that is not base 64 encoded")
	}
}
