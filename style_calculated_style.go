// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/CalculatedStyle.java

package ufo

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// CalculatedStyleI lists the non-private methods of CalculatedStyle, which
// EmptyStyle extends.
type CalculatedStyleI interface {
	AsCalculatedStyle() *CalculatedStyle

	DeriveStyle(matched *CascadedStyle) CalculatedStyleI
	GetParent() CalculatedStyleI
	ToString() string
	AsColor(cssName *CSSName) FSColor
	AsFloat(cssName *CSSName) float32
	AsString(cssName *CSSName) string
	AsStringArray(cssName *CSSName) []string
	SetDefaultValue(cssName *CSSName, fsDerivedValue FSDerivedValue)
	HasAbsoluteUnit(cssName *CSSName) bool
	IsIdent(cssName *CSSName, val *IdentValue) bool
	GetIdent(cssName *CSSName) *IdentValue
	GetOpacity() float32
	GetDisplay() *IdentValue
	GetColor() FSColor
	GetBackgroundColor() FSColor
	GetBackgroundSize() *BackgroundSize
	GetBackgroundPosition() *BackgroundPosition
	HasTransform() bool
	GetTransforms() []*PropertyValue
	GetTransformOrigin() *BackgroundPosition
	GetCounterReset() []*CounterData
	GetCounterIncrement() []*CounterData
	GetBorder(ctx CssContext) *BorderPropertySet
	DisableOSBorder() bool
	GetFont(ctx CssContext) *FontSpecification
	GetFontSpecification() *FontSpecification
	GetIntPropertyProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) int
	GetFloatPropertyProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) float32
	GetFloatPropertyProportionalWidth(cssName *CSSName, parentWidth float32, ctx CssContext) float32
	GetFloatPropertyProportionalHeight(cssName *CSSName, parentHeight float32, ctx CssContext) float32
	GetLineHeight(ctx CssContext) float32
	LetterSpacing(ctx CssContext) float32
	GetMarginRect(cbWidth float32, ctx CssContext) RectPropertySetI
	GetMarginRectWithUseCache(cbWidth float32, ctx CssContext, useCache bool) RectPropertySetI
	GetPaddingRect(cbWidth float32, ctx CssContext) RectPropertySetI
	GetStringProperty(cssName *CSSName) string
	IsLength(cssName *CSSName) bool
	IsLengthOrNumber(cssName *CSSName) bool
	ValueByName(cssName *CSSName) FSDerivedValue
	GetCachedPadding() RectPropertySetI
	GetMarginBorderPadding(cssCtx CssContext, cbWidth int, edge *CalculatedStyleEdge) int
	Padding() *NullableInsets
	GetWhitespace() *IdentValue
	GetFSFont(cssContext CssContext) FSFont
	GetFSFontMetrics(c CssContext) FSFontMetrics
	GetWordWrap() *IdentValue
	GetWordBreak() *IdentValue
	GetHyphens() *IdentValue
	IsClearLeft() bool
	IsClearRight() bool
	IsCleared() bool
	GetBackgroundRepeat() *IdentValue
	GetBackgroundAttachment() *IdentValue
	IsFixedBackground() bool
	IsInline() bool
	IsInlineBlock() bool
	IsTable() bool
	IsInlineTable() bool
	IsUnderTableLayout() bool
	IsTableCell() bool
	IsTableSection() bool
	IsTableCaption() bool
	IsTableHeader() bool
	IsTableFooter() bool
	IsTableRow() bool
	IsDisplayNone() bool
	IsSpecifiedAsBlock() bool
	IsBlockEquivalent() bool
	IsLaidOutInInlineContext() bool
	IsNeedAutoMarginResolution() bool
	IsAbsolute() bool
	IsFixed() bool
	IsFloated() bool
	IsFloatedLeft() bool
	IsFloatedRight() bool
	IsRelative() bool
	IsPositionedOrFloated() bool
	IsPositioned() bool
	IsAutoWidth() bool
	IsAbsoluteWidth() bool
	IsAutoHeight() bool
	IsAutoLeftMargin() bool
	IsAutoRightMargin() bool
	IsAutoZIndex() bool
	EstablishesBFC() bool
	RequiresLayer() bool
	IsRunning() bool
	IsLinearGradient() bool
	GetLinearGradient(cssContext CssContext, w int, h int) *FSLinearGradient
	GetRunningName() string
	IsOverflowApplies() bool
	IsOverflowVisible() bool
	IsHorizontalBackgroundRepeat() bool
	IsVerticalBackgroundRepeat() bool
	IsTopAuto() bool
	IsBottomAuto() bool
	IsListItem() bool
	IsVisible() bool
	IsForcePageBreakBefore() bool
	IsForcePageBreakAfter() bool
	IsAvoidPageBreakInside() bool
	CreateAnonymousStyle(display *IdentValue) CalculatedStyleI
	MayHaveFirstLine() bool
	MayHaveFirstLetter() bool
	IsNonFlowContent() bool
	IsMayCollapseMarginsWithChildren() bool
	IsAbsFixedOrInlineBlockEquiv() bool
	IsMaxWidthNone() bool
	IsMaxHeightNone() bool
	IsBorderBox() bool
	GetMinWidth(c CssContext, cbWidth int) int
	GetMaxWidth(c CssContext, cbWidth int) int
	GetMinHeight(c CssContext, cbHeight int) int
	GetMaxHeight(c CssContext, cbHeight int) int
	IsCollapseBorders() bool
	GetBorderHSpacing(c CssContext) int
	GetBorderVSpacing(c CssContext) int
	GetRowSpan() int
	GetColSpan() int
	AsLength(c CssContext, cssName *CSSName) *Length
	IsShowEmptyCells() bool
	IsHasBackground() bool
	GetTextDecorations() []FSDerivedValue
	GetTextUnderlinePosition() *IdentValue
	GetTextUnderlineOffset() FSDerivedValue
	IsPaginateTable() bool
	IsTextJustify() bool
	IsListMarkerInside() bool
	IsKeepWithInline() bool
	IsDynamicAutoWidth() bool
	IsDynamicAutoWidthApplicable() bool
	IsCanBeShrunkToFit() bool
}

