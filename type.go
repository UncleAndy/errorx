package errorx

import (
	"encoding"
)

// A distinct error type
// Belongs to a namespace, may be a descendant of another type in the same namespace
// May contain or inherit modifiers that alter the default properties for any error of this type
// May contain or inherit traits that all errors of this type will possess
type Type struct {
	namespace Namespace
	parent    *Type
	id        int64
	fullName  string
	traits    map[Trait]bool
	modifiers modifiers
}

var _ encoding.TextMarshaler = (*Type)(nil)

// Defines a new distinct type within a namespace
func NewType(namespace Namespace, name string, traits ...Trait) *Type {
	return newType(namespace, nil, name, traits...)
}

// Defines a new subtype within a namespace of a parent type
func (t *Type) NewSubtype(name string, traits ...Trait) *Type {
	return newType(t.namespace, t, name, traits...)
}

// One-time modification of defaults in error creation
func (t *Type) ApplyModifiers(modifiers ...TypeModifier) *Type {
	t.modifiers = t.modifiers.ReplaceWith(newTypeModifiers(modifiers...))
	return t
}

// Create an error of this type with a message
// Without args, leaves the original message intact, so a message may be generated or provided externally
// With args, a formatting is performed, and it is therefore expected a format string to be constant
func (t *Type) New(message string, args ...interface{}) *Error {
	return NewErrorBuilder(t).
		WithConditionallyFormattedMessage(message, args...).
		Create()
}

// Create an error of this type without any message
// May be used when other information is sufficient, such as error type and stack trace
func (t *Type) NewWithNoMessage() *Error {
	return NewErrorBuilder(t).
		Create()
}

// Create an error of this type with another as original cause
// As far as type checks are concerned, this error is the only one visible, with original present only in error message
// The original error will, however, pass its dynamic properties
// Without args, leaves the original message intact, so a message may be generated or provided externally
// With args, a formatting is performed, and it is therefore expected a format string to be constant
func (t *Type) Wrap(err error, message string, args ...interface{}) *Error {
	return NewErrorBuilder(t).
		WithConditionallyFormattedMessage(message, args...).
		WithCause(err).
		Create()
}

// Create an error of this type with another as original cause and with no additional message
// May be used when other information is sufficient, such as error type, cause and its stack trace and message
// As far as type checks are concerned, this error is the only one visible, with original visible only in error message
// The original error will, however, pass its dynamic properties
func (t *Type) WrapWithNoMessage(err error) *Error {
	return NewErrorBuilder(t).
		WithCause(err).
		Create()
}

// Type check for errors
// Returns true either both are exactly the same type, or if the same is true for one of current type's ancestors
func (t *Type) IsOfType(other *Type) bool {
	current := t
	for current != nil {
		if current.id == other.id {
			return true
		}

		current = current.parent
	}

	return false
}

// Type check for errors
// Returns true either both are exactly the same type, or if the same is true for one of current type's ancestors
// For an error that does not have an errorx type, returns false
func IsOfType(err error, t *Type) bool {
	e := Cast(err)
	return e != nil && e.IsOfType(t)
}

// A parent type, if present
func (t *Type) Supertype() *Type {
	return t.parent
}

// A fully qualified name if type, is not presumed to be unique, see TypeSubscriber
func (t *Type) FullName() string {
	return t.fullName
}

// A base namespace this type belongs to
func (t *Type) RootNamespace() Namespace {
	return t.namespace
}

func (t *Type) String() string {
	return t.FullName()
}

func (t *Type) MarshalText() (text []byte, err error) {
	return []byte(t.String()), nil
}

func newType(namespace Namespace, parent *Type, name string, traits ...Trait) *Type {
	collectModifiers := func() modifiers {
		if parent == nil {
			return newInheritedModifiers(namespace.modifiers)
		} else {
			return newInheritedModifiers(parent.modifiers)
		}
	}

	collectTraits := func() map[Trait]bool {
		result := make(map[Trait]bool)
		if parent != nil {
			for trait := range parent.traits {
				result[trait] = true
			}
		}

		for trait := range namespace.collectTraits() {
			result[trait] = true
		}

		for _, trait := range traits {
			result[trait] = true
		}

		return result
	}

	createFullName := func() string {
		if parent == nil {
			return namespace.FullName() + "." + name
		} else {
			return parent.FullName() + "." + name
		}
	}

	t := &Type{
		id:        nextInternalID(),
		namespace: namespace,
		parent:    parent,
		fullName:  createFullName(),
		traits:    collectTraits(),
		modifiers: collectModifiers(),
	}

	globalRegistry.registerType(t)
	return t
}