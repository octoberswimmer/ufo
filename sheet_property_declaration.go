// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/PropertyDeclaration.java

package ufo

import "strconv"

// PropertyDeclaration represents a single property declared in a CSS rule
// set. A PropertyDeclaration is created from a CSS value and is immutable. The
// declaration knows its origin, importance and specificity, and thus is
// prepared to be sorted out among properties of the same name, within a
// matched group, for the CSS cascade, into a CascadedStyle.
//
// Java types the value as org.w3c.dom.css.CSSPrimitiveValue. PropertyValue is
// the only implementation and GetFingerprint casts to it, so the Go field is
// a *PropertyValue.
type PropertyDeclaration struct {
	propName          string
	cssName           *CSSName
	cssPrimitiveValue *PropertyValue

	// Whether the property was declared as important! by the user.
	important bool

	origin      StylesheetInfoOrigin
	identVal    *IdentValue
	identIsSet  bool
	fingerprint string
}

// PropertyDeclarationImportanceAndOriginCount is ImportanceAndOrigin of
// stylesheet - how many different
const PropertyDeclarationImportanceAndOriginCount = 6

const (
	// ImportanceAndOrigin of stylesheet - user agent
	propertyDeclarationUserAgent = 1

	// ImportanceAndOrigin of stylesheet - user normal
	propertyDeclarationUserNormal = 2

	// ImportanceAndOrigin of stylesheet - author normal
	propertyDeclarationAuthorNormal = 3

	// ImportanceAndOrigin of stylesheet - author important
	propertyDeclarationAuthorImportant = 4

	// ImportanceAndOrigin of stylesheet - user important
	propertyDeclarationUserImportant = 5
)

// NewPropertyDeclaration creates a new instance of PropertyDeclaration.
//
// cssName is the name of the CSS property, value the value to wrap, imp true
// if the property was declared important! and false if not, and orig the
// origin of the property declaration, that is, the origin of the style sheet
// where it was declared.
func NewPropertyDeclaration(cssName *CSSName, value *PropertyValue, imp bool, orig StylesheetInfoOrigin) *PropertyDeclaration {
	return &PropertyDeclaration{
		propName:          cssName.ToString(),
		cssName:           cssName,
		cssPrimitiveValue: value,
		important:         imp,
		origin:            orig,
	}
}

// String converts to a String representation of the object.
func (p *PropertyDeclaration) String() string {
	return p.GetPropertyName() + ": " + p.GetValue().ToString()
}

func (p *PropertyDeclaration) ToString() string {
	return p.String()
}

func (p *PropertyDeclaration) AsIdentValue() *IdentValue {
	if !p.identIsSet {
		p.identVal = IdentValueGetByIdentString(p.cssPrimitiveValue.GetCssText())
		p.identIsSet = true
	}
	return p.identVal
}

func (p *PropertyDeclaration) GetDeclarationStandardText() string {
	return p.cssName.ToString() + ": " + p.cssPrimitiveValue.GetCssText() + ";"
}

func (p *PropertyDeclaration) GetFingerprint() string {
	if p.fingerprint == "" {
		// The Java expression is 'P' + cssName.FS_ID + ':' + fingerprint + ';'.
		// Its first two operators add a char, an int and a char, which is
		// integer addition, so the text starts with the decimal number
		// 80 + FS_ID + 58 and contains neither "P" nor ":".
		p.fingerprint = strconv.Itoa('P'+p.cssName.FS_ID+':') + p.cssPrimitiveValue.GetFingerprint() + ";"
	}
	return p.fingerprint
}

// GetImportanceAndOrigin returns an int representing the combined origin and
// importance of the property as declared. The int is assigned such that
// default origin and importance is 0, and highest an important! property
// defined by the user (origin is StylesheetInfoOriginUser). The combined value
// would allow this property to be sequenced in the CSS cascade along with
// other properties matched to the same element with the same property name.
// In that sort, the highest sequence number returned from this method would
// take priority in the cascade, so that a user important! property would
// override a user non-important! property, and so on. The actual integer
// value returned by this method is unimportant, but has the lowest value of 0
// and increments sequentially by 1 for each increase in origin/importance.
func (p *PropertyDeclaration) GetImportanceAndOrigin() int {
	switch p.origin {
	case StylesheetInfoOriginUserAgent:
		return propertyDeclarationUserAgent
	case StylesheetInfoOriginUser:
		if p.important {
			return propertyDeclarationUserImportant
		}
		return propertyDeclarationUserNormal
	case StylesheetInfoOriginAuthor:
		if p.important {
			return propertyDeclarationAuthorImportant
		}
		return propertyDeclarationAuthorNormal
	}
	panic(NewXRRuntimeException("unknown StylesheetInfo.Origin"))
}

// GetPropertyName returns the CSS name of this property, e.g. "font-family".
func (p *PropertyDeclaration) GetPropertyName() string {
	return p.propName
}

// GetCSSName gets the cSSName attribute of the PropertyDeclaration object.
func (p *PropertyDeclaration) GetCSSName() *CSSName {
	return p.cssName
}

// GetValue returns the specified value for this property. Specified means the
// value as entered by the user. Modifying the value returned here will result
// in indeterminate behavior--consider it immutable.
func (p *PropertyDeclaration) GetValue() *PropertyValue {
	return p.cssPrimitiveValue
}

func (p *PropertyDeclaration) IsImportant() bool {
	return p.important
}

func (p *PropertyDeclaration) GetOrigin() StylesheetInfoOrigin {
	return p.origin
}
