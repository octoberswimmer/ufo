// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/swing/NaiveUserAgent.java

package pdf

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/octoberswimmer/ufo"
)

const naiveUserAgentDefaultImageCacheSize = 16

var naiveUserAgentClasspathPrefix = regexp.MustCompile("classpath:/?")

// NaiveUserAgentI lists the non-private methods of NaiveUserAgent, which
// ITextUserAgent extends.
type NaiveUserAgentI interface {
	ufo.UserAgentCallback
	ShrinkImageCache()
	ClearImageCache()
	ResolveAndOpenStream(uri string) io.ReadCloser
	OpenStream(uri string) (io.ReadCloser, error)
	NeedsRedirect(status int) bool
	CreateImageResource(uri string, img image.Image) *ufo.ImageResource
	ResolveClasspathUrl(uri string) *url.URL
	DocumentStarted()
	DocumentLoaded()
	OnLayoutException(t error)
	OnRenderException(t error)
	AsNaiveUserAgent() *NaiveUserAgent
}

// NaiveUserAgent is a simple implementation of ufo.UserAgentCallback which
// places no restrictions on what XML, CSS or images are loaded, and reports
// visited links without any filtering. The most straightforward process
// available is used to load the resources in question--either using os or
// net/http.
//
// The NaiveUserAgent has a small cache for images, the size of which (number
// of images) can be passed as a constructor argument. There is no automatic
// cleaning of the cache; call ShrinkImageCache to remove the least-accessed
// elements--for example, you might do this when a new document is about to be
// loaded. The NaiveUserAgent also has the methods of Java's DocumentListener
// (the event package is not ported); DocumentStarted attempts to shrink its
// cache.
//
// This class is meant as a starting point--it will work out of the box, but
// you should really implement your own, tuned to your application's needs.
//
// The class is in Flying Saucer's swing package, which is not ported. It is
// ported here, in the package of its one subclass, ITextUserAgent. What
// depends on AWT is left out: CreateImageResource can't wrap a decoded image
// (AWTFSImage is not ported), and OpenConnection and OnHttpConnection have
// no counterpart because net/http follows redirects itself.
//
// Author: Torbjoern Gannholm
type NaiveUserAgent struct {
	// self is the outermost object, the receiver of calls to the methods a
	// subclass may override.
	self NaiveUserAgentI

	// imageCache is a (simple) LRU cache
	imageCache         *naiveUserAgentImageCache
	imageCacheCapacity int
	// baseURL is "" where Java has null.
	baseURL string
}

var _ NaiveUserAgentI = (*NaiveUserAgent)(nil)

// naiveUserAgentImageCache stands for the access-ordered LinkedHashMap that
// Java uses: keys holds the keys from the least recently to the most recently
// accessed.
type naiveUserAgentImageCache struct {
	entries map[string]*ufo.ImageResource
	keys    []string
}

func newNaiveUserAgentImageCache() *naiveUserAgentImageCache {
	return &naiveUserAgentImageCache{entries: map[string]*ufo.ImageResource{}}
}

func (c *naiveUserAgentImageCache) touch(key string) {
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
	c.keys = append(c.keys, key)
}

// get returns nil when the key is absent or was put with a nil resource.
func (c *naiveUserAgentImageCache) get(key string) *ufo.ImageResource {
	resource, ok := c.entries[key]
	if !ok {
		return nil
	}
	c.touch(key)
	return resource
}

func (c *naiveUserAgentImageCache) put(key string, resource *ufo.ImageResource) {
	c.entries[key] = resource
	c.touch(key)
}

func (c *naiveUserAgentImageCache) size() int {
	return len(c.keys)
}

func (c *naiveUserAgentImageCache) removeEldest() {
	if len(c.keys) == 0 {
		return
	}
	delete(c.entries, c.keys[0])
	c.keys = c.keys[1:]
}

func (c *naiveUserAgentImageCache) clear() {
	c.entries = map[string]*ufo.ImageResource{}
	c.keys = nil
}

// NewNaiveUserAgent creates a new instance of NaiveUserAgent with a max image
// cache of 16 images.
func NewNaiveUserAgent() *NaiveUserAgent {
	return NewNaiveUserAgentWithImgCacheSize(naiveUserAgentDefaultImageCacheSize)
}

