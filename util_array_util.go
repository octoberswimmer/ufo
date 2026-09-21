// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/ArrayUtil.java

package ufo

// The three Java overloads of ArrayUtil.cloneOrEmpty differ only by the
// element type of the array, so each Go function carries the type in its
// name. A nil slice stands for a null Java array.

func ArrayUtilCloneOrEmptyStringArray(source []string) []string {
	if source == nil {
		return ConstantsEmptyStrArr
	}
	result := make([]string, len(source))
	copy(result, source)
	return result
}

func ArrayUtilCloneOrEmptyByteArray(source []byte) []byte {
	if source == nil {
		return ConstantsEmptyByteArr
	}
	result := make([]byte, len(source))
	copy(result, source)
	return result
}

func ArrayUtilCloneOrEmptyIntArray(source []int) []int {
	if source == nil {
		return ConstantsEmptyIntArr
	}
	result := make([]int, len(source))
	copy(result, source)
	return result
}
