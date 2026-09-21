// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/util/ConfigurationTest.java

package ufo

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConfigurationStringValue(t *testing.T) {
	if got := ConfigurationValueForWithDefaultVal("xr.css.user-agent-default-css", "the-default"); got != "/resources/css/" {
		t.Errorf("got %q", got)
	}
	if got := ConfigurationValueForWithDefaultVal("xr.css.user-agent-default-CSS", "the-default"); got != "the-default" {
		t.Errorf("got %q", got)
	}
}

func TestConfigurationByteValue(t *testing.T) {
	if got := ConfigurationValueAsByte("xr.test-config-byte", 15); got != 8 {
		t.Errorf("got %d", got)
	}
	if got := ConfigurationValueAsByte("xr.test-config-BYTE", 15); got != 15 {
		t.Errorf("got %d", got)
	}
}

func TestConfigurationShortValue(t *testing.T) {
	if got := ConfigurationValueAsShort("xr.test-config-short", 20); got != 16 {
		t.Errorf("got %d", got)
	}
	if got := ConfigurationValueAsShort("xr.test-config-SHORT", 20); got != 20 {
		t.Errorf("got %d", got)
	}
}

func TestConfigurationIntValue(t *testing.T) {
	if got := ConfigurationValueAsInt("xr.test-config-int", 25); got != 100 {
		t.Errorf("got %d", got)
	}
	if got := ConfigurationValueAsInt("xr.test-config-INT", 25); got != 25 {
		t.Errorf("got %d", got)
	}
}

func TestConfigurationLongValue(t *testing.T) {
	if got := ConfigurationValueAsLong("xr.test-config-long", 30); got != 2000 {
		t.Errorf("got %d", got)
	}
	if got := ConfigurationValueAsLong("xr.test-config-LONG", 30); got != 30 {
		t.Errorf("got %d", got)
	}
}

func TestConfigurationFloatValue(t *testing.T) {
	if got := ConfigurationValueAsFloat("xr.test-config-float", 45.5); got != 3000.25 {
		t.Errorf("got %v", got)
	}
	if got := ConfigurationValueAsFloat("xr.test-config-FLOAT", 45.5); got != 45.5 {
		t.Errorf("got %v", got)
	}
}

func TestConfigurationDoubleValue(t *testing.T) {
	if got := ConfigurationValueAsDouble("xr.test-config-double", 50.75); got != 4000.50 {
		t.Errorf("got %v", got)
	}
	if got := ConfigurationValueAsDouble("xr.test-config-DOUBLE", 50.75); got != 50.75 {
		t.Errorf("got %v", got)
	}
}

func TestConfigurationTypes(t *testing.T) {
	if !ConfigurationIsTrue("xr.test-config-boolean", false) {
		t.Errorf("xr.test-config-boolean is not true")
	}
	if ConfigurationIsTrue("xr.test-config-BOOLEAN", false) {
		t.Errorf("xr.test-config-BOOLEAN is true")
	}
}

// The tests below are not in the Java suite.

func TestConfigurationValueForAndKeysByPrefix(t *testing.T) {
	if got := ConfigurationValueFor("xr.image.scale"); got != "LOW" {
		t.Errorf("xr.image.scale = %q", got)
	}
	if got := ConfigurationValueFor("xr.no-such-key"); got != "" {
		t.Errorf("xr.no-such-key = %q", got)
	}
	if got := ConfigurationValueAsChar("xr.renderer.missing-character-replacement", 'x'); got != '#' {
		t.Errorf("xr.renderer.missing-character-replacement = %q", got)
	}
	want := []string{"xr.test-config-boolean", "xr.test-config-byte", "xr.test-config-double", "xr.test-config-float",
		"xr.test-config-int", "xr.test-config-long", "xr.test-config-short"}
	if got := ConfigurationKeysByPrefix("xr.test-config-"); !reflect.DeepEqual(got, want) {
		t.Errorf("keys %v, want %v", got, want)
	}
	if !ConfigurationIsFalse("xr.cache.stylesheets", true) {
		t.Errorf("xr.cache.stylesheets is not false")
	}
	if got := ConfigurationValueFromClassConstant("xr.image.render-quality", "fallback"); got != "fallback" {
		t.Errorf("class constant = %v", got)
	}
}

// TestConfigurationDefaultFileMatchesFallback checks that every fallback
// property has the value of the embedded default file, except the one that
// differs in Java too.
func TestConfigurationDefaultFileMatchesFallback(t *testing.T) {
	c := &Configuration{logLevel: LevelOff}
	defaults := c.loadDefaultProperties()
	for key, value := range c.newFallbackProperties() {
		if key == "xr.renderer.missing-character-replacement" || key == "xr.text.aa-smoothing-level" {
			continue
		}
		if defaults[key] != value {
			t.Errorf("%s: default file has %q, fallback has %q", key, defaults[key], value)
		}
	}
}