// NewNaiveUserAgentWithImgCacheSize creates a new NaiveUserAgent with a cache
// of a specific size.
//
// imgCacheSize is the number of images to hold in cache before LRU images are
// released.
func NewNaiveUserAgentWithImgCacheSize(imgCacheSize int) *NaiveUserAgent {
	a := &NaiveUserAgent{}
	initNaiveUserAgent(a, imgCacheSize)
	a.self = a
	return a
}

// initNaiveUserAgent initializes the fields of a; constructors of subclasses
// call it and then SetSelf.
func initNaiveUserAgent(a *NaiveUserAgent, imgCacheSize int) {
	a.imageCacheCapacity = imgCacheSize

	// note we do *not* evict in put()--users of this class must call ShrinkImageCache().
	// that's because we don't know when is a good time to flush the cache
	a.imageCache = newNaiveUserAgentImageCache()
}

// SetSelf sets the outermost object.
func (a *NaiveUserAgent) SetSelf(self NaiveUserAgentI) {
	a.self = self
}

func (a *NaiveUserAgent) AsNaiveUserAgent() *NaiveUserAgent {
	return a
}

// ShrinkImageCache drops the least-recently used items, if the image cache
// has more items than the limit specified for this class, until it reaches
// the desired size.
func (a *NaiveUserAgent) ShrinkImageCache() {
	ovr := a.imageCache.size() - a.imageCacheCapacity
	for a.imageCache.size() > 0 && ovr > 0 {
		ovr--
		a.imageCache.removeEldest()
	}
}

// ClearImageCache empties the image cache entirely.
func (a *NaiveUserAgent) ClearImageCache() {
	a.imageCache.clear()
}

// ResolveAndOpenStream gets a reader for the resource identified, or nil
// when it can't be opened. It is protected in Java.
func (a *NaiveUserAgent) ResolveAndOpenStream(uri string) io.ReadCloser {
	resolvedUri := a.self.ResolveURI(uri)
	var stream io.ReadCloser
	var err error
	if ufo.FontUtilIsEmbeddedBase64Font(uri) {
		if data := ufo.FontUtilGetEmbeddedBase64Data(uri); data != nil {
			stream = io.NopCloser(data)
		}
	} else {
		stream, err = a.self.OpenStream(resolvedUri)
	}
	if err != nil {
		// A MalformedURLException in Java: the URL can't be parsed or has no
		// scheme.
		if parsed, parseErr := url.Parse(resolvedUri); parseErr != nil || parsed.Scheme == "" {
			ufo.XRLogExceptionWithTh("bad URL given: "+resolvedUri, err)
		} else if errors.Is(err, fs.ErrNotExist) {
			ufo.XRLogException("item at URI " + resolvedUri + " not found: " + err.Error())
		} else {
			ufo.XRLogExceptionWithTh("IO problem for "+resolvedUri, err)
		}
		return nil
	}
	return stream
}

// OpenStream opens the "file", "http", "https" and "classpath" URLs that the
// core's IOUtil opens, asking for "*/*". It is protected in Java, where it
// calls openConnection(uri).getInputStream(). ufo.IOUtilStreamAtUrl sets a
// connect timeout of 10 seconds and a read timeout of 30 seconds, which
// Java's NaiveUserAgent does not.
func (a *NaiveUserAgent) OpenStream(uri string) (io.ReadCloser, error) {
	return ufo.IOUtilStreamAtUrl(uri)
}

// NeedsRedirect verifies that return code of connection represents a
// redirection. It is protected and final in Java.
//
// It returns true if the return code is a 3xx that names a new location.
func (a *NaiveUserAgent) NeedsRedirect(status int) bool {
	return status == http.StatusFound || status == http.StatusMovedPermanently || status == http.StatusSeeOther
}

// GetCSSResource retrieves the CSS located at the given URI.  It's assumed
// the URI does point to a CSS file--the URI will be accessed, opened, read
// and then passed into the CSS parser. The result is packed up into an
// CSSResource for later consumption.
func (a *NaiveUserAgent) GetCSSResource(uri string) *ufo.CSSResource {
	stream := a.self.ResolveAndOpenStream(uri)
	if stream == nil {
		return ufo.NewCSSResource(nil)
	}
	return ufo.NewCSSResource(stream)
}

