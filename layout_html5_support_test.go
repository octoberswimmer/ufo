// Tests for layout_html5_support.go. Flying Saucer has no JUnit test for
// Html5Support.

package ufo

import (
	"reflect"
	"testing"
)

type html5SupportTestLogger struct {
	messages []string
}

func (l *html5SupportTestLogger) Log(where string, level *Level, msg string) {
	l.messages = append(l.messages, level.String()+" "+msg)
}

func (l *html5SupportTestLogger) LogWithTh(where string, level *Level, msg string, th error) {
	l.Log(where, level, msg)
}

func (l *html5SupportTestLogger) SetLevel(logger string, level *Level) {}

func TestHtml5SupportLogUnsupportedFeatures(t *testing.T) {
	previousImpl := XRLogGetLoggerImpl()
	previousEnabled := XRLogIsLoggingEnabled()
	t.Cleanup(func() {
		XRLogSetLoggerImpl(previousImpl)
		XRLogSetLoggingEnabled(previousEnabled)
	})
	logger := &html5SupportTestLogger{}
	XRLogSetLoggerImpl(logger)
	XRLogSetLoggingEnabled(true)

	Html5SupportLogUnsupportedFeatures(nil, nil)
	if len(logger.messages) != 0 {
		t.Errorf("empty sets logged %q", logger.messages)
	}

	Html5SupportLogUnsupportedFeatures([]string{"nav", "video"}, []string{"display: flex", "resize"})
	want := []string{
		"WARNING Encountered HTML5 elements which are not supported by FlyingSaucer: <nav>, <video>. Rendering may be incorrect.",
		"WARNING Encountered CSS3 features not supported by FlyingSaucer: \"display: flex\", \"resize\". Rendering may be incorrect.",
	}
	if !reflect.DeepEqual(logger.messages, want) {
		t.Errorf("got %q, want %q", logger.messages, want)
	}
}
