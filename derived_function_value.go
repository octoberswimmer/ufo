// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/FunctionValue.java

package ufo

type FunctionValue struct {
	DerivedValue
	function *FSFunction
}

func NewFunctionValue(name *CSSName, value *PropertyValue) *FunctionValue {
	cssText := value.GetCssText()
	return &FunctionValue{
		DerivedValue: NewDerivedValue(name, value.GetPrimitiveType(), &cssText, &cssText),

		function: value.GetFunction(),
	}
}

func (f *FunctionValue) GetFunction() *FSFunction {
	return f.function
}
