package errorx

func NotRecognisedType() *Type { return notRecognisedType }

func CaseNoError() Trait { return caseNoError }
func CaseNoTrait() Trait { return caseNoTrait }

var (
	notRecognisedType = syntheticErrors.NewType("non.recognised")

	caseNoError = RegisterTrait("synthetic.no.error")
	caseNoTrait = RegisterTrait("synthetic.no.trait")
)

// Used to perform a switch around the type of an error
// For nil errors, returns nil
// For error types not in the 'types' list, including non-errorx errors, NotRecognisedType() is returned
// It is safe to treat NotRecognisedType() as 'any other type of not-nil error' case
// The effect is equivalent to a series of IsOfType() checks
// NB: if more than one provided types matches with error, the first match in the providers list is recognised
func TypeSwitch(err error, types ...*Type) *Type {
	typed := Cast(err)

	switch {
	case err == nil:
		return nil
	case typed == nil:
		return NotRecognisedType()
	default:
		for _, t := range types {
			if typed.IsOfType(t) {
				return t
			}
		}

		return NotRecognisedType()
	}
}

// Used to perform a switch around the trait of an error
// For nil errors, returns CaseNoError()
// For error types that lack any of the provided traits, including non-errorx errors, CaseNoTrait() is returned
// It is safe to treat CaseNoTrait() as 'any other kind of not-nil error' case
// The effect is equivalent to a series of HasTrait() checks
// NB: if more than one provided types matches with error, the first match in the providers list is recognised
func TraitSwitch(err error, traits ...Trait) Trait {
	typed := Cast(err)

	switch {
	case err == nil:
		return CaseNoError()
	case typed == nil:
		return CaseNoTrait()
	default:
		for _, t := range traits {
			if typed.HasTrait(t) {
				return t
			}
		}

		return CaseNoTrait()
	}
}
