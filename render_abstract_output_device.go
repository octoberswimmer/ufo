// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/AbstractOutputDevice.java

package ufo

import (
	"fmt"
	"math"

	"github.com/octoberswimmer/ufo/geom"
)

// AbstractOutputDeviceI lists the non-private methods of AbstractOutputDevice:
// the OutputDevice interface, the protected methods, and the font
// specification accessors. The subclass passes itself as this interface to
// InitAbstractOutputDevice; AbstractOutputDevice calls the methods that are
// abstract in Java (DrawLine, SetColor, Fill, Clip, DrawImage, ...) through it.
type AbstractOutputDeviceI interface {
	OutputDevice

	AsAbstractOutputDevice() *AbstractOutputDevice

	// DrawLine is the protected abstract method drawLine. It is exported so
	// that a type in another package (pdf) can implement it.
	DrawLine(x1 int, y1 int, x2 int, y2 int)

	JustificationInfo(c *RenderingContext, inlineText *InlineText) *JustificationInfo

	GetFontSpecification() *FontSpecification
	SetFontSpecification(fs *FontSpecification)
}

// AbstractOutputDevice is an abstract implementation of an OutputDevice. It
// provides complete implementations for many OutputDevice methods.
//
// It is a struct to embed. The Java type parameters
// <T extends FSImage, FontType extends FSFont> are the interfaces FSImage and
// FSFont here.
type AbstractOutputDevice struct {
	self AbstractOutputDeviceI

	// fontSpec may be nil.
	fontSpec *FontSpecification
}

// InitAbstractOutputDevice sets the outermost object, through which
// AbstractOutputDevice calls the methods its subclass implements. Every
// constructor of a subclass calls it.
func InitAbstractOutputDevice(a *AbstractOutputDevice, self AbstractOutputDeviceI) {
	a.self = self
}

func (a *AbstractOutputDevice) AsAbstractOutputDevice() *AbstractOutputDevice {
	return a
}

func (a *AbstractOutputDevice) DrawText(c *RenderingContext, inlineText *InlineText) {
	iB := inlineText.GetParent()
	text := inlineText.GetSubstring()

	if text != "" {
		a.self.SetColor(iB.GetStyle().GetColor())
		a.self.SetFont(iB.GetStyle().GetFSFont(c))
		a.self.SetFontSpecification(iB.GetStyle().GetFontSpecification())
		info := a.self.JustificationInfo(c, inlineText)
		if info != nil {
			c.GetTextRenderer().DrawStringWithInfo(
				c.GetOutputDevice(),
				text,
				float32(iB.GetAbsX()+inlineText.GetX()), float32(iB.GetAbsY()+iB.GetBaseline()),
				info)
		} else {
			c.GetTextRenderer().DrawString(
				c.GetOutputDevice(),
				text,
				float32(iB.GetAbsX()+inlineText.GetX()), float32(iB.GetAbsY()+iB.GetBaseline()))
		}
	}

	if c.DebugDrawFontMetrics() {
		a.drawFontMetrics(c, inlineText)
	}
}

// JustificationInfo returns the per-character adjustments to draw inlineText
// with: the line's justification (if any) plus the style's letter-spacing (if
// any), or nil when neither applies.
func (a *AbstractOutputDevice) JustificationInfo(c *RenderingContext, inlineText *InlineText) *JustificationInfo {
	iB := inlineText.GetParent()
	var info *JustificationInfo
	if iB.GetStyle().IsTextJustify() {
		info = iB.GetLineBox().GetJustificationInfo()
	}

	letterSpacing := iB.GetStyle().LetterSpacing(c)
	if letterSpacing == 0.0 {
		return info
	}
	if info == nil {
		return NewJustificationInfo(letterSpacing, letterSpacing)
	}
	return NewJustificationInfo(info.NonSpaceAdjust()+letterSpacing, info.SpaceAdjust()+letterSpacing)
}

func (a *AbstractOutputDevice) drawFontMetrics(c *RenderingContext, inlineText *InlineText) {
	iB := inlineText.GetParent()
	text := inlineText.GetSubstring()

	a.self.SetColor(NewFSRGBColor(0xFF, 0x33, 0xFF))

	fm := iB.GetStyle().GetFSFontMetrics(nil)
	width := LayoutTextUtilTextWidth(c, iB.GetStyle(), iB.GetStyle().GetFSFont(c), text)
	x := iB.GetAbsX() + inlineText.GetX()
	y := iB.GetAbsY() + iB.GetBaseline()

	a.self.DrawLine(x, y, x+width, y)

	y += int(math.Ceil(float64(fm.GetDescent())))
	a.self.DrawLine(x, y, x+width, y)

	y -= int(math.Ceil(float64(fm.GetDescent())))
	y -= int(math.Ceil(float64(fm.GetAscent())))
	a.self.DrawLine(x, y, x+width, y)
}

