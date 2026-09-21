// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/FontUtil.java

package ufo

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/url"
	"strings"
)

func FontUtilIsEmbeddedBase64Font(uri string) bool {
	return strings.HasPrefix(uri, "data:font/")
}

// FontUtilGetEmbeddedBase64Data returns the decoded data of a base 64 data
// URI, or nil when the URI is not base 64 encoded.
func FontUtilGetEmbeddedBase64Data(uri string) io.Reader {
	b64Index := strings.Index(uri, "base64,")
	if b64Index != -1 {
		b64encoded := uri[b64Index+len("base64,"):]
		return bytes.NewReader(utilDecodeBase64DataUriPayload(b64encoded))
	}
	XRLogLoadWithLevel(LevelSevere, "Embedded css fonts must be encoded in base 64.")
	return nil
}

// utilDecodeBase64DataUriPayload does what FontUtil.getEmbeddedBase64Data and
// ImageUtil.getEmbeddedBase64Image both do with the text after "base64,":
// URLDecoder.decode(text, US_ASCII) when the text contains a '%', then
// Base64.getDecoder().decode. Both Java calls throw
// IllegalArgumentException on malformed input, which is a panic here.
func utilDecodeBase64DataUriPayload(b64encoded string) []byte {
	if strings.Contains(b64encoded, "%") {
		// URLDecoder turns '+' into a space and %XX into a byte, as
		// url.QueryUnescape does.
		decoded, err := url.QueryUnescape(b64encoded)
		if err != nil {
			panic(NewXRRuntimeExceptionWithCause("URLDecoder: "+err.Error(), err))
		}
		b64encoded = decoded
	}
	// The JDK's basic decoder accepts input with or without the trailing
	// padding.
	unpadded := b64encoded
	for i := 0; i < 2 && strings.HasSuffix(unpadded, "="); i++ {
		unpadded = unpadded[:len(unpadded)-1]
	}
	data, err := base64.RawStdEncoding.DecodeString(unpadded)
	if err != nil {
		panic(NewXRRuntimeExceptionWithCause("Illegal base64 data: "+err.Error(), err))
	}
	return data
}
