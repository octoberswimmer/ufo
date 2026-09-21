// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/XRLogger.java

package ufo

// XRLogger is an interface whose implementations log Flying Saucer log
// messages.
type XRLogger interface {
	Log(where string, level *Level, msg string)

	// LogWithTh logs msg together with the error that caused it; th may be nil.
	LogWithTh(where string, level *Level, msg string, th error)

	SetLevel(logger string, level *Level)
}
