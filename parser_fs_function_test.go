// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/FSFunctionTest.java

package ufo

import "testing"

var (
	fsFunctionTestParameter = NewPropertyValueFloat(CSSPrimitiveValueCssIdent, 0.0, "header")
	fsFunctionTestFunction  = NewFSFunction("element", []*PropertyValue{fsFunctionTestParameter})
)

func TestFSFunction_is(t *testing.T) {
	if !fsFunctionTestFunction.Is("element") {
		t.Error(`Is("element") = false, want true`)
	}
	if fsFunctionTestFunction.Is("counter") {
		t.Error(`Is("counter") = true, want false`)
	}
}

func TestFSFunction_getName(t *testing.T) {
	if got := fsFunctionTestFunction.GetName(); got != "element" {
		t.Errorf("GetName() = %q, want %q", got, "element")
	}
}

func TestFSFunction_getParameters(t *testing.T) {
	got := fsFunctionTestFunction.GetParameters()
	if len(got) != 1 || got[0] != fsFunctionTestParameter {
		t.Errorf("GetParameters() = %v, want exactly the one parameter", got)
	}
	if got := NewFSFunction("element", []*PropertyValue{}).GetParameters(); len(got) != 0 {
		t.Errorf("GetParameters() = %v, want empty", got)
	}
}

func TestFSFunction_stringRepresentation(t *testing.T) {
	if got := fsFunctionTestFunction.ToString(); got != "element(header,)" {
		t.Errorf("ToString() = %q, want %q", got, "element(header,)")
	}
}