// CalculatedStyle is a set of properties that apply to a single Element, derived from all matched
// properties following the rules for CSS cascade, inheritance, importance,
// specificity and sequence. A derived style is just like a style but
// (presumably) has additional information that allows relative properties to be
// assigned values, e.g. font attributes. Property values are fully resolved
// when this style is created. A property retrieved by name should always have
// only one value in this class (e.g. one-one map). Any methods to retrieve
// property values from an instance of this class require a valid CssContext
// be given to it, for some cases of property
// resolution. Generally, a programmer will not use this class directly, but
// will retrieve properties using a StyleReference
// implementation.
type CalculatedStyle struct {
	// self is the outermost object (a *CalculatedStyle or an *EmptyStyle). It
	// is the value handed out wherever Java passes this as a CalculatedStyle.
	// EmptyStyle overrides nothing, so methods here call each other directly.
	self CalculatedStyleI

	// The parent-style we inherit from; nil for a root style.
	parent CalculatedStyleI

	border  *BorderPropertySet
	margin  RectPropertySetI
	padding RectPropertySetI

	lineHeight         float32
	lineHeightResolved bool

	letterSpacing         float32
	letterSpacingResolved bool

	fsFont        FSFont
	fsFontMetrics FSFontMetrics

	marginsAllowed bool
	paddingAllowed bool
	bordersAllowed bool

	backgroundSize *BackgroundSize

	// Cache child styles of this style that have the same cascaded properties.
	// Java uses a ConcurrentHashMap; here the map is guarded by childCacheMutex.
	childCache      map[string]CalculatedStyleI
	childCacheMutex sync.Mutex

	// Our main array of property values defined in this style, keyed
	// by the CSSName assigned ID.
	derivedValuesById []FSDerivedValue

	// The derived Font for this style
	font *FontSpecification
}

// initCalculatedStyle ports the private constructor CalculatedStyle(parent)
// together with the field initializers.
func (c *CalculatedStyle) initCalculatedStyle(parent CalculatedStyleI) {
	c.marginsAllowed = true
	c.paddingAllowed = true
	c.bordersAllowed = true
	c.childCache = make(map[string]CalculatedStyleI)
	c.derivedValuesById = make([]FSDerivedValue, CSSNameCountCSSNames())
	c.parent = parent
}

// NewCalculatedStyle ports the protected default constructor; as the instance
// is immutable after use, don't use this for class instantiation externally.
func NewCalculatedStyle() *CalculatedStyle {
	c := &CalculatedStyle{}
	c.initCalculatedStyle(nil)
	c.self = c
	return c
}

// newCalculatedStyleWithParentMatched is the constructor for the
// CalculatedStyle object. To get a derived style, use
// the Styler objects getDerivedStyle which will cache styles
func newCalculatedStyleWithParentMatched(parent CalculatedStyleI, matched *CascadedStyle) *CalculatedStyle {
	c := &CalculatedStyle{}
	c.initCalculatedStyle(parent)
	c.self = c

	c.derive(matched)

	display := c.GetDisplay()
	c.paddingAllowed = c.checkPaddingAllowed(display)
	c.marginsAllowed = calculatedStyleCheckMarginsAllowed(display)
	c.bordersAllowed = calculatedStyleCheckBordersAllowed(display)
	return c
}

func (c *CalculatedStyle) AsCalculatedStyle() *CalculatedStyle {
	return c
}

func (c *CalculatedStyle) checkPaddingAllowed(display *IdentValue) bool {
	return display != IdentValueTableHeaderGroup && display != IdentValueTableRowGroup &&
		display != IdentValueTableFooterGroup && display != IdentValueTableRow &&
		(!calculatedStyleIsTable(display) || !c.IsCollapseBorders())
}

func calculatedStyleIsTable(display *IdentValue) bool {
	return display == IdentValueTable || display == IdentValueInlineTable
}

func calculatedStyleCheckMarginsAllowed(display *IdentValue) bool {
	return !cssKnowledgeContains(cssKnowledgeMarginsNotAllowed, display)
}

func calculatedStyleCheckBordersAllowed(display *IdentValue) bool {
	return !cssKnowledgeContains(cssKnowledgeBordersNotAllowed, display)
}

// DeriveStyle derives a child style from this style.
//
// depends on the ability to return the identical CascadedStyle each time a child style is needed
//
// matched is the CascadedStyle to apply. Returns the derived child style.
func (c *CalculatedStyle) DeriveStyle(matched *CascadedStyle) CalculatedStyleI {
	fingerprint := matched.GetFingerprint()
	c.childCacheMutex.Lock()
	defer c.childCacheMutex.Unlock()
	if cached, ok := c.childCache[fingerprint]; ok {
		return cached
	}
	var derived CalculatedStyleI = newCalculatedStyleWithParentMatched(c.self, matched)
	c.childCache[fingerprint] = derived
	return derived
}

// GetParent returns nil for a root style.
func (c *CalculatedStyle) GetParent() CalculatedStyleI {
	return c.parent
}

func (c *CalculatedStyle) ToString() string {
	return c.genStyleKey()
}

func (c *CalculatedStyle) String() string {
	return c.ToString()
}

// AsColor may return nil.
func (c *CalculatedStyle) AsColor(cssName *CSSName) FSColor {
	prop := c.ValueByName(cssName)
	if prop == FSDerivedValue(IdentValueTransparent) {
		return FSRGBColorTransparent
	}
	return prop.AsColor()
}

func (c *CalculatedStyle) AsFloat(cssName *CSSName) float32 {
	return c.ValueByName(cssName).AsFloat()
}

func (c *CalculatedStyle) AsString(cssName *CSSName) string {
	return c.ValueByName(cssName).AsString()
}

func (c *CalculatedStyle) AsStringArray(cssName *CSSName) []string {
	return c.ValueByName(cssName).AsStringArray()
}

func (c *CalculatedStyle) SetDefaultValue(cssName *CSSName, fsDerivedValue FSDerivedValue) {
	if c.derivedValuesById[cssName.FS_ID] == nil {
		c.derivedValuesById[cssName.FS_ID] = fsDerivedValue
	}
}

