package policymappings

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

// Fixtures follow Insights policy-mapping API JSON.
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

			var mapping PolicyMapping
			require.NoError(t, json.Unmarshal(raw, &mapping))

			got, err := mapping.GetYamlBytes()
			require.NoError(t, err)

			again, err := mapping.GetYamlBytes()
			require.NoError(t, err)
			assert.Equal(t, string(got), string(again))

			wantPath := filepath.Join("testdata/api", strings.TrimSuffix(entry.Name(), ".json")+".yaml")
			want, err := os.ReadFile(wantPath)
			require.NoError(t, err)
			assert.Equal(t, string(want), string(got))

			dec := yaml.NewDecoder(bytes.NewReader(got))
			var back PolicyMapping
			require.NoError(t, dec.Decode(&back))
			require.ErrorIs(t, dec.Decode(&PolicyMapping{}), io.EOF)

			assert.Equal(t, mapping.Name, back.Name)
			assert.Equal(t, mapping.Type, back.Type)
			assert.Equal(t, mapping.Spec, back.Spec)
			assert.NotContains(t, string(got), "createdAt")
			assert.NotContains(t, string(got), "updatedAt")
			if mapping.Spec.Enabled == nil {
				assert.NotContains(t, string(got), "enabled:")
			}
			if mapping.Spec.Block == nil {
				assert.NotContains(t, string(got), "block:")
			}
		})
	}
}
