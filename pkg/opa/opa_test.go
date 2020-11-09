// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
