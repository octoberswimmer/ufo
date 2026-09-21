// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/OneToFourPropertyBuilders.java

package ufo

// oneToFourPropertyBuildersOneToFourPropertyBuilder is the Java abstract class
// OneToFourPropertyBuilder. The abstract getProperties() and
// getPropertyBuilder() are the function fields of the same names, set by the
// constructor of each subclass.
type oneToFourPropertyBuildersOneToFourPropertyBuilder struct {
	AbstractPropertyBuilder
	getProperties      func() []*CSSName
	getPropertyBuilder func() PropertyBuilder
}

func newOneToFourPropertyBuildersOneToFourPropertyBuilder(getProperties func() []*CSSName, getPropertyBuilder func() PropertyBuilder) *oneToFourPropertyBuildersOneToFourPropertyBuilder {
	b := &oneToFourPropertyBuildersOneToFourPropertyBuilder{getProperties: getProperties, getPropertyBuilder: getPropertyBuilder}
	b.self = b
	return b
}

func (b *oneToFourPropertyBuildersOneToFourPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	result := make([]*PropertyDeclaration, 0, 4)
	b.AssertFoundUpToValues(cssName, values, 4)

	builder := b.getPropertyBuilder()

	props := b.getProperties()

	var decl1 *PropertyDeclaration
	var decl2 *PropertyDeclaration
	var decl3 *PropertyDeclaration
	var decl4 *PropertyDeclaration
	switch len(values) {
	case 1:
		decl1 = builder.BuildDeclarations(
			cssName, values, origin, important)[0]

		result = append(result, b.CopyOf(decl1, props[0]))
		result = append(result, b.CopyOf(decl1, props[1]))
		result = append(result, b.CopyOf(decl1, props[2]))
		result = append(result, b.CopyOf(decl1, props[3]))

	case 2:
		decl1 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[0:1], origin, important, false)[0]
		decl2 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[1:2], origin, important, false)[0]

		result = append(result, b.CopyOf(decl1, props[0]))
		result = append(result, b.CopyOf(decl2, props[1]))
		result = append(result, b.CopyOf(decl1, props[2]))
		result = append(result, b.CopyOf(decl2, props[3]))

	case 3:
		decl1 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[0:1], origin, important, false)[0]
		decl2 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[1:2], origin, important, false)[0]
		decl3 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[2:3], origin, important, false)[0]

		result = append(result, b.CopyOf(decl1, props[0]))
		result = append(result, b.CopyOf(decl2, props[1]))
		result = append(result, b.CopyOf(decl3, props[2]))
		result = append(result, b.CopyOf(decl2, props[3]))

	case 4:
		decl1 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[0:1], origin, important, false)[0]
		decl2 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[1:2], origin, important, false)[0]
		decl3 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[2:3], origin, important, false)[0]
		decl4 = builder.BuildDeclarationsWithInheritAllowed(
			cssName, values[3:4], origin, important, false)[0]

		result = append(result, b.CopyOf(decl1, props[0]))
		result = append(result, b.CopyOf(decl2, props[1]))
		result = append(result, b.CopyOf(decl3, props[2]))
		result = append(result, b.CopyOf(decl4, props[3]))
	}

	return result
}

type OneToFourPropertyBuildersBorderColor struct {
	oneToFourPropertyBuildersOneToFourPropertyBuilder
}

func NewOneToFourPropertyBuildersBorderColor() *OneToFourPropertyBuildersBorderColor {
	b := &OneToFourPropertyBuildersBorderColor{*newOneToFourPropertyBuildersOneToFourPropertyBuilder(
		func() []*CSSName {
			return []*CSSName{
				CSSNameBorderTopColor,
				CSSNameBorderRightColor,
				CSSNameBorderBottomColor,
				CSSNameBorderLeftColor}
		},
		func() PropertyBuilder {
			return PrimitivePropertyBuildersColorBuilder
		})}
	b.self = b
	return b
}

type OneToFourPropertyBuildersBorderStyle struct {
	oneToFourPropertyBuildersOneToFourPropertyBuilder
}

func NewOneToFourPropertyBuildersBorderStyle() *OneToFourPropertyBuildersBorderStyle {
	b := &OneToFourPropertyBuildersBorderStyle{*newOneToFourPropertyBuildersOneToFourPropertyBuilder(
		func() []*CSSName {
			return []*CSSName{
				CSSNameBorderTopStyle,
				CSSNameBorderRightStyle,
				CSSNameBorderBottomStyle,
				CSSNameBorderLeftStyle}
		},
		func() PropertyBuilder {
			return PrimitivePropertyBuildersBorderStyleBuilder
		})}
	b.self = b
	return b
}

type OneToFourPropertyBuildersBorderWidth struct {
	oneToFourPropertyBuildersOneToFourPropertyBuilder
}

func NewOneToFourPropertyBuildersBorderWidth() *OneToFourPropertyBuildersBorderWidth {
	b := &OneToFourPropertyBuildersBorderWidth{*newOneToFourPropertyBuildersOneToFourPropertyBuilder(
		func() []*CSSName {
			return []*CSSName{
				CSSNameBorderTopWidth,
				CSSNameBorderRightWidth,
				CSSNameBorderBottomWidth,
				CSSNameBorderLeftWidth}
		},
		func() PropertyBuilder {
			return PrimitivePropertyBuildersBorderWidthBuilder
		})}
	b.self = b
	return b
}

type OneToFourPropertyBuildersBorderRadius struct {
	oneToFourPropertyBuildersOneToFourPropertyBuilder
}

func NewOneToFourPropertyBuildersBorderRadius() *OneToFourPropertyBuildersBorderRadius {
	b := &OneToFourPropertyBuildersBorderRadius{*newOneToFourPropertyBuildersOneToFourPropertyBuilder(
		func() []*CSSName {
			return []*CSSName{
				CSSNameBorderTopLeftRadius,
				CSSNameBorderTopRightRadius,
				CSSNameBorderBottomRightRadius,
				CSSNameBorderBottomLeftRadius}
		},
		func() PropertyBuilder {
			return PrimitivePropertyBuildersBorderRadiusBuilder
		})}
	b.self = b
	return b
}

type OneToFourPropertyBuildersMargin struct {
	oneToFourPropertyBuildersOneToFourPropertyBuilder
}

func NewOneToFourPropertyBuildersMargin() *OneToFourPropertyBuildersMargin {
	b := &OneToFourPropertyBuildersMargin{*newOneToFourPropertyBuildersOneToFourPropertyBuilder(
		func() []*CSSName {
			return []*CSSName{
				CSSNameMarginTop,
				CSSNameMarginRight,
				CSSNameMarginBottom,
				CSSNameMarginLeft}
		},
		func() PropertyBuilder {
			return PrimitivePropertyBuildersMarginBuilder
		})}
	b.self = b
	return b
}

type OneToFourPropertyBuildersPadding struct {
	oneToFourPropertyBuildersOneToFourPropertyBuilder
}

func NewOneToFourPropertyBuildersPadding() *OneToFourPropertyBuildersPadding {
	b := &OneToFourPropertyBuildersPadding{*newOneToFourPropertyBuildersOneToFourPropertyBuilder(
		func() []*CSSName {
			return []*CSSName{
				CSSNamePaddingTop,
				CSSNamePaddingRight,
				CSSNamePaddingBottom,
				CSSNamePaddingLeft}
		},
		func() PropertyBuilder {
			return PrimitivePropertyBuildersPaddingBuilder
		})}
	b.self = b
	return b
}
