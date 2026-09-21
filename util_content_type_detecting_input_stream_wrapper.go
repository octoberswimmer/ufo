// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/ContentTypeDetectingInputStreamWrapper.java

package ufo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

var (
	contentTypeDetectingMagicBytesPdf = []byte("%PDF")
	contentTypeDetectingMagicBytesXml = []byte("<?xm")
	contentTypeDetectingMagicBytesSvg = []byte("<svg")
	contentTypeDetectingCommentStart  = []byte("<!--")
	contentTypeDetectingCommentEnd    = []byte("-->")
	contentTypeDetectingNoData        = []byte{}
	contentTypeDetectingUtf8Bom       = []byte{0xEF, 0xBB, 0xBF}
)

const (
	contentTypeDetectingMaxMagicBytes      = 4
	contentTypeDetectingMaxPrologScanBytes = 4096
)

// ContentTypeDetectingInputStreamWrapper is a buffered reader over a stream
// that tells whether the stream starts like a PDF or an SVG document without
// consuming what it looked at. Java extends BufferedInputStream and looks
// ahead with mark and reset; here the look-ahead is bufio.Reader.Peek, which
// leaves the read position where it was.
type ContentTypeDetectingInputStreamWrapper struct {
	*bufio.Reader
	source     io.Reader
	firstBytes []byte
}

// ContentTypeDetectingInputStreamWrapperDetectContentType returns nil for a
// nil stream.
func ContentTypeDetectingInputStreamWrapperDetectContentType(is io.Reader) (*ContentTypeDetectingInputStreamWrapper, error) {
	if is == nil {
		return nil, nil
	}
	return newContentTypeDetectingInputStreamWrapper(is)
}

func newContentTypeDetectingInputStreamWrapper(source io.Reader) (*ContentTypeDetectingInputStreamWrapper, error) {
	w := &ContentTypeDetectingInputStreamWrapper{
		// The buffer holds the longest look-ahead, the prolog scan.
		Reader: bufio.NewReaderSize(source, 2*contentTypeDetectingMaxPrologScanBytes),
		source: source,
	}
	if err := contentTypeDetectingSkipUtf8BomIfPresent(w.Reader); err != nil {
		return nil, err
	}
	firstBytes, err := contentTypeDetectingReadFirstBytes(w.Reader, contentTypeDetectingMaxMagicBytes)
	if err != nil {
		return nil, err
	}
	w.firstBytes = firstBytes
	return w, nil
}