func (a *AbstractOutputDevice) DrawTextDecorationWithIBDecoration(
	c *RenderingContext, iB *InlineLayoutBox, decoration *TextDecoration) {
	a.self.SetColor(iB.GetStyle().GetColor())

	edge := iB.GetContentAreaEdge(iB.GetAbsX(), iB.GetAbsY(), c)

	a.self.FillRect(edge.X, iB.GetAbsY()+decoration.GetOffset(),
		edge.Width, decoration.GetThickness())
}

func (a *AbstractOutputDevice) DrawTextDecoration(c *RenderingContext, lineBox *LineBox) {
	a.self.SetColor(lineBox.GetStyle().GetColor())
	parent := lineBox.GetParent()
	decorations := lineBox.GetTextDecorations()
	for _, textDecoration := range decorations {
		if parent.GetStyle().IsIdent(
			CSSNameFsTextDecorationExtent, IdentValueBlock) {
			a.self.FillRect(
				lineBox.GetAbsX(),
				lineBox.GetAbsY()+textDecoration.GetOffset(),
				parent.GetAbsX()+parent.GetTx()+parent.GetContentWidth()-lineBox.GetAbsX(),
				textDecoration.GetThickness())
		} else {
			a.self.FillRect(
				lineBox.GetAbsX(), lineBox.GetAbsY()+textDecoration.GetOffset(),
				lineBox.GetContentWidth(),
				textDecoration.GetThickness())
		}
	}
}

func (a *AbstractOutputDevice) DrawDebugOutline(c *RenderingContext, box BoxI, color FSColor) {
	a.self.SetColor(color)
	rect := box.GetMarginEdgeWithLeftTop(box.GetAbsX(), box.GetAbsY(), c, 0, 0)
	rect.Height -= 1
	rect.Width -= 1
	a.self.DrawRect(rect.X, rect.Y, rect.Width, rect.Height)
}

func (a *AbstractOutputDevice) PaintCollapsedBorder(
	c *RenderingContext, border *BorderPropertySet, bounds *geom.Rectangle, side int) {
	BorderPainterPaint(bounds, side, border, c, 0)
}

func (a *AbstractOutputDevice) PaintBorder(c *RenderingContext, box BoxI) {
	if !box.GetStyle().IsVisible() {
		return
	}

	borderBounds := box.GetPaintingBorderEdge(c)

	BorderPainterPaint(borderBounds, box.GetBorderSides(), box.GetBorder(c), c, 0)
}

func (a *AbstractOutputDevice) PaintBorderWithStyleEdgeSides(c *RenderingContext, style CalculatedStyleI, edge *geom.Rectangle, sides int) {
	BorderPainterPaint(edge, sides, style.GetBorder(c), c, 0)
}

// getBackgroundImage returns nil when the style has no background image or the
// image cannot be loaded. Java catches every Exception raised while loading
// and logs it; a panic raised here is recovered and logged in the same way.
func (a *AbstractOutputDevice) getBackgroundImage(c *RenderingContext, style CalculatedStyleI) (result FSImage) {
	if !style.IsIdent(CSSNameBackgroundImage, IdentValueNone) {
		uri := style.GetStringProperty(CSSNameBackgroundImage)
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				UuPException(err)
				result = nil
			}
		}()
		return c.GetUac().GetImageResource(uri).GetImage()
	}
	return nil
}

func (a *AbstractOutputDevice) PaintBackgroundWithStyleBoundsBgImageContainerBorder(
	c *RenderingContext, style CalculatedStyleI,
	bounds *geom.Rectangle, bgImageContainer *geom.Rectangle, border *BorderPropertySet) {
	a.paintBackground0(c, style, bounds, bgImageContainer, border)
}

func (a *AbstractOutputDevice) PaintBackground(c *RenderingContext, box BoxI) {
	if !box.GetStyle().IsVisible() {
		return
	}

	backgroundBounds := box.GetPaintingBorderEdge(c)
	border := box.GetStyle().GetBorder(c)
	a.paintBackground0(c, box.GetStyle(), backgroundBounds, backgroundBounds, border)
}

