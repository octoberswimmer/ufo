// Tests for render_abstract_output_device.go. Flying Saucer has no JUnit test
// for AbstractOutputDevice. testdata/render/abstract_output_device_background.txt
// is the call log of a Java subclass of AbstractOutputDevice (flying-saucer-core
// 10.6.0-SNAPSHOT) that records the abstract methods paintBackground calls,
// for the same styles, rectangles and clips as the test below.

package ufo

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/geom"
)

// abstractOutputDeviceTestImage is an image of a given size. Scale logs its
// arguments and resolves -1 from the aspect ratio.
type abstractOutputDeviceTestImage struct {
	out  *strings.Builder
	w, h int
}

func (i *abstractOutputDeviceTestImage) GetWidth() int  { return i.w }
func (i *abstractOutputDeviceTestImage) GetHeight() int { return i.h }
func (i *abstractOutputDeviceTestImage) Scale(width int, height int) FSImage {
	fmt.Fprintf(i.out, "scale %d %d\n", width, height)
	nw, nh := width, height
	if nw == -1 {
		nw = i.w * nh / i.h
	}
	if nh == -1 {
		nh = i.h * nw / i.w
	}
	return &abstractOutputDeviceTestImage{i.out, nw, nh}
}

// abstractOutputDeviceTestDevice embeds AbstractOutputDevice as the PDF output
// device does and records the calls of the methods that are abstract in Java.
// A call of a method it does not define panics on the nil embedded interface.
type abstractOutputDeviceTestDevice struct {
	AbstractOutputDevice
	OutputDevice
	out   *strings.Builder
	clip  geom.Shape
	lines [][4]int
}

func newAbstractOutputDeviceTestDevice(out *strings.Builder) *abstractOutputDeviceTestDevice {
	d := &abstractOutputDeviceTestDevice{out: out}
	InitAbstractOutputDevice(&d.AbstractOutputDevice, d)
	return d
}

// The methods AbstractOutputDevice implements are also in the embedded nil
// OutputDevice, at the same depth, so the device names the ones to use.
func (d *abstractOutputDeviceTestDevice) DrawText(c *RenderingContext, inlineText *InlineText) {
	d.AbstractOutputDevice.DrawText(c, inlineText)
}
func (d *abstractOutputDeviceTestDevice) DrawTextDecoration(c *RenderingContext, lineBox *LineBox) {
	d.AbstractOutputDevice.DrawTextDecoration(c, lineBox)
}
func (d *abstractOutputDeviceTestDevice) DrawTextDecorationWithIBDecoration(c *RenderingContext, iB *InlineLayoutBox, decoration *TextDecoration) {
	d.AbstractOutputDevice.DrawTextDecorationWithIBDecoration(c, iB, decoration)
}
func (d *abstractOutputDeviceTestDevice) DrawDebugOutline(c *RenderingContext, box BoxI, color FSColor) {
	d.AbstractOutputDevice.DrawDebugOutline(c, box, color)
}
func (d *abstractOutputDeviceTestDevice) PaintCollapsedBorder(c *RenderingContext, border *BorderPropertySet, bounds *geom.Rectangle, side int) {
	d.AbstractOutputDevice.PaintCollapsedBorder(c, border, bounds, side)
}
func (d *abstractOutputDeviceTestDevice) PaintBorder(c *RenderingContext, box BoxI) {
	d.AbstractOutputDevice.PaintBorder(c, box)
}
func (d *abstractOutputDeviceTestDevice) PaintBorderWithStyleEdgeSides(c *RenderingContext, style CalculatedStyleI, edge *geom.Rectangle, sides int) {
	d.AbstractOutputDevice.PaintBorderWithStyleEdgeSides(c, style, edge, sides)
}
func (d *abstractOutputDeviceTestDevice) PaintBackground(c *RenderingContext, box BoxI) {
	d.AbstractOutputDevice.PaintBackground(c, box)
}
func (d *abstractOutputDeviceTestDevice) PaintBackgroundWithStyleBoundsBgImageContainerBorder(c *RenderingContext, style CalculatedStyleI, bounds *geom.Rectangle, bgImageContainer *geom.Rectangle, border *BorderPropertySet) {
	d.AbstractOutputDevice.PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, style, bounds, bgImageContainer, border)
}

func (d *abstractOutputDeviceTestDevice) DrawLine(x1 int, y1 int, x2 int, y2 int) {
	d.lines = append(d.lines, [4]int{x1, y1, x2, y2})
}

