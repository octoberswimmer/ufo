package geom

import "math"

// AffineTransform is java.awt.geom.AffineTransform: the matrix
//
//	[ m00 m01 m02 ]
//	[ m10 m11 m12 ]
//	[  0   0   1  ]
//
// applied to column vectors. The JDK tracks which parts of the matrix are
// non-trivial and skips the trivial ones; this type derives the same state
// from the matrix values where the JDK result depends on it (CreateInverse,
// InverseTransform).
type AffineTransform struct {
	m00, m10, m01, m11, m02, m12 float64
}

const (
	applyIdentity  = 0
	applyTranslate = 1
	applyScale     = 2
	applyShear     = 4
)

// NewAffineTransform is new AffineTransform(): the identity.
func NewAffineTransform() *AffineTransform {
	return &AffineTransform{m00: 1, m11: 1}
}

// NewAffineTransformWithM00M10M01M11M02M12 is new AffineTransform(m00, m10,
// m01, m11, m02, m12). Note the column order of the arguments.
func NewAffineTransformWithM00M10M01M11M02M12(m00, m10, m01, m11, m02, m12 float64) *AffineTransform {
	return &AffineTransform{m00: m00, m10: m10, m01: m01, m11: m11, m02: m02, m12: m12}
}

// NewAffineTransformFromValues is the same as
// NewAffineTransformWithM00M10M01M11M02M12.
func NewAffineTransformFromValues(m00, m10, m01, m11, m02, m12 float64) *AffineTransform {
	return NewAffineTransformWithM00M10M01M11M02M12(m00, m10, m01, m11, m02, m12)
}

// NewAffineTransformWithFlatmatrix is new AffineTransform(double[]
// flatmatrix): { m00 m10 m01 m11 [m02 m12] }. The translation is zero when the
// slice has fewer than six values.
func NewAffineTransformWithFlatmatrix(flatmatrix []float64) *AffineTransform {
	t := &AffineTransform{
		m00: flatmatrix[0],
		m10: flatmatrix[1],
		m01: flatmatrix[2],
		m11: flatmatrix[3],
	}
	if len(flatmatrix) > 5 {
		t.m02 = flatmatrix[4]
		t.m12 = flatmatrix[5]
	}
	return t
}

// NewAffineTransformFromAffineTransform is new AffineTransform(AffineTransform).
func NewAffineTransformFromAffineTransform(tx *AffineTransform) *AffineTransform {
	c := *tx
	return &c
}

// AffineTransformGetTranslateInstance is AffineTransform.getTranslateInstance.
func AffineTransformGetTranslateInstance(tx, ty float64) *AffineTransform {
	t := NewAffineTransform()
	t.SetToTranslation(tx, ty)
	return t
}

// AffineTransformGetScaleInstance is AffineTransform.getScaleInstance.
func AffineTransformGetScaleInstance(sx, sy float64) *AffineTransform {
	t := NewAffineTransform()
	t.SetToScale(sx, sy)
	return t
}

// AffineTransformGetShearInstance is AffineTransform.getShearInstance.
func AffineTransformGetShearInstance(shx, shy float64) *AffineTransform {
	t := NewAffineTransform()
	t.SetToShear(shx, shy)
	return t
}

// AffineTransformGetRotateInstance is AffineTransform.getRotateInstance(theta).
func AffineTransformGetRotateInstance(theta float64) *AffineTransform {
	t := NewAffineTransform()
	t.SetToRotation(theta)
	return t
}

// AffineTransformGetRotateInstanceWithAnchorxAnchory is
// AffineTransform.getRotateInstance(theta, anchorx, anchory).
func AffineTransformGetRotateInstanceWithAnchorxAnchory(theta, anchorx, anchory float64) *AffineTransform {
	t := NewAffineTransform()
	t.SetToRotationWithAnchorxAnchory(theta, anchorx, anchory)
	return t
}

// state derives the JDK's state flags from the matrix values, as the JDK's
// updateState does.
func (t *AffineTransform) state() int {
	s := applyIdentity
	if t.m01 == 0.0 && t.m10 == 0.0 {
		if t.m00 != 1.0 || t.m11 != 1.0 {
			s = applyScale
		}
	} else {
		s = applyShear
		if t.m00 != 0.0 || t.m11 != 0.0 {
			s |= applyScale
		}
	}
	if t.m02 != 0.0 || t.m12 != 0.0 {
		s |= applyTranslate
	}
	return s
}

