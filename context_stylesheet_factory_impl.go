// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/context/StylesheetFactoryImpl.java

package ufo

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

// StylesheetFactoryImpl is a Factory class for Cascading Style Sheets. Sheets
// are parsed using a single parser instance for all sheets. Sheets are cached
// by URI using LRU test, but timestamp of file is not checked.
type StylesheetFactoryImpl struct {
	// userAgentCallback is the UserAgentCallback to resolve uris.
	userAgentCallback UserAgentCallback

	// cache is an LRU cache. Java wraps it in Collections.synchronizedMap;
	// cacheLock is that wrapper's lock.
	cache     *StylesheetCache
	cacheLock sync.Mutex

	cssParser *CSSParser

	// unsupportedCssFeatures is a LinkedHashSet in Java: the slice keeps the
	// insertion order and the map holds the members.
	unsupportedCssFeatures    []string
	unsupportedCssFeaturesSet map[string]struct{}
}

func NewStylesheetFactoryImpl(userAgentCallback UserAgentCallback) *StylesheetFactoryImpl {
	f := &StylesheetFactoryImpl{
		userAgentCallback:         userAgentCallback,
		cache:                     NewStylesheetCache(),
		unsupportedCssFeaturesSet: make(map[string]struct{}),
	}
	f.cssParser = NewCSSParserWithCss3FeatureListener(
		CSSErrorHandlerFunc(func(uri string, message string) {
			XRLogCssParse(LevelWarning, "("+uri+") "+message)
		}),
		func(feature string) {
			if _, ok := f.unsupportedCssFeaturesSet[feature]; !ok {
				f.unsupportedCssFeaturesSet[feature] = struct{}{}
				f.unsupportedCssFeatures = append(f.unsupportedCssFeatures, feature)
			}
		})
	return f
}

// GetUnsupportedCssFeatures returns the unsupported CSS features the parser
// reported, each once, in the order they were first reported.
func (f *StylesheetFactoryImpl) GetUnsupportedCssFeatures() []string {
	return f.unsupportedCssFeatures
}

func (f *StylesheetFactoryImpl) Parse(reader io.Reader, info *StylesheetInfo) *Stylesheet {
	return f.ParseWithUriOrigin(reader, info.GetUri(), info.GetOrigin())
}

func (f *StylesheetFactoryImpl) ParseWithUriOrigin(reader io.Reader, uri string, origin StylesheetInfoOrigin) *Stylesheet {
	sheet, err := f.cssParser.ParseStylesheet(uri, origin, reader)
	if err != nil {
		XRLogCssParse(LevelWarning, "Couldn't parse stylesheet at URI "+uri+": "+err.Error(), err)
		return NewStylesheet(uri, origin)
	}
	return sheet
}

// parse returns nil if uri could not be loaded.
func (f *StylesheetFactoryImpl) parse(info *StylesheetInfo) *Stylesheet {
	var cr *CSSResource
	if css := info.GetContent(); css != nil {
		cr = NewCSSResource(strings.NewReader(*css))
	} else {
		cr = f.userAgentCallback.GetCSSResource(info.GetUri())
	}
	// Whether by accident or design, InputStream will never be null
	// since the null resource stream is wrapped in a BufferedInputStream
	inputSource := cr.GetResourceInputSource()
	if inputSource == nil {
		return nil
	}

	is := inputSource.GetByteStream()
	if is == nil {
		return nil
	}
	if closer, ok := is.(io.Closer); ok {
		defer closer.Close()
	}
	charset := ConfigurationValueForWithDefaultVal("xr.stylesheets.charset-name", "UTF-8")
	reader, err := stylesheetFactoryImplNewInputStreamReader(is, charset)
	if err != nil {
		panic(NewXRRuntimeExceptionWithCause(err.Error(), err))
	}
	return f.Parse(reader, info)
}

