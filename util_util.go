// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/Util.java

package ufo

import "strings"

func UtilReplace(source, target, replacement string) string {
	var output strings.Builder
	n := 0
	for {
		off := strings.Index(source[n:], target)
		if off == -1 {
			output.WriteString(source[n:])
			break
		}
		off += n
		output.WriteString(source[n:off])
		output.WriteString(replacement)
		n = off + len(target)
	}
	return output.String()
}

// UtilIsNullOrEmpty reports whether str is empty. A null Java string is the
// empty string here.
func UtilIsNullOrEmpty(str string) bool {
	return str == ""
}
