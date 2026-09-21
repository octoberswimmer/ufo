package geom

import (
	"math"
	"testing"
)

// The expected values in this file were produced by OpenJDK 25 running the
// same operations on the java.awt.geom classes.

// seg is a path segment: the type followed by its coordinates.
type seg []float64

// sinCosTolerance is the largest difference accepted in a coordinate computed
// from a sine or cosine. The JDK and Go use different implementations of those
// functions, and their results can differ in the last bit.
const sinCosTolerance = 1e-12

func collectDouble(pi PathIterator) []seg {
	var out []seg
	coords := make([]float64, 6)
	for ; !pi.IsDone(); pi.Next() {
		segType := pi.CurrentSegmentDouble(coords)
		s := seg{float64(segType)}
		s = append(s, coords[:segmentPointCounts[segType]*2]...)
		out = append(out, s)
	}
	return out
}

func collectFloat(pi PathIterator) []seg {
	var out []seg
	coords := make([]float32, 6)
	for ; !pi.IsDone(); pi.Next() {
		segType := pi.CurrentSegmentFloat(coords)
		s := seg{float64(segType)}
		for _, c := range coords[:segmentPointCounts[segType]*2] {
			s = append(s, float64(c))
		}
		out = append(out, s)
	}
	return out
}

func assertSegments(t *testing.T, name string, got, want []seg, tolerance float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: %d segments, want %d\ngot  %v\nwant %v", name, len(got), len(want), got, want)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) || got[i][0] != want[i][0] {
			t.Errorf("%s: segment %d = %v, want %v", name, i, got[i], want[i])
			continue
		}
		for j := 1; j < len(want[i]); j++ {
			if math.Abs(got[i][j]-want[i][j]) > tolerance {
				t.Errorf("%s: segment %d = %v, want %v", name, i, got[i], want[i])
				break
			}
		}
	}
}

func TestRectanglePathIterator(t *testing.T) {
	pi := NewRectangle(1, 2, 3, 4).GetPathIterator(nil)
	if pi.GetWindingRule() != PathIteratorWindNonZero {
		t.Errorf("winding rule = %d", pi.GetWindingRule())
	}
	assertSegments(t, "rect", collectDouble(pi), []seg{
		{0, 1.0, 2.0},
		{1, 4.0, 2.0},
		{1, 4.0, 6.0},
		{1, 1.0, 6.0},
		{1, 1.0, 2.0},
		{4},
	}, 0)
	assertSegments(t, "negative width", collectDouble(NewRectangle(1, 2, -3, 4).GetPathIterator(nil)), nil, 0)
	assertSegments(t, "zero width", collectDouble(NewRectangle(1, 2, 0, 4).GetPathIterator(nil)), []seg{
		{0, 1.0, 2.0},
		{1, 1.0, 2.0},
		{1, 1.0, 6.0},
		{1, 1.0, 6.0},
		{1, 1.0, 2.0},
		{4},
	}, 0)
}

func TestPathIteratorPanicsWhenDone(t *testing.T) {
	pi := NewRectangle(1, 2, -3, 4).GetPathIterator(nil)
	defer func() {
		if _, ok := recover().(*NoSuchElementException); !ok {
			t.Errorf("expected a *NoSuchElementException panic")
		}
	}()
	pi.CurrentSegmentDouble(make([]float64, 6))
}

func TestLine2DPathIteratorWithTransform(t *testing.T) {
	l := NewLine2DFloat(1.1, 2.2, 3.3, 4.4)
	assertSegments(t, "line", collectDouble(l.GetPathIterator(AffineTransformGetScaleInstance(2, 3))), []seg{
		{0, 2.200000047683716, 6.6000001430511475},
		{1, 6.599999904632568, 13.200000286102295},
	}, 0)
	if got := l.GetBounds(); !got.Equals(NewRectangle(1, 2, 3, 3)) {
		t.Errorf("GetBounds = %v", got)
	}
	if got := NewLine2DDouble(5, 1, 2, 7).GetBounds2D(); !got.Equals(NewRectangle2DDouble(2, 1, 3, 6)) {
		t.Errorf("GetBounds2D = %v", got)
	}
}

