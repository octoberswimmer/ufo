// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/constants/IdentValue.java

package ufo

import "sync"

// identValueAllIdentValues is a ConcurrentHashMap in Java. It is written only
// while the package-level variables below are initialized and read after
// that; the mutex covers the reads and writes all the same.
var (
	identValueAllIdentValues     = map[string]*IdentValue{}
	identValueAllIdentValuesLock sync.RWMutex
	identValueMaxAssigned        int
)

type IdentValue struct {
	ident string
	FS_ID int
}

// The variables are initialized in declaration order (none of the
// initializers depends on another package-level variable that is declared
// later), so every FS_ID is the same as in Java.
var (
	IdentValueAbsolute             = identValueAddValue("absolute")
	IdentValueAlways               = identValueAddValue("always")
	IdentValueArmenian             = identValueAddValue("armenian")
	IdentValueAuto                 = identValueAddValue("auto")
	IdentValueAvoid                = identValueAddValue("avoid")
	IdentValueBaseline             = identValueAddValue("baseline")
	IdentValueBlink                = identValueAddValue("blink")
	IdentValueBlock                = identValueAddValue("block")
	IdentValueBold                 = identValueAddValue("bold")
	IdentValueBolder               = identValueAddValue("bolder")
	IdentValueBorderBox            = identValueAddValue("border-box")
	IdentValueBoth                 = identValueAddValue("both")
	IdentValueBottom               = identValueAddValue("bottom")
	IdentValueBreakAll             = identValueAddValue("break-all")
	IdentValueCapitalize           = identValueAddValue("capitalize")
	IdentValueCenter               = identValueAddValue("center")
	IdentValueCircle               = identValueAddValue("circle")
	IdentValueCjkIdeographic       = identValueAddValue("cjk-ideographic")
	IdentValueCloseQuote           = identValueAddValue("close-quote")
	IdentValueCollapse             = identValueAddValue("collapse")
	IdentValueCompact              = identValueAddValue("compact")
	IdentValueContain              = identValueAddValue("contain")
	IdentValueContentBox           = identValueAddValue("content-box")
	IdentValueCover                = identValueAddValue("cover")
	IdentValueCreate               = identValueAddValue("create")
	IdentValueDashed               = identValueAddValue("dashed")
	IdentValueDecimal              = identValueAddValue("decimal")
	IdentValueDecimalLeadingZero   = identValueAddValue("decimal-leading-zero")
	IdentValueDisc                 = identValueAddValue("disc")
	IdentValueDotted               = identValueAddValue("dotted")
	IdentValueDouble               = identValueAddValue("double")
	IdentValueDynamic              = identValueAddValue("dynamic")
	IdentValueFixed                = identValueAddValue("fixed")
	IdentValueFontWeight100        = identValueAddValue("100")
	IdentValueFontWeight200        = identValueAddValue("200")
	IdentValueFontWeight300        = identValueAddValue("300")
	IdentValueFontWeight400        = identValueAddValue("400")
	IdentValueFontWeight500        = identValueAddValue("500")
	IdentValueFontWeight600        = identValueAddValue("600")
	IdentValueFontWeight700        = identValueAddValue("700")
	IdentValueFontWeight800        = identValueAddValue("800")
	IdentValueFontWeight900        = identValueAddValue("900")
	IdentValueFsContentPlaceholder = identValueAddValue("-fs-content-placeholder")
	IdentValueFsInitialValue       = identValueAddValue("-fs-initial-value")
	IdentValueGeorgian             = identValueAddValue("georgian")
	IdentValueGroove               = identValueAddValue("groove")
	IdentValueHebrew               = identValueAddValue("hebrew")
	IdentValueHidden               = identValueAddValue("hidden")
	IdentValueHide                 = identValueAddValue("hide")
	IdentValueHiragana             = identValueAddValue("hiragana")
	IdentValueHiraganaIroha        = identValueAddValue("hiragana-iroha")
	IdentValueInherit              = identValueAddValue("inherit")
	IdentValueInline               = identValueAddValue("inline")
	IdentValueInlineBlock          = identValueAddValue("inline-block")
	IdentValueInlineTable          = identValueAddValue("inline-table")
	IdentValueInset                = identValueAddValue("inset")
	IdentValueInside               = identValueAddValue("inside")
	IdentValueItalic               = identValueAddValue("italic")
	IdentValueJustify              = identValueAddValue("justify")
	IdentValueKatakana             = identValueAddValue("katakana")
	IdentValueKatakanaIroha        = identValueAddValue("katakana-iroha")
	IdentValueKeep                 = identValueAddValue("keep")
	IdentValueLandscape            = identValueAddValue("landscape")
	IdentValueLeft                 = identValueAddValue("left")
	IdentValueLighter              = identValueAddValue("lighter")
	IdentValueLine                 = identValueAddValue("line")
	IdentValueLinearGradient       = identValueAddValue("linear-gradient")
	IdentValueLineThrough          = identValueAddValue("line-through")
	IdentValueListItem             = identValueAddValue("list-item")
	IdentValueLowerAlpha           = identValueAddValue("lower-alpha")
	IdentValueLowerGreek           = identValueAddValue("lower-greek")
	IdentValueLowerLatin           = identValueAddValue("lower-latin")
	IdentValueLowerRoman           = identValueAddValue("lower-roman")
	IdentValueLowercase            = identValueAddValue("lowercase")
	IdentValueLtr                  = identValueAddValue("ltr")
	IdentValueMarker               = identValueAddValue("marker")
	IdentValueMiddle               = identValueAddValue("middle")
	IdentValueNoCloseQuote         = identValueAddValue("no-close-quote")
	IdentValueNoOpenQuote          = identValueAddValue("no-open-quote")
	IdentValueNoRepeat             = identValueAddValue("no-repeat")
	IdentValueNone                 = identValueAddValue("none")
	IdentValueNormal               = identValueAddValue("normal")
	IdentValueNowrap               = identValueAddValue("nowrap")
	IdentValueBreakWord            = identValueAddValue("break-word")
	IdentValueOblique              = identValueAddValue("oblique")
	IdentValueOpenQuote            = identValueAddValue("open-quote")
	IdentValueOutset               = identValueAddValue("outset")
	IdentValueOutside              = identValueAddValue("outside")
	IdentValueOverline             = identValueAddValue("overline")
	IdentValuePaginate             = identValueAddValue("paginate")
	IdentValuePointer              = identValueAddValue("pointer")
	IdentValuePortrait             = identValueAddValue("portrait")
	IdentValuePre                  = identValueAddValue("pre")
	IdentValuePreLine              = identValueAddValue("pre-line")
	IdentValuePreWrap              = identValueAddValue("pre-wrap")
	IdentValueRelative             = identValueAddValue("relative")
	IdentValueRepeat               = identValueAddValue("repeat")
	IdentValueRepeatX              = identValueAddValue("repeat-x")
	IdentValueRepeatY              = identValueAddValue("repeat-y")
	IdentValueRidge                = identValueAddValue("ridge")
	IdentValueRight                = identValueAddValue("right")
	IdentValueRunIn                = identValueAddValue("run-in")
	IdentValueScroll               = identValueAddValue("scroll")
	IdentValueSeparate             = identValueAddValue("separate")
	IdentValueShow                 = identValueAddValue("show")
	IdentValueSmallCaps            = identValueAddValue("small-caps")
	IdentValueSolid                = identValueAddValue("solid")
	IdentValueSquare               = identValueAddValue("square")
	IdentValueStatic               = identValueAddValue("static")
	IdentValueSub                  = identValueAddValue("sub")
	IdentValueSuper                = identValueAddValue("super")
	IdentValueTable                = identValueAddValue("table")
	IdentValueTableCaption         = identValueAddValue("table-caption")
	IdentValueTableCell            = identValueAddValue("table-cell")
	IdentValueTableColumn          = identValueAddValue("table-column")
	IdentValueTableColumnGroup     = identValueAddValue("table-column-group")
	IdentValueTableFooterGroup     = identValueAddValue("table-footer-group")
	IdentValueTableHeaderGroup     = identValueAddValue("table-header-group")
	IdentValueTableRow             = identValueAddValue("table-row")
	IdentValueTableRowGroup        = identValueAddValue("table-row-group")
	IdentValueTextBottom           = identValueAddValue("text-bottom")
	IdentValueTextTop              = identValueAddValue("text-top")
	IdentValueThick                = identValueAddValue("thick")
	IdentValueThin                 = identValueAddValue("thin")
	IdentValueTop                  = identValueAddValue("top")
	IdentValueTransparent          = identValueAddValue("transparent")
	IdentValueUnder                = identValueAddValue("under")
	IdentValueUnderline            = identValueAddValue("underline")
	IdentValueUpperAlpha           = identValueAddValue("upper-alpha")
	IdentValueUpperGreek           = identValueAddValue("upper-greek")
	IdentValueUpperLatin           = identValueAddValue("upper-latin")
	IdentValueUpperRoman           = identValueAddValue("upper-roman")
	IdentValueUppercase            = identValueAddValue("uppercase")
	IdentValueVisible              = identValueAddValue("visible")
	IdentValueCrosshair            = identValueAddValue("crosshair")
	IdentValueDefault              = identValueAddValue("default")
	IdentValueEmbed                = identValueAddValue("embed")
	IdentValueEResize              = identValueAddValue("e-resize")
	IdentValueHelp                 = identValueAddValue("help")
	IdentValueLarge                = identValueAddValue("large")
	IdentValueLarger               = identValueAddValue("larger")
	IdentValueMedium               = identValueAddValue("medium")
	IdentValueMove                 = identValueAddValue("move")
	IdentValueNResize              = identValueAddValue("n-resize")
	IdentValueNeResize             = identValueAddValue("ne-resize")
	IdentValueNwResize             = identValueAddValue("nw-resize")
	IdentValueProgress             = identValueAddValue("progress")
	IdentValueSResize              = identValueAddValue("s-resize")
	IdentValueSeResize             = identValueAddValue("se-resize")
	IdentValueSmall                = identValueAddValue("small")
	IdentValueSmaller              = identValueAddValue("smaller")
	IdentValueStart                = identValueAddValue("start")
	IdentValueSwResize             = identValueAddValue("sw-resize")
	IdentValueText                 = identValueAddValue("text")
	IdentValueWResize              = identValueAddValue("w-resize")
	IdentValueWait                 = identValueAddValue("wait")
	IdentValueXLarge               = identValueAddValue("x-large")
	IdentValueXSmall               = identValueAddValue("x-small")
	IdentValueXxLarge              = identValueAddValue("xx-large")
	IdentValueXxSmall              = identValueAddValue("xx-small")
	IdentValueManual               = identValueAddValue("manual")
)

