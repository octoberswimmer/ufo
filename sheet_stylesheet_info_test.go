// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/StylesheetInfo.java.
// Flying Saucer has no JUnit test for it.

package ufo

import (
	"reflect"
	"testing"
)

func TestStylesheetInfo_mediaTypes(t *testing.T) {
	cases := []struct {
		media    string
		expected []string
	}{
		{"", []string{"all"}},
		{"print", []string{"print"}},
		{" Print , SCREEN ", []string{"print", "screen"}},
		{"print,,screen", []string{"print", "", "screen"}},
		// String.split drops the trailing empty strings.
		{"print,,", []string{"print"}},
		{",", []string{}},
	}
	for _, c := range cases {
		if actual := StylesheetInfoMediaTypes(c.media); !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("StylesheetInfoMediaTypes(%q) = %#v, expected %#v", c.media, actual, c.expected)
		}
	}
}

func TestStylesheetInfo_appliesToMedia(t *testing.T) {
	printOnly := NewStylesheetInfo(StylesheetInfoOriginAuthor, "a.css", []string{"print"}, nil)
	all := NewStylesheetInfo(StylesheetInfoOriginAuthor, "a.css", []string{"all"}, nil)
	cases := []struct {
		info     *StylesheetInfo
		media    string
		expected bool
	}{
		{printOnly, "print", true},
		{printOnly, "PRINT", true},
		{printOnly, "screen", false},
		{printOnly, "All", true},
		{all, "screen", true},
	}
	for i, c := range cases {
		if actual := c.info.AppliesToMedia(c.media); actual != c.expected {
			t.Errorf("case %d: AppliesToMedia(%q) = %v", i, c.media, actual)
		}
	}
}

func TestStylesheetInfo_accessors(t *testing.T) {
	content := "p { color: red }"
	info := NewStylesheetInfo(StylesheetInfoOriginUser, "a.css", []string{"print"}, &content)
	if info.GetUri() != "a.css" || info.GetOrigin() != StylesheetInfoOriginUser ||
		!reflect.DeepEqual(info.GetMedia(), []string{"print"}) || *info.GetContent() != content {
		t.Errorf("the accessors give other values than the constructor got")
	}
	if info.String() != "CSS a.css" {
		t.Errorf("String() = %q", info.String())
	}
	if NewStylesheetInfo(StylesheetInfoOriginUser, "a.css", nil, nil).GetContent() != nil {
		t.Errorf("GetContent() is not nil for a sheet without content")
	}
}

func TestStylesheetInfoOrigin_string(t *testing.T) {
	for origin, expected := range map[StylesheetInfoOrigin]string{
		StylesheetInfoOriginUserAgent: "USER_AGENT",
		StylesheetInfoOriginUser:      "USER",
		StylesheetInfoOriginAuthor:    "AUTHOR",
	} {
		if origin.String() != expected {
			t.Errorf("String() = %q, expected %q", origin.String(), expected)
		}
	}
}
