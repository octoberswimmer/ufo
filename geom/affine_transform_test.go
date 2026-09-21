package geom

import (
	"errors"
	"math"
	"testing"
)

// The expected values in this file were produced by OpenJDK 25 running the
// same operations on java.awt.geom.AffineTransform. Matrices are in GetMatrix
// order: m00, m10, m01, m11, m02, m12.

func assertMatrix(t *testing.T, name string, at *AffineTransform, want [6]float64, tolerance float64) {
	t.Helper()
	var got [6]float64
	at.GetMatrix(got[:])
	for i := range want {
		if math.Abs(got[i]-want[i]) > tolerance {
			t.Errorf("%s: matrix = %v, want %v", name, got, want)
			return
		}
	}
}

func TestAffineTransformInstances(t *testing.T) {
	tests := []struct {
		name      string
		at        *AffineTransform
		want      [6]float64
		tolerance float64
		str       string
	}{
		{"identity", NewAffineTransform(), [6]float64{1, 0, 0, 1, 0, 0}, 0,
			"AffineTransform[[1.0, 0.0, 0.0], [0.0, 1.0, 0.0]]"},
		{"translate", AffineTransformGetTranslateInstance(3, 4), [6]float64{1, 0, 0, 1, 3, 4}, 0,
			"AffineTransform[[1.0, 0.0, 3.0], [0.0, 1.0, 4.0]]"},
		{"scale", AffineTransformGetScaleInstance(2, -1), [6]float64{2, 0, 0, -1, 0, 0}, 0,
			"AffineTransform[[2.0, 0.0, 0.0], [0.0, -1.0, 0.0]]"},
		// Quadrant rotations have exact entries.
		{"rotate 90", AffineTransformGetRotateInstance(math.Pi / 2), [6]float64{0, 1, -1, 0, 0, 0}, 0,
			"AffineTransform[[0.0, -1.0, 0.0], [1.0, 0.0, 0.0]]"},
		{"rotate 180", AffineTransformGetRotateInstance(math.Pi), [6]float64{-1, 0, 0, -1, 0, 0}, 0,
			"AffineTransform[[-1.0, -0.0, 0.0], [0.0, -1.0, 0.0]]"},
		{"rotate 270", AffineTransformGetRotateInstance(-math.Pi / 2), [6]float64{0, -1, 1, 0, 0, 0}, 0,
			"AffineTransform[[0.0, 1.0, 0.0], [-1.0, 0.0, 0.0]]"},
		{"rotate 30", AffineTransformGetRotateInstance(math.Pi / 6),
			[6]float64{0.8660254037844387, 0.49999999999999994, -0.49999999999999994, 0.8660254037844387, 0, 0}, sinCosTolerance,
			"AffineTransform[[0.866025403784439, -0.5, 0.0], [0.5, 0.866025403784439, 0.0]]"},
		{"rotate 30 around (10, 20)", AffineTransformGetRotateInstanceWithAnchorxAnchory(math.Pi/6, 10, 20),
			[6]float64{0.8660254037844387, 0.49999999999999994, -0.49999999999999994, 0.8660254037844387, 11.339745962155611, -2.3205080756887733}, sinCosTolerance,
			// The last digits depend on the sine and cosine implementation.
			""},
		{"shear", AffineTransformGetShearInstance(0.5, 0.25), [6]float64{1, 0.25, 0.5, 1, 0, 0}, 0,
			"AffineTransform[[1.0, 0.5, 0.0], [0.25, 1.0, 0.0]]"},
		{"six values", NewAffineTransformWithM00M10M01M11M02M12(1, 2, 3, 4, 5, 6), [6]float64{1, 2, 3, 4, 5, 6}, 0,
			"AffineTransform[[1.0, 3.0, 5.0], [2.0, 4.0, 6.0]]"},
		{"flat matrix of four", NewAffineTransformWithFlatmatrix([]float64{1, 2, 3, 4}), [6]float64{1, 2, 3, 4, 0, 0}, 0,
			"AffineTransform[[1.0, 3.0, 0.0], [2.0, 4.0, 0.0]]"},
		{"flat matrix of six", NewAffineTransformWithFlatmatrix([]float64{1, 2, 3, 4, 5, 6}), [6]float64{1, 2, 3, 4, 5, 6}, 0,
			"AffineTransform[[1.0, 3.0, 5.0], [2.0, 4.0, 6.0]]"},
	}
	for _, tc := range tests {
		assertMatrix(t, tc.name, tc.at, tc.want, tc.tolerance)
		if got := tc.at.String(); tc.str != "" && got != tc.str {
			t.Errorf("%s: String = %q, want %q", tc.name, got, tc.str)
		}
	}
}

func TestAffineTransformMutatorsMatchJDK(t *testing.T) {
	at := NewAffineTransform()
	if !at.IsIdentity() {
		t.Errorf("a new transform is not the identity")
	}
	at.Translate(10, 20)
	at.Scale(2, 3)
	assertMatrix(t, "translate, scale", at, [6]float64{2.0, 0.0, 0.0, 3.0, 10.0, 20.0}, 0)
	if at.IsIdentity() || at.GetDeterminant() != 6 {
		t.Errorf("IsIdentity or GetDeterminant is wrong: %v", at.GetDeterminant())
	}
	at.Rotate(math.Pi / 2)
	assertMatrix(t, "quadrant rotate", at, [6]float64{0.0, 3.0, -2.0, 0.0, 10.0, 20.0}, 0)
	at.Rotate(0.7)
	assertMatrix(t, "rotate", at, [6]float64{-1.288435374475382, 2.2945265618534654, -1.529684374568977, -1.932653061713073, 10.0, 20.0}, sinCosTolerance)
	at.Shear(0.1, 0.2)
	assertMatrix(t, "shear", at, [6]float64{-1.5943722493891774, 1.9079959495108507, -1.6585279120165153, -1.7032004055277266, 10.0, 20.0}, sinCosTolerance)
	at.RotateWithAnchorxAnchory(0.3, 5, 6)
	anchor := [6]float64{-2.0132904984034723, 1.319448015844342, -1.1132830159480853, -2.1909808530026122, 8.823121868660895, 25.86942235318186}
	assertMatrix(t, "rotate around anchor", at, anchor, sinCosTolerance)
	if got := at.GetDeterminant(); math.Abs(got-5.88) > sinCosTolerance {
		t.Errorf("GetDeterminant = %v", got)
	}

	if at.GetScaleX() != at.m00 || at.GetShearY() != at.m10 || at.GetShearX() != at.m01 ||
		at.GetScaleY() != at.m11 || at.GetTranslateX() != at.m02 || at.GetTranslateY() != at.m12 {
		t.Errorf("the element getters do not match the matrix")
	}

	inv, err := at.CreateInverse()
	if err != nil {
		t.Fatalf("CreateInverse: %v", err)
	}
	assertMatrix(t, "inverse", inv, [6]float64{-0.37261579132697487, -0.22439592106196293, 0.18933384624967436, -0.34239634326589663, -1.6103226973198073, 10.837468174490624}, sinCosTolerance)

	p := at.Transform(NewPoint2DDouble(3, 4), nil)
	if math.Abs(p.X - -1.6698816903418638) > sinCosTolerance || math.Abs(p.Y-21.063842988704437) > sinCosTolerance {
		t.Errorf("Transform = %v", p)
	}
	d := at.DeltaTransform(NewPoint2DDouble(3, 4), nil)
	if math.Abs(d.X - -10.493003559002759) > sinCosTolerance || math.Abs(d.Y - -4.805579364477423) > sinCosTolerance {
		t.Errorf("DeltaTransform = %v", d)
	}
	ip, err := at.InverseTransform(NewPoint2DDouble(3, 4), nil)
	if err != nil || math.Abs(ip.X - -1.9708346863020345) > sinCosTolerance || math.Abs(ip.Y-8.79469503824115) > sinCosTolerance {
		t.Errorf("InverseTransform = %v, %v", ip, err)
	}
	// Transform may write into its source.
	src := NewPoint2DDouble(3, 4)
	if got := at.Transform(src, src); got != src || !src.Equals(p) {
		t.Errorf("Transform in place = %v, want %v", src, p)
	}
}

