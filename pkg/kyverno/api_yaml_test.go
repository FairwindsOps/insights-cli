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

package kyverno

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// API list responses are flat JSON (name, kind, spec, labels), not Kubernetes YAML.
// Download rebuilds a policy document from that JSON. These fixtures are Insights
// examples plus one payload of scalars that YAML encoders tend to retype.
func TestDownloadedYAMLFromAPIJSON(t *testing.T) {
	entries, err := os.ReadDir("testdata/api")
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata/api", entry.Name()))
			require.NoError(t, err)

			var policy KyvernoPolicy
			require.NoError(t, json.Unmarshal(raw, &policy))

			got, err := policy.GetYamlBytes()
			require.NoError(t, err)

			again, err := policy.GetYamlBytes()
			require.NoError(t, err)
			assert.Equal(t, string(got), string(again))

			wantPath := filepath.Join("testdata/api", strings.TrimSuffix(entry.Name(), ".json")+".yaml")
			want, err := os.ReadFile(wantPath)
			require.NoError(t, err)
			assert.Equal(t, string(want), string(got))

			path := filepath.Join(t.TempDir(), "policy.yaml")
			require.NoError(t, os.WriteFile(path, got, 0o644))
			back, err := ReadPolicyFromFile(path)
			require.NoError(t, err)

			assert.Equal(t, policy.Name, back.Name)
			assert.Equal(t, policy.Kind, back.Kind)
			assert.Equal(t, policy.APIVersion, back.APIVersion)
			assert.Equal(t, policy.Namespace, back.Namespace)
			assert.Equal(t, policy.Labels, back.Labels)
			assert.Equal(t, policy.Annotations, back.Annotations)
			assert.Nil(t, back.CreatedAt)
			assert.Nil(t, back.ManagedByInsights)
			assert.NotContains(t, string(got), "managedByInsights")
			assert.NotContains(t, string(got), "appliedAppGroups")
			assert.NotContains(t, string(got), "createdAt")

			if entry.Name() == "edge-policy.json" {
				assertEdgePolicyYAML(t, string(got), back)
			}
		})
	}
}

func assertEdgePolicyYAML(t *testing.T, yamlText string, back KyvernoPolicy) {
	t.Helper()

	// 2-space indent, sequence items indented under the key (v3 encoder style).
	assert.Contains(t, yamlText, "  name: edge-policy\n  namespace: production\n")
	assert.Contains(t, yamlText, "  rules:\n    - cel:\n")
	assert.NotContains(t, yamlText, "null")
	assert.NotContains(t, yamlText, "foreach")
	assert.NotContains(t, yamlText, "conditions:")
	assert.NotContains(t, yamlText, "ignored-name")
	assert.NotContains(t, yamlText, "ignored-namespace")
	assert.NotContains(t, yamlText, "2025-06-10")
	assert.Contains(t, yamlText, "generateName: edge-")
	assert.Contains(t, yamlText, "namespace: production")

	assert.Equal(t, "yes", back.Labels["flag"])
	assert.Equal(t, "on", back.Labels["switch"])
	assert.Equal(t, "no", back.Labels["off"])
	assert.Equal(t, "false", back.Labels["quoted-false"])
	assert.Equal(t, "1", back.Labels["tier"])
	assert.Equal(t, "", back.Labels["empty"])
	assert.Equal(t, "?*", back.Annotations["example.com/pattern"])
	assert.Equal(t, "!*:latest", back.Annotations["example.com/image"])
	assert.Equal(t, "Pod {{ request.object.metadata.name }} must set app", back.Annotations["example.com/template"])
	assert.Equal(t, "edge-", back.Metadata["generateName"])

	privileged := lookup(t, back.Spec, "rules", "0", "validate", "pattern", "spec", "containers", "0", "securityContext", "privileged")
	assert.Equal(t, "false", privileged)
	runAsNonRoot := lookup(t, back.Spec, "rules", "0", "validate", "pattern", "spec", "containers", "0", "securityContext", "runAsNonRoot")
	assert.Equal(t, true, runAsNonRoot)
	assert.Equal(t, 2.0, asFloat(t, lookup(t, back.Spec, "rules", "0", "validate", "pattern", "spec", "replicas")))
	assert.Equal(t, 1.5, asFloat(t, lookup(t, back.Spec, "rules", "0", "validate", "pattern", "spec", "ratio")))
	assert.Equal(t, 10.0, asFloat(t, lookup(t, back.Spec, "webhookTimeoutSeconds")))
	assert.Equal(t, true, back.Spec["background"])
	assert.Equal(t, false, back.Spec["schemaValidation"])
	assert.Equal(t, true, back.Status["ready"])
	assert.Equal(t, "", back.Status["message"])

	expression := lookup(t, back.Spec, "rules", "0", "cel", "expressions", "0", "expression")
	assert.Equal(t, "has(object.metadata.annotations) &&\n(\"app.kubernetes.io/name\" in object.metadata.annotations) &&\nobject.spec.replicas > 0\n", expression)
}

func lookup(t *testing.T, v any, path ...string) any {
	t.Helper()
	cur := v
	for _, p := range path {
		switch n := cur.(type) {
		case map[string]any:
			next, ok := n[p]
			if !ok {
				t.Fatalf("missing %q in %#v", p, n)
			}
			cur = next
		case []any:
			i, err := strconv.Atoi(p)
			if err != nil || i < 0 || i >= len(n) {
				t.Fatalf("missing index %q in %#v", p, n)
			}
			cur = n[i]
		default:
			t.Fatalf("cannot walk %T with %q", cur, p)
		}
	}
	return cur
}

func asFloat(t *testing.T, v any) float64 {
	t.Helper()
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case uint64:
		return float64(n)
	default:
		t.Fatalf("got %T (%v), want a number", v, v)
		return 0
	}
}
