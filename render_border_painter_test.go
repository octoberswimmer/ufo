// Tests for render_border_painter.go. Flying Saucer has no JUnit test for
// BorderPainter. The expected values here were produced by running the Java
// BorderPainter (flying-saucer-core 10.6.0-SNAPSHOT on OpenJDK 25) over the same
// cases and dumping every path segment as the hexadecimal bits of its double
// coordinates. The Go output matched the Java output bit for bit; the
// testdata files hold one SHA-256 hash of that dump per group of cases.

package ufo

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/geom"
)

// borderPainterTestDumpShape writes one line per path segment: the segment
// type followed by the bits of each coordinate in hexadecimal.
func borderPainterTestDumpShape(out *strings.Builder, s geom.Shape) {
	c := make([]float64, 6)
	for it := s.GetPathIterator(nil); !it.IsDone(); it.Next() {
		segType := it.CurrentSegmentDouble(c)
		n := 2
		switch segType {
		case geom.PathIteratorSegClose:
			n = 0
		case geom.PathIteratorSegCubicto:
			n = 6
		case geom.PathIteratorSegQuadto:
			n = 4
		}
		fmt.Fprintf(out, "%d", segType)
		for i := 0; i < n; i++ {
			fmt.Fprintf(out, " %x", math.Float64bits(c[i]))
		}
		out.WriteString("\n")
	}
}

func borderPainterTestReadHashes(t *testing.T, path string, keyFields int) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	result := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != keyFields+1 {
			t.Fatalf("%s: malformed line %q", path, scanner.Text())
		}
		result[strings.Join(fields[:keyFields], " ")] = fields[keyFields]
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func borderPainterTestHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

var borderPainterTestBounds = [][]int{{0, 0, 100, 50}, {10, 20, 200, 100}, {-5, 7, 33, 91}, {3, 4, 10, 10}}
var borderPainterTestWidths = [][]float32{{0, 0, 0, 0}, {1, 1, 1, 1}, {2, 4, 6, 8}, {5, 0, 5, 0}, {0.5, 1.5, 2.25, 3.75}, {10, 10, 10, 10}, {3, 1, 0, 7}}

// Each entry is the (left, right) radius pair of the top left, top right,
// bottom right and bottom left corner. The entry of 60s exceeds every test
// bounds, so BorderPropertySet.NormalizedInstance scales it down.
var borderPainterTestRadii = [][]float32{{0, 0, 0, 0, 0, 0, 0, 0}, {5, 5, 5, 5, 5, 5, 5, 5}, {10, 20, 5, 5, 0, 0, 8, 3}, {60, 60, 60, 60, 60, 60, 60, 60}, {2.5, 7.25, 0, 4, 12, 1, 6, 6}}
var borderPainterTestSides = []int{BorderPainterTop, BorderPainterRight, BorderPainterBottom, BorderPainterLeft}

type borderPainterTestShapeParams struct {
	drawInterior bool
	scaledOffset float32
	widthScale   float32
	overlap      bool
}

var borderPainterTestThird = float32(1) / 3.0

var borderPainterTestParams = []borderPainterTestShapeParams{
	{false, 0, 1, true}, {true, 0, 1, true}, {false, 0.5, 1, true}, {true, 0, 1, false},
	{true, 1, 0.5, true}, {true, 2, borderPainterTestThird, true}, {true, 0, borderPainterTestThird, true}, {false, 1, 1, false}}

func borderPainterTestBorder(widths []float32, radii []float32, styles *borderPropertySetStyles, colors *borderPropertySetColors) *BorderPropertySet {
	corners := &borderPropertySetCorners{
		NewBorderRadiusCorner(radii[0], radii[1]), NewBorderRadiusCorner(radii[2], radii[3]),
		NewBorderRadiusCorner(radii[4], radii[5]), NewBorderRadiusCorner(radii[6], radii[7])}
	return newBorderPropertySetWithSides(widths[0], widths[1], widths[2], widths[3], styles, corners, colors)
}

