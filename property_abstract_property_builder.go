// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/AbstractPropertyBuilder.java

package ufo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// AbstractPropertyBuilder is the base embedded by every property builder.
//
// The Java four-argument buildDeclarations calls the abstract five-argument
// one, so the struct holds self, the outermost builder, which every
// constructor of every builder sets.
type AbstractPropertyBuilder struct {
	self PropertyBuilder
}

func (b *AbstractPropertyBuilder) BuildDeclarations(cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool) []*PropertyDeclaration {
	return b.self.BuildDeclarationsWithInheritAllowed(cssName, values, origin, important, true)
}

func (b *AbstractPropertyBuilder) AssertFoundUpToValues(cssName *CSSName, values []*PropertyValue, max int) {
	found := len(values)
	if found < 1 || found > max {
		panic(NewCSSParseException(fmt.Sprintf("Found %d values for %s when between %d and %d value(s) were expected",
			found, cssName.ToString(), 1, max), -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentType(cssName *CSSName, value *PropertyValue) {
	if value.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentOrURIType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && typ != CSSPrimitiveValueCssUri {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or a URI", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentOrColorType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && typ != CSSPrimitiveValueCssRgbcolor {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or a color", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentOrIntegerType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent &&
		typ != CSSPrimitiveValueCssNumber ||
		typ == CSSPrimitiveValueCssNumber &&
			int(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) !=
				propertyBuilderRound(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or an integer", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckInteger(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssNumber ||
		int(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) != propertyBuilderRound(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an integer", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentOrLengthType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && !b.IsLength(value) {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or a length", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentOrNumberType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && typ != CSSPrimitiveValueCssNumber {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or a length", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentLengthOrPercentType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && !b.IsLength(value) && typ != CSSPrimitiveValueCssPercentage {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier, length, or percentage", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckLengthOrPercentType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if !b.IsLength(value) && typ != CSSPrimitiveValueCssPercentage {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be a length or percentage", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckLengthType(cssName *CSSName, value *PropertyValue) {
	if !b.IsLength(value) {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be a length", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckNumberType(cssName *CSSName, value *PropertyValue) {
	if value.GetPrimitiveType() != CSSPrimitiveValueCssNumber {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be a number", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentOrString(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssString && typ != CSSPrimitiveValueCssIdent {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or string", -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdentLengthNumberOrPercentType(cssName *CSSName, value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent &&
		!b.IsLength(value) &&
		typ != CSSPrimitiveValueCssPercentage &&
		typ != CSSPrimitiveValueCssNumber {
		panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier, length, or percentage", -1))
	}
}

func AbstractPropertyBuilderCheckValueBetween(cssName *CSSName, value float32, min float32, max float32) {
	if value > max || value < min {
		panic(NewCSSParseException(fmt.Sprintf("%s must be between %f and %f, but received: %f", cssName.ToString(), min, max, value), -1))
	}
}

func (b *AbstractPropertyBuilder) IsLength(value *PropertyValue) bool {
	unit := value.GetPrimitiveType()
	switch unit {
	case CSSPrimitiveValueCssEms, CSSPrimitiveValueCssExs, CSSPrimitiveValueCssPx, CSSPrimitiveValueCssIn,
		CSSPrimitiveValueCssCm, CSSPrimitiveValueCssMm, CSSPrimitiveValueCssPt, CSSPrimitiveValueCssPc:
		return true
	case CSSPrimitiveValueCssNumber:
		return value.GetFloatValueWithUnitType(CSSPrimitiveValueCssIn) == 0.0
	default:
		return false
	}
}

func (b *AbstractPropertyBuilder) CheckValidity(cssName *CSSName, validValues BitSet, value *IdentValue) {
	if !validValues.Get(value.FS_ID) {
		panic(NewCSSParseException("Ident "+value.ToString()+" is an invalid or unsupported value for "+cssName.ToString(), -1))
	}
}

func (b *AbstractPropertyBuilder) CheckIdent(value *PropertyValue) *IdentValue {
	result := IdentValueValueOf(value.GetStringValue())
	if result == nil {
		panic(NewCSSParseException("Value "+value.GetStringValue()+" is not a recognized identifier", -1))
	}
	value.SetIdentValue(result)
	return result
}

func (b *AbstractPropertyBuilder) CopyOf(decl *PropertyDeclaration, newName *CSSName) *PropertyDeclaration {
	return NewPropertyDeclaration(newName, decl.GetValue(), decl.IsImportant(), decl.GetOrigin())
}

func (b *AbstractPropertyBuilder) CheckInheritAllowed(value *PropertyValue, inheritAllowed bool) {
	if value.GetCssValueType() == CSSValueCssInherit && !inheritAllowed {
		panic(NewCSSParseException("Invalid use of inherit", -1))
	}
}

// CheckInheritAll returns nil when values is not a single inherit value.
func (b *AbstractPropertyBuilder) CheckInheritAll(all []*CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]
		b.CheckInheritAllowed(value, inheritAllowed)
		if value.GetCssValueType() == CSSValueCssInherit {
			result := make([]*PropertyDeclaration, 0, len(all))
			for _, cssName := range all {
				result = append(result, NewPropertyDeclaration(cssName, value, important, origin))
			}
			return result
		}
	}

	return nil
}

// propertyBuilderRound is Java's Math.round(float).
func propertyBuilderRound(f float32) int {
	return int(math.Floor(float64(f) + 0.5))
}

// propertyBuilderFloatToString formats f the way Java's Float.toString does,
// which the builders rely on when they concatenate a float into CSS text or
// an error message ("50.0%").
func propertyBuilderFloatToString(f float32) string {
	d := float64(f)
	switch {
	case math.IsNaN(d):
		return "NaN"
	case math.IsInf(d, 1):
		return "Infinity"
	case math.IsInf(d, -1):
		return "-Infinity"
	}
	abs := math.Abs(d)
	if abs != 0 && (abs < 1e-3 || abs >= 1e7) {
		s := strconv.FormatFloat(d, 'e', -1, 32)
		mantissa, exponent, _ := strings.Cut(s, "e")
		if !strings.Contains(mantissa, ".") {
			mantissa += ".0"
		}
		exp, _ := strconv.Atoi(exponent)
		return mantissa + "E" + strconv.Itoa(exp)
	}
	s := strconv.FormatFloat(d, 'f', -1, 32)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// propertyBuilderValueList converts values to the untyped list that
// NewPropertyValueList takes (Java's List<?>). The copy also stands in for the
// places where Java wraps the values in a new ArrayList.
func propertyBuilderValueList(values []*PropertyValue) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}
