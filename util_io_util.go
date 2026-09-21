// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/IOUtil.java

package ufo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func IOUtilCopyFile(page string, outputDir string) error {
	outputFile := filepath.Join(outputDir, filepath.Base(page))
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()
	buffered := bufio.NewWriter(out)
	in, err := os.Open(page)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := ioUtilCopyBytes(bufio.NewReader(in), buffered); err != nil {
		return err
	}
	if err := buffered.Flush(); err != nil {
		return err
	}
	return out.Close()
}

func ioUtilCopyBytes(in io.Reader, out io.Writer) error {
	buf := make([]byte, 1024)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func IOUtilDeleteAllFiles(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		// File.listFiles returns null for a path that is not a readable
		// directory, and the Java method then does nothing.
		return nil
	}
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("Cleanup directory %s, can't delete file %s", dir, path)
		}
	}
	return nil
}

// ioUtilReadCloser joins a reader with the closer of the stream it reads.
type ioUtilReadCloser struct {
	io.Reader
	closer io.Closer
}

func (r *ioUtilReadCloser) Close() error {
	return r.closer.Close()
}

// ioUtilMalformedURLError is the counterpart of java.net.MalformedURLException
// and java.net.URISyntaxException.
type ioUtilMalformedURLError struct {
	msg string
}

func (e *ioUtilMalformedURLError) Error() string {
	return e.msg
}

// ioUtilOpenURL is URL.openConnection().getInputStream() for the schemes
// that the port reads: "file", "http", "https", and "classpath" for the
// files that GeneralUtilGetURLFromClasspath names. A zero timeout means no
// timeout, as for a java.net.URLConnection that was not given one.
func ioUtilOpenURL(uri string, connectTimeout, readTimeout time.Duration, accept string) (io.ReadCloser, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, &ioUtilMalformedURLError{msg: err.Error()}
	}
	switch parsed.Scheme {
	case "":
		return nil, &ioUtilMalformedURLError{msg: "no protocol: " + uri}
	case "file":
		path := parsed.Path
		if parsed.Opaque != "" {
			path = parsed.Opaque
		}
		return os.Open(filepath.FromSlash(path))
	case generalUtilClasspathScheme:
		name := generalUtilClasspathName(parsed.Path)
		if parsed.Opaque != "" {
			name = generalUtilClasspathName(parsed.Opaque)
		}
		return generalUtilClasspathOpen(name)
	case "http", "https":
		client := &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: connectTimeout}).DialContext,
				ResponseHeaderTimeout: readTimeout,
			},
		}
		request, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			return nil, &ioUtilMalformedURLError{msg: err.Error()}
		}
		if accept != "" {
			request.Header.Set("Accept", accept)
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
			// HttpURLConnection.getInputStream throws FileNotFoundException
			// for these two status codes.
			response.Body.Close()
			return nil, &fs.PathError{Op: "open", Path: uri, Err: fs.ErrNotExist}
		}
		if response.StatusCode >= 400 {
			response.Body.Close()
			return nil, fmt.Errorf("Server returned HTTP response code: %d for URL: %s", response.StatusCode, uri)
		}
		return response.Body, nil
	default:
		return nil, &ioUtilMalformedURLError{msg: "unknown protocol: " + parsed.Scheme}
	}
}

// ioUtilLogOpenFailure logs a failure to open uri the way the catch clauses
// of IOUtil.openStreamAtUrl and IOUtil.getInputStream do.
func ioUtilLogOpenFailure(uri string, err error) {
	var malformed *ioUtilMalformedURLError
	if errors.As(err, &malformed) {
		XRLogExceptionWithTh("bad URL given: "+uri, err)
	} else if errors.Is(err, fs.ErrNotExist) {
		XRLogException("item at URI " + uri + " not found (caused by: " + err.Error() + ")")
	} else {
		XRLogExceptionWithTh("IO problem for "+uri, err)
	}
}

// IOUtilOpenStreamAtUrl attempts to open a connection, and a stream, to the
// URI provided. timeouts will be set for opening the connection and reading
// from it. will return the stream, or nil if unable to open or read or a
// timeout occurred. Does not buffer the stream.
func IOUtilOpenStreamAtUrl(uri string) io.ReadCloser {
	stream, err := IOUtilStreamAtUrl(uri)
	if err != nil {
		ioUtilLogOpenFailure(uri, err)
		return nil
	}
	return stream
}

func IOUtilStreamAtUrl(uri string) (io.ReadCloser, error) {
	return ioUtilOpenURL(uri, 10*1000*time.Millisecond, 30*1000*time.Millisecond, "*/*")
}

// IOUtilGetInputStream gets a buffered stream for the resource identified,
// or nil when uri is empty or the resource can't be opened.
func IOUtilGetInputStream(uri string) io.ReadCloser {
	if uri == "" {
		return nil
	}
	stream, err := ioUtilOpenURL(uri, 0, 0, "")
	if err != nil {
		ioUtilLogOpenFailure(uri, err)
		return nil
	}
	return &ioUtilReadCloser{Reader: bufio.NewReader(stream), closer: stream}
}

// IOUtilReadBytesString ports IOUtil.readBytes(String uri). It returns nil
// when the resource can't be read.
func IOUtilReadBytesString(uri string) []byte {
	is := IOUtilGetInputStream(uri)
	if is == nil {
		return nil
	}
	defer is.Close()
	result, err := IOUtilReadBytesInputStream(is)
	if err != nil {
		XRLogLoadWithTh(LevelWarning, "Unable to read "+uri, err)
		return nil
	}
	return result
}

// IOUtilReadBytesPath ports IOUtil.readBytes(Path file).
func IOUtilReadBytesPath(file string) ([]byte, error) {
	is, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer is.Close()
	return IOUtilReadBytesInputStream(is)
}

// IOUtilReadBytesInputStream ports IOUtil.readBytes(InputStream is).
func IOUtilReadBytesInputStream(is io.Reader) ([]byte, error) {
	var result bytes.Buffer
	if err := ioUtilCopyBytes(is, &result); err != nil {
		return nil, err
	}
	if result.Len() == 0 {
		return []byte{}, nil
	}
	return result.Bytes(), nil
}

// Deprecated: close the stream where it was opened instead.
func IOUtilClose(in io.Closer) {
	if in != nil {
		_ = in.Close()
	}
}