func TestEllipse2DPathIterator(t *testing.T) {
	e := NewEllipse2DFloat(10, 20, 30, 40)
	assertSegments(t, "ellipse", collectDouble(e.GetPathIterator(nil)), []seg{
		{0, 40.0, 40.0},
		{3, 40.0, 51.045694996615865, 33.2842712474619, 60.0, 25.0, 60.0},
		{3, 16.715728752538098, 60.0, 10.0, 51.045694996615865, 10.0, 40.0},
		{3, 10.0, 28.954305003384135, 16.715728752538098, 20.0, 25.0, 20.0},
		{3, 33.2842712474619, 20.0, 40.0, 28.954305003384135, 40.0, 40.0},
		{4},
	}, 0)
	assertSegments(t, "ellipse float, translated", collectFloat(e.GetPathIterator(AffineTransformGetTranslateInstance(0.1, 0.2))), []seg{
		{0, 40.099998474121094, 40.20000076293945},
		{3, 40.099998474121094, 51.245697021484375, 33.38426971435547, 60.20000076293945, 25.100000381469727, 60.20000076293945},
		{3, 16.81572914123535, 60.20000076293945, 10.100000381469727, 51.245697021484375, 10.100000381469727, 40.20000076293945},
		{3, 10.100000381469727, 29.154306411743164, 16.81572914123535, 20.200000762939453, 25.100000381469727, 20.200000762939453},
		{3, 33.38426971435547, 20.200000762939453, 40.099998474121094, 29.154306411743164, 40.099998474121094, 40.20000076293945},
		{4},
	}, 0)
	assertSegments(t, "negative height", collectDouble(NewEllipse2DDouble(0, 0, 5, -1).GetPathIterator(nil)), nil, 0)
	if got := NewEllipse2DDouble(0.5, 0.5, 2, 2).GetBounds(); !got.Equals(NewRectangle(0, 0, 3, 3)) {
		t.Errorf("GetBounds = %v", got)
	}
	if !NewEllipse2DDouble(0, 0, 10, 10).Contains(5, 5) || NewEllipse2DDouble(0, 0, 10, 10).Contains(1, 1) {
		t.Errorf("Contains is wrong")
	}
}

func TestArc2DPathIterator(t *testing.T) {
	tests := []struct {
		name string
		arc  *Arc2D
		want []seg
	}{
		{"90 degrees, open", NewArc2DDouble(0, 0, 100, 100, 0, 90, Arc2DOpen), []seg{
			{0, 100.0, 50.0},
			{3, 100.0, 22.38576250846033, 77.61423749153967, 0.0, 50.0, 0.0},
		}},
		{"-135 degrees, pie", NewArc2DDouble(10, 20, 100, 50, 30, -135, Arc2DPie), []seg{
			{0, 103.30127018922194, 32.5},
			{3, 113.41282630946668, 41.25686447192397, 111.97871772108589, 52.196998901679564, 99.66766701456176, 60.21903572521802},
			{3, 87.35661630803762, 68.24107254875648, 66.59307414590533, 71.765208956769, 47.05904774487396, 69.1481456572267},
			{1, 60.0, 45.0},
			{4},
		}},
		{"400 degrees, chord", NewArc2DDouble(0, 0, 10, 10, 45, 400, Arc2DChord), []seg{
			{0, 8.535533905932738, 1.4644660940672627},
			{3, 6.582912447176389, -0.4881553646890868, 3.417087552823612, -0.4881553646890868, 1.4644660940672627, 1.4644660940672622},
			{3, -0.4881553646890868, 3.4170875528236118, -0.4881553646890877, 6.582912447176387, 1.4644660940672614, 8.535533905932738},
			{3, 3.4170875528236113, 10.488155364689087, 6.582912447176387, 10.488155364689089, 8.535533905932738, 8.535533905932738},
			{3, 10.488155364689087, 6.582912447176389, 10.488155364689089, 3.417087552823613, 8.535533905932738, 1.4644660940672631},
			{4},
		}},
		{"-360 degrees, open", NewArc2DDouble(0, 0, 10, 10, 0, -360, Arc2DOpen), []seg{
			{0, 10.0, 5.0},
			{3, 10.0, 7.761423749153966, 7.761423749153967, 10.0, 5.0, 10.0},
			{3, 2.238576250846034, 10.0, 8.881784197001252e-16, 7.761423749153967, 0.0, 5.000000000000001},
			{3, 0.0, 2.238576250846034, 2.238576250846032, 8.881784197001252e-16, 4.999999999999999, 0.0},
			{3, 7.761423749153966, 0.0, 10.0, 2.238576250846032, 10.0, 4.999999999999999},
		}},
		{"0 degrees, pie", NewArc2DDouble(0, 0, 10, 10, 10, 0, Arc2DPie), []seg{
			{0, 9.92403876506104, 4.131759111665349},
			{1, 5.0, 5.0},
			{4},
		}},
	}
	for _, tc := range tests {
		pi := tc.arc.GetPathIterator(nil)
		if pi.GetWindingRule() != PathIteratorWindNonZero {
			t.Errorf("%s: winding rule = %d", tc.name, pi.GetWindingRule())
		}
		assertSegments(t, tc.name, collectDouble(pi), tc.want, sinCosTolerance)
	}

	assertSegments(t, "float", collectFloat(NewArc2DFloat(1.5, 2.5, 7.25, 9.75, 91, -46, Arc2DOpen).GetPathIterator(nil)), []seg{
		{0, 5.061735153198242, 2.5007424354553223},
		{3, 6.044938087463379, 2.4776628017425537, 6.992926597595215, 2.992748260498047, 7.688261985778809, 3.927854537963867},
	}, 0)
	assertSegments(t, "negative width", collectDouble(NewArc2DDouble(0, 0, -1, 10, 0, 90, Arc2DPie).GetPathIterator(nil)), nil, 0)
}