// TestBorderPainterGenerateBorderShapeAgainstJava covers generateBorderBounds
// (inside and outside) and generateBorderShape for every side and eight
// parameter combinations, over 4 bounds x 7 width sets x 5 radius sets: 4760
// paths.
func TestBorderPainterGenerateBorderShapeAgainstJava(t *testing.T) {
	want := borderPainterTestReadHashes(t, "testdata/render/border_painter_paths.sha256", 3)
	checked := 0
	for bi, b := range borderPainterTestBounds {
		bounds := geom.NewRectangle(b[0], b[1], b[2], b[3])
		for wi, w := range borderPainterTestWidths {
			for ri, r := range borderPainterTestRadii {
				border := borderPainterTestBorder(w, r, borderPropertySetNoStyles, borderPropertySetNoColors)
				var out strings.Builder
				for inside := 0; inside < 2; inside++ {
					fmt.Fprintf(&out, "bounds %d %d %d %d\n", bi, wi, ri, inside)
					borderPainterTestDumpShape(&out, BorderPainterGenerateBorderBounds(bounds, border, inside == 1))
				}
				for _, side := range borderPainterTestSides {
					for pi, p := range borderPainterTestParams {
						fmt.Fprintf(&out, "shape %d %d %d %d %d\n", bi, wi, ri, side, pi)
						borderPainterTestDumpShape(&out, BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, side, border, p.drawInterior, p.scaledOffset, p.widthScale, p.overlap))
					}
				}
				key := fmt.Sprintf("%d %d %d", bi, wi, ri)
				if got := borderPainterTestHash(out.String()); got != want[key] {
					t.Errorf("bounds %v widths %v radii %v: dump hash %s, Java gives %s", b, w, r, got, want[key])
				}
				checked++
			}
		}
	}
	if checked != len(want) {
		t.Errorf("checked %d groups, testdata has %d", checked, len(want))
	}
}

type borderPainterTestSeg struct {
	segType int
	coords  []float64
}

func borderPainterTestSegments(s geom.Shape) []borderPainterTestSeg {
	var result []borderPainterTestSeg
	c := make([]float64, 6)
	for it := s.GetPathIterator(nil); !it.IsDone(); it.Next() {
		segType := it.CurrentSegmentDouble(c)
		n := 2
		switch segType {
		case geom.PathIteratorSegClose:
			n = 0
		case geom.PathIteratorSegCubicto:
			n = 6
		case geom.PathIteratorSegQuadto:
			n = 4
		}
		result = append(result, borderPainterTestSeg{segType, append([]float64(nil), c[:n]...)})
	}
	return result
}

func borderPainterTestAssertSegments(t *testing.T, name string, got []borderPainterTestSeg, want []borderPainterTestSeg) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: %d segments, want %d: %v", name, len(got), len(want), got)
		return
	}
	for i := range want {
		if got[i].segType != want[i].segType || len(got[i].coords) != len(want[i].coords) {
			t.Errorf("%s: segment %d = %v, want %v", name, i, got[i], want[i])
			continue
		}
		for j := range want[i].coords {
			if got[i].coords[j] != want[i].coords[j] {
				t.Errorf("%s: segment %d = %v, want %v", name, i, got[i], want[i])
				break
			}
		}
	}
}