func setConfigurationSystemPropertiesForTest(t *testing.T, properties map[string]string) {
	t.Helper()
	configurationSystemPropertiesMu.Lock()
	saved := configurationSystemProperties
	configurationSystemProperties = properties
	configurationSystemPropertiesMu.Unlock()
	t.Cleanup(func() {
		configurationSystemPropertiesMu.Lock()
		configurationSystemProperties = saved
		configurationSystemPropertiesMu.Unlock()
	})
}

// TestConfigurationPrecedence checks the order of the sources: the default
// file, then the override file named by xr.conf, then system properties.
func TestConfigurationPrecedence(t *testing.T) {
	overrideFile := filepath.Join(t.TempDir(), "override.conf")
	err := os.WriteFile(overrideFile, []byte(strings.Join([]string{
		"# override",
		"xr.test-config-int = 7",
		"xr.test-config-long = 8",
		"xr.added-by-override : yes",
	}, "\n")), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	// A home directory without an override file.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("xr.test-config-short", "5")
	setConfigurationSystemPropertiesForTest(t, map[string]string{
		"xr.conf":              overrideFile,
		"xr.test-config-long":  "9",
		"xr.not-in-defaults":   "ignored",
		"xr.added-by-override": "from system property",
	})

	c := newConfiguration()
	tests := map[string]string{
		"xr.test-config-byte":  "8",                    // default file
		"xr.test-config-int":   "7",                    // override file
		"xr.test-config-long":  "9",                    // system property over override file
		"xr.test-config-short": "5",                    // environment variable
		"xr.added-by-override": "from system property", // key added by the override file
	}
	for key, want := range tests {
		if got := c.properties[key]; got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if _, ok := c.properties["xr.not-in-defaults"]; ok {
		t.Errorf("a system property that is not a key of the loaded properties was added")
	}
}

func TestConfigurationUserHomeOverrideFile(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".flyingsaucer"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := os.WriteFile(filepath.Join(home, ".flyingsaucer", "local.xhtmlrenderer.conf"), []byte("xr.test-config-int=11\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	setConfigurationSystemPropertiesForTest(t, map[string]string{})

	c := newConfiguration()
	if got := c.properties["xr.test-config-int"]; got != "11" {
		t.Errorf("xr.test-config-int = %q, want 11", got)
	}
}

type configurationTestLogger struct {
	messages []string
}

func (l *configurationTestLogger) Log(level *Level, msg string) {
	l.messages = append(l.messages, level.GetName()+" "+msg)
}

func TestConfigurationShowConfigQueuesStartupMessages(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	setConfigurationSystemPropertiesForTest(t, map[string]string{"show-config": "INFO"})
	c := newConfiguration()
	if len(c.startupLogRecords) == 0 {
		t.Fatalf("no startup log records")
	}
	first := c.startupLogRecords[0]
	if first.level != LevelInfo || first.msg != "Configuration loaded from resources/conf/xhtmlrenderer.conf" {
		t.Errorf("first record: %s %q", first.level.GetName(), first.msg)
	}
	for _, record := range c.startupLogRecords {
		if record.level.IntValue() < LevelInfo.IntValue() {
			t.Errorf("record below INFO: %s %q", record.level.GetName(), record.msg)
		}
	}
}

func TestConfigurationLoadProperties(t *testing.T) {
	source := "# comment\n" +
		"! another comment\n" +
		"\n" +
		"plain=value\n" +
		"  spaced   =   value with spaces  \n" +
		"colon:value\n" +
		"blank value\n" +
		"empty=\n" +
		"keyonly\n" +
		"continued = first \\\n" +
		"    second\n" +
		"escaped\\ key\\=x = tab\\there\\u0041\\n\n" +
		"latin1 = caf\xe9\n" +
		"astral = \\ud83d\\ude00\n" +
		"windows=line\r\n" +
		"last=no newline"
	got, err := configurationLoadProperties(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"plain":         "value",
		"spaced":        "value with spaces  ",
		"colon":         "value",
		"blank":         "value",
		"empty":         "",
		"keyonly":       "",
		"continued":     "first second",
		"escaped key=x": "tab\there" + "A\n",
		"latin1":        "café",
		"astral":        "\U0001F600",
		"windows":       "line",
		"last":          "no newline",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestConfigurationParseJavaFloat(t *testing.T) {
	valid := map[string]float64{"3000.25F": 3000.25, "4000.50D": 4000.5, " 1e3 ": 1000, "-.5f": -0.5, "7": 7}
	for input, want := range valid {
		got, err := configurationParseJavaFloat(input, 64)
		if err != nil || got != want {
			t.Errorf("%q: got %v, %v, want %v", input, got, err, want)
		}
	}
	for _, input := range []string{"", "abc", "0x10", "1_000", "inf", "1.5FF"} {
		if _, err := configurationParseJavaFloat(input, 64); err == nil {
			t.Errorf("%q: no error", input)
		}
	}
}
