// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/XRRuntimeException.java

package ufo

// XRRuntimeException is the general runtime exception used in XHTMLRenderer.
// Auto-logs messages to plumbing.exception hierarchy. It is used as a panic
// value and implements error.
type XRRuntimeException struct {
	msg   string
	cause error
}

// NewXRRuntimeException instantiates a new exception with a "reason" message.
//
// msg is the reason the exception is being thrown.
func NewXRRuntimeException(msg string) *XRRuntimeException {
	e := &XRRuntimeException{msg: msg}
	e.log(msg)
	return e
}

// NewXRRuntimeExceptionWithCause instantiates a new exception with a "reason"
// message.
//
// msg is the reason the exception is being thrown. cause is the error that
// caused this exception to be thrown, e.g. an I/O error.
func NewXRRuntimeExceptionWithCause(msg string, cause error) *XRRuntimeException {
	e := &XRRuntimeException{msg: msg, cause: cause}
	e.logWithCause(msg, cause)
	return e
}

// log logs the exception message.
func (e *XRRuntimeException) log(msg string) {
	XRLogException("Unhandled exception. " + msg)
}

// logWithCause logs the exception's message, plus the error that caused the
// exception to be thrown.
func (e *XRRuntimeException) logWithCause(msg string, cause error) {
	XRLogExceptionWithTh("Unhandled exception. "+msg, cause)
}

func (e *XRRuntimeException) GetMessage() string {
	return e.msg
}

func (e *XRRuntimeException) GetCause() error {
	return e.cause
}

// Error implements error with the exception message.
func (e *XRRuntimeException) Error() string {
	return e.msg
}

// Unwrap returns the cause, for errors.Is and errors.As.
func (e *XRRuntimeException) Unwrap() error {
	return e.cause
}
