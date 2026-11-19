package asset_repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextArray_Value(t *testing.T) {
	q := string(rune(doubleQuote))
	bs := string(rune(backslash))

	tests := []struct {
		name string
		in   textArray
		want string
	}{
		{"empty", textArray{}, "{}"},
		{"plain values", textArray{"SS-1", "TX-1"}, "{" + q + "SS-1" + q + "," + q + "TX-1" + q + "}"},
		{"commas, spaces and braces stay inside quotes", textArray{"a, b", "{x}"}, "{" + q + "a, b" + q + "," + q + "{x}" + q + "}"},
		{"a quote is escaped", textArray{"say " + q + "hi" + q}, "{" + q + "say " + bs + q + "hi" + bs + q + q + "}"},
		{"a backslash is escaped", textArray{"C:" + bs + "path"}, "{" + q + "C:" + bs + bs + "path" + q + "}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
