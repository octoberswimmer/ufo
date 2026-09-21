// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/XRLog.java

package ufo

import (
	"fmt"
	"sync/atomic"
)

// XRLog is a utility class for logging. It gives access to the various logs
// (plumbing.load, .init, .render) through one function per log.
//
// Overloads: PORTING.md derives the plain name both for the Java overload
// with the fewest parameters, XRLog.load(msg), and, in its logging example,
// for XRLog.load(level, msg). Each per-log function is therefore variadic and
// accepts exactly the three Java argument lists:
//
//	XRLogLoad(msg string)                          logged at the default level of the log
//	XRLogLoad(level *Level, msg string)
//	XRLogLoad(level *Level, msg string, th error)
//
// Any other argument list panics. XRLogLoadWithLevel and XRLogLoadWithTh are
// the typed forms of the second and third.
//
// Names: the Java logger-name constants (XRLog.LOAD, XRLog.CSS_PARSE, ...)
// derive the same Go names as the static methods (XRLog.load,
// XRLog.cssParse, ...). The methods keep the derived names and the constants
// are named XRLogLogger<Name>.

var xrLogLoggerNames = make([]string, 0, 20)

var (
	XRLogLoggerConfig      = xrLogRegisterLoggerByName("org.xhtmlrenderer.config")
	XRLogLoggerException   = xrLogRegisterLoggerByName("org.xhtmlrenderer.exception")
	XRLogLoggerGeneral     = xrLogRegisterLoggerByName("org.xhtmlrenderer.general")
	XRLogLoggerInit        = xrLogRegisterLoggerByName("org.xhtmlrenderer.init")
	XRLogLoggerJunit       = xrLogRegisterLoggerByName("org.xhtmlrenderer.junit")
	XRLogLoggerLoad        = xrLogRegisterLoggerByName("org.xhtmlrenderer.load")
	XRLogLoggerMatch       = xrLogRegisterLoggerByName("org.xhtmlrenderer.match")
	XRLogLoggerCascade     = xrLogRegisterLoggerByName("org.xhtmlrenderer.cascade")
	XRLogLoggerXmlEntities = xrLogRegisterLoggerByName("org.xhtmlrenderer.load.xml-entities")
	XRLogLoggerCssParse    = xrLogRegisterLoggerByName("org.xhtmlrenderer.css-parse")
	XRLogLoggerLayout      = xrLogRegisterLoggerByName("org.xhtmlrenderer.layout")
	XRLogLoggerRender      = xrLogRegisterLoggerByName("org.xhtmlrenderer.render")
)

var xrLogLoggerImpl XRLogger = NewJDKXRLogger()

// State of the loggingEnabled flag. Java reads
// Configuration.isTrue("xr.util-logging.loggingEnabled", true) in the static
// initializer of XRLog. Here the value is read on the first call of
// XRLogIsLoggingEnabled, so that package initialization of XRLog does not
// depend on Configuration, whose own initialization logs through XRLog.
// While the configuration value is being read, logging counts as enabled,
// which is the Java default.
const (
	xrLogLoggingUnset int32 = iota
	xrLogLoggingResolving
	xrLogLoggingOn
	xrLogLoggingOff
)

var xrLogLoggingEnabled atomic.Int32

func xrLogRegisterLoggerByName(loggerName string) string {
	xrLogLoggerNames = append(xrLogLoggerNames, loggerName)
	return loggerName
}

// XRLogListRegisteredLoggers returns a list of all loggers that will be
// accessed by XRLog. Each entry is a string with a logger name; example name
// might be "org.xhtmlrenderer.config".
//
// Returns the list of loggers, never nil.
func XRLogListRegisteredLoggers() []string {
	result := make([]string, len(xrLogLoggerNames))
	copy(result, xrLogLoggerNames)
	return result
}

