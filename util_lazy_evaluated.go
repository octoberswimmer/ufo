// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/LazyEvaluated.java

package ufo

import "sync"

// LazyEvaluated holds a value that is computed by the supplier on the first
// call of Get. Java detects "not computed yet" by the value being null and
// throws when the supplier returns null; a Go type parameter has no null, so
// a flag records that the supplier ran.
type LazyEvaluated[T any] struct {
	mu       sync.Mutex
	computed bool
	value    T
	supplier func() T
}

func (l *LazyEvaluated[T]) Get() T {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.computed {
		l.value = l.supplier()
		l.computed = true
	}
	return l.value
}

func LazyEvaluatedLazy[T any](supplier func() T) *LazyEvaluated[T] {
	return &LazyEvaluated[T]{supplier: supplier}
}
