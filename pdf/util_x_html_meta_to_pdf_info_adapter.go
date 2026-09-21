// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/util/XHtmlMetaToPdfInfoAdapter.java

package pdf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

const (
	xHtmlMetaToPdfInfoAdapterHtmlTagTitle         = "title"
	xHtmlMetaToPdfInfoAdapterHtmlTagHead          = "head"
	xHtmlMetaToPdfInfoAdapterHtmlTagMeta          = "meta"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeyTitle     = "title"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeyDcTitle   = "DC.title"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeyCreator   = "creator"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeyDcCreator = "DC.creator"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeySubject   = "subject"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeyDcSubject = "DC.subject"
	xHtmlMetaToPdfInfoAdapterHtmlMetaKeyKeywords  = "keywords"
	xHtmlMetaToPdfInfoAdapterHtmlMetaAttrName     = "name"
	xHtmlMetaToPdfInfoAdapterHtmlMetaAttrContent  = "content"
)

// XHtmlMetaToPdfInfoAdapter is a PDF creation listener that parses meta data
// elements from an (X)HTML document and appends them to the info dictionary
// of a PDF document.
//
// The XHTML document is parsed for relevant PDF meta data during
// construction, then adds the meta data to the PDF document when the PDF
// document is closed by the calling ITextRenderer.
//
// Valid (X)HTML tags are:
//   - TITLE
//
// Valid (X)HTML meta tag attribute names are:
//   - TITLE (optional), DC.TITLE
//   - CREATOR, AUTHOR, DC.CREATOR
//   - SUBJECT, DC.SUBJECT
//   - KEYWORDS
//
// Valid PDF meta names are defined in Adobe's PDF Reference (Sixth Edition),
// section "10.2.1 - Document Information Dictionary", table 10.2, pg.844
// http://www.adobe.com/devnet/pdf/pdf_reference.html
//
// # Usage
//
//	// Create document model
//	var doc *dom.Document = ...
//
//	// Create new PDF renderer
//	renderer := pdf.NewITextRenderer()
//
//	// Add PDF creation listener
//	pdfCreationListener := pdf.NewXHtmlMetaToPdfInfoAdapter(doc)
//	renderer.SetListener(pdfCreationListener)
//
//	// Add document to renderer
//	renderer.SetDocument(doc, "")
//
//	// Layout PDF document
//	renderer.Layout()
//
//	// Write PDF document
//	renderer.CreatePDF(outputStream, true)
//
// # Notes
//
// This class was derived from a sample PDF creation listener at
// "http://markmail.org/message/46t3bw7q6mbhvra2" by Jesse Keller
// <jesse.keller@roche.com>.
//
// Author: Tim Telcik <tim.telcik@permeance.com.au>
//
// See DefaultPDFCreationListener, PDFCreationListener, ITextRenderer,
// https://markmail.org/message/46t3bw7q6mbhvra2,
// https://www.adobe.com/devnet/pdf/pdf_reference.html and
// https://www.seoconsultants.com/meta-tags/dublin/
type XHtmlMetaToPdfInfoAdapter struct {
	DefaultPDFCreationListener

	// pdfInfoValues is a HashMap in Java. pdfInfoKeys keeps the keys in the
	// order they were first put, which is the order they are written to the
	// info dictionary in.
	pdfInfoValues map[writer.PdfName]*writer.PdfString
	pdfInfoKeys   []writer.PdfName
}

// NewXHtmlMetaToPdfInfoAdapter creates a new adapter from the given XHTML
// document.
func NewXHtmlMetaToPdfInfoAdapter(doc *dom.Document) *XHtmlMetaToPdfInfoAdapter {
	a := &XHtmlMetaToPdfInfoAdapter{pdfInfoValues: map[writer.PdfName]*writer.PdfString{}}
	a.parseHtmlTags(doc)
	return a
}

// OnClose is the PDFCreationListener onClose event handler.
func (a *XHtmlMetaToPdfInfoAdapter) OnClose(renderer *ITextRenderer) {
	ufo.XRLogRender(ufo.LevelFinest, "handling onClose event ...")
	a.addPdfMetaValuesToPdfDocument(renderer)
}

func (a *XHtmlMetaToPdfInfoAdapter) put(pdfName writer.PdfName, pdfString *writer.PdfString) {
	if _, ok := a.pdfInfoValues[pdfName]; !ok {
		a.pdfInfoKeys = append(a.pdfInfoKeys, pdfName)
	}
	a.pdfInfoValues[pdfName] = pdfString
}

func (a *XHtmlMetaToPdfInfoAdapter) parseHtmlTags(doc *dom.Document) {
	ufo.XRLogRender(ufo.LevelFinest, "parsing (X)HTML tags ...")
	a.parseHtmlTitleTag(doc)
	a.parseHtmlMetaTags(doc)
	if ufo.XRLogIsLoggingEnabled() {
		var b strings.Builder
		for i, key := range a.pdfInfoKeys {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(string(key) + "=" + a.pdfInfoValues[key].String())
		}
		ufo.XRLogRender(ufo.LevelFinest, "PDF info map = {"+b.String()+"}")
	}
}

