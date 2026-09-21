// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/constants/ValueConstants.java

package ufo

import "errors"

// ValueConstants is a utility class for working with CSSValue instances.

var valueConstantsSacTypesStrings = map[int16]string{
	// HACK: this is a quick way to perform the lookup, but dumb if the short assigned are > 100; but the compiler will tell us that (PWW 21-01-05)
	CSSPrimitiveValueCssEms:        "em",
	CSSPrimitiveValueCssExs:        "ex",
	CSSPrimitiveValueCssPx:         "px",
	CSSPrimitiveValueCssPercentage: "%",
	CSSPrimitiveValueCssIn:         "in",
	CSSPrimitiveValueCssCm:         "cm",
	CSSPrimitiveValueCssMm:         "mm",
	CSSPrimitiveValueCssPt:         "pt",
	CSSPrimitiveValueCssPc:         "pc",
}

// ValueConstantsStringForSACPrimitiveType returns "" for a type that has no
// unit string (Java returns null).
func ValueConstantsStringForSACPrimitiveType(primitiveType int16) string {
	return valueConstantsSacTypesStrings[primitiveType]
}

// ValueConstantsIsAbsoluteUnit returns true if the specified type absolute
// (even if we have a computed value for it), meaning that either the value
// can be used directly (e.g. pixels) or there is a fixed context-independent
// conversion for it (e.g. inches). Proportional types (e.g. %) return false.
//
// primitiveType is the CSSValue type to check.
func ValueConstantsIsAbsoluteUnit(primitiveType int16) bool {
	// TODO: check this list...

	// note, all types are included here to make sure none are missed
	switch primitiveType {

	// proportional length or size
	case CSSPrimitiveValueCssPercentage:
		return false

	// refer to values known to the DerivedValue instance (tobe)
	case CSSPrimitiveValueCssEms,
		CSSPrimitiveValueCssExs:
		return true

	// length
	case CSSPrimitiveValueCssIn,
		CSSPrimitiveValueCssCm,
		CSSPrimitiveValueCssMm,
		CSSPrimitiveValueCssPt,
		CSSPrimitiveValueCssPc,
		CSSPrimitiveValueCssPx:
		return true

	// color
	case CSSPrimitiveValueCssRgbcolor:
		return true

	// ?
	case CSSPrimitiveValueCssAttr,
		CSSPrimitiveValueCssDimension,
		CSSPrimitiveValueCssNumber,
		CSSPrimitiveValueCssRect:
		return true

	// counters
	case CSSPrimitiveValueCssCounter:
		return true

	// angles
	case CSSPrimitiveValueCssDeg,
		CSSPrimitiveValueCssGrad,
		CSSPrimitiveValueCssRad:
		return true

	// aural - freq
	case CSSPrimitiveValueCssHz,
		CSSPrimitiveValueCssKhz:
		return true

	// time
	case CSSPrimitiveValueCssS,
		CSSPrimitiveValueCssMs:
		return true

	// URI
	case CSSPrimitiveValueCssUri,

		CSSPrimitiveValueCssIdent,
		CSSPrimitiveValueCssString:
		return true

	case CSSPrimitiveValueCssUnknown:
		XRLogCascade(LevelWarning, "Asked whether type was absolute, given CSS_UNKNOWN as the type. "+
			"Might be one of those funny values like background-position.")
		GeneralUtilDumpShortException(errors.New("Taking a thread dump..."))
		return false

	default:
		return false
	}
}

// ValueConstantsIsNumber returns true if the SAC primitive value type is a
// number unit--a unit that can only contain a numeric value. This is a
// shorthand way of saying, did the user declare this as a number unit (like
// px)?
func ValueConstantsIsNumber(cssPrimitiveType int16) bool {
	switch cssPrimitiveType {
	// relative length or size
	case CSSPrimitiveValueCssEms, CSSPrimitiveValueCssExs, CSSPrimitiveValueCssPercentage:
		// relatives will be treated separately from lengths;
		return false
	// length
	case CSSPrimitiveValueCssPx, CSSPrimitiveValueCssIn, CSSPrimitiveValueCssCm, CSSPrimitiveValueCssMm, CSSPrimitiveValueCssPt, CSSPrimitiveValueCssPc:
		return true
	default:
		return false
	}
}
