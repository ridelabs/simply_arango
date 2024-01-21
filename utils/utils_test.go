package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripWS(t *testing.T) {
	assert.Equal(t, "a b c d ", StripExtraWS("\na    \t\tb       c\n\n     d     "))
}
