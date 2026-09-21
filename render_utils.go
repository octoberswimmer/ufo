// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/Utils.java

package ufo

import "strings"

func UtilsAppendPositioningInfo(style CalculatedStyleI, result *strings.Builder) {
	if style.IsRelative() {
		result.WriteString("(relative) ")
	}
	if style.IsFixed() {
		result.WriteString("(fixed) ")
	}
	if style.IsAbsolute() {
		result.WriteString("(absolute) ")
	}
	if style.IsFloated() {
		result.WriteString("(floated) ")
	}
}
