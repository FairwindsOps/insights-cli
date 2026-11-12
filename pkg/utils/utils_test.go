package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvertBoolPointer(t *testing.T) {
	assert.Nil(t, InvertBoolPointer(nil))
	assert.True(t, *InvertBoolPointer(new(false)))
	assert.False(t, *InvertBoolPointer(new(true)))
}

func TestGetYamlBytesScalarFormatting(t *testing.T) {
	type row struct {
		Note  string `yaml:"note"`
		Flag  string `yaml:"flag"`
		Count int    `yaml:"count"`
		On    *bool  `yaml:"on,omitempty"`
		Off   *bool  `yaml:"off,omitempty"`
		Skip  *bool  `yaml:"skip,omitempty"`
	}
	on, off := true, false
	got, err := GetYamlBytes(row{
		Note:  "team: platform # keep",
		Flag:  "yes",
		Count: 2,
		On:    &on,
		Off:   &off,
	})
	require.NoError(t, err)

	want, err := os.ReadFile("testdata/scalars.yaml")
	require.NoError(t, err)
	assert.Equal(t, string(want), string(got))
}
