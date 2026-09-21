// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/HSBColor.java

package ufo

import "math"

type HSBColor struct {
	hue        float32
	saturation float32
	brightness float32
}

func NewHSBColor(hue float32, saturation float32, brightness float32) *HSBColor {
	return &HSBColor{hue: hue, saturation: saturation, brightness: brightness}
}

func (c *HSBColor) Hue() float32 {
	return c.hue
}

func (c *HSBColor) Saturation() float32 {
	return c.saturation
}

func (c *HSBColor) Brightness() float32 {
	return c.brightness
}

// ToRGB is taken from java.awt.Color to avoid dependency on it. The float32
// conversions around products keep the compiler from fusing a multiplication
// and an addition into one operation, which Java does not do.
func (c *HSBColor) ToRGB() *FSRGBColor {
	hue, saturation, brightness := c.hue, c.saturation, c.brightness
	r, g, b := 0, 0, 0
	if saturation == 0 {
		r = int(float32(brightness*255.0) + 0.5)
		g = r
		b = r
	} else {
		h := (hue - float32(math.Floor(float64(hue)))) * 6.0
		f := h - float32(math.Floor(float64(h)))
		p := brightness * (1.0 - saturation)
		q := brightness * (1.0 - float32(saturation*f))
		t := brightness * (1.0 - float32(saturation*(1.0-f)))
		switch int(h) {
		case 0:
			r = int(float32(brightness*255.0) + 0.5)
			g = int(float32(t*255.0) + 0.5)
			b = int(float32(p*255.0) + 0.5)
		case 1:
			r = int(float32(q*255.0) + 0.5)
			g = int(float32(brightness*255.0) + 0.5)
			b = int(float32(p*255.0) + 0.5)
		case 2:
			r = int(float32(p*255.0) + 0.5)
			g = int(float32(brightness*255.0) + 0.5)
			b = int(float32(t*255.0) + 0.5)
		case 3:
			r = int(float32(p*255.0) + 0.5)
			g = int(float32(q*255.0) + 0.5)
			b = int(float32(brightness*255.0) + 0.5)
		case 4:
			r = int(float32(t*255.0) + 0.5)
			g = int(float32(p*255.0) + 0.5)
			b = int(float32(brightness*255.0) + 0.5)
		case 5:
			r = int(float32(brightness*255.0) + 0.5)
			g = int(float32(p*255.0) + 0.5)
			b = int(float32(q*255.0) + 0.5)
		}
	}
	return NewFSRGBColor(r, g, b)
}

// Equals compares the three components, as the Java record does.
func (c *HSBColor) Equals(o *HSBColor) bool {
	return o != nil && c.hue == o.hue && c.saturation == o.saturation && c.brightness == o.brightness
}
