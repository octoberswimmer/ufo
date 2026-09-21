// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/FSFunction.java

package ufo

import "strings"

type FSFunction struct {
	name       string
	parameters []*PropertyValue
}

func NewFSFunction(name string, parameters []*PropertyValue) *FSFunction {
	return &FSFunction{name: name, parameters: parameters}
}

func (f *FSFunction) Is(name string) bool {
	return f.name == name
}

func (f *FSFunction) GetName() string {
	return f.name
}

func (f *FSFunction) GetParameters() []*PropertyValue {
	return f.parameters
}

func (f *FSFunction) ToString() string {
	var result strings.Builder
	result.WriteString(f.name)
	result.WriteByte('(')
	for _, parameter := range f.parameters {
		result.WriteString(parameter.ToString()) // HACK
		result.WriteByte(',')
	}
	result.WriteByte(')')
	return result.String()
}

func (f *FSFunction) String() string {
	return f.ToString()
}
