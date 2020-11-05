package opa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotEqual(t *testing.T) {
	floatVal1 := 1.0
	floatVal2 := 2.0
	stringVal1 := "one"
	stringVal2 := "two"
	floatPointer1 := &floatVal1
	floatPointer2 := &floatVal1
	assert.False(t, notEqual(floatPointer1, floatPointer2))
	floatPointer2 = &floatVal2
	assert.True(t, notEqual(floatPointer1, floatPointer2))
	floatPointer1 = nil
	assert.True(t, notEqual(floatPointer1, floatPointer2))
	floatPointer2 = nil
	assert.False(t, notEqual(floatPointer1, floatPointer2))
	stringPointer1 := &stringVal1
	stringPointer2 := &stringVal2
	assert.True(t, notEqual(stringPointer1, stringPointer2))
	assert.False(t, notEqual(stringPointer1, stringPointer1))
	assert.True(t, notEqual(stringPointer1, floatPointer2))
}

func TestCompareCheck(t *testing.T) {
	// TODO implement
}

func TestTargetsNotEqual(t *testing.T) {
	// TODO implement
}
