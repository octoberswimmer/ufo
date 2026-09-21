// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/FSCMYKColor.java

package ufo

import "fmt"

type FSCMYKColor struct {
	cyan    float32
	magenta float32
	yellow  float32
	black   float32
}

func NewFSCMYKColor(cyan float32, magenta float32, yellow float32, black float32) *FSCMYKColor {
	c := &FSCMYKColor{}
	c.cyan = c.validateColor(cyan, "Cyan")
	c.magenta = c.validateColor(magenta, "Magenta")
	c.yellow = c.validateColor(yellow, "Yellow")
	c.black = c.validateColor(black, "Black")
	return c
}

func (c *FSCMYKColor) validateColor(v float32, name string) float32 {
	if v < 0 || v > 1 {
		panic(NewXRRuntimeException(fmt.Sprintf("%s %s is out of range [0, 1]", name, fsRGBColorFloatToString(v))))
	}
	return v
}

func (c *FSCMYKColor) GetCyan() float32 {
	return c.cyan
}

func (c *FSCMYKColor) GetMagenta() float32 {
	return c.magenta
}

func (c *FSCMYKColor) GetYellow() float32 {
	return c.yellow
}

func (c *FSCMYKColor) GetBlack() float32 {
	return c.black
}

func (c *FSCMYKColor) ToString() string {
	return "cmyk(" + fsRGBColorFloatToString(c.cyan) + ", " + fsRGBColorFloatToString(c.magenta) + ", " +
		fsRGBColorFloatToString(c.yellow) + ", " + fsRGBColorFloatToString(c.black) + ")"
}

func (c *FSCMYKColor) String() string {
	return c.ToString()
}

func (c *FSCMYKColor) LightenColor() FSColor {
	return NewFSCMYKColor(c.cyan*0.8, c.magenta*0.8, c.yellow*0.8, c.black)
}

func (c *FSCMYKColor) DarkenColor() FSColor {
	return NewFSCMYKColor(
		min(1.0, c.cyan/0.8), min(1.0, c.magenta/0.8),
		min(1.0, c.yellow/0.8), c.black)
}
