// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/Configuration.java

package ufo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"
)

// Configuration stores runtime configuration information for application
// parameters that may vary on restarting. This implements the Singleton
// pattern, but through package functions. That is, the first time
// Configuration is used, the properties are loaded into the Singleton
// instance. Subsequent calls to ConfigurationValueFor retrieve values from
// the Singleton.
//
// Properties may be overridden using a second properties file, or
// individually using system properties. What each Java source of
// configuration became:
//
//   - The default properties file resources/conf/xhtmlrenderer.conf, which
//     Java loads from the classpath (the jar), is the same file embedded in
//     the package (see generalUtilClasspath).
//   - A Java system property (java -Dxr.property-name=new_value) is a value
//     set with ConfigurationSetProperty(key, value) before the first use of
//     Configuration, or else the environment variable of the same name as the
//     property (env xr.property-name=new_value program). Java reads system
//     properties for "show-config", "xr.conf" and every key of the default
//     file that starts with "xr."; the same keys are read here.
//   - The override file named by the system property xr.conf is the file
//     named by that property as described above. It is a file path or a URL.
//   - The override file in the user home directory,
//     ${user.home}/.flyingsaucer/local.xhtmlrenderer.conf, is the same path
//     under os.UserHomeDir. As in Java it is read only when xr.conf is not
//     set.
//
// The order in which these are read is: default properties; override
// configuration properties or else the properties file in the user home
// directory; and system properties.
//
// You can override as many properties as you like.
//
// Note that overrides from system properties are driven by the property
// names in the default configuration file. Specifying a property name not in
// that file will have no effect: the property will not be loaded or available
// for lookup. Configuration is NOT used to control logging levels or output.
//
// There are convenience conversion functions for all the primitive types, in
// functions like ConfigurationValueAsInt. A default must always be provided
// for these functions. The default is returned if the value is not found, or
// if the conversion from string fails. If the value is not present, or the
// conversion fails, a warning message is written to the log.
type Configuration struct {
	// Our backing data store of properties.
	properties map[string]string

	// The log Level for Configuration messages; taken from show-config
	// system property.
	logLevel *Level

	// List of log records for messages from Configuration startup; used to
	// hold these temporarily as we can't use XRLog while starting up, as it
	// depends on Configuration.
	startupLogRecords []configurationLogRecord

	// Logger we use internally related to configuration.
	configLogger ConfigurationLogger
}

// ConfigurationLogger receives the messages of Configuration about its own
// loading. It stands for the java.util.logging.Logger that
// Configuration.setConfigLogger takes.
type ConfigurationLogger interface {
	Log(level *Level, msg string)
}

// configurationLogRecord stands for java.util.logging.LogRecord.
type configurationLogRecord struct {
	level *Level
	msg   string
}

// The location of our default properties file; must be among the embedded
// files.
const configurationSfFileName = "resources/conf/xhtmlrenderer.conf"

var (
	// The Singleton instance of the class. Java creates it when the class is
	// loaded; here it is created on the first use, so that a program can call
	// ConfigurationSetProperty first.
	configurationSInstance     *Configuration
	configurationSInstanceOnce sync.Once

	// configurationSystemProperties holds what ConfigurationSetProperty set.
	configurationSystemProperties   = map[string]string{}
	configurationSystemPropertiesMu sync.Mutex
)

// ConfigurationSetProperty sets a system property, the counterpart of
// System.setProperty and of -Dkey=value on the java command line. As in Java,
// Configuration reads system properties once, when it is first used; a
// property set later has no effect on it.
func ConfigurationSetProperty(key, value string) {
	configurationSystemPropertiesMu.Lock()
	defer configurationSystemPropertiesMu.Unlock()
	configurationSystemProperties[key] = value
}

