// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/CalculatedStyle.java.
// Flying Saucer has no JUnit test for the class; these cover derivation,
// inheritance, initial values, length resolution and the per-style caches.

package ufo

import "testing"

// calculatedStyleTestContext is a CssContext with 20 dots per pixel and the
// resolution the PDF renderer uses (96 px per inch).
type calculatedStyleTestContext struct {
	fontsResolved int
}

func (c *calculatedStyleTestContext) GetMmPerDot() float32 {
	return 25.4 / (96 * 20)
}

func (c *calculatedStyleTestContext) GetDotsPerPixel() int {
	return 20
}

func (c *calculatedStyleTestContext) GetFontSize2D(font *FontSpecification) float32 {
	return font.Size()
}

func (c *calculatedStyleTestContext) GetXHeight(parentFont *FontSpecification) float32 {
	return parentFont.Size() / 2
}

func (c *calculatedStyleTestContext) GetFont(font *FontSpecification) FSFont {
	c.fontsResolved++
	return calculatedStyleTestFont{size: font.Size()}
}

func (c *calculatedStyleTestContext) GetCss() *StyleReference {
	return nil
}

func (c *calculatedStyleTestContext) GetTextRenderer() TextRenderer {
	return nil
}

func (c *calculatedStyleTestContext) GetFontContext() FontContext {
	return nil
}

func (c *calculatedStyleTestContext) GetFSFontMetrics(font FSFont) FSFontMetrics {
	return calculatedStyleTestFontMetrics{size: font.GetSize2D()}
}

type calculatedStyleTestFont struct {
	size float32
}

func (f calculatedStyleTestFont) GetSize2D() float32 {
	return f.size
}

type calculatedStyleTestFontMetrics struct {
	size float32
}

func (m calculatedStyleTestFontMetrics) GetAscent() float32                 { return m.size }
func (m calculatedStyleTestFontMetrics) GetDescent() float32                { return m.size / 2 }
func (m calculatedStyleTestFontMetrics) GetStrikethroughOffset() float32    { return 0 }
func (m calculatedStyleTestFontMetrics) GetStrikethroughThickness() float32 { return 0 }
func (m calculatedStyleTestFontMetrics) GetUnderlineOffset() float32        { return 0 }
func (m calculatedStyleTestFontMetrics) GetUnderlineThickness() float32     { return 0 }

func calculatedStyleTestDeclaration(cssName *CSSName, value *PropertyValue) *PropertyDeclaration {
	return NewPropertyDeclaration(cssName, value, false, StylesheetInfoOriginAuthor)
}

func calculatedStyleTestLength(cssName *CSSName, unit int16, value float32, cssText string) *PropertyDeclaration {
	return calculatedStyleTestDeclaration(cssName, NewPropertyValueFloat(unit, value, cssText))
}

func calculatedStyleTestIdent(cssName *CSSName, ident *IdentValue) *PropertyDeclaration {
	return calculatedStyleTestDeclaration(cssName, NewPropertyValueIdentValue(ident))
}

func TestCalculatedStyle_deriveStyleCachesByFingerprint(t *testing.T) {
	root := NewEmptyStyle()
	matched := CascadedStyleCreateLayoutStyle(calculatedStyleTestIdent(CSSNameDisplay, IdentValueBlock))

	first := root.DeriveStyle(matched)
	second := root.DeriveStyle(matched)
	if first != second {
		t.Errorf("DeriveStyle returned two styles for one CascadedStyle")
	}
	if first.GetParent() != CalculatedStyleI(root) {
		t.Errorf("GetParent() = %v, want the EmptyStyle the child was derived from", first.GetParent())
	}
	if root.GetParent() != nil {
		t.Errorf("root GetParent() = %v, want nil", root.GetParent())
	}

	other := root.DeriveStyle(CascadedStyleCreateLayoutStyle(calculatedStyleTestIdent(CSSNameDisplay, IdentValueInline)))
	if other == first {
		t.Errorf("DeriveStyle returned one style for two different CascadedStyles")
	}
}