func (d *abstractOutputDeviceTestDevice) SetColor(color FSColor) {
	fmt.Fprintf(d.out, "setColor %s\n", color.ToString())
}

func (d *abstractOutputDeviceTestDevice) SetOpacity(opacity float32) {
	fmt.Fprintf(d.out, "setOpacity %x\n", math.Float32bits(opacity))
}

func (d *abstractOutputDeviceTestDevice) DrawImage(image FSImage, x int, y int) {
	fmt.Fprintf(d.out, "drawImage %d %d %d %d\n", image.GetWidth(), image.GetHeight(), x, y)
}

func (d *abstractOutputDeviceTestDevice) DrawLinearGradient(gradient *FSLinearGradient, x int, y int, width int, height int) {
	fmt.Fprintf(d.out, "drawLinearGradient %d %d %d %d\n", x, y, width, height)
}

func (d *abstractOutputDeviceTestDevice) Fill(s geom.Shape) {
	d.out.WriteString("fill\n")
}

func (d *abstractOutputDeviceTestDevice) Clip(s geom.Shape) {
	d.out.WriteString("clip\n")
}

func (d *abstractOutputDeviceTestDevice) GetClip() geom.Shape {
	d.out.WriteString("getClip\n")
	return d.clip
}

func (d *abstractOutputDeviceTestDevice) SetClip(s geom.Shape) {
	switch {
	case s == nil:
		d.out.WriteString("setClip null\n")
	case s == d.clip:
		d.out.WriteString("setClip old\n")
	default:
		d.out.WriteString("setClip area\n")
	}
}

var _ AbstractOutputDeviceI = (*abstractOutputDeviceTestDevice)(nil)

type abstractOutputDeviceTestUac struct {
	out *strings.Builder
}

func (u *abstractOutputDeviceTestUac) GetCSSResource(uri string) *CSSResource { return nil }
func (u *abstractOutputDeviceTestUac) GetImageResource(uri string) *ImageResource {
	fmt.Fprintf(u.out, "getImageResource %s\n", uri)
	if strings.Contains(uri, "missing") {
		return nil
	}
	if strings.Contains(uri, "empty") {
		return NewImageResource(uri, &abstractOutputDeviceTestImage{u.out, 0, 0})
	}
	if strings.Contains(uri, "none") {
		return NewImageResource(uri, nil)
	}
	return NewImageResource(uri, &abstractOutputDeviceTestImage{u.out, 16, 12})
}
func (u *abstractOutputDeviceTestUac) GetXMLResource(uri string) *XMLResource { return nil }
func (u *abstractOutputDeviceTestUac) GetBinaryResource(uri string) []byte    { return nil }
func (u *abstractOutputDeviceTestUac) IsVisited(uri string) bool              { return false }
func (u *abstractOutputDeviceTestUac) SetBaseURL(url string)                  {}
func (u *abstractOutputDeviceTestUac) GetBaseURL() string                     { return "" }
func (u *abstractOutputDeviceTestUac) ResolveURI(uri string) string           { return uri }

func abstractOutputDeviceTestStyle(t *testing.T, declarations string) CalculatedStyleI {
	t.Helper()
	parser := NewCSSParser(CSSErrorHandlerFunc(func(uri string, message string) {
		t.Errorf("CSS error: %s", message)
	}))
	sheet, err := parser.ParseStylesheet("", StylesheetInfoOriginAuthor, strings.NewReader("@page { "+declarations+" }"))
	if err != nil {
		t.Fatal(err)
	}
	matcher := NewMatcher(nil, nil, nil, []*Stylesheet{sheet}, "print")
	return NewEmptyStyle().DeriveStyle(matcher.GetPageCascadedStyle("", "first").GetPageStyle())
}