// Concatenate applies its argument first: T.concatenate(S) maps a point
// through S and then through T. PreConcatenate applies its argument last.
func TestAffineTransformConcatenationOrder(t *testing.T) {
	c := AffineTransformGetTranslateInstance(10, 20)
	c.Concatenate(AffineTransformGetScaleInstance(2, 3))
	assertMatrix(t, "T.Concatenate(S)", c, [6]float64{2.0, 0.0, 0.0, 3.0, 10.0, 20.0}, 0)
	if p := c.Transform(NewPoint2DDouble(1, 1), nil); !p.Equals(NewPoint2DDouble(12, 23)) {
		t.Errorf("T.Concatenate(S) maps (1, 1) to %v", p)
	}

	c = AffineTransformGetTranslateInstance(10, 20)
	c.PreConcatenate(AffineTransformGetScaleInstance(2, 3))
	assertMatrix(t, "T.PreConcatenate(S)", c, [6]float64{2.0, 0.0, 0.0, 3.0, 20.0, 60.0}, 0)
	if p := c.Transform(NewPoint2DDouble(1, 1), nil); !p.Equals(NewPoint2DDouble(22, 63)) {
		t.Errorf("T.PreConcatenate(S) maps (1, 1) to %v", p)
	}

	general := NewAffineTransformWithM00M10M01M11M02M12(-2.0132904984034723, 1.319448015844342, -1.1132830159480853, -2.1909808530026122, 8.823121868660895, 25.86942235318186)
	c.Concatenate(general)
	assertMatrix(t, "general Concatenate", c, [6]float64{-4.026580996806945, 3.958344047533026, -2.2265660318961706, -6.572942559007837, 37.64624373732179, 137.6082670595456}, 0)
	c.PreConcatenate(AffineTransformGetRotateInstance(1.1))
	assertMatrix(t, "general PreConcatenate", c, [6]float64{-5.35414687157437, -1.7930291130088587, 4.847893069682788, -4.965793286407716, -105.56131026327552, 95.96918571167282}, sinCosTolerance)
}

func TestAffineTransformCreateInverse(t *testing.T) {
	tests := []struct {
		name string
		at   *AffineTransform
		want [6]float64
	}{
		{"scale and translate", NewAffineTransformWithM00M10M01M11M02M12(3, 0, 0, 7, 11, 13),
			[6]float64{0.3333333333333333, 0.0, 0.0, 0.14285714285714285, -3.6666666666666665, -1.8571428571428572}},
		{"shear and translate", NewAffineTransformWithM00M10M01M11M02M12(0, 3, 7, 0, 11, 13),
			[6]float64{0.0, 0.14285714285714285, 0.3333333333333333, 0.0, -4.333333333333333, -1.5714285714285714}},
		{"translate", NewAffineTransformWithM00M10M01M11M02M12(1, 0, 0, 1, 11, 13),
			[6]float64{1.0, 0.0, 0.0, 1.0, -11.0, -13.0}},
		{"identity", NewAffineTransform(), [6]float64{1, 0, 0, 1, 0, 0}},
	}
	for _, tc := range tests {
		inv, err := tc.at.CreateInverse()
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		assertMatrix(t, tc.name, inv, tc.want, 0)
	}
}

func TestAffineTransformCreateInverseOfSingularMatrix(t *testing.T) {
	tests := []struct {
		matrix  []float64
		message string
	}{
		{[]float64{1, 2, 2, 4, 0, 0}, "Determinant is 0.0"},
		{[]float64{0, 0, 0, 5, 1, 1}, "Determinant is 0"},
		{[]float64{0, 0, 3, 0, 1, 1}, "Determinant is 0"},
	}
	for _, tc := range tests {
		at := NewAffineTransformWithFlatmatrix(tc.matrix)
		inv, err := at.CreateInverse()
		var nte *NoninvertibleTransformException
		if inv != nil || !errors.As(err, &nte) || nte.Message != tc.message {
			t.Errorf("CreateInverse(%v) = %v, %v; want the error %q", tc.matrix, inv, err, tc.message)
		}
		if _, err := at.InverseTransform(NewPoint2DDouble(1, 1), nil); err == nil {
			t.Errorf("InverseTransform(%v) returned no error", tc.matrix)
		}
	}
}

// This is the sequence ITextOutputDevice uses to turn its current transform
// into a PDF text matrix.
func TestAffineTransformPDFNormalizeSequence(t *testing.T) {
	mx := make([]float64, 6)
	NewAffineTransform().GetMatrix(mx)
	mx[3] = -1
	mx[5] = 842.5
	nm := NewAffineTransformWithFlatmatrix(mx)
	cur := AffineTransformGetTranslateInstance(12.5, 30)
	cur.Scale(0.05, 0.05)
	nm.Concatenate(cur)
	nm.Concatenate(AffineTransformGetScaleInstance(1, -1))
	nm.Scale(20, 20)
	assertMatrix(t, "normalize", nm, [6]float64{1.0, 0.0, 0.0, 1.0, 12.5, 812.5}, 0)
}

