// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/GeneralUtil.java

package ufo

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// generalUtilClasspath holds the files that Flying Saucer loads from its jar
// (flying-saucer-core/src/main/resources). A Java classpath resource name
// "resources/conf/xhtmlrenderer.conf" is the same path in this file system.
//
//go:embed resources
var generalUtilClasspath embed.FS

// generalUtilClasspathEntries are the file systems that stand for the entries
// of a Java class path, searched in order: the embedded core resources first,
// then those added with GeneralUtilAddClasspathEntry.
var (
	generalUtilClasspathMu      sync.RWMutex
	generalUtilClasspathEntries = []fs.FS{generalUtilClasspath}
)

// GeneralUtilAddClasspathEntry appends a file system to the class path that
// "classpath:" URLs and GeneralUtilGetURLFromClasspath resolve against, as
// adding a directory or jar to the Java class path does. Flying Saucer's tests
// load their resources this way.
func GeneralUtilAddClasspathEntry(fsys fs.FS) {
	generalUtilClasspathMu.Lock()
	defer generalUtilClasspathMu.Unlock()
	generalUtilClasspathEntries = append(generalUtilClasspathEntries, fsys)
}

// generalUtilClasspathOpen opens the named file from the first class path
// entry that has it.
func generalUtilClasspathOpen(name string) (fs.File, error) {
	generalUtilClasspathMu.RLock()
	entries := generalUtilClasspathEntries
	generalUtilClasspathMu.RUnlock()
	var firstErr error
	for _, entry := range entries {
		f, err := entry.Open(name)
		if err == nil {
			return f, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return nil, firstErr
}

// generalUtilClasspathScheme is the scheme of the URLs that
// GeneralUtilGetURLFromClasspath returns. Java returns a "jar:" or "file:"
// URL of the resource; the embedded files have no such URL. The IOUtil
// functions that open a URL open this scheme from the embedded files.
const generalUtilClasspathScheme = "classpath"

func generalUtilClasspathName(resource string) string {
	return strings.TrimPrefix(resource, "/")
}

func GeneralUtilCiEquals(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}

// GeneralUtilOpenStreamFromClasspath opens a file embedded from the
// resources directory, or returns nil when there is no such file. The obj
// parameter, whose class loader Java uses, is not used.
func GeneralUtilOpenStreamFromClasspath(obj any, resource string) io.ReadCloser {
	stream, err := generalUtilClasspathOpen(generalUtilClasspathName(resource))
	if err != nil {
		return nil
	}
	return stream
}

// GeneralUtilGetURLFromClasspath returns a URL with the scheme "classpath"
// for a file embedded from the resources directory, or nil when there is no
// such file. The obj parameter, whose class loader Java uses, is not used.
func GeneralUtilGetURLFromClasspath(obj any, resource string) *url.URL {
	name := generalUtilClasspathName(resource)
	f, err := generalUtilClasspathOpen(name)
	if err != nil {
		return nil
	}
	f.Close()
	return &url.URL{Scheme: generalUtilClasspathScheme, Path: "/" + name}
}

// GeneralUtilDumpShortException dumps an exception to the console, only the
// last 5 lines of the stack trace. A Go error carries no stack trace, so the
// lines are the stack of the caller.
func GeneralUtilDumpShortException(ex error) {
	s := ""
	if ex != nil {
		s = ex.Error()
	}
	if s == "" || strings.TrimSpace(s) == "null" {
		s = "{no ex. message}"
	}
	pcs := make([]uintptr, 5)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder
	for {
		frame, more := frames.Next()
		sb.WriteString("  ")
		sb.WriteString(frame.Function)
		sb.WriteString("(ln ")
		sb.WriteString(strconv.Itoa(frame.Line))
		sb.WriteString(")\n")
		if !more {
			break
		}
	}
	log.Printf("%s, %T\n%s", s, ex, sb.String())
}

func GeneralUtilIsMacOSX() bool {
	return runtime.GOOS == "darwin"
}

func GeneralUtilHtmlEscapeSpace(uri string) string {
	var sbURI strings.Builder
	sbURI.Grow(int(float64(len(uri)) * 1.5))
	for _, ch := range uri {
		if ch == ' ' {
			sbURI.WriteString("%20")
		} else if ch == '\\' {
			sbURI.WriteByte('/')
		} else {
			sbURI.WriteRune(ch)
		}
	}
	return sbURI.String()
}

// GeneralUtilParseIntRelaxed parses an integer from a string using less
// restrictive rules about which characters we won't accept. This scavenges
// the supplied string for any numeric character, while dropping all others.
//
// It returns the number represented by the passed string, or 0 if the string
// is empty, white-space only, contains only non-numeric characters, or simply
// evaluates to 0 after parsing (e.g. "0").
func GeneralUtilParseIntRelaxed(s string) int {
	return GeneralUtilParseIntRelaxedWithDefaultValue(s, 0)
}

func GeneralUtilParseIntRelaxedWithDefaultValue(s string, defaultValue int) int {
	// An edge-case short circuit...
	if s == "" || javaStringTrim(s) == "" {
		return defaultValue
	}

	var buffer strings.Builder
	for _, c := range s {
		if unicode.IsDigit(c) {
			// Integer.parseInt reads a digit of any script by its
			// numeric value.
			buffer.WriteByte(byte('0' + generalUtilDigitValue(c)))
		} else {
			// If we hit a non-numeric with numbers already in the
			// buffer, we're done.
			if buffer.Len() > 0 {
				break
			}
		}
	}

	if buffer.Len() == 0 {
		return defaultValue
	}

	value, err := strconv.ParseInt(buffer.String(), 10, 32)
	if err != nil {
		// The only way we get here now is if s > Integer.MAX_VALUE
		return math.MaxInt32
	}
	return int(value)
}

// generalUtilDigitValue is Character.digit(c, 10) for a character of the
// category Nd. The decimal digits of a script are ten consecutive code
// points starting at its zero, and every range of the Nd table starts at a
// zero and holds a multiple of ten code points.
func generalUtilDigitValue(c rune) int {
	for _, r := range unicode.Nd.R16 {
		if c >= rune(r.Lo) && c <= rune(r.Hi) {
			return int(c-rune(r.Lo)) % 10
		}
	}
	for _, r := range unicode.Nd.R32 {
		if c >= rune(r.Lo) && c <= rune(r.Hi) {
			return int(c-rune(r.Lo)) % 10
		}
	}
	panic(NewXRRuntimeException(fmt.Sprintf("not a decimal digit: %q", c)))
}

// javaStringTrim is String.trim(): it removes the leading and trailing
// characters up to U+0020.
func javaStringTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}

// GeneralUtilEscapeHTML converts any special characters into their
// corresponding HTML entities, for example < to &lt;. This is done using a
// character by character test, so you may consider other approaches for
// large documents. Make sure you declare the entities that might appear in
// this replacement, e.g. the latin-1 entities.
// This method was taken from a code-samples website, written and hosted by
// Real Gagnon, at http://www.rgagnon.com/javadetails/java-0306.html.
func GeneralUtilEscapeHTML(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '&':
			sb.WriteString("&amp;")
		case '"':
			sb.WriteString("&quot;")
		// be careful with this one (non-breaking white space)
		case ' ':
			sb.WriteString("&nbsp;")
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}
