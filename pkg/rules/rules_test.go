// Copyright 2022 FairwindsOps Inc
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

package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FWI-3100 - the rule variable was not being reset - so previous values were being re-used
func TestGetRulesFromFiles(t *testing.T) {
	folderRules := map[string][]string{
		"testdata": {"testdata/rule1.yaml", "testdata/rule2.yaml"},
	}
	rules, err := getRulesFromFiles(folderRules)
	assert.NoError(t, err)

	assert.Equal(t, "us-east-1", rules[0].Cluster)
	assert.Equal(t, "", rules[1].Cluster)
}