// TODO: doc
func (c *CalculatedStyle) HasAbsoluteUnit(cssName *CSSName) (isAbs bool) {
	defer func() {
		if e := recover(); e != nil {
			XRLogLayout(LevelWarning, "Property "+cssName.ToString()+" has an assignment we don't understand, "+
				"and can't tell if it's an absolute unit or not. Assuming it is not. Exception was: "+
				calculatedStyleExceptionMessage(e))
			isAbs = false
		}
	}()
	isAbs = c.ValueByName(cssName).HasAbsoluteUnit()
	return isAbs
}

// calculatedStyleExceptionMessage returns what Java's Exception.getMessage()
// returns for the recovered panic value.
func calculatedStyleExceptionMessage(e any) string {
	switch v := e.(type) {
	case *XRRuntimeException:
		return v.GetMessage()
	case error:
		return v.Error()
	default:
		return fmt.Sprint(e)
	}
}

// IsIdent gets the ident attribute of the CalculatedStyle object
func (c *CalculatedStyle) IsIdent(cssName *CSSName, val *IdentValue) bool {
	return c.ValueByName(cssName) == FSDerivedValue(val)
}

// GetIdent gets the ident attribute of the CalculatedStyle object
func (c *CalculatedStyle) GetIdent(cssName *CSSName) *IdentValue {
	return c.ValueByName(cssName).AsIdentValue()
}

// GetOpacity is a convenience property accessor; returns a Opacity
// Uses the actual value (computed actual value) for this
// element.
func (c *CalculatedStyle) GetOpacity() float32 {
	opacity := c.AsFloat(CSSNameOpacity)

	for parentStyle := c.GetParent(); parentStyle != nil; parentStyle = parentStyle.GetParent() {
		opacity = opacity * parentStyle.AsFloat(CSSNameOpacity)
	}

	return opacity
}

func (c *CalculatedStyle) GetDisplay() *IdentValue {
	return c.GetIdent(CSSNameDisplay)
}

// GetColor is a convenience property accessor; returns a Color initialized with the
// foreground color Uses the actual value (computed actual value) for this
// element. May return nil.
func (c *CalculatedStyle) GetColor() FSColor {
	return c.AsColor(CSSNameColor)
}

// GetBackgroundColor is a convenience property accessor; returns a Color initialized with the
// background color value; Uses the actual value (computed actual value) for
// this element. Returns nil for a transparent background.
func (c *CalculatedStyle) GetBackgroundColor() FSColor {
	prop := c.ValueByName(CSSNameBackgroundColor)
	if prop == FSDerivedValue(IdentValueTransparent) {
		return nil
	} else {
		return c.AsColor(CSSNameBackgroundColor)
	}
}

func (c *CalculatedStyle) GetBackgroundSize() *BackgroundSize {
	if c.backgroundSize == nil {
		c.backgroundSize = c.createBackgroundSize()
	}

	return c.backgroundSize
}

func (c *CalculatedStyle) createBackgroundSize() *BackgroundSize {
	value := c.ValueByName(CSSNameBackgroundSize)
	if ident, ok := value.(*IdentValue); ok {
		if ident == IdentValueCover {
			return NewBackgroundSizeWithContainCoverBothAuto(false, true, false)
		} else if ident == IdentValueContain {
			return NewBackgroundSizeWithContainCoverBothAuto(true, false, false)
		}
	} else {
		valueList := value.(*ListValue)
		values := listValueGetPropertyValues(valueList)
		firstAuto := values[0].GetIdentValue() == IdentValueAuto
		secondAuto := values[1].GetIdentValue() == IdentValueAuto

		if firstAuto && secondAuto {
			return NewBackgroundSizeWithContainCoverBothAuto(false, false, true)
		} else {
			return NewBackgroundSize(values[0], values[1])
		}
	}

	panic(NewXRRuntimeException("Cannot created background size for " + fmt.Sprint(value)))
}

func (c *CalculatedStyle) GetBackgroundPosition() *BackgroundPosition {
	result := c.ValueByName(CSSNameBackgroundPosition).(*ListValue)
	values := listValueGetPropertyValues(result)

	return NewBackgroundPosition(values[0], values[1])
}

func (c *CalculatedStyle) HasTransform() bool {
	return !c.IsIdent(CSSNameTransform, IdentValueNone)
}

// GetTransforms returns the parsed transform functions, in the order they should be applied, or an empty
// list if transform: none. Each value's FSFunction name is one of
// matrix, translate, translateX, translateY, scale, scaleX, scaleY, rotate, skew, skewX, skewY.
func (c *CalculatedStyle) GetTransforms() []*PropertyValue {
	if !c.HasTransform() {
		return []*PropertyValue{}
	}

	value := c.ValueByName(CSSNameTransform)
	return listValueGetPropertyValues(value.(*ListValue))
}

func (c *CalculatedStyle) GetTransformOrigin() *BackgroundPosition {
	result := c.ValueByName(CSSNameTransformOrigin).(*ListValue)
	values := listValueGetPropertyValues(result)

	return NewBackgroundPosition(values[0], values[1])
}

// GetCounterReset returns nil for counter-reset: none.
func (c *CalculatedStyle) GetCounterReset() []*CounterData {
	value := c.ValueByName(CSSNameCounterReset)

	if value == FSDerivedValue(IdentValueNone) {
		return nil
	} else {
		return listValueGetCounterData(value.(*ListValue))
	}
}

// GetCounterIncrement returns nil for counter-increment: none.
func (c *CalculatedStyle) GetCounterIncrement() []*CounterData {
	value := c.ValueByName(CSSNameCounterIncrement)

	if value == FSDerivedValue(IdentValueNone) {
		return nil
	} else {
		return listValueGetCounterData(value.(*ListValue))
	}
}

// GetBorder accepts a nil ctx.
func (c *CalculatedStyle) GetBorder(ctx CssContext) *BorderPropertySet {
	if !c.bordersAllowed {
		return BorderPropertySetEmptyBorder
	} else {
		return calculatedStyleGetBorderProperty(c, ctx)
	}
}

func (c *CalculatedStyle) DisableOSBorder() bool {
	border := c.GetBorder(nil)
	return border.LeftStyle() != nil || border.RightStyle() != nil || border.TopStyle() != nil || border.BottomStyle() != nil
}

