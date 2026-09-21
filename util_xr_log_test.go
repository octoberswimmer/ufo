// Tests of util_xr_log.go and util_jdk_xr_logger.go. The Java suite has no
// tests of these classes.

package ufo

import (
	"errors"
	"testing"
)

type xrLogTestRecord struct {
	where string
	level *Level
	msg   string
	th    error
}

type xrLogTestLogger struct {
	records []xrLogTestRecord
}

func (l *xrLogTestLogger) Log(where string, level *Level, msg string) {
	l.records = append(l.records, xrLogTestRecord{where, level, msg, nil})
}

func (l *xrLogTestLogger) LogWithTh(where string, level *Level, msg string, th error) {
	l.records = append(l.records, xrLogTestRecord{where, level, msg, th})
}

func (l *xrLogTestLogger) SetLevel(logger string, level *Level) {}

func TestXRLog_overloads(t *testing.T) {
	previousImpl := XRLogGetLoggerImpl()
	previousEnabled := XRLogIsLoggingEnabled()
	defer func() {
		XRLogSetLoggerImpl(previousImpl)
		XRLogSetLoggingEnabled(previousEnabled)
	}()
	logger := &xrLogTestLogger{}
	XRLogSetLoggerImpl(logger)
	XRLogSetLoggingEnabled(true)

	cause := errors.New("cause")
	XRLogCssParse("a")
	XRLogCssParse(LevelWarning, "b")
	XRLogCssParse(LevelSevere, "c", cause)
	XRLogJunit("d")
	XRLogException("e")
	XRLogLoadWithTh(LevelWarning, "f", cause)

	want := []xrLogTestRecord{
		{"org.xhtmlrenderer.css-parse", LevelInfo, "a", nil},
		{"org.xhtmlrenderer.css-parse", LevelWarning, "b", nil},
		{"org.xhtmlrenderer.css-parse", LevelSevere, "c", cause},
		{"org.xhtmlrenderer.junit", LevelFinest, "d", nil},
		{"org.xhtmlrenderer.exception", LevelWarning, "e", nil},
		{"org.xhtmlrenderer.load", LevelWarning, "f", cause},
	}
	if len(logger.records) != len(want) {
		t.Fatalf("logged %d records, want %d", len(logger.records), len(want))
	}
	for i := range want {
		if logger.records[i] != want[i] {
			t.Errorf("record %d = %v, want %v", i, logger.records[i], want[i])
		}
	}

	XRLogSetLoggingEnabled(false)
	XRLogCssParse(LevelSevere, "not logged")
	if len(logger.records) != len(want) {
		t.Error("a message was logged while logging was disabled")
	}
}

func TestXRLog_listRegisteredLoggers(t *testing.T) {
	loggers := XRLogListRegisteredLoggers()
	if len(loggers) != 12 || loggers[0] != "org.xhtmlrenderer.config" || loggers[11] != "org.xhtmlrenderer.render" {
		t.Errorf("XRLogListRegisteredLoggers() = %v", loggers)
	}
}

func TestJDKXRLogger_levels(t *testing.T) {
	logger := NewJDKXRLogger()
	logger.configured = true
	if logger.isLoggable("org.xhtmlrenderer.load", LevelInfo) {
		t.Error("INFO is below the WARNING threshold of the output and must be discarded")
	}
	if !logger.isLoggable("org.xhtmlrenderer.load", LevelWarning) {
		t.Error("WARNING must be logged")
	}
	logger.SetLevel("org.xhtmlrenderer", LevelSevere)
	if logger.isLoggable("org.xhtmlrenderer.load.xml-entities", LevelWarning) {
		t.Error("the level of a parent logger must apply to a logger without a level")
	}
	logger.SetLevel("org.xhtmlrenderer.load", LevelOff)
	if logger.isLoggable("org.xhtmlrenderer.load", LevelSevere) {
		t.Error("a logger with level OFF must discard every message")
	}
	if LoggerUtilParseLogLevel("FINER", LevelOff) != LevelFiner || LoggerUtilParseLogLevel("bogus", LevelOff) != LevelOff {
		t.Error("LoggerUtilParseLogLevel does not parse level names")
	}
}
