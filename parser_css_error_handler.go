// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/CSSErrorHandler.java

package ufo

type CSSErrorHandler interface {
	// Error reports a CSS parse problem; uri is "" when there is none.
	Error(uri string, message string)
}

// CSSErrorHandlerFunc adapts a function to CSSErrorHandler. It stands for
// the Java lambdas that implement the interface.
type CSSErrorHandlerFunc func(uri string, message string)

func (f CSSErrorHandlerFunc) Error(uri string, message string) {
	f(uri, message)
}