// GetImageResource retrieves the image located at the given URI. It's
// assumed the URI does point to an image--the URI will be accessed, opened,
// read and then passed into the image-parsing routines. The result is packed
// up into an ImageResource for later consumption.
func (a *NaiveUserAgent) GetImageResource(imageLocation string) *ufo.ImageResource {
	if ufo.ImageUtilIsEmbeddedBase64Image(imageLocation) {
		return a.self.CreateImageResource("", ufo.ImageUtilLoadEmbeddedBase64Image(imageLocation))
	}

	cached := a.imageCache.get(imageLocation)
	if cached != nil {
		//TODO: check that cached image is still valid
		return cached
	}

	uri := a.self.ResolveURI(imageLocation)
	is := a.self.ResolveAndOpenStream(uri)
	if is != nil {
		defer is.Close()
		img, _, err := image.Decode(is)
		if err == nil && img == nil {
			err = fmt.Errorf("image.Decode() returned nil for URI %s", uri)
		}
		if err == nil {
			ir := a.self.CreateImageResource(uri, img)
			a.imageCache.put(imageLocation, ir)
			return ir
		}
		if errors.Is(err, fs.ErrNotExist) {
			ufo.XRLogException(fmt.Sprintf("Can't read image file; image at URI '%s' not found (caused by: %s)", uri, err))
		} else {
			ufo.XRLogExceptionWithTh(fmt.Sprintf("Can't read image file; unexpected problem for URI '%s'", uri), err)
		}
	}

	return a.self.CreateImageResource(uri, nil)
}

// CreateImageResource is the factory method to generate ImageResources from a
// given Image. May be overridden in subclass. It is protected in Java.
//
// uri is the URI for the image, resolved to an absolute URI. img is the image
// to package; may be nil (for example, if image could not be loaded).
//
// Java wraps the image with AWTFSImageFactory, a Swing class that is not
// ported, so a non-nil image panics here; ITextUserAgent does not reach this
// method.
func (a *NaiveUserAgent) CreateImageResource(uri string, img image.Image) *ufo.ImageResource {
	if img != nil {
		panic(ufo.NewXRRuntimeException("not ported: AWTFSImageFactory.createImage"))
	}
	return ufo.NewImageResource(uri, nil)
}

// GetXMLResource retrieves the XML located at the given URI. It's assumed the
// URI does point to an XML--the URI will be accessed, opened, read and then
// passed into the XML parser configured for Flying Saucer. The result is
// packed up into an XMLResource for later consumption.
func (a *NaiveUserAgent) GetXMLResource(uri string) *ufo.XMLResource {
	inputStream := a.self.ResolveAndOpenStream(uri)
	if inputStream == nil {
		return ufo.XMLResourceLoadInputStream(nil)
	}
	defer inputStream.Close()
	return ufo.XMLResourceLoadInputStream(inputStream)
}

// GetBinaryResource returns nil when the resource can't be opened.
func (a *NaiveUserAgent) GetBinaryResource(uri string) []byte {
	is := a.self.ResolveAndOpenStream(uri)
	if is == nil {
		return nil
	}
	defer is.Close()
	result, err := ufo.IOUtilReadBytesInputStream(is)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(fmt.Sprintf("Can't read binary resource from '%s'", uri), err))
	}
	return result
}

// IsVisited returns true if the given URI was visited, meaning it was
// requested at some point since initialization.
//
// It always returns false; visits are not tracked in the NaiveUserAgent.
func (a *NaiveUserAgent) IsVisited(uri string) bool {
	return false
}

// SetBaseURL sets the URL relative to which URIs are resolved.
func (a *NaiveUserAgent) SetBaseURL(url string) {
	a.baseURL = url
}

// naiveUserAgentFileToURL is new File(path).toURI().toURL().toExternalForm():
// "file:" and the escaped absolute path, with a trailing slash for a
// directory.
func naiveUserAgentFileToURL(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	slashed := filepath.ToSlash(abs)
	if !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}
	if info, err := os.Stat(abs); err == nil && info.IsDir() && !strings.HasSuffix(slashed, "/") {
		slashed += "/"
	}
	return "file:" + (&url.URL{Path: slashed}).EscapedPath(), nil
}