// newConfiguration will parse default configuration file, system properties,
// override properties, etc. and result in a usable Configuration instance.
//
// It panics if any stage of loading configuration fails. This could happen,
// for example, if the default configuration file was not readable.
func newConfiguration() *Configuration {
	c := &Configuration{logLevel: LevelOff}
	defer func() {
		if r := recover(); r != nil {
			c.handleUnexpectedExceptionOnInit(r)
			panic(r)
		}
	}()

	// read logging level from System properties
	// here we are trying to see if user wants to see logging about
	// what configuration was loaded, e.g. debugging for config itself
	val, ok := c.getSystemProperty("show-config")
	if ok {
		c.logLevel = LoggerUtilParseLogLevel(val, LevelOff)
	}
	c.properties = c.loadDefaultProperties()

	sysOverrideFile, ok := c.getSystemPropertyOverrideFileName()
	if ok {
		c.loadOverrideProperties(sysOverrideFile)
	} else {
		userHomeOverrideFileName, ok := c.getUserHomeOverrideFileName()
		if ok {
			c.loadOverrideProperties(userHomeOverrideFileName)
		}
	}
	c.loadSystemProperties()
	c.logAfterLoad()
	return c
}

func (c *Configuration) handleUnexpectedExceptionOnInit(e any) {
	fmt.Fprintf(os.Stderr, "Could not initialize configuration for Flying Saucer library. Message is: %v\n", e)
	log.Printf("%v", e)
}

// ConfigurationSetConfigLogger sets the logger which we use for
// Configuration-related logging. Before this is called the first time, all
// internal log records are queued up; they are flushed to the logger when
// this method is first called. Afterwards, all log events are written to this
// logger. This queueing behavior helps avoid order-of-operations bugs related
// to loading configuration information related to logging.
func ConfigurationSetConfigLogger(logger ConfigurationLogger) {
	config := configurationInstance()
	config.configLogger = logger
	for _, lr := range config.startupLogRecords {
		logger.Log(lr.level, lr.msg)
	}
	config.startupLogRecords = nil
}

// println is used internally for logging status/info about the class.
func (c *Configuration) println(level *Level, msg string) {
	if c.logLevel != LevelOff {
		if c.configLogger == nil {
			c.startupLogRecords = append(c.startupLogRecords, configurationLogRecord{level: level, msg: msg})
		} else {
			c.configLogger.Log(level, msg)
		}
	}
}

// info is used internally to log a message about the class at level INFO.
func (c *Configuration) info(msg string) {
	if c.logLevel.IntValue() <= LevelInfo.IntValue() {
		c.println(LevelInfo, msg)
	}
}

// warning is used internally to log a message about the class at level
// WARNING.
func (c *Configuration) warning(msg string) {
	if c.logLevel.IntValue() <= LevelWarning.IntValue() {
		c.println(LevelWarning, msg)
	}
}

// warningWithTh is used internally to log a message about the class at level
// WARNING, in case an error was raised.
func (c *Configuration) warningWithTh(msg string, th error) {
	c.warning(msg)
	log.Printf("%s: %v", msg, th)
}

// fine is used internally to log a message about the class at level FINE.
func (c *Configuration) fine(msg string) {
	if c.logLevel.IntValue() <= LevelFine.IntValue() {
		c.println(LevelFine, msg)
	}
}

// finer is used internally to log a message about the class at level FINER.
func (c *Configuration) finer(msg string) {
	if c.logLevel.IntValue() <= LevelFiner.IntValue() {
		c.println(LevelFiner, msg)
	}
}

// loadDefaultProperties loads the default set of properties, which may be
// overridden.
func (c *Configuration) loadDefaultProperties() map[string]string {
	readStream := GeneralUtilOpenStreamFromClasspath(nil, configurationSfFileName)
	if readStream == nil {
		log.Printf("No configuration files found in classpath using URL %s, resorting to hard-coded fallback properties.", configurationSfFileName)
		return c.newFallbackProperties()
	}
	defer readStream.Close()
	properties, err := configurationLoadProperties(readStream)
	if err != nil {
		// Java throws a plain RuntimeException. NewXRRuntimeException is
		// not used because it logs through XRLog, which reads the
		// Configuration that is being created here.
		panic(fmt.Errorf("Could not load properties file for configuration.: %w", err))
	}
	c.info("Configuration loaded from " + configurationSfFileName)
	return properties
}

