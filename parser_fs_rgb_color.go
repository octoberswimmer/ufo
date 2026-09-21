// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/FSRGBColor.java

package ufo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	FSRGBColorTransparent = NewFSRGBColor(0, 0, 0)
	FSRGBColorRed         = NewFSRGBColor(255, 0, 0)
	FSRGBColorGreen       = NewFSRGBColor(0, 255, 0)
	FSRGBColorBlue        = NewFSRGBColor(0, 0, 255)
)

type FSRGBColor struct {
	red   int
	green int
	blue  int
	alpha float32
}

func NewFSRGBColor(red int, green int, blue int) *FSRGBColor {
	return NewFSRGBColorWithAlpha(red, green, blue, 1.0)
}

func NewFSRGBColorWithAlpha(red int, green int, blue int, alpha float32) *FSRGBColor {
	c := &FSRGBColor{}
	c.red = c.validateColor("Red", red)
	c.green = c.validateColor("Green", green)
	c.blue = c.validateColor("Blue", blue)
	c.alpha = c.validateAlpha(alpha)
	return c
}

func (c *FSRGBColor) validateColor(name string, color int) int {
	if color < 0 || color > 255 {
		panic(NewXRRuntimeException(fmt.Sprintf("%s %d is out of range [0, 255]", name, color)))
	}
	return color
}

func (c *FSRGBColor) validateAlpha(alpha float32) float32 {
	if alpha < 0 || alpha > 1 {
		panic(NewXRRuntimeException(fmt.Sprintf("alpha %s is out of range [0, 1]", fsRGBColorFloatToString(alpha))))
	}
	return alpha
}

// NewFSRGBColorInt ports the constructor FSRGBColor(int color), which takes
// the color as 0xRRGGBB.
func NewFSRGBColorInt(color int) *FSRGBColor {
	return NewFSRGBColor((color&0xff0000)>>16, (color&0x00ff00)>>8, color&0xff)
}

func (c *FSRGBColor) GetBlue() int {
	return c.blue
}

func (c *FSRGBColor) GetGreen() int {
	return c.green
}

func (c *FSRGBColor) GetRed() int {
	return c.red
}

func (c *FSRGBColor) GetAlpha() float32 {
	return c.alpha
}

func (c *FSRGBColor) ToString() string {
	if c.alpha != 1 {
		return "rgba(" + strconv.Itoa(c.red) + "," + strconv.Itoa(c.green) + "," + strconv.Itoa(c.blue) + "," + fsRGBColorFloatToString(c.alpha) + ")"
	}
	return "#" + c.toStringInt(c.red) + c.toStringInt(c.green) + c.toStringInt(c.blue)
}

func (c *FSRGBColor) String() string {
	return c.ToString()
}

func (c *FSRGBColor) toStringInt(color int) string {
	return fmt.Sprintf("%02x", color)
}

func (c *FSRGBColor) Equals(o any) bool {
	that, ok := o.(*FSRGBColor)
	if !ok || that == nil {
		return false
	}
	if c == that {
		return true
	}
	return c.blue == that.blue && c.green == that.green && c.red == that.red && c.alpha == that.alpha
}

// HashCode returns the value of java.util.Objects.hash(red, green, blue).
func (c *FSRGBColor) HashCode() int {
	var result int32 = 1
	for _, v := range []int{c.red, c.green, c.blue} {
		result = 31*result + int32(v)
	}
	return int(result)
}

func (c *FSRGBColor) LightenColor() FSColor {
	hsb := c.ToHSB()
	sLighter := float32(float32(0.35*hsb.Brightness()) * hsb.Saturation())
	bLighter := 0.6999 + float32(0.3*hsb.Brightness())
	return NewHSBColor(hsb.Hue(), sLighter, bLighter).ToRGB()
}

func (c *FSRGBColor) DarkenColor() FSColor {
	hsb := c.ToHSB()
	hBase := hsb.Hue()
	sBase := hsb.Saturation()
	bBase := hsb.Brightness()
	bDarker := float32(0.56 * bBase)

	return NewHSBColor(hBase, sBase, bDarker).ToRGB()
}

func (c *FSRGBColor) ToHSB() *HSBColor {
	return fsRGBColorRGBtoHSB(c.GetRed(), c.GetGreen(), c.GetBlue())
}

// Taken from java.awt.Color to avoid dependency on it
func fsRGBColorRGBtoHSB(r int, g int, b int) *HSBColor {
	cmax := fsRGBColorMax(r, g, b)
	cmin := fsRGBColorMin(r, g, b)
	brightness := cmax / 255.0
	var saturation float32
	if cmax == 0.0 {
		saturation = 0.0
	} else {
		saturation = (cmax - cmin) / cmax
	}
	var hue float32
	if saturation == 0 {
		hue = 0
	} else {
		hue = fsRGBColorCalculateHue(r, g, b, cmax, cmin)
	}
	return NewHSBColor(hue, saturation, brightness)
}

func fsRGBColorCalculateHue(r int, g int, b int, cmax float32, cmin float32) float32 {
	redc := (cmax - float32(r)) / (cmax - cmin)
	greenc := (cmax - float32(g)) / (cmax - cmin)
	bluec := (cmax - float32(b)) / (cmax - cmin)
	var hue1 float32
	if float32(r) == cmax {
		hue1 = bluec - greenc
	} else if float32(g) == cmax {
		hue1 = 2.0 + redc - bluec
	} else {
		hue1 = 4.0 + greenc - redc
	}

	hue2 := hue1 / 6.0
	if hue1 < 0 {
		return hue2 + 1.0
	}
	return hue2
}

func fsRGBColorMax(a int, b int, c int) float32 {
	return float32(max(max(a, b), c))
}

func fsRGBColorMin(a int, b int, c int) float32 {
	return float32(min(min(a, b), c))
}

// fsRGBColorFloatToString formats f as java.lang.Float.toString does: the
// shortest decimal that identifies the float, with at least one digit after
// the decimal point, in plain notation for 1e-3 <= |f| < 1e7 and in the form
// 1.0E-4 outside that range.
func fsRGBColorFloatToString(f float32) string {
	f64 := float64(f)
	switch {
	case math.IsNaN(f64):
		return "NaN"
	case math.IsInf(f64, 1):
		return "Infinity"
	case math.IsInf(f64, -1):
		return "-Infinity"
	case f64 == 0:
		if math.Signbit(f64) {
			return "-0.0"
		}
		return "0.0"
	}
	abs := math.Abs(f64)
	if abs >= 1e-3 && abs < 1e7 {
		s := strconv.FormatFloat(f64, 'f', -1, 32)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}
	s := strconv.FormatFloat(f64, 'e', -1, 32)
	mantissa, exponent, _ := strings.Cut(s, "e")
	if !strings.Contains(mantissa, ".") {
		mantissa += ".0"
	}
	exp, err := strconv.Atoi(exponent)
	if err != nil {
		panic(NewXRRuntimeExceptionWithCause("cannot format float "+s, err))
	}
	return mantissa + "E" + strconv.Itoa(exp)
}
