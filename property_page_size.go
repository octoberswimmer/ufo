// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/PageSize.java

package ufo

type PageSize struct {
	width  *PropertyValue
	height *PropertyValue
}

func NewPageSize(width *PropertyValue, height *PropertyValue) *PageSize {
	return &PageSize{width: width, height: height}
}

func (p *PageSize) Width() *PropertyValue {
	return p.width
}

func (p *PageSize) Height() *PropertyValue {
	return p.height
}

// ISO A5 media: 148mm wide and 210 mm high
var PageSizeA5 = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 148.0, "148mm"),
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 210.0, "210mm"))

// IS0 A4 media: 210 mm wide and 297 mm high
var PageSizeA4 = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 210.0, "210mm"),
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 297.0, "297mm"))

// ISO A3 media: 297mm wide and 420mm high
var PageSizeA3 = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 297.0, "297mm"),
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 420.0, "420mm"))

// ISO B3 media: 176mm wide by 250mm high
var PageSizeB3 = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 176.0, "176mm"),
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 250, "250mm"))

// ISO B4 media: 250mm wide by 353mm high
var PageSizeB4 = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 250.0, "250mm"),
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 353.0, "353mm"))

// ISO B5 media: 176mm wide by 250 high
var PageSizeB5 = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 176.0, "176mm"),
	NewPropertyValueFloat(CSSPrimitiveValueCssMm, 250.0, "250mm"))

// North American letter media: 8.5 inches wide and 11 inches high
var PageSizeLetter = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssIn, 8.5, "8.5in"),
	NewPropertyValueFloat(CSSPrimitiveValueCssIn, 11.0, "11in"))

// North American legal: 8.5 inches wide by 14 inches high
var PageSizeLegal = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssIn, 8.5, "8.5in"),
	NewPropertyValueFloat(CSSPrimitiveValueCssIn, 14.0, "14in"))

// North American ledger: 11 inches wide by 17 inches high
var PageSizeLedger = NewPageSize(
	NewPropertyValueFloat(CSSPrimitiveValueCssIn, 11.0, "11in"),
	NewPropertyValueFloat(CSSPrimitiveValueCssIn, 17.0, "17in"))

var pageSizeSizeMap = map[string]*PageSize{
	"a3":     PageSizeA3,
	"a4":     PageSizeA4,
	"a5":     PageSizeA5,
	"b3":     PageSizeB3,
	"b4":     PageSizeB4,
	"b5":     PageSizeB5,
	"letter": PageSizeLetter,
	"legal":  PageSizeLegal,
	"ledger": PageSizeLedger,
}

// PageSizeGetPageSize returns nil when pageSize is not a known page size name.
func PageSizeGetPageSize(pageSize string) *PageSize {
	return pageSizeSizeMap[pageSize]
}
