package writer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"math"
)

// imageData is the pixel data of an image, shared by the copies
// ImageGetInstanceFromImage makes, so a writer embeds it once however many
// differently scaled copies are drawn.
type imageData struct {
	width, height int
	// jpeg holds a JPEG file written as is with DCTDecode; otherwise pixels
	// holds 8-bit samples, components per pixel, row by row.
	jpeg       []byte
	pixels     []byte
	components int
	// alpha holds one 8-bit opacity sample per pixel, nil when every pixel is
	// opaque.
	alpha []byte
	// invertCMYK is set for an Adobe CMYK JPEG, whose samples are inverted.
	invertCMYK bool
}

// Image is com.lowagie.text.Image: pixel data with a size in points used
// when it is drawn. A new image is 1 point per pixel.
type Image struct {
	data                      *imageData
	plainWidth, plainHeight   float32
	scaledWidth, scaledHeight float32
	absoluteX, absoluteY      float32
	dpiX, dpiY                int
}

// ImageGetInstance decodes a JPEG, PNG or GIF file, as Image.getInstance(byte[]).
// A JPEG is embedded as it is (DCTDecode); its size and number of color
// components are read from the frame header. A PNG or GIF is decoded and
// written as 8-bit DeviceRGB (DeviceGray for a gray image) with an SMask
// when any pixel is not opaque. Only the first frame of a GIF is used.
func ImageGetInstance(data []byte) (*Image, error) {
	switch {
	case len(data) > 3 && data[0] == 0xff && data[1] == 0xd8:
		return newJPEGImage(data)
	case len(data) > 8 && string(data[1:4]) == "PNG":
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decoding PNG: %w", err)
		}
		out := ImageGetInstanceFromGoImage(img)
		out.dpiX, out.dpiY = pngDPI(data)
		return out, nil
	case len(data) > 6 && string(data[:3]) == "GIF":
		img, err := gif.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decoding GIF: %w", err)
		}
		return ImageGetInstanceFromGoImage(img), nil
	}
	return nil, newDocumentException("The byte array is not a recognized imageformat.")
}

// ImageGetInstanceFromImage returns a copy of an image that shares its pixel
// data and has its own size and position, as Image.getInstance(Image).
func ImageGetInstanceFromImage(img *Image) *Image {
	if img == nil {
		return nil
	}
	c := *img
	return &c
}

// ImageGetInstanceFromGoImage returns the image for decoded pixels.
func ImageGetInstanceFromGoImage(src image.Image) *Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	d := &imageData{width: w, height: h}
	_, isGray := src.(*image.Gray)
	if isGray {
		d.components = 1
	} else {
		d.components = 3
	}
	d.pixels = make([]byte, 0, w*h*d.components)
	alpha := make([]byte, 0, w*h)
	opaque := true
	switch img := src.(type) {
	case *image.NRGBA:
		// The samples are stored without premultiplication already.
		for y := 0; y < h; y++ {
			row := img.Pix[y*img.Stride : y*img.Stride+4*w]
			for x := 0; x < 4*w; x += 4 {
				d.pixels = append(d.pixels, row[x], row[x+1], row[x+2])
				alpha = append(alpha, row[x+3])
				if row[x+3] != 0xff {
					opaque = false
				}
			}
		}
	default:
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				// NRGBA is the color without premultiplication, which is what
				// the image samples next to an SMask have to be.
				c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
				if isGray {
					d.pixels = append(d.pixels, c.R)
				} else {
					d.pixels = append(d.pixels, c.R, c.G, c.B)
				}
				alpha = append(alpha, c.A)
				if c.A != 0xff {
					opaque = false
				}
			}
		}
	}
	if !opaque {
		d.alpha = alpha
	}
	return newImage(d)
}

func newImage(d *imageData) *Image {
	nan := float32(math.NaN())
	return &Image{
		data:       d,
		plainWidth: float32(d.width), plainHeight: float32(d.height),
		scaledWidth: float32(d.width), scaledHeight: float32(d.height),
		absoluteX: nan, absoluteY: nan,
	}
}

