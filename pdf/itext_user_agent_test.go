// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/ITextUserAgentTest.java

package pdf

import "testing"

func iTextUserAgentTestAssertSvgSize(t *testing.T, svg string, width int, height int) {
	t.Helper()
	agent := NewITextUserAgent(NewITextOutputDevice(1.0), 1)
	size, err := agent.GetOriginalSvgSize("https://some.test.com", []byte(svg))
	if err != nil {
		t.Fatal(err)
	}
	if size.Width() != width || size.Height() != height {
		t.Errorf("size = %dx%d, want %dx%d", size.Width(), size.Height(), width, height)
	}
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="200">
</svg>
`, 300, 200)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inCentimeters(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="4cm" height="8cm">
</svg>
`, 151, 302)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inMillimeters(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="40mm" height="80mm">
</svg>
`, 151, 302)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inQuarterMillimeters(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="4000Q" height="8000Q">
</svg>
`, 3780, 7559)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inInches(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="4in" height="8in">
</svg>
`, 4*96, 8*96)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inPicas(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="100pc" height="200pc">
</svg>
`, 1600, 3200)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inPoints(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="100pt" height="200pt">
</svg>
`, 133, 267)
}

func TestITextUserAgent_svg_withWidthAndHeightAttributes_inPixels(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" width="4px" height="8px">
</svg>
`, 4, 8)
}

func TestITextUserAgent_svg_withViewBoxAttribute(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 600 500">
</svg>
`, 600, 500)
}

func TestITextUserAgent_svg_withoutSpecifiedSize(t *testing.T) {
	iTextUserAgentTestAssertSvgSize(t, `<svg xmlns="http://www.w3.org/2000/svg">
</svg>
`, 300, 150)
}
