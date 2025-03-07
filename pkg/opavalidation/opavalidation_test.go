package opavalidation_test

import (
	"context"
	"os"
	"testing"

	"github.com/fairwindsops/insights-cli/pkg/opavalidation"
	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
	"github.com/stretchr/testify/assert"
)

func TestValidateRego(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		description                 string
		regoFileName                string
		objectFileName              string
		expectNumActionItems        int
		expectNumInvalidActionItems int
		expectError                 bool
	}{
		{
			description:    "no package",
			regoFileName:   "nopackage.rego",
			objectFileName: "nopackage.yaml",
			expectError:    true,
		},
		{
			description:    "invalid rego",
			regoFileName:   "invalidrego.rego",
			objectFileName: "invalidrego.yaml",
			expectError:    true,
		},
		{
			description:                 "invalid action item",
			regoFileName:                "invalidactionitem.rego",
			objectFileName:              "invalidactionitem.yaml",
			expectNumActionItems:        2,
			expectNumInvalidActionItems: 1,
		},
		{
			description:    "missing action item",
			regoFileName:   "missingactionitem.rego",
			objectFileName: "missingactionitem.yaml",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			b, err := os.ReadFile("testdata/" + tc.regoFileName)
			if err != nil {
				t.Fatalf("error reading %s: %v", tc.regoFileName, err)
			}
			regoAsString := string(b)
			objectAsBytes, err := os.ReadFile("testdata/" + tc.objectFileName)
			if err != nil {
				t.Fatalf("error reading %s: %v", tc.objectFileName, err)
			}
			gotActionItems, gotErr := opavalidation.ValidateRego(context.TODO(), regoAsString, "v0", objectAsBytes, fwrego.InsightsInfo{}, "TestEvent", "", nil)
			if !tc.expectError && gotErr != nil {
				t.Fatal(gotErr)
			}
			if tc.expectError && gotErr == nil {
				t.Fatal("expected an error")
			}
			gotNumActionItems := len(gotActionItems)
			if tc.expectNumActionItems != gotNumActionItems {
				t.Fatalf("want %d action items, got %d", tc.expectNumActionItems, gotNumActionItems)
			}
			gotNumInvalidActionItems := gotActionItems.NumInvalid()
			if tc.expectNumInvalidActionItems != gotNumInvalidActionItems {
				t.Fatalf("want %d invalid action items, got %d", tc.expectNumInvalidActionItems, gotNumInvalidActionItems)
			}
		})
	}
}

func TestRunWithLibs(t *testing.T) {
	ais, err := opavalidation.Run("v0", "testdata/fileWithLib.rego", "testdata/pod1.yaml", opavalidation.ExpectActionItemOptions{}, fwrego.InsightsInfo{}, "", "testdata/libs")
	assert.NoError(t, err)
	assert.Len(t, ais, 0)
	ais, err = opavalidation.Run("v0", "testdata/fileWithLib.rego", "testdata/pod2.yaml", opavalidation.ExpectActionItemOptions{}, fwrego.InsightsInfo{}, "", "testdata/libs")
	assert.Error(t, err)
	assert.Equal(t, "1 action items were returned but none are expected", err.Error())
	assert.Len(t, ais, 1)
	assert.Equal(t, "Label is missing", ais[0].Title)
}

func TestMultipleRules(t *testing.T) {
	ais, err := opavalidation.Run("v0", "test/multipleRules.rego", "testdata/pod1.yaml", opavalidation.ExpectActionItemOptions{}, fwrego.InsightsInfo{}, "", "")
	assert.NoError(t, err)
	assert.Len(t, ais, 0)
}
