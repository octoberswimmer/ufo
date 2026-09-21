// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/StylesheetInfo.java

package ufo

import "strings"

// StylesheetInfo is a reference to a stylesheet. If no stylesheet is set, the
// matcher will try to find the stylesheet by uri, first from the
// StylesheetFactory cache, then by loading the uri if it is not cached.
//
// Therefore, either a stylesheet must be set, or an uri must be set.
//
// Origin defaults to USER_AGENT and media defaults to "all".
type StylesheetInfo struct {
	uri        string
	origin     StylesheetInfoOrigin
	mediaTypes []string
	// content is nil when the stylesheet has no inline content.
	content *string
}

// StylesheetInfoOrigin is the origin of a stylesheet (the Java enum
// StylesheetInfo.Origin).
type StylesheetInfoOrigin int

const (
	StylesheetInfoOriginUserAgent StylesheetInfoOrigin = iota
	StylesheetInfoOriginUser
	StylesheetInfoOriginAuthor
)

// String returns the name of the Java enum constant.
func (o StylesheetInfoOrigin) String() string {
	switch o {
	case StylesheetInfoOriginUserAgent:
		return "USER_AGENT"
	case StylesheetInfoOriginUser:
		return "USER"
	case StylesheetInfoOriginAuthor:
		return "AUTHOR"
	}
	panic(NewXRRuntimeException("unknown StylesheetInfo.Origin"))
}

func (o StylesheetInfoOrigin) ToString() string {
	return o.String()
}

// NewStylesheetInfo creates a StylesheetInfo. content is nil when there is no
// inline content.
func NewStylesheetInfo(origin StylesheetInfoOrigin, uri string, mediaTypes []string, content *string) *StylesheetInfo {
	return &StylesheetInfo{
		origin:     origin,
		uri:        uri,
		mediaTypes: mediaTypes,
		content:    content,
	}
}

// AppliesToMedia checks if this stylesheet applies to the given medium.
// media is a single media identifier. It returns true if the stylesheet
// referenced applies to the medium.
func (s *StylesheetInfo) AppliesToMedia(media string) bool {
	mLowerCase := strings.ToLower(media)
	return mLowerCase == "all" ||
		stylesheetInfoContains(s.mediaTypes, "all") || stylesheetInfoContains(s.mediaTypes, mLowerCase)
}

func stylesheetInfoContains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func StylesheetInfoMediaTypes(media string) []string {
	if media == "" {
		//default for HTML is "screen", but that is silly and firefox seems to assume "all"
		return []string{"all"}
	}

	parts := strings.Split(media, ",")
	// String.split drops trailing empty strings.
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	result := make([]string, 0, len(parts))
	for _, mediaType := range parts {
		// String.trim removes the characters up to U+0020 from both ends.
		trimmed := strings.TrimFunc(mediaType, func(r rune) bool { return r <= ' ' })
		result = append(result, strings.ToLower(trimmed))
	}
	return result
}

// GetUri gets the uri attribute of the StylesheetInfo object.
func (s *StylesheetInfo) GetUri() string {
	return s.uri
}

// GetMedia gets the media attribute of the StylesheetInfo object.
func (s *StylesheetInfo) GetMedia() []string {
	return s.mediaTypes
}

// GetOrigin gets the origin attribute of the StylesheetInfo object.
func (s *StylesheetInfo) GetOrigin() StylesheetInfoOrigin {
	return s.origin
}

// GetContent returns the inline content, or nil when there is none
// (Optional.empty in Java).
func (s *StylesheetInfo) GetContent() *string {
	return s.content
}

func (s *StylesheetInfo) String() string {
	return "CSS " + s.uri
}

func (s *StylesheetInfo) ToString() string {
	return s.String()
}
