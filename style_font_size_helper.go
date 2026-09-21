// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/FontSizeHelper.java

package ufo

// The Java maps are LinkedHashMaps whose key order GetNextSmaller and
// GetNextLarger observe; fontSizeHelperProportionalFontSizeOrder is the
// insertion order of the proportional map.
var fontSizeHelperProportionalFontSizeOrder = []*IdentValue{
	IdentValueXxSmall,
	IdentValueXSmall,
	IdentValueSmall,
	IdentValueMedium,
	IdentValueLarge,
	IdentValueXLarge,
	IdentValueXxLarge,
}

// XXX Should come from (or be influenced by) the UA.  These sizes
// correspond to the Firefox defaults
var fontSizeHelperProportionalFontSizes = map[*IdentValue]*PropertyValue{
	IdentValueXxSmall: NewPropertyValueFloat(CSSPrimitiveValueCssPx, 9.0, "9px"),
	IdentValueXSmall:  NewPropertyValueFloat(CSSPrimitiveValueCssPx, 10.0, "10px"),
	IdentValueSmall:   NewPropertyValueFloat(CSSPrimitiveValueCssPx, 13.0, "13px"),
	IdentValueMedium:  NewPropertyValueFloat(CSSPrimitiveValueCssPx, 16.0, "16px"),
	IdentValueLarge:   NewPropertyValueFloat(CSSPrimitiveValueCssPx, 18.0, "18px"),
	IdentValueXLarge:  NewPropertyValueFloat(CSSPrimitiveValueCssPx, 24.0, "24px"),
	IdentValueXxLarge: NewPropertyValueFloat(CSSPrimitiveValueCssPx, 32.0, "32px"),
}

var fontSizeHelperFixedFontSizes = map[*IdentValue]*PropertyValue{
	IdentValueXxSmall: NewPropertyValueFloat(CSSPrimitiveValueCssPx, 9.0, "9px"),
	IdentValueXSmall:  NewPropertyValueFloat(CSSPrimitiveValueCssPx, 10.0, "10px"),
	IdentValueSmall:   NewPropertyValueFloat(CSSPrimitiveValueCssPx, 12.0, "12px"),
	IdentValueMedium:  NewPropertyValueFloat(CSSPrimitiveValueCssPx, 13.0, "13px"),
	IdentValueLarge:   NewPropertyValueFloat(CSSPrimitiveValueCssPx, 16.0, "16px"),
	IdentValueXLarge:  NewPropertyValueFloat(CSSPrimitiveValueCssPx, 20.0, "20px"),
	IdentValueXxLarge: NewPropertyValueFloat(CSSPrimitiveValueCssPx, 26.0, "26px"),
}

var fontSizeHelperDefaultSmaller = NewPropertyValueFloat(CSSPrimitiveValueCssEms, 0.8, "0.8em")
var fontSizeHelperDefaultLarger = NewPropertyValueFloat(CSSPrimitiveValueCssEms, 1.2, "1.2em")

// FontSizeHelperGetNextSmaller returns nil when there is no smaller size.
func FontSizeHelperGetNextSmaller(absFontSize *IdentValue) *IdentValue {
	var prev *IdentValue
	for _, ident := range fontSizeHelperProportionalFontSizeOrder {
		if ident == absFontSize {
			return prev
		}
		prev = ident
	}
	return nil
}

// FontSizeHelperGetNextLarger returns nil when there is no larger size.
func FontSizeHelperGetNextLarger(absFontSize *IdentValue) *IdentValue {
	for i := 0; i < len(fontSizeHelperProportionalFontSizeOrder); {
		ident := fontSizeHelperProportionalFontSizeOrder[i]
		i++
		if ident == absFontSize && i < len(fontSizeHelperProportionalFontSizeOrder) {
			return fontSizeHelperProportionalFontSizeOrder[i]
		}
	}
	return nil
}

func FontSizeHelperResolveAbsoluteFontSize(fontSize *IdentValue, fontFamilies []string) *PropertyValue {
	monospace := fontSizeHelperIsMonospace(fontFamilies)

	if monospace {
		return fontSizeHelperFixedFontSizes[fontSize]
	} else {
		return fontSizeHelperProportionalFontSizes[fontSize]
	}
}

func FontSizeHelperGetDefaultRelativeFontSize(fontSize *IdentValue) *PropertyValue {
	if fontSize == IdentValueLarger {
		return fontSizeHelperDefaultLarger
	} else if fontSize == IdentValueSmaller {
		return fontSizeHelperDefaultSmaller
	} else {
		return nil
	}
}

func fontSizeHelperIsMonospace(fontFamilies []string) bool {
	for _, fontFamily := range fontFamilies {
		if fontFamily == "monospace" {
			return true
		}
	}
	return false
}