func (c *CalculatedStyle) GetFont(ctx CssContext) *FontSpecification {
	if c.font == nil {
		families := c.ValueByName(CSSNameFontFamily).AsStringArray()

		var size float32
		fontSize := c.ValueByName(CSSNameFontSize)
		if identFontSize, ok := fontSize.(*IdentValue); ok {
			resolved := c.resolveAbsoluteFontSize()
			var replacement *PropertyValue
			if resolved != nil {
				replacement = FontSizeHelperResolveAbsoluteFontSize(resolved, families)
			} else {
				replacement = FontSizeHelperGetDefaultRelativeFontSize(identFontSize)
			}

			size = LengthValueCalcFloatProportionalValue(c.self, CSSNameFontSize, replacement.GetCssText(),
				replacement.GetFloatValue(), replacement.GetPrimitiveType(), 0, ctx)
		} else {
			size = c.GetFloatPropertyProportionalTo(CSSNameFontSize, 0, ctx)
		}

		c.font = NewFontSpecification(
			size,
			c.GetIdent(CSSNameFontWeight),
			families,
			c.GetIdent(CSSNameFontStyle),
			c.GetIdent(CSSNameFontVariant),
		)
	}
	return c.font
}

// GetFontSpecification returns nil until GetFont has been called.
func (c *CalculatedStyle) GetFontSpecification() *FontSpecification {
	return c.font
}

func (c *CalculatedStyle) resolveAbsoluteFontSize() *IdentValue {
	fontSize := c.ValueByName(CSSNameFontSize)
	fontSizeIdent, ok := fontSize.(*IdentValue)
	if !ok {
		return nil
	}
	if PrimitivePropertyBuildersAbsoluteFontSizes.Get(fontSizeIdent.FS_ID) {
		return fontSizeIdent
	}

	parent := c.GetParent().AsCalculatedStyle().resolveAbsoluteFontSize()
	if parent != nil {
		if fontSizeIdent == IdentValueSmaller {
			return FontSizeHelperGetNextSmaller(parent)
		} else if fontSize == FSDerivedValue(IdentValueLarger) {
			return FontSizeHelperGetNextLarger(parent)
		}
	}

	return nil
}

func (c *CalculatedStyle) GetIntPropertyProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) int {
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(cssName, baseValue, ctx))
}

func (c *CalculatedStyle) GetFloatPropertyProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) float32 {
	return c.ValueByName(cssName).GetFloatProportionalTo(cssName, baseValue, ctx)
}

func (c *CalculatedStyle) GetFloatPropertyProportionalWidth(cssName *CSSName, parentWidth float32, ctx CssContext) float32 {
	return c.ValueByName(cssName).GetFloatProportionalTo(cssName, parentWidth, ctx)
}

func (c *CalculatedStyle) GetFloatPropertyProportionalHeight(cssName *CSSName, parentHeight float32, ctx CssContext) float32 {
	return c.ValueByName(cssName).GetFloatProportionalTo(cssName, parentHeight, ctx)
}

func (c *CalculatedStyle) GetLineHeight(ctx CssContext) float32 {
	if !c.lineHeightResolved {
		if c.IsIdent(CSSNameLineHeight, IdentValueNormal) {
			lineHeight1 := c.GetFont(ctx).Size() * 1.1
			// Make sure rasterized characters will (probably) fit inside
			// the line box
			metrics := c.GetFSFontMetrics(ctx)
			lineHeight2 := float32(math.Ceil(float64(metrics.GetDescent() + metrics.GetAscent())))
			c.lineHeight = max(lineHeight1, lineHeight2)
		} else if c.IsLength(CSSNameLineHeight) {
			//could be more elegant, I suppose
			c.lineHeight = c.GetFloatPropertyProportionalHeight(CSSNameLineHeight, 0, ctx)
		} else {
			//must be a number
			c.lineHeight = c.GetFont(ctx).Size() * c.ValueByName(CSSNameLineHeight).AsFloat()
		}
		c.lineHeightResolved = true
	}
	return c.lineHeight
}

// LetterSpacing returns the additional spacing applied after each character of inline text
// (the letter-spacing property), resolved to dots.
// Returns 0.0 for letter-spacing: normal.
func (c *CalculatedStyle) LetterSpacing(ctx CssContext) float32 {
	if !c.letterSpacingResolved {
		if c.IsIdent(CSSNameLetterSpacing, IdentValueNormal) {
			c.letterSpacing = 0.0
		} else {
			c.letterSpacing = c.GetFloatPropertyProportionalWidth(CSSNameLetterSpacing, 0, ctx)
		}
		c.letterSpacingResolved = true
	}
	return c.letterSpacing
}

// GetMarginRect is a convenience property accessor; returns a Border initialized with the
// four-sided margin width. Uses the actual value (computed actual value)
// for this element.
func (c *CalculatedStyle) GetMarginRect(cbWidth float32, ctx CssContext) RectPropertySetI {
	return c.GetMarginRectWithUseCache(cbWidth, ctx, true)
}

func (c *CalculatedStyle) GetMarginRectWithUseCache(cbWidth float32, ctx CssContext, useCache bool) RectPropertySetI {
	if !c.marginsAllowed {
		return RectPropertySetAllZeros
	} else {
		return calculatedStyleGetMarginProperty(
			c, cbWidth, ctx, useCache)
	}
}

// GetPaddingRect is a convenience property accessor; returns a Border initialized with the
// four-sided padding width. Uses the actual value (computed actual value)
// for this element.
func (c *CalculatedStyle) GetPaddingRect(cbWidth float32, ctx CssContext) RectPropertySetI {
	if !c.paddingAllowed {
		return RectPropertySetAllZeros
	} else {
		return calculatedStyleGetPaddingProperty(c, cbWidth, ctx)
	}
}

func (c *CalculatedStyle) GetStringProperty(cssName *CSSName) string {
	return c.ValueByName(cssName).AsString()
}

func (c *CalculatedStyle) IsLength(cssName *CSSName) bool {
	val := c.ValueByName(cssName)
	_, ok := val.(*LengthValue)
	return ok
}

func (c *CalculatedStyle) IsLengthOrNumber(cssName *CSSName) bool {
	val := c.ValueByName(cssName)
	_, isNumber := val.(*NumberValue)
	_, isLength := val.(*LengthValue)
	return isNumber || isLength
}