// loadOverrideProperties loads overriding property values from a second
// configuration file; this is optional. See class documentation.
//
// uri is the path to the file, or the URL, where properties are defined.
func (c *Configuration) loadOverrideProperties(uri string) {
	var temp map[string]string
	if _, statErr := os.Stat(uri); statErr == nil {
		absolutePath, err := filepath.Abs(uri)
		if err != nil {
			absolutePath = uri
		}
		c.info("Found config override file " + absolutePath)
		readStream, err := os.Open(uri)
		if err == nil {
			temp, err = configurationLoadProperties(bufio.NewReader(readStream))
			readStream.Close()
		}
		if err != nil {
			c.warningWithTh("Error while loading override properties file; skipping.", err)
			return
		}
	} else {
		in, err := ioUtilOpenURL(uri, 0, 0, "")
		if err == nil {
			c.info("Found config override URI " + uri)
			temp, err = configurationLoadProperties(bufio.NewReader(in))
			in.Close()
		}
		var malformed *ioUtilMalformedURLError
		if errors.As(err, &malformed) {
			c.warning(fmt.Sprintf("URI for override properties is malformed, skipping: '%s' (caused by: %s)", uri, err))
			return
		}
		if err != nil {
			c.warningWithTh("Overridden properties could not be loaded from URI: "+uri, err)
			return
		}
	}

	// override existing properties
	cnt := 0
	for _, key := range configurationSortedKeys(c.properties) {
		val, ok := temp[key]
		if ok {
			c.properties[key] = val
			c.finer("  " + key + " -> " + val)
			cnt++
		}
	}
	c.finer("Configuration: " + strconv.Itoa(cnt) + " properties overridden from secondary properties file.")
	// and add any new properties we don't already know about (needed for custom logging
	// configuration)
	allRead := configurationSortedKeys(temp)

	cnt = 0
	for _, key := range allRead {
		val, ok := temp[key]
		if ok {
			c.properties[key] = val
			c.finer("  (+)" + key + " -> " + val)
			cnt++
		}
	}
	c.finer("Configuration: " + strconv.Itoa(cnt) + " properties added from secondary properties file.")
}

func (c *Configuration) getSystemPropertyOverrideFileName() (string, bool) {
	return c.getSystemProperty("xr.conf")
}

func (c *Configuration) getUserHomeOverrideFileName() (string, bool) {
	confFileName := "local.xhtmlrenderer.conf"
	home, err := os.UserHomeDir()
	if err != nil {
		// can happen in a sandbox
		log.Printf("Cannot read file '%s': %v", confFileName, err)
		return "", false
	}
	return filepath.Join(home, ".flyingsaucer", confFileName), true
}

