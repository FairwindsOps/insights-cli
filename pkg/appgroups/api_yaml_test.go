package appgroups

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Fixtures follow Insights app-group API JSON, including fields the CLI does not store.
// Download writes YAML through utils.GetYamlBytes. Push reads it back with gopkg.in/yaml.v3.
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

			var group AppGroup
			require.NoError(t, json.Unmarshal(raw, &group))

			got, err := group.GetYamlBytes()
			require.NoError(t, err)

			again, err := group.GetYamlBytes()
			require.NoError(t, err)
			assert.Equal(t, string(got), string(again))

			wantPath := filepath.Join("testdata/api", strings.TrimSuffix(entry.Name(), ".json")+".yaml")
			want, err := os.ReadFile(wantPath)
			require.NoError(t, err)
			assert.Equal(t, string(want), string(got))

			dec := yaml.NewDecoder(bytes.NewReader(got))
			var back AppGroup
			require.NoError(t, dec.Decode(&back))
			require.ErrorIs(t, dec.Decode(&AppGroup{}), io.EOF)

			assert.Equal(t, group.Name, back.Name)
			assert.Equal(t, group.Type, back.Type)
			assert.NotContains(t, string(got), "createdAt")
			assert.NotContains(t, string(got), "updatedAt")

			if entry.Name() == "from-api.json" {
				assert.Equal(t, "AppGroup", back.Type)
				assert.Equal(t, []string{"*"}, back.Spec.Match[0].Namespaces)
				labels := back.Spec.Match[0].Labels[0]
				assert.Equal(t, "true", labels["policy-applied"])
				assert.Equal(t, "yes", labels["flag"])
				assert.Equal(t, "team: platform", labels["note"])
				assert.Equal(t, "1", labels["tier"])
				nsLabels := back.Spec.Match[0].NamespaceLabels[0]
				assert.Equal(t, "true", nsLabels["policy-applied"])
				assert.Equal(t, true, nsLabels["critical"])
				assert.Equal(t, 1.0, asFloat(t, nsLabels["priority"]))
				return
			}

			assert.Equal(t, group.Spec, back.Spec)
		})
	}
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
	default:
		t.Fatalf("got %T (%v), want a number", v, v)
		return 0
	}
}