func newIdentValue(ident string) *IdentValue {
	v := &IdentValue{ident: ident, FS_ID: identValueMaxAssigned}
	identValueMaxAssigned++
	return v
}

// ToString returns a string representation of the object, in this case, the
// ident as a string (as it appears in the CSS spec).
func (v *IdentValue) ToString() string {
	return v.ident
}

func (v *IdentValue) String() string {
	return v.ident
}

// IdentValueGetByIdentString returns the Singleton IdentValue that
// corresponds to the given string, e.g. for "normal" will return
// IdentValueNormal. Use this when you have the string but need to look up the
// Singleton. If the string doesn't match an ident in the CSS spec, it panics
// with an XRRuntimeException.
//
// ident is the identifier to retrieve the Singleton IdentValue for.
func IdentValueGetByIdentString(ident string) *IdentValue {
	val := IdentValueValueOf(ident)
	if val == nil {
		panic(NewXRRuntimeException("Ident named " + ident + " has no IdentValue instance assigned to it."))
	}
	return val
}

func IdentValueLooksLikeIdent(ident string) bool {
	return IdentValueValueOf(ident) != nil
}

// IdentValueValueOf may return nil.
func IdentValueValueOf(ident string) *IdentValue {
	identValueAllIdentValuesLock.RLock()
	defer identValueAllIdentValuesLock.RUnlock()
	return identValueAllIdentValues[ident]
}