// GetMatrix writes { m00 m10 m01 m11 m02 m12 } into flatmatrix. A slice with
// fewer than six elements receives the first four values only.
func (t *AffineTransform) GetMatrix(flatmatrix []float64) {
	flatmatrix[0] = t.m00
	flatmatrix[1] = t.m10
	flatmatrix[2] = t.m01
	flatmatrix[3] = t.m11
	if len(flatmatrix) > 5 {
		flatmatrix[4] = t.m02
		flatmatrix[5] = t.m12
	}
}

func (t *AffineTransform) GetScaleX() float64 {
	return t.m00
}

func (t *AffineTransform) GetScaleY() float64 {
	return t.m11
}

func (t *AffineTransform) GetShearX() float64 {
	return t.m01
}

func (t *AffineTransform) GetShearY() float64 {
	return t.m10
}

func (t *AffineTransform) GetTranslateX() float64 {
	return t.m02
}

func (t *AffineTransform) GetTranslateY() float64 {
	return t.m12
}

func (t *AffineTransform) GetDeterminant() float64 {
	switch t.state() &^ applyTranslate {
	case applyShear | applyScale:
		return float64(t.m00*t.m11) - float64(t.m01*t.m10)
	case applyShear:
		return -(t.m01 * t.m10)
	case applyScale:
		return t.m00 * t.m11
	}
	return 1.0
}

func (t *AffineTransform) IsIdentity() bool {
	return t.state() == applyIdentity
}

// Translate concatenates a translation: [this] = [this] x [translate].
func (t *AffineTransform) Translate(tx, ty float64) {
	t.m02 = float64(tx*t.m00) + float64(ty*t.m01) + t.m02
	t.m12 = float64(tx*t.m10) + float64(ty*t.m11) + t.m12
}

// Scale concatenates a scale: [this] = [this] x [scale].
func (t *AffineTransform) Scale(sx, sy float64) {
	t.m00 *= sx
	t.m10 *= sx
	t.m01 *= sy
	t.m11 *= sy
}

// Shear concatenates a shear: [this] = [this] x [shear].
func (t *AffineTransform) Shear(shx, shy float64) {
	M0 := t.m00
	M1 := t.m01
	t.m00 = M0 + float64(M1*shy)
	t.m01 = float64(M0*shx) + M1
	M0 = t.m10
	M1 = t.m11
	t.m10 = M0 + float64(M1*shy)
	t.m11 = float64(M0*shx) + M1
}

// Rotate concatenates a rotation by theta radians. A rotation by a multiple
// of 90 degrees, detected by sin(theta) or cos(theta) being exactly 1 or -1,
// moves and negates matrix entries without multiplying, as the JDK does, so
// it introduces no rounding error.
func (t *AffineTransform) Rotate(theta float64) {
	sin := math.Sin(theta)
	if sin == 1.0 || sin == -1.0 {
		if sin == 1.0 {
			t.rotate90()
		} else {
			t.rotate270()
		}
		return
	}
	cos := math.Cos(theta)
	if cos == -1.0 {
		t.rotate180()
	} else if cos != 1.0 {
		M0 := t.m00
		M1 := t.m01
		t.m00 = float64(cos*M0) + float64(sin*M1)
		t.m01 = float64(-sin*M0) + float64(cos*M1)
		M0 = t.m10
		M1 = t.m11
		t.m10 = float64(cos*M0) + float64(sin*M1)
		t.m11 = float64(-sin*M0) + float64(cos*M1)
	}
}

func (t *AffineTransform) rotate90() {
	M0 := t.m00
	t.m00 = t.m01
	t.m01 = -M0
	M0 = t.m10
	t.m10 = t.m11
	t.m11 = -M0
}

func (t *AffineTransform) rotate180() {
	t.m00 = -t.m00
	t.m11 = -t.m11
	if t.m01 != 0.0 || t.m10 != 0.0 {
		// If there was a shear, then this rotation has no
		// effect on the state.
		t.m01 = -t.m01
		t.m10 = -t.m10
	}
}

