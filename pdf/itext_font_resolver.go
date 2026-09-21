// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextFontResolver.java

package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

const (
	iTextFontResolverOtf      = ".otf"
	iTextFontResolverTtf      = ".ttf"
	iTextFontResolverAfm      = ".afm"
	iTextFontResolverPfm      = ".pfm"
	iTextFontResolverPfb      = ".pfb"
	iTextFontResolverPfa      = ".pfa"
	iTextFontResolverTtc      = ".ttc"
	iTextFontResolverTtcComma = ".ttc,"
)

// ITextFontResolverI lists the non-private methods of ITextFontResolver,
// which CJKFontResolver extends. A Java value typed ITextFontResolver is an
// ITextFontResolverI.
type ITextFontResolverI interface {
	ufo.FontResolver
	GetFonts() map[string]*FontFamily
	FlushFontFaceFonts()
	GetNonEmbeddedFontFaceFamilies() []string
	AddEmbedFontFace(fontFamily string, encoding string)
	ResetEmbedFontFace()
	ImportFontFaces(fontFaces []*ufo.FontFaceRule, userAgentCallback ufo.UserAgentCallback)
	AddFontDirectory(dir string, embedded bool) error
	AddFontDirectoryWithEncoding(dir string, encoding string, embedded bool) error
	AddFont(path string, embedded bool) error
	AddFontWithEncoding(path string, encoding string, embedded bool) error
	AddFontWithEncodingPathToPFB(path string, encoding string, embedded bool, pathToPFB string) error
	AddFontWithFontFamilyNameOverrideEncodingPathToPFB(path string, fontFamilyNameOverride string,
		encoding string, embedded bool, pathToPFB string) error
	AddFontBaseFont(font *writer.BaseFont, path string, fontFamilyNameOverride string)
	NormalizeFontFamily(fontFamily string) string
	LoadFonts() map[string]*FontFamily
	AsITextFontResolver() *ITextFontResolver
}

type ITextFontResolver struct {
	// self is the outermost object, so that GetFonts calls the LoadFonts of
	// a subclass (CJKFontResolver).
	self ITextFontResolverI

	embedFontFaces map[string]string

	// fontFamiliesLock stands for synchronized (_fontFamilies).
	fontFamiliesLock sync.Mutex
	fontFamilies     map[string]*FontFamily

	// fontCache is a ConcurrentHashMap in Java.
	fontCacheLock sync.Mutex
	fontCache     map[string]*FontDescription
}

var _ ITextFontResolverI = (*ITextFontResolver)(nil)

func NewITextFontResolver() *ITextFontResolver {
	r := &ITextFontResolver{}
	initITextFontResolver(r)
	r.self = r
	return r
}

// initITextFontResolver initializes the fields of r; constructors of
// subclasses call it and then SetSelf.
func initITextFontResolver(r *ITextFontResolver) {
	r.embedFontFaces = map[string]string{}
	r.fontFamilies = map[string]*FontFamily{}
	r.fontCache = map[string]*FontDescription{}
}

// SetSelf sets the outermost object, the receiver of calls to LoadFonts.
func (r *ITextFontResolver) SetSelf(self ITextFontResolverI) {
	r.self = self
}

func (r *ITextFontResolver) AsITextFontResolver() *ITextFontResolver {
	return r
}

func (r *ITextFontResolver) GetFonts() map[string]*FontFamily {
	r.fontFamiliesLock.Lock()
	defer r.fontFamiliesLock.Unlock()
	if len(r.fontFamilies) == 0 {
		for name, family := range r.self.LoadFonts() {
			r.fontFamilies[name] = family
		}
	}
	return r.fontFamilies
}

// ITextFontResolverGetDistinctFontFamilyNames is a utility method which uses
// the writer to determine the family name(s) for the font at the given path.
// The APIs seem to indicate there can be more than one name, but this method
// will return a set of them. Use a name from this list when referencing the
// font in CSS for PDF output. Note that family names as reported here may
// vary from those reported by other font libraries, e.g. "Arial Unicode MS"
// and "ArialUnicodeMS".
//
// path is the local path to the font file; encoding and embedded are the same
// as what you would use for AddFontWithEncoding. It returns the set of all
// family names for the font file.
func ITextFontResolverGetDistinctFontFamilyNames(path string, encoding string, embedded bool) map[string]struct{} {
	font, err := writer.CreateFont(path, encoding, embedded)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(
			fmt.Sprintf("Failed to read font family names from %s (encoding: %s, embedded: %t)", path, encoding, embedded), err))
	}
	fontFamilyNames := TrueTypeUtilGetFamilyNames(font)
	result := map[string]struct{}{}
	for _, name := range fontFamilyNames {
		result[name] = struct{}{}
	}
	return result
}

