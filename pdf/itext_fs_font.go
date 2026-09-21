// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextFSFont.java

package pdf

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/octoberswimmer/ufo"
)

type ITextFSFont struct {
	font *FontDescription
	size float32
}

var _ ufo.FSFont = (*ITextFSFont)(nil)

func NewITextFSFont(font *FontDescription, size float32) *ITextFSFont {
	return &ITextFSFont{font: font, size: size}
}

func (f *ITextFSFont) GetSize2D() float32 {
	return f.size
}

func (f *ITextFSFont) GetFontDescription() *FontDescription {
	return f.font
}

func (f *ITextFSFont) String() string {
	return fmt.Sprintf("%s:%s", f.font, iTextFSFontFloatToString(f.size))
}

func (f *ITextFSFont) ToString() string {
	return f.String()
}

// iTextFSFontFloatToString formats a float32 as Java's Float.toString does.
// The core's propertyBuilderFloatToString does the same but is unexported.
func iTextFSFontFloatToString(f float32) string {
	d := float64(f)
	switch {
	case math.IsNaN(d):
		return "NaN"
	case math.IsInf(d, 1):
		return "Infinity"
	case math.IsInf(d, -1):
		return "-Infinity"
	}
	abs := math.Abs(d)
	if abs != 0 && (abs < 1e-3 || abs >= 1e7) {
		s := strconv.FormatFloat(d, 'e', -1, 32)
		mantissa, exponent, _ := strings.Cut(s, "e")
		if !strings.Contains(mantissa, ".") {
			mantissa += ".0"
		}
		exp, _ := strconv.Atoi(exponent)
		return mantissa + "E" + strconv.Itoa(exp)
	}
	s := strconv.FormatFloat(d, 'f', -1, 32)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