// stylesheetFactoryImplNewInputStreamReader ports
// new InputStreamReader(is, charsetName) for the charsets the standard
// library can decode: UTF-8, US-ASCII and ISO-8859-1. Any other name is the
// error Java reports as UnsupportedEncodingException.
func stylesheetFactoryImplNewInputStreamReader(is io.Reader, charsetName string) (io.Reader, error) {
	switch strings.ToUpper(charsetName) {
	case "UTF-8", "UTF8", "US-ASCII", "ASCII":
		return is, nil
	case "ISO-8859-1", "ISO8859_1", "LATIN1":
		return &stylesheetFactoryImplLatin1Reader{source: bufio.NewReader(is)}, nil
	}
	return nil, fmt.Errorf("%s", charsetName)
}

// stylesheetFactoryImplLatin1Reader decodes ISO-8859-1 bytes to UTF-8.
type stylesheetFactoryImplLatin1Reader struct {
	source  *bufio.Reader
	pending []byte
}

func (r *stylesheetFactoryImplLatin1Reader) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		if len(r.pending) > 0 {
			copied := copy(p[n:], r.pending)
			r.pending = r.pending[copied:]
			n += copied
			continue
		}
		b, err := r.source.ReadByte()
		if err != nil {
			if n > 0 {
				return n, nil
			}
			return 0, err
		}
		if b < 0x80 {
			p[n] = b
			n++
		} else {
			r.pending = []byte{0xC0 | b>>6, 0x80 | b&0x3F}
		}
	}
	return n, nil
}

func (f *StylesheetFactoryImpl) ParseStyleDeclaration(origin StylesheetInfoOrigin, styleDeclaration string) *Ruleset {
	return f.cssParser.ParseDeclaration(origin, styleDeclaration)
}

// PutStylesheet adds a stylesheet to the factory cache. Will overwrite older
// entry for same key.
//
// key is the key to use to reference sheet later; must be unique in factory.
// sheet is the sheet to cache.
func (f *StylesheetFactoryImpl) PutStylesheet(key string, sheet *Stylesheet) {
	f.cacheLock.Lock()
	defer f.cacheLock.Unlock()
	f.cache.Put(key, sheet)
}

// ContainsStylesheet returns true if a Stylesheet with this key has been put
// in the cache. Note that the Stylesheet may be nil.
//
// TODO: work out how to handle caching properly, with cache invalidation
func (f *StylesheetFactoryImpl) ContainsStylesheet(key string) bool {
	f.cacheLock.Lock()
	defer f.cacheLock.Unlock()
	return f.cache.ContainsKey(key)
}

// RemoveCachedStylesheet removes a cached sheet by its key.
//
// key is the key for this sheet; same as key passed to PutStylesheet.
func (f *StylesheetFactoryImpl) RemoveCachedStylesheet(key string) {
	f.cacheLock.Lock()
	defer f.cacheLock.Unlock()
	f.cache.Remove(key)
}

func (f *StylesheetFactoryImpl) FlushCachedStylesheets() {
	f.cacheLock.Lock()
	defer f.cacheLock.Unlock()
	f.cache.Clear()
}

// GetStylesheet returns a cached sheet by its key; loads and caches it if not
// in cache; nil if not able to load.
//
// info is the StylesheetInfo for this sheet.
//
// TODO: this looks a bit odd
func (f *StylesheetFactoryImpl) GetStylesheet(info *StylesheetInfo) *Stylesheet {
	XRLogLoad("Requesting stylesheet: " + info.GetUri())

	f.cacheLock.Lock()
	s := f.cache.Get(info.GetUri())
	f.cacheLock.Unlock()
	if s == nil && !f.ContainsStylesheet(info.GetUri()) {
		s = f.parse(info)
		f.PutStylesheet(info.GetUri(), s)
	}
	return s
}

func (f *StylesheetFactoryImpl) SetUserAgentCallback(userAgent UserAgentCallback) {
	f.userAgentCallback = userAgent
}

func (f *StylesheetFactoryImpl) SetSupportCMYKColors(b bool) {
	f.cssParser.SetSupportCMYKColors(b)
}

var _ StylesheetFactory = (*StylesheetFactoryImpl)(nil)