func configurationSortedKeys(properties map[string]string) []string {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// loadSystemProperties loads overriding property values from system
// properties; this is optional. See class documentation.
func (c *Configuration) loadSystemProperties() {
	c.fine("Overriding loaded configuration from System properties.")
	cnt := 0
	for _, key := range configurationSortedKeys(c.properties) {
		if !strings.HasPrefix(key, "xr.") {
			continue
		}

		val, ok := c.getSystemProperty(key)
		if ok {
			c.properties[key] = val
			c.finer("  Overrode value for " + key)
			cnt++
		}
	}
	c.fine("Configuration: " + strconv.Itoa(cnt) + " properties overridden from System properties.")

	// add any additional properties we don't already know about (e.g. used for extended logging properties)
	//
	// The Java loop walks the keys of the loaded properties, not the keys of
	// the system properties, and skips every key the loaded properties
	// contain, so it adds nothing. The port keeps that: a system property
	// whose name is not a key of the loaded properties is not read.
	cnt = 0
	for _, key := range configurationSortedKeys(c.properties) {
		if _, contains := c.properties[key]; strings.HasPrefix(key, "xr.") && !contains {
			val, _ := c.getSystemProperty(key)
			c.properties[key] = val
			c.finer("  (+) " + key)
			cnt++
		}
	}
	c.fine("Configuration: " + strconv.Itoa(cnt) + " FS properties added from System properties.")
}

// logAfterLoad writes a log of loaded properties to the configuration logger.
func (c *Configuration) logAfterLoad() {
	c.finer("Configuration contains " + strconv.Itoa(len(c.properties)) + " keys.")
	c.finer("List of configuration properties, after override:")
	for _, key := range configurationSortedKeys(c.properties) {
		val := c.properties[key]
		c.finer("  " + key + " = " + val)
	}
	c.finer("Properties list complete.")
}

// ConfigurationValueFor returns the value for key in the Configuration. A
// warning is issued to the log if the property is not defined, and the
// result is "", which stands for the Java null.
func ConfigurationValueFor(key string) string {
	val, _ := configurationLookup(key)
	return val
}

// configurationLookup is Configuration.valueFor(String) with the distinction
// between a missing property and an empty one.
func configurationLookup(key string) (string, bool) {
	conf := configurationInstance()
	val, ok := conf.properties[key]
	if !ok {
		conf.warning("CONFIGURATION: no value found for key " + key)
	}
	return val, ok
}

// ConfigurationValueAsByte returns the value for key in the Configuration as
// a byte, or the default provided value if not found or if the value is not a
// valid byte. A warning is issued to the log if the property is not defined,
// or if the conversion from String fails.
func ConfigurationValueAsByte(key string, defaultVal int8) int {
	val, ok := configurationLookup(key)
	if !ok {
		return int(defaultVal)
	}

	parsed, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		XRLogException("Property '" + key + "' was requested as a byte, but " +
			"value of '" + val + "' is not a byte. Check configuration.")
		return int(defaultVal)
	}
	return int(parsed)
}

// ConfigurationValueAsShort returns the value for key in the Configuration as
// a short, or the default provided value if not found or if the value is not
// a valid short. A warning is issued to the log if the property is not
// defined, or if the conversion from String fails.
func ConfigurationValueAsShort(key string, defaultVal int16) int {
	val, ok := configurationLookup(key)
	if !ok {
		return int(defaultVal)
	}

	parsed, err := strconv.ParseInt(val, 10, 16)
	if err != nil {
		XRLogException("Property '" + key + "' was requested as a short, but " +
			"value of '" + val + "' is not a short. Check configuration.")
		return int(defaultVal)
	}
	return int(parsed)
}

// ConfigurationValueAsInt returns the value for key in the Configuration as
// an integer, or a default value if not found or if the value is not a valid
// integer. A warning is issued to the log if the property is not defined, or
// if the conversion from String fails.
func ConfigurationValueAsInt(key string, defaultVal int) int {
	val, ok := configurationLookup(key)
	if !ok {
		return defaultVal
	}

	parsed, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		XRLogException("Property '" + key + "' was requested as an integer, but " +
			"value of '" + val + "' is not an integer. Check configuration.")
		return defaultVal
	}
	return int(parsed)
}

// ConfigurationValueAsChar returns the value for key in the Configuration as
// a character, or a default value if not found. A warning is issued to the
// log if the property is not defined, or if the configuration value is too
// long to be a char. If the configuration value is longer than a single
// character, only the first character is returned.
func ConfigurationValueAsChar(key string, defaultVal rune) rune {
	val, ok := configurationLookup(key)
	if !ok {
		return defaultVal
	}

	runes := []rune(val)
	if len(runes) > 1 {
		XRLogException("Property '" + key + "' was requested as a character. The value of '" +
			val + "' is too long to be a char. Returning only the first character.")
	}

	if len(runes) == 0 {
		// String.charAt(0) throws StringIndexOutOfBoundsException.
		panic(NewXRRuntimeException("Property '" + key + "' was requested as a character, but its value is empty."))
	}
	return runes[0]
}

