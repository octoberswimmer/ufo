// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/PrimitivePropertyBuilders.java

package ufo

import (
	"strconv"
	"strings"
)

// BitSet stands in for java.util.BitSet as the property builders use it: a
// set of IdentValue FS_ID numbers.
type BitSet map[int]struct{}

func (s BitSet) Get(bitIndex int) bool {
	_, ok := s[bitIndex]
	return ok
}

func (s BitSet) Set(bitIndex int) {
	s[bitIndex] = struct{}{}
}

func (s BitSet) Or(set BitSet) {
	for bitIndex := range set {
		s[bitIndex] = struct{}{}
	}
}

// none | hidden | dotted | dashed | solid | double | groove | ridge | inset | outset
var PrimitivePropertyBuildersBorderStyles = primitivePropertyBuildersSetFor(
	IdentValueNone, IdentValueHidden, IdentValueDotted, IdentValueDashed, IdentValueSolid,
	IdentValueDouble, IdentValueGroove, IdentValueRidge, IdentValueInset, IdentValueOutset)

// thin | medium | thick
var PrimitivePropertyBuildersBorderWidths = primitivePropertyBuildersSetFor(IdentValueThin, IdentValueMedium, IdentValueThick)

// normal | small-caps | inherit
var PrimitivePropertyBuildersFontVariants = primitivePropertyBuildersSetFor(IdentValueNormal, IdentValueSmallCaps)

// normal | italic | oblique | inherit
var PrimitivePropertyBuildersFontStyles = primitivePropertyBuildersSetFor(IdentValueNormal, IdentValueItalic, IdentValueOblique)

var PrimitivePropertyBuildersFontWeights = primitivePropertyBuildersSetFor(IdentValueNormal, IdentValueBold, IdentValueBolder, IdentValueLighter)

var PrimitivePropertyBuildersPageOrientations = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValuePortrait, IdentValueLandscape)

// inside | outside | inherit
var PrimitivePropertyBuildersListStylePositions = primitivePropertyBuildersSetFor(IdentValueInside, IdentValueOutside)

// disc | circle | square | decimal
// | decimal-leading-zero | lower-roman | upper-roman
// | lower-greek | lower-latin | upper-latin | armenian
// | georgian | lower-alpha | upper-alpha | none | inherit
var PrimitivePropertyBuildersListStyleTypes = primitivePropertyBuildersSetFor(IdentValueDisc, IdentValueCircle, IdentValueSquare,
	IdentValueDecimal, IdentValueDecimalLeadingZero, IdentValueLowerRoman, IdentValueUpperRoman, IdentValueLowerGreek,
	IdentValueLowerLatin, IdentValueUpperLatin, IdentValueArmenian, IdentValueGeorgian, IdentValueLowerAlpha, IdentValueUpperAlpha, IdentValueNone)

// repeat | repeat-x | repeat-y | no-repeat | inherit
var PrimitivePropertyBuildersBackgroundRepeats = primitivePropertyBuildersSetFor(IdentValueRepeat, IdentValueRepeatX, IdentValueRepeatY, IdentValueNoRepeat)

// scroll | fixed | inherit
var PrimitivePropertyBuildersBackgroundAttachments = primitivePropertyBuildersSetFor(IdentValueScroll, IdentValueFixed)

// left | right | top | bottom | center
var PrimitivePropertyBuildersBackgroundPositions = primitivePropertyBuildersSetFor(IdentValueLeft, IdentValueRight, IdentValueTop, IdentValueBottom, IdentValueCenter)

var PrimitivePropertyBuildersAbsoluteFontSizes = primitivePropertyBuildersSetFor(IdentValueXxSmall, IdentValueXSmall, IdentValueSmall, IdentValueMedium, IdentValueLarge, IdentValueXLarge, IdentValueXxLarge)

var PrimitivePropertyBuildersRelativeFontSizes = primitivePropertyBuildersSetFor(IdentValueSmaller, IdentValueLarger)

// The Java constant PrimitivePropertyBuilders.COLOR would be named
// PrimitivePropertyBuildersColor, which is the Go name of the nested class
// PrimitivePropertyBuilders.Color, so the six builder constants carry a
// Builder suffix.
var PrimitivePropertyBuildersColorBuilder PropertyBuilder = newPrimitivePropertyBuildersGenericColor()
var PrimitivePropertyBuildersBorderStyleBuilder PropertyBuilder = newPrimitivePropertyBuildersGenericBorderStyle()
var PrimitivePropertyBuildersBorderWidthBuilder PropertyBuilder = newPrimitivePropertyBuildersGenericBorderWidth()
var PrimitivePropertyBuildersBorderRadiusBuilder PropertyBuilder = newPrimitivePropertyBuildersNonNegativeLengthLike()
var PrimitivePropertyBuildersMarginBuilder PropertyBuilder = newPrimitivePropertyBuildersLengthLikeWithAuto()
var PrimitivePropertyBuildersPaddingBuilder PropertyBuilder = newPrimitivePropertyBuildersNonNegativeLengthLike()

func primitivePropertyBuildersSetFor(values ...*IdentValue) BitSet {
	result := make(BitSet, len(values))
	for _, ident := range values {
		result.Set(ident.FS_ID)
	}
	return result
}

func primitivePropertyBuildersAssertFoundSingleValue(cssName *CSSName, values []*PropertyValue) {
	if len(values) != 1 {
		panic(NewCSSParseException("Found "+strconv.Itoa(len(values))+" value(s) for "+
			cssName.ToString()+" when "+strconv.Itoa(1)+" value(s) were expected", -1))
	}
}

// primitivePropertyBuildersSingleIdent is the Java abstract class SingleIdent.
// The set returned by the abstract getAllowed() is the allowed field.
type primitivePropertyBuildersSingleIdent struct {
	AbstractPropertyBuilder
	allowed BitSet
}

func newPrimitivePropertyBuildersSingleIdent(allowed BitSet) *primitivePropertyBuildersSingleIdent {
	b := &primitivePropertyBuildersSingleIdent{allowed: allowed}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersSingleIdent) getAllowed() BitSet {
	return b.allowed
}

func (b *primitivePropertyBuildersSingleIdent) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentType(cssName, value)
		ident := b.CheckIdent(value)

		b.CheckValidity(cssName, b.getAllowed(), ident)
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

var primitivePropertyBuildersGenericColorAllowed = primitivePropertyBuildersSetFor(IdentValueTransparent)

type primitivePropertyBuildersGenericColor struct {
	AbstractPropertyBuilder
}

func newPrimitivePropertyBuildersGenericColor() *primitivePropertyBuildersGenericColor {
	b := &primitivePropertyBuildersGenericColor{}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersGenericColor) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrColorType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			color := ConversionsGetColor(value.GetStringValue())
			if color != nil {
				return []*PropertyDeclaration{
					NewPropertyDeclaration(
						cssName,
						NewPropertyValueFSColor(color),
						important,
						origin)}
			}

			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, primitivePropertyBuildersGenericColorAllowed, ident)
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

type primitivePropertyBuildersGenericBorderStyle struct {
	primitivePropertyBuildersSingleIdent
}

func newPrimitivePropertyBuildersGenericBorderStyle() *primitivePropertyBuildersGenericBorderStyle {
	b := &primitivePropertyBuildersGenericBorderStyle{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersBorderStyles)}
	b.self = b
	return b
}

type primitivePropertyBuildersGenericBorderWidth struct {
	AbstractPropertyBuilder
}

func newPrimitivePropertyBuildersGenericBorderWidth() *primitivePropertyBuildersGenericBorderWidth {
	b := &primitivePropertyBuildersGenericBorderWidth{}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersGenericBorderWidth) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrLengthType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, PrimitivePropertyBuildersBorderWidths, ident)

			return []*PropertyDeclaration{
				NewPropertyDeclaration(
					cssName, ConversionsGetBorderWidth(ident.ToString()), important, origin)}
		} else {
			if value.GetFloatValue() < 0.0 {
				panic(NewCSSParseException(cssName.ToString()+" may not be negative", -1))
			}
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

type primitivePropertyBuildersGenericBorderCornerRadius struct {
	AbstractPropertyBuilder
}

func newPrimitivePropertyBuildersGenericBorderCornerRadius() *primitivePropertyBuildersGenericBorderCornerRadius {
	b := &primitivePropertyBuildersGenericBorderCornerRadius{}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersGenericBorderCornerRadius) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	b.AssertFoundUpToValues(cssName, values, 2)

	first := values[0]
	var second *PropertyValue
	if len(values) == 2 {
		second = values[1]
	}

	b.CheckInheritAllowed(first, inheritAllowed)

	if second != nil {
		b.CheckInheritAllowed(second, false)
	}

	b.CheckLengthOrPercentType(cssName, first)
	if second == nil {
		return primitivePropertyBuildersCreateTwoValueResponse(cssName, first, first, origin, important)
	} else {
		b.CheckLengthOrPercentType(cssName, second)
		return primitivePropertyBuildersCreateTwoValueResponse(cssName, first, second, origin, important)
	}
}