// ValueByName returns a FSDerivedValue by name. Because we are a derived
// style, the property will already be resolved at this point.
//
// cssName is the CSS property name, e.g. "font-family"
func (c *CalculatedStyle) ValueByName(cssName *CSSName) FSDerivedValue {
	val := c.derivedValuesById[cssName.FS_ID]

	needInitialValue := val == FSDerivedValue(IdentValueFsInitialValue)

	// but the property may not be defined for this Element
	if val == nil || needInitialValue {
		// if it is inheritable (like color) and we are not root, ask our parent
		// for the value
		inherited := false
		if !needInitialValue && CSSNamePropertyInherits(cssName) &&
			c.parent != nil {
			val = c.parent.ValueByName(cssName)
			inherited = val != nil
		}
		if !inherited {
			// otherwise, use the initial value (defined by the CSS2 Spec)
			initialValue := CSSNameInitialValue(cssName)
			if initialValue == "" {
				panic(NewXRRuntimeException("Property '" + cssName.ToString() + "' has no initial values assigned. " +
					"Check CSSName declarations."))
			}
			if initialValue[0] == '=' {
				ref := CSSNameCssProperty(initialValue[1:])
				val = c.ValueByName(ref)
			} else {
				val = cssName.InitialDerivedValue()
			}
		}
		c.derivedValuesById[cssName.FS_ID] = val
	}
	return val
}

// derive implements cascade/inherit/important logic. This should result in the
// element for this style having a value for *each and every* (visual)
// property in the CSS2 spec. The implementation is based on the notion that
// the matched styles are given to us in a perfectly sorted order, such that
// properties appearing later in the rule-set always override properties
// appearing earlier. It also assumes that all properties in the CSS2 spec
// are defined somewhere across all the matched styles; for example, that
// the full-property set is given in the user-agent CSS that is always
// loaded with styles. The current implementation makes no attempt to check
// either of these assumptions. When this method exits, the derived property
// list for this class will be populated with the properties defined for
// this element, properly cascaded.
func (c *CalculatedStyle) derive(matched *CascadedStyle) {
	for _, pd := range matched.GetCascadedPropertyDeclarations() {
		val := c.deriveValue(pd.GetCSSName(), pd.GetValue())
		c.derivedValuesById[pd.GetCSSName().FS_ID] = val
	}
}

func (c *CalculatedStyle) deriveValue(cssName *CSSName, value *PropertyValue) FSDerivedValue {
	return DerivedValueFactoryNewDerivedValue(c.self, cssName, value)
}

func (c *CalculatedStyle) genStyleKey() string {
	var sb strings.Builder
	for i := 0; i < len(c.derivedValuesById); i++ {
		name := CSSNameGetByID(i)
		val := c.derivedValuesById[i]
		if val != nil {
			sb.WriteString(name.ToString())
		} else {
			sb.WriteString("(no prop assigned in this pos)")
		}
		sb.WriteString("|\n")
	}
	return sb.String()

}

func (c *CalculatedStyle) GetCachedPadding() RectPropertySetI {
	if c.padding == nil {
		panic(NewXRRuntimeException("No padding property cached yet; should have called getPropertyRect() at least once before."))
	} else {
		return c.padding
	}
}

func calculatedStyleGetPaddingProperty(style *CalculatedStyle,
	cbWidth float32,
	ctx CssContext) RectPropertySetI {
	if style.padding == nil {
		style.padding = calculatedStyleNewRectInstance(style, CSSNamePaddingSideProperties, cbWidth, ctx).
			ResetNegativeValues()
	}

	return style.padding
}

func calculatedStyleGetMarginProperty(style *CalculatedStyle,
	cbWidth float32,
	ctx CssContext,
	useCache bool) RectPropertySetI {
	if !useCache {
		return calculatedStyleNewRectInstance(style, CSSNameMarginSideProperties, cbWidth, ctx)
	} else {
		if style.margin == nil {
			style.margin = calculatedStyleNewRectInstance(style, CSSNameMarginSideProperties, cbWidth, ctx)
		}

		return style.margin
	}
}

func calculatedStyleNewRectInstance(style *CalculatedStyle,
	sides *CSSNameCSSSideProperties,
	cbWidth float32,
	ctx CssContext) RectPropertySetI {
	rect := RectPropertySetNewInstance(style.self, sides, cbWidth, ctx)

	if rect.IsAllZeros() {
		rect = RectPropertySetAllZeros
	}
	return rect
}

func calculatedStyleGetBorderProperty(style *CalculatedStyle,
	ctx CssContext) *BorderPropertySet {
	if style.border == nil {
		style.border = BorderPropertySetNewInstance(style.self, ctx).ResetNegativeValues().(*BorderPropertySet)
	}
	return style.border
}

// CalculatedStyleEdge ports the enum CalculatedStyle.Edge.
type CalculatedStyleEdge struct {
	name    string
	ordinal int
}

var (
	CalculatedStyleEdgeLeft   = &CalculatedStyleEdge{"LEFT", 0}
	CalculatedStyleEdgeRight  = &CalculatedStyleEdge{"RIGHT", 1}
	CalculatedStyleEdgeTop    = &CalculatedStyleEdge{"TOP", 2}
	CalculatedStyleEdgeBottom = &CalculatedStyleEdge{"BOTTOM", 3}
)

func (e *CalculatedStyleEdge) Name() string {
	return e.name
}

func (e *CalculatedStyleEdge) Ordinal() int {
	return e.ordinal
}

func (e *CalculatedStyleEdge) String() string {
	return e.name
}

func (e *CalculatedStyleEdge) GetMarginBorderPadding(margin RectPropertySetI, border *BorderPropertySet, padding RectPropertySetI) int {
	switch e {
	case CalculatedStyleEdgeLeft:
		return calculatedStyleFloatToInt(margin.Left() + border.Left() + padding.Left())
	case CalculatedStyleEdgeRight:
		return calculatedStyleFloatToInt(margin.Right() + border.Right() + padding.Right())
	case CalculatedStyleEdgeTop:
		return calculatedStyleFloatToInt(margin.Top() + border.Top() + padding.Top())
	default:
		return calculatedStyleFloatToInt(margin.Bottom() + border.Bottom() + padding.Bottom())
	}
}

