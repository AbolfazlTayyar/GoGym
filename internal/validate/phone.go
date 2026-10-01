// Package validate holds input formats shared across modules.
package validate

import "regexp"

// International forms (+989…) are rejected, not normalized, so one number can't register twice past UNIQUE.
var iranMobilePattern = regexp.MustCompile(`^09\d{9}$`)

const MsgIranMobile = "must be an Iranian mobile number, e.g. 09372144430"

func IsIranMobile(s string) bool {
	return iranMobilePattern.MatchString(s)
}