// primitivePropertyBuildersLengthWithIdent is the Java abstract class
// LengthWithIdent. The set returned by the abstract getAllowed() is the
// allowed field, and a subclass that overrides isNegativeValuesAllowed() sets
// negativeValuesAllowed in its constructor.
type primitivePropertyBuildersLengthWithIdent struct {
	AbstractPropertyBuilder
	allowed               BitSet
	negativeValuesAllowed bool
}

func newPrimitivePropertyBuildersLengthWithIdent(allowed BitSet) *primitivePropertyBuildersLengthWithIdent {
	b := &primitivePropertyBuildersLengthWithIdent{allowed: allowed, negativeValuesAllowed: true}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersLengthWithIdent) getAllowed() BitSet {
	return b.allowed
}

func (b *primitivePropertyBuildersLengthWithIdent) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrLengthType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, b.getAllowed(), ident)
		} else if !b.isNegativeValuesAllowed() && value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" may not be negative", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *primitivePropertyBuildersLengthWithIdent) isNegativeValuesAllowed() bool {
	return b.negativeValuesAllowed
}

// primitivePropertyBuildersLengthLikeWithIdent is the Java abstract class
// LengthLikeWithIdent, with the same fields as
// primitivePropertyBuildersLengthWithIdent.
type primitivePropertyBuildersLengthLikeWithIdent struct {
	AbstractPropertyBuilder
	allowed               BitSet
	negativeValuesAllowed bool
}

func newPrimitivePropertyBuildersLengthLikeWithIdent(allowed BitSet) *primitivePropertyBuildersLengthLikeWithIdent {
	b := &primitivePropertyBuildersLengthLikeWithIdent{allowed: allowed, negativeValuesAllowed: true}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersLengthLikeWithIdent) getAllowed() BitSet {
	return b.allowed
}

func (b *primitivePropertyBuildersLengthLikeWithIdent) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentLengthOrPercentType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, b.getAllowed(), ident)
		} else if !b.isNegativeValuesAllowed() && value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" may not be negative", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *primitivePropertyBuildersLengthLikeWithIdent) isNegativeValuesAllowed() bool {
	return b.negativeValuesAllowed
}

// primitivePropertyBuildersLengthLike is the Java class LengthLike. A subclass
// that overrides isNegativeValuesAllowed() sets negativeValuesAllowed in its
// constructor.
type primitivePropertyBuildersLengthLike struct {
	AbstractPropertyBuilder
	negativeValuesAllowed bool
}

func newPrimitivePropertyBuildersLengthLike() *primitivePropertyBuildersLengthLike {
	b := &primitivePropertyBuildersLengthLike{negativeValuesAllowed: true}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersLengthLike) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckLengthOrPercentType(cssName, value)

		if !b.isNegativeValuesAllowed() && value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" may not be negative", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *primitivePropertyBuildersLengthLike) isNegativeValuesAllowed() bool {
	return b.negativeValuesAllowed
}

type primitivePropertyBuildersNonNegativeLengthLike struct {
	primitivePropertyBuildersLengthLike
}

func newPrimitivePropertyBuildersNonNegativeLengthLike() *primitivePropertyBuildersNonNegativeLengthLike {
	b := &primitivePropertyBuildersNonNegativeLengthLike{*newPrimitivePropertyBuildersLengthLike()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type primitivePropertyBuildersColOrRowSpan struct {
	AbstractPropertyBuilder
}

func newPrimitivePropertyBuildersColOrRowSpan() *primitivePropertyBuildersColOrRowSpan {
	b := &primitivePropertyBuildersColOrRowSpan{}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersColOrRowSpan) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckNumberType(cssName, value)

		if value.GetFloatValue() < 1 {
			panic(NewCSSParseException("colspan/rowspan must be greater than zero", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

// primitivePropertyBuildersPlainInteger is the Java class PlainInteger. A
// subclass that overrides isNegativeValuesAllowed() sets negativeValuesAllowed
// in its constructor.
type primitivePropertyBuildersPlainInteger struct {
	AbstractPropertyBuilder
	negativeValuesAllowed bool
}

func newPrimitivePropertyBuildersPlainInteger() *primitivePropertyBuildersPlainInteger {
	b := &primitivePropertyBuildersPlainInteger{negativeValuesAllowed: true}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersPlainInteger) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckInteger(cssName, value)

		if !b.isNegativeValuesAllowed() && value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" may not be negative", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *primitivePropertyBuildersPlainInteger) isNegativeValuesAllowed() bool {
	return b.negativeValuesAllowed
}

// primitivePropertyBuildersLength is the Java class Length.
type primitivePropertyBuildersLength struct {
	AbstractPropertyBuilder
	negativeValuesAllowed bool
}

func newPrimitivePropertyBuildersLength() *primitivePropertyBuildersLength {
	b := &primitivePropertyBuildersLength{negativeValuesAllowed: true}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersLength) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckLengthType(cssName, value)

		if !b.isNegativeValuesAllowed() && value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" may not be negative", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *primitivePropertyBuildersLength) isNegativeValuesAllowed() bool {
	return b.negativeValuesAllowed
}

// The Java file has the classes SingleString, SingleStringWithIdent and
// SingleStringWithNone commented out at this point; they are not ported.

// <length> | <percentage> | auto | inherit
var primitivePropertyBuildersLengthLikeWithAutoAllowed = primitivePropertyBuildersSetFor(IdentValueAuto)

type primitivePropertyBuildersLengthLikeWithAuto struct {
	primitivePropertyBuildersLengthLikeWithIdent
}

func newPrimitivePropertyBuildersLengthLikeWithAuto() *primitivePropertyBuildersLengthLikeWithAuto {
	b := &primitivePropertyBuildersLengthLikeWithAuto{*newPrimitivePropertyBuildersLengthLikeWithIdent(primitivePropertyBuildersLengthLikeWithAutoAllowed)}
	b.self = b
	return b
}

// <length> | normal | inherit
var primitivePropertyBuildersLengthWithNormalAllowed = primitivePropertyBuildersSetFor(IdentValueNormal)

type primitivePropertyBuildersLengthWithNormal struct {
	primitivePropertyBuildersLengthWithIdent
}

func newPrimitivePropertyBuildersLengthWithNormal() *primitivePropertyBuildersLengthWithNormal {
	b := &primitivePropertyBuildersLengthWithNormal{*newPrimitivePropertyBuildersLengthWithIdent(primitivePropertyBuildersLengthWithNormalAllowed)}
	b.self = b
	return b
}

// <length> | <percentage> | none | inherit
var primitivePropertyBuildersLengthLikeWithNoneAllowed = primitivePropertyBuildersSetFor(IdentValueNone)

type primitivePropertyBuildersLengthLikeWithNone struct {
	primitivePropertyBuildersLengthLikeWithIdent
}

func newPrimitivePropertyBuildersLengthLikeWithNone() *primitivePropertyBuildersLengthLikeWithNone {
	b := &primitivePropertyBuildersLengthLikeWithNone{*newPrimitivePropertyBuildersLengthLikeWithIdent(primitivePropertyBuildersLengthLikeWithNoneAllowed)}
	b.self = b
	return b
}

// <uri> | none | inherit
var primitivePropertyBuildersGenericURIWithNoneAllowed = primitivePropertyBuildersSetFor(IdentValueNone)

type primitivePropertyBuildersGenericURIWithNone struct {
	AbstractPropertyBuilder
}

func newPrimitivePropertyBuildersGenericURIWithNone() *primitivePropertyBuildersGenericURIWithNone {
	b := &primitivePropertyBuildersGenericURIWithNone{}
	b.self = b
	return b
}

func (b *primitivePropertyBuildersGenericURIWithNone) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrURIType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, primitivePropertyBuildersGenericURIWithNoneAllowed, ident)
		}
	}
	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