func (c *CalculatedStyle) GetMarginBorderPadding(cssCtx CssContext, cbWidth int, edge *CalculatedStyleEdge) int {
	border := c.GetBorder(cssCtx)
	margin := c.GetMarginRect(float32(cbWidth), cssCtx)
	padding := c.GetPaddingRect(float32(cbWidth), cssCtx)

	return edge.GetMarginBorderPadding(margin, border, padding)
}

// getLengthValue returns nil when the property is not a length.
func (c *CalculatedStyle) getLengthValue(cssName *CSSName) *int {
	widthValue := c.ValueByName(cssName)
	if length, ok := widthValue.(*LengthValue); ok {
		result := calculatedStyleFloatToInt(length.AsFloat())
		return &result
	}

	return nil
}

func (c *CalculatedStyle) Padding() *NullableInsets {
	paddingTop := c.getLengthValue(CSSNamePaddingTop)
	paddingLeft := c.getLengthValue(CSSNamePaddingLeft)
	paddingBottom := c.getLengthValue(CSSNamePaddingBottom)
	paddingRight := c.getLengthValue(CSSNamePaddingRight)
	return NewNullableInsets(paddingTop, paddingLeft, paddingBottom, paddingRight)
}

func (c *CalculatedStyle) GetWhitespace() *IdentValue {
	return c.GetIdent(CSSNameWhiteSpace)
}

func (c *CalculatedStyle) GetFSFont(cssContext CssContext) FSFont {
	if c.fsFont == nil {
		c.fsFont = cssContext.GetFont(c.GetFont(cssContext))
	}
	return c.fsFont
}

func (c *CalculatedStyle) GetFSFontMetrics(ctx CssContext) FSFontMetrics {
	if c.fsFontMetrics == nil {
		c.fsFontMetrics = ctx.GetFSFontMetrics(c.GetFSFont(ctx))
	}
	return c.fsFontMetrics
}

func (c *CalculatedStyle) GetWordWrap() *IdentValue {
	return c.GetIdent(CSSNameWordWrap)
}

func (c *CalculatedStyle) GetWordBreak() *IdentValue {
	return c.GetIdent(CSSNameWordBreak)
}

func (c *CalculatedStyle) GetHyphens() *IdentValue {
	return c.GetIdent(CSSNameHyphens)
}

func (c *CalculatedStyle) IsClearLeft() bool {
	clear := c.GetIdent(CSSNameClear)
	return clear == IdentValueLeft || clear == IdentValueBoth
}

func (c *CalculatedStyle) IsClearRight() bool {
	clear := c.GetIdent(CSSNameClear)
	return clear == IdentValueRight || clear == IdentValueBoth
}

func (c *CalculatedStyle) IsCleared() bool {
	return !c.IsIdent(CSSNameClear, IdentValueNone)
}

func (c *CalculatedStyle) GetBackgroundRepeat() *IdentValue {
	return c.GetIdent(CSSNameBackgroundRepeat)
}

func (c *CalculatedStyle) GetBackgroundAttachment() *IdentValue {
	return c.GetIdent(CSSNameBackgroundAttachment)
}

func (c *CalculatedStyle) IsFixedBackground() bool {
	return c.GetIdent(CSSNameBackgroundAttachment) == IdentValueFixed
}

func (c *CalculatedStyle) IsInline() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueInline) &&
		!(c.IsFloated() || c.IsAbsolute() || c.IsFixed() || c.IsRunning())
}

func (c *CalculatedStyle) IsInlineBlock() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueInlineBlock)
}

func (c *CalculatedStyle) IsTable() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueTable)
}

func (c *CalculatedStyle) IsInlineTable() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueInlineTable)
}

func (c *CalculatedStyle) IsUnderTableLayout() bool {
	return cssKnowledgeContains(cssKnowledgeUnderTableLayout, c.GetDisplay())
}

func (c *CalculatedStyle) IsTableCell() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueTableCell)
}

func (c *CalculatedStyle) IsTableSection() bool {
	return cssKnowledgeContains(cssKnowledgeTableSections, c.GetDisplay())
}

func (c *CalculatedStyle) IsTableCaption() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueTableCaption)
}

func (c *CalculatedStyle) IsTableHeader() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueTableHeaderGroup)
}

func (c *CalculatedStyle) IsTableFooter() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueTableFooterGroup)
}

func (c *CalculatedStyle) IsTableRow() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueTableRow)
}

func (c *CalculatedStyle) IsDisplayNone() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueNone)
}

func (c *CalculatedStyle) IsSpecifiedAsBlock() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueBlock)
}

func (c *CalculatedStyle) IsBlockEquivalent() bool {
	if c.IsFloated() || c.IsAbsolute() || c.IsFixed() {
		return true
	} else {
		return cssKnowledgeContains(cssKnowledgeBlockEquivalents, c.GetDisplay())
	}
}

func (c *CalculatedStyle) IsLaidOutInInlineContext() bool {
	if c.IsFloated() || c.IsAbsolute() || c.IsFixed() || c.IsRunning() {
		return true
	} else {
		return cssKnowledgeContains(cssKnowledgeLaidOutInInlineContext, c.GetDisplay())
	}
}

func (c *CalculatedStyle) IsNeedAutoMarginResolution() bool {
	return !(c.IsAbsolute() || c.IsFixed() || c.IsFloated() || c.IsInlineBlock())
}

func (c *CalculatedStyle) IsAbsolute() bool {
	return c.IsIdent(CSSNamePosition, IdentValueAbsolute)
}

func (c *CalculatedStyle) IsFixed() bool {
	return c.IsIdent(CSSNamePosition, IdentValueFixed)
}

func (c *CalculatedStyle) IsFloated() bool {
	if c.IsUnderTableLayout() {
		return false
	}
	floatVal := c.GetIdent(CSSNameFloat)
	return floatVal == IdentValueLeft || floatVal == IdentValueRight
}

func (c *CalculatedStyle) IsFloatedLeft() bool {
	return c.IsIdent(CSSNameFloat, IdentValueLeft)
}

func (c *CalculatedStyle) IsFloatedRight() bool {
	return c.IsIdent(CSSNameFloat, IdentValueRight)
}

func (c *CalculatedStyle) IsRelative() bool {
	return c.IsIdent(CSSNamePosition, IdentValueRelative)
}

func (c *CalculatedStyle) IsPositionedOrFloated() bool {
	return c.IsAbsolute() || c.IsFixed() || c.IsFloated() || c.IsRelative()
}