// The control points of a 90 degree arc of a circle lie at the distance
// 4/3*tan(pi/8) times the radius from the end points.
func TestArc2DQuarterCircleControlPointDistance(t *testing.T) {
	segs := collectDouble(NewArc2DDouble(-1, -1, 2, 2, 0, 90, Arc2DOpen).GetPathIterator(nil))
	if len(segs) != 2 {
		t.Fatalf("segments = %v", segs)
	}
	k := 4.0 / 3.0 * math.Tan(math.Pi/8)
	want := seg{3, 1, -k, k, -1, 0, -1}
	assertSegments(t, "unit quarter circle", segs[1:], []seg{want}, sinCosTolerance)
}

func TestArc2DBounds(t *testing.T) {
	pie := NewArc2DDouble(10, 20, 100, 50, 30, -135, Arc2DPie)
	b := pie.GetBounds2D()
	want := NewRectangle2DDouble(47.05904774487396, 32.5, 62.940952255126035, 37.5)
	if math.Abs(b.X-want.X) > sinCosTolerance || math.Abs(b.Y-want.Y) > sinCosTolerance ||
		math.Abs(b.Width-want.Width) > sinCosTolerance || math.Abs(b.Height-want.Height) > sinCosTolerance {
		t.Errorf("GetBounds2D = %v, want %v", b, want)
	}
	if got := pie.GetBounds(); !got.Equals(NewRectangle(10, 20, 100, 50)) {
		t.Errorf("GetBounds = %v", got)
	}
	start, end := pie.GetStartPoint(), pie.GetEndPoint()
	if math.Abs(start.X-103.30127018922194) > sinCosTolerance || math.Abs(start.Y-32.5) > sinCosTolerance ||
		math.Abs(end.X-47.05904774487396) > sinCosTolerance || math.Abs(end.Y-69.1481456572267) > sinCosTolerance {
		t.Errorf("start, end = %v, %v", start, end)
	}

	open := NewArc2DDouble(10, 20, 100, 50, 100, 200, Arc2DOpen)
	b = open.GetBounds2D()
	want = NewRectangle2DDouble(10.0, 20.3798061746948, 75.0, 49.6201938253052)
	if math.Abs(b.X-want.X) > sinCosTolerance || math.Abs(b.Y-want.Y) > sinCosTolerance ||
		math.Abs(b.Width-want.Width) > sinCosTolerance || math.Abs(b.Height-want.Height) > sinCosTolerance {
		t.Errorf("GetBounds2D = %v, want %v", b, want)
	}
	if !open.ContainsAngle(-90) || open.ContainsAngle(45) || !open.ContainsAngle(720+120) {
		t.Errorf("ContainsAngle is wrong")
	}
}