// newJPEGImage reads the size and the number of components from the first
// start-of-frame marker, the resolution from the JFIF header and the
// inverted-CMYK convention from the Adobe marker.
func newJPEGImage(data []byte) (*Image, error) {
	d := &imageData{jpeg: data}
	dpiX, dpiY := 0, 0
	pos := 2
	found := false
	for pos+4 <= len(data) {
		if data[pos] != 0xff {
			pos++
			continue
		}
		marker := data[pos+1]
		if marker == 0xff {
			pos++
			continue
		}
		if marker == 0xd8 || marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
			pos += 2
			continue
		}
		length := int(binary.BigEndian.Uint16(data[pos+2:]))
		seg := pos + 4
		if length < 2 || pos+2+length > len(data) {
			break
		}
		switch {
		case marker == 0xe0 && length >= 16 && string(data[seg:seg+5]) == "JFIF\x00":
			units := data[seg+7]
			dx, dy := int(binary.BigEndian.Uint16(data[seg+8:])), int(binary.BigEndian.Uint16(data[seg+10:]))
			switch units {
			case 1:
				dpiX, dpiY = dx, dy
			case 2:
				dpiX, dpiY = int(float64(dx)*2.54+0.5), int(float64(dy)*2.54+0.5)
			}
		case marker == 0xee && length >= 14 && string(data[seg:seg+5]) == "Adobe":
			d.invertCMYK = true
		case marker >= 0xc0 && marker <= 0xcf && marker != 0xc4 && marker != 0xc8 && marker != 0xcc:
			if length < 8 {
				return nil, newDocumentException("the JPEG frame header is too short")
			}
			if data[seg] != 8 {
				return nil, newDocumentException("a JPEG with %d bits per component is not supported", data[seg])
			}
			d.height = int(binary.BigEndian.Uint16(data[seg+1:]))
			d.width = int(binary.BigEndian.Uint16(data[seg+3:]))
			d.components = int(data[seg+5])
			found = true
		}
		if found {
			break
		}
		pos += 2 + length
	}
	if !found || d.width == 0 || d.height == 0 {
		return nil, newDocumentException("the JPEG has no frame header")
	}
	if d.components != 1 && d.components != 3 && d.components != 4 {
		return nil, newDocumentException("a JPEG with %d color components is not supported", d.components)
	}
	img := newImage(d)
	img.dpiX, img.dpiY = dpiX, dpiY
	return img, nil
}

// pngDPI reads the pHYs chunk; 0 when it is absent or not in meters.
func pngDPI(data []byte) (int, int) {
	pos := 8
	for pos+12 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[pos:]))
		kind := string(data[pos+4 : pos+8])
		if kind == "IDAT" || kind == "IEND" {
			break
		}
		if kind == "pHYs" && length >= 9 && pos+8+9 <= len(data) {
			if data[pos+16] != 1 {
				return 0, 0
			}
			x := float64(binary.BigEndian.Uint32(data[pos+8:]))
			y := float64(binary.BigEndian.Uint32(data[pos+12:]))
			return int(x*0.0254 + 0.5), int(y*0.0254 + 0.5)
		}
		pos += 12 + length
	}
	return 0, 0
}

// GetWidth returns the width in pixels.
func (img *Image) GetWidth() float32 { return float32(img.data.width) }

// GetHeight returns the height in pixels.
func (img *Image) GetHeight() float32 { return float32(img.data.height) }

// GetPlainWidth returns the width the image is scaled to, in points.
func (img *Image) GetPlainWidth() float32 { return img.plainWidth }

// GetPlainHeight returns the height the image is scaled to, in points.
func (img *Image) GetPlainHeight() float32 { return img.plainHeight }

// GetScaledWidth returns the width the image takes when drawn. Images are
// not rotated, so it equals GetPlainWidth.
func (img *Image) GetScaledWidth() float32 { return img.scaledWidth }

// GetScaledHeight returns the height the image takes when drawn.
func (img *Image) GetScaledHeight() float32 { return img.scaledHeight }

// ScaleAbsolute sets the drawn size in points.
func (img *Image) ScaleAbsolute(newWidth, newHeight float32) {
	img.plainWidth, img.plainHeight = newWidth, newHeight
	img.scaledWidth, img.scaledHeight = newWidth, newHeight
}

// ScaleAbsoluteWidth sets the drawn width in points.
func (img *Image) ScaleAbsoluteWidth(newWidth float32) {
	img.ScaleAbsolute(newWidth, img.plainHeight)
}

// ScaleAbsoluteHeight sets the drawn height in points.
func (img *Image) ScaleAbsoluteHeight(newHeight float32) {
	img.ScaleAbsolute(img.plainWidth, newHeight)
}

// ScalePercent sets the drawn size to a percentage of the pixel size.
func (img *Image) ScalePercent(percent float32) { img.ScalePercentXY(percent, percent) }

