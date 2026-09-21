// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/BorderPropertySet.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

var borderPropertySetNoCorners = &borderPropertySetCorners{BorderRadiusCornerUndefined, BorderRadiusCornerUndefined, BorderRadiusCornerUndefined, BorderRadiusCornerUndefined}
var borderPropertySetNoStyles = &borderPropertySetStyles{nil, nil, nil, nil}
var borderPropertySetNoColors = &borderPropertySetColors{FSRGBColorTransparent, FSRGBColorTransparent, FSRGBColorTransparent, FSRGBColorTransparent}
var BorderPropertySetEmptyBorder = newBorderPropertySetWithSides(0.0, 0.0, 0.0, 0.0, borderPropertySetNoStyles, borderPropertySetNoCorners, borderPropertySetNoColors)

// borderPropertySetStyles ports the private record BorderPropertySet.Styles;
// each side may be nil.
type borderPropertySetStyles struct {
	top    *IdentValue
	right  *IdentValue
	bottom *IdentValue
	left   *IdentValue
}

func (s *borderPropertySetStyles) hasHidden() bool {
	return s.top == IdentValueHidden || s.right == IdentValueHidden || s.bottom == IdentValueHidden || s.left == IdentValueHidden
}

// borderPropertySetColors ports the private record BorderPropertySet.Colors.
type borderPropertySetColors struct {
	top    FSColor
	right  FSColor
	bottom FSColor
	left   FSColor
}

func (c *borderPropertySetColors) lighten() *borderPropertySetColors {
	return &borderPropertySetColors{
		c.top.LightenColor(),
		c.right.LightenColor(),
		c.bottom.LightenColor(),
		c.left.LightenColor(),
	}
}

func (c *borderPropertySetColors) darken() *borderPropertySetColors {
	return &borderPropertySetColors{
		c.top.DarkenColor(),
		c.right.DarkenColor(),
		c.bottom.DarkenColor(),
		c.left.DarkenColor(),
	}
}

// borderPropertySetCorners ports the private record BorderPropertySet.Corners.
type borderPropertySetCorners struct {
	topLeft     *BorderRadiusCorner
	topRight    *BorderRadiusCorner
	bottomRight *BorderRadiusCorner
	bottomLeft  *BorderRadiusCorner
}

// User: patrick
// Date: Oct 21, 2005
type BorderPropertySet struct {
	RectPropertySet

	styles  *borderPropertySetStyles
	colors  *borderPropertySetColors
	corners *borderPropertySetCorners
}

func NewBorderPropertySet(border *BorderPropertySet) *BorderPropertySet {
	return newBorderPropertySetWithSides(border.Top(), border.Right(), border.Bottom(), border.Left(), border.styles, border.corners, border.colors)
}

func newBorderPropertySetWithSides(
	top float32,
	right float32,
	bottom float32,
	left float32,
	styles *borderPropertySetStyles,
	corners *borderPropertySetCorners,
	colors *borderPropertySetColors,
) *BorderPropertySet {
	return &BorderPropertySet{
		RectPropertySet: RectPropertySet{top: top, right: right, bottom: bottom, left: left},
		styles:          styles,
		colors:          colors,
		corners:         corners,
	}
}

func NewBorderPropertySetWithTopRightBottomLeft(
	top *CollapsedBorderValue,
	right *CollapsedBorderValue,
	bottom *CollapsedBorderValue,
	left *CollapsedBorderValue,
) *BorderPropertySet {
	return newBorderPropertySetWithSides(float32(top.Width()),
		float32(right.Width()),
		float32(bottom.Width()),
		float32(left.Width()),
		&borderPropertySetStyles{top.Style(), right.Style(), bottom.Style(), left.Style()},
		borderPropertySetNoCorners,
		&borderPropertySetColors{top.Color(), right.Color(), bottom.Color(), left.Color()},
	)
}

