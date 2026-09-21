// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/UserAgentCallback.java

package ufo

// UserAgentCallback is to be implemented by any user agent using the
// renderer. "User agent" is a term defined by the W3C in the documentation
// for XHTML and CSS; in most cases, you can think of this as the rendering
// component for a browser.
//
// This interface defines a simple callback mechanism for Flying Saucer to
// interact with a user agent. The toolkit provides a default implementation
// for this interface (NaiveUserAgent) which in most cases you can leave as
// is.
//
// The user agent in this case is responsible for retrieving external
// resources. For privacy reasons, if using the library in an application that
// can access URIs in an unrestricted fashion, you may decide to restrict
// access to XML, CSS or images retrieved from external sources; that's one of
// the purposes of the UAC.
type UserAgentCallback interface {
	// GetCSSResource retrieves the CSS at the given URI. This is a
	// synchronous call.
	GetCSSResource(uri string) *CSSResource

	// GetImageResource retrieves the Image at the given URI. This is a
	// synchronous call.
	GetImageResource(uri string) *ImageResource

	// GetXMLResource retrieves the XML at the given URI. This is a
	// synchronous call. It returns nil when the XML could not be loaded.
	GetXMLResource(uri string) *XMLResource

	// GetBinaryResource retrieves a binary resource located at a given URI
	// and returns its contents as a byte slice or nil if the resource could
	// not be loaded.
	GetBinaryResource(uri string) []byte

	// IsVisited normally returns true if the user agent has visited this
	// URI. UserAgent should consider if it should answer truthfully or not
	// for privacy reasons.
	IsVisited(uri string) bool

	// SetBaseURL does not need to be a correct URL, only an identifier that
	// the implementation can resolve. url is a URL against which relative
	// references can be resolved.
	SetBaseURL(url string)

	// GetBaseURL returns the base uri, possibly in the implementations
	// private uri-space.
	GetBaseURL() string

	// ResolveURI is used to find an uri that may be relative to the BaseURL.
	// The returned value will always only be used via methods in the same
	// implementation of this interface, therefore may be a private
	// uri-space. uri is an absolute or relative (to baseURL) uri to be
	// resolved. It returns the full uri in uri-spaces known to the current
	// implementation, or "" when the uri cannot be resolved.
	ResolveURI(uri string) string
}