type PrimitivePropertyBuildersBackgroundAttachment struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersBackgroundAttachment() *PrimitivePropertyBuildersBackgroundAttachment {
	b := &PrimitivePropertyBuildersBackgroundAttachment{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersBackgroundAttachments)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBackgroundColor struct {
	primitivePropertyBuildersGenericColor
}

func NewPrimitivePropertyBuildersBackgroundColor() *PrimitivePropertyBuildersBackgroundColor {
	b := &PrimitivePropertyBuildersBackgroundColor{*newPrimitivePropertyBuildersGenericColor()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBackgroundImage struct {
	primitivePropertyBuildersGenericURIWithNone
}

func NewPrimitivePropertyBuildersBackgroundImage() *PrimitivePropertyBuildersBackgroundImage {
	b := &PrimitivePropertyBuildersBackgroundImage{*newPrimitivePropertyBuildersGenericURIWithNone()}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersBackgroundImage) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {

	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]

	if !strings.HasPrefix(value.ToString(), IdentValueLinearGradient.AsString()) {
		return b.primitivePropertyBuildersGenericURIWithNone.BuildDeclarationsWithInheritAllowed(cssName, values, origin, important, inheritAllowed)
	}

	BuilderUtilCheckFunctionsAllowed(value.GetFunction(), "linear-gradient")
	return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
}

var primitivePropertyBuildersBackgroundSizeAllAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueContain, IdentValueCover)

type PrimitivePropertyBuildersBackgroundSize struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersBackgroundSize() *PrimitivePropertyBuildersBackgroundSize {
	b := &PrimitivePropertyBuildersBackgroundSize{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersBackgroundSize) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	b.AssertFoundUpToValues(cssName, values, 2)

	first := values[0]
	var second *PropertyValue
	if len(values) == 2 {
		second = values[1]
	}

	b.CheckInheritAllowed(first, inheritAllowed)
	if len(values) == 1 &&
		first.GetCssValueType() == CSSValueCssInherit {
		return []*PropertyDeclaration{
			NewPropertyDeclaration(cssName, first, important, origin)}
	}

	if second != nil {
		b.CheckInheritAllowed(second, false)
	}

	b.CheckIdentLengthOrPercentType(cssName, first)
	if second == nil {
		if first.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			firstIdent := b.CheckIdent(first)
			b.CheckValidity(cssName, primitivePropertyBuildersBackgroundSizeAllAllowed, firstIdent)

			if firstIdent == IdentValueContain || firstIdent == IdentValueCover {
				return []*PropertyDeclaration{
					NewPropertyDeclaration(cssName, first, important, origin)}
			} else {
				return primitivePropertyBuildersCreateTwoValueResponse(CSSNameBackgroundSize, first, first, origin, important)
			}
		} else {
			return primitivePropertyBuildersCreateTwoValueResponse(CSSNameBackgroundSize, first, NewPropertyValueIdentValue(IdentValueAuto), origin, important)
		}
	} else {
		b.CheckIdentLengthOrPercentType(cssName, second)

		if first.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			firstIdent := b.CheckIdent(first)
			if firstIdent != IdentValueAuto {
				panic(NewCSSParseException("The only ident value allowed here is 'auto'", -1))
			}
		} else if first.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" values cannot be negative", -1))
		}

		if second.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			secondIdent := b.CheckIdent(second)
			if secondIdent != IdentValueAuto {
				panic(NewCSSParseException("The only ident value allowed here is 'auto'", -1))
			}
		} else if second.GetFloatValue() < 0.0 {
			panic(NewCSSParseException(cssName.ToString()+" values cannot be negative", -1))
		}

		return primitivePropertyBuildersCreateTwoValueResponse(CSSNameBackgroundSize, first, second, origin, important)
	}
}

type PrimitivePropertyBuildersBackgroundPosition struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersBackgroundPosition() *PrimitivePropertyBuildersBackgroundPosition {
	b := &PrimitivePropertyBuildersBackgroundPosition{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersBackgroundPosition) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	b.AssertFoundUpToValues(cssName, values, 2)

	first := values[0]
	var second *PropertyValue
	if len(values) == 2 {
		second = values[1]
	}

	b.CheckInheritAllowed(first, inheritAllowed)
	if len(values) == 1 &&
		first.GetCssValueType() == CSSValueCssInherit {
		return []*PropertyDeclaration{
			NewPropertyDeclaration(cssName, first, important, origin)}
	}

	if second != nil {
		b.CheckInheritAllowed(second, false)
	}

	b.CheckIdentLengthOrPercentType(cssName, first)
	if second == nil {
		if b.IsLength(first) || first.GetPrimitiveType() == CSSPrimitiveValueCssPercentage {
			responseValues := make([]*PropertyValue, 0, 2)
			responseValues = append(responseValues, first)
			responseValues = append(responseValues, NewPropertyValueFloat(
				CSSPrimitiveValueCssPercentage, 50.0, "50%"))
			return []*PropertyDeclaration{NewPropertyDeclaration(
				CSSNameBackgroundPosition,
				NewPropertyValueList(propertyBuilderValueList(responseValues)), important, origin)}
		}
	} else {
		b.CheckIdentLengthOrPercentType(cssName, second)
	}

	var firstIdent *IdentValue
	if first.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
		firstIdent = b.CheckIdent(first)
		b.CheckValidity(cssName, b.getAllowed(), firstIdent)
	}

	var secondIdent *IdentValue
	if second == nil {
		secondIdent = IdentValueCenter
	} else if second.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
		secondIdent = b.CheckIdent(second)
		b.CheckValidity(cssName, b.getAllowed(), secondIdent)
	}

	if firstIdent == nil && secondIdent == nil {
		return []*PropertyDeclaration{NewPropertyDeclaration(
			CSSNameBackgroundPosition, NewPropertyValueList(propertyBuilderValueList(values)), important, origin)}
	} else if firstIdent != nil && secondIdent != nil {
		if firstIdent == IdentValueTop || firstIdent == IdentValueBottom ||
			secondIdent == IdentValueLeft || secondIdent == IdentValueRight {
			temp := firstIdent
			firstIdent = secondIdent
			secondIdent = temp
		}

		b.checkIdentPosition(cssName, firstIdent, secondIdent)

		return b.createTwoPercentValueResponse(
			b.getPercentForIdent(firstIdent),
			b.getPercentForIdent(secondIdent),
			important,
			origin)
	} else {
		b.checkIdentPosition(cssName, firstIdent, secondIdent)

		responseValues := make([]*PropertyValue, 0, 2)

		if firstIdent == nil {
			responseValues = append(responseValues, first)
			responseValues = append(responseValues, b.createValueForIdent(secondIdent))
		} else {
			responseValues = append(responseValues, b.createValueForIdent(firstIdent))
			responseValues = append(responseValues, second)
		}

		return []*PropertyDeclaration{NewPropertyDeclaration(
			CSSNameBackgroundPosition,
			NewPropertyValueList(propertyBuilderValueList(responseValues)), important, origin)}
	}
}

func (b *PrimitivePropertyBuildersBackgroundPosition) checkIdentPosition(cssName *CSSName, firstIdent *IdentValue, secondIdent *IdentValue) {
	if firstIdent == IdentValueTop || firstIdent == IdentValueBottom ||
		secondIdent == IdentValueLeft || secondIdent == IdentValueRight {
		panic(NewCSSParseException("Invalid combination of keywords in "+cssName.ToString(), -1))
	}
}

func (b *PrimitivePropertyBuildersBackgroundPosition) getPercentForIdent(ident *IdentValue) float32 {
	var percent float32 = 0.0

	if ident == IdentValueCenter {
		percent = 50.0
	} else if ident == IdentValueBottom || ident == IdentValueRight {
		percent = 100.0
	}

	return percent
}

func (b *PrimitivePropertyBuildersBackgroundPosition) createValueForIdent(ident *IdentValue) *PropertyValue {
	percent := b.getPercentForIdent(ident)
	return NewPropertyValueFloat(
		CSSPrimitiveValueCssPercentage, percent, propertyBuilderFloatToString(percent)+"%")
}

func (b *PrimitivePropertyBuildersBackgroundPosition) createTwoPercentValueResponse(
	percent1 float32, percent2 float32, important bool, origin StylesheetInfoOrigin) []*PropertyDeclaration {
	value1 := NewPropertyValueFloat(
		CSSPrimitiveValueCssPercentage, percent1, propertyBuilderFloatToString(percent1)+"%")
	value2 := NewPropertyValueFloat(
		CSSPrimitiveValueCssPercentage, percent2, propertyBuilderFloatToString(percent2)+"%")

	values := make([]*PropertyValue, 0, 2)
	values = append(values, value1)
	values = append(values, value2)

	result := NewPropertyDeclaration(
		CSSNameBackgroundPosition,
		NewPropertyValueList(propertyBuilderValueList(values)), important, origin)

	return []*PropertyDeclaration{result}
}

func (b *PrimitivePropertyBuildersBackgroundPosition) getAllowed() BitSet {
	return PrimitivePropertyBuildersBackgroundPositions
}

type PrimitivePropertyBuildersBackgroundRepeat struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersBackgroundRepeat() *PrimitivePropertyBuildersBackgroundRepeat {
	b := &PrimitivePropertyBuildersBackgroundRepeat{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersBackgroundRepeats)}
	b.self = b
	return b
}