func (t *AffineTransform) rotate270() {
	M0 := t.m00
	t.m00 = -t.m01
	t.m01 = M0
	M0 = t.m10
	t.m10 = -t.m11
	t.m11 = M0
}

// RotateWithAnchorxAnchory is rotate(theta, anchorx, anchory): a rotation
// around the anchor point.
func (t *AffineTransform) RotateWithAnchorxAnchory(theta, anchorx, anchory float64) {
	t.Translate(anchorx, anchory)
	t.Rotate(theta)
	t.Translate(-anchorx, -anchory)
}

func (t *AffineTransform) SetToIdentity() {
	*t = AffineTransform{m00: 1, m11: 1}
}

func (t *AffineTransform) SetToTranslation(tx, ty float64) {
	*t = AffineTransform{m00: 1, m11: 1, m02: tx, m12: ty}
}

func (t *AffineTransform) SetToScale(sx, sy float64) {
	*t = AffineTransform{m00: sx, m11: sy}
}

func (t *AffineTransform) SetToShear(shx, shy float64) {
	*t = AffineTransform{m00: 1, m01: shx, m10: shy, m11: 1}
}

// SetToRotation sets the matrix to a rotation by theta radians. For a
// multiple of 90 degrees the entries are exactly 0, 1 and -1.
func (t *AffineTransform) SetToRotation(theta float64) {
	sin := math.Sin(theta)
	var cos float64
	if sin == 1.0 || sin == -1.0 {
		cos = 0.0
	} else {
		cos = math.Cos(theta)
		if cos == -1.0 || cos == 1.0 {
			sin = 0.0
		}
	}
	*t = AffineTransform{m00: cos, m10: sin, m01: -sin, m11: cos}
}

// SetToRotationWithAnchorxAnchory is setToRotation(theta, anchorx, anchory).
func (t *AffineTransform) SetToRotationWithAnchorxAnchory(theta, anchorx, anchory float64) {
	t.SetToRotation(theta)
	sin := t.m10
	oneMinusCos := 1.0 - t.m00
	t.m02 = float64(anchorx*oneMinusCos) + float64(anchory*sin)
	t.m12 = float64(anchory*oneMinusCos) - float64(anchorx*sin)
}

// SetTransform copies the matrix of tx.
func (t *AffineTransform) SetTransform(tx *AffineTransform) {
	*t = *tx
}

// SetTransformWithM00M10M01M11M02M12 is setTransform(m00, m10, m01, m11, m02,
// m12).
func (t *AffineTransform) SetTransformWithM00M10M01M11M02M12(m00, m10, m01, m11, m02, m12 float64) {
	*t = AffineTransform{m00: m00, m10: m10, m01: m01, m11: m11, m02: m02, m12: m12}
}

// Concatenate sets [this] = [this] x [tx]: tx is applied to a point first,
// then the original transform.
func (t *AffineTransform) Concatenate(tx *AffineTransform) {
	T00, T01, T02 := tx.m00, tx.m01, tx.m02
	T10, T11, T12 := tx.m10, tx.m11, tx.m12
	M0 := t.m00
	M1 := t.m01
	t.m00 = float64(T00*M0) + float64(T10*M1)
	t.m01 = float64(T01*M0) + float64(T11*M1)
	t.m02 += float64(T02*M0) + float64(T12*M1)
	M0 = t.m10
	M1 = t.m11
	t.m10 = float64(T00*M0) + float64(T10*M1)
	t.m11 = float64(T01*M0) + float64(T11*M1)
	t.m12 += float64(T02*M0) + float64(T12*M1)
}

// PreConcatenate sets [this] = [tx] x [this]: the original transform is
// applied to a point first, then tx.
func (t *AffineTransform) PreConcatenate(tx *AffineTransform) {
	T00, T01, T02 := tx.m00, tx.m01, tx.m02
	T10, T11, T12 := tx.m10, tx.m11, tx.m12
	M0 := t.m00
	M1 := t.m10
	t.m00 = float64(T00*M0) + float64(T01*M1)
	t.m10 = float64(T10*M0) + float64(T11*M1)
	M0 = t.m01
	M1 = t.m11
	t.m01 = float64(T00*M0) + float64(T01*M1)
	t.m11 = float64(T10*M0) + float64(T11*M1)
	M0 = t.m02
	M1 = t.m12
	t.m02 = float64(T00*M0) + float64(T01*M1) + T02
	t.m12 = float64(T10*M0) + float64(T11*M1) + T12
}