func (a *AbstractOutputDevice) paintBackground0(
	c *RenderingContext, style CalculatedStyleI,
	backgroundBounds *geom.Rectangle, bgImageContainer *geom.Rectangle,
	border *BorderPropertySet) {
	if !ConfigurationIsTrue("xr.renderer.draw.backgrounds", true) {
		return
	}

	a.self.SetOpacity(style.GetOpacity())

	backgroundColor := style.GetBackgroundColor()

	var backgroundLinearGradient *FSLinearGradient
	var backgroundImage FSImage

	if style.IsLinearGradient() {
		// TODO: Is this the correct width to use?
		backgroundLinearGradient = style.GetLinearGradient(c, bgImageContainer.Width, bgImageContainer.Height)
	} else {
		backgroundImage = a.getBackgroundImage(c, style)
	}

	// If the image width or height is zero, then there's nothing to draw.
	// Also prevents infinite loop when trying to tile an image with zero size.
	if backgroundImage == nil || backgroundImage.GetHeight() == 0 || backgroundImage.GetWidth() == 0 {
		backgroundImage = nil
	}

	if (backgroundColor == nil || backgroundColor == FSColor(FSRGBColorTransparent)) &&
		backgroundImage == nil && backgroundLinearGradient == nil {
		return
	}

	borderBounds := geom.NewArea(BorderPainterGenerateBorderBounds(backgroundBounds, border, false))

	oldclip := a.self.GetClip()
	if oldclip != nil {
		// we need to respect the clip sent to us, get the intersection between the old and the new
		borderBounds.Intersect(geom.NewArea(oldclip))
	}

	if backgroundColor != nil && backgroundColor != FSColor(FSRGBColorTransparent) {
		a.self.SetColor(backgroundColor)
		a.self.Fill(borderBounds)
	}

	if backgroundImage != nil || backgroundLinearGradient != nil {
		a.self.SetClip(borderBounds)
		localBGImageContainer := bgImageContainer
		if style.IsFixedBackground() {
			localBGImageContainer = c.GetViewportRectangle()
		}

		xoff := localBGImageContainer.X
		yoff := localBGImageContainer.Y

		if border != nil {
			xoff += calculatedStyleFloatToInt(border.Left())
			yoff += calculatedStyleFloatToInt(border.Top())
		}

		a.self.Clip(borderBounds)

		if backgroundLinearGradient != nil {
			a.self.DrawLinearGradient(backgroundLinearGradient,
				backgroundBounds.X, backgroundBounds.Y, backgroundBounds.Width, backgroundBounds.Height)
			a.self.SetClip(oldclip)
			return
		}

		if backgroundImage != nil {
			backgroundImage = a.scaleBackgroundImage(c, style, localBGImageContainer, backgroundImage)
		}

		imageWidth := float32(backgroundImage.GetWidth())
		imageHeight := float32(backgroundImage.GetHeight())

		position := style.GetBackgroundPosition()
		xoff += a.calcOffset(
			c, style, position.GetHorizontal(), float32(localBGImageContainer.Width), imageWidth)
		yoff += a.calcOffset(
			c, style, position.GetVertical(), float32(localBGImageContainer.Height), imageHeight)

		hrepeat := style.IsHorizontalBackgroundRepeat()
		vrepeat := style.IsVerticalBackgroundRepeat()

		if !hrepeat && !vrepeat {
			imageBounds := geom.NewRectangle(xoff, yoff, calculatedStyleFloatToInt(imageWidth), calculatedStyleFloatToInt(imageHeight))
			if imageBounds.Intersects(backgroundBounds) {
				a.self.DrawImage(backgroundImage, xoff, yoff)
			}
		} else if hrepeat && vrepeat {
			a.paintTiles(
				backgroundImage,
				a.adjustTo(backgroundBounds.X, xoff, calculatedStyleFloatToInt(imageWidth)),
				a.adjustTo(backgroundBounds.Y, yoff, calculatedStyleFloatToInt(imageHeight)),
				backgroundBounds.X+backgroundBounds.Width,
				backgroundBounds.Y+backgroundBounds.Height)
		} else if hrepeat {
			xoff = a.adjustTo(backgroundBounds.X, xoff, calculatedStyleFloatToInt(imageWidth))
			imageBounds := geom.NewRectangle(xoff, yoff, calculatedStyleFloatToInt(imageWidth), calculatedStyleFloatToInt(imageHeight))
			if imageBounds.Intersects(backgroundBounds) {
				a.paintHorizontalBand(
					backgroundImage,
					xoff,
					yoff,
					backgroundBounds.X+backgroundBounds.Width)
			}
		} else if vrepeat {
			yoff = a.adjustTo(backgroundBounds.Y, yoff, calculatedStyleFloatToInt(imageHeight))
			imageBounds := geom.NewRectangle(xoff, yoff, calculatedStyleFloatToInt(imageWidth), calculatedStyleFloatToInt(imageHeight))
			if imageBounds.Intersects(backgroundBounds) {
				a.paintVerticalBand(
					backgroundImage,
					xoff,
					yoff,
					backgroundBounds.Y+backgroundBounds.Height)
			}
		}

		a.self.SetClip(oldclip)
	}
}