func TestArc2DRejectsUnknownType(t *testing.T) {
	defer func() {
		e, ok := recover().(*IllegalArgumentException)
		if !ok || e.Message != "invalid type for Arc: 3" {
			t.Errorf("panic = %v", e)
		}
	}()
	NewArc2DDouble(0, 0, 1, 1, 0, 90, 3)
}

// This builds a path the way BorderPainter.generateBorderShape does: arcs and
// lines appended with connect, then a translation, a quadrant rotation and
// another translation, all in a Path2D.Float.
func TestPath2DFloatBuiltLikeBorderPainter(t *testing.T) {
	p := NewPath2DFloat()
	if p.GetCurrentPoint() != nil {
		t.Errorf("GetCurrentPoint of an empty path = %v", p.GetCurrentPoint())
	}
	p.Append(NewArc2DFloat(-3, -2, 14.5, 9.25, 136, -46, Arc2DOpen), true)
	p.Append(NewArc2DFloat(80.3, -2, 10, 12, 90, -45, Arc2DOpen), true)
	p.LineTo(float64(float32(88.125)), float64(float32(5.1)))
	// The move of this line goes to the current point and is dropped.
	p.Append(NewLine2DFloat(88.125, 5.1, 0, 3.3), true)
	p.ClosePath()
	cp := p.GetCurrentPoint()
	if cp == nil || float32(cp.X) != -0.96521354 || float32(cp.Y) != -0.58779496 {
		t.Errorf("GetCurrentPoint after ClosePath = %v", cp)
	}
	p.Transform(AffineTransformGetTranslateInstance(-50.5, -20.25))
	p.Transform(AffineTransformGetRotateInstance(math.Pi / 2))
	p.Transform(AffineTransformGetTranslateInstance(50.5+7, 20.25+9))

	assertSegments(t, "border", collectDouble(p.GetPathIterator(nil)), []seg{
		{0, 78.33779907226562, -22.215213775634766},
		{3, 79.24029541015625, -20.84902572631836, 79.75, -18.966705322265625, 79.75, -17.0},
		{1, 79.75, 64.05000305175781},
		{3, 79.75, 65.37608337402344, 79.11785888671875, 66.64785766601562, 77.99264526367188, 67.58554077148438},
		{1, 72.6500015258789, 66.875},
		{1, 74.44999694824219, -21.25},
		{4},
	}, 0)
	assertSegments(t, "border, rotated, float", collectFloat(p.GetPathIterator(AffineTransformGetRotateInstance(0.3))), []seg{
		{0, 81.40399932861328, 1.9273982048034668},
		{3, 81.86245727539062, 3.4992735385894775, 81.79312896728516, 5.448150634765625, 81.21192932128906, 7.327016353607178},
		{1, 57.260013580322266, 84.75704193115234},
		{3, 56.86812973022461, 86.02389526367188, 55.888389587402344, 87.05205535888672, 54.53632736206055, 87.6153335571289},
		{1, 49.64228439331055, 85.35767364501953},
		{1, 77.40460205078125, 1.7005780935287476},
		{4},
	}, 0)
	if got, want := p.GetBounds2D(), NewRectangle2DDouble(72.6500015258789, -22.215213775634766, 7.099998474121094, 89.80075454711914); !got.Equals(want) {
		t.Errorf("GetBounds2D = %v, want %v", got, want)
	}
	if got := p.GetBounds(); !got.Equals(NewRectangle(72, -23, 8, 91)) {
		t.Errorf("GetBounds = %v", got)
	}
}

func buildMixedPath() *Path2D {
	q := NewPath2DDoubleWithRule(Path2DWindEvenOdd)
	q.MoveTo(0, 0)
	// A move directly after a move replaces it.
	q.MoveTo(1, 1)
	q.QuadTo(5, 9, 10, 1)
	q.CurveTo(12, -4, 3, -6, 1, 1)
	q.ClosePath()
	// A second close is dropped.
	q.ClosePath()
	q.MoveTo(20, 20)
	q.LineTo(30, 20)
	q.LineTo(30, 30)
	return q
}

