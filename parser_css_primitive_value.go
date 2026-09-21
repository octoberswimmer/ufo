// Ported from the constants of the W3C DOM Level 2 Style interfaces
// org.w3c.dom.css.CSSPrimitiveValue and org.w3c.dom.css.CSSValue that
// flying-saucer-core uses. The numeric values are the W3C ones.

package ufo

// Unit types of org.w3c.dom.css.CSSPrimitiveValue.
const (
	CSSPrimitiveValueCssUnknown    int16 = 0
	CSSPrimitiveValueCssNumber     int16 = 1
	CSSPrimitiveValueCssPercentage int16 = 2
	CSSPrimitiveValueCssEms        int16 = 3
	CSSPrimitiveValueCssExs        int16 = 4
	CSSPrimitiveValueCssPx         int16 = 5
	CSSPrimitiveValueCssCm         int16 = 6
	CSSPrimitiveValueCssMm         int16 = 7
	CSSPrimitiveValueCssIn         int16 = 8
	CSSPrimitiveValueCssPt         int16 = 9
	CSSPrimitiveValueCssPc         int16 = 10
	CSSPrimitiveValueCssDeg        int16 = 11
	CSSPrimitiveValueCssRad        int16 = 12
	CSSPrimitiveValueCssGrad       int16 = 13
	CSSPrimitiveValueCssMs         int16 = 14
	CSSPrimitiveValueCssS          int16 = 15
	CSSPrimitiveValueCssHz         int16 = 16
	CSSPrimitiveValueCssKhz        int16 = 17
	CSSPrimitiveValueCssDimension  int16 = 18
	CSSPrimitiveValueCssString     int16 = 19
	CSSPrimitiveValueCssUri        int16 = 20
	CSSPrimitiveValueCssIdent      int16 = 21
	CSSPrimitiveValueCssAttr       int16 = 22
	CSSPrimitiveValueCssCounter    int16 = 23
	CSSPrimitiveValueCssRect       int16 = 24
	CSSPrimitiveValueCssRgbcolor   int16 = 25
)

// Value types of org.w3c.dom.css.CSSValue.
const (
	CSSValueCssInherit        int16 = 0
	CSSValueCssPrimitiveValue int16 = 1
	CSSValueCssValueList      int16 = 2
	CSSValueCssCustom         int16 = 3
)
