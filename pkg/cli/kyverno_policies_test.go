// Copyright 2025 FairwindsOps Inc
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

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKyvernoPoliciesCommandStructure(t *testing.T) {
	// Test that the kyverno-policies command exists and has the expected subcommands
	assert.NotNil(t, kyvernoPoliciesCmd, "kyverno-policies command should exist")

	// Test that cluster subcommand exists
	assert.NotNil(t, kyvernoPoliciesClusterCmd, "kyverno-policies cluster command should exist")

	// Test that local listing command exists
	assert.NotNil(t, listLocalKyvernoPoliciesCmd, "list-local command should exist")

	// Test that cluster subcommands exist
	assert.NotNil(t, listClusterKyvernoPoliciesCmd, "cluster list command should exist")
	assert.NotNil(t, listClusterKyvernoPoliciesWithAppGroupsCmd, "cluster list-with-app-groups command should exist")
	assert.NotNil(t, exportClusterKyvernoPoliciesYamlCmd, "cluster export-yaml command should exist")
	assert.NotNil(t, validateClusterKyvernoPoliciesCmd, "cluster validate command should exist")
}

func TestKyvernoPoliciesCommandNames(t *testing.T) {
	// Test command names
	assert.Equal(t, "kyverno-policies", kyvernoPoliciesCmd.Use)
	assert.Equal(t, "cluster", kyvernoPoliciesClusterCmd.Use)
	assert.Equal(t, "list-local", listLocalKyvernoPoliciesCmd.Use)
	assert.Equal(t, "list <cluster>", listClusterKyvernoPoliciesCmd.Use)
	assert.Equal(t, "list-with-app-groups <cluster>", listClusterKyvernoPoliciesWithAppGroupsCmd.Use)
	assert.Equal(t, "export-yaml <cluster>", exportClusterKyvernoPoliciesYamlCmd.Use)
	assert.Equal(t, "validate <cluster>", validateClusterKyvernoPoliciesCmd.Use)
}

func TestKyvernoPoliciesCommandDescriptions(t *testing.T) {
	// Test that commands have proper descriptions
	assert.NotEmpty(t, kyvernoPoliciesCmd.Short, "kyverno-policies command should have a short description")
	assert.NotEmpty(t, kyvernoPoliciesClusterCmd.Short, "cluster command should have a short description")
	assert.NotEmpty(t, listLocalKyvernoPoliciesCmd.Short, "list-local command should have a short description")
	assert.NotEmpty(t, listClusterKyvernoPoliciesCmd.Short, "cluster list command should have a short description")
	assert.NotEmpty(t, listClusterKyvernoPoliciesWithAppGroupsCmd.Short, "cluster list-with-app-groups command should have a short description")
	assert.NotEmpty(t, exportClusterKyvernoPoliciesYamlCmd.Short, "cluster export-yaml command should have a short description")
	assert.NotEmpty(t, validateClusterKyvernoPoliciesCmd.Short, "cluster validate command should have a short description")
}
