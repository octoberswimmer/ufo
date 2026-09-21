// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/swing/NaiveUserAgentTest.java
// (NaiveUserAgent is ported into this package).

package pdf

import (
	"strings"
	"testing"
)

func naiveUserAgentTestResolve(baseUri string, uri string) string {
	userAgent := NewNaiveUserAgent()
	userAgent.SetBaseURL(baseUri)
	return userAgent.ResolveURI(uri)
}

func naiveUserAgentTestAssertResolve(t *testing.T, baseUri string, uri string, expected string) {
	t.Helper()
	if actual := naiveUserAgentTestResolve(baseUri, uri); actual != expected {
		t.Errorf("resolve(%q, %q) = %q, want %q", baseUri, uri, actual, expected)
	}
}

// Java's null base URI is "" in the port.
func TestNaiveUserAgent_basicResolve(t *testing.T) {
	// absolute uris should be unchanged
	naiveUserAgentTestAssertResolve(t, "", "http://www.example.com", "http://www.example.com")
	naiveUserAgentTestAssertResolve(t, "ftp://www.example.com/other", "http://www.example.com", "http://www.example.com")

	// by default relative uris resolves as file
	if actual := naiveUserAgentTestResolve("", "www.example.com"); !strings.HasPrefix(actual, "file:") {
		t.Errorf("resolve(null, %q) = %q, want a file: URL", "www.example.com", actual)
	}

	// relative uris without slash
	naiveUserAgentTestAssertResolve(t, "ftp://www.example.com/other", "test", "ftp://www.example.com/test")

	// relative uris with slash
	naiveUserAgentTestAssertResolve(t, "ftp://www.example.com/other/", "test", "ftp://www.example.com/other/test")
	naiveUserAgentTestAssertResolve(t, "ftp://www.example.com/other/", "/test", "ftp://www.example.com/test")
}

func TestNaiveUserAgent_customProtocolResolve(t *testing.T) {
	// absolute uris should be unchanged
	naiveUserAgentTestAssertResolve(t, "", "custom://www.example.com", "custom://www.example.com")
	naiveUserAgentTestAssertResolve(t, "ftp://www.example.com/other", "custom://www.example.com", "custom://www.example.com")

	// relative uris without slash
	naiveUserAgentTestAssertResolve(t, "custom://www.example.com/other", "test", "custom://www.example.com/test")

	// relative uris with slash
	naiveUserAgentTestAssertResolve(t, "custom://www.example.com/other/", "test", "custom://www.example.com/other/test")
	naiveUserAgentTestAssertResolve(t, "custom://www.example.com/other/", "/test", "custom://www.example.com/test")
}

// This reproduces https://code.google.com/archive/p/flying-saucer/issues/262
//
// Below test was green with 9.0.6 and turned red in 9.0.7
func TestNaiveUserAgent_jarFileUriResolve(t *testing.T) {
	// absolute uris should be unchanged
	naiveUserAgentTestAssertResolve(t, "", "jar:file:/path/jarfile.jar!/foo/index.xhtml", "jar:file:/path/jarfile.jar!/foo/index.xhtml")
	naiveUserAgentTestAssertResolve(t, "ftp://www.example.com/other", "jar:file:/path/jarfile.jar!/foo/index.xhtml", "jar:file:/path/jarfile.jar!/foo/index.xhtml")

	// relative uris without slash
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/index.xhtml", "other.xhtml", "jar:file:/path/jarfile.jar!/foo/other.xhtml")

	// relative uris with slash
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/", "other.xhtml", "jar:file:/path/jarfile.jar!/foo/other.xhtml")
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/", "/other.xhtml", "jar:file:/path/jarfile.jar!/other.xhtml")
}

// Resolution against opaque base URLs beyond the Java test, with the results
// of the Java NaiveUserAgent (JDK 25) for the same input.
func TestNaiveUserAgent_jarFileUriResolve_matchesJava(t *testing.T) {
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/bar/index.xhtml", "../other.xhtml", "jar:file:/path/jarfile.jar!/foo/other.xhtml")
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/bar/index.xhtml", "./x/../y.css", "jar:file:/path/jarfile.jar!/foo/bar/y.css")
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/index.xhtml", "../../up.css", "jar:file:/path/jarfile.jar!/up.css")
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/index.xhtml", "#frag", "jar:file:/path/jarfile.jar!/foo/index.xhtml#frag")
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/index.xhtml#top", "a.css#b", "jar:file:/path/jarfile.jar!/foo/a.css#b")
	naiveUserAgentTestAssertResolve(t, "jar:file:/path/jarfile.jar!/foo/", "..", "jar:file:/path/jarfile.jar!/")
	naiveUserAgentTestAssertResolve(t, "mailto:someone@example.com", "x.css", "mailto:x.css")
}

// Java's class loader returns a "file:" URL of the test resource. The port
// has no class loader: the class path is a list of file systems
// (ufo.GeneralUtilAddClasspathEntry; TestMain adds testdata), and a class
// path URL has the scheme "classpath".
func TestNaiveUserAgent_resolveClasspathUrl(t *testing.T) {
	agent := NewNaiveUserAgent()
	for _, uri := range []string{"classpath:transgrey.png", "classpath:/transgrey.png"} {
		u := agent.ResolveClasspathUrl(uri)
		if u == nil || u.Scheme != "classpath" || !strings.HasSuffix(u.Path, "/transgrey.png") {
			t.Errorf("ResolveClasspathUrl(%q) = %v, want a classpath URL of transgrey.png", uri, u)
		}
	}
}