func (c *CalculatedStyle) IsPositioned() bool {
	return c.IsAbsolute() || c.IsFixed() || c.IsRelative()
}

func (c *CalculatedStyle) IsAutoWidth() bool {
	return c.IsIdent(CSSNameWidth, IdentValueAuto)
}

func (c *CalculatedStyle) IsAbsoluteWidth() bool {
	return c.ValueByName(CSSNameWidth).HasAbsoluteUnit()
}

func (c *CalculatedStyle) IsAutoHeight() bool {
	return c.IsIdent(CSSNameHeight, IdentValueAuto)
}

func (c *CalculatedStyle) IsAutoLeftMargin() bool {
	return c.IsIdent(CSSNameMarginLeft, IdentValueAuto)
}

func (c *CalculatedStyle) IsAutoRightMargin() bool {
	return c.IsIdent(CSSNameMarginRight, IdentValueAuto)
}

func (c *CalculatedStyle) IsAutoZIndex() bool {
	return c.IsIdent(CSSNameZIndex, IdentValueAuto)
}

func (c *CalculatedStyle) EstablishesBFC() bool {
	value := c.ValueByName(CSSNamePosition)

	if _, ok := value.(*FunctionValue); ok { // running(header)
		return false
	} else {
		display := c.GetDisplay()
		position := value.(*IdentValue)

		return c.IsFloated() ||
			position == IdentValueAbsolute || position == IdentValueFixed ||
			display == IdentValueInlineBlock || display == IdentValueTableCell ||
			!c.IsIdent(CSSNameOverflow, IdentValueVisible)
	}
}

func (c *CalculatedStyle) RequiresLayer() bool {
	if c.HasTransform() {
		return true
	}

	value := c.ValueByName(CSSNamePosition)

	if _, ok := value.(*FunctionValue); ok { // running(header)
		return false
	} else {
		position := c.GetIdent(CSSNamePosition)

		if position == IdentValueAbsolute ||
			position == IdentValueRelative || position == IdentValueFixed {
			return true
		}

		overflow := c.GetIdent(CSSNameOverflow)
		return (overflow == IdentValueScroll || overflow == IdentValueAuto) &&
			c.IsOverflowApplies()
	}
}

func (c *CalculatedStyle) IsRunning() bool {
	value := c.ValueByName(CSSNamePosition)
	_, ok := value.(*FunctionValue)
	return ok
}

func (c *CalculatedStyle) IsLinearGradient() bool {
	value := c.ValueByName(CSSNameBackgroundImage)
	function, ok := value.(*FunctionValue)
	return ok &&
		GeneralUtilCiEquals(function.GetFunction().GetName(), "linear-gradient")
}

func (c *CalculatedStyle) GetLinearGradient(cssContext CssContext, w int, h int) *FSLinearGradient {
	value := c.ValueByName(CSSNameBackgroundImage).(*FunctionValue)
	return NewFSLinearGradient(value.GetFunction(), c.self, w, h, cssContext)
}

func (c *CalculatedStyle) GetRunningName() string {
	value := c.ValueByName(CSSNamePosition).(*FunctionValue)
	function := value.GetFunction()
	param := function.GetParameters()[0]
	return param.GetStringValue()
}

func (c *CalculatedStyle) IsOverflowApplies() bool {
	return cssKnowledgeContains(cssKnowledgeOverflowApplicable, c.GetDisplay())
}

func (c *CalculatedStyle) IsOverflowVisible() bool {
	return c.ValueByName(CSSNameOverflow) == FSDerivedValue(IdentValueVisible)
}

func (c *CalculatedStyle) IsHorizontalBackgroundRepeat() bool {
	value := c.GetIdent(CSSNameBackgroundRepeat)
	return value == IdentValueRepeatX || value == IdentValueRepeat
}

func (c *CalculatedStyle) IsVerticalBackgroundRepeat() bool {
	value := c.GetIdent(CSSNameBackgroundRepeat)
	return value == IdentValueRepeatY || value == IdentValueRepeat
}

func (c *CalculatedStyle) IsTopAuto() bool {
	return c.IsIdent(CSSNameTop, IdentValueAuto)
}

func (c *CalculatedStyle) IsBottomAuto() bool {
	return c.IsIdent(CSSNameBottom, IdentValueAuto)
}

func (c *CalculatedStyle) IsListItem() bool {
	return c.IsIdent(CSSNameDisplay, IdentValueListItem)
}

func (c *CalculatedStyle) IsVisible() bool {
	return c.IsIdent(CSSNameVisibility, IdentValueVisible)
}

func (c *CalculatedStyle) IsForcePageBreakBefore() bool {
	val := c.GetIdent(CSSNamePageBreakBefore)
	return val == IdentValueAlways || val == IdentValueLeft ||
		val == IdentValueRight
}

func (c *CalculatedStyle) IsForcePageBreakAfter() bool {
	val := c.GetIdent(CSSNamePageBreakAfter)
	return val == IdentValueAlways || val == IdentValueLeft ||
		val == IdentValueRight
}

func (c *CalculatedStyle) IsAvoidPageBreakInside() bool {
	return c.IsIdent(CSSNamePageBreakInside, IdentValueAvoid)
}

func (c *CalculatedStyle) CreateAnonymousStyle(display *IdentValue) CalculatedStyleI {
	return c.DeriveStyle(CascadedStyleCreateAnonymousStyle(display))
}

func (c *CalculatedStyle) MayHaveFirstLine() bool {
	return cssKnowledgeContains(cssKnowledgeMayHaveFirstLine, c.GetDisplay())
}

func (c *CalculatedStyle) MayHaveFirstLetter() bool {
	return cssKnowledgeContains(cssKnowledgeMayHaveFirstLetter, c.GetDisplay())
}

func (c *CalculatedStyle) IsNonFlowContent() bool {
	return c.IsFloated() || c.IsAbsolute() || c.IsFixed() || c.IsRunning()
}

func (c *CalculatedStyle) IsMayCollapseMarginsWithChildren() bool {
	return c.IsIdent(CSSNameOverflow, IdentValueVisible) &&
		!(c.IsFloated() || c.IsAbsolute() || c.IsFixed() || c.IsInlineBlock())
}

