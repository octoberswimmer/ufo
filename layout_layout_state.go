// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/LayoutState.java

package ufo

// LayoutState is a bean which captures all state necessary to lay out an
// arbitrary box. Mutable objects must be copied when provided to this class.
// It is far too expensive to maintain a bean of this class for each box.
// It is only created as needed.
type LayoutState struct {
	firstLines   *StyleTracker
	firstLetters *StyleTracker

	currentMarkerData *MarkerData

	bfcs []*BlockFormattingContext

	pageName         string
	extraSpaceTop    int
	extraSpaceBottom int
	noPageBreak      int
}

// NewLayoutStateWithPageNameExtraSpaceTopExtraSpaceBottomNoPageBreak copies
// blockFormattingContexts, as Java's new ArrayDeque<>(collection) does.
// currentMarkerData may be nil; pageName is "" for Java's null.
func NewLayoutStateWithPageNameExtraSpaceTopExtraSpaceBottomNoPageBreak(firstLines *StyleTracker, firstLetters *StyleTracker, currentMarkerData *MarkerData,
	blockFormattingContexts []*BlockFormattingContext,
	pageName string, extraSpaceTop int, extraSpaceBottom int, noPageBreak int) *LayoutState {
	bfcs := make([]*BlockFormattingContext, len(blockFormattingContexts))
	copy(bfcs, blockFormattingContexts)
	return &LayoutState{
		firstLines:        firstLines,
		firstLetters:      firstLetters,
		currentMarkerData: currentMarkerData,
		bfcs:              bfcs,
		pageName:          pageName,
		extraSpaceTop:     extraSpaceTop,
		extraSpaceBottom:  extraSpaceBottom,
		noPageBreak:       noPageBreak,
	}
}

func NewLayoutState(firstLines *StyleTracker, firstLetters *StyleTracker, currentMarkerData *MarkerData,
	blockFormattingContexts []*BlockFormattingContext) *LayoutState {
	return NewLayoutStateWithPageNameExtraSpaceTopExtraSpaceBottomNoPageBreak(firstLines, firstLetters, currentMarkerData, blockFormattingContexts, "", 0, 0, 0)
}

// GetBFCs returns the contexts from the outermost to the innermost (the
// iteration order of Java's Deque).
func (l *LayoutState) GetBFCs() []*BlockFormattingContext {
	return l.bfcs
}

// GetCurrentMarkerData may return nil.
func (l *LayoutState) GetCurrentMarkerData() *MarkerData {
	return l.currentMarkerData
}

func (l *LayoutState) GetFirstLetters() *StyleTracker {
	return l.firstLetters
}

func (l *LayoutState) GetFirstLines() *StyleTracker {
	return l.firstLines
}

// GetPageName returns "" for Java's null.
func (l *LayoutState) GetPageName() string {
	return l.pageName
}

func (l *LayoutState) GetExtraSpaceTop() int {
	return l.extraSpaceTop
}

func (l *LayoutState) GetExtraSpaceBottom() int {
	return l.extraSpaceBottom
}

func (l *LayoutState) GetNoPageBreak() int {
	return l.noPageBreak
}