func IdentValueGetIdentCount() int {
	identValueAllIdentValuesLock.RLock()
	defer identValueAllIdentValuesLock.RUnlock()
	return len(identValueAllIdentValues)
}

// identValueAddValue adds a feature to the Value attribute of the IdentValue
// class.
//
// ident is the feature to be added to the Value attribute.
func identValueAddValue(ident string) *IdentValue {
	val := newIdentValue(ident)
	identValueAllIdentValuesLock.Lock()
	defer identValueAllIdentValuesLock.Unlock()
	identValueAllIdentValues[ident] = val
	return val
}

/*
 * METHODS USED TO SUPPORT IdentValue as an FSDerivedValue, used in CalculatedStyle.
 * Most of these panic--makes use of the interface easier in CS (avoids casting)
 */

func (v *IdentValue) IsDeclaredInherit() bool {
	return v == IdentValueInherit
}

func (v *IdentValue) ComputedValue() FSDerivedValue {
	return v
}

func (v *IdentValue) AsFloat() float32 {
	panic(NewXRRuntimeException("Ident value is never a float; wrong class used for derived value."))
}

func (v *IdentValue) AsColor() FSColor {
	panic(NewXRRuntimeException("Ident value is never a color; wrong class used for derived value."))
}

func (v *IdentValue) GetFloatProportionalTo(cssName *CSSName,
	baseValue float32,
	ctx CssContext) float32 {
	panic(NewXRRuntimeException("Ident value (" + v.ToString() + ") is never a length; wrong class used for derived value."))
}

func (v *IdentValue) AsString() string {
	return v.ToString()
}

func (v *IdentValue) AsStringArray() []string {
	panic(NewXRRuntimeException("Ident value is never a string array; wrong class used for derived value."))
}

func (v *IdentValue) AsIdentValue() *IdentValue {
	return v
}

func (v *IdentValue) HasAbsoluteUnit() bool {
	// log and return false
	panic(NewXRRuntimeException("Ident value is never an absolute unit; wrong class used for derived value; this " +
		"ident value is a " + v.AsString()))
}

func (v *IdentValue) IsIdent() bool {
	return true
}

func (v *IdentValue) IsDependentOnFontSize() bool {
	return false
}