func (c *CalculatedStyle) IsAbsFixedOrInlineBlockEquiv() bool {
	return c.IsAbsolute() || c.IsFixed() || c.IsInlineBlock() || c.IsInlineTable()
}

func (c *CalculatedStyle) IsMaxWidthNone() bool {
	return c.IsIdent(CSSNameMaxWidth, IdentValueNone)
}

func (c *CalculatedStyle) IsMaxHeightNone() bool {
	return c.IsIdent(CSSNameMaxHeight, IdentValueNone)
}

func (c *CalculatedStyle) IsBorderBox() bool {
	return c.IsIdent(CSSNameBoxSizing, IdentValueBorderBox)
}

func (c *CalculatedStyle) GetMinWidth(ctx CssContext, cbWidth int) int {
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(CSSNameMinWidth, float32(cbWidth), ctx))
}

func (c *CalculatedStyle) GetMaxWidth(ctx CssContext, cbWidth int) int {
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(CSSNameMaxWidth, float32(cbWidth), ctx))
}

func (c *CalculatedStyle) GetMinHeight(ctx CssContext, cbHeight int) int {
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(CSSNameMinHeight, float32(cbHeight), ctx))
}

func (c *CalculatedStyle) GetMaxHeight(ctx CssContext, cbHeight int) int {
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(CSSNameMaxHeight, float32(cbHeight), ctx))
}

func (c *CalculatedStyle) IsCollapseBorders() bool {
	return c.IsIdent(CSSNameBorderCollapse, IdentValueCollapse)
}

func (c *CalculatedStyle) GetBorderHSpacing(ctx CssContext) int {
	if c.IsCollapseBorders() {
		return 0
	}
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(CSSNameFsBorderSpacingHorizontal, 0, ctx))
}

func (c *CalculatedStyle) GetBorderVSpacing(ctx CssContext) int {
	if c.IsCollapseBorders() {
		return 0
	}
	return calculatedStyleFloatToInt(c.GetFloatPropertyProportionalTo(CSSNameFsBorderSpacingVertical, 0, ctx))
}

func (c *CalculatedStyle) GetRowSpan() int {
	result := calculatedStyleFloatToInt(c.AsFloat(CSSNameFsRowspan))
	if result > 0 {
		return result
	}
	return 1
}

func (c *CalculatedStyle) GetColSpan() int {
	result := calculatedStyleFloatToInt(c.AsFloat(CSSNameFsColspan))
	if result > 0 {
		return result
	}
	return 1
}

func (c *CalculatedStyle) AsLength(ctx CssContext, cssName *CSSName) *Length {
	value := c.ValueByName(cssName)
	_, isLength := value.(*LengthValue)
	_, isNumber := value.(*NumberValue)
	if isLength || isNumber {
		if value.HasAbsoluteUnit() {
			return NewLength(int64(calculatedStyleFloatToInt(value.GetFloatProportionalTo(cssName, 0, ctx))), LengthLengthTypeFixed)
		} else {
			return NewLength(int64(calculatedStyleFloatToInt(value.AsFloat())), LengthLengthTypePercent)
		}
	}

	return LengthZero
}

func (c *CalculatedStyle) IsShowEmptyCells() bool {
	return c.IsCollapseBorders() || c.IsIdent(CSSNameEmptyCells, IdentValueShow)
}

func (c *CalculatedStyle) IsHasBackground() bool {
	return !(c.IsIdent(CSSNameBackgroundColor, IdentValueTransparent) &&
		c.IsIdent(CSSNameBackgroundImage, IdentValueNone))
}

// GetTextDecorations returns nil for text-decoration: none.
func (c *CalculatedStyle) GetTextDecorations() []FSDerivedValue {
	value := c.ValueByName(CSSNameTextDecoration)
	if value == FSDerivedValue(IdentValueNone) {
		return nil
	} else {
		idents := listValueGetPropertyValues(value.(*ListValue))
		result := make([]FSDerivedValue, 0, len(idents))
		for _, ident := range idents {
			result = append(result, DerivedValueFactoryNewDerivedValue(
				c.self, CSSNameTextDecoration, ident))
		}
		return result
	}
}

func (c *CalculatedStyle) GetTextUnderlinePosition() *IdentValue {
	return c.GetIdent(CSSNameTextUnderlinePosition)
}

func (c *CalculatedStyle) GetTextUnderlineOffset() FSDerivedValue {
	return c.ValueByName(CSSNameTextUnderlineOffset)
}

// getCursor() is not ported: it returns a java.awt.Cursor and is only used by
// the Swing renderer.

func (c *CalculatedStyle) IsPaginateTable() bool {
	return c.IsIdent(CSSNameFsTablePaginate, IdentValuePaginate)
}

func (c *CalculatedStyle) IsTextJustify() bool {
	return c.IsIdent(CSSNameTextAlign, IdentValueJustify) &&
		!(c.IsIdent(CSSNameWhiteSpace, IdentValuePre) ||
			c.IsIdent(CSSNameWhiteSpace, IdentValuePreLine))
}

func (c *CalculatedStyle) IsListMarkerInside() bool {
	return c.IsIdent(CSSNameListStylePosition, IdentValueInside)
}

func (c *CalculatedStyle) IsKeepWithInline() bool {
	return c.IsIdent(CSSNameFsKeepWithInline, IdentValueKeep)
}

func (c *CalculatedStyle) IsDynamicAutoWidth() bool {
	return c.IsIdent(CSSNameFsDynamicAutoWidth, IdentValueDynamic)
}

func (c *CalculatedStyle) IsDynamicAutoWidthApplicable() bool {
	return c.IsDynamicAutoWidth() && c.IsAutoWidth() && !c.IsCanBeShrunkToFit()
}

func (c *CalculatedStyle) IsCanBeShrunkToFit() bool {
	return c.IsInlineBlock() || c.IsFloated() || c.IsAbsolute() || c.IsFixed()
}

// calculatedStyleFloatToInt is Java's (int) cast of a float: truncation toward
// zero, NaN to 0, and saturation at the int range.
func calculatedStyleFloatToInt(f float32) int {
	if f != f {
		return 0
	}
	if f >= math.MaxInt32 {
		return math.MaxInt32
	}
	if f <= math.MinInt32 {
		return math.MinInt32
	}
	return int(f)
}
