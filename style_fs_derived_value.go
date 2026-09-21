// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/FSDerivedValue.java

package ufo

// FSDerivedValue is the marker interface for all derived values. All methods
// for any possible style are declared here, which doesn't make complete
// sense, as, for example, a length can't return a value for AsColor().
// This is done so that CalculatedStyle can just look up an
// FSDerivedValue, without casting, and call the appropriate function
// without a cast to the appropriate subtype.
// The users of CalculatedStyle have to then make sure they don't
// make meaningless calls like AsColor(CSSNameHeight). DerivedValue
// and IdentValue, the two implementations of this interface, just
// panic with an XRRuntimeException if they can't handle the call.
//
// NOTE: When resolving proportional property values, implementations of this
// interface must be prepared to handle calls with different base values.
type FSDerivedValue interface {
	IsDeclaredInherit() bool

	AsFloat() float32

	// AsColor may return nil.
	AsColor() FSColor

	GetFloatProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) float32
	AsString() string
	AsStringArray() []string
	AsIdentValue() *IdentValue
	HasAbsoluteUnit() bool
	IsDependentOnFontSize() bool
	IsIdent() bool
}