// ResolveFont may return nil.
func (r *ITextFontResolver) ResolveFont(renderingContext *ufo.SharedContext, spec *ufo.FontSpecification) ufo.FSFont {
	return r.resolveFontFamilies(spec.Families(), spec.Size(), spec.FontWeight(), spec.FontStyle())
}

func (r *ITextFontResolver) FlushCache() {
	r.fontFamiliesLock.Lock()
	for name := range r.fontFamilies {
		delete(r.fontFamilies, name)
	}
	r.fontFamiliesLock.Unlock()
	r.clearFontCache()
}

func (r *ITextFontResolver) clearFontCache() {
	r.fontCacheLock.Lock()
	defer r.fontCacheLock.Unlock()
	for key := range r.fontCache {
		delete(r.fontCache, key)
	}
}

func (r *ITextFontResolver) FlushFontFaceFonts() {
	r.clearFontCache()

	fonts := r.GetFonts()
	for name, family := range fonts {
		family.removeFontDescriptionsIf((*FontDescription).IsFromFontFace)
		if len(family.GetFontDescriptions()) == 0 {
			delete(fonts, name)
		}
	}
}

// GetNonEmbeddedFontFaceFamilies returns the names of @font-face families
// that were registered without embedding (see -fs-pdf-font-embed). PDF/A
// requires every font used in the document to be embedded, so callers
// requesting PDF/A conformance should ensure this returns an empty list
// before generating the PDF.
//
// Java streams over the values of a HashMap, whose order is not defined; the
// names are sorted here.
func (r *ITextFontResolver) GetNonEmbeddedFontFaceFamilies() []string {
	seen := map[string]struct{}{}
	result := []string{}
	for _, family := range r.GetFonts() {
		for _, description := range family.GetFontDescriptions() {
			if !description.IsFromFontFace() {
				continue
			}
			if description.GetFont().IsEmbedded() {
				continue
			}
			name := description.GetFont().GetPostscriptFontName()
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				result = append(result, name)
			}
		}
	}
	sort.Strings(result)
	return result
}

func (r *ITextFontResolver) AddEmbedFontFace(fontFamily string, encoding string) {
	r.embedFontFaces[fontFamily] = encoding
}

func (r *ITextFontResolver) ResetEmbedFontFace() {
	for name := range r.embedFontFaces {
		delete(r.embedFontFaces, name)
	}
}

func (r *ITextFontResolver) ImportFontFaces(fontFaces []*ufo.FontFaceRule, userAgentCallback ufo.UserAgentCallback) {
	for _, rule := range fontFaces {
		r.importFontFace(rule, userAgentCallback)
	}
}