// ConfigurationValueAsLong returns the value for key in the Configurations a
// long, or the default provided value if not found or if the value is not a
// valid long. A warning is issued to the log if the property is not defined,
// or if the conversion from String fails.
func ConfigurationValueAsLong(key string, defaultVal int64) int64 {
	val, ok := configurationLookup(key)
	if !ok {
		return defaultVal
	}

	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		XRLogException("Property '" + key + "' was requested as a long, but " +
			"value of '" + val + "' is not a long. Check configuration.")
		return defaultVal
	}
	return parsed
}

// ConfigurationValueAsFloat returns the value for key in the Configuration as
// a float, or the default provided value if not found or if the value is not
// a valid float. A warning is issued to the log if the property is not
// defined, or if the conversion from String fails.
func ConfigurationValueAsFloat(key string, defaultVal float32) float32 {
	val, ok := configurationLookup(key)
	if !ok {
		return defaultVal
	}

	parsed, err := configurationParseJavaFloat(val, 32)
	if err != nil {
		XRLogException("Property '" + key + "' was requested as a float, but " +
			"value of '" + val + "' is not a float. Check configuration.")
		return defaultVal
	}
	return float32(parsed)
}

// ConfigurationValueAsDouble returns the value for key in the Configuration
// as a double, or the default provided value if not found or if the value is
// not a valid double. A warning is issued to the log if the property is not
// defined, or if the conversion from String fails.
func ConfigurationValueAsDouble(key string, defaultVal float64) float64 {
	val, ok := configurationLookup(key)
	if !ok {
		return defaultVal
	}

	parsed, err := configurationParseJavaFloat(val, 64)
	if err != nil {
		XRLogException("Property '" + key + "' was requested as a double, but " +
			"value of '" + val + "' is not a double. Check configuration.")
		return defaultVal
	}
	return parsed
}

// configurationParseJavaFloat is Float.parseFloat and Double.parseDouble for
// decimal input: surrounding whitespace is dropped, and the number may end
// with a type suffix (f, F, d or D), as the values "3000.25F" and "4000.50D"
// of the default configuration do.
func configurationParseJavaFloat(val string, bitSize int) (float64, error) {
	s := javaStringTrim(val)
	switch s {
	case "NaN", "+NaN", "-NaN", "Infinity", "+Infinity", "-Infinity":
		return strconv.ParseFloat(s, bitSize)
	}
	if s != "" {
		switch s[len(s)-1] {
		case 'f', 'F', 'd', 'D':
			s = s[:len(s)-1]
		}
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		// strconv also reads "inf", "nan", hexadecimal digits and
		// underscores, which the decimal form of Java does not have.
		if !(ch >= '0' && ch <= '9') && ch != '.' && ch != 'e' && ch != 'E' && ch != '+' && ch != '-' {
			return 0, fmt.Errorf("For input string: %q", val)
		}
	}
	parsed, err := strconv.ParseFloat(s, bitSize)
	var numError *strconv.NumError
	if errors.As(err, &numError) && numError.Err == strconv.ErrRange {
		// Java returns an infinity or a zero for a value out of range.
		return parsed, nil
	}
	return parsed, err
}

// ConfigurationValueForWithDefaultVal returns the value for key in the
// Configuration, or the default provided value if not found. A warning is
// issued to the log if the property is not defined, and if the default is
// empty.
func ConfigurationValueForWithDefaultVal(key string, defaultVal string) string {
	conf := configurationInstance()
	val, ok := conf.properties[key]
	if !ok {
		val = defaultVal
		if val == "" {
			conf.warning("CONFIGURATION: no value found for key " + key + " and no default given.")
		}
	}
	return val
}

// ConfigurationKeysByPrefix returns all configuration keys that start with
// prefix, sorted. The result is empty if no such keys are found.
func ConfigurationKeysByPrefix(prefix string) []string {
	conf := configurationInstance()
	result := []string{}
	for _, key := range configurationSortedKeys(conf.properties) {
		if strings.HasPrefix(key, prefix) {
			result = append(result, key)
		}
	}
	return result
}

