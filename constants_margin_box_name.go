// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/constants/MarginBoxName.java

package ufo

var (
	marginBoxNameAll         = map[string]*MarginBoxName{}
	marginBoxNameMaxAssigned int
)

type MarginBoxName struct {
	FS_ID int

	ident         string
	textAlign     *IdentValue
	verticalAlign *IdentValue
}

// The variables are initialized in declaration order, so every FS_ID is the
// same as in Java.
var (
	MarginBoxNameTopLeftCorner     = marginBoxNameAddValue("top-left-corner", IdentValueRight, IdentValueMiddle)
	MarginBoxNameTopLeft           = marginBoxNameAddValue("top-left", IdentValueLeft, IdentValueMiddle)
	MarginBoxNameTopCenter         = marginBoxNameAddValue("top-center", IdentValueCenter, IdentValueMiddle)
	MarginBoxNameTopRight          = marginBoxNameAddValue("top-right", IdentValueRight, IdentValueMiddle)
	MarginBoxNameTopRightCorner    = marginBoxNameAddValue("top-right-corner", IdentValueLeft, IdentValueMiddle)
	MarginBoxNameBottomLeftCorner  = marginBoxNameAddValue("bottom-left-corner", IdentValueRight, IdentValueMiddle)
	MarginBoxNameBottomLeft        = marginBoxNameAddValue("bottom-left", IdentValueLeft, IdentValueMiddle)
	MarginBoxNameBottomCenter      = marginBoxNameAddValue("bottom-center", IdentValueCenter, IdentValueMiddle)
	MarginBoxNameBottomRight       = marginBoxNameAddValue("bottom-right", IdentValueRight, IdentValueMiddle)
	MarginBoxNameBottomRightCorner = marginBoxNameAddValue("bottom-right-corner", IdentValueLeft, IdentValueMiddle)
	MarginBoxNameLeftTop           = marginBoxNameAddValue("left-top", IdentValueCenter, IdentValueTop)
	MarginBoxNameLeftMiddle        = marginBoxNameAddValue("left-middle", IdentValueCenter, IdentValueMiddle)
	MarginBoxNameLeftBottom        = marginBoxNameAddValue("left-bottom", IdentValueCenter, IdentValueBottom)
	MarginBoxNameRightTop          = marginBoxNameAddValue("right-top", IdentValueCenter, IdentValueTop)
	MarginBoxNameRightMiddle       = marginBoxNameAddValue("right-middle", IdentValueCenter, IdentValueMiddle)
	MarginBoxNameRightBottom       = marginBoxNameAddValue("right-bottom", IdentValueCenter, IdentValueBottom)

	// HACK to support page level XMP metadata.  For ease of implementation, it reuses
	// the margin box infrastructure, but is instead embedded in the PDF vs. being displayed
	// on the screen.
	MarginBoxNameFsPdfXmpMetadata = marginBoxNameAddValue("-fs-pdf-xmp-metadata", IdentValueTop, IdentValueLeft)
)

func newMarginBoxName(ident string, textAlign *IdentValue, verticalAlign *IdentValue) *MarginBoxName {
	m := &MarginBoxName{
		ident:         ident,
		textAlign:     textAlign,
		verticalAlign: verticalAlign,
	}

	m.FS_ID = marginBoxNameMaxAssigned
	marginBoxNameMaxAssigned++
	return m
}

func marginBoxNameAddValue(ident string, textAlign *IdentValue, verticalAlign *IdentValue) *MarginBoxName {
	val := newMarginBoxName(ident, textAlign, verticalAlign)
	marginBoxNameAll[ident] = val
	return val
}

func (m *MarginBoxName) ToString() string {
	return m.ident
}

func (m *MarginBoxName) String() string {
	return m.ident
}

// MarginBoxNameValueOf may return nil.
func MarginBoxNameValueOf(ident string) *MarginBoxName {
	return marginBoxNameAll[ident]
}

func (m *MarginBoxName) HashCode() int {
	return m.FS_ID
}

func (m *MarginBoxName) Equals(o any) bool {
	marginBoxName, ok := o.(*MarginBoxName)
	if !ok || marginBoxName == nil {
		return false
	}

	return m.FS_ID == marginBoxName.FS_ID
}

func (m *MarginBoxName) GetInitialTextAlign() *IdentValue {
	return m.textAlign
}

func (m *MarginBoxName) GetInitialVerticalAlign() *IdentValue {
	return m.verticalAlign
}