func TestCalculatedStyle_valueByNameInheritsAndUsesInitialValues(t *testing.T) {
	root := NewEmptyStyle()
	parent := root.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestDeclaration(CSSNameColor, NewPropertyValueFSColor(FSRGBColorBlue)),
		calculatedStyleTestIdent(CSSNameDisplay, IdentValueBlock),
		calculatedStyleTestLength(CSSNameWidth, CSSPrimitiveValueCssPx, 10, "10px"),
	))
	child := parent.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameFloat, IdentValueLeft),
	))

	// color inherits
	// DerivedValueFactory caches colors by their CSS text across the whole
	// package, as Java's static cache does, so compare values, not instances.
	if got := child.GetColor(); !FSRGBColorBlue.Equals(got) {
		t.Errorf("child GetColor() = %v, want the parent's blue", got)
	}
	// display and width do not inherit; their initial values are inline and auto
	if got := child.GetDisplay(); got != IdentValueInline {
		t.Errorf("child GetDisplay() = %v, want inline", got)
	}
	if !child.IsAutoWidth() {
		t.Errorf("child IsAutoWidth() = false, want true")
	}
	if parent.IsAutoWidth() {
		t.Errorf("parent IsAutoWidth() = true, want false")
	}
	if !parent.IsLength(CSSNameWidth) || !parent.IsAbsoluteWidth() {
		t.Errorf("parent width is not an absolute length")
	}
	if !child.IsFloated() || !child.IsFloatedLeft() || child.IsInline() || !child.IsBlockEquivalent() {
		t.Errorf("float: left child: IsFloated=%v IsFloatedLeft=%v IsInline=%v IsBlockEquivalent=%v",
			child.IsFloated(), child.IsFloatedLeft(), child.IsInline(), child.IsBlockEquivalent())
	}
	// background-color: transparent is reported as no color
	if got := child.GetBackgroundColor(); got != nil {
		t.Errorf("GetBackgroundColor() = %v, want nil", got)
	}
	if got := child.AsColor(CSSNameBackgroundColor); got != FSColor(FSRGBColorTransparent) {
		t.Errorf("AsColor(background-color) = %v, want FSRGBColorTransparent", got)
	}
	if child.IsHasBackground() {
		t.Errorf("IsHasBackground() = true, want false")
	}
}

func TestCalculatedStyle_lengthsResolveAgainstTheContext(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	root := NewEmptyStyle()
	parent := root.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameDisplay, IdentValueBlock),
		calculatedStyleTestLength(CSSNameFontSize, CSSPrimitiveValueCssPx, 10, "10px"),
	))
	child := parent.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameDisplay, IdentValueBlock),
		calculatedStyleTestLength(CSSNameFontSize, CSSPrimitiveValueCssEms, 2, "2em"),
		calculatedStyleTestLength(CSSNameWidth, CSSPrimitiveValueCssPercentage, 50, "50%"),
		calculatedStyleTestLength(CSSNameHeight, CSSPrimitiveValueCssIn, 1, "1in"),
		calculatedStyleTestLength(CSSNameMarginLeft, CSSPrimitiveValueCssEms, 1.5, "1.5em"),
		calculatedStyleTestLength(CSSNamePaddingTop, CSSPrimitiveValueCssPx, -3, "-3px"),
		calculatedStyleTestLength(CSSNamePaddingLeft, CSSPrimitiveValueCssPercentage, 10, "10%"),
	))

	// 10px at 20 dots per pixel
	if got := parent.GetFont(ctx).Size(); got != 200 {
		t.Errorf("parent font size = %v, want 200", got)
	}
	// 2em of the parent's font size
	if got := child.GetFont(ctx).Size(); got != 400 {
		t.Errorf("child font size = %v, want 400", got)
	}
	if got := child.GetFloatPropertyProportionalWidth(CSSNameWidth, 1000, ctx); got != 500 {
		t.Errorf("width 50%% of 1000 = %v, want 500", got)
	}
	// 1in = 96px = 1920 dots
	if got := child.GetFloatPropertyProportionalHeight(CSSNameHeight, 0, ctx); got != 1920 {
		t.Errorf("height 1in = %v, want 1920", got)
	}

	margin := child.GetMarginRect(1000, ctx)
	// 1.5em of the element's own font size
	if margin.Left() != 600 || margin.Right() != 0 || margin.Top() != 0 || margin.Bottom() != 0 {
		t.Errorf("margin = %s, want left=600 and the other sides 0", margin.ToString())
	}
	if child.GetMarginRect(2000, ctx) != margin {
		t.Errorf("GetMarginRect did not return the cached margin")
	}
	if child.GetMarginRectWithUseCache(1000, ctx, false) == margin {
		t.Errorf("GetMarginRectWithUseCache(useCache=false) returned the cached margin")
	}

	padding := child.GetPaddingRect(1000, ctx)
	// the negative top padding is reset to 0
	if padding.Top() != 0 || padding.Left() != 100 {
		t.Errorf("padding = %s, want top=0 left=100", padding.ToString())
	}
	if child.GetCachedPadding() != padding {
		t.Errorf("GetCachedPadding() did not return the padding computed by GetPaddingRect")
	}

	if got := child.GetMarginBorderPadding(ctx, 1000, CalculatedStyleEdgeLeft); got != 700 {
		t.Errorf("GetMarginBorderPadding(left) = %v, want 700", got)
	}

	insets := child.Padding()
	if insets.Top() == nil || *insets.Top() != -3 || insets.Left() == nil || *insets.Left() != 10 {
		t.Errorf("Padding() top=%v left=%v, want -3 and 10", insets.Top(), insets.Left())
	}
	if insets.Right() != nil && *insets.Right() != 0 {
		t.Errorf("Padding() right = %v, want nil or 0", *insets.Right())
	}

	// line-height: normal is max(1.1 * font size, ceil(ascent + descent))
	if got := child.GetLineHeight(ctx); got != 600 {
		t.Errorf("GetLineHeight() = %v, want 600", got)
	}
	if got := child.LetterSpacing(ctx); got != 0 {
		t.Errorf("LetterSpacing() = %v, want 0", got)
	}

	// the FSFont is resolved once per style
	before := ctx.fontsResolved
	child.GetFSFont(ctx)
	child.GetFSFont(ctx)
	if ctx.fontsResolved != before {
		t.Errorf("GetFSFont resolved the font again: %d calls, want %d", ctx.fontsResolved, before)
	}

	length := child.AsLength(ctx, CSSNameWidth)
	if !length.IsPercent() || length.Value() != 50 {
		t.Errorf("AsLength(width) = %s, want percent 50", length)
	}
	length = child.AsLength(ctx, CSSNameHeight)
	if !length.IsFixed() || length.Value() != 1920 {
		t.Errorf("AsLength(height) = %s, want fixed 1920", length)
	}
	if got := parent.AsLength(ctx, CSSNameWidth); got != LengthZero {
		t.Errorf("AsLength(width: auto) = %s, want LengthZero", got)
	}
}