func (r *ITextFontResolver) importFontFace(rule *ufo.FontFaceRule, userAgentCallback ufo.UserAgentCallback) {
	style := rule.GetCalculatedStyle()

	src := style.ValueByName(ufo.CSSNameSrc)
	if src == ufo.FSDerivedValue(ufo.IdentValueNone) {
		return
	}

	fontSources := r.parseFontSources(src)
	if len(fontSources) == 0 {
		ufo.XRLogException("No valid font sources found in src property")
		return
	}

	// Try each font source in order until one works
	for _, fontSrc := range fontSources {
		uri := fontSrc.uri
		format := fontSrc.format

		// Check if format is supported
		if !r.fontSupported(uri, format) {
			continue
		}
		font1 := userAgentCallback.GetBinaryResource(uri)
		if font1 == nil {
			ufo.XRLogException("Could not load font " + uri)
			continue
		}

		var font2 []byte
		metricsSrc := style.ValueByName(ufo.CSSNameFsFontMetricSrc)
		if metricsSrc != ufo.FSDerivedValue(ufo.IdentValueNone) {
			font2 = userAgentCallback.GetBinaryResource(metricsSrc.AsString())
			if font2 == nil {
				ufo.XRLogException("Could not load font metric data " + uri)
				continue
			}
		}

		if font2 != nil {
			font1, font2 = font2, font1
		}

		embedded := style.IsIdent(ufo.CSSNameFsPdfFontEmbed, ufo.IdentValueEmbed)
		encoding := style.GetStringProperty(ufo.CSSNameFsPdfFontEncoding)
		// fontFamily is "" where Java has null.
		fontFamily := ""
		if rule.HasFontFamily() {
			fontFamily = style.ValueByName(ufo.CSSNameFontFamily).AsString()
		}
		if embedEncoding, ok := r.embedFontFaces[fontFamily]; ok {
			embedded = true
			encoding = embedEncoding
		}
		var fontWeight *ufo.IdentValue
		if rule.HasFontWeight() {
			fontWeight = style.GetIdent(ufo.CSSNameFontWeight)
		}
		var fontStyle *ufo.IdentValue
		if rule.HasFontStyle() {
			fontStyle = style.GetIdent(ufo.CSSNameFontStyle)
		}

		err := r.addFontFaceFont(fontFamily, fontWeight, fontStyle, uri, format, encoding, embedded, font1, font2)
		if err == nil {
			// Successfully added font, no need to try other sources
			return
		}
		ufo.XRLogExceptionWithTh("Could not load font "+uri, err)
		// Continue to try next font source
	}

	ufo.XRLogException("Failed to load any font from src property")
}

// iTextFontResolverFontSrc represents a font source with URI and optional
// format ("" when there is none).
type iTextFontResolverFontSrc struct {
	uri    string
	format string
}

// parseFontSources parses font sources from the src property value. Handles
// both single values and comma-separated lists of url()/format() pairs.
func (r *ITextFontResolver) parseFontSources(src ufo.FSDerivedValue) []iTextFontResolverFontSrc {
	var result []iTextFontResolverFontSrc

	// Check if it's a list value (multiple sources)
	if listValue, ok := src.(*ufo.ListValue); ok {
		values := listValue.GetValues()
		if values != nil {
			// currentUriSet stands for currentUri != null.
			currentUriSet := false
			currentUri := ""
			currentFormat := ""

			for _, value := range values {
				if propValue, ok := value.(*ufo.PropertyValue); ok {
					if propValue.GetPrimitiveType() == ufo.CSSPrimitiveValueCssUri {
						// If we have a pending URI, add it before starting a new one
						if currentUriSet {
							result = append(result, iTextFontResolverFontSrc{uri: currentUri, format: currentFormat})
							currentFormat = ""
						}
						currentUri = propValue.GetStringValue()
						currentUriSet = true
					} else if propValue.GetPropertyValueType() == ufo.PropertyValueTypeValueTypeFunction {
						function := propValue.GetFunction()
						if function != nil && function.Is("format") {
							params := function.GetParameters()
							if len(params) != 0 && params[0].GetStringValue() != "" {
								currentFormat = params[0].GetStringValue()
							}
						}
					}
				}
			}

			// Add the last URI if present
			if currentUriSet {
				result = append(result, iTextFontResolverFontSrc{uri: currentUri, format: currentFormat})
			}
		}
	} else {
		// Single value (just a URI)
		result = append(result, iTextFontResolverFontSrc{uri: src.AsString(), format: ""})
	}

	return result
}

// AddFontDirectory adds all fonts from given directory with encoding
// "CP1252" (don't ask me why :) )
func (r *ITextFontResolver) AddFontDirectory(dir string, embedded bool) error {
	return r.AddFontDirectoryWithEncoding(dir, writer.BaseFontCp1252, embedded)
}

// AddFontDirectoryWithEncoding adds all fonts from given directory (all files
// with extension ".otf" and ".ttf")
func (r *ITextFontResolver) AddFontDirectoryWithEncoding(dir string, encoding string, embedded bool) (err error) {
	defer iTextRendererRecover(&err)
	info, statErr := os.Stat(dir)
	if statErr != nil || !info.IsDir() {
		panic(ufo.NewXRRuntimeException(fmt.Sprintf("%s is not a directory", dir)))
	}
	files, err := r.filesWithExtensions(dir, iTextFontResolverOtf, iTextFontResolverTtf)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := r.AddFontWithEncoding(file, encoding, embedded); err != nil {
			return err
		}
	}
	return nil
}

