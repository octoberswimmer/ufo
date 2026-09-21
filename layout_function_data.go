// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/FunctionData.java

package ufo

// FunctionData contains all the information necessary to invoke a
// ContentFunction.
type FunctionData struct {
	contentFunction ContentFunction
	function        *FSFunction
}

func NewFunctionData(contentFunction ContentFunction, function *FSFunction) *FunctionData {
	return &FunctionData{contentFunction: contentFunction, function: function}
}

func (f *FunctionData) GetContentFunction() ContentFunction {
	return f.contentFunction
}

func (f *FunctionData) GetFunction() *FSFunction {
	return f.function
}
