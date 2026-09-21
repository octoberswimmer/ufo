// Tests of parser_property_value.go, parser_css_parse_exception.go and the
// float formatting of parser_fs_rgb_color.go. The Java suite has no tests of
// these classes.

package ufo

import "testing"

func TestPropertyValue_constructors(t *testing.T) {
	number := NewPropertyValueFloat(CSSPrimitiveValueCssNumber, 2, "2")
	if number.GetPropertyValueType() != PropertyValueTypeValueTypeNumber || number.GetFloatValue() != 2 {
		t.Errorf("number: type %v, value %v", number.GetPropertyValueType(), number.GetFloatValue())
	}
	// A number of 0 is a length, as in Java.
	zero := NewPropertyValueFloat(CSSPrimitiveValueCssNumber, 0, "0")
	if zero.GetPropertyValueType() != PropertyValueTypeValueTypeLength {
		t.Errorf("zero: type %v", zero.GetPropertyValueType())
	}
	inherit := NewPropertyValueString(CSSPrimitiveValueCssIdent, "INHERIT", "INHERIT")
	if inherit.GetCssValueType() != CSSValueCssInherit || inherit.GetPropertyValueType() != PropertyValueTypeValueTypeIdent {
		t.Errorf("inherit: css value type %d, type %v", inherit.GetCssValueType(), inherit.GetPropertyValueType())
	}
	color := NewPropertyValueFSColor(FSRGBColorRed)
	if color.GetCssText() != "#ff0000" || color.GetPrimitiveType() != CSSPrimitiveValueCssRgbcolor || color.GetFSColor() != FSColor(FSRGBColorRed) {
		t.Errorf("color: text %q, primitive type %d", color.GetCssText(), color.GetPrimitiveType())
	}
	list := NewPropertyValueList([]any{number, color})
	if list.GetCssText() != "[2, #ff0000]" || len(list.GetValues()) != 2 || list.GetCssValueType() != CSSValueCssCustom {
		t.Errorf("list: text %q, %d values", list.GetCssText(), len(list.GetValues()))
	}
	operator := NewPropertyValueStringWithOperatorToken(CSSPrimitiveValueCssString, "a", "'a'", TokenTkComma)
	if operator.GetOperator() != TokenTkComma || len(operator.GetStringArrayValue()) != 0 {
		t.Errorf("operator: %v, string array %v", operator.GetOperator(), operator.GetStringArrayValue())
	}
}

func TestPropertyValue_getFingerprint(t *testing.T) {
	ident := NewPropertyValueString(CSSPrimitiveValueCssIdent, "auto", "auto")
	if got, want := ident.GetFingerprint(), "I3"; got != want {
		t.Errorf("GetFingerprint() = %q, want %q", got, want)
	}
	if ident.GetIdentValue() != IdentValueAuto {
		t.Error("GetFingerprint did not resolve the ident value")
	}
	length := NewPropertyValueFloat(CSSPrimitiveValueCssPx, 1, "1px")
	if got := length.GetFingerprint(); got != "1px" {
		t.Errorf("GetFingerprint() = %q, want %q", got, "1px")
	}
}

func TestCSSParseException_getMessage(t *testing.T) {
	cases := []struct {
		e    *CSSParseException
		want string
	}{
		{NewCSSParseException("Bad value", 4), "Bad value at line 5."},
		{NewCSSParseExceptionToken(TokenTkComma, TokenTkSemicolon, 0), "Found a comma where ; was expected at line 1."},
		{NewCSSParseExceptionTokenArray(nil, []*Token{TokenTkIdent, TokenTkString}, 0),
			"Found end of file where an identifier or a string was expected at line 1."},
		{NewCSSParseExceptionTokenArray(TokenTkEof, []*Token{TokenTkIdent, TokenTkString, TokenTkUri}, 1),
			"Found end of file where one of an identifier, a string, or a URI was expected at line 2."},
	}
	for _, c := range cases {
		if got := c.e.GetMessage(); got != c.want {
			t.Errorf("GetMessage() = %q, want %q", got, c.want)
		}
	}
	if !cases[3].e.IsEOF() || cases[1].e.IsEOF() {
		t.Error("IsEOF does not compare the found token with TokenTkEof")
	}
}

func TestFSRGBColor_floatToString(t *testing.T) {
	cases := []struct {
		f    float32
		want string
	}{
		{1, "1.0"}, {0, "0.0"}, {0.5, "0.5"}, {0.42, "0.42"}, {100, "100.0"},
		{0.001, "0.001"}, {0.0001, "1.0E-4"}, {1.5e7, "1.5E7"}, {-2.25, "-2.25"},
	}
	for _, c := range cases {
		if got := fsRGBColorFloatToString(c.f); got != c.want {
			t.Errorf("fsRGBColorFloatToString(%v) = %q, want %q", c.f, got, c.want)
		}
	}
	if got := NewFSCMYKColor(0, 0.5, 1, 0.25).ToString(); got != "cmyk(0.0, 0.5, 1.0, 0.25)" {
		t.Errorf("FSCMYKColor.ToString() = %q", got)
	}
}
