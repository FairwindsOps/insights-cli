package opavalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetObjectFileNamesForPolicy(t *testing.T) {
	opts := ExpectActionItemOptions{
		Default:              true,
		SuccessFileExtension: ".success.yaml",
		FailureFileExtension: ".failure.yaml",
	}

	f, ok, err := opts.getObjectFileNamesForPolicy("testdata/multiple-validations/rego.rego")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Len(t, f, 6)
}