// filesWithExtensions returns the absolute paths of the files in directory f
// whose lower-cased name ends with one of the extensions.
func (r *ITextFontResolver) filesWithExtensions(f string, extensions ...string) ([]string, error) {
	entries, err := os.ReadDir(f)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(f)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, entry := range entries {
		lower := strings.ToLower(entry.Name())
		for _, extension := range extensions {
			if strings.HasSuffix(lower, extension) {
				result = append(result, filepath.Join(abs, entry.Name()))
				break
			}
		}
	}
	return result, nil
}

// AddFont adds the font with encoding "CP1252" (don't ask me why :) )
func (r *ITextFontResolver) AddFont(path string, embedded bool) error {
	return r.AddFontWithEncoding(path, writer.BaseFontCp1252, embedded)
}

func (r *ITextFontResolver) AddFontWithEncoding(path string, encoding string, embedded bool) error {
	return r.AddFontWithEncodingPathToPFB(path, encoding, embedded, "")
}

// AddFontWithEncodingPathToPFB ports addFont(path, encoding, embedded,
// pathToPFB); pathToPFB is "" where Java passes null.
func (r *ITextFontResolver) AddFontWithEncodingPathToPFB(path string, encoding string, embedded bool, pathToPFB string) error {
	return r.AddFontWithFontFamilyNameOverrideEncodingPathToPFB(path, "", encoding, embedded, pathToPFB)
}

// AddFontWithFontFamilyNameOverrideEncodingPathToPFB ports addFont(path,
// fontFamilyNameOverride, encoding, embedded, pathToPFB);
// fontFamilyNameOverride and pathToPFB are "" where Java passes null.
//
// A Type 1 font (.afm or .pfm) is not supported by the writer: the error of
// writer.CreateFontWithBytes, a *writer.UnsupportedFeatureError, is returned.
// A font whose description can't be read makes TrueTypeUtil panic with an
// *ufo.XRRuntimeException in Java's terms; it is returned as the error.
func (r *ITextFontResolver) AddFontWithFontFamilyNameOverrideEncodingPathToPFB(path string, fontFamilyNameOverride string,
	encoding string, embedded bool, pathToPFB string) (err error) {
	defer iTextRendererRecover(&err)
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, iTextFontResolverOtf) || strings.HasSuffix(lower, iTextFontResolverTtf) || strings.Contains(lower, iTextFontResolverTtcComma) {
		font, err := writer.CreateFont(path, encoding, embedded)
		if err != nil {
			return err
		}
		r.AddFontBaseFont(font, path, fontFamilyNameOverride)
	} else if strings.HasSuffix(lower, iTextFontResolverTtc) {
		names, err := writer.EnumerateTTCNames(path)
		if err != nil {
			return err
		}
		for i := 0; i < len(names); i++ {
			if err := r.AddFontWithFontFamilyNameOverrideEncodingPathToPFB(fmt.Sprintf("%s,%d", path, i), fontFamilyNameOverride, encoding, embedded, ""); err != nil {
				return err
			}
		}
	} else if strings.HasSuffix(lower, iTextFontResolverAfm) || strings.HasSuffix(lower, iTextFontResolverPfm) {
		if embedded && pathToPFB == "" {
			return fmt.Errorf("When embedding a font, path to PFB/PFA file must be specified (path: %s)", path)
		}

		// Java reads pathToPFB even when it is null, which fails; here no
		// file is read then.
		var pfb []byte
		if pathToPFB != "" {
			pfb, err = r.readFile(pathToPFB)
			if err != nil {
				return err
			}
		}
		font, err := writer.CreateFontWithBytes(
			path, encoding, embedded, false, nil, pfb)
		if err != nil {
			return err
		}

		fontFamilyName := fontFamilyNameOverride
		if fontFamilyName == "" {
			fontFamilyName = font.GetFamilyFontName()[0][3]
		}
		fontFamily := r.getFontFamily(fontFamilyName)

		description := NewFontDescription(font)
		// XXX Need to set weight, underline position, etc.  This information
		// is contained in the AFM file (and even parsed by Type1Font), but
		// unfortunately it isn't exposed to the caller.
		fontFamily.AddFontDescription(description)
	} else {
		return fmt.Errorf("Unsupported font type: %s", path)
	}
	return nil
}