// The expected segments are the Java output for the same arguments.
func TestBorderPainterGenerateBorderShapeSegments(t *testing.T) {
	// A one pixel top border without radii, not drawn as an interior: a line.
	border := borderPainterTestBorder([]float32{1, 1, 1, 1}, borderPainterTestRadii[0], borderPropertySetNoStyles, borderPropertySetNoColors)
	borderPainterTestAssertSegments(t, "top line",
		borderPainterTestSegments(BorderPainterGenerateBorderShape(geom.NewRectangle(0, 0, 100, 50), BorderPainterTop, border, false)),
		[]borderPainterTestSeg{
			{0, []float64{0.0, 0.0}},
			{1, []float64{100.0, 0.0}},
		})

	// The right side of a border with differing widths and radii, with its
	// interior.
	border = borderPainterTestBorder([]float32{2, 4, 6, 8}, borderPainterTestRadii[2], borderPropertySetNoStyles, borderPropertySetNoColors)
	borderPainterTestAssertSegments(t, "right side",
		borderPainterTestSegments(BorderPainterGenerateBorderShape(geom.NewRectangle(10, 20, 200, 100), BorderPainterRight, border, true)),
		[]borderPainterTestSeg{
			{0, []float64{207.42404174804688, 20.62689971923828}},
			{3, []float64{209.01370239257812, 21.50806427001953, 210.0, 23.182456970214844, 210.0, 25.0}},
			{1, []float64{210.0, 120.0}},
			{1, []float64{206.0, 114.0}},
			{1, []float64{206.0, 25.0}},
			{3, []float64{206.0, 23.909473419189453, 205.802734375, 22.90483856201172, 205.48480224609375, 22.376140594482422}},
			{4, []float64{}},
		})

	// The inner bounds of a border with fractional widths.
	border = borderPainterTestBorder([]float32{0.5, 1.5, 2.25, 3.75}, borderPainterTestRadii[4], borderPropertySetNoStyles, borderPropertySetNoColors)
	borderPainterTestAssertSegments(t, "inner bounds",
		borderPainterTestSegments(BorderPainterGenerateBorderBounds(geom.NewRectangle(10, 20, 200, 100), border, true)),
		[]borderPainterTestSeg{
			{0, []float64{16.60688018798828, 20.534053802490234}},
			{3, []float64{16.818954467773438, 20.511398315429688, 17.03424835205078, 20.5, 17.25, 20.5}},
			{1, []float64{208.5, 20.5}},
			{1, []float64{208.5, 117.75}},
			{1, []float64{16.0, 117.75}},
			{3, []float64{15.55499267578125, 117.75, 15.119979858398438, 117.53006744384766, 14.749969482421875, 117.11801147460938}},
			{3, []float64{14.125152587890625, 116.42219543457031, 13.75, 115.25243377685547, 13.75, 114.0}},
			{1, []float64{13.75, 22.5}},
			{3, []float64{13.75, 21.537155151367188, 14.950592041015625, 20.710975646972656, 16.60688018798828, 20.53404998779297}},
		})

	// The left side of the inner third of a double border whose radii are
	// scaled down to fit the bounds.
	border = borderPainterTestBorder([]float32{3, 1, 0, 7}, borderPainterTestRadii[3], borderPropertySetNoStyles, borderPropertySetNoColors)
	borderPainterTestAssertSegments(t, "left side, inner third",
		borderPainterTestSegments(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(geom.NewRectangle(-5, 7, 33, 91), BorderPainterLeft, border, true, 2, borderPainterTestThird, true)),
		[]borderPainterTestSeg{
			{0, []float64{11.70651912689209, 97.99542999267578}},
			{3, []float64{8.532808303833008, 98.13587951660156, 5.469902038574219, 95.03790283203125, 3.205906867980957, 89.39751434326172}},
			{3, []float64{0.9419107437133789, 83.75711822509766, -0.33333396911621094, 76.04725646972656, -0.33333396911621094, 68.0}},
			{1, []float64{-0.33333396911621094, 37.0}},
			{3, []float64{-0.33333396911621094, 26.294837951660156, 2.246278762817383, 16.52660369873047, 6.312607288360596, 11.833763122558594}},
			{1, []float64{7.335474014282227, 12.732559204101562}},
			{3, []float64{4.070957183837891, 17.257789611816406, 2.0, 26.677169799804688, 2.0, 37.0}},
			{1, []float64{2.0, 68.0}},
			{3, []float64{2.0, 76.04725646972656, 3.0237884521484375, 83.75711822509766, 4.841361999511719, 89.39751434326172}},
			{3, []float64{6.658936023712158, 95.03790283203125, 9.117889404296875, 98.13587951660156, 11.66579818725586, 97.99542999267578}},
			{4, []float64{}},
		})
}

func TestBorderPainterGenerateBorderShapeNoSide(t *testing.T) {
	defer func() {
		e, ok := recover().(*XRRuntimeException)
		if !ok || e.Error() != "No side found" {
			t.Errorf("panic = %v", e)
		}
	}()
	BorderPainterGenerateBorderShape(geom.NewRectangle(0, 0, 10, 10), 0, BorderPropertySetEmptyBorder, false)
}

// borderPainterTestDevice records the OutputDevice calls that
// BorderPainterPaint makes. A call of any other method panics on the nil
// embedded interface.
type borderPainterTestDevice struct {
	OutputDevice
	out           strings.Builder
	initialStroke geom.Stroke
}

func borderPainterTestFloatBits(f float32) string {
	return fmt.Sprintf("%x", math.Float32bits(f))
}

func (d *borderPainterTestDevice) SetColor(color FSColor) {
	fmt.Fprintf(&d.out, "setColor %s\n", color.ToString())
}