func newBorderPropertySetWithStyleCtx(
	style CalculatedStyleI,
	ctx CssContext,
) *BorderPropertySet {
	top := borderPropertySetCalculate(style, CSSNameBorderTopStyle, CSSNameBorderTopWidth, ctx)
	right := borderPropertySetCalculate(style, CSSNameBorderRightStyle, CSSNameBorderRightWidth, ctx)
	bottom := borderPropertySetCalculate(style, CSSNameBorderBottomStyle, CSSNameBorderBottomWidth, ctx)
	left := borderPropertySetCalculate(style, CSSNameBorderLeftStyle, CSSNameBorderLeftWidth, ctx)
	styles := &borderPropertySetStyles{
		style.GetIdent(CSSNameBorderTopStyle),
		style.GetIdent(CSSNameBorderRightStyle),
		style.GetIdent(CSSNameBorderBottomStyle),
		style.GetIdent(CSSNameBorderLeftStyle),
	}
	corners := &borderPropertySetCorners{
		NewBorderRadiusCornerWithFromValStyleCtx(CSSNameBorderTopLeftRadius, style, ctx),
		NewBorderRadiusCornerWithFromValStyleCtx(CSSNameBorderTopRightRadius, style, ctx),
		NewBorderRadiusCornerWithFromValStyleCtx(CSSNameBorderBottomRightRadius, style, ctx),
		NewBorderRadiusCornerWithFromValStyleCtx(CSSNameBorderBottomLeftRadius, style, ctx),
	}
	colors := &borderPropertySetColors{
		style.AsColor(CSSNameBorderTopColor),
		style.AsColor(CSSNameBorderRightColor),
		style.AsColor(CSSNameBorderBottomColor),
		style.AsColor(CSSNameBorderLeftColor),
	}
	return newBorderPropertySetWithSides(top, right, bottom, left, styles, corners, colors)
}

func borderPropertySetCalculate(style CalculatedStyleI, borderStyle *CSSName, borderWidth *CSSName, ctx CssContext) float32 {
	if style.IsIdent(borderStyle, IdentValueNone) || style.IsIdent(borderStyle, IdentValueHidden) {
		return 0
	}
	return style.GetFloatPropertyProportionalHeight(borderWidth, 0, ctx)
}

// Deprecated: use Lighten.
func (b *BorderPropertySet) LightenWithStyle(style *IdentValue) *BorderPropertySet {
	return b.Lighten()
}

// Lighten returns the colors for brighter parts of each side for a particular decoration style
func (b *BorderPropertySet) Lighten() *BorderPropertySet {
	return newBorderPropertySetWithSides(
		b.Top(), b.Right(), b.Bottom(), b.Left(),
		b.styles,
		b.corners,
		b.colors.lighten(),
	)
}

// Deprecated: use Darken.
func (b *BorderPropertySet) DarkenWithStyle(style *IdentValue) *BorderPropertySet {
	return b.Darken()
}

// Darken returns the colors for brighter parts of each side for a particular decoration style
func (b *BorderPropertySet) Darken() *BorderPropertySet {
	return newBorderPropertySetWithSides(
		b.Top(), b.Right(), b.Bottom(), b.Left(),
		b.styles,
		b.corners,
		b.colors.darken(),
	)
}

// BorderPropertySetNewInstance accepts a nil ctx.
func BorderPropertySetNewInstance(
	style CalculatedStyleI,
	ctx CssContext,
) *BorderPropertySet {
	result := newBorderPropertySetWithStyleCtx(style, ctx)
	if result.IsAllZeros() && !result.HasHidden() && !result.HasBorderRadius() {
		return BorderPropertySetEmptyBorder
	}
	return result

}

func (b *BorderPropertySet) ToString() string {
	return "BorderPropertySet[top=" + lengthValueFloatToString(b.Top()) +
		",right=" + lengthValueFloatToString(b.Right()) +
		",bottom=" + lengthValueFloatToString(b.Bottom()) +
		",left=" + lengthValueFloatToString(b.Left()) + "]"
}

func (b *BorderPropertySet) String() string {
	return b.ToString()
}

func (b *BorderPropertySet) NoTop() bool {
	return b.styles.top == IdentValueNone || calculatedStyleFloatToInt(b.Top()) == 0
}

func (b *BorderPropertySet) NoRight() bool {
	return b.styles.right == IdentValueNone || calculatedStyleFloatToInt(b.Right()) == 0
}

func (b *BorderPropertySet) NoBottom() bool {
	return b.styles.bottom == IdentValueNone || calculatedStyleFloatToInt(b.Bottom()) == 0
}

func (b *BorderPropertySet) NoLeft() bool {
	return b.styles.left == IdentValueNone || calculatedStyleFloatToInt(b.Left()) == 0
}

// TopStyle may return nil.
func (b *BorderPropertySet) TopStyle() *IdentValue {
	return b.styles.top
}