func TestCalculatedStyle_absoluteFontSizeKeywords(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	root := NewEmptyStyle()
	parent := root.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameFontSize, IdentValueLarge),
	))
	smaller := parent.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameFontSize, IdentValueSmaller),
	))
	larger := parent.DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameFontSize, IdentValueLarger),
		calculatedStyleTestIdent(CSSNameDisplay, IdentValueBlock),
	))

	// large = 18px, the next smaller is medium = 16px, the next larger x-large = 24px
	if got := parent.GetFont(ctx).Size(); got != 18*20 {
		t.Errorf("large = %v, want %v", got, 18*20)
	}
	if got := smaller.GetFont(ctx).Size(); got != 16*20 {
		t.Errorf("smaller than large = %v, want %v", got, 16*20)
	}
	if got := larger.GetFont(ctx).Size(); got != 24*20 {
		t.Errorf("larger than large = %v, want %v", got, 24*20)
	}
}

func TestCalculatedStyle_tableRowAllowsNoMarginsPaddingOrBorders(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	row := NewEmptyStyle().DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameDisplay, IdentValueTableRow),
		calculatedStyleTestLength(CSSNameMarginLeft, CSSPrimitiveValueCssPx, 5, "5px"),
		calculatedStyleTestLength(CSSNamePaddingLeft, CSSPrimitiveValueCssPx, 5, "5px"),
		calculatedStyleTestIdent(CSSNameBorderLeftStyle, IdentValueSolid),
		calculatedStyleTestLength(CSSNameBorderLeftWidth, CSSPrimitiveValueCssPx, 5, "5px"),
	))

	if got := row.GetMarginRect(100, ctx); got != RectPropertySetI(RectPropertySetAllZeros) {
		t.Errorf("GetMarginRect() = %s, want RectPropertySetAllZeros", got.ToString())
	}
	if got := row.GetPaddingRect(100, ctx); got != RectPropertySetI(RectPropertySetAllZeros) {
		t.Errorf("GetPaddingRect() = %s, want RectPropertySetAllZeros", got.ToString())
	}
	if got := row.GetBorder(ctx); got != BorderPropertySetEmptyBorder {
		t.Errorf("GetBorder() = %s, want BorderPropertySetEmptyBorder", got.ToString())
	}
	if !row.IsTableRow() || !row.IsUnderTableLayout() || row.IsFloated() {
		t.Errorf("IsTableRow=%v IsUnderTableLayout=%v IsFloated=%v", row.IsTableRow(), row.IsUnderTableLayout(), row.IsFloated())
	}
}

func TestCalculatedStyle_border(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	block := NewEmptyStyle().DeriveStyle(CascadedStyleCreateLayoutStyle(
		calculatedStyleTestIdent(CSSNameDisplay, IdentValueBlock),
		calculatedStyleTestIdent(CSSNameBorderLeftStyle, IdentValueSolid),
		calculatedStyleTestLength(CSSNameBorderLeftWidth, CSSPrimitiveValueCssPx, 5, "5px"),
		calculatedStyleTestLength(CSSNameBorderTopWidth, CSSPrimitiveValueCssPx, 5, "5px"),
		calculatedStyleTestDeclaration(CSSNameBorderLeftColor, NewPropertyValueFSColor(FSRGBColorRed)),
	))

	border := block.GetBorder(ctx)
	// border-top-style is none, so its width counts as 0
	if border.Left() != 100 || border.Top() != 0 {
		t.Errorf("border = %s, want left=100 top=0", border.ToString())
	}
	if border.LeftStyle() != IdentValueSolid || border.NoLeft() || !border.NoTop() {
		t.Errorf("LeftStyle=%v NoLeft=%v NoTop=%v", border.LeftStyle(), border.NoLeft(), border.NoTop())
	}
	if !FSRGBColorRed.Equals(border.LeftColor()) {
		t.Errorf("LeftColor() = %v, want red", border.LeftColor())
	}
	if border.HasBorderRadius() {
		t.Errorf("HasBorderRadius() = true, want false")
	}
	if block.GetBorder(ctx) != border {
		t.Errorf("GetBorder did not return the cached border")
	}
}
