package opavalidation_test

import (
	"context"
	"os"
	"testing"

	"github.com/fairwindsops/insights-cli/pkg/opavalidation"
	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
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
			gotActionItems, gotErr := opavalidation.ValidateRego(context.TODO(), regoAsString, objectAsBytes, fwrego.InsightsInfo{}, "TestEvent", "", nil)
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