// ResolveURI resolves the URI; if absolute, leaves as is, if relative,
// returns an absolute URI based on the baseUrl for the agent.
//
// uri is a URI, possibly relative.
//
// It returns a URI as string, resolved, or "" if there was an exception (for
// example if the URI is malformed), where Java returns null. "" is also
// returned for "", where Java returns null for null.
func (a *NaiveUserAgent) ResolveURI(uri string) string {
	if uri == "" {
		return ""
	}

	if a.baseURL == "" { //first try to set a base URL
		result, err := url.Parse(uri)
		if err != nil {
			ufo.XRLogExceptionWithTh("The default NaiveUserAgent could not use the URL as base url: "+uri, err)
		} else if result.IsAbs() && result.Scheme != "file" {
			a.self.SetBaseURL(result.String())
		}

		if a.baseURL == "" { // still not set -> fallback to current working directory
			dir, err := os.Getwd()
			var base string
			if err == nil {
				base, err = naiveUserAgentFileToURL(dir)
			}
			if err != nil {
				ufo.XRLogException(fmt.Sprintf("The default NaiveUserAgent doesn't know how to resolve the base URL for '%s': %s", uri, err))
				return ""
			}
			a.self.SetBaseURL(base)
		}
	}

	// baseURL is guaranteed to be set at this point.
	// test if the URI is valid; if not, try to assign the base url as its parent
	var t error
	result, err := url.Parse(uri)
	if err == nil {
		if result.IsAbs() {
			if result.Scheme == "classpath" {
				// The core opens "classpath" URLs from its embedded
				// resources, which is what Java does when a URLStreamHandler
				// for the classpath protocol is registered; the URL is left
				// as it is when it names an embedded file.
				resource := a.self.ResolveClasspathUrl(uri)
				if resource != nil {
					return resource.String()
				}
			}
			return result.String()
		}
		ufo.XRLogLoad(uri + " is not a URL; may be relative. Testing using parent URL " + a.baseURL)
		var baseURI *url.URL
		baseURI, err = url.Parse(a.baseURL)
		if err == nil {
			if baseURI.Opaque == "" {
				// uri.resolve(child) only works for opaque URIs.
				// Otherwise, it would simply return child.
				return naiveUserAgentJavaForm(baseURI.ResolveReference(result))
			}
			// Fall back to previous resolution using URL. Java builds
			// new URL(new URL(_baseURL), uri), which resolves against the
			// scheme-specific part of an opaque base such as "jar:file:...".
			// net/url can't resolve against an opaque URL. Of the JDK's URL
			// handlers, those for "jar" (naiveUserAgentResolveJarURL) and
			// "mailto" accept an opaque context.
			switch {
			case strings.EqualFold(baseURI.Scheme, "jar"):
				var resolved string
				resolved, t = naiveUserAgentResolveJarURL(a.baseURL, uri)
				if t == nil {
					return resolved
				}
			case strings.EqualFold(baseURI.Scheme, "mailto"):
				// mailto.Handler.parseURL takes the spec, without its
				// reference, as the address.
				address := uri
				if i := strings.IndexByte(address, '#'); i >= 0 {
					address = address[:i]
				}
				if strings.TrimSpace(address) != "" {
					return "mailto:" + address
				}
				t = errors.New("No email address")
			default:
				t = fmt.Errorf("unknown protocol: %s", baseURI.Scheme)
			}
		} else {
			t = err
		}
	} else {
		t = err
	}
	ufo.XRLogExceptionWithTh("The default NaiveUserAgent cannot resolve the URL "+uri+" with base URL "+a.baseURL, t)
	return ""
}

// naiveUserAgentResolveJarURL is new URL(new URL(base), spec).toExternalForm()
// for a base URL with the "jar" scheme and a relative spec: the JDK's
// sun.net.www.protocol.jar.Handler.parseURL resolves spec against the entry
// path after "!/" of the base, a spec starting with "/" replacing the whole
// entry path, and removes "." and ".." segments from the entry path.
func naiveUserAgentResolveJarURL(base string, spec string) (string, error) {
	// URL.getFile() of the base: everything after "jar:" up to the reference.
	ctxFile := base[len("jar:"):]
	if i := strings.IndexByte(ctxFile, '#'); i >= 0 {
		ctxFile = ctxFile[:i]
	}

	var file string
	ref := ""
	hasRef := false
	refOnly := false
	if refPos := strings.IndexByte(spec, '#'); refPos >= 0 {
		ref = spec[refPos+1:]
		hasRef = true
		refOnly = refPos == 0
		spec = spec[:refPos]
	}
	if refOnly {
		file = ctxFile
	} else {
		// parseContextSpec
		if strings.HasPrefix(spec, "/") {
			bangSlash := naiveUserAgentIndexOfBangSlash(ctxFile)
			if bangSlash == -1 {
				return "", fmt.Errorf("malformed context url:%s: no !/", base)
			}
			ctxFile = ctxFile[:bangSlash]
		} else {
			lastSlash := strings.LastIndexByte(ctxFile, '/')
			if lastSlash == -1 {
				return "", fmt.Errorf("malformed context url:%s", base)
			} else if lastSlash < len(ctxFile)-1 {
				ctxFile = ctxFile[:lastSlash+1]
			}
		}
		file = ctxFile + spec

		bangSlash := naiveUserAgentIndexOfBangSlash(file)
		if bangSlash == -1 {
			return "", fmt.Errorf("no !/ in spec")
		}
		file = file[:bangSlash] + naiveUserAgentCanonizeString(file[bangSlash:])
	}
	result := "jar:" + file
	if hasRef {
		result += "#" + ref
	}
	return result, nil
}

