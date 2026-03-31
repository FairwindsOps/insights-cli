package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvertBoolPointer(t *testing.T) {
	assert.Nil(t, InvertBoolPointer(nil))
	assert.True(t, *InvertBoolPointer(new(false)))
	assert.False(t, *InvertBoolPointer(new(true)))
}