// collapse | separate | inherit
var primitivePropertyBuildersBorderCollapseAllowed = primitivePropertyBuildersSetFor(IdentValueCollapse, IdentValueSeparate)

type PrimitivePropertyBuildersBorderCollapse struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersBorderCollapse() *PrimitivePropertyBuildersBorderCollapse {
	b := &PrimitivePropertyBuildersBorderCollapse{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersBorderCollapseAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderTopColor struct {
	primitivePropertyBuildersGenericColor
}

func NewPrimitivePropertyBuildersBorderTopColor() *PrimitivePropertyBuildersBorderTopColor {
	b := &PrimitivePropertyBuildersBorderTopColor{*newPrimitivePropertyBuildersGenericColor()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderRightColor struct {
	primitivePropertyBuildersGenericColor
}

func NewPrimitivePropertyBuildersBorderRightColor() *PrimitivePropertyBuildersBorderRightColor {
	b := &PrimitivePropertyBuildersBorderRightColor{*newPrimitivePropertyBuildersGenericColor()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderBottomColor struct {
	primitivePropertyBuildersGenericColor
}

func NewPrimitivePropertyBuildersBorderBottomColor() *PrimitivePropertyBuildersBorderBottomColor {
	b := &PrimitivePropertyBuildersBorderBottomColor{*newPrimitivePropertyBuildersGenericColor()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderLeftColor struct {
	primitivePropertyBuildersGenericColor
}

func NewPrimitivePropertyBuildersBorderLeftColor() *PrimitivePropertyBuildersBorderLeftColor {
	b := &PrimitivePropertyBuildersBorderLeftColor{*newPrimitivePropertyBuildersGenericColor()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderTopStyle struct {
	primitivePropertyBuildersGenericBorderStyle
}

func NewPrimitivePropertyBuildersBorderTopStyle() *PrimitivePropertyBuildersBorderTopStyle {
	b := &PrimitivePropertyBuildersBorderTopStyle{*newPrimitivePropertyBuildersGenericBorderStyle()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderRightStyle struct {
	primitivePropertyBuildersGenericBorderStyle
}

func NewPrimitivePropertyBuildersBorderRightStyle() *PrimitivePropertyBuildersBorderRightStyle {
	b := &PrimitivePropertyBuildersBorderRightStyle{*newPrimitivePropertyBuildersGenericBorderStyle()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderBottomStyle struct {
	primitivePropertyBuildersGenericBorderStyle
}

func NewPrimitivePropertyBuildersBorderBottomStyle() *PrimitivePropertyBuildersBorderBottomStyle {
	b := &PrimitivePropertyBuildersBorderBottomStyle{*newPrimitivePropertyBuildersGenericBorderStyle()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderLeftStyle struct {
	primitivePropertyBuildersGenericBorderStyle
}

func NewPrimitivePropertyBuildersBorderLeftStyle() *PrimitivePropertyBuildersBorderLeftStyle {
	b := &PrimitivePropertyBuildersBorderLeftStyle{*newPrimitivePropertyBuildersGenericBorderStyle()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderTopWidth struct {
	primitivePropertyBuildersGenericBorderWidth
}

func NewPrimitivePropertyBuildersBorderTopWidth() *PrimitivePropertyBuildersBorderTopWidth {
	b := &PrimitivePropertyBuildersBorderTopWidth{*newPrimitivePropertyBuildersGenericBorderWidth()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderRightWidth struct {
	primitivePropertyBuildersGenericBorderWidth
}

func NewPrimitivePropertyBuildersBorderRightWidth() *PrimitivePropertyBuildersBorderRightWidth {
	b := &PrimitivePropertyBuildersBorderRightWidth{*newPrimitivePropertyBuildersGenericBorderWidth()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderBottomWidth struct {
	primitivePropertyBuildersGenericBorderWidth
}

func NewPrimitivePropertyBuildersBorderBottomWidth() *PrimitivePropertyBuildersBorderBottomWidth {
	b := &PrimitivePropertyBuildersBorderBottomWidth{*newPrimitivePropertyBuildersGenericBorderWidth()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderLeftWidth struct {
	primitivePropertyBuildersGenericBorderWidth
}

func NewPrimitivePropertyBuildersBorderLeftWidth() *PrimitivePropertyBuildersBorderLeftWidth {
	b := &PrimitivePropertyBuildersBorderLeftWidth{*newPrimitivePropertyBuildersGenericBorderWidth()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderTopLeftRadius struct {
	primitivePropertyBuildersGenericBorderCornerRadius
}

func NewPrimitivePropertyBuildersBorderTopLeftRadius() *PrimitivePropertyBuildersBorderTopLeftRadius {
	b := &PrimitivePropertyBuildersBorderTopLeftRadius{*newPrimitivePropertyBuildersGenericBorderCornerRadius()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderTopRightRadius struct {
	primitivePropertyBuildersGenericBorderCornerRadius
}

func NewPrimitivePropertyBuildersBorderTopRightRadius() *PrimitivePropertyBuildersBorderTopRightRadius {
	b := &PrimitivePropertyBuildersBorderTopRightRadius{*newPrimitivePropertyBuildersGenericBorderCornerRadius()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderBottomRightRadius struct {
	primitivePropertyBuildersGenericBorderCornerRadius
}

func NewPrimitivePropertyBuildersBorderBottomRightRadius() *PrimitivePropertyBuildersBorderBottomRightRadius {
	b := &PrimitivePropertyBuildersBorderBottomRightRadius{*newPrimitivePropertyBuildersGenericBorderCornerRadius()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBorderBottomLeftRadius struct {
	primitivePropertyBuildersGenericBorderCornerRadius
}

func NewPrimitivePropertyBuildersBorderBottomLeftRadius() *PrimitivePropertyBuildersBorderBottomLeftRadius {
	b := &PrimitivePropertyBuildersBorderBottomLeftRadius{*newPrimitivePropertyBuildersGenericBorderCornerRadius()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersBottom struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersBottom() *PrimitivePropertyBuildersBottom {
	b := &PrimitivePropertyBuildersBottom{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

// top | bottom | inherit
var primitivePropertyBuildersCaptionSideAllowed = primitivePropertyBuildersSetFor(IdentValueTop, IdentValueBottom)

type PrimitivePropertyBuildersCaptionSide struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersCaptionSide() *PrimitivePropertyBuildersCaptionSide {
	b := &PrimitivePropertyBuildersCaptionSide{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersCaptionSideAllowed)}
	b.self = b
	return b
}

// none | left | right | both | inherit
var primitivePropertyBuildersClearAllowed = primitivePropertyBuildersSetFor(IdentValueNone, IdentValueLeft, IdentValueRight, IdentValueBoth)

type PrimitivePropertyBuildersClear struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersClear() *PrimitivePropertyBuildersClear {
	b := &PrimitivePropertyBuildersClear{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersClearAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersColor struct {
	primitivePropertyBuildersGenericColor
}

func NewPrimitivePropertyBuildersColor() *PrimitivePropertyBuildersColor {
	b := &PrimitivePropertyBuildersColor{*newPrimitivePropertyBuildersGenericColor()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersColumnCount struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersColumnCount() *PrimitivePropertyBuildersColumnCount {
	b := &PrimitivePropertyBuildersColumnCount{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersColumnCount) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrIntegerType(cssName, value)
		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			if ident != IdentValueAuto {
				panic(NewCSSParseException("Only auto is allowed for "+cssName.ToString(), -1))
			}
		} else if value.GetFloatValue() < 1.0 {
			panic(NewCSSParseException(cssName.ToString()+" must be at least 1", -1))
		}
	}
	return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
}

type PrimitivePropertyBuildersColumnGap struct {
	primitivePropertyBuildersLengthWithNormal
}

func NewPrimitivePropertyBuildersColumnGap() *PrimitivePropertyBuildersColumnGap {
	b := &PrimitivePropertyBuildersColumnGap{*newPrimitivePropertyBuildersLengthWithNormal()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

var primitivePropertyBuildersColumnWidthAllowed = primitivePropertyBuildersSetFor(IdentValueAuto)

type PrimitivePropertyBuildersColumnWidth struct {
	primitivePropertyBuildersLengthWithIdent
}

func NewPrimitivePropertyBuildersColumnWidth() *PrimitivePropertyBuildersColumnWidth {
	b := &PrimitivePropertyBuildersColumnWidth{*newPrimitivePropertyBuildersLengthWithIdent(primitivePropertyBuildersColumnWidthAllowed)}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

// [ [<uri> ,]* [ auto | crosshair | default | pointer | move | e-resize
// | ne-resize | nw-resize | n-resize | se-resize | sw-resize | s-resize
// | w-resize | text | wait | help | progress ] ] | inherit
var primitivePropertyBuildersCursorAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueCrosshair, IdentValueDefault, IdentValuePointer, IdentValueMove, IdentValueEResize, IdentValueNeResize, IdentValueNwResize, IdentValueNResize, IdentValueSeResize, IdentValueSwResize, IdentValueSResize, IdentValueWResize, IdentValueText, IdentValueWait, IdentValueHelp, IdentValueProgress)

type PrimitivePropertyBuildersCursor struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersCursor() *PrimitivePropertyBuildersCursor {
	b := &PrimitivePropertyBuildersCursor{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersCursorAllowed)}
	b.self = b
	return b
}

// inline | block | list-item | run-in | inline-block | table | inline-table
// | table-row-group | table-header-group
// | table-footer-group | table-row | table-column-group | table-column
// | table-cell | table-caption | none | inherit
// (run-in is not in the allowed set)
var primitivePropertyBuildersDisplayAllowed = primitivePropertyBuildersSetFor(IdentValueInline, IdentValueBlock, IdentValueListItem, IdentValueInlineBlock, IdentValueTable, IdentValueInlineTable, IdentValueTableRowGroup, IdentValueTableHeaderGroup, IdentValueTableFooterGroup, IdentValueTableRow, IdentValueTableColumnGroup, IdentValueTableColumn, IdentValueTableCell, IdentValueTableCaption, IdentValueNone)

type PrimitivePropertyBuildersDisplay struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersDisplay() *PrimitivePropertyBuildersDisplay {
	b := &PrimitivePropertyBuildersDisplay{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersDisplayAllowed)}
	b.self = b
	return b
}

// show | hide | inherit
var primitivePropertyBuildersEmptyCellsAllowed = primitivePropertyBuildersSetFor(IdentValueShow, IdentValueHide)

type PrimitivePropertyBuildersEmptyCells struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersEmptyCells() *PrimitivePropertyBuildersEmptyCells {
	b := &PrimitivePropertyBuildersEmptyCells{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersEmptyCellsAllowed)}
	b.self = b
	return b
}

// left | right | none | inherit
var primitivePropertyBuildersFloatAllowed = primitivePropertyBuildersSetFor(IdentValueLeft, IdentValueRight, IdentValueNone)

type PrimitivePropertyBuildersFloat struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFloat() *PrimitivePropertyBuildersFloat {
	b := &PrimitivePropertyBuildersFloat{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFloatAllowed)}
	b.self = b
	return b
}

// [[ <family-name> | <generic-family> ] [, <family-name>| <generic-family>]* ] | inherit
type PrimitivePropertyBuildersFontFamily struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersFontFamily() *PrimitivePropertyBuildersFontFamily {
	b := &PrimitivePropertyBuildersFontFamily{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersFontFamily) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]
		b.CheckInheritAllowed(value, inheritAllowed)
		if value.GetCssValueType() == CSSValueCssInherit {
			return []*PropertyDeclaration{
				NewPropertyDeclaration(cssName, value, important, origin)}
		}
	}

	// Both Opera and Firefox parse "Century Gothic" Arial sans-serif as
	// [Century Gothic], [Arial sans-serif] (i.e. the comma is assumed
	// after a string).  Seems wrong per the spec, but FF (at least)
	// does it in standards mode, so we do too.
	consecutiveIdents := []string{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		operator := value.GetOperator()
		if operator != nil && operator != TokenTkComma {
			panic(NewCSSParseException("Invalid font-family definition", -1))
		}

		if operator != nil {
			if len(consecutiveIdents) != 0 {
				normalized = append(normalized, b.concat(consecutiveIdents, ' '))
				consecutiveIdents = consecutiveIdents[:0]
			}
		}

		b.CheckInheritAllowed(value, false)
		typ := value.GetPrimitiveType()
		if typ == CSSPrimitiveValueCssString {
			if len(consecutiveIdents) != 0 {
				normalized = append(normalized, b.concat(consecutiveIdents, ' '))
				consecutiveIdents = consecutiveIdents[:0]
			}
			normalized = append(normalized, value.GetStringValue())
		} else if typ == CSSPrimitiveValueCssIdent {
			consecutiveIdents = append(consecutiveIdents, value.GetStringValue())
		} else {
			panic(NewCSSParseException("Invalid font-family definition", -1))
		}
	}
	if len(consecutiveIdents) != 0 {
		normalized = append(normalized, b.concat(consecutiveIdents, ' '))
	}

	text := b.concat(normalized, ',')
	result := NewPropertyValueStringWithStringArrayValueOperatorToken(
		CSSPrimitiveValueCssString, text, text, normalized, nil) // HACK cssText can be wrong

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, result, important, origin)}
}

func (b *PrimitivePropertyBuildersFontFamily) concat(strs []string, separator rune) string {
	var buf strings.Builder
	for i, s := range strs {
		buf.WriteString(s)
		if i < len(strs)-1 {
			buf.WriteRune(separator)
		}
	}
	return buf.String()
}

// <absolute-size> | <relative-size> | <length> | <percentage> | inherit
var primitivePropertyBuildersFontSizeAllowed = func() BitSet {
	allowed := BitSet{}
	allowed.Or(PrimitivePropertyBuildersAbsoluteFontSizes)
	allowed.Or(PrimitivePropertyBuildersRelativeFontSizes)
	return allowed
}()

type PrimitivePropertyBuildersFontSize struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersFontSize() *PrimitivePropertyBuildersFontSize {
	b := &PrimitivePropertyBuildersFontSize{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersFontSize) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentLengthOrPercentType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, primitivePropertyBuildersFontSizeAllowed, ident)
		} else if value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException("font-size may not be negative", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

type PrimitivePropertyBuildersFontStyle struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFontStyle() *PrimitivePropertyBuildersFontStyle {
	b := &PrimitivePropertyBuildersFontStyle{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersFontStyles)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFontVariant struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFontVariant() *PrimitivePropertyBuildersFontVariant {
	b := &PrimitivePropertyBuildersFontVariant{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersFontVariants)}
	b.self = b
	return b
}

// normal | bold | bolder | lighter | 100 | 200 | 300 | 400 | 500 | 600 | 700 | 800 | 900 | inherit
type PrimitivePropertyBuildersFontWeight struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersFontWeight() *PrimitivePropertyBuildersFontWeight {
	b := &PrimitivePropertyBuildersFontWeight{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersFontWeight) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrNumberType(cssName, value)

		typ := value.GetPrimitiveType()
		if typ == CSSPrimitiveValueCssIdent {
			b.CheckIdentType(cssName, value)
			ident := b.CheckIdent(value)

			b.CheckValidity(cssName, b.getAllowed(), ident)
		} else if typ == CSSPrimitiveValueCssNumber {
			weight := ConversionsGetNumericFontWeight(value.GetFloatValue())
			if weight == nil {
				panic(NewCSSParseException(value.ToString()+" is not a valid font weight", -1))
			}

			replacement := NewPropertyValueString(
				CSSPrimitiveValueCssIdent, weight.ToString(), weight.ToString())
			replacement.SetIdentValue(weight)
			return []*PropertyDeclaration{
				NewPropertyDeclaration(cssName, replacement, important, origin)}

		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *PrimitivePropertyBuildersFontWeight) getAllowed() BitSet {
	return PrimitivePropertyBuildersFontWeights
}

type PrimitivePropertyBuildersFSBorderSpacingHorizontal struct {
	primitivePropertyBuildersLength
}

func NewPrimitivePropertyBuildersFSBorderSpacingHorizontal() *PrimitivePropertyBuildersFSBorderSpacingHorizontal {
	b := &PrimitivePropertyBuildersFSBorderSpacingHorizontal{*newPrimitivePropertyBuildersLength()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSBorderSpacingVertical struct {
	primitivePropertyBuildersLength
}

func NewPrimitivePropertyBuildersFSBorderSpacingVertical() *PrimitivePropertyBuildersFSBorderSpacingVertical {
	b := &PrimitivePropertyBuildersFSBorderSpacingVertical{*newPrimitivePropertyBuildersLength()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSFontMetricSrc struct {
	primitivePropertyBuildersGenericURIWithNone
}

func NewPrimitivePropertyBuildersFSFontMetricSrc() *PrimitivePropertyBuildersFSFontMetricSrc {
	b := &PrimitivePropertyBuildersFSFontMetricSrc{*newPrimitivePropertyBuildersGenericURIWithNone()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSPageHeight struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersFSPageHeight() *PrimitivePropertyBuildersFSPageHeight {
	b := &PrimitivePropertyBuildersFSPageHeight{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSPageWidth struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersFSPageWidth() *PrimitivePropertyBuildersFSPageWidth {
	b := &PrimitivePropertyBuildersFSPageWidth{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

// start | auto
var primitivePropertyBuildersFSPageSequenceAllowed = primitivePropertyBuildersSetFor(IdentValueStart, IdentValueAuto)

type PrimitivePropertyBuildersFSPageSequence struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSPageSequence() *PrimitivePropertyBuildersFSPageSequence {
	b := &PrimitivePropertyBuildersFSPageSequence{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSPageSequenceAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSPageOrientation struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSPageOrientation() *PrimitivePropertyBuildersFSPageOrientation {
	b := &PrimitivePropertyBuildersFSPageOrientation{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersPageOrientations)}
	b.self = b
	return b
}

// auto | embed
var primitivePropertyBuildersFSPDFFontEmbedAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueEmbed)

type PrimitivePropertyBuildersFSPDFFontEmbed struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSPDFFontEmbed() *PrimitivePropertyBuildersFSPDFFontEmbed {
	b := &PrimitivePropertyBuildersFSPDFFontEmbed{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSPDFFontEmbedAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSPDFFontEncoding struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersFSPDFFontEncoding() *PrimitivePropertyBuildersFSPDFFontEncoding {
	b := &PrimitivePropertyBuildersFSPDFFontEncoding{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersFSPDFFontEncoding) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrString(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			// Convert to string
			return []*PropertyDeclaration{
				NewPropertyDeclaration(
					cssName,
					NewPropertyValueString(
						CSSPrimitiveValueCssString,
						value.GetStringValue(),
						value.GetCssText()),
					important,
					origin)}
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

type PrimitivePropertyBuildersFSTableCellColspan struct {
	primitivePropertyBuildersColOrRowSpan
}

func NewPrimitivePropertyBuildersFSTableCellColspan() *PrimitivePropertyBuildersFSTableCellColspan {
	b := &PrimitivePropertyBuildersFSTableCellColspan{*newPrimitivePropertyBuildersColOrRowSpan()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSTableCellRowspan struct {
	primitivePropertyBuildersColOrRowSpan
}

func NewPrimitivePropertyBuildersFSTableCellRowspan() *PrimitivePropertyBuildersFSTableCellRowspan {
	b := &PrimitivePropertyBuildersFSTableCellRowspan{*newPrimitivePropertyBuildersColOrRowSpan()}
	b.self = b
	return b
}

var primitivePropertyBuildersFSTablePaginateAllowed = primitivePropertyBuildersSetFor(IdentValuePaginate, IdentValueAuto)

type PrimitivePropertyBuildersFSTablePaginate struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSTablePaginate() *PrimitivePropertyBuildersFSTablePaginate {
	b := &PrimitivePropertyBuildersFSTablePaginate{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSTablePaginateAllowed)}
	b.self = b
	return b
}

var primitivePropertyBuildersFSTextDecorationExtentAllowed = primitivePropertyBuildersSetFor(IdentValueLine, IdentValueBlock)

type PrimitivePropertyBuildersFSTextDecorationExtent struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSTextDecorationExtent() *PrimitivePropertyBuildersFSTextDecorationExtent {
	b := &PrimitivePropertyBuildersFSTextDecorationExtent{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSTextDecorationExtentAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersFSFitImagesToWidth struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersFSFitImagesToWidth() *PrimitivePropertyBuildersFSFitImagesToWidth {
	b := &PrimitivePropertyBuildersFSFitImagesToWidth{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersHeight struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersHeight() *PrimitivePropertyBuildersHeight {
	b := &PrimitivePropertyBuildersHeight{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

var primitivePropertyBuildersFSDynamicAutoWidthAllowed = primitivePropertyBuildersSetFor(IdentValueDynamic, IdentValueStatic)

type PrimitivePropertyBuildersFSDynamicAutoWidth struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSDynamicAutoWidth() *PrimitivePropertyBuildersFSDynamicAutoWidth {
	b := &PrimitivePropertyBuildersFSDynamicAutoWidth{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSDynamicAutoWidthAllowed)}
	b.self = b
	return b
}

// auto | keep
var primitivePropertyBuildersFSKeepWithInlineAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueKeep)

type PrimitivePropertyBuildersFSKeepWithInline struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSKeepWithInline() *PrimitivePropertyBuildersFSKeepWithInline {
	b := &PrimitivePropertyBuildersFSKeepWithInline{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSKeepWithInlineAllowed)}
	b.self = b
	return b
}

// none | create
var primitivePropertyBuildersFSNamedDestinationAllowed = primitivePropertyBuildersSetFor(IdentValueNone, IdentValueCreate)

type PrimitivePropertyBuildersFSNamedDestination struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersFSNamedDestination() *PrimitivePropertyBuildersFSNamedDestination {
	b := &PrimitivePropertyBuildersFSNamedDestination{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersFSNamedDestinationAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersLeft struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersLeft() *PrimitivePropertyBuildersLeft {
	b := &PrimitivePropertyBuildersLeft{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersLetterSpacing struct {
	primitivePropertyBuildersLengthWithNormal
}

func NewPrimitivePropertyBuildersLetterSpacing() *PrimitivePropertyBuildersLetterSpacing {
	b := &PrimitivePropertyBuildersLetterSpacing{*newPrimitivePropertyBuildersLengthWithNormal()}
	b.self = b
	return b
}

// normal | <number> | <length> | <percentage> | inherit
var primitivePropertyBuildersLineHeightAllowed = primitivePropertyBuildersSetFor(IdentValueNormal)

type PrimitivePropertyBuildersLineHeight struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersLineHeight() *PrimitivePropertyBuildersLineHeight {
	b := &PrimitivePropertyBuildersLineHeight{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersLineHeight) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentLengthNumberOrPercentType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, primitivePropertyBuildersLineHeightAllowed, ident)
		} else if value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException("line-height may not be negative", -1))
		}
	}
	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

type PrimitivePropertyBuildersListStyleImage struct {
	primitivePropertyBuildersGenericURIWithNone
}

func NewPrimitivePropertyBuildersListStyleImage() *PrimitivePropertyBuildersListStyleImage {
	b := &PrimitivePropertyBuildersListStyleImage{*newPrimitivePropertyBuildersGenericURIWithNone()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersListStylePosition struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersListStylePosition() *PrimitivePropertyBuildersListStylePosition {
	b := &PrimitivePropertyBuildersListStylePosition{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersListStylePositions)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersListStyleType struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersListStyleType() *PrimitivePropertyBuildersListStyleType {
	b := &PrimitivePropertyBuildersListStyleType{*newPrimitivePropertyBuildersSingleIdent(PrimitivePropertyBuildersListStyleTypes)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersMarginTop struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersMarginTop() *PrimitivePropertyBuildersMarginTop {
	b := &PrimitivePropertyBuildersMarginTop{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersMarginRight struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersMarginRight() *PrimitivePropertyBuildersMarginRight {
	b := &PrimitivePropertyBuildersMarginRight{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersMarginBottom struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersMarginBottom() *PrimitivePropertyBuildersMarginBottom {
	b := &PrimitivePropertyBuildersMarginBottom{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersMarginLeft struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersMarginLeft() *PrimitivePropertyBuildersMarginLeft {
	b := &PrimitivePropertyBuildersMarginLeft{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersMaxHeight struct {
	primitivePropertyBuildersLengthLikeWithNone
}

func NewPrimitivePropertyBuildersMaxHeight() *PrimitivePropertyBuildersMaxHeight {
	b := &PrimitivePropertyBuildersMaxHeight{*newPrimitivePropertyBuildersLengthLikeWithNone()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersMaxWidth struct {
	primitivePropertyBuildersLengthLikeWithNone
}

func NewPrimitivePropertyBuildersMaxWidth() *PrimitivePropertyBuildersMaxWidth {
	b := &PrimitivePropertyBuildersMaxWidth{*newPrimitivePropertyBuildersLengthLikeWithNone()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersMinHeight struct {
	primitivePropertyBuildersNonNegativeLengthLike
}

func NewPrimitivePropertyBuildersMinHeight() *PrimitivePropertyBuildersMinHeight {
	b := &PrimitivePropertyBuildersMinHeight{*newPrimitivePropertyBuildersNonNegativeLengthLike()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersMinWidth struct {
	primitivePropertyBuildersNonNegativeLengthLike
}

func NewPrimitivePropertyBuildersMinWidth() *PrimitivePropertyBuildersMinWidth {
	b := &PrimitivePropertyBuildersMinWidth{*newPrimitivePropertyBuildersNonNegativeLengthLike()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersOrphans struct {
	primitivePropertyBuildersPlainInteger
}

func NewPrimitivePropertyBuildersOrphans() *PrimitivePropertyBuildersOrphans {
	b := &PrimitivePropertyBuildersOrphans{*newPrimitivePropertyBuildersPlainInteger()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersOpacity struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersOpacity() *PrimitivePropertyBuildersOpacity {
	b := &PrimitivePropertyBuildersOpacity{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersOpacity) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	b.CheckNumberType(cssName, value)
	AbstractPropertyBuilderCheckValueBetween(cssName, value.GetFloatValue(), 0, 1)

	return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
}

// We only support visible or hidden for now
// visible | hidden | scroll | auto | inherit
// (scroll and auto are not in the allowed set)
var primitivePropertyBuildersOverflowAllowed = primitivePropertyBuildersSetFor(IdentValueVisible, IdentValueHidden)

type PrimitivePropertyBuildersOverflow struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersOverflow() *PrimitivePropertyBuildersOverflow {
	b := &PrimitivePropertyBuildersOverflow{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersOverflowAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersPaddingTop struct {
	primitivePropertyBuildersNonNegativeLengthLike
}

func NewPrimitivePropertyBuildersPaddingTop() *PrimitivePropertyBuildersPaddingTop {
	b := &PrimitivePropertyBuildersPaddingTop{*newPrimitivePropertyBuildersNonNegativeLengthLike()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersPaddingRight struct {
	primitivePropertyBuildersNonNegativeLengthLike
}

func NewPrimitivePropertyBuildersPaddingRight() *PrimitivePropertyBuildersPaddingRight {
	b := &PrimitivePropertyBuildersPaddingRight{*newPrimitivePropertyBuildersNonNegativeLengthLike()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersPaddingBottom struct {
	primitivePropertyBuildersNonNegativeLengthLike
}

func NewPrimitivePropertyBuildersPaddingBottom() *PrimitivePropertyBuildersPaddingBottom {
	b := &PrimitivePropertyBuildersPaddingBottom{*newPrimitivePropertyBuildersNonNegativeLengthLike()}
	b.self = b
	return b
}

type PrimitivePropertyBuildersPaddingLeft struct {
	primitivePropertyBuildersNonNegativeLengthLike
}

func NewPrimitivePropertyBuildersPaddingLeft() *PrimitivePropertyBuildersPaddingLeft {
	b := &PrimitivePropertyBuildersPaddingLeft{*newPrimitivePropertyBuildersNonNegativeLengthLike()}
	b.self = b
	return b
}

// auto | always | avoid | left | right | inherit
var primitivePropertyBuildersPageBreakBeforeAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueAlways, IdentValueAvoid, IdentValueLeft, IdentValueRight)

type PrimitivePropertyBuildersPageBreakBefore struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersPageBreakBefore() *PrimitivePropertyBuildersPageBreakBefore {
	b := &PrimitivePropertyBuildersPageBreakBefore{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersPageBreakBeforeAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersPage struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersPage() *PrimitivePropertyBuildersPage {
	b := &PrimitivePropertyBuildersPage{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersPage) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentType(cssName, value)

		if value.GetStringValue() != "auto" {
			// Treat as string since it won't be a proper IdentValue
			value = NewPropertyValueString(
				CSSPrimitiveValueCssString, value.GetStringValue(), value.GetCssText())
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

// auto | always | avoid | left | right | inherit
var primitivePropertyBuildersPageBreakAfterAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueAlways, IdentValueAvoid, IdentValueLeft, IdentValueRight)

type PrimitivePropertyBuildersPageBreakAfter struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersPageBreakAfter() *PrimitivePropertyBuildersPageBreakAfter {
	b := &PrimitivePropertyBuildersPageBreakAfter{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersPageBreakAfterAllowed)}
	b.self = b
	return b
}

// avoid | auto | inherit
var primitivePropertyBuildersPageBreakInsideAllowed = primitivePropertyBuildersSetFor(IdentValueAvoid, IdentValueAuto)

type PrimitivePropertyBuildersPageBreakInside struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersPageBreakInside() *PrimitivePropertyBuildersPageBreakInside {
	b := &PrimitivePropertyBuildersPageBreakInside{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersPageBreakInsideAllowed)}
	b.self = b
	return b
}

// static | relative | absolute | fixed | inherit
var primitivePropertyBuildersPositionAllowed = primitivePropertyBuildersSetFor(IdentValueStatic, IdentValueRelative, IdentValueAbsolute, IdentValueFixed)

type PrimitivePropertyBuildersPosition struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersPosition() *PrimitivePropertyBuildersPosition {
	b := &PrimitivePropertyBuildersPosition{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersPosition) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			b.CheckIdentType(cssName, value)
			ident := b.CheckIdent(value)

			b.CheckValidity(cssName, b.getAllowed(), ident)
		} else if value.GetPropertyValueType() == PropertyValueTypeValueTypeFunction {
			function := value.GetFunction()
			if function.Is("running") {
				params := function.GetParameters()
				if len(params) == 1 {
					param := params[0]
					if param.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
						panic(NewCSSParseException("The running function takes an identifier as a parameter", -1))
					}
				} else {
					panic(NewCSSParseException("The running function takes one parameter", -1))
				}
			} else {
				panic(NewCSSParseException("Only the running function is supported here", -1))
			}
		} else {
			panic(NewCSSParseException("Value for "+cssName.ToString()+" must be an identifier or function", -1))
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func (b *PrimitivePropertyBuildersPosition) getAllowed() BitSet {
	return primitivePropertyBuildersPositionAllowed
}

type PrimitivePropertyBuildersRight struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersRight() *PrimitivePropertyBuildersRight {
	b := &PrimitivePropertyBuildersRight{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

// <uri> | none | inherit
// Also supports: url('font.woff') format('woff'), url('font.ttf') format('truetype')
var primitivePropertyBuildersSrcAllowed = primitivePropertyBuildersSetFor(IdentValueNone)

type PrimitivePropertyBuildersSrc struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersSrc() *PrimitivePropertyBuildersSrc {
	b := &PrimitivePropertyBuildersSrc{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersSrc) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	// Handle inherit case
	if len(values) == 1 {
		value := values[0]
		b.CheckInheritAllowed(value, inheritAllowed)
		if value.GetCssValueType() == CSSValueCssInherit {
			return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
		}

		// Handle single value case (none or single URL)
		b.CheckIdentOrURIType(cssName, value)
		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, primitivePropertyBuildersSrcAllowed, ident)
		}
		return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
	}

	// Handle multiple values (e.g., url() format() pairs)
	// Wrap all values into a PropertyValue list
	listValue := NewPropertyValueList(propertyBuilderValueList(values))
	return []*PropertyDeclaration{NewPropertyDeclaration(cssName, listValue, important, origin)}
}

type PrimitivePropertyBuildersTabSize struct {
	primitivePropertyBuildersPlainInteger
}

func NewPrimitivePropertyBuildersTabSize() *PrimitivePropertyBuildersTabSize {
	b := &PrimitivePropertyBuildersTabSize{*newPrimitivePropertyBuildersPlainInteger()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersTop struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersTop() *PrimitivePropertyBuildersTop {
	b := &PrimitivePropertyBuildersTop{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

// auto | fixed | inherit
var primitivePropertyBuildersTableLayoutAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueFixed)

type PrimitivePropertyBuildersTableLayout struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersTableLayout() *PrimitivePropertyBuildersTableLayout {
	b := &PrimitivePropertyBuildersTableLayout{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersTableLayoutAllowed)}
	b.self = b
	return b
}

// left | right | center | justify | inherit
var primitivePropertyBuildersTextAlignAllowed = primitivePropertyBuildersSetFor(IdentValueLeft, IdentValueRight, IdentValueCenter, IdentValueJustify)

type PrimitivePropertyBuildersTextAlign struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersTextAlign() *PrimitivePropertyBuildersTextAlign {
	b := &PrimitivePropertyBuildersTextAlign{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersTextAlignAllowed)}
	b.self = b
	return b
}

// none | [ underline || overline || line-through || blink ] | inherit
// (none and blink are not in the allowed set)
var primitivePropertyBuildersTextDecorationAllowed = primitivePropertyBuildersSetFor(
	IdentValueUnderline,
	IdentValueOverline, IdentValueLineThrough)

type PrimitivePropertyBuildersTextDecoration struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersTextDecoration() *PrimitivePropertyBuildersTextDecoration {
	b := &PrimitivePropertyBuildersTextDecoration{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersTextDecoration) getAllowed() BitSet {
	return primitivePropertyBuildersTextDecorationAllowed
}

func (b *PrimitivePropertyBuildersTextDecoration) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]
		goWithSingle := false
		if value.GetCssValueType() == CSSValueCssInherit {
			goWithSingle = true
		} else {
			b.CheckIdentType(CSSNameTextDecoration, value)
			ident := b.CheckIdent(value)
			if ident == IdentValueNone {
				goWithSingle = true
			}
		}

		if goWithSingle {
			return []*PropertyDeclaration{
				NewPropertyDeclaration(cssName, value, important, origin)}
		}
	}

	for _, value := range values {
		b.CheckInheritAllowed(value, false)
		b.CheckIdentType(cssName, value)
		ident := b.CheckIdent(value)
		if ident == IdentValueNone {
			panic(NewCSSParseException("Value none may not be used in this position", -1))
		}
		b.CheckValidity(cssName, b.getAllowed(), ident)
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, NewPropertyValueList(propertyBuilderValueList(values)), important, origin)}
}

type PrimitivePropertyBuildersTextIndent struct {
	primitivePropertyBuildersLengthLike
}

func NewPrimitivePropertyBuildersTextIndent() *PrimitivePropertyBuildersTextIndent {
	b := &PrimitivePropertyBuildersTextIndent{*newPrimitivePropertyBuildersLengthLike()}
	b.self = b
	return b
}

// auto | under
var primitivePropertyBuildersTextUnderlinePositionAllowed = primitivePropertyBuildersSetFor(IdentValueAuto, IdentValueUnder)

type PrimitivePropertyBuildersTextUnderlinePosition struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersTextUnderlinePosition() *PrimitivePropertyBuildersTextUnderlinePosition {
	b := &PrimitivePropertyBuildersTextUnderlinePosition{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersTextUnderlinePositionAllowed)}
	b.self = b
	return b
}

// auto | <length> | <percentage>
type PrimitivePropertyBuildersTextUnderlineOffset struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersTextUnderlineOffset() *PrimitivePropertyBuildersTextUnderlineOffset {
	b := &PrimitivePropertyBuildersTextUnderlineOffset{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.self = b
	return b
}

// capitalize | uppercase | lowercase | none | inherit
var primitivePropertyBuildersTextTransformAllowed = primitivePropertyBuildersSetFor(IdentValueCapitalize, IdentValueUppercase, IdentValueLowercase, IdentValueNone)

type PrimitivePropertyBuildersTextTransform struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersTextTransform() *PrimitivePropertyBuildersTextTransform {
	b := &PrimitivePropertyBuildersTextTransform{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersTextTransformAllowed)}
	b.self = b
	return b
}

// baseline | sub | super | top | text-top | middle
// | bottom | text-bottom | <percentage> | <length> | inherit
var primitivePropertyBuildersVerticalAlignAllowed = primitivePropertyBuildersSetFor(IdentValueBaseline, IdentValueSub, IdentValueSuper, IdentValueTop, IdentValueTextTop, IdentValueMiddle, IdentValueBottom, IdentValueTextBottom)

type PrimitivePropertyBuildersVerticalAlign struct {
	primitivePropertyBuildersLengthLikeWithIdent
}

func NewPrimitivePropertyBuildersVerticalAlign() *PrimitivePropertyBuildersVerticalAlign {
	b := &PrimitivePropertyBuildersVerticalAlign{*newPrimitivePropertyBuildersLengthLikeWithIdent(primitivePropertyBuildersVerticalAlignAllowed)}
	b.self = b
	return b
}

// visible | hidden | collapse | inherit
var primitivePropertyBuildersVisibilityAllowed = primitivePropertyBuildersSetFor(IdentValueVisible, IdentValueHidden, IdentValueCollapse)

type PrimitivePropertyBuildersVisibility struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersVisibility() *PrimitivePropertyBuildersVisibility {
	b := &PrimitivePropertyBuildersVisibility{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersVisibilityAllowed)}
	b.self = b
	return b
}

// normal | pre | nowrap | pre-wrap | pre-line | inherit
var primitivePropertyBuildersWhiteSpaceAllowed = primitivePropertyBuildersSetFor(IdentValueNormal, IdentValuePre, IdentValueNowrap, IdentValuePreWrap, IdentValuePreLine)

type PrimitivePropertyBuildersWhiteSpace struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersWhiteSpace() *PrimitivePropertyBuildersWhiteSpace {
	b := &PrimitivePropertyBuildersWhiteSpace{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersWhiteSpaceAllowed)}
	b.self = b
	return b
}

// normal | break-all
var primitivePropertyBuildersWordBreakAllowed = primitivePropertyBuildersSetFor(IdentValueNormal, IdentValueBreakAll)

type PrimitivePropertyBuildersWordBreak struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersWordBreak() *PrimitivePropertyBuildersWordBreak {
	b := &PrimitivePropertyBuildersWordBreak{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersWordBreakAllowed)}
	b.self = b
	return b
}

// normal | break-word
var primitivePropertyBuildersWordWrapAllowed = primitivePropertyBuildersSetFor(IdentValueNormal, IdentValueBreakWord)

type PrimitivePropertyBuildersWordWrap struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersWordWrap() *PrimitivePropertyBuildersWordWrap {
	b := &PrimitivePropertyBuildersWordWrap{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersWordWrapAllowed)}
	b.self = b
	return b
}

// none | manual | auto
var primitivePropertyBuildersHyphensAllowed = primitivePropertyBuildersSetFor(IdentValueNone, IdentValueManual, IdentValueAuto)

type PrimitivePropertyBuildersHyphens struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersHyphens() *PrimitivePropertyBuildersHyphens {
	b := &PrimitivePropertyBuildersHyphens{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersHyphensAllowed)}
	b.self = b
	return b
}

// border-box | content-box
var primitivePropertyBuildersBoxSizingAllowed = primitivePropertyBuildersSetFor(IdentValueBorderBox, IdentValueContentBox)

type PrimitivePropertyBuildersBoxSizing struct {
	primitivePropertyBuildersSingleIdent
}

func NewPrimitivePropertyBuildersBoxSizing() *PrimitivePropertyBuildersBoxSizing {
	b := &PrimitivePropertyBuildersBoxSizing{*newPrimitivePropertyBuildersSingleIdent(primitivePropertyBuildersBoxSizingAllowed)}
	b.self = b
	return b
}

type PrimitivePropertyBuildersWidows struct {
	primitivePropertyBuildersPlainInteger
}

func NewPrimitivePropertyBuildersWidows() *PrimitivePropertyBuildersWidows {
	b := &PrimitivePropertyBuildersWidows{*newPrimitivePropertyBuildersPlainInteger()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersWidth struct {
	primitivePropertyBuildersLengthLikeWithAuto
}

func NewPrimitivePropertyBuildersWidth() *PrimitivePropertyBuildersWidth {
	b := &PrimitivePropertyBuildersWidth{*newPrimitivePropertyBuildersLengthLikeWithAuto()}
	b.negativeValuesAllowed = false
	b.self = b
	return b
}

type PrimitivePropertyBuildersWordSpacing struct {
	primitivePropertyBuildersLengthWithNormal
}

func NewPrimitivePropertyBuildersWordSpacing() *PrimitivePropertyBuildersWordSpacing {
	b := &PrimitivePropertyBuildersWordSpacing{*newPrimitivePropertyBuildersLengthWithNormal()}
	b.self = b
	return b
}

// auto | <integer> | inherit
var primitivePropertyBuildersZIndexAllowed = primitivePropertyBuildersSetFor(IdentValueAuto)

type PrimitivePropertyBuildersZIndex struct {
	AbstractPropertyBuilder
}

func NewPrimitivePropertyBuildersZIndex() *PrimitivePropertyBuildersZIndex {
	b := &PrimitivePropertyBuildersZIndex{}
	b.self = b
	return b
}

func (b *PrimitivePropertyBuildersZIndex) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	primitivePropertyBuildersAssertFoundSingleValue(cssName, values)
	value := values[0]
	b.CheckInheritAllowed(value, inheritAllowed)
	if value.GetCssValueType() != CSSValueCssInherit {
		b.CheckIdentOrIntegerType(cssName, value)

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			b.CheckValidity(cssName, primitivePropertyBuildersZIndexAllowed, ident)
		}
	}

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName, value, important, origin)}
}

func primitivePropertyBuildersCreateTwoValueResponse(
	cssName *CSSName, value1 *PropertyValue,
	value2 *PropertyValue,
	origin StylesheetInfoOrigin, important bool) []*PropertyDeclaration {

	return []*PropertyDeclaration{
		NewPropertyDeclaration(cssName,
			NewPropertyValueList(propertyBuilderValueList([]*PropertyValue{value1, value2})), important, origin),
	}
}
