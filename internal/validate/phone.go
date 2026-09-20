// Package validate holds input formats shared across modules, so a rule like
// "what counts as a phone number" is defined once rather than re-derived in
// each handler that accepts one.
package validate

import "regexp"

// iranMobilePattern matches an Iranian mobile number in national format: a
// leading 0, the mobile marker 9, then 9 more digits — e.g. 09372144430.
//
// International forms (+989…, 00989…) are deliberately rejected rather than
// normalized: the app stores one canonical form, and accepting several that
// map to the same subscriber would let the same person register twice past
// the phone UNIQUE constraint.
var iranMobilePattern = regexp.MustCompile(`^09\d{9}$`)

// IsIranMobile reports whether s is an Iranian mobile number in national
// format (09XXXXXXXXX).
func IsIranMobile(s string) bool {
	return iranMobilePattern.MatchString(s)
}