func TestAbstractOutputDevicePaintBackgroundAgainstJava(t *testing.T) {
	wantBytes, err := os.ReadFile("testdata/render/abstract_output_device_background.txt")
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	sharedContext := NewSharedContextWithUac(&abstractOutputDeviceTestUac{&out})
	sharedContext.SetDPI(96)
	sharedContext.SetDotsPerPixel(1)
	dev := newAbstractOutputDeviceTestDevice(&out)
	rc := NewRenderingContext(sharedContext, dev, nil, nil, 0)

	cases := []string{
		"background-color: red",
		"color: red",
		"background-color: transparent",
		"background-color: red; opacity: 0.5",
		"background-image: url(img.png)",
		"background-image: url(img.png); background-repeat: no-repeat; background-position: 50% 50%",
		"background-image: url(img.png); background-repeat: repeat-x; background-position: 5px 3px",
		"background-image: url(img.png); background-repeat: repeat-y; background-position: 100% 0%; border: 3px solid black",
		"background-image: url(img.png); background-repeat: no-repeat; background-size: contain",
		"background-image: url(img.png); background-repeat: no-repeat; background-size: cover",
		"background-image: url(img.png); background-repeat: no-repeat; background-size: 50% auto",
		"background-image: url(img.png); background-size: 20px 30px",
		"background-image: url(img.png); background-repeat: repeat-x; background-position: -40px 70%",
		"background-image: url(img.png); background-repeat: repeat-y; background-position: 30px 35px",
		"background-image: linear-gradient(to right, red, blue)",
		"background-color: blue; background-image: url(img.png); background-repeat: no-repeat; background-position: -50px -50px",
		"background-color: blue; background-image: url(empty.png)",
		"background-color: blue; background-image: url(none.png)",
		"background-color: blue; background-image: url(missing.png)",
		"background-image: url(missing.png)",
		"background-image: url(img.png); border: 2px solid red; border-radius: 10px; border-left-width: 7px",
	}
	// Each entry is the background bounds and the background image container.
	rects := [][2]*geom.Rectangle{
		{geom.NewRectangle(10, 20, 100, 60), geom.NewRectangle(10, 20, 100, 60)},
		{geom.NewRectangle(10, 20, 100, 60), geom.NewRectangle(-5, 3, 40, 45)},
	}
	for ci, declarations := range cases {
		style := abstractOutputDeviceTestStyle(t, declarations)
		for ri, r := range rects {
			for withClip := 0; withClip < 2; withClip++ {
				if withClip == 1 {
					dev.clip = geom.NewRectangle(0, 0, 50, 50)
				} else {
					dev.clip = nil
				}
				fmt.Fprintf(&out, "case %d %d %d\n", ci, ri, withClip)
				dev.PaintBackgroundWithStyleBoundsBgImageContainerBorder(rc, style, r[0], r[1], style.GetBorder(rc))
			}
		}
	}
	// A style without a border: neither call reaches the output device.
	out.WriteString("case nullborder\n")
	style := abstractOutputDeviceTestStyle(t, "background-image: url(img.png); background-repeat: no-repeat")
	dev.clip = nil
	dev.PaintCollapsedBorder(rc, style.GetBorder(rc), rects[0][0], BorderPainterTop)
	dev.PaintBorderWithStyleEdgeSides(rc, style, rects[0][0], BorderPainterAll)

	gotLines := strings.Split(out.String(), "\n")
	wantLines := strings.Split(string(wantBytes), "\n")
	current := ""
	for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
		got, want := "<end>", "<end>"
		if i < len(gotLines) {
			got = gotLines[i]
		}
		if i < len(wantLines) {
			want = wantLines[i]
		}
		if strings.HasPrefix(want, "case ") {
			current = want
		}
		if got != want {
			t.Fatalf("line %d (in %q): got %q, Java gives %q", i+1, current, got, want)
		}
	}
}

func TestAbstractOutputDeviceAdjustTo(t *testing.T) {
	a := &AbstractOutputDevice{}
	cases := []struct{ target, current, imageDim, want int }{
		{10, 10, 16, 10},
		{10, 15, 16, -1},
		{10, 100, 16, 4},
		{10, 26, 16, 10},
		{10, 5, 16, 5},
		{10, -40, 16, 8},
		{10, -38, 16, 10},
		{0, -16, 16, 0},
	}
	for _, tc := range cases {
		if got := a.adjustTo(tc.target, tc.current, tc.imageDim); got != tc.want {
			t.Errorf("adjustTo(%d, %d, %d) = %d, want %d", tc.target, tc.current, tc.imageDim, got, tc.want)
		}
	}
}

func TestAbstractOutputDeviceFontSpecification(t *testing.T) {
	var out strings.Builder
	dev := newAbstractOutputDeviceTestDevice(&out)
	if dev.GetFontSpecification() != nil {
		t.Errorf("GetFontSpecification of a new device = %v", dev.GetFontSpecification())
	}
	fs := &FontSpecification{}
	dev.SetFontSpecification(fs)
	if dev.GetFontSpecification() != fs {
		t.Errorf("GetFontSpecification = %v, want %v", dev.GetFontSpecification(), fs)
	}
	if dev.AsAbstractOutputDevice() != &dev.AbstractOutputDevice {
		t.Errorf("AsAbstractOutputDevice does not return the embedded struct")
	}
}