func (a *AbstractOutputDevice) adjustTo(target int, current int, imageDim int) int {
	result := current
	if result > target {
		for result > target {
			result -= imageDim
		}
	} else if result < target {
		for result < target {
			result += imageDim
		}
		if result != target {
			result -= imageDim
		}
	}
	return result
}

func (a *AbstractOutputDevice) paintTiles(image FSImage, left int, top int, right int, bottom int) {
	width := image.GetWidth()
	height := image.GetHeight()

	for x := left; x < right; x += width {
		for y := top; y < bottom; y += height {
			a.self.DrawImage(image, x, y)
		}
	}
}

func (a *AbstractOutputDevice) paintVerticalBand(image FSImage, left int, top int, bottom int) {
	height := image.GetHeight()

	for y := top; y < bottom; y += height {
		a.self.DrawImage(image, left, y)
	}
}

func (a *AbstractOutputDevice) paintHorizontalBand(image FSImage, left int, top int, right int) {
	width := image.GetWidth()

	for x := left; x < right; x += width {
		a.self.DrawImage(image, x, top)
	}
}

func (a *AbstractOutputDevice) calcOffset(c CssContext, style CalculatedStyleI, value *PropertyValue, boundsDim float32, imageDim float32) int {
	if value.GetPrimitiveType() == CSSPrimitiveValueCssPercentage {
		percent := value.GetFloatValue() / 100.0
		return propertyBuilderRound(float32(boundsDim*percent) - float32(imageDim*percent))
	} else { /* it's a <length> */
		return calculatedStyleFloatToInt(LengthValueCalcFloatProportionalValue(
			style,
			CSSNameBackgroundPosition,
			value.GetCssText(),
			value.GetFloatValue(),
			value.GetPrimitiveType(),
			0,
			c))
	}
}

func (a *AbstractOutputDevice) scaleBackgroundImage(c CssContext, style CalculatedStyleI, backgroundContainer *geom.Rectangle, image FSImage) FSImage {
	backgroundSize := style.GetBackgroundSize()

	if !backgroundSize.IsBothAuto() {
		if backgroundSize.IsCover() || backgroundSize.IsContain() {
			testHeight := int(float64(image.GetHeight()) * float64(backgroundContainer.Width) / float64(image.GetWidth()))
			if backgroundSize.IsContain() {
				if testHeight > backgroundContainer.Height {
					return image.Scale(-1, backgroundContainer.Height)
				} else {
					return image.Scale(backgroundContainer.Width, -1)
				}
			} else if backgroundSize.IsCover() {
				if testHeight > backgroundContainer.Height {
					return image.Scale(backgroundContainer.Width, -1)
				} else {
					return image.Scale(-1, backgroundContainer.Height)
				}
			}
		} else {
			scaledWidth := a.calcBackgroundSizeLength(c, style, backgroundSize.GetWidth(), float32(backgroundContainer.Width))
			scaledHeight := a.calcBackgroundSizeLength(c, style, backgroundSize.GetHeight(), float32(backgroundContainer.Height))

			return image.Scale(scaledWidth, scaledHeight)
		}
	}
	return image
}

func (a *AbstractOutputDevice) calcBackgroundSizeLength(c CssContext, style CalculatedStyleI, value *PropertyValue, boundsDim float32) int {
	if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent { // 'auto'
		return -1
	} else if value.GetPrimitiveType() == CSSPrimitiveValueCssPercentage {
		percent := value.GetFloatValue() / 100.0
		return propertyBuilderRound(boundsDim * percent)
	} else {
		return calculatedStyleFloatToInt(LengthValueCalcFloatProportionalValue(
			style,
			CSSNameBackgroundSize,
			value.GetCssText(),
			value.GetFloatValue(),
			value.GetPrimitiveType(),
			0,
			c))
	}
}

// GetFontSpecification gets the FontSpecification for this
// AbstractOutputDevice.
func (a *AbstractOutputDevice) GetFontSpecification() *FontSpecification {
	return a.fontSpec
}

// SetFontSpecification sets the FontSpecification for this
// AbstractOutputDevice.
func (a *AbstractOutputDevice) SetFontSpecification(fs *FontSpecification) {
	a.fontSpec = fs
}