func TestPath2DSegmentsAndBounds(t *testing.T) {
	q := buildMixedPath()
	pi := q.GetPathIterator(nil)
	if pi.GetWindingRule() != PathIteratorWindEvenOdd {
		t.Errorf("winding rule = %d", pi.GetWindingRule())
	}
	assertSegments(t, "q", collectDouble(pi), []seg{
		{0, 1.0, 1.0},
		{2, 5.0, 9.0, 10.0, 1.0},
		{3, 12.0, -4.0, 3.0, -6.0, 1.0, 1.0},
		{4},
		{0, 20.0, 20.0},
		{1, 30.0, 20.0},
		{1, 30.0, 30.0},
	}, 0)
	if cp := q.GetCurrentPoint(); !cp.Equals(NewPoint2DDouble(30, 30)) {
		t.Errorf("GetCurrentPoint = %v", cp)
	}
	// The bounds follow the curves, not their control points.
	want := NewRectangle2DDouble(1.0, -3.5308289965298663, 29.0, 33.53082899652986)
	if got := q.GetBounds2D(); !got.Equals(want) {
		t.Errorf("GetBounds2D = %v, want %v", got, want)
	}
	if got := NewPath2DDouble().GetBounds2D(); !got.Equals(NewRectangle2D()) {
		t.Errorf("GetBounds2D of an empty path = %v", got)
	}

	c := q.Clone()
	c.Reset()
	if c.GetCurrentPoint() != nil || c.GetWindingRule() != Path2DWindEvenOdd || q.GetCurrentPoint() == nil {
		t.Errorf("Clone/Reset is wrong")
	}
}

func TestShapeIntersectsMatchesJDK(t *testing.T) {
	q := buildMixedPath()
	ellipse := NewEllipse2DDouble(0, -5, 12, 8)
	line := NewLine2DDouble(0, 0, 30, 25)
	tests := []struct {
		rect                [4]float64
		path, ellipse, line bool
	}{
		{[4]float64{4.0, 3.0, 1.0, 1.0}, true, false, true},
		{[4]float64{4.0, 5.4, 1.0, 1.0}, false, false, false},
		{[4]float64{4.0, 5.6, 1.0, 1.0}, false, false, false},
		{[4]float64{0.0, 0.0, 1.0, 1.0}, false, true, true},
		{[4]float64{11.0, -3.0, 0.4, 0.4}, false, true, false},
		{[4]float64{10.2, -2.5, 0.4, 0.4}, false, true, false},
		{[4]float64{25.0, 21.0, 2.0, 2.0}, true, false, true},
		{[4]float64{22.0, 27.0, 2.0, 2.0}, false, false, false},
		{[4]float64{28.0, 22.0, 1.0, 1.0}, true, false, false},
		{[4]float64{-5.0, -10.0, 50.0, 50.0}, true, true, true},
		{[4]float64{4.0, 3.0, 0.0, 1.0}, false, false, false},
		{[4]float64{19.0, 19.0, 2.0, 2.0}, true, false, false},
	}
	for _, tc := range tests {
		r := tc.rect
		if got := q.IntersectsWithXYWH(r[0], r[1], r[2], r[3]); got != tc.path {
			t.Errorf("path intersects %v = %v, want %v", r, got, tc.path)
		}
		if got := ellipse.IntersectsWithXYWH(r[0], r[1], r[2], r[3]); got != tc.ellipse {
			t.Errorf("ellipse intersects %v = %v, want %v", r, got, tc.ellipse)
		}
		if got := line.IntersectsWithXYWH(r[0], r[1], r[2], r[3]); got != tc.line {
			t.Errorf("line intersects %v = %v, want %v", r, got, tc.line)
		}
	}
	if !q.Intersects(NewRectangle(4, 3, 1, 1)) || q.Intersects(NewRectangle(22, 27, 2, 2)) {
		t.Errorf("Intersects(*Rectangle) is wrong")
	}
}

