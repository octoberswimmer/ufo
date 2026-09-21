// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/BuilderUtil.java

package ufo

import "fmt"

var builderUtilLengthValues = map[int16]struct{}{
	int16(CSSPrimitiveValueCssEms): {},
	int16(CSSPrimitiveValueCssExs): {},
	int16(CSSPrimitiveValueCssPx):  {},
	int16(CSSPrimitiveValueCssIn):  {},
	int16(CSSPrimitiveValueCssCm):  {},
	int16(CSSPrimitiveValueCssMm):  {},
	int16(CSSPrimitiveValueCssPt):  {},
	int16(CSSPrimitiveValueCssPc):  {},
}

func BuilderUtilIsLength(value *PropertyValue) bool {
	unit := value.GetPrimitiveType()
	_, ok := builderUtilLengthValues[int16(unit)]
	return ok ||
		unit == CSSPrimitiveValueCssNumber && value.GetFloatValueWithUnitType(CSSPrimitiveValueCssIn) == 0.0
}

func BuilderUtilCheckFunctionsAllowed(fn *FSFunction, allowed ...string) {
	for _, allow := range allowed {
		if allow == fn.GetName() {
			return
		}
	}

	panic(NewCSSParseException(fmt.Sprintf("Function %s not supported here", fn.GetName()), -1))
}
