// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/UrlAwareLineBreakIterator.java

package ufo

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const urlAwareLineBreakIteratorBreakingChars = ".,:;!?- \n\r\t/"

// UrlAwareLineBreakIterator is a BreakIterator implementation that improves
// line breaking for URLs. Break points are supported before path fragments.
//
// Positions are byte offsets. Every character this class looks for ("://",
// "/" and the breaking characters) is one byte long in UTF-8, so a range
// start or stop moved past such a character stays on a character boundary.
type UrlAwareLineBreakIterator struct {
	delegate     *BreakIterator
	text         string
	currentRange *urlAwareLineBreakIteratorRange
}

var _ BreakIteratorI = (*UrlAwareLineBreakIterator)(nil)

func NewUrlAwareLineBreakIterator(text string) *UrlAwareLineBreakIterator {
	i := &UrlAwareLineBreakIterator{delegate: BreakIteratorGetLineInstance()}
	i.SetText(text)
	return i
}

func (i *UrlAwareLineBreakIterator) Preceding(offset int) int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) Last() int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) Previous() int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) Next() int {
	i.checkNotAheadOfDelegate()

	searchRange := i.currentRange // the range in which we search for slashes

	if i.isDelegateInSync() {
		reachedEnd := i.advanceDelegate()
		if reachedEnd {
			return BreakIteratorDone
		}

		if "://" == i.substring(newUrlAwareLineBreakIteratorRangeWithOffsets(i.currentRange.getStop(), -1, 2)) {
			searchRange = searchRange.withStart(i.currentRange.getStop() + 2)
			i.advanceDelegate() // no reached-end check needed here, because there are at least two slashes ahead
		}
	}
	searchRange = searchRange.withStop(i.currentRange.getStop())

	searchRange = i.trimSearchRange(searchRange)
	nextSlash := i.findSlashInRange(searchRange)
	if nextSlash > -1 {
		i.currentRange = i.currentRange.withStart(nextSlash)
	} else {
		i.currentRange = i.currentRange.withStart(i.delegate.Current())
	}

	return i.currentRange.getStart()
}

func (i *UrlAwareLineBreakIterator) trimSearchRange(searchRange *urlAwareLineBreakIteratorRange) *urlAwareLineBreakIteratorRange {
	// Exclude leading breaking characters (should really only be a slash).
	for searchRange.getStart() < i.currentRange.getStop() {
		r, _ := utf8.DecodeRuneInString(i.text[searchRange.getStart():])
		if !(strings.IndexRune(urlAwareLineBreakIteratorBreakingChars, r) > -1) {
			break
		}
		// a breaking character is one byte long
		searchRange = searchRange.incrementStart()
	}

	// Exclude trailing breaking characters.
	for searchRange.getStop() > searchRange.getStart() {
		r, _ := utf8.DecodeLastRuneInString(i.text[:searchRange.getStop()])
		if !(strings.IndexRune(urlAwareLineBreakIteratorBreakingChars, r) > -1) {
			break
		}
		// a breaking character is one byte long
		searchRange = searchRange.decrementStop()
	}

	return searchRange
}

func (i *UrlAwareLineBreakIterator) findSlashInRange(searchRange *urlAwareLineBreakIteratorRange) int {
	// String.indexOf(ch, fromIndex) treats a negative fromIndex as 0 and
	// returns -1 for a fromIndex past the end.
	from := searchRange.getStart()
	if from < 0 {
		from = 0
	}
	nextSlash := -1
	if from < len(i.text) {
		if n := strings.IndexByte(i.text[from:], '/'); n > -1 {
			nextSlash = from + n
		}
	}
	if nextSlash < searchRange.getStop() {
		return nextSlash
	}
	return -1
}

func (i *UrlAwareLineBreakIterator) substring(r *urlAwareLineBreakIteratorRange) string {
	return i.text[max(0, r.getStart()):min(len(i.text), r.getStop())]
}

func (i *UrlAwareLineBreakIterator) checkNotAheadOfDelegate() {
	// This is a sanity check. We should never be in this state.
	if i.currentRange.getStart() > i.delegate.Current() {
		panic(NewXRRuntimeException("Iterator ahead of delegate."))
	}
}

func (i *UrlAwareLineBreakIterator) isDelegateInSync() bool {
	return i.currentRange.getStart() == i.delegate.Current()
}

func (i *UrlAwareLineBreakIterator) advanceDelegate() bool {
	next := i.delegate.Next()
	i.currentRange = i.currentRange.withStop(next)
	return next == BreakIteratorDone
}

func (i *UrlAwareLineBreakIterator) NextWithN(n int) int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) IsBoundary(offset int) bool {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) Following(offset int) int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) First() int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) Current() int {
	panic(NewXRRuntimeException("Not yet implemented"))
}

func (i *UrlAwareLineBreakIterator) GetText() string {
	return i.delegate.GetText()
}

func (i *UrlAwareLineBreakIterator) SetText(newText string) {
	i.delegate.SetText(newText)
	i.text = newText
	i.currentRange = newUrlAwareLineBreakIteratorRange(i.delegate.Current(), i.delegate.Current())
}

// urlAwareLineBreakIteratorRange is the private class
// UrlAwareLineBreakIterator.Range.
type urlAwareLineBreakIteratorRange struct {
	start int
	stop  int
}

func newUrlAwareLineBreakIteratorRange(start int, stop int) *urlAwareLineBreakIteratorRange {
	return &urlAwareLineBreakIteratorRange{start: start, stop: max(start, stop)}
}

func newUrlAwareLineBreakIteratorRangeWithOffsets(referencePoint int, startOffset int, stopOffset int) *urlAwareLineBreakIteratorRange {
	return newUrlAwareLineBreakIteratorRange(referencePoint+startOffset, referencePoint+stopOffset)
}

func (r *urlAwareLineBreakIteratorRange) withStart(start int) *urlAwareLineBreakIteratorRange {
	return newUrlAwareLineBreakIteratorRange(start, r.stop)
}

func (r *urlAwareLineBreakIteratorRange) withStop(stop int) *urlAwareLineBreakIteratorRange {
	return newUrlAwareLineBreakIteratorRange(r.start, stop)
}

func (r *urlAwareLineBreakIteratorRange) incrementStart() *urlAwareLineBreakIteratorRange {
	newStart := r.start + 1
	return newUrlAwareLineBreakIteratorRange(newStart, max(newStart, r.stop))
}

func (r *urlAwareLineBreakIteratorRange) decrementStop() *urlAwareLineBreakIteratorRange {
	newStop := r.stop - 1
	return newUrlAwareLineBreakIteratorRange(min(r.start, newStop), newStop)
}

func (r *urlAwareLineBreakIteratorRange) getStart() int {
	return r.start
}

func (r *urlAwareLineBreakIteratorRange) getStop() int {
	return r.stop
}

func (r *urlAwareLineBreakIteratorRange) String() string {
	return "[" + strconv.Itoa(r.start) + ", " + strconv.Itoa(r.stop) + ")"
}

func (r *urlAwareLineBreakIteratorRange) ToString() string {
	return r.String()
}
