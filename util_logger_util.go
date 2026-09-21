// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/LoggerUtil.java

package ufo

// LoggerUtil is a utility class for working with logging levels.

func LoggerUtilParseLogLevel(val string, defaultLogLevel *Level) *Level {
	switch val {
	case "ALL":
		return LevelAll
	case "CONFIG":
		return LevelConfig
	case "FINE":
		return LevelFine
	case "FINER":
		return LevelFiner
	case "FINEST":
		return LevelFinest
	case "INFO":
		return LevelInfo
	case "OFF":
		return LevelOff
	case "SEVERE":
		return LevelSevere
	case "WARNING":
		return LevelWarning
	default:
		return defaultLogLevel
	}
}
