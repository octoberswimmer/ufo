// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/simple/extend/StyleBuilder.java

package ufo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// Java applies the pattern with Matcher.matches, hence the anchors.
var styleBuilderReInteger = regexp.MustCompile(`^\d+$`)

type StyleBuilder struct {
	style strings.Builder
}

func NewStyleBuilder() *StyleBuilder {
	return &StyleBuilder{}
}

// styleBuilderTrim ports String.trim, which removes the characters up to
// U+0020 from both ends.
func styleBuilderTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}

func (b *StyleBuilder) Append(e *dom.Element, attributeName string, styleName string) {
	attributeValue := styleBuilderTrim(e.GetAttribute(attributeName))
	if attributeValue != "" {
		b.AppendStyle(styleName, attributeValue)
	}
}

func (b *StyleBuilder) AppendUrl(e *dom.Element, attributeName string, styleName string) {
	attributeValue := styleBuilderTrim(e.GetAttribute(attributeName))
	if attributeValue != "" {
		b.AppendStyle(styleName, "url("+attributeValue+")")
	}
}

func (b *StyleBuilder) AppendLength(e *dom.Element, attributeName string, styleName string) {
	b.AppendLengthWithSuffix(e, attributeName, styleName, ";")
}

func (b *StyleBuilder) AppendLengthWithSuffix(e *dom.Element, attributeName string, styleName string, suffix string) {
	attributeValue := styleBuilderTrim(e.GetAttribute(attributeName))
	if attributeValue != "" {
		b.AppendStyleWithSuffix(styleName, b.ConvertToLength(attributeValue), suffix)
	}
}

func (b *StyleBuilder) AppendStyle(styleName string, attributeValue string) {
	b.AppendStyleWithSuffix(styleName, attributeValue, ";")
}

func (b *StyleBuilder) AppendStyleWithSuffix(styleName string, attributeValue string, suffix string) {
	b.style.WriteString(styleName)
	b.style.WriteString(strings.ToLower(attributeValue))
	b.style.WriteString(suffix)
}

func (b *StyleBuilder) AppendWidth(e *dom.Element) {
	b.AppendLength(e, "width", "width: ")
}

func (b *StyleBuilder) AppendHeight(e *dom.Element) {
	b.AppendLength(e, "height", "height: ")
}

func (b *StyleBuilder) ConvertToLength(value string) string {
	if b.IsInteger(value) {
		return value + "px"
	}
	return value
}

func (b *StyleBuilder) IsInteger(value string) bool {
	return styleBuilderReInteger.MatchString(value)
}

func (b *StyleBuilder) AppendRawStyle(cssStyle string) {
	b.style.WriteString(cssStyle)
}

func (b *StyleBuilder) ApplyTableContentAlign(e *dom.Element) {
	b.Append(e, "align", "text-align: ")
	b.Append(e, "valign", "vertical-align: ")
}

func (b *StyleBuilder) ApplyFloatingAlign(e *dom.Element) {
	s := strings.ToLower(styleBuilderTrim(e.GetAttribute("align")))
	switch s {
	case "":
	case "left", "right":
		b.AppendStyle("float: ", s)
	case "center":
		b.AppendRawStyle("margin-left: auto; margin-right: auto;")
	default:
		panic(NewXRRuntimeException(fmt.Sprintf("Unknown align attribute: '%s'", s)))
	}
}

func (b *StyleBuilder) String() string {
	return b.style.String()
}

func (b *StyleBuilder) ToString() string {
	return b.String()
}
