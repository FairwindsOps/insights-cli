// Copyright 2023 FairwindsOps Inc
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

	"github.com/fairwindsops/insights-plugins/plugins/opa/pkg/opa"
	"github.com/stretchr/testify/assert"

	"github.com/fairwindsops/insights-cli/pkg/models"
)

func TestCompareCheck(t *testing.T) {
	results := compareChecks(nil, nil)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))

	checks := []models.CustomCheckModel{
		{
			CheckName: "Check1",
		},
	}
	results = compareChecks(checks, nil)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 1, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))

	apiChecks := []opa.OPACustomCheck{
		{
			Name: "Check1",
		},
	}
	results = compareChecks(checks, apiChecks)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))

	results = compareChecks(nil, apiChecks)
	assert.Equal(t, 1, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
}