// CreateInverse returns the inverse transform, or a
// *NoninvertibleTransformException when the matrix has no inverse.
func (t *AffineTransform) CreateInverse() (*AffineTransform, error) {
	switch t.state() {
	case applyShear | applyScale | applyTranslate, applyShear | applyScale:
		det := float64(t.m00*t.m11) - float64(t.m01*t.m10)
		if math.Abs(det) <= math.SmallestNonzeroFloat64 {
			return nil, &NoninvertibleTransformException{Message: "Determinant is " + javaDoubleString(det)}
		}
		inv := &AffineTransform{
			m00: t.m11 / det,
			m10: -t.m10 / det,
			m01: -t.m01 / det,
			m11: t.m00 / det,
		}
		if t.state()&applyTranslate != 0 {
			inv.m02 = (float64(t.m01*t.m12) - float64(t.m11*t.m02)) / det
			inv.m12 = (float64(t.m10*t.m02) - float64(t.m00*t.m12)) / det
		}
		return inv, nil
	case applyShear | applyTranslate, applyShear:
		if t.m01 == 0.0 || t.m10 == 0.0 {
			return nil, &NoninvertibleTransformException{Message: "Determinant is 0"}
		}
		inv := &AffineTransform{m10: 1.0 / t.m01, m01: 1.0 / t.m10}
		if t.state()&applyTranslate != 0 {
			inv.m02 = -t.m12 / t.m10
			inv.m12 = -t.m02 / t.m01
		}
		return inv, nil
	case applyScale | applyTranslate, applyScale:
		if t.m00 == 0.0 || t.m11 == 0.0 {
			return nil, &NoninvertibleTransformException{Message: "Determinant is 0"}
		}
		inv := &AffineTransform{m00: 1.0 / t.m00, m11: 1.0 / t.m11}
		if t.state()&applyTranslate != 0 {
			inv.m02 = -t.m02 / t.m00
			inv.m12 = -t.m12 / t.m11
		}
		return inv, nil
	case applyTranslate:
		return &AffineTransform{m00: 1, m11: 1, m02: -t.m02, m12: -t.m12}, nil
	}
	return NewAffineTransform(), nil
}

// Transform applies the transform to ptSrc and stores the result in ptDst,
// which it returns. A nil ptDst is allocated. ptSrc and ptDst may be the same
// point.
func (t *AffineTransform) Transform(ptSrc, ptDst *Point2D) *Point2D {
	if ptDst == nil {
		ptDst = &Point2D{}
	}
	// Copy source coords into local variables in case src == dst
	x := ptSrc.X
	y := ptSrc.Y
	ptDst.X = float64(x*t.m00) + float64(y*t.m01) + t.m02
	ptDst.Y = float64(x*t.m10) + float64(y*t.m11) + t.m12
	return ptDst
}

// TransformDouble is transform(double[] srcPts, int srcOff, double[] dstPts,
// int dstOff, int numPts). The source and destination may be the same slice;
// when the destination section starts inside the source section, the source
// is copied first so that no point is overwritten before it is read.
func (t *AffineTransform) TransformDouble(srcPts []float64, srcOff int, dstPts []float64, dstOff int, numPts int) {
	if numPts > 0 && len(srcPts) > 0 && len(dstPts) > 0 && &srcPts[0] == &dstPts[0] &&
		dstOff > srcOff && dstOff < srcOff+numPts*2 {
		copy(dstPts[dstOff:dstOff+numPts*2], srcPts[srcOff:srcOff+numPts*2])
		srcOff = dstOff
	}
	for ; numPts > 0; numPts-- {
		x := srcPts[srcOff]
		y := srcPts[srcOff+1]
		srcOff += 2
		dstPts[dstOff] = float64(t.m00*x) + float64(t.m01*y) + t.m02
		dstPts[dstOff+1] = float64(t.m10*x) + float64(t.m11*y) + t.m12
		dstOff += 2
	}
}

