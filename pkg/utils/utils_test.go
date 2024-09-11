package utils

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestInvertBoolPointer(t *testing.T) {
	assert.Nil(t, InvertBoolPointer(nil))
	assert.True(t, *InvertBoolPointer(lo.ToPtr(false)))
	assert.False(t, *InvertBoolPointer(lo.ToPtr(true)))
}