// ConfigurationIsTrue returns true if the value is "true" (ignores case), or
// the default provided value if not found or if the value is not a valid
// boolean (true or false, ignores case). A warning is issued to the log if
// the property is not defined, and if the default is null.
func ConfigurationIsTrue(key string, defaultVal bool) bool {
	val, ok := configurationLookup(key)
	if !ok {
		return defaultVal
	}

	if !strings.Contains("true|false", val) {
		XRLogException("Property '" + key + "' was requested as a boolean, but " +
			"value of '" + val + "' is not a boolean. Check configuration.")
		return defaultVal
	}
	return strings.EqualFold(val, "true")
}

// ConfigurationIsFalse returns true if the value is not "true" (ignores
// case), or the default provided value if not found or if the value is not a
// valid boolean (true or false, ignores case). A warning is issued to the log
// if the property is not defined, or the value is not a valid boolean.
func ConfigurationIsFalse(key string, defaultVal bool) bool {
	return !ConfigurationIsTrue(key, defaultVal)
}

// configurationInstance returns the singleton instance of the class.
func configurationInstance() *Configuration {
	configurationSInstanceOnce.Do(func() {
		configurationSInstance = newConfiguration()
	})
	return configurationSInstance
}

// ConfigurationValueFromClassConstant, given a property, resolves the value
// to a public constant field on some class. The property value must be the
// fully qualified name of the class and field, e.g.
// aKey=java.awt.RenderingHints.VALUE_INTERPOLATION_NEAREST_NEIGHBOR.
//
// Java finds the class and the field by reflection. The values of the default
// configuration name constants of java.awt.RenderingHints, which have no
// counterpart here, and Go cannot look a package variable up by name, so every
// class name is one that Class.forName does not find: the function logs the
// warning of that case and returns defaultValue.
func ConfigurationValueFromClassConstant(key string, defaultValue any) any {
	conf := configurationInstance()
	val, ok := configurationLookup(key)
	if !ok {
		return defaultValue
	}
	idx := strings.LastIndex(val, ".")
	if idx < 0 {
		conf.warning("Property key " + key + " for object value constant is not properly formatted; " +
			"should be FQN<dot>constant, is " + val)
		return defaultValue
	}
	className := val[:idx]
	conf.warning("Property for object value constant " + key + " is not a FQN: " + className +
		", caused by: java.lang.ClassNotFoundException: " + className)
	return defaultValue
}