func (d *borderPainterTestDevice) SetStroke(s geom.Stroke) {
	stroke := s.(*geom.BasicStroke)
	fmt.Fprintf(&d.out, "setStroke %s %d %d %s", borderPainterTestFloatBits(stroke.GetLineWidth()), stroke.GetEndCap(), stroke.GetLineJoin(), borderPainterTestFloatBits(stroke.GetMiterLimit()))
	for _, dash := range stroke.GetDashArray() {
		fmt.Fprintf(&d.out, " d%s", borderPainterTestFloatBits(dash))
	}
	fmt.Fprintf(&d.out, " %s\n", borderPainterTestFloatBits(stroke.GetDashPhase()))
}

func (d *borderPainterTestDevice) GetStroke() geom.Stroke {
	d.out.WriteString("getStroke\n")
	return d.initialStroke
}

func (d *borderPainterTestDevice) GetClip() geom.Shape {
	d.out.WriteString("getClip\n")
	return nil
}

func (d *borderPainterTestDevice) SetClip(s geom.Shape) {
	if s == nil {
		d.out.WriteString("setClip null\n")
	} else {
		d.out.WriteString("setClip area\n")
	}
}

func (d *borderPainterTestDevice) Fill(s geom.Shape) {
	d.out.WriteString("fill\n")
	borderPainterTestDumpShape(&d.out, s)
}

func (d *borderPainterTestDevice) Draw(s geom.Shape) {
	d.out.WriteString("draw\n")
	borderPainterTestDumpShape(&d.out, s)
}

func (d *borderPainterTestDevice) DrawBorderLine(bounds geom.Shape, side int, width int, solid bool) {
	fmt.Fprintf(&d.out, "drawBorderLine %d %d %t\n", side, width, solid)
	borderPainterTestDumpShape(&d.out, bounds)
}

// TestBorderPainterPaintAgainstJava compares the sequence of output device
// calls (colors, strokes, clips, filled and drawn paths) that paint makes with
// the sequence the Java BorderPainter.paint makes on a recording OutputDevice,
// for every border style, four width sets, two color sets, four side masks and
// two pattern offsets: 640 calls of paint.
func TestBorderPainterPaintAgainstJava(t *testing.T) {
	want := borderPainterTestReadHashes(t, "testdata/render/border_painter_paint.sha256", 2)

	styles := []*IdentValue{IdentValueNone, IdentValueHidden, IdentValueSolid, IdentValueDashed, IdentValueDotted, IdentValueDouble, IdentValueRidge, IdentValueGroove, IdentValueInset, IdentValueOutset}
	widths := [][]float32{{1, 1, 1, 1}, {3, 2, 1, 4}, {0, 2, 0.5, 2}, {6, 6, 6, 6}}
	colors := []*borderPropertySetColors{
		{NewFSRGBColor(255, 0, 0), NewFSRGBColor(0, 255, 0), NewFSRGBColor(0, 0, 255), NewFSRGBColor(10, 20, 30)},
		{NewFSRGBColor(255, 0, 0), FSRGBColorTransparent, NewFSRGBColor(0, 0, 255), FSRGBColorTransparent}}
	sideMasks := []int{BorderPainterAll, BorderPainterTop | BorderPainterBottom, BorderPainterLeft, 0}
	xOffsets := []int{0, 7}
	bounds := geom.NewRectangle(10, 20, 120, 60)
	radii := []float32{8, 8, 0, 0, 4, 12, 0, 0}

	checked := 0
	for si := range styles {
		// The left side takes the next style, so that one border mixes two styles.
		st := &borderPropertySetStyles{styles[si], styles[si], styles[si], styles[(si+1)%len(styles)]}
		for wi, w := range widths {
			device := &borderPainterTestDevice{initialStroke: geom.NewBasicStrokeWithWidth(3.0)}
			ctx := NewRenderingContext(nil, device, nil, nil, 0)
			for ci, cl := range colors {
				border := borderPainterTestBorder(w, radii, st, cl)
				for mi, mask := range sideMasks {
					for _, xOffset := range xOffsets {
						fmt.Fprintf(&device.out, "paint %d %d %d %d %d\n", si, wi, ci, mi, xOffset)
						BorderPainterPaint(bounds, mask, border, ctx, xOffset)
					}
				}
			}
			key := fmt.Sprintf("%d %d", si, wi)
			if got := borderPainterTestHash(device.out.String()); got != want[key] {
				t.Errorf("style %v widths %v: call log hash %s, Java gives %s\n%s", styles[si], w, got, want[key], device.out.String())
			}
			checked++
		}
	}
	if checked != len(want) {
		t.Errorf("checked %d groups, testdata has %d", checked, len(want))
	}
}
