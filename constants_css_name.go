// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/constants/CSSName.java

package ufo

import (
	"sort"
	"sync"
)

// A CSSName is a Singleton representing a single CSS property name, like
// border-width. The class declares a Singleton static instance for each CSS
// Level 2 property. A CSSName instance has the property name available from
// the ToString method, as well as a unique (among all CSSName instances)
// integer id ranging from 0...n instances, incremented by 1, available using
// the public field FS_ID (e.g. CSSNameColor.FS_ID).
//
// Initialization order. The Java class runs three steps in its static
// initializer, in this order: (1) one addProperty call per property, each
// with a new property builder; (2) filling ALL_PROPERTIES; (3) parsing the
// initial value of every implemented primitive property with a CSSParser and
// deriving a value from it with DerivedValueFactory. Go initializes
// package-level variables in an order it derives from their dependencies and
// runs init functions after that, so the steps are arranged as follows.
//
//   - Step 1 without the builders happens in the initializers of the
//     CSSName... variables below. cssNameAddProperty depends on nothing
//     outside this file, so the variables are initialized in declaration
//     order, which is the Java order, and every FS_ID is the same as in Java.
//     Package-level variables of other files that refer to a CSSName (the
//     CSSName arrays of the shorthand property builders) get the pointer,
//     with propName, initialValue, the flags and FS_ID set. Step 2 happens in
//     the same call: the property is appended to cssNameAllProperties, where
//     its index is its FS_ID.
//   - The builders are attached by the init function of this file, in the
//     Java order. Were the builder constructors called from the variable
//     initializers, a constructor that refers to a CSSName or to anything
//     initialized from one would reorder the initialization and change the
//     FS_ID values, or fail to compile with an initialization cycle.
//   - Step 3 runs on the first call of InitialDerivedValue, once for all
//     properties, in the Java order (sorted by property name). It needs the
//     CSS parser, the property builders, StylesheetInfo and
//     DerivedValueFactory to be initialized, which is not the case while this
//     file's package-level variables or init function run.
type CSSName struct {
	// The CSS 2 property name, e.g. "border"
	propName string

	// A (String) initial value from the CSS 2.1 specification
	initialValue string

	// True if the property inherits by default, false if not inherited
	propertyInherits bool

	initialDerivedValue FSDerivedValue

	implemented bool

	// builder may be nil.
	builder PropertyBuilder

	// Unique integer id for a CSSName.
	FS_ID int
}

// marker values, used for initialization
const (
	cssNamePrimitive    = 0
	cssNameShorthand    = 1
	cssNameInherits     = 2
	cssNameNotInherited = 3
)

var (
	// Used to assign unique int id values to new CSSNames created in this class
	cssNameMaxAssigned int

	// All CSS properties, indexed by FS_ID
	cssNameAllProperties []*CSSName

	// Map of all CSS properties. A TreeMap in Java; the functions that
	// iterate over it sort the keys on read.
	cssNameAllPropertyNames = map[string]*CSSName{}

	// Map of all non-shorthand CSS properties. A TreeMap in Java; the
	// functions that iterate over it sort the keys on read.
	cssNameAllPrimitivePropertyNames = map[string]*CSSName{}

	cssNameInitialDerivedValuesOnce sync.Once
)

