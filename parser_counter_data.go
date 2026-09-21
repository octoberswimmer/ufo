// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/CounterData.java

package ufo

type CounterData struct {
	name  string
	value int
}

func NewCounterData(name string, value int) *CounterData {
	return &CounterData{name: name, value: value}
}

func (c *CounterData) GetName() string {
	return c.name
}

func (c *CounterData) GetValue() int {
	return c.value
}
