package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetListEnv(t *testing.T) {
	const key = "GOGYM_TEST_LIST"

	tests := map[string]struct {
		value string
		want  []string
	}{
		"empty":  {value: "", want: nil},
		"single": {value: "10.0.0.5", want: []string{"10.0.0.5"}},
		"trims and drops blanks": {
			value: " 10.0.0.5 , ,172.16.0.0/12,",
			want:  []string{"10.0.0.5", "172.16.0.0/12"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(key, tc.value)
			assert.Equal(t, tc.want, getListEnv(key))
		})
	}
}