// ScalePercentXY is scalePercent(percentX, percentY).
func (img *Image) ScalePercentXY(percentX, percentY float32) {
	img.ScaleAbsolute(img.GetWidth()*percentX/100, img.GetHeight()*percentY/100)
}

// ScaleToFit sets the largest drawn size of the same proportions that fits
// in fitWidth by fitHeight.
func (img *Image) ScaleToFit(fitWidth, fitHeight float32) {
	img.ScalePercent(100)
	percentX := fitWidth * 100 / img.scaledWidth
	percentY := fitHeight * 100 / img.scaledHeight
	if percentX < percentY {
		img.ScalePercent(percentX)
	} else {
		img.ScalePercent(percentY)
	}
}

// SetAbsolutePosition sets where AddImageAbsolute draws the image.
func (img *Image) SetAbsolutePosition(absoluteX, absoluteY float32) {
	img.absoluteX, img.absoluteY = absoluteX, absoluteY
}

// HasAbsolutePosition reports whether SetAbsolutePosition was called.
func (img *Image) HasAbsolutePosition() bool {
	return !math.IsNaN(float64(img.absoluteX)) && !math.IsNaN(float64(img.absoluteY))
}

// GetAbsoluteX returns the x position, NaN when it was not set.
func (img *Image) GetAbsoluteX() float32 { return img.absoluteX }

// GetAbsoluteY returns the y position, NaN when it was not set.
func (img *Image) GetAbsoluteY() float32 { return img.absoluteY }

// GetDpiX returns the horizontal resolution recorded in the file (JFIF
// density, PNG pHYs), 0 when there is none.
func (img *Image) GetDpiX() int { return img.dpiX }

// GetDpiY returns the vertical resolution recorded in the file.
func (img *Image) GetDpiY() int { return img.dpiY }

// SetDpi overrides the resolution.
func (img *Image) SetDpi(dpiX, dpiY int) { img.dpiX, img.dpiY = dpiX, dpiY }

// GetColorComponents returns the number of color components per pixel: 1, 3
// or, for a CMYK JPEG, 4.
func (img *Image) GetColorComponents() int { return img.data.components }

// HasAlpha reports whether the image is written with an SMask.
func (img *Image) HasAlpha() bool { return img.data.alpha != nil }

// IsJpeg reports whether the image is embedded as a JPEG file.
func (img *Image) IsJpeg() bool { return img.data.jpeg != nil }

// addImage writes the image XObject (and its SMask) the first time an image
// with this pixel data is drawn.
func (w *PdfWriter) addImage(img *Image) (*xobjectEntry, error) {
	d := img.data
	if e, ok := w.images[d]; ok {
		return e, nil
	}
	if !w.opened || w.closed {
		return nil, newDocumentException("the document is not open")
	}
	var s *PdfStream
	if d.jpeg != nil {
		s = NewPdfStream(d.jpeg)
		s.Put("Filter", PdfName("DCTDecode"))
	} else {
		s = w.newStream(d.pixels)
	}
	s.Put("Type", PdfName("XObject"))
	s.Put("Subtype", PdfName("Image"))
	s.Put("Width", PdfNumber(d.width))
	s.Put("Height", PdfNumber(d.height))
	s.Put("BitsPerComponent", PdfNumber(8))
	switch d.components {
	case 1:
		s.Put("ColorSpace", PdfName("DeviceGray"))
	case 4:
		s.Put("ColorSpace", PdfName("DeviceCMYK"))
		if d.invertCMYK {
			s.Put("Decode", NewPdfArrayFloats(1, 0, 1, 0, 1, 0, 1, 0))
		}
	default:
		s.Put("ColorSpace", PdfName("DeviceRGB"))
	}
	if d.alpha != nil {
		mask := w.newStream(d.alpha)
		mask.Put("Type", PdfName("XObject"))
		mask.Put("Subtype", PdfName("Image"))
		mask.Put("Width", PdfNumber(d.width))
		mask.Put("Height", PdfNumber(d.height))
		mask.Put("BitsPerComponent", PdfNumber(8))
		mask.Put("ColorSpace", PdfName("DeviceGray"))
		maskRef := w.newReference()
		w.writeObject(maskRef, mask)
		s.Put("SMask", maskRef)
	}
	w.xobjectCount++
	e := &xobjectEntry{name: PdfName(fmt.Sprintf("Im%d", w.xobjectCount)), ref: w.newReference()}
	w.writeObject(e.ref, s)
	w.images[d] = e
	return e, w.err
}