// xrLogChannel implements the three Java overloads of a per-log method.
func xrLogChannel(where string, defaultLevel *Level, args []any) {
	switch len(args) {
	case 1:
		if msg, ok := args[0].(string); ok {
			XRLogLog(where, defaultLevel, msg)
			return
		}
	case 2:
		level, ok1 := args[0].(*Level)
		msg, ok2 := args[1].(string)
		if ok1 && ok2 {
			XRLogLog(where, level, msg)
			return
		}
	case 3:
		level, ok1 := args[0].(*Level)
		msg, ok2 := args[1].(string)
		if ok1 && ok2 {
			if args[2] == nil {
				XRLogLogWithTh(where, level, msg, nil)
				return
			}
			if th, ok := args[2].(error); ok {
				XRLogLogWithTh(where, level, msg, th)
				return
			}
			XRLogLogWithTh(where, level, msg, fmt.Errorf("%v", args[2]))
			return
		}
	}
	panic(fmt.Sprintf("XRLog: arguments must be (msg), (level, msg) or (level, msg, th), got %v", args))
}

// XRLogCssParse ports XRLog.cssParse(msg), XRLog.cssParse(level, msg) and
// XRLog.cssParse(level, msg, th); the message-only form logs at LevelInfo.
func XRLogCssParse(args ...any) {
	xrLogChannel(XRLogLoggerCssParse, LevelInfo, args)
}

func XRLogCssParseWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerCssParse, level, msg)
}

func XRLogCssParseWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerCssParse, level, msg, th)
}

// XRLogXmlEntities ports XRLog.xmlEntities(msg), XRLog.xmlEntities(level, msg) and
// XRLog.xmlEntities(level, msg, th); the message-only form logs at LevelInfo.
func XRLogXmlEntities(args ...any) {
	xrLogChannel(XRLogLoggerXmlEntities, LevelInfo, args)
}

func XRLogXmlEntitiesWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerXmlEntities, level, msg)
}

func XRLogXmlEntitiesWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerXmlEntities, level, msg, th)
}

// XRLogCascade ports XRLog.cascade(msg), XRLog.cascade(level, msg) and
// XRLog.cascade(level, msg, th); the message-only form logs at LevelInfo.
func XRLogCascade(args ...any) {
	xrLogChannel(XRLogLoggerCascade, LevelInfo, args)
}

func XRLogCascadeWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerCascade, level, msg)
}

func XRLogCascadeWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerCascade, level, msg, th)
}

func XRLogException(msg string) {
	XRLogExceptionWithTh(msg, nil)
}

func XRLogExceptionWithTh(msg string, th error) {
	XRLogLogWithTh(XRLogLoggerException, LevelWarning, msg, th)
}

// XRLogGeneral ports XRLog.general(msg), XRLog.general(level, msg) and
// XRLog.general(level, msg, th); the message-only form logs at LevelInfo.
func XRLogGeneral(args ...any) {
	xrLogChannel(XRLogLoggerGeneral, LevelInfo, args)
}

func XRLogGeneralWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerGeneral, level, msg)
}

func XRLogGeneralWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerGeneral, level, msg, th)
}

// XRLogInit ports XRLog.init(msg), XRLog.init(level, msg) and
// XRLog.init(level, msg, th); the message-only form logs at LevelInfo.
func XRLogInit(args ...any) {
	xrLogChannel(XRLogLoggerInit, LevelInfo, args)
}

func XRLogInitWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerInit, level, msg)
}

func XRLogInitWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerInit, level, msg, th)
}

// XRLogJunit ports XRLog.junit(msg), XRLog.junit(level, msg) and
// XRLog.junit(level, msg, th); the message-only form logs at LevelFinest.
func XRLogJunit(args ...any) {
	xrLogChannel(XRLogLoggerJunit, LevelFinest, args)
}

func XRLogJunitWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerJunit, level, msg)
}

func XRLogJunitWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerJunit, level, msg, th)
}

// XRLogLoad ports XRLog.load(msg), XRLog.load(level, msg) and
// XRLog.load(level, msg, th); the message-only form logs at LevelInfo.
func XRLogLoad(args ...any) {
	xrLogChannel(XRLogLoggerLoad, LevelInfo, args)
}

func XRLogLoadWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerLoad, level, msg)
}

func XRLogLoadWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerLoad, level, msg, th)
}

// XRLogMatch ports XRLog.match(msg), XRLog.match(level, msg) and
// XRLog.match(level, msg, th); the message-only form logs at LevelInfo.
func XRLogMatch(args ...any) {
	xrLogChannel(XRLogLoggerMatch, LevelInfo, args)
}

func XRLogMatchWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerMatch, level, msg)
}

func XRLogMatchWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerMatch, level, msg, th)
}

// XRLogLayout ports XRLog.layout(msg), XRLog.layout(level, msg) and
// XRLog.layout(level, msg, th); the message-only form logs at LevelInfo.
func XRLogLayout(args ...any) {
	xrLogChannel(XRLogLoggerLayout, LevelInfo, args)
}

func XRLogLayoutWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerLayout, level, msg)
}

func XRLogLayoutWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerLayout, level, msg, th)
}

// XRLogRender ports XRLog.render(msg), XRLog.render(level, msg) and
// XRLog.render(level, msg, th); the message-only form logs at LevelInfo.
func XRLogRender(args ...any) {
	xrLogChannel(XRLogLoggerRender, LevelInfo, args)
}

func XRLogRenderWithLevel(level *Level, msg string) {
	XRLogLog(XRLogLoggerRender, level, msg)
}

func XRLogRenderWithTh(level *Level, msg string, th error) {
	XRLogLogWithTh(XRLogLoggerRender, level, msg, th)
}

func XRLogLog(where string, level *Level, msg string) {
	if XRLogIsLoggingEnabled() {
		xrLogLoggerImpl.Log(where, level, msg)
	}
}

func XRLogLogWithTh(where string, level *Level, msg string, th error) {
	if XRLogIsLoggingEnabled() {
		xrLogLoggerImpl.LogWithTh(where, level, msg, th)
	}
}

func XRLogSetLevel(log string, level *Level) {
	xrLogLoggerImpl.SetLevel(log, level)
}

// XRLogIsLoggingEnabled reports whether logging is on or off.
//
// Returns true if logging is enabled, false if not. Corresponds to
// configuration file property xr.util-logging.loggingEnabled, or to the value
// passed to XRLogSetLoggingEnabled.
func XRLogIsLoggingEnabled() bool {
	switch xrLogLoggingEnabled.Load() {
	case xrLogLoggingOn:
		return true
	case xrLogLoggingOff:
		return false
	case xrLogLoggingResolving:
		return true
	}
	if !xrLogLoggingEnabled.CompareAndSwap(xrLogLoggingUnset, xrLogLoggingResolving) {
		return XRLogIsLoggingEnabled()
	}
	enabled := ConfigurationIsTrue("xr.util-logging.loggingEnabled", true)
	if enabled {
		xrLogLoggingEnabled.CompareAndSwap(xrLogLoggingResolving, xrLogLoggingOn)
	} else {
		xrLogLoggingEnabled.CompareAndSwap(xrLogLoggingResolving, xrLogLoggingOff)
	}
	return xrLogLoggingEnabled.Load() != xrLogLoggingOff
}

// XRLogSetLoggingEnabled turns logging on or off, without affecting logging
// configuration.
//
// loggingEnabled is the flag whether logging is enabled or not; if false, all
// logging calls fail silently. Corresponds to configuration file property
// xr.util-logging.loggingEnabled.
func XRLogSetLoggingEnabled(loggingEnabled bool) {
	if loggingEnabled {
		xrLogLoggingEnabled.Store(xrLogLoggingOn)
	} else {
		xrLogLoggingEnabled.Store(xrLogLoggingOff)
	}
}

func XRLogGetLoggerImpl() XRLogger {
	return xrLogLoggerImpl
}

func XRLogSetLoggerImpl(loggerImpl XRLogger) {
	if loggerImpl == nil {
		panic(NewXRRuntimeException("loggerImpl must not be nil"))
	}
	xrLogLoggerImpl = loggerImpl
}