// Close closes the wrapped stream when it has a Close method.
func (w *ContentTypeDetectingInputStreamWrapper) Close() error {
	if closer, ok := w.source.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// contentTypeDetectingPeek returns up to count bytes from the read position
// without consuming them. Reaching the end of the stream first is not an
// error.
func contentTypeDetectingPeek(in *bufio.Reader, count int) ([]byte, error) {
	head, err := in.Peek(count)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return head, nil
}

// If the wrapped stream begins with a UTF-8 BOM (EF BB BF), consume it so
// firstBytes captures the real first glyph. Otherwise the magic-byte
// comparison fails for BOM-prefixed XML / SVG written by text editors that
// auto-prepend a BOM (e.g. Notepad on Windows), and IsSvg returns false for
// what is clearly an SVG.
func contentTypeDetectingSkipUtf8BomIfPresent(in *bufio.Reader) error {
	head, err := contentTypeDetectingPeek(in, len(contentTypeDetectingUtf8Bom))
	if err != nil {
		return err
	}
	if len(head) == len(contentTypeDetectingUtf8Bom) && bytes.Equal(head, contentTypeDetectingUtf8Bom) {
		// BOM consumed; subsequent reads see "<svg…" / "<?xml…"
		_, err := in.Discard(len(contentTypeDetectingUtf8Bom))
		return err
	}
	return nil
}

func contentTypeDetectingReadFirstBytes(in *bufio.Reader, count int) ([]byte, error) {
	head, err := contentTypeDetectingPeek(in, count)
	if err != nil {
		return nil, err
	}
	if len(head) <= 0 {
		return contentTypeDetectingNoData, nil
	}
	// A shorter result means there is not enough data in the stream.
	return bytes.Clone(head), nil
}

func (w *ContentTypeDetectingInputStreamWrapper) streamStartsWithMagicBytes(magic []byte) bool {
	return bytes.Equal(w.firstBytes, magic)
}

func (w *ContentTypeDetectingInputStreamWrapper) IsPdf() bool {
	return w.streamStartsWithMagicBytes(contentTypeDetectingMagicBytesPdf)
}

func (w *ContentTypeDetectingInputStreamWrapper) IsSvg() (bool, error) {
	if w.streamStartsWithMagicBytes(contentTypeDetectingMagicBytesXml) ||
		w.streamStartsWithMagicBytes(contentTypeDetectingMagicBytesSvg) {
		return true, nil
	}
	return w.startsWithSvgAfterLeadingComments()
}

// contentTypeDetectingScan reads the looked-ahead bytes of the prolog scan.
// remaining is the number of bytes the scan may still read before it exceeds
// the cap, the int[] remaining of the Java methods.
type contentTypeDetectingScan struct {
	data      []byte
	pos       int
	remaining int
}

// Handles SVGs whose prolog opens with an XML comment (e.g. a real XML
// editor or export tool commenting out the declaration:
// <!--<?xml version="1.0" encoding="UTF-8" standalone="no"?>-->)
// before the <svg> or <?xml magic bytes.
//
// The scan is capped at contentTypeDetectingMaxPrologScanBytes: source may
// be attacker-controlled (e.g. a remote URL referenced from untrusted
// HTML/CSS), so an unbounded search for a comment terminator that never
// arrives would let the buffer keep growing for as long as bytes keep
// arriving, which exhausts memory. Exceeding the cap returns an error rather
// than silently misdetecting, since at that point continuing to buffer is
// unsafe, and genuinely running out of input (EOF within the cap) already
// resolves to false on its own.
//
// The scan reads looked-ahead bytes only, so callers that go on to read the
// full SVG body still see it from the very start, comments included.
func (w *ContentTypeDetectingInputStreamWrapper) startsWithSvgAfterLeadingComments() (bool, error) {
	data, err := contentTypeDetectingPeek(w.Reader, contentTypeDetectingMaxPrologScanBytes)
	if err != nil {
		return false, err
	}
	scan := &contentTypeDetectingScan{data: data, remaining: contentTypeDetectingMaxPrologScanBytes}
	b, err := scan.readBounded()
	if err != nil {
		return false, err
	}
	b, err = scan.skipWhitespace(b)
	if err != nil {
		return false, err
	}
	for b != -1 {
		rest, err := scan.readBoundedNBytes(contentTypeDetectingMaxMagicBytes - 1)
		if err != nil {
			return false, err
		}
		if len(rest) < contentTypeDetectingMaxMagicBytes-1 {
			return false, nil // fewer than 4 bytes left
		}
		if contentTypeDetectingMatches(b, rest, contentTypeDetectingCommentStart) {
			closed, err := scan.skipToCommentEnd()
			if err != nil {
				return false, err
			}
			if !closed {
				return false, nil // EOF within the scan cap, comment never closed
			}
			b, err = scan.readBounded()
			if err != nil {
				return false, err
			}
			b, err = scan.skipWhitespace(b)
			if err != nil {
				return false, err
			}
			continue
		}
		return contentTypeDetectingMatches(b, rest, contentTypeDetectingMagicBytesXml) ||
			contentTypeDetectingMatches(b, rest, contentTypeDetectingMagicBytesSvg), nil
	}
	return false, nil
}

// contentTypeDetectingMatches compares firstByte followed by rest against a
// 4-byte magic pattern.
func contentTypeDetectingMatches(firstByte int, rest []byte, pattern []byte) bool {
	if firstByte != int(pattern[0]) {
		return false
	}
	for i := range rest {
		if rest[i] != pattern[i+1] {
			return false
		}
	}
	return true
}

func (s *contentTypeDetectingScan) skipWhitespace(b int) (int, error) {
	for b != -1 && contentTypeDetectingIsXmlWhitespace(byte(b)) {
		var err error
		b, err = s.readBounded()
		if err != nil {
			return -1, err
		}
	}
	return b, nil
}

// skipToCommentEnd returns true once --> is found; false on EOF (within the
// scan cap) first.
func (s *contentTypeDetectingScan) skipToCommentEnd() (bool, error) {
	matched := 0
	for {
		b, err := s.readBounded()
		if err != nil {
			return false, err
		}
		if b == -1 {
			return false, nil
		}
		if b == int(contentTypeDetectingCommentEnd[matched]) {
			matched++
			if matched == len(contentTypeDetectingCommentEnd) {
				return true, nil
			}
		} else {
			if b == int(contentTypeDetectingCommentEnd[0]) {
				matched = 1
			} else {
				matched = 0
			}
		}
	}
}

// readBounded returns the next byte, or -1 at the end of the stream.
func (s *contentTypeDetectingScan) readBounded() (int, error) {
	if s.remaining <= 0 {
		return -1, fmt.Errorf("SVG detection aborted: prolog scan exceeded %d bytes", contentTypeDetectingMaxPrologScanBytes)
	}
	s.remaining--
	if s.pos >= len(s.data) {
		return -1, nil
	}
	b := s.data[s.pos]
	s.pos++
	return int(b), nil
}

func (s *contentTypeDetectingScan) readBoundedNBytes(length int) ([]byte, error) {
	if s.remaining < length {
		return nil, fmt.Errorf("SVG detection aborted: prolog scan exceeded %d bytes", contentTypeDetectingMaxPrologScanBytes)
	}
	s.remaining -= length
	end := min(s.pos+length, len(s.data))
	result := s.data[s.pos:end]
	s.pos = end
	return result, nil
}

func contentTypeDetectingIsXmlWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
