// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextUserAgent.java

package pdf

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

const iTextUserAgentImageCacheCapacity = 32

var iTextUserAgentReSizeWithUnits = regexp.MustCompile(`^(.+?)(\D{1,2})$`)

type ITextUserAgent struct {
	NaiveUserAgent

	outputDevice *ITextOutputDevice
	dotsPerPixel int
}

var _ NaiveUserAgentI = (*ITextUserAgent)(nil)

func NewITextUserAgent(outputDevice *ITextOutputDevice, dotsPerPixel int) *ITextUserAgent {
	a := &ITextUserAgent{}
	initNaiveUserAgent(&a.NaiveUserAgent, ufo.ConfigurationValueAsInt("xr.image.cache-capacity", iTextUserAgentImageCacheCapacity))
	a.outputDevice = outputDevice
	a.dotsPerPixel = dotsPerPixel
	a.SetSelf(a)
	return a
}

// GetDotsPerPixel is package-private in Java.
func (a *ITextUserAgent) GetDotsPerPixel() int {
	return a.dotsPerPixel
}

func (a *ITextUserAgent) GetImageResource(uriStr string) *ufo.ImageResource {
	unresolvedUri := uriStr
	if !ufo.ImageUtilIsEmbeddedBase64Image(uriStr) {
		uriStr = a.self.ResolveURI(uriStr)
	}
	resource := a.imageCache.get(unresolvedUri)

	if resource == nil {
		resource = a.loadImageResource(uriStr)
		a.imageCache.put(unresolvedUri, resource)
	}
	if resource != nil {
		image := a.makeSafeCopy(resource.GetImage())
		return ufo.NewImageResource(resource.GetImageUri(), image)
	} else {
		return ufo.NewImageResource(uriStr, nil)
	}
}

// makeSafeCopy returns a nil interface for a nil image.
func (a *ITextUserAgent) makeSafeCopy(image ufo.FSImage) ufo.FSImage {
	if mutable, ok := image.(*ITextFSImage); ok {
		if mutable == nil {
			return nil
		}
		return mutable.Clone()
	}
	return image
}

// loadImageResource returns nil when the image can't be loaded.
func (a *ITextUserAgent) loadImageResource(uriStr string) *ufo.ImageResource {
	if ufo.ImageUtilIsEmbeddedBase64Image(uriStr) {
		return a.loadEmbeddedBase64ImageResource(uriStr)
	}
	resource, err := a.loadImageResourceFromStream(uriStr)
	if err != nil {
		ufo.XRLogExceptionWithTh(fmt.Sprintf("Could not load image from '%s'", uriStr), err)
		return nil
	}
	return resource
}

// loadImageResourceFromStream is the try block of Java's loadImageResource;
// the error it returns is what the catch clauses log.
func (a *ITextUserAgent) loadImageResourceFromStream(uriStr string) (*ufo.ImageResource, error) {
	stream := a.self.ResolveAndOpenStream(uriStr)
	if stream == nil {
		return nil, nil
	}
	defer stream.Close()
	cis, err := ufo.ContentTypeDetectingInputStreamWrapperDetectContentType(stream)
	if err != nil {
		return nil, err
	}
	if cis != nil {
		if cis.IsPdf() {
			uri, err := url.Parse(uriStr)
			if err != nil {
				return nil, err
			}
			// The writer does not import PDF pages: GetReader returns its
			// *writer.UnsupportedFeatureError.
			reader, err := a.outputDevice.GetReader(uri)
			if err != nil {
				return nil, err
			}
			// Java continues with reader.getPageSizeWithRotation(1) and wraps
			// the page in a PDFAsImage of that size scaled by the dots per
			// point of the output device. writer.PdfReader has no page size
			// accessor.
			_ = reader
			panic(ufo.NewXRRuntimeException("not ported: PdfReader.getPageSizeWithRotation"))
		}
		isSvg, err := cis.IsSvg()
		if err != nil {
			return nil, err
		}
		if isSvg {
			image, err := a.readCsv(uriStr, cis)
			if err != nil {
				return nil, err
			}
			return ufo.NewImageResource(uriStr, image), nil
		} else {
			data, err := ufo.IOUtilReadBytesInputStream(cis)
			if err != nil {
				return nil, err
			}
			image, err := writer.ImageGetInstance(data)
			if err != nil {
				return nil, err
			}
			a.scaleToOutputResolution(image)
			return ufo.NewImageResource(uriStr, NewITextFSImage(image)), nil
		}
	}
	return nil, nil
}

