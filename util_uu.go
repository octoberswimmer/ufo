// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/Uu.java

package ufo

// UuPString ports Uu.p(String).
func UuPString(message string) {
	if XRLogIsLoggingEnabled() {
		XRLogGeneral(message)
	}
}

// UuPException ports Uu.p(Exception).
func UuPException(e error) {
	if XRLogIsLoggingEnabled() {
		XRLogExceptionWithTh(e.Error(), e)
	}
}
