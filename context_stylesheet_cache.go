// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/context/StylesheetCache.java

package ufo

import "container/list"

const stylesheetCacheCacheCapacity = 16

// StylesheetCache is a map from a stylesheet key to the stylesheet that holds
// at most 16 entries. Java extends LinkedHashMap in access order with
// removeEldestEntry: Get and Put move the entry to the most recently used
// end, ContainsKey does not, and a Put that makes the size exceed the
// capacity removes the least recently used entry. The order is kept in a
// linked list whose front is the eldest entry. A key may be mapped to a nil
// stylesheet.
type StylesheetCache struct {
	order   *list.List
	entries map[string]*list.Element
}

type stylesheetCacheEntry struct {
	key   string
	value *Stylesheet
}

func NewStylesheetCache() *StylesheetCache {
	return &StylesheetCache{
		order:   list.New(),
		entries: make(map[string]*list.Element, stylesheetCacheCacheCapacity),
	}
}

// Put maps key to sheet and returns the stylesheet the key was mapped to
// before, or nil.
func (c *StylesheetCache) Put(key string, sheet *Stylesheet) *Stylesheet {
	if element, ok := c.entries[key]; ok {
		entry := element.Value.(*stylesheetCacheEntry)
		old := entry.value
		entry.value = sheet
		c.order.MoveToBack(element)
		return old
	}
	c.entries[key] = c.order.PushBack(&stylesheetCacheEntry{key: key, value: sheet})
	if c.removeEldestEntry() {
		eldest := c.order.Front()
		c.order.Remove(eldest)
		delete(c.entries, eldest.Value.(*stylesheetCacheEntry).key)
	}
	return nil
}

func (c *StylesheetCache) removeEldestEntry() bool {
	return c.Size() > stylesheetCacheCacheCapacity
}

// Get returns the stylesheet the key is mapped to, or nil when the key is
// absent or mapped to nil.
func (c *StylesheetCache) Get(key string) *Stylesheet {
	element, ok := c.entries[key]
	if !ok {
		return nil
	}
	c.order.MoveToBack(element)
	return element.Value.(*stylesheetCacheEntry).value
}

func (c *StylesheetCache) ContainsKey(key string) bool {
	_, ok := c.entries[key]
	return ok
}

// Remove removes the key and returns the stylesheet it was mapped to, or nil.
func (c *StylesheetCache) Remove(key string) *Stylesheet {
	element, ok := c.entries[key]
	if !ok {
		return nil
	}
	c.order.Remove(element)
	delete(c.entries, key)
	return element.Value.(*stylesheetCacheEntry).value
}

func (c *StylesheetCache) Clear() {
	c.order.Init()
	clear(c.entries)
}

func (c *StylesheetCache) Size() int {
	return len(c.entries)
}
