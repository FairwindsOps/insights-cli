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
	results := compareChecks(nil, nil, nil)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
	assert.Equal(t, 0, len(results.InstanceDelete))
	assert.Equal(t, 0, len(results.InstanceInsert))
	assert.Equal(t, 0, len(results.InstanceUpdate))
	checks := []models.CustomCheckModel{
		{
			CheckName: "Check1",
			Instances: []models.CustomCheckInstanceModel{
				{
					CheckName:    "Check1",
					InstanceName: "instance2",
				},
			},
		},
	}
	results = compareChecks(checks, nil, nil)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 1, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
	assert.Equal(t, 0, len(results.InstanceDelete))
	assert.Equal(t, 1, len(results.InstanceInsert))
	assert.Equal(t, 0, len(results.InstanceUpdate))
	apiChecks := []opa.OPACustomCheck{
		{
			Name: "Check1",
		},
	}
	results = compareChecks(checks, apiChecks, nil)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
	assert.Equal(t, 0, len(results.InstanceDelete))
	assert.Equal(t, 1, len(results.InstanceInsert))
	assert.Equal(t, 0, len(results.InstanceUpdate))
	results = compareChecks(nil, apiChecks, nil)
	assert.Equal(t, 1, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
	assert.Equal(t, 0, len(results.InstanceDelete))
	assert.Equal(t, 0, len(results.InstanceInsert))
	assert.Equal(t, 0, len(results.InstanceUpdate))
	apiInstances := []opa.CheckSetting{
		{
			CheckName: "Check1",
		},
		{
			CheckName: "Check1",
		},
		{
			CheckName: "Check2",
		},
	}
	apiInstances[0].AdditionalData.Name = "instance2"
	results = compareChecks(nil, apiChecks, apiInstances)
	assert.Equal(t, 1, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
	assert.Equal(t, 2, len(results.InstanceDelete))
	assert.Equal(t, 0, len(results.InstanceInsert))
	assert.Equal(t, 0, len(results.InstanceUpdate))
	results = compareChecks(checks, apiChecks, apiInstances)
	assert.Equal(t, 0, len(results.CheckDelete))
	assert.Equal(t, 0, len(results.CheckInsert))
	assert.Equal(t, 0, len(results.CheckUpdate))
	assert.Equal(t, 1, len(results.InstanceDelete))
	assert.Equal(t, 0, len(results.InstanceInsert))
	assert.Equal(t, 0, len(results.InstanceUpdate))
	// TODO implement checks for updates, deletes, and instances
}

func TestTargetsNotEqual(t *testing.T) {
	apiTarget := []string{
		"core/type1",
		"core/type2",
		"other/type1",
		"other/type2",
	}
	fileTargets := []models.KubernetesTarget{
		{
			Kinds: []string{
				"type1",
				"type2",
			},
			APIGroups: []string{
				"core",
				"other",
			},
		},
	}
	assert.False(t, targetsNotEqual(apiTarget, fileTargets))
	apiTarget = []string{
		"core/type2",
		"core/type1",
		"other/type1",
	}
	assert.True(t, targetsNotEqual(apiTarget, fileTargets))
}