// AddFontBaseFont ports addFont(BaseFont, String, String);
// fontFamilyNameOverride is "" where Java passes null.
func (r *ITextFontResolver) AddFontBaseFont(font *writer.BaseFont, path string, fontFamilyNameOverride string) {
	fontFamilyNames := iTextFontResolverGetFontFamilyNames(font, fontFamilyNameOverride)

	for _, fontFamilyName := range fontFamilyNames {
		r.getFontFamily(fontFamilyName).
			AddFontDescription(TrueTypeUtilExtractDescription(path, font, nil))
	}
}

func iTextFontResolverGetFontFamilyNames(font *writer.BaseFont, fontFamilyNameOverride string) []string {
	if fontFamilyNameOverride != "" {
		return []string{fontFamilyNameOverride}
	} else {
		return TrueTypeUtilGetFamilyNames(font)
	}
}

// fontSupported takes format "" where Java passes null.
func (r *ITextFontResolver) fontSupported(uri string, format string) bool {
	if format != "" {
		return format == "opentype" || format == "truetype"
	}
	lower := strings.ToLower(uri)
	if ufo.FontUtilIsEmbeddedBase64Font(uri) {
		return ufo.SupportedEmbeddedFontTypesIsSupported(uri)
	} else {
		return strings.HasSuffix(lower, iTextFontResolverOtf) || strings.HasSuffix(lower, iTextFontResolverTtf)
	}
}

// getFontName takes format and fontFamilyName "" where Java passes null.
func (r *ITextFontResolver) getFontName(uri string, format string, fontFamilyName string) string {
	if fontFamilyName != "" {
		if ufo.FontUtilIsEmbeddedBase64Font(uri) {
			return fontFamilyName + ufo.SupportedEmbeddedFontTypesGetExtension(uri)
		} else if format != "" {
			lower := strings.ToLower(uri)
			if !strings.HasSuffix(lower, iTextFontResolverOtf) && !strings.HasSuffix(lower, iTextFontResolverTtf) {
				ext := ""
				switch format {
				case "opentype":
					ext = iTextFontResolverOtf
				case "truetype":
					ext = iTextFontResolverTtf
				}
				return fontFamilyName + ext
			}
		}
	}
	return uri
}

// addFontFaceFont takes ttfAfm, the font as a byte array, and pfb, possibly
// nil.
func (r *ITextFontResolver) addFontFaceFont(fontFamilyNameOverride string, fontWeightOverride *ufo.IdentValue,
	fontStyleOverride *ufo.IdentValue, uri string, format string,
	encoding string, embedded bool, ttfAfm []byte, pfb []byte) error {
	fontName := r.getFontName(uri, format, fontFamilyNameOverride)
	font, err := writer.CreateFontWithBytes(fontName, encoding, embedded, false, ttfAfm, pfb)
	if err != nil {
		return err
	}

	fontFamilyNames := iTextFontResolverGetFontFamilyNames(font, fontFamilyNameOverride)

	for _, fontFamilyName := range fontFamilyNames {
		fontFamily := r.getFontFamily(fontFamilyName)
		fontFamily.AddFontDescription(
			iTextFontResolverFontDescription(fontWeightOverride, fontStyleOverride, uri, ttfAfm, font),
		)
	}
	return nil
}

func iTextFontResolverFontDescription(fontWeightOverride *ufo.IdentValue, fontStyleOverride *ufo.IdentValue,
	uri string, ttfAfm []byte, font *writer.BaseFont) *FontDescription {
	return TrueTypeUtilExtractDescriptionWithContents(uri, ttfAfm, font, true, fontWeightOverride, fontStyleOverride)
}

func (r *ITextFontResolver) readFile(path string) ([]byte, error) {
	return ufo.IOUtilReadBytesPath(path)
}

func (r *ITextFontResolver) getFontFamily(fontFamilyName string) *FontFamily {
	fontFamily := r.GetFonts()[fontFamilyName]
	if fontFamily == nil {
		fontFamily = NewFontFamily(fontFamilyName)
		r.GetFonts()[fontFamilyName] = fontFamily
	}
	return fontFamily
}

