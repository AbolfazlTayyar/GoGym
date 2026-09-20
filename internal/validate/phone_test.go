package validate_test

import (
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/validate"
	"github.com/stretchr/testify/assert"
)

func TestIsIranMobile(t *testing.T) {
	valid := []string{
		"09372144430",
		"09163503284",
		"09121234567",
		"09901234567",
		"09011234567",
	}
	for _, phone := range valid {
		assert.True(t, validate.IsIranMobile(phone), "expected %q to be valid", phone)
	}

	invalid := []string{
		"",
		"0937214443",     // one digit short
		"093721444301",   // one digit long
		"08372144430",    // doesn't start with 09
		"+989372144430",  // international form is not normalized, so not accepted
		"00989372144430", // ditto
		"9372144430",     // missing leading 0
		"0937 214 4430",  // spaces
		"0937-214-4430",  // separators
		"0937214443a",    // non-digit
	}
	for _, phone := range invalid {
		assert.False(t, validate.IsIranMobile(phone), "expected %q to be invalid", phone)
	}
}