func TestAffineTransformCloneEqualsAndSetters(t *testing.T) {
	at := NewAffineTransformWithM00M10M01M11M02M12(1, 2, 3, 4, 5, 6)
	c := at.Clone()
	if !c.Equals(at) || c == at {
		t.Errorf("Clone is wrong")
	}
	c.Translate(1, 1)
	if c.Equals(at) || at.GetTranslateX() != 5 {
		t.Errorf("Clone shares state with the original")
	}
	if !NewAffineTransformFromAffineTransform(at).Equals(at) || !NewAffineTransformFromValues(1, 2, 3, 4, 5, 6).Equals(at) {
		t.Errorf("the copying constructors are wrong")
	}
	c.SetTransform(at)
	if !c.Equals(at) {
		t.Errorf("SetTransform is wrong")
	}
	c.SetToIdentity()
	if !c.IsIdentity() {
		t.Errorf("SetToIdentity is wrong")
	}
	c.SetTransformWithM00M10M01M11M02M12(1, 2, 3, 4, 5, 6)
	if !c.Equals(at) || c.Equals(nil) {
		t.Errorf("SetTransformWithM00M10M01M11M02M12 or Equals is wrong")
	}
}

func TestAffineTransformArrays(t *testing.T) {
	at := NewAffineTransformWithM00M10M01M11M02M12(2, 0, 0, 3, 1, 1)
	src := []float64{0, 0, 1, 1, 2, 2}
	dst := make([]float64, 8)
	at.TransformDouble(src, 2, dst, 4, 2)
	if want := []float64{0, 0, 0, 0, 3, 4, 5, 7}; !equalFloats(dst, want) {
		t.Errorf("TransformDouble = %v, want %v", dst, want)
	}
	// Overlapping sections of one slice, destination after source.
	buf := []float64{1, 1, 2, 2, 0, 0}
	at.TransformDouble(buf, 0, buf, 2, 2)
	if want := []float64{1, 1, 3, 4, 5, 7}; !equalFloats(buf, want) {
		t.Errorf("overlapping TransformDouble = %v, want %v", buf, want)
	}
	fsrc := []float32{0.1, 0.2}
	at.TransformFloat(fsrc, 0, fsrc, 0, 1)
	if fsrc[0] != float32(2*float64(float32(0.1))+1) || fsrc[1] != float32(3*float64(float32(0.2))+1) {
		t.Errorf("TransformFloat = %v", fsrc)
	}
}

func equalFloats(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAffineTransformCreateTransformedShape(t *testing.T) {
	at := NewAffineTransformWithM00M10M01M11M02M12(2, 0, 0, 3, 5, 7)
	if at.CreateTransformedShape(nil) != nil {
		t.Errorf("CreateTransformedShape(nil) is not nil")
	}
	var noArea *Area
	if at.CreateTransformedShape(noArea) != nil {
		t.Errorf("CreateTransformedShape of a nil *Area is not a nil Shape")
	}
	s := at.CreateTransformedShape(NewRectangle(1, 2, 3, 4))
	if _, ok := s.(*Path2D); !ok {
		t.Fatalf("CreateTransformedShape returned %T, want *Path2D", s)
	}
	assertSegments(t, "transformed rectangle", collectDouble(s.GetPathIterator(nil)), []seg{
		{0, 7.0, 13.0},
		{1, 13.0, 13.0},
		{1, 13.0, 25.0},
		{1, 7.0, 25.0},
		{1, 7.0, 13.0},
		{4},
	}, 0)
	if !s.GetBounds2D().Equals(NewRectangle2DDouble(7, 13, 6, 12)) || !s.GetBounds().Equals(NewRectangle(7, 13, 6, 12)) {
		t.Errorf("bounds = %v, %v", s.GetBounds2D(), s.GetBounds())
	}
	// The source is not modified.
	src := NewPath2DDoubleFromShape(NewRectangle(1, 2, 3, 4))
	at.CreateTransformedShape(src)
	if !src.GetBounds().Equals(NewRectangle(1, 2, 3, 4)) {
		t.Errorf("CreateTransformedShape modified its argument")
	}
}