// resolveFontFamilies ports the private resolveFont(String[], float,
// IdentValue, IdentValue); families may be nil.
func (r *ITextFontResolver) resolveFontFamilies(families []string, size float32, weight *ufo.IdentValue, style *ufo.IdentValue) ufo.FSFont {
	if !(style == ufo.IdentValueNormal || style == ufo.IdentValueOblique ||
		style == ufo.IdentValueItalic) {
		style = ufo.IdentValueNormal
	}
	if families != nil {
		for _, family := range families {
			font := r.resolveFontFamily(family, size, weight, style)
			if font != nil {
				return font
			}
		}
	}

	iTextFontResolverLogDebug("Could not resolve font %s:%s:%s - fallback to Serif", iTextFontResolverArraysToString(families), weight, style)
	return r.resolveFontFamily("Serif", size, weight, style)
}

// iTextFontResolverArraysToString formats as java.util.Arrays.toString.
func iTextFontResolverArraysToString(values []string) string {
	if values == nil {
		return "null"
	}
	return "[" + strings.Join(values, ", ") + "]"
}

// iTextFontResolverLogDebug stands for the slf4j log.debug calls of the Java
// class: the message goes to the general XRLog channel at LevelFine. The
// message is only formatted when logging is enabled.
func iTextFontResolverLogDebug(format string, args ...any) {
	if !ufo.XRLogIsLoggingEnabled() {
		return
	}
	ufo.XRLogGeneral(ufo.LevelFine, fmt.Sprintf(format, args...))
}

// NormalizeFontFamily is package-private in Java.
func (r *ITextFontResolver) NormalizeFontFamily(fontFamily string) string {
	result := r.stripQuotes(fontFamily)

	if strings.EqualFold(result, "serif") {
		return "Serif"
	} else if strings.EqualFold(result, "sans-serif") {
		return "SansSerif"
	} else if strings.EqualFold(result, "monospace") {
		return "Monospaced"
	}

	return result
}

// stripQuotes strips off the leading and trailing quote if they are there
func (r *ITextFontResolver) stripQuotes(text string) string {
	result := text
	if strings.HasPrefix(result, "\"") {
		result = result[1:]
	}
	if strings.HasSuffix(result, "\"") {
		result = result[0 : len(result)-1]
	}
	return result
}

// resolveFontFamily ports the private resolveFont(String, float, IdentValue,
// IdentValue). It returns a nil interface when the family is not known.
func (r *ITextFontResolver) resolveFontFamily(fontFamily string, size float32, weight *ufo.IdentValue, style *ufo.IdentValue) ufo.FSFont {
	normalizedFontFamily := r.NormalizeFontFamily(fontFamily)

	cacheKey := fmt.Sprintf("%s-%s-%s", normalizedFontFamily, weight, style)
	r.fontCacheLock.Lock()
	result := r.fontCache[cacheKey]
	r.fontCacheLock.Unlock()

	if result != nil {
		iTextFontResolverLogDebug("Resolved font (from cache) %s:%s:%s -> %s", fontFamily, weight, style, result)
		return NewITextFSFont(result, size)
	}

	family := r.GetFonts()[normalizedFontFamily]
	if family != nil {
		desiredWeight := ITextFontResolverConvertWeightToInt(weight)
		result = family.Match(desiredWeight, style)
		if result != nil {
			iTextFontResolverLogDebug("Resolved font %s:%s(%d):%s -> %s", fontFamily, weight, desiredWeight, style, result)
			r.fontCacheLock.Lock()
			r.fontCache[cacheKey] = result
			r.fontCacheLock.Unlock()
			return NewITextFSFont(result, size)
		}
	}

	return nil
}

func ITextFontResolverConvertWeightToInt(weight *ufo.IdentValue) int {
	if weight == ufo.IdentValueNormal {
		return 400
	} else if weight == ufo.IdentValueBold {
		return 700
	} else if weight == ufo.IdentValueFontWeight100 {
		return 100
	} else if weight == ufo.IdentValueFontWeight200 {
		return 200
	} else if weight == ufo.IdentValueFontWeight300 {
		return 300
	} else if weight == ufo.IdentValueFontWeight400 {
		return 400
	} else if weight == ufo.IdentValueFontWeight500 {
		return 500
	} else if weight == ufo.IdentValueFontWeight600 {
		return 600
	} else if weight == ufo.IdentValueFontWeight700 {
		return 700
	} else if weight == ufo.IdentValueFontWeight800 {
		return 800
	} else if weight == ufo.IdentValueFontWeight900 {
		return 900
	} else if weight == ufo.IdentValueLighter {
		// FIXME
		return 400
	} else if weight == ufo.IdentValueBolder {
		// FIXME
		return 700
	}
	panic(ufo.NewXRRuntimeException(fmt.Sprintf("Cannot convert weight to integer: %s", weight)))
}

