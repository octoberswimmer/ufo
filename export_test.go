package ufo

// Test-only access for the external test package ufo_test, whose tests lay
// documents out with the real PDF renderer (package pdf imports this one, so
// only an external test package can use both).

// LayerFloatsForTest returns the floats a layer holds (Java's private
// Layer.getFloats, which the Java-side dump reads by reflection).
func LayerFloatsForTest(l *Layer) []BlockBoxI { return l.floats }

// BoxSimpleClassNameForTest is the Java simple class name of a box.
func BoxSimpleClassNameForTest(b BoxI) string { return boxSimpleClassName(b) }