// naiveUserAgentIndexOfBangSlash returns the index of the "/" of the last
// "!/" in spec, or -1 (jar.Handler.indexOfBangSlash).
func naiveUserAgentIndexOfBangSlash(spec string) int {
	for indexOfBang := len(spec); indexOfBang >= 0; indexOfBang-- {
		indexOfBang = strings.LastIndexByte(spec[:min(indexOfBang+1, len(spec))], '!')
		if indexOfBang == -1 {
			return -1
		}
		if indexOfBang != len(spec)-1 && spec[indexOfBang+1] == '/' {
			return indexOfBang + 1
		}
	}
	return -1
}

// naiveUserAgentCanonizeString removes "/../" and "/./" segments and a
// trailing "/.." or "/." (sun.net.www.ParseUtil.canonizeString).
func naiveUserAgentCanonizeString(file string) string {
	if len(file) == 0 || (!strings.Contains(file, "./") && file[len(file)-1] != '.') {
		return file
	}
	// Remove embedded /../
	for {
		i := strings.Index(file, "/../")
		if i < 0 {
			break
		}
		if lim := strings.LastIndexByte(file[:i], '/'); lim >= 0 {
			file = file[:lim] + file[i+3:]
		} else {
			file = file[i+3:]
		}
	}
	// Remove embedded /./
	for {
		i := strings.Index(file, "/./")
		if i < 0 {
			break
		}
		file = file[:i] + file[i+2:]
	}
	// Remove trailing ..
	for strings.HasSuffix(file, "/..") {
		i := strings.Index(file, "/..")
		if lim := strings.LastIndexByte(file[:i], '/'); lim >= 0 {
			file = file[:lim+1]
		} else {
			file = file[:i]
		}
	}
	// Remove trailing .
	if strings.HasSuffix(file, "/.") {
		file = file[:len(file)-1]
	}
	return file
}

// naiveUserAgentJavaForm writes a resolved URL as java.net.URI.toString
// does: a "file" URL without an authority keeps the single-slash form
// ("file:/dir/page.html") that the base URL had, where url.URL.String writes
// "file:///dir/page.html".
func naiveUserAgentJavaForm(u *url.URL) string {
	if u.Scheme == "file" && u.Host == "" && u.User == nil && u.Opaque == "" {
		s := "file:" + u.EscapedPath()
		if u.RawQuery != "" {
			s += "?" + u.RawQuery
		}
		if u.Fragment != "" {
			s += "#" + u.EscapedFragment()
		}
		return s
	}
	return u.String()
}

// ResolveClasspathUrl returns the URL of a file embedded from the core's
// resources directory, or nil. It is package-private in Java, where the
// context class loader looks the resource up.
func (a *NaiveUserAgent) ResolveClasspathUrl(uri string) *url.URL {
	// replaceFirst
	path := uri
	if loc := naiveUserAgentClasspathPrefix.FindStringIndex(uri); loc != nil {
		path = uri[:loc[0]] + uri[loc[1]:]
	}
	return ufo.GeneralUtilGetURLFromClasspath(a, path)
}

// GetBaseURL returns the current baseUrl for this class.
func (a *NaiveUserAgent) GetBaseURL() string {
	return a.baseURL
}

func (a *NaiveUserAgent) DocumentStarted() {
	a.self.ShrinkImageCache()
}

func (a *NaiveUserAgent) DocumentLoaded() { /* ignore*/ }

func (a *NaiveUserAgent) OnLayoutException(t error) { /* ignore*/ }

func (a *NaiveUserAgent) OnRenderException(t error) { /* ignore*/ }