// newFallbackProperties returns properties filled with values of last
// resort--in case we can't read default properties file for some reason; this
// is to prevent Configuration init from throwing any exceptions, or ending up
// with a completely empty configuration instance.
func (c *Configuration) newFallbackProperties() map[string]string {
	props := map[string]string{}
	props["xr.css.user-agent-default-css"] = "/resources/css/"
	props["xr.test.files.hamlet"] = "/demos/browser/xhtml/hamlet.xhtml"
	props["xr.simple-log-format"] = "{1} {2}:: {5}"
	props["xr.simple-log-format-throwable"] = "{1} {2}:: {5}"
	props["xr.test-config-byte"] = "8"
	props["xr.test-config-short"] = "16"
	props["xr.test-config-int"] = "100"
	props["xr.test-config-long"] = "2000"
	props["xr.test-config-float"] = "3000.25F"
	props["xr.test-config-double"] = "4000.50D"
	props["xr.test-config-boolean"] = "true"
	props["xr.util-logging.loggingEnabled"] = "false"
	props["xr.util-logging.handlers"] = "java.util.logging.ConsoleHandler"
	props["xr.util-logging.use-parent-handler"] = "false"
	props["xr.util-logging.java.util.logging.ConsoleHandler.level"] = "INFO"
	props["xr.util-logging.java.util.logging.ConsoleHandler.formatter"] = "org.xhtmlrenderer.util.XRSimpleLogFormatter"
	props["xr.util-logging.org.xhtmlrenderer.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.config.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.exception.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.general.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.init.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.load.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.load.xml-entities.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.match.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.cascade.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.css-parse.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.layout.level"] = "ALL"
	props["xr.util-logging.org.xhtmlrenderer.render.level"] = "ALL"
	props["xr.load.xml-reader"] = "default"
	props["xr.load.configure-features"] = "false"
	props["xr.load.validation"] = "false"
	props["xr.load.string-interning"] = "false"
	props["xr.load.namespaces"] = "false"
	props["xr.load.namespace-prefixes"] = "false"
	props["xr.layout.whitespace.experimental"] = "true"
	props["xr.layout.bad-sizing-hack"] = "false"
	props["xr.renderer.viewport-repaint"] = "true"
	props["xr.renderer.draw.backgrounds"] = "true"
	props["xr.renderer.draw.borders"] = "true"
	props["xr.renderer.debug.box-outlines"] = "false"
	props["xr.renderer.replace-missing-characters"] = "false"
	props["xr.renderer.missing-character-replacement"] = "false"
	props["xr.text.scale"] = "1.0"
	props["xr.text.aa-smoothing-level"] = "1"
	props["xr.text.aa-fontsize-threshhold"] = "0"
	props["xr.text.aa-rendering-hint"] = "RenderingHints.VALUE_TEXT_ANTIALIAS_HGRB"
	props["xr.cache.stylesheets"] = "false"
	props["xr.incremental.enabled"] = "false"
	props["xr.incremental.lazyimage"] = "false"
	props["xr.incremental.debug.layoutdelay"] = "0"
	props["xr.incremental.repaint.print-timing"] = "false"
	props["xr.use.threads"] = "false"
	props["xr.use.listeners"] = "true"
	props["xr.image.buffered"] = "false"
	props["xr.image.scale"] = "LOW"
	props["xr.image.render-quality"] = "java.awt.RenderingHints.VALUE_INTERPOLATION_NEAREST_NEIGHBOR"
	return props
}

// getSystemProperty returns the system property of the given name: the value
// set with ConfigurationSetProperty, or else the environment variable of
// that name. The second result is false where Java returns null.
func (c *Configuration) getSystemProperty(name string) (string, bool) {
	configurationSystemPropertiesMu.Lock()
	val, ok := configurationSystemProperties[name]
	configurationSystemPropertiesMu.Unlock()
	if ok {
		return val, true
	}
	return os.LookupEnv(name)
}

// configurationLoadProperties reads the format of java.util.Properties.load:
// one "key=value", "key:value" or "key value" per logical line; a line that
// ends with an odd number of backslashes continues on the next line; a line
// whose first non-blank character is '#' or '!' is a comment; backslash
// escapes \t, \n, \r, \f and \uXXXX; the bytes are ISO 8859-1.
func configurationLoadProperties(r io.Reader) (map[string]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// ISO 8859-1: each byte is the code point of that value.
	chars := make([]rune, len(data))
	for i, b := range data {
		chars[i] = rune(b)
	}

	properties := map[string]string{}
	for _, line := range configurationLogicalLines(chars) {
		// The key ends at the first unescaped '=', ':' or blank.
		keyLen := 0
		valueStart := len(line)
		hasSep := false
		precedingBackslash := false
		for keyLen < len(line) {
			ch := line[keyLen]
			if (ch == '=' || ch == ':') && !precedingBackslash {
				valueStart = keyLen + 1
				hasSep = true
				break
			} else if configurationIsBlank(ch) && !precedingBackslash {
				valueStart = keyLen + 1
				break
			}
			if ch == '\\' {
				precedingBackslash = !precedingBackslash
			} else {
				precedingBackslash = false
			}
			keyLen++
		}
		// Blanks before the value are dropped, and so is one '=' or ':'
		// among them when the key ended at a blank.
		for valueStart < len(line) {
			ch := line[valueStart]
			if !configurationIsBlank(ch) {
				if !hasSep && (ch == '=' || ch == ':') {
					hasSep = true
				} else {
					break
				}
			}
			valueStart++
		}
		key, err := configurationLoadConvert(line[:keyLen])
		if err != nil {
			return nil, err
		}
		value, err := configurationLoadConvert(line[valueStart:])
		if err != nil {
			return nil, err
		}
		properties[key] = value
	}
	return properties, nil
}