var (
	// Unique CSSName instance for CSS2 property.
	// TODO: UA dependent
	CSSNameColor = cssNameAddProperty("color", cssNamePrimitive, "black", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameBackgroundColor = cssNameAddProperty("background-color", cssNamePrimitive, "transparent", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBackgroundImage = cssNameAddProperty("background-image", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBackgroundRepeat = cssNameAddProperty("background-repeat", cssNamePrimitive, "repeat", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBackgroundAttachment = cssNameAddProperty("background-attachment", cssNamePrimitive, "scroll", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBackgroundPosition = cssNameAddProperty("background-position", cssNamePrimitive, "0% 0%", cssNameNotInherited)

	CSSNameBackgroundSize = cssNameAddProperty("background-size", cssNamePrimitive, "auto auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderCollapse = cssNameAddProperty("border-collapse", cssNamePrimitive, "separate", cssNameInherits)

	// Unique CSSName instance for fictitious property.
	CSSNameFsBorderSpacingHorizontal = cssNameAddProperty("-fs-border-spacing-horizontal", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsBorderSpacingVertical = cssNameAddProperty("-fs-border-spacing-vertical", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsDynamicAutoWidth = cssNameAddProperty("-fs-dynamic-auto-width", cssNamePrimitive, "static", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsFontMetricSrc = cssNameAddProperty("-fs-font-metric-src", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsKeepWithInline = cssNameAddProperty("-fs-keep-with-inline", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsPageWidth = cssNameAddProperty("-fs-page-width", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsPageHeight = cssNameAddProperty("-fs-page-height", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsPageSequence = cssNameAddProperty("-fs-page-sequence", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsPdfFontEmbed = cssNameAddProperty("-fs-pdf-font-embed", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsPdfFontEncoding = cssNameAddProperty("-fs-pdf-font-encoding", cssNamePrimitive, "Cp1252", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsPageOrientation = cssNameAddProperty("-fs-page-orientation", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsTablePaginate = cssNameAddProperty("-fs-table-paginate", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsTextDecorationExtent = cssNameAddProperty("-fs-text-decoration-extent", cssNamePrimitive, "line", cssNameNotInherited)

	// Used for forcing images to scale to a certain width
	CSSNameFsFitImagesToWidth = cssNameAddProperty("-fs-fit-images-to-width", cssNamePrimitive, "auto", cssNameNotInherited)

	// Used to control creation of named destinations for boxes having the id attribute set.
	CSSNameFsNamedDestination = cssNameAddProperty("-fs-named-destination", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBottom = cssNameAddProperty("bottom", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameCaptionSide = cssNameAddProperty("caption-side", cssNamePrimitive, "top", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameClear = cssNameAddProperty("clear", cssNamePrimitive, "none", cssNameNotInherited)

	CSSNameColumnCount = cssNameAddProperty("column-count", cssNamePrimitive, "auto", cssNameNotInherited)

	CSSNameColumnGap = cssNameAddProperty("column-gap", cssNamePrimitive, "normal", cssNameNotInherited)

	CSSNameColumnWidth = cssNameAddProperty("column-width", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameClip = cssNameAddPropertyWithImplemented("clip", cssNamePrimitive, "auto", cssNameNotInherited, false)

	// Unique CSSName instance for CSS2 property.
	CSSNameContent = cssNameAddProperty("content", cssNamePrimitive, "normal", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameCounterIncrement = cssNameAddPropertyWithImplemented("counter-increment", cssNamePrimitive, "none", cssNameNotInherited, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameCounterReset = cssNameAddPropertyWithImplemented("counter-reset", cssNamePrimitive, "none", cssNameNotInherited, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameCursor = cssNameAddPropertyWithImplemented("cursor", cssNamePrimitive, "auto", cssNameInherits, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameDirection = cssNameAddPropertyWithImplemented("direction", cssNamePrimitive, "ltr", cssNameInherits, false)

	// Unique CSSName instance for CSS2 property.
	CSSNameDisplay = cssNameAddProperty("display", cssNamePrimitive, "inline", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameEmptyCells = cssNameAddProperty("empty-cells", cssNamePrimitive, "show", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameFloat = cssNameAddProperty("float", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFontStyle = cssNameAddProperty("font-style", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameFontVariant = cssNameAddProperty("font-variant", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameFontWeight = cssNameAddProperty("font-weight", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameFontSize = cssNameAddProperty("font-size", cssNamePrimitive, "medium", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameLineHeight = cssNameAddProperty("line-height", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	// TODO: UA dependent
	CSSNameFontFamily = cssNameAddProperty("font-family", cssNamePrimitive, "serif", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsColspan = cssNameAddProperty("-fs-table-cell-colspan", cssNamePrimitive, "1", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFsRowspan = cssNameAddProperty("-fs-table-cell-rowspan", cssNamePrimitive, "1", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameHeight = cssNameAddProperty("height", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameLeft = cssNameAddProperty("left", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameLetterSpacing = cssNameAddPropertyWithImplemented("letter-spacing", cssNamePrimitive, "normal", cssNameInherits, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameListStyleType = cssNameAddProperty("list-style-type", cssNamePrimitive, "disc", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameListStylePosition = cssNameAddProperty("list-style-position", cssNamePrimitive, "outside", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameListStyleImage = cssNameAddProperty("list-style-image", cssNamePrimitive, "none", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameMaxHeight = cssNameAddProperty("max-height", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameMaxWidth = cssNameAddProperty("max-width", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameMinHeight = cssNameAddProperty("min-height", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	// TODO: UA dependent
	CSSNameMinWidth = cssNameAddProperty("min-width", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameOrphans = cssNameAddPropertyWithImplemented("orphans", cssNamePrimitive, "2", cssNameInherits, true)

	// The inherit flag carries the Java comment "PR22 - INHERITS".
	CSSNameOpacity = cssNameAddPropertyWithImplemented("opacity", cssNamePrimitive, "1", cssNameNotInherited, true)

	// Unique CSSName instance for CSS2 property.
	// Java: /* "invert", */ "black"  XXX Wrong (but doesn't matter for now)
	CSSNameOutlineColor = cssNameAddPropertyWithImplemented("outline-color", cssNamePrimitive, "black", cssNameNotInherited, false)

	// Unique CSSName instance for CSS2 property.
	CSSNameOutlineStyle = cssNameAddPropertyWithImplemented("outline-style", cssNamePrimitive, "none", cssNameNotInherited, false)

	// Unique CSSName instance for CSS2 property.
	CSSNameOutlineWidth = cssNameAddPropertyWithImplemented("outline-width", cssNamePrimitive, "medium", cssNameNotInherited, false)

	// Unique CSSName instance for CSS2 property.
	CSSNameOverflow = cssNameAddProperty("overflow", cssNamePrimitive, "visible", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePage = cssNameAddProperty("page", cssNamePrimitive, "auto", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNamePageBreakAfter = cssNameAddProperty("page-break-after", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePageBreakBefore = cssNameAddProperty("page-break-before", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePageBreakInside = cssNameAddProperty("page-break-inside", cssNamePrimitive, "auto", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNamePosition = cssNameAddProperty("position", cssNamePrimitive, "static", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	// TODO: UA dependent
	CSSNameQuotes = cssNameAddProperty("quotes", cssNamePrimitive, "none", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameRight = cssNameAddProperty("right", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameSrc = cssNameAddProperty("src", cssNamePrimitive, "none", cssNameNotInherited)

	// Used for controlling tab size in pre tags. See <a href="http://dev.w3.org/csswg/css3-text/#tab-size">...</a>
	CSSNameTabSize = cssNameAddProperty("tab-size", cssNamePrimitive, "8", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameTableLayout = cssNameAddProperty("table-layout", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	// TODO: UA dependent
	CSSNameTextAlign = cssNameAddProperty("text-align", cssNamePrimitive, "left", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameTextDecoration = cssNameAddProperty("text-decoration", cssNamePrimitive, "none", cssNameNotInherited)

	// CSS text-underline-position property.
	CSSNameTextUnderlinePosition = cssNameAddProperty("text-underline-position", cssNamePrimitive, "auto", cssNameInherits)

	// CSS text-underline-offset property.
	CSSNameTextUnderlineOffset = cssNameAddProperty("text-underline-offset", cssNamePrimitive, "auto", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameTextIndent = cssNameAddProperty("text-indent", cssNamePrimitive, "0", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameTextTransform = cssNameAddProperty("text-transform", cssNamePrimitive, "none", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameTop = cssNameAddProperty("top", cssNamePrimitive, "auto", cssNameNotInherited)

	CSSNameTransform = cssNameAddProperty("transform", cssNamePrimitive, "none", cssNameNotInherited)

	CSSNameTransformOrigin = cssNameAddProperty("transform-origin", cssNamePrimitive, "50% 50%", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameUnicodeBidi = cssNameAddPropertyWithImplemented("unicode-bidi", cssNamePrimitive, "normal", cssNameNotInherited, false)

	// Unique CSSName instance for CSS2 property.
	CSSNameVerticalAlign = cssNameAddProperty("vertical-align", cssNamePrimitive, "baseline", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameVisibility = cssNameAddProperty("visibility", cssNamePrimitive, "visible", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameWhiteSpace = cssNameAddProperty("white-space", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS3 property.
	CSSNameWordBreak = cssNameAddProperty("word-break", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS3 property.
	CSSNameWordWrap = cssNameAddProperty("word-wrap", cssNamePrimitive, "normal", cssNameInherits)

	// Unique CSSName instance for CSS3 property.
	CSSNameHyphens = cssNameAddProperty("hyphens", cssNamePrimitive, "none", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameWidows = cssNameAddPropertyWithImplemented("widows", cssNamePrimitive, "2", cssNameInherits, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameWidth = cssNameAddProperty("width", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameWordSpacing = cssNameAddPropertyWithImplemented("word-spacing", cssNamePrimitive, "normal", cssNameInherits, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameZIndex = cssNameAddProperty("z-index", cssNamePrimitive, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderTopColor = cssNameAddProperty("border-top-color", cssNamePrimitive, "=color", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderRightColor = cssNameAddProperty("border-right-color", cssNamePrimitive, "=color", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderBottomColor = cssNameAddProperty("border-bottom-color", cssNamePrimitive, "=color", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderLeftColor = cssNameAddProperty("border-left-color", cssNamePrimitive, "=color", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderTopStyle = cssNameAddProperty("border-top-style", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderRightStyle = cssNameAddProperty("border-right-style", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderBottomStyle = cssNameAddProperty("border-bottom-style", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderLeftStyle = cssNameAddProperty("border-left-style", cssNamePrimitive, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderTopWidth = cssNameAddProperty("border-top-width", cssNamePrimitive, "medium", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderRightWidth = cssNameAddProperty("border-right-width", cssNamePrimitive, "medium", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderBottomWidth = cssNameAddProperty("border-bottom-width", cssNamePrimitive, "medium", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderLeftWidth = cssNameAddProperty("border-left-width", cssNamePrimitive, "medium", cssNameNotInherited)

	// Unique CSSName instance for CSS3 property.
	CSSNameBorderTopLeftRadius = cssNameAddPropertyWithImplemented("border-top-left-radius", cssNamePrimitive, "0 0", cssNameNotInherited, true)

	// Unique CSSName instance for CSS3 property.
	CSSNameBorderTopRightRadius = cssNameAddPropertyWithImplemented("border-top-right-radius", cssNamePrimitive, "0 0", cssNameNotInherited, true)

	// Unique CSSName instance for CSS3 property.
	CSSNameBorderBottomRightRadius = cssNameAddPropertyWithImplemented("border-bottom-right-radius", cssNamePrimitive, "0 0", cssNameNotInherited, true)

	// Unique CSSName instance for CSS3 property.
	CSSNameBorderBottomLeftRadius = cssNameAddPropertyWithImplemented("border-bottom-left-radius", cssNamePrimitive, "0 0", cssNameNotInherited, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameMarginTop = cssNameAddProperty("margin-top", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameMarginRight = cssNameAddProperty("margin-right", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameMarginBottom = cssNameAddProperty("margin-bottom", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameMarginLeft = cssNameAddProperty("margin-left", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePaddingTop = cssNameAddProperty("padding-top", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePaddingRight = cssNameAddProperty("padding-right", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePaddingBottom = cssNameAddProperty("padding-bottom", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNamePaddingLeft = cssNameAddProperty("padding-left", cssNamePrimitive, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBackgroundShorthand = cssNameAddProperty("background", cssNameShorthand, "transparent none repeat scroll 0% 0%", cssNameNotInherited)

	// Unique CSSName instance for CSS3 property.
	CSSNameBorderRadiusShorthand = cssNameAddPropertyWithImplemented("border-radius", cssNameShorthand, "0px", cssNameNotInherited, true)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderWidthShorthand = cssNameAddProperty("border-width", cssNameShorthand, "medium", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderStyleShorthand = cssNameAddProperty("border-style", cssNameShorthand, "none", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderShorthand = cssNameAddProperty("border", cssNameShorthand, "medium none black", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderTopShorthand = cssNameAddProperty("border-top", cssNameShorthand, "medium none black", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderRightShorthand = cssNameAddProperty("border-right", cssNameShorthand, "medium none black", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderBottomShorthand = cssNameAddProperty("border-bottom", cssNameShorthand, "medium none black", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderLeftShorthand = cssNameAddProperty("border-left", cssNameShorthand, "medium none black", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderColorShorthand = cssNameAddProperty("border-color", cssNameShorthand, "black", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBorderSpacing = cssNameAddProperty("border-spacing", cssNameShorthand, "0", cssNameInherits)

	CSSNameColumnsShorthand = cssNameAddProperty("columns", cssNameShorthand, "auto auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameFontShorthand = cssNameAddProperty("font", cssNameShorthand, "", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameListStyleShorthand = cssNameAddProperty("list-style", cssNameShorthand, "disc outside none", cssNameInherits)

	// Unique CSSName instance for CSS2 property.
	CSSNameMarginShorthand = cssNameAddProperty("margin", cssNameShorthand, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameOutlineShorthand = cssNameAddPropertyWithImplemented("outline", cssNameShorthand, "invert none medium", cssNameNotInherited, false)

	// Unique CSSName instance for CSS2 property.
	CSSNamePaddingShorthand = cssNameAddProperty("padding", cssNameShorthand, "0", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameSizeShorthand = cssNameAddProperty("size", cssNameShorthand, "auto", cssNameNotInherited)

	// Unique CSSName instance for CSS2 property.
	CSSNameBoxSizing = cssNameAddProperty("box-sizing", cssNamePrimitive, "content-box", cssNameNotInherited)
)

// init attaches the property builders, in the order of the Java declarations.
// See the comment on CSSName for why this is not done in the variable
// initializers.
func init() {
	CSSNameColor.builder = NewPrimitivePropertyBuildersColor()
	CSSNameBackgroundColor.builder = NewPrimitivePropertyBuildersBackgroundColor()
	CSSNameBackgroundImage.builder = NewPrimitivePropertyBuildersBackgroundImage()
	CSSNameBackgroundRepeat.builder = NewPrimitivePropertyBuildersBackgroundRepeat()
	CSSNameBackgroundAttachment.builder = NewPrimitivePropertyBuildersBackgroundAttachment()
	CSSNameBackgroundPosition.builder = NewPrimitivePropertyBuildersBackgroundPosition()
	CSSNameBackgroundSize.builder = NewPrimitivePropertyBuildersBackgroundSize()
	CSSNameBorderCollapse.builder = NewPrimitivePropertyBuildersBorderCollapse()
	CSSNameFsBorderSpacingHorizontal.builder = NewPrimitivePropertyBuildersFSBorderSpacingHorizontal()
	CSSNameFsBorderSpacingVertical.builder = NewPrimitivePropertyBuildersFSBorderSpacingVertical()
	CSSNameFsDynamicAutoWidth.builder = NewPrimitivePropertyBuildersFSDynamicAutoWidth()
	CSSNameFsFontMetricSrc.builder = NewPrimitivePropertyBuildersFSFontMetricSrc()
	CSSNameFsKeepWithInline.builder = NewPrimitivePropertyBuildersFSKeepWithInline()
	CSSNameFsPageWidth.builder = NewPrimitivePropertyBuildersFSPageWidth()
	CSSNameFsPageHeight.builder = NewPrimitivePropertyBuildersFSPageHeight()
	CSSNameFsPageSequence.builder = NewPrimitivePropertyBuildersFSPageSequence()
	CSSNameFsPdfFontEmbed.builder = NewPrimitivePropertyBuildersFSPDFFontEmbed()
	CSSNameFsPdfFontEncoding.builder = NewPrimitivePropertyBuildersFSPDFFontEncoding()
	CSSNameFsPageOrientation.builder = NewPrimitivePropertyBuildersFSPageOrientation()
	CSSNameFsTablePaginate.builder = NewPrimitivePropertyBuildersFSTablePaginate()
	CSSNameFsTextDecorationExtent.builder = NewPrimitivePropertyBuildersFSTextDecorationExtent()
	CSSNameFsFitImagesToWidth.builder = NewPrimitivePropertyBuildersFSFitImagesToWidth()
	CSSNameFsNamedDestination.builder = NewPrimitivePropertyBuildersFSNamedDestination()
	CSSNameBottom.builder = NewPrimitivePropertyBuildersBottom()
	CSSNameCaptionSide.builder = NewPrimitivePropertyBuildersCaptionSide()
	CSSNameClear.builder = NewPrimitivePropertyBuildersClear()
	CSSNameColumnCount.builder = NewPrimitivePropertyBuildersColumnCount()
	CSSNameColumnGap.builder = NewPrimitivePropertyBuildersColumnGap()
	CSSNameColumnWidth.builder = NewPrimitivePropertyBuildersColumnWidth()
	// CSSNameClip has no builder.
	CSSNameContent.builder = NewContentPropertyBuilder()
	CSSNameCounterIncrement.builder = NewCounterPropertyBuilderCounterIncrement()
	CSSNameCounterReset.builder = NewCounterPropertyBuilderCounterReset()
	CSSNameCursor.builder = NewPrimitivePropertyBuildersCursor()
	// CSSNameDirection has no builder.
	CSSNameDisplay.builder = NewPrimitivePropertyBuildersDisplay()
	CSSNameEmptyCells.builder = NewPrimitivePropertyBuildersEmptyCells()
	CSSNameFloat.builder = NewPrimitivePropertyBuildersFloat()
	CSSNameFontStyle.builder = NewPrimitivePropertyBuildersFontStyle()
	CSSNameFontVariant.builder = NewPrimitivePropertyBuildersFontVariant()
	CSSNameFontWeight.builder = NewPrimitivePropertyBuildersFontWeight()
	CSSNameFontSize.builder = NewPrimitivePropertyBuildersFontSize()
	CSSNameLineHeight.builder = NewPrimitivePropertyBuildersLineHeight()
	CSSNameFontFamily.builder = NewPrimitivePropertyBuildersFontFamily()
	CSSNameFsColspan.builder = NewPrimitivePropertyBuildersFSTableCellColspan()
	CSSNameFsRowspan.builder = NewPrimitivePropertyBuildersFSTableCellRowspan()
	CSSNameHeight.builder = NewPrimitivePropertyBuildersHeight()
	CSSNameLeft.builder = NewPrimitivePropertyBuildersLeft()
	CSSNameLetterSpacing.builder = NewPrimitivePropertyBuildersLetterSpacing()
	CSSNameListStyleType.builder = NewPrimitivePropertyBuildersListStyleType()
	CSSNameListStylePosition.builder = NewPrimitivePropertyBuildersListStylePosition()
	CSSNameListStyleImage.builder = NewPrimitivePropertyBuildersListStyleImage()
	CSSNameMaxHeight.builder = NewPrimitivePropertyBuildersMaxHeight()
	CSSNameMaxWidth.builder = NewPrimitivePropertyBuildersMaxWidth()
	CSSNameMinHeight.builder = NewPrimitivePropertyBuildersMinHeight()
	CSSNameMinWidth.builder = NewPrimitivePropertyBuildersMinWidth()
	CSSNameOrphans.builder = NewPrimitivePropertyBuildersOrphans()
	CSSNameOpacity.builder = NewPrimitivePropertyBuildersOpacity()
	// CSSNameOutlineColor has no builder.
	// CSSNameOutlineStyle has no builder.
	// CSSNameOutlineWidth has no builder.
	CSSNameOverflow.builder = NewPrimitivePropertyBuildersOverflow()
	CSSNamePage.builder = NewPrimitivePropertyBuildersPage()
	CSSNamePageBreakAfter.builder = NewPrimitivePropertyBuildersPageBreakAfter()
	CSSNamePageBreakBefore.builder = NewPrimitivePropertyBuildersPageBreakBefore()
	CSSNamePageBreakInside.builder = NewPrimitivePropertyBuildersPageBreakInside()
	CSSNamePosition.builder = NewPrimitivePropertyBuildersPosition()
	CSSNameQuotes.builder = NewQuotesPropertyBuilder()
	CSSNameRight.builder = NewPrimitivePropertyBuildersRight()
	CSSNameSrc.builder = NewPrimitivePropertyBuildersSrc()
	CSSNameTabSize.builder = NewPrimitivePropertyBuildersTabSize()
	CSSNameTableLayout.builder = NewPrimitivePropertyBuildersTableLayout()
	CSSNameTextAlign.builder = NewPrimitivePropertyBuildersTextAlign()
	CSSNameTextDecoration.builder = NewPrimitivePropertyBuildersTextDecoration()
	CSSNameTextUnderlinePosition.builder = NewPrimitivePropertyBuildersTextUnderlinePosition()
	CSSNameTextUnderlineOffset.builder = NewPrimitivePropertyBuildersTextUnderlineOffset()
	CSSNameTextIndent.builder = NewPrimitivePropertyBuildersTextIndent()
	CSSNameTextTransform.builder = NewPrimitivePropertyBuildersTextTransform()
	CSSNameTop.builder = NewPrimitivePropertyBuildersTop()
	CSSNameTransform.builder = NewTransformPropertyBuilder()
	CSSNameTransformOrigin.builder = NewTransformOriginPropertyBuilder()
	// CSSNameUnicodeBidi has no builder.
	CSSNameVerticalAlign.builder = NewPrimitivePropertyBuildersVerticalAlign()
	CSSNameVisibility.builder = NewPrimitivePropertyBuildersVisibility()
	CSSNameWhiteSpace.builder = NewPrimitivePropertyBuildersWhiteSpace()
	CSSNameWordBreak.builder = NewPrimitivePropertyBuildersWordBreak()
	CSSNameWordWrap.builder = NewPrimitivePropertyBuildersWordWrap()
	CSSNameHyphens.builder = NewPrimitivePropertyBuildersHyphens()
	CSSNameWidows.builder = NewPrimitivePropertyBuildersWidows()
	CSSNameWidth.builder = NewPrimitivePropertyBuildersWidth()
	CSSNameWordSpacing.builder = NewPrimitivePropertyBuildersWordSpacing()
	CSSNameZIndex.builder = NewPrimitivePropertyBuildersZIndex()
	CSSNameBorderTopColor.builder = NewPrimitivePropertyBuildersBorderTopColor()
	CSSNameBorderRightColor.builder = NewPrimitivePropertyBuildersBorderLeftColor()
	CSSNameBorderBottomColor.builder = NewPrimitivePropertyBuildersBorderBottomColor()
	CSSNameBorderLeftColor.builder = NewPrimitivePropertyBuildersBorderLeftColor()
	CSSNameBorderTopStyle.builder = NewPrimitivePropertyBuildersBorderTopStyle()
	CSSNameBorderRightStyle.builder = NewPrimitivePropertyBuildersBorderRightStyle()
	CSSNameBorderBottomStyle.builder = NewPrimitivePropertyBuildersBorderBottomStyle()
	CSSNameBorderLeftStyle.builder = NewPrimitivePropertyBuildersBorderLeftStyle()
	CSSNameBorderTopWidth.builder = NewPrimitivePropertyBuildersBorderTopWidth()
	CSSNameBorderRightWidth.builder = NewPrimitivePropertyBuildersBorderRightWidth()
	CSSNameBorderBottomWidth.builder = NewPrimitivePropertyBuildersBorderBottomWidth()
	CSSNameBorderLeftWidth.builder = NewPrimitivePropertyBuildersBorderLeftWidth()
	CSSNameBorderTopLeftRadius.builder = NewPrimitivePropertyBuildersBorderTopLeftRadius()
	CSSNameBorderTopRightRadius.builder = NewPrimitivePropertyBuildersBorderTopRightRadius()
	CSSNameBorderBottomRightRadius.builder = NewPrimitivePropertyBuildersBorderBottomRightRadius()
	CSSNameBorderBottomLeftRadius.builder = NewPrimitivePropertyBuildersBorderBottomLeftRadius()
	CSSNameMarginTop.builder = NewPrimitivePropertyBuildersMarginTop()
	CSSNameMarginRight.builder = NewPrimitivePropertyBuildersMarginRight()
	CSSNameMarginBottom.builder = NewPrimitivePropertyBuildersMarginBottom()
	CSSNameMarginLeft.builder = NewPrimitivePropertyBuildersMarginLeft()
	CSSNamePaddingTop.builder = NewPrimitivePropertyBuildersPaddingTop()
	CSSNamePaddingRight.builder = NewPrimitivePropertyBuildersPaddingRight()
	CSSNamePaddingBottom.builder = NewPrimitivePropertyBuildersPaddingBottom()
	CSSNamePaddingLeft.builder = NewPrimitivePropertyBuildersPaddingLeft()
	CSSNameBackgroundShorthand.builder = NewBackgroundPropertyBuilder()
	CSSNameBorderRadiusShorthand.builder = NewOneToFourPropertyBuildersBorderRadius()
	CSSNameBorderWidthShorthand.builder = NewOneToFourPropertyBuildersBorderWidth()
	CSSNameBorderStyleShorthand.builder = NewOneToFourPropertyBuildersBorderStyle()
	CSSNameBorderShorthand.builder = NewBorderPropertyBuildersBorder()
	CSSNameBorderTopShorthand.builder = NewBorderPropertyBuildersBorderTop()
	CSSNameBorderRightShorthand.builder = NewBorderPropertyBuildersBorderRight()
	CSSNameBorderBottomShorthand.builder = NewBorderPropertyBuildersBorderBottom()
	CSSNameBorderLeftShorthand.builder = NewBorderPropertyBuildersBorderLeft()
	CSSNameBorderColorShorthand.builder = NewOneToFourPropertyBuildersBorderColor()
	CSSNameBorderSpacing.builder = NewBorderSpacingPropertyBuilder()
	CSSNameColumnsShorthand.builder = NewColumnsPropertyBuilder()
	CSSNameFontShorthand.builder = NewFontPropertyBuilder()
	CSSNameListStyleShorthand.builder = NewListStylePropertyBuilder()
	CSSNameMarginShorthand.builder = NewOneToFourPropertyBuildersMargin()
	// CSSNameOutlineShorthand has no builder.
	CSSNamePaddingShorthand.builder = NewOneToFourPropertyBuildersPadding()
	CSSNameSizeShorthand.builder = NewSizePropertyBuilder()
	CSSNameBoxSizing.builder = NewPrimitivePropertyBuildersBoxSizing()
}

var CSSNameMarginSideProperties = NewCSSNameCSSSideProperties(
	CSSNameMarginTop,
	CSSNameMarginRight,
	CSSNameMarginBottom,
	CSSNameMarginLeft)

var CSSNamePaddingSideProperties = NewCSSNameCSSSideProperties(
	CSSNamePaddingTop,
	CSSNamePaddingRight,
	CSSNamePaddingBottom,
	CSSNamePaddingLeft)

var CSSNameBorderSideProperties = NewCSSNameCSSSideProperties(
	CSSNameBorderTopWidth,
	CSSNameBorderRightWidth,
	CSSNameBorderBottomWidth,
	CSSNameBorderLeftWidth)

var CSSNameBorderStyleProperties = NewCSSNameCSSSideProperties(
	CSSNameBorderTopStyle,
	CSSNameBorderRightStyle,
	CSSNameBorderBottomStyle,
	CSSNameBorderLeftStyle)

var CSSNameBorderColorProperties = NewCSSNameCSSSideProperties(
	CSSNameBorderTopColor,
	CSSNameBorderRightColor,
	CSSNameBorderBottomColor,
	CSSNameBorderLeftColor)

func newCSSName(
	propName string, initialValue string, inherits bool,
	implemented bool, builder PropertyBuilder) *CSSName {
	c := &CSSName{}
	c.propName = propName
	c.FS_ID = cssNameMaxAssigned
	cssNameMaxAssigned++
	c.initialValue = initialValue
	c.propertyInherits = inherits
	c.implemented = implemented
	c.builder = builder
	return c
}

// ToString returns a string representation of the object, in this case,
// always the full CSS property name in lowercase.
func (c *CSSName) ToString() string {
	return c.propName
}

func (c *CSSName) String() string {
	return c.propName
}

// CSSNameCountCSSNames returns a count of all CSS properties known to this
// class, shorthand and primitive.
func CSSNameCountCSSNames() int {
	return cssNameMaxAssigned
}

// CSSNameCountCSSPrimitiveNames returns a count of all CSS primitive
// (non-shorthand) properties known to this class.
func CSSNameCountCSSPrimitiveNames() int {
	return len(cssNameAllPrimitivePropertyNames)
}

// CSSNameAllCSS2PropertyNames returns ALL CSS 2 visual property names, sorted
// (the Java method returns an iterator over the keys of a TreeMap).
func CSSNameAllCSS2PropertyNames() []string {
	return cssNameSortedKeys(cssNameAllPropertyNames)
}

// CSSNameAllCSS2PrimitivePropertyNames returns ALL primitive (non-shorthand)
// CSS 2 visual property names, sorted (the Java method returns an iterator
// over the keys of a TreeMap).
func CSSNameAllCSS2PrimitivePropertyNames() []string {
	return cssNameSortedKeys(cssNameAllPrimitivePropertyNames)
}

func cssNameSortedKeys(m map[string]*CSSName) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// CSSNamePropertyInherits returns true if the named property inherits by
// default, according to the CSS2 spec.
// CLEAN: method is now unnecessary
func CSSNamePropertyInherits(cssName *CSSName) bool {
	return cssName.propertyInherits
}

// CSSNameInitialValue returns the initial value of the named property,
// according to the CSS2 spec, as a String. Casting must be taken care of by
// the caller, as there is too much variation in value-types.
// CLEAN: method is now unnecessary
func CSSNameInitialValue(cssName *CSSName) string {
	return cssName.initialValue
}

// InitialDerivedValue may return nil: for a shorthand property, for a
// property that is not implemented, and for a property whose initial value
// starts with "=".
func (c *CSSName) InitialDerivedValue() FSDerivedValue {
	cssNameInitialDerivedValuesOnce.Do(cssNameInitInitialDerivedValues)
	return c.initialDerivedValue
}

func CSSNameIsImplemented(cssName *CSSName) bool {
	return cssName.implemented
}

// CSSNameGetPropertyBuilder may return nil.
func CSSNameGetPropertyBuilder(cssName *CSSName) PropertyBuilder {
	return cssName.builder
}

// CSSNameGetByPropertyName gets the byPropertyName attribute of the CSSName
// class. It may return nil.
func CSSNameGetByPropertyName(propName string) *CSSName {
	return cssNameAllPropertyNames[propName]
}

func CSSNameCssProperty(propName string) *CSSName {
	cssName := CSSNameGetByPropertyName(propName)
	if cssName == nil {
		panic(NewXRRuntimeException("Unknown CSS property: " + propName))
	}
	return cssName
}

func CSSNameGetByID(id int) *CSSName {
	return cssNameAllProperties[id]
}

// cssNameAddProperty ports addProperty(propName, type, initialValue, inherit,
// builder) without the builder, which the init function of this file
// attaches.
func cssNameAddProperty(
	propName string,
	propType int,
	initialValue string,
	inherit int,
) *CSSName {
	return cssNameAddPropertyWithImplemented(propName, propType, initialValue, inherit, true)
}

// cssNameAddPropertyWithImplemented adds a feature to the Property attribute
// of the CSSName class. It ports addProperty(propName, type, initialValue,
// inherit, implemented, builder) without the builder, which the init function
// of this file attaches.
//
// propName is the feature to be added to the Property attribute.
func cssNameAddPropertyWithImplemented(
	propName string,
	propType int,
	initialValue string,
	inherit int,
	implemented bool,
) *CSSName {
	cssName := newCSSName(
		propName, initialValue, inherit == cssNameInherits, implemented, nil)

	cssNameAllPropertyNames[propName] = cssName

	if propType == cssNamePrimitive {
		cssNameAllPrimitivePropertyNames[propName] = cssName
	}

	// Java fills ALL_PROPERTIES from ALL_PROPERTY_NAMES in a static block that
	// runs after the last addProperty call. The properties are created in
	// FS_ID order, so appending here gives the same array.
	cssNameAllProperties = append(cssNameAllProperties, cssName)

	return cssName
}

// cssNameInitInitialDerivedValues ports the last static block of the Java
// class. It runs on the first call of InitialDerivedValue; see the comment on
// CSSName.
func cssNameInitInitialDerivedValues() {
	parser := NewCSSParser(CSSErrorHandlerFunc(func(uri string, message string) {
		XRLogCssParse("(" + uri + ") " + message)
	}))
	for _, propName := range cssNameSortedKeys(cssNameAllPrimitivePropertyNames) {
		cssName := cssNameAllPrimitivePropertyNames[propName]
		if cssName.initialValue[0] != '=' && cssName.implemented {
			value := parser.ParsePropertyValue(
				cssName, StylesheetInfoOriginUserAgent, cssName.initialValue)

			if value == nil {
				XRLogException("Unable to derive initial value for " + cssName.ToString())
			} else {
				cssName.initialDerivedValue = DerivedValueFactoryNewDerivedValue(
					nil,
					cssName,
					value)
			}
		}
	}
}

// CompareTo is assumed to be consistent with Equals because CSSName is in
// essence an enum.
func (c *CSSName) CompareTo(object *CSSName) int {
	if object == nil {
		panic(NewXRRuntimeException("Cannot compare " + c.ToString() + " to null"))
	}
	return c.FS_ID - object.FS_ID
}

// FIXME equals, hashcode
func (c *CSSName) Equals(o any) bool {
	cssName, ok := o.(*CSSName)
	if !ok || cssName == nil {
		return false
	}
	if c == cssName {
		return true
	}

	return c.FS_ID == cssName.FS_ID
}

func (c *CSSName) HashCode() int {
	return c.FS_ID
}

// CSSNameCSSSideProperties ports the Java record CSSName.CSSSideProperties.
type CSSNameCSSSideProperties struct {
	top    *CSSName
	right  *CSSName
	bottom *CSSName
	left   *CSSName
}

func NewCSSNameCSSSideProperties(top *CSSName, right *CSSName, bottom *CSSName, left *CSSName) *CSSNameCSSSideProperties {
	return &CSSNameCSSSideProperties{top: top, right: right, bottom: bottom, left: left}
}

func (p *CSSNameCSSSideProperties) Top() *CSSName {
	return p.top
}

func (p *CSSNameCSSSideProperties) Right() *CSSName {
	return p.right
}

func (p *CSSNameCSSSideProperties) Bottom() *CSSName {
	return p.bottom
}

func (p *CSSNameCSSSideProperties) Left() *CSSName {
	return p.left
}

// CSSNameUnsupportedCss3Properties holds the known CSS3 property names not
// supported by FlyingSaucer (HTML4/CSS2 only). Used to give a targeted warning
// instead of a generic "unrecognized property" message.
var CSSNameUnsupportedCss3Properties = map[string]struct{}{
	// Flexbox
	"flex":            {},
	"flex-direction":  {},
	"flex-wrap":       {},
	"flex-flow":       {},
	"flex-grow":       {},
	"flex-shrink":     {},
	"flex-basis":      {},
	"justify-content": {},
	"align-items":     {},
	"align-content":   {},
	"align-self":      {},
	"order":           {},
	// CSS Grid
	"grid":                  {},
	"grid-template":         {},
	"grid-template-columns": {},
	"grid-template-rows":    {},
	"grid-template-areas":   {},
	"grid-column":           {},
	"grid-row":              {},
	"grid-area":             {},
	"gap":                   {},
	"column-gap":            {},
	"row-gap":               {},
	"grid-auto-columns":     {},
	"grid-auto-rows":        {},
	"grid-auto-flow":        {},
	// Transitions
	"transition":                 {},
	"transition-property":        {},
	"transition-duration":        {},
	"transition-timing-function": {},
	"transition-delay":           {},
	// Animations
	"animation":                 {},
	"animation-name":            {},
	"animation-duration":        {},
	"animation-timing-function": {},
	"animation-delay":           {},
	"animation-iteration-count": {},
	"animation-direction":       {},
	"animation-fill-mode":       {},
	"animation-play-state":      {},
	// Transforms (3D only; 2D transform/transform-origin are supported)
	"transform-style": {},
	// Visual effects
	"box-shadow":      {},
	"text-shadow":     {},
	"filter":          {},
	"backdrop-filter": {},
	// Other common CSS3 properties
	"user-select":     {},
	"pointer-events":  {},
	"resize":          {},
	"will-change":     {},
	"object-fit":      {},
	"object-position": {},
	"appearance":      {},
}

// CSSNameUnsupportedCss3DisplayValues holds the CSS3 values for the display
// property (flexbox and grid layouts) not supported by FlyingSaucer.
var CSSNameUnsupportedCss3DisplayValues = map[string]struct{}{"flex": {}, "inline-flex": {}, "grid": {}, "inline-grid": {}}
