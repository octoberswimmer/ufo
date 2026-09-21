// Ported from the JDK class java.util.logging.Level, which
// flying-saucer-core/src/main/java/org/xhtmlrenderer/util/XRLog.java and its
// callers use to grade log messages.

package ufo

import "math"

// Level is a logging level with the name and the integer value of the
// matching java.util.logging.Level constant. A message is logged when its
// level value is at least the level value of the logger it is sent to.
type Level struct {
	name  string
	value int
}

var (
	LevelOff     = &Level{name: "OFF", value: math.MaxInt32}
	LevelSevere  = &Level{name: "SEVERE", value: 1000}
	LevelWarning = &Level{name: "WARNING", value: 900}
	LevelInfo    = &Level{name: "INFO", value: 800}
	LevelConfig  = &Level{name: "CONFIG", value: 700}
	LevelFine    = &Level{name: "FINE", value: 500}
	LevelFiner   = &Level{name: "FINER", value: 400}
	LevelFinest  = &Level{name: "FINEST", value: 300}
	LevelAll     = &Level{name: "ALL", value: math.MinInt32}
)

// GetName returns the name of the level, e.g. "WARNING".
func (l *Level) GetName() string {
	return l.name
}

// IntValue returns the integer value of the level.
func (l *Level) IntValue() int {
	return l.value
}

func (l *Level) ToString() string {
	return l.name
}

func (l *Level) String() string {
	return l.name
}
