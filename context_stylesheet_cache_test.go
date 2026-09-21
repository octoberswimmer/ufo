// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/context/StylesheetCacheTest.java

package ufo

import (
	"strconv"
	"testing"
)

func TestStylesheetCache_holdsNoMoreThan16Entries(t *testing.T) {
	cache := NewStylesheetCache()
	for i := 0; i < 17; i++ {
		cache.Put("key#"+strconv.Itoa(i), NewStylesheet("https://"+strconv.Itoa(i), StylesheetInfoOriginAuthor))
	}

	if cache.Size() != 16 {
		t.Errorf("size = %d, want 16", cache.Size())
	}
	if cache.ContainsKey("key#0") {
		t.Errorf("cache contains key#0")
	}
	if !cache.ContainsKey("key#1") {
		t.Errorf("cache does not contain key#1")
	}
	if !cache.ContainsKey("key#16") {
		t.Errorf("cache does not contain key#16")
	}
}

// The Java class inherits access ordering from LinkedHashMap; the JUnit test
// covers insertion order only.
func TestStylesheetCache_evictsLeastRecentlyUsedEntry(t *testing.T) {
	cache := NewStylesheetCache()
	for i := 0; i < 16; i++ {
		cache.Put("key#"+strconv.Itoa(i), NewStylesheet("https://"+strconv.Itoa(i), StylesheetInfoOriginAuthor))
	}
	cache.Get("key#0")
	cache.ContainsKey("key#1")
	cache.Put("key#2", nil)
	cache.Put("key#16", NewStylesheet("https://16", StylesheetInfoOriginAuthor))
	cache.Put("key#17", NewStylesheet("https://17", StylesheetInfoOriginAuthor))

	if cache.Size() != 16 {
		t.Errorf("size = %d, want 16", cache.Size())
	}
	for _, key := range []string{"key#0", "key#2", "key#4", "key#16", "key#17"} {
		if !cache.ContainsKey(key) {
			t.Errorf("cache does not contain %s", key)
		}
	}
	for _, key := range []string{"key#1", "key#3"} {
		if cache.ContainsKey(key) {
			t.Errorf("cache contains %s", key)
		}
	}
	if cache.Get("key#2") != nil {
		t.Errorf("key#2 is not mapped to nil")
	}
}