// xHtmlMetaToPdfInfoAdapterRootHead returns the first head element of the
// document. Java throws a NullPointerException for a document without one.
func xHtmlMetaToPdfInfoAdapterRootHead(doc *dom.Document) *dom.Element {
	headNodeList := doc.GetDocumentElement().GetElementsByTagName(xHtmlMetaToPdfInfoAdapterHtmlTagHead)
	ufo.XRLogRender(ufo.LevelFinest, "headNodeList="+fmt.Sprint(headNodeList))
	if len(headNodeList) == 0 {
		panic(ufo.NewXRRuntimeException("The document has no head element"))
	}
	return headNodeList[0]
}

func (a *XHtmlMetaToPdfInfoAdapter) parseHtmlTitleTag(doc *dom.Document) {
	rootHeadNodeElement := xHtmlMetaToPdfInfoAdapterRootHead(doc)
	titleNodeList := rootHeadNodeElement.GetElementsByTagName(xHtmlMetaToPdfInfoAdapterHtmlTagTitle)
	ufo.XRLogRender(ufo.LevelFinest, "titleNodeList="+fmt.Sprint(titleNodeList))
	if len(titleNodeList) != 0 {
		titleElement := titleNodeList[0]
		ufo.XRLogRender(ufo.LevelFinest, "titleElement="+fmt.Sprint(titleElement))
		ufo.XRLogRender(ufo.LevelFinest, "titleElement.name="+titleElement.GetTagName())
		ufo.XRLogRender(ufo.LevelFinest, "titleElement.value="+titleElement.GetNodeValue())
		ufo.XRLogRender(ufo.LevelFinest, "titleElement.content="+titleElement.GetTextContent())
		titleContent := titleElement.GetTextContent()
		pdfName := writer.PdfNameTitle
		pdfString := writer.NewPdfString(titleContent)
		a.put(pdfName, pdfString)
	}
}

func (a *XHtmlMetaToPdfInfoAdapter) parseHtmlMetaTags(doc *dom.Document) {
	rootHeadNodeElement := xHtmlMetaToPdfInfoAdapterRootHead(doc)
	metaNodeList := rootHeadNodeElement.GetElementsByTagName(xHtmlMetaToPdfInfoAdapterHtmlTagMeta)
	ufo.XRLogRender(ufo.LevelFinest, "metaNodeList="+fmt.Sprint(metaNodeList))

	for inode := 0; inode < len(metaNodeList); inode++ {
		ufo.XRLogRender(ufo.LevelFinest, "node "+strconv.Itoa(inode)+" = "+metaNodeList[inode].GetNodeName())
		thisNode := metaNodeList[inode]
		ufo.XRLogRender(ufo.LevelFinest, "node "+fmt.Sprint(thisNode))
		metaName := thisNode.GetAttribute(xHtmlMetaToPdfInfoAdapterHtmlMetaAttrName)
		metaContent := thisNode.GetAttribute(xHtmlMetaToPdfInfoAdapterHtmlMetaAttrContent)
		ufo.XRLogRender(ufo.LevelFinest, "metaName="+metaName+", metaContent="+metaContent)
		if metaName != "" && metaContent != "" {

			if strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeyTitle, metaName) ||
				strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeyDcTitle, metaName) {
				pdfName := writer.PdfNameTitle
				pdfString := writer.NewPdfStringWithEncoding(metaContent, writer.PdfObjectTextUnicode)
				a.put(pdfName, pdfString)

			} else if strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeyCreator, metaName) ||
				strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeyDcCreator, metaName) {
				pdfName := writer.PdfNameAuthor
				pdfString := writer.NewPdfStringWithEncoding(metaContent, writer.PdfObjectTextUnicode)
				a.put(pdfName, pdfString)

			} else if strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeySubject, metaName) ||
				strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeyDcSubject, metaName) {
				pdfName := writer.PdfNameSubject
				pdfString := writer.NewPdfStringWithEncoding(metaContent, writer.PdfObjectTextUnicode)
				a.put(pdfName, pdfString)

			} else if strings.EqualFold(xHtmlMetaToPdfInfoAdapterHtmlMetaKeyKeywords, metaName) {
				pdfName := writer.PdfNameKeywords
				pdfString := writer.NewPdfStringWithEncoding(metaContent, writer.PdfObjectTextUnicode)
				a.put(pdfName, pdfString)
			}
		}
	}
}

// addPdfMetaValuesToPdfDocument adds the PDF meta values to the target PDF
// document.
func (a *XHtmlMetaToPdfInfoAdapter) addPdfMetaValuesToPdfDocument(renderer *ITextRenderer) {
	for _, pdfName := range a.pdfInfoKeys {
		pdfString := a.pdfInfoValues[pdfName]
		ufo.XRLogRender(ufo.LevelFinest, "pdfName="+string(pdfName)+", pdfString="+pdfString.String())
		renderer.GetOutputDevice().GetWriter().GetInfo().Put(pdfName, pdfString)
	}
	if ufo.XRLogIsLoggingEnabled() {
		ufo.XRLogRender(ufo.LevelFinest, "added "+fmt.Sprint(renderer.GetOutputDevice().GetWriter().GetInfo().GetKeys()))
	}
}
