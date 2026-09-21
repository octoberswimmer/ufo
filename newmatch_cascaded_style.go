// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/CascadedStyle.java

package ufo

import (
	"sort"
	"strings"
)

// CascadedStyle holds a set of PropertyDeclarations for each unique CSS
// property name. What properties belong in the set is not determined, except
// that multiple entries are resolved into a single set using cascading rules.
// The set is cascaded during instantiation, so once you have a CascadedStyle,
// the PropertyDeclarations you retrieve from it will have been resolved
// following the CSS cascading rules. Note that this class knows nothing about
// CSS selector-matching rules. Before creating a CascadedStyle, you will need
// to determine which PropertyDeclarations belong in the set--for example, by
// matching Rulesets to document elements via their selectors. You can get
// individual properties by using PropertyByName or a slice of properties with
// GetCascadedPropertyDeclarations. Check for individual property assignments
// using HasProperty. A CascadedStyle is immutable, as properties can not be
// added or removed from it once instantiated.
type CascadedStyle struct {
	// cascadedProperties is a TreeMap in Java, ordered by CSSName.compareTo
	// (the FS_ID). The Go map is unordered; the two methods that observe the
	// order, GetCascadedPropertyDeclarations and GetFingerprint, read it
	// through sortedValues, which sorts by CSSName.CompareTo.
	cascadedProperties map[*CSSName]*PropertyDeclaration

	fingerprint    string
	hasFingerprint bool
}

// CascadedStyleCreateAnonymousStyle creates a CascadedStyle, setting the
// display property to the value of the display parameter.
func CascadedStyleCreateAnonymousStyle(display *IdentValue) *CascadedStyle {
	val := NewPropertyValueIdentValue(display)

	props := []*PropertyDeclaration{
		NewPropertyDeclaration(CSSNameDisplay, val, true, StylesheetInfoOriginUser)}

	return NewCascadedStyle(props)
}

// CascadedStyleCreateLayoutStyle creates a CascadedStyle using the provided
// property declarations. It is used when a box requires a style that does not
// correspond to anything in the parsed stylesheets. The declarations are
// PropertyDeclaration objects created with
// CascadedStyleCreateLayoutPropertyDeclaration.
//
// It stands for both Java overloads createLayoutStyle(PropertyDeclaration...)
// and createLayoutStyle(List<PropertyDeclaration>).
func CascadedStyleCreateLayoutStyle(declarations ...*PropertyDeclaration) *CascadedStyle {
	return NewCascadedStyle(declarations)
}

// CascadedStyleCreateLayoutStyleWithStartingPoint creates a CascadedStyle
// using style information from startingPoint and then adding the property
// declarations from decls. The declarations are PropertyDeclaration objects
// created with CascadedStyleCreateLayoutPropertyDeclaration.
func CascadedStyleCreateLayoutStyleWithStartingPoint(startingPoint *CascadedStyle, decls []*PropertyDeclaration) *CascadedStyle {
	return newCascadedStyleWithStartingPoint(startingPoint.cascadedProperties, decls)
}

// CascadedStyleCreateLayoutPropertyDeclaration creates a PropertyDeclaration
// suitable for passing to CascadedStyleCreateLayoutStyle or
// CascadedStyleCreateLayoutStyleWithStartingPoint.
func CascadedStyleCreateLayoutPropertyDeclaration(cssName *CSSName, display *IdentValue) *PropertyDeclaration {
	val := NewPropertyValueIdentValue(display)
	// Urk... kind of ugly, but we really want this value to be used
	return NewPropertyDeclaration(cssName, val, true, StylesheetInfoOriginUser)
}

// NewCascadedStyle constructs a new CascadedStyle, given a slice of
// PropertyDeclarations already sorted by specificity of the CSS selector they
// came from. The slice can have multiple PropertyDeclarations with the same
// name; the property cascade will be resolved during instantiation, resulting
// in a set of PropertyDeclarations. Once instantiated, properties may be
// retrieved using the normal API for the class.
func NewCascadedStyle(iter []*PropertyDeclaration) *CascadedStyle {
	return newCascadedStyleWithStartingPoint(nil, iter)
}

func newCascadedStyleWithStartingPoint(startingPoint map[*CSSName]*PropertyDeclaration, iter []*PropertyDeclaration) *CascadedStyle {
	//do a bucket-sort on importance and origin
	//properties should already be in order of specificity
	buckets := make([][]*PropertyDeclaration, PropertyDeclarationImportanceAndOriginCount)

	for _, prop := range iter {
		i := prop.GetImportanceAndOrigin()
		buckets[i] = append(buckets[i], prop)
	}

	cascadedProperties := make(map[*CSSName]*PropertyDeclaration, len(startingPoint))
	for cssName, prop := range startingPoint {
		cascadedProperties[cssName] = prop
	}
	for _, bucket := range buckets {
		for _, prop := range bucket {
			cascadedProperties[prop.GetCSSName()] = prop
		}
	}
	return &CascadedStyle{cascadedProperties: cascadedProperties}
}

// CascadedStyleEmptyCascadedStyle is an empty singleton, used to negate
// inheritance of properties
var CascadedStyleEmptyCascadedStyle = &CascadedStyle{cascadedProperties: map[*CSSName]*PropertyDeclaration{}}

// HasProperty returns true if property has been defined in this style.
// cssName is the CSS property name, e.g. "font-family".
func (s *CascadedStyle) HasProperty(cssName *CSSName) bool {
	_, ok := s.cascadedProperties[cssName]
	return ok
}

// PropertyByName returns a PropertyDeclaration by CSS property name, e.g.
// "font-family". Properties are already cascaded during instantiation, so this
// will return the actual property (and corresponding value) to use for
// CSS-based layout and rendering. It returns nil if the property is not
// declared in this set.
func (s *CascadedStyle) PropertyByName(cssName *CSSName) *PropertyDeclaration {
	return s.cascadedProperties[cssName]
}

// GetIdent gets the ident attribute of the CascadedStyle object
func (s *CascadedStyle) GetIdent(cssName *CSSName) *IdentValue {
	pd := s.PropertyByName(cssName)
	if pd == nil {
		return nil
	}
	return pd.AsIdentValue()
}

// sortedValues is TreeMap.values(): the declarations in the order of their
// CSSName.
func (s *CascadedStyle) sortedValues() []*PropertyDeclaration {
	list := make([]*PropertyDeclaration, 0, len(s.cascadedProperties))
	for _, prop := range s.cascadedProperties {
		list = append(list, prop)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].GetCSSName().CompareTo(list[j].GetCSSName()) < 0
	})
	return list
}

// GetCascadedPropertyDeclarations returns the PropertyDeclarations already
// matched in this CascadedStyle. For a given property name, there may be no
// match, in which case there will be no PropertyDeclaration for that property
// name in the result.
func (s *CascadedStyle) GetCascadedPropertyDeclarations() []*PropertyDeclaration {
	return s.sortedValues()
}

func (s *CascadedStyle) CountAssigned() int { return len(s.cascadedProperties) }

func (s *CascadedStyle) GetFingerprint() string {
	if !s.hasFingerprint {
		var sb strings.Builder
		for _, propertyDeclaration := range s.sortedValues() {
			sb.WriteString(propertyDeclaration.GetFingerprint())
		}
		s.fingerprint = sb.String()
		s.hasFingerprint = true
	}
	return s.fingerprint
}
