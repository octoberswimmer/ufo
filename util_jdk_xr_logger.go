// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/JDKXRLogger.java

package ufo

import (
	"log"
	"strings"
	"sync"
)

// JDKXRLogger is the default XRLogger. The Java class logs through
// java.util.logging; this one keeps the per-logger levels that
// java.util.logging keeps and writes to the standard logger of Go's log
// package.
//
// What is kept from the Java class: the logger levels are read from the
// configuration keys "xr.util-logging.<logger name>.level"; a logger without a
// level takes the level of its parent (the name up to the last '.'), and the
// root logger has level INFO; a level string that cannot be parsed sets OFF.
//
// What differs:
//   - java.util.logging handlers and formatters are not ported. The keys
//     "xr.util-logging.handlers", "...use-parent-handler", "...formatter" and
//     the handler level are read but have no effect. The one output is the
//     standard logger, which, as PORTING.md specifies, discards messages below
//     WARNING. A message is written as "<logger name> <level>:: <message>",
//     the default format of XRSimpleLogFormatter, followed by the error text
//     when there is an error.
//   - The Java constructor reads the configuration. Here that happens on the
//     first use of the logger, because the default instance is created while
//     package-level variables are initialized, when Configuration may not be
//     initialized yet.
//   - Configuration.setConfigLogger takes a java.util.logging.Logger in Java
//     and a ConfigurationLogger here; it is given a jdkXRLoggerNamedLogger,
//     which logs to this logger under the name XRLogLoggerConfig.
type JDKXRLogger struct {
	mu         sync.Mutex
	configured bool
	levels     map[string]*Level

	// handlerLevel is the level below which the output discards messages.
	handlerLevel *Level
}

func NewJDKXRLogger() *JDKXRLogger {
	return &JDKXRLogger{
		levels:       map[string]*Level{},
		handlerLevel: LevelWarning,
	}
}

func (l *JDKXRLogger) Log(where string, level *Level, msg string) {
	l.configure()
	if l.isLoggable(where, level) {
		log.Print(where + " " + level.GetName() + ":: " + msg)
	}
}

func (l *JDKXRLogger) LogWithTh(where string, level *Level, msg string, th error) {
	l.configure()
	if l.isLoggable(where, level) {
		if th != nil {
			log.Print(where + " " + level.GetName() + ":: " + msg + ": " + th.Error())
		} else {
			log.Print(where + " " + level.GetName() + ":: " + msg)
		}
	}
}

func (l *JDKXRLogger) SetLevel(logger string, level *Level) {
	l.configure()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.levels[logger] = level
}

// isLoggable applies the level of the logger named where, as
// java.util.logging.Logger.isLoggable does, and then the level of the output,
// as java.util.logging.Handler.isLoggable does.
func (l *JDKXRLogger) isLoggable(where string, level *Level) bool {
	loggerLevel := l.getEffectiveLevel(where)
	if level.IntValue() < loggerLevel.IntValue() || loggerLevel == LevelOff {
		return false
	}
	if level.IntValue() < l.handlerLevel.IntValue() || l.handlerLevel == LevelOff {
		return false
	}
	return true
}

// getEffectiveLevel returns the level of the named logger, or of the nearest
// parent that has one, or INFO, the default level of the root logger.
func (l *JDKXRLogger) getEffectiveLevel(logger string) *Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	name := logger
	for {
		if level, ok := l.levels[name]; ok && level != nil {
			return level
		}
		i := strings.LastIndexByte(name, '.')
		if i < 0 {
			return LevelInfo
		}
		name = name[:i]
	}
}

// configure does what the Java constructor does. It marks the logger as
// configured before it reads the configuration, so that a message logged
// while the configuration is read does not start a second configuration.
func (l *JDKXRLogger) configure() {
	l.mu.Lock()
	if l.configured {
		l.mu.Unlock()
		return
	}
	l.configured = true
	l.mu.Unlock()

	props := jdkXRLoggerRetrieveLoggingProperties()

	if !XRLogIsLoggingEnabled() {
		ConfigurationSetConfigLogger(&jdkXRLoggerNamedLogger{logger: l, name: XRLogLoggerConfig})
		return
	}

	l.initializeJDKLogManager(props)

	ConfigurationSetConfigLogger(&jdkXRLoggerNamedLogger{logger: l, name: XRLogLoggerConfig})
}

// jdkXRLoggerNamedLogger stands for the java.util.logging.Logger that
// Logger.getLogger(name) returns: it logs to a JDKXRLogger under one logger
// name, whether or not XRLog logging is enabled.
type jdkXRLoggerNamedLogger struct {
	logger *JDKXRLogger
	name   string
}

func (n *jdkXRLoggerNamedLogger) Log(level *Level, msg string) {
	n.logger.Log(n.name, level, msg)
}

func jdkXRLoggerRetrieveLoggingProperties() map[string]string {
	// pull logging properties from configuration
	// they are all prefixed as shown
	prefix := "xr.util-logging."
	props := map[string]string{}
	for _, fullKey := range ConfigurationKeysByPrefix(prefix) {
		key := fullKey[len(prefix):]
		value := ConfigurationValueFor(fullKey)
		props[key] = value
	}
	return props
}

func (l *JDKXRLogger) initializeJDKLogManager(fsLoggingProperties map[string]string) {
	// load our properties into our log manager
	for key, prop := range fsLoggingProperties {
		if strings.HasSuffix(key, "level") {
			l.configureLogLevel(key[:strings.LastIndexByte(key, '.')], prop)
		}
		// Keys ending in "handlers" and "formatter" configure
		// java.util.logging handlers in Java; see the comment on JDKXRLogger.
	}
}

// configureLogLevel parses the levelValue into a Level instance and assigns
// it to the logger named by loggerName; if the levelValue is invalid (e.g.
// misspelled), assigns LevelOff to the logger.
func (l *JDKXRLogger) configureLogLevel(loggerName string, levelValue string) {
	level := LoggerUtilParseLogLevel(levelValue, LevelOff)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.levels[loggerName] = level
}