func (a *ITextUserAgent) loadEmbeddedBase64ImageResource(uri string) *ufo.ImageResource {
	data := ufo.ImageUtilGetEmbeddedBase64Image(uri)
	if data == nil {
		// requireNonNull
		panic(ufo.NewXRRuntimeException("Can't read embedded base64 image from " + uri))
	}
	image, err := writer.ImageGetInstance(data)
	if err != nil {
		ufo.XRLogExceptionWithTh("Can't read embedded base64 image from "+uri, err)
		return ufo.NewImageResource("", nil)
	}
	a.scaleToOutputResolution(image)
	return ufo.NewImageResource(uri, NewITextFSImage(image))
}

// readCsv reads an SVG image. Java returns an SvgImage, which draws the
// document with Batik. SvgImage is not ported and the writer does not draw
// SVG: after reading the size as Java does, the
// *writer.UnsupportedFeatureError of writer.ImageGetInstanceFromSvg is
// returned.
func (a *ITextUserAgent) readCsv(uri string, in io.Reader) (ufo.FSImage, error) {
	svgBytes, err := ufo.IOUtilReadBytesInputStream(in)
	if err != nil {
		return nil, err
	}
	if _, err := a.GetOriginalSvgSize(uri, svgBytes); err != nil {
		return nil, err
	}
	_, err = writer.ImageGetInstanceFromSvg(svgBytes)
	return nil, err
}

// GetOriginalSvgSize is package-private in Java.
func (a *ITextUserAgent) GetOriginalSvgSize(uri string, svgImage []byte) (*ufo.Size, error) {
	document, err := dom.ParseXML(bytes.NewReader(svgImage))
	if err != nil {
		return nil, err
	}
	return ITextUserAgentGetSvgSize(document.GetDocumentElement()), nil
}

// ITextUserAgentGetSvgSize is package-private in Java.
func ITextUserAgentGetSvgSize(svgRoot *dom.Element) *ufo.Size {
	width := svgRoot.GetAttribute("width")
	height := svgRoot.GetAttribute("height")
	if width != "" && height != "" {
		return ufo.NewSize(ITextUserAgentParseSize(width), ITextUserAgentParseSize(height))
	}
	viewBox := strings.SplitN(svgRoot.GetAttribute("viewBox"), " ", 4)
	if len(viewBox) >= 4 {
		return ufo.NewSize(ITextUserAgentParseSize(viewBox[2]), ITextUserAgentParseSize(viewBox[3]))
	}

	return ufo.NewSize(300, 150) // default size in most browsers
}

func (a *ITextUserAgent) scaleToOutputResolution(image *writer.Image) {
	factor := float32(a.dotsPerPixel)
	if factor != 1.0 {
		image.ScaleAbsolute(image.GetPlainWidth()*factor, image.GetPlainHeight()*factor)
	}
}

// ITextUserAgentParseSize is taken from
// https://developer.mozilla.org/en-US/docs/Learn_web_development/Core/Styling_basics/Values_and_units#absolute_length_units
//
// cssValue is e.g. "100px", "200pt", "300mm", "400cm", "500Q", "600pc",
// "700in". It returns the size in pixels. It is package-private in Java.
func ITextUserAgentParseSize(cssValue string) int {
	value := iTextUserAgentParseCssValue(cssValue)

	var size float64
	switch value.units {
	case "cm":
		size = 37.8 * float64(value.value)
	case "mm":
		size = 96.0 / 25.4 * float64(value.value)
	case "Q":
		size = 96.0 / 25.4 / 4 * float64(value.value)
	case "in":
		size = 96.0 * float64(value.value)
	case "pt":
		size = 96.0 / 72 * float64(value.value)
	case "pc":
		size = 96.0 / 72 * 12 * float64(value.value)
	default:
		size = float64(value.value)
	}
	return int(math.Floor(size + 0.5))
}

func iTextUserAgentParseCssValue(cssValue string) iTextUserAgentCssSize {
	matcher := iTextUserAgentReSizeWithUnits.FindStringSubmatch(cssValue)
	value := cssValue
	units := "px"
	if matcher != nil {
		value = matcher[1]
		units = matcher[2]
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 32)
	if err != nil {
		// NumberFormatException
		panic(ufo.NewXRRuntimeExceptionWithCause(fmt.Sprintf("For input string: \"%s\"", value), err))
	}
	return iTextUserAgentCssSize{value: float32(parsed), units: units}
}

// iTextUserAgentCssSize ports the private record ITextUserAgent.CssSize.
type iTextUserAgentCssSize struct {
	value float32
	units string
}
