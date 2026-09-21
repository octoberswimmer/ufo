// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/PDFEncryption.java

package pdf

import (
	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// PDFEncryption holds the encryption settings of a document. The writer does
// not implement encryption: ITextRenderer.CreatePDF passes these values to
// writer.PdfWriter.SetEncryption and returns its
// *writer.UnsupportedFeatureError.
type PDFEncryption struct {
	userPassword      []byte
	ownerPassword     []byte
	allowedPrivileges int
	encryptionType    int
}

func NewPDFEncryption(userPassword []byte, ownerPassword []byte) *PDFEncryption {
	return NewPDFEncryptionWithAllowedPrivileges(userPassword, ownerPassword,
		writer.PdfWriterAllowPrinting|writer.PdfWriterAllowCopy|writer.PdfWriterAllowFillIn)
}

func NewPDFEncryptionWithAllowedPrivileges(userPassword []byte, ownerPassword []byte, allowedPrivileges int) *PDFEncryption {
	return NewPDFEncryptionWithAllowedPrivilegesEncryptionType(userPassword, ownerPassword, allowedPrivileges,
		writer.PdfWriterStandardEncryption128)
}

func NewPDFEncryptionWithAllowedPrivilegesEncryptionType(userPassword []byte, ownerPassword []byte, allowedPrivileges int, encryptionType int) *PDFEncryption {
	return &PDFEncryption{
		userPassword:      ufo.ArrayUtilCloneOrEmptyByteArray(userPassword),
		ownerPassword:     ufo.ArrayUtilCloneOrEmptyByteArray(ownerPassword),
		allowedPrivileges: allowedPrivileges,
		encryptionType:    encryptionType,
	}
}

func (e *PDFEncryption) GetUserPassword() []byte {
	return ufo.ArrayUtilCloneOrEmptyByteArray(e.userPassword)
}

func (e *PDFEncryption) GetOwnerPassword() []byte {
	return ufo.ArrayUtilCloneOrEmptyByteArray(e.ownerPassword)
}

func (e *PDFEncryption) GetAllowedPrivileges() int {
	return e.allowedPrivileges
}

func (e *PDFEncryption) GetEncryptionType() int {
	return e.encryptionType
}