// TransformFloat is transform(float[] srcPts, int srcOff, float[] dstPts, int
// dstOff, int numPts). The arithmetic is done in float64 and each result is
// rounded to float32.
func (t *AffineTransform) TransformFloat(srcPts []float32, srcOff int, dstPts []float32, dstOff int, numPts int) {
	if numPts > 0 && len(srcPts) > 0 && len(dstPts) > 0 && &srcPts[0] == &dstPts[0] &&
		dstOff > srcOff && dstOff < srcOff+numPts*2 {
		copy(dstPts[dstOff:dstOff+numPts*2], srcPts[srcOff:srcOff+numPts*2])
		srcOff = dstOff
	}
	for ; numPts > 0; numPts-- {
		x := float64(srcPts[srcOff])
		y := float64(srcPts[srcOff+1])
		srcOff += 2
		dstPts[dstOff] = float32(float64(t.m00*x) + float64(t.m01*y) + t.m02)
		dstPts[dstOff+1] = float32(float64(t.m10*x) + float64(t.m11*y) + t.m12)
		dstOff += 2
	}
}

// DeltaTransform applies the transform without its translation, which is the
// transform of a distance vector.
func (t *AffineTransform) DeltaTransform(ptSrc, ptDst *Point2D) *Point2D {
	if ptDst == nil {
		ptDst = &Point2D{}
	}
	x := ptSrc.X
	y := ptSrc.Y
	ptDst.X = float64(x*t.m00) + float64(y*t.m01)
	ptDst.Y = float64(x*t.m10) + float64(y*t.m11)
	return ptDst
}

// InverseTransform applies the inverse of the transform to ptSrc. It returns a
// *NoninvertibleTransformException when the matrix has no inverse.
func (t *AffineTransform) InverseTransform(ptSrc, ptDst *Point2D) (*Point2D, error) {
	inv, err := t.CreateInverse()
	if err != nil {
		return nil, err
	}
	if ptDst == nil {
		ptDst = &Point2D{}
	}
	x := ptSrc.X
	y := ptSrc.Y
	switch t.state() {
	case applyShear | applyScale | applyTranslate, applyShear | applyScale:
		x -= t.m02
		y -= t.m12
		det := float64(t.m00*t.m11) - float64(t.m01*t.m10)
		ptDst.X = (float64(x*t.m11) - float64(y*t.m01)) / det
		ptDst.Y = (float64(y*t.m00) - float64(x*t.m10)) / det
	case applyShear | applyTranslate, applyShear:
		x -= t.m02
		y -= t.m12
		ptDst.X = y / t.m10
		ptDst.Y = x / t.m01
	case applyScale | applyTranslate, applyScale:
		x -= t.m02
		y -= t.m12
		ptDst.X = x / t.m00
		ptDst.Y = y / t.m11
	default:
		inv.Transform(&Point2D{X: x, Y: y}, ptDst)
	}
	return ptDst, nil
}

// CreateTransformedShape returns the shape transformed by this transform, or
// nil for a nil shape. As in the JDK, the result is a new Path2D (the Double
// variant) built from the shape's path, even for a Rectangle. The exception is
// an *Area, which has no single path: the result is an *Area whose member
// shapes are the transformed members.
func (t *AffineTransform) CreateTransformedShape(pSrc Shape) Shape {
	if pSrc == nil {
		return nil
	}
	if a, ok := pSrc.(*Area); ok {
		if a == nil {
			return nil
		}
		return a.CreateTransformedArea(t)
	}
	return NewPath2DDoubleFromShapeWithAt(pSrc, t)
}

func (t *AffineTransform) Clone() *AffineTransform {
	c := *t
	return &c
}

func (t *AffineTransform) Equals(o *AffineTransform) bool {
	return o != nil && t.m00 == o.m00 && t.m01 == o.m01 && t.m02 == o.m02 &&
		t.m10 == o.m10 && t.m11 == o.m11 && t.m12 == o.m12
}

// matround rounds to 15 decimal places, as the JDK does for toString.
func matround(matval float64) float64 {
	return math.RoundToEven(matval*1e15) / 1e15
}

func (t *AffineTransform) String() string {
	return "AffineTransform[[" +
		javaDoubleString(matround(t.m00)) + ", " +
		javaDoubleString(matround(t.m01)) + ", " +
		javaDoubleString(matround(t.m02)) + "], [" +
		javaDoubleString(matround(t.m10)) + ", " +
		javaDoubleString(matround(t.m11)) + ", " +
		javaDoubleString(matround(t.m12)) + "]]"
}