// LoadFonts is protected in Java; CJKFontResolver overrides it.
func (r *ITextFontResolver) LoadFonts() map[string]*FontFamily {
	result := map[string]*FontFamily{}
	r.addCourier(result)
	r.addTimes(result)
	r.addHelvetica(result)
	r.addSymbol(result)
	r.addZapfDingbats(result)
	return result
}

// createFont passes writer.BaseFontWinansi ("Cp1252") where Java passes the
// literal "winansi": OpenPDF normalizes that spelling to Cp1252 and the writer
// accepts the normalized name only.
func (r *ITextFontResolver) createFont(name string) *writer.BaseFont {
	return r.createFontWithEncodingEmbedded(name, writer.BaseFontWinansi, true)
}

func (r *ITextFontResolver) createFontWithEncodingEmbedded(name string, encoding string, embedded bool) *writer.BaseFont {
	font, err := writer.CreateFont(name, encoding, embedded)
	if err != nil {
		panic(ufo.NewXRRuntimeExceptionWithCause(
			fmt.Sprintf("Failed to load font %s (encoding: %s, embedded: %t)", name, encoding, embedded), err))
	}
	return font
}

func (r *ITextFontResolver) addCourier(result map[string]*FontFamily) {
	courier := NewFontFamily("Courier")

	courier.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontCourierBoldoblique), ufo.IdentValueOblique, 700))
	courier.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontCourierOblique), ufo.IdentValueOblique, 400))
	courier.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontCourierBold), ufo.IdentValueNormal, 700))
	courier.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontCourier), ufo.IdentValueNormal, 400))

	result["DialogInput"] = courier
	result["Monospaced"] = courier
	result["Courier"] = courier
}

func (r *ITextFontResolver) addTimes(result map[string]*FontFamily) {
	times := NewFontFamily("Times")

	times.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontTimesBolditalic), ufo.IdentValueItalic, 700))
	times.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontTimesItalic), ufo.IdentValueItalic, 400))
	times.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontTimesBold), ufo.IdentValueNormal, 700))
	times.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontTimesRoman), ufo.IdentValueNormal, 400))

	result["Serif"] = times
	result["TimesRoman"] = times
}

func (r *ITextFontResolver) addHelvetica(result map[string]*FontFamily) {
	helvetica := NewFontFamily("Helvetica")

	helvetica.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontHelveticaBoldoblique), ufo.IdentValueOblique, 700))
	helvetica.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontHelveticaOblique), ufo.IdentValueOblique, 400))
	helvetica.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontHelveticaBold), ufo.IdentValueNormal, 700))
	helvetica.AddFontDescription(NewFontDescriptionWithStyleWeight(
		r.createFont(writer.BaseFontHelvetica), ufo.IdentValueNormal, 400))

	result["Dialog"] = helvetica
	result["SansSerif"] = helvetica
	result["Helvetica"] = helvetica
}

func (r *ITextFontResolver) addSymbol(result map[string]*FontFamily) {
	fontFamily := NewFontFamily("Symbol")
	fontFamily.AddFontDescription(NewFontDescriptionWithStyleWeight(r.createFontWithEncodingEmbedded(writer.BaseFontSymbol, writer.BaseFontCp1252, false), ufo.IdentValueNormal, 400))
	result["Symbol"] = fontFamily
}

func (r *ITextFontResolver) addZapfDingbats(result map[string]*FontFamily) {
	fontFamily := NewFontFamily("ZapfDingbats")
	fontFamily.AddFontDescription(NewFontDescriptionWithStyleWeight(r.createFontWithEncodingEmbedded(writer.BaseFontZapfdingbats, writer.BaseFontCp1252, false), ufo.IdentValueNormal, 400))
	result["ZapfDingbats"] = fontFamily
}