func configurationIsBlank(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\f'
}

// configurationLogicalLines splits the text into logical lines: comment
// lines and blank lines are dropped, leading blanks are dropped, and a line
// continued with a trailing backslash is joined with the next one, whose
// leading blanks are dropped.
func configurationLogicalLines(chars []rune) [][]rune {
	var lines [][]rune
	var current []rune
	skipWhiteSpace := true
	isNewLine := true
	isCommentLine := false
	appendedLineBegin := false
	precedingBackslash := false
	flush := func() {
		if len(current) > 0 {
			lines = append(lines, current)
		}
		current = nil
	}
	for i := 0; i < len(chars); i++ {
		ch := chars[i]
		if skipWhiteSpace {
			if configurationIsBlank(ch) {
				continue
			}
			if !appendedLineBegin && (ch == '\r' || ch == '\n') {
				continue
			}
			skipWhiteSpace = false
			appendedLineBegin = false
		}
		if isNewLine {
			isNewLine = false
			if ch == '#' || ch == '!' {
				isCommentLine = true
			}
		}
		if isCommentLine {
			if ch == '\r' || ch == '\n' {
				isCommentLine = false
				isNewLine = true
				skipWhiteSpace = true
			}
			continue
		}
		if ch != '\n' && ch != '\r' {
			current = append(current, ch)
			if ch == '\\' {
				precedingBackslash = !precedingBackslash
			} else {
				precedingBackslash = false
			}
			continue
		}
		// End of a natural line.
		if precedingBackslash {
			// Drop the backslash and continue with the next line.
			current = current[:len(current)-1]
			skipWhiteSpace = true
			appendedLineBegin = true
			precedingBackslash = false
			if ch == '\r' && i+1 < len(chars) && chars[i+1] == '\n' {
				i++
			}
			continue
		}
		flush()
		isNewLine = true
		skipWhiteSpace = true
	}
	if precedingBackslash && len(current) > 0 {
		current = current[:len(current)-1]
	}
	flush()
	return lines
}

// configurationLoadConvert replaces the backslash escapes of a key or a
// value.
func configurationLoadConvert(in []rune) (string, error) {
	var out strings.Builder
	// units collects consecutive \uXXXX escapes, which are UTF-16 code units.
	var units []uint16
	flushUnits := func() {
		if len(units) > 0 {
			out.WriteString(string(utf16.Decode(units)))
			units = nil
		}
	}
	for i := 0; i < len(in); i++ {
		ch := in[i]
		if !(ch == '\\' && i+1 < len(in) && in[i+1] == 'u') {
			flushUnits()
		}
		if ch != '\\' || i+1 >= len(in) {
			if ch != '\\' {
				out.WriteRune(ch)
			}
			continue
		}
		i++
		ch = in[i]
		switch ch {
		case 'u':
			if i+4 >= len(in) {
				return "", errors.New("Malformed \\uxxxx encoding.")
			}
			value, err := strconv.ParseUint(string(in[i+1:i+5]), 16, 16)
			if err != nil {
				return "", errors.New("Malformed \\uxxxx encoding.")
			}
			i += 4
			// A following \uXXXX may be the second half of a surrogate
			// pair, so the units are decoded together.
			units = append(units, uint16(value))
		case 't':
			out.WriteByte('\t')
		case 'r':
			out.WriteByte('\r')
		case 'n':
			out.WriteByte('\n')
		case 'f':
			out.WriteByte('\f')
		default:
			out.WriteRune(ch)
		}
	}
	flushUnits()
	return out.String(), nil
}
