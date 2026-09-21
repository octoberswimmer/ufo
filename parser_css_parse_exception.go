// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/CSSParseException.java

package ufo

import (
	"strconv"
	"strings"
)

// CSSParseException is used as a panic value inside the CSS parser, which
// recovers it where the Java code catches it. It implements error.
type CSSParseException struct {
	found    *Token
	expected []*Token
	line     int

	// genericMessage is nil when the exception was built from tokens.
	genericMessage *string

	cause error

	callerNotified bool
}

func NewCSSParseException(message string, line int) *CSSParseException {
	return NewCSSParseExceptionWithCause(message, line, nil)
}

func NewCSSParseExceptionWithCause(message string, line int, cause error) *CSSParseException {
	return &CSSParseException{
		found:          nil,
		expected:       nil,
		line:           line,
		genericMessage: &message,
		cause:          cause,
	}
}

// NewCSSParseExceptionToken ports CSSParseException(Token found, Token
// expected, int line).
func NewCSSParseExceptionToken(found *Token, expected *Token, line int) *CSSParseException {
	return &CSSParseException{
		found:          found,
		expected:       []*Token{expected},
		line:           line,
		genericMessage: nil,
	}
}

// NewCSSParseExceptionTokenArray ports CSSParseException(Token found,
// Token[] expected, int line).
func NewCSSParseExceptionTokenArray(found *Token, expected []*Token, line int) *CSSParseException {
	e := &CSSParseException{
		found:          found,
		line:           line,
		genericMessage: nil,
	}
	if expected == nil {
		e.expected = []*Token{}
	} else {
		e.expected = make([]*Token, len(expected))
		copy(e.expected, expected)
	}
	return e
}

func (e *CSSParseException) GetMessage() string {
	if e.genericMessage != nil {
		return *e.genericMessage + " at line " + strconv.Itoa(e.line+1) + "."
	}
	var found string
	if e.found == nil {
		found = "end of file"
	} else {
		found = e.found.GetExternalName()
	}
	return "Found " + found + " where " +
		e.descr(e.expected) + " was expected at line " + strconv.Itoa(e.line+1) + "."
}

// Error implements error with the text of GetMessage.
func (e *CSSParseException) Error() string {
	return e.GetMessage()
}

func (e *CSSParseException) GetCause() error {
	return e.cause
}

// Unwrap returns the cause, for errors.Is and errors.As.
func (e *CSSParseException) Unwrap() error {
	return e.cause
}

func (e *CSSParseException) descr(tokens []*Token) string {
	if len(tokens) == 1 {
		return tokens[0].GetExternalName()
	}
	var result strings.Builder
	if len(tokens) > 2 {
		result.WriteString("one of ")
	}
	for i := 0; i < len(tokens); i++ {
		result.WriteString(tokens[i].GetExternalName())
		if i < len(tokens)-2 {
			result.WriteString(", ")
		} else if i == len(tokens)-2 {
			if len(tokens) > 2 {
				result.WriteString(", or ")
			} else {
				result.WriteString(" or ")
			}
		}
	}
	return result.String()
}

func (e *CSSParseException) GetFound() *Token {
	return e.found
}

func (e *CSSParseException) GetLine() int {
	return e.line
}

func (e *CSSParseException) SetLine(i int) {
	e.line = i
}

func (e *CSSParseException) IsEOF() bool {
	return e.found == TokenTkEof
}

func (e *CSSParseException) IsCallerNotified() bool {
	return e.callerNotified
}

func (e *CSSParseException) SetCallerNotified(callerNotified bool) {
	e.callerNotified = callerNotified
}