func TestPath2DIntersectsHonoursWindingRule(t *testing.T) {
	tests := []struct {
		rule                         int
		inHole, inRing, across, away bool
	}{
		{Path2DWindEvenOdd, false, true, true, false},
		{Path2DWindNonZero, true, true, true, false},
	}
	for _, tc := range tests {
		h := NewPath2DDoubleWithRule(tc.rule)
		h.Append(NewRectangle(0, 0, 30, 30), false)
		h.Append(NewRectangle(10, 10, 10, 10), false)
		got := [4]bool{
			h.IntersectsWithXYWH(12, 12, 5, 5),
			h.IntersectsWithXYWH(2, 2, 5, 5),
			h.IntersectsWithXYWH(8, 8, 5, 5),
			h.IntersectsWithXYWH(40, 40, 5, 5),
		}
		want := [4]bool{tc.inHole, tc.inRing, tc.across, tc.away}
		if got != want {
			t.Errorf("rule %d: intersects = %v, want %v", tc.rule, got, want)
		}
	}
}

func TestPath2DRequiresInitialMoveTo(t *testing.T) {
	for name, f := range map[string]func(p *Path2D){
		"LineTo":    func(p *Path2D) { p.LineTo(1, 1) },
		"QuadTo":    func(p *Path2D) { p.QuadTo(1, 1, 2, 2) },
		"CurveTo":   func(p *Path2D) { p.CurveTo(1, 1, 2, 2, 3, 3) },
		"ClosePath": func(p *Path2D) { p.ClosePath() },
	} {
		func() {
			defer func() {
				e, ok := recover().(*IllegalPathStateException)
				if !ok || e.Message != "missing initial moveto in path definition" {
					t.Errorf("%s: panic = %v", name, e)
				}
			}()
			f(NewPath2DFloat())
		}()
	}
}

func TestPath2DRejectsUnknownWindingRule(t *testing.T) {
	defer func() {
		if _, ok := recover().(*IllegalArgumentException); !ok {
			t.Errorf("expected an *IllegalArgumentException panic")
		}
	}()
	NewPath2DDoubleWithRule(2)
}

func TestPath2DAppendWithoutConnectStartsASubpath(t *testing.T) {
	p := NewGeneralPath()
	p.MoveTo(0, 0)
	p.LineTo(5, 5)
	p.Append(NewLine2DDouble(5, 5, 9, 9), false)
	assertSegments(t, "no connect", collectDouble(p.GetPathIterator(nil)), []seg{
		{0, 0, 0}, {1, 5, 5}, {0, 5, 5}, {1, 9, 9},
	}, 0)

	p = NewGeneralPath()
	p.MoveTo(0, 0)
	p.LineTo(5, 5)
	p.Append(NewLine2DDouble(6, 6, 9, 9), true)
	assertSegments(t, "connect to a different point", collectDouble(p.GetPathIterator(nil)), []seg{
		{0, 0, 0}, {1, 5, 5}, {1, 6, 6}, {1, 9, 9},
	}, 0)
}

func TestPath2DFloatRoundsStoredCoordinates(t *testing.T) {
	f := NewPath2DFloat()
	f.MoveTo(0.1, 0.2)
	d := NewPath2DDouble()
	d.MoveTo(0.1, 0.2)
	if cp := f.GetCurrentPoint(); cp.X != float64(float32(0.1)) || cp.Y != float64(float32(0.2)) {
		t.Errorf("Path2D.Float stored %v", cp)
	}
	if cp := d.GetCurrentPoint(); cp.X != 0.1 || cp.Y != 0.2 {
		t.Errorf("Path2D.Double stored %v", cp)
	}
	// A Double path made from a Float path and a transform keeps full precision.
	fromFloat := NewPath2DDoubleFromShapeWithAt(f, AffineTransformGetScaleInstance(0.1, 0.1))
	if cp := fromFloat.GetCurrentPoint(); cp.X != float64(float32(0.1))*0.1 {
		t.Errorf("Path2D.Double from Path2D.Float stored %v", cp)
	}
	transformed, ok := f.CreateTransformedShape(AffineTransformGetScaleInstance(0.1, 0.1)).(*Path2D)
	if !ok {
		t.Fatalf("CreateTransformedShape did not return a *Path2D")
	}
	if cp := transformed.GetCurrentPoint(); cp.X != float64(float32(float64(float32(0.1))*0.1)) {
		t.Errorf("transformed Path2D.Float stored %v", cp)
	}
}