// RightStyle may return nil.
func (b *BorderPropertySet) RightStyle() *IdentValue {
	return b.styles.right
}

// BottomStyle may return nil.
func (b *BorderPropertySet) BottomStyle() *IdentValue {
	return b.styles.bottom
}

// LeftStyle may return nil.
func (b *BorderPropertySet) LeftStyle() *IdentValue {
	return b.styles.left
}

func (b *BorderPropertySet) TopColor() FSColor {
	return b.colors.top
}

func (b *BorderPropertySet) RightColor() FSColor {
	return b.colors.right
}

func (b *BorderPropertySet) BottomColor() FSColor {
	return b.colors.bottom
}

func (b *BorderPropertySet) LeftColor() FSColor {
	return b.colors.left
}

func (b *BorderPropertySet) HasHidden() bool {
	return b.styles.hasHidden()
}

func (b *BorderPropertySet) HasBorderRadius() bool {
	return b.GetTopLeft().HasRadius() || b.GetTopRight().HasRadius() || b.GetBottomLeft().HasRadius() || b.GetBottomRight().HasRadius()
}

func (b *BorderPropertySet) GetBottomRight() *BorderRadiusCorner {
	return b.corners.bottomRight
}

func (b *BorderPropertySet) GetBottomLeft() *BorderRadiusCorner {
	return b.corners.bottomLeft
}

func (b *BorderPropertySet) GetTopRight() *BorderRadiusCorner {
	return b.corners.topRight
}

func (b *BorderPropertySet) GetTopLeft() *BorderRadiusCorner {
	return b.corners.topLeft
}

func (b *BorderPropertySet) NormalizedInstance(bounds *geom.Rectangle) *BorderPropertySet {
	var factor float32 = 1
	boundsWidth := float32(bounds.Width)
	boundsHeight := float32(bounds.Height)

	// top
	factor = min(factor, boundsWidth/b.getSideLength(b.corners.topLeft, b.corners.topRight, boundsWidth))
	// bottom
	factor = min(factor, boundsWidth/b.getSideLength(b.corners.bottomRight, b.corners.bottomLeft, boundsWidth))
	// right
	factor = min(factor, boundsHeight/b.getSideLength(b.corners.topRight, b.corners.bottomRight, boundsHeight))
	// left
	factor = min(factor, boundsHeight/b.getSideLength(b.corners.bottomLeft, b.corners.bottomRight, boundsHeight))

	normalizedCorners := &borderPropertySetCorners{
		NewBorderRadiusCorner(factor*b.corners.topLeft.GetMaxLeft(boundsHeight), factor*b.corners.topLeft.GetMaxRight(boundsWidth)),
		NewBorderRadiusCorner(factor*b.corners.topRight.GetMaxLeft(boundsWidth), factor*b.corners.topRight.GetMaxRight(boundsHeight)),
		NewBorderRadiusCorner(factor*b.corners.bottomRight.GetMaxLeft(boundsHeight), factor*b.corners.bottomRight.GetMaxRight(boundsWidth)),
		NewBorderRadiusCorner(factor*b.corners.bottomLeft.GetMaxLeft(boundsWidth), factor*b.corners.bottomLeft.GetMaxRight(boundsHeight)),
	}
	return newBorderPropertySetWithSides(b.Top(), b.Right(), b.Bottom(), b.Left(),
		b.styles,
		normalizedCorners,
		b.colors,
	)
}

// getSideLength is a helper function for normalizeBorderRadius. Gets the max side width for each of the corners or the side width whichever is larger
func (b *BorderPropertySet) getSideLength(left *BorderRadiusCorner, right *BorderRadiusCorner, sideLength float32) float32 {
	return max(sideLength, left.GetMaxRight(sideLength)+right.GetMaxLeft(sideLength))
}

// ResetNegativeValues overrides RectPropertySet.ResetNegativeValues. Java
// narrows the return type to BorderPropertySet; a Go method that satisfies
// RectPropertySetI cannot, so the returned value always holds a
// *BorderPropertySet.
func (b *BorderPropertySet) ResetNegativeValues() RectPropertySetI {
	return newBorderPropertySetWithSides(rectPropertySetMax(0, b.Top()), rectPropertySetMax(0, b.Right()), rectPropertySetMax(0, b.Bottom()), rectPropertySetMax(0, b.Left()), b.styles, b.corners, b.colors)
}
