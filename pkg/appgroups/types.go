package appgroups

// copied from FW Insights openAPI codegen

type AppGroup struct {
	Name string       `json:"name,omitempty" yaml:"name,omitempty"`
	Spec AppGroupSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
	Type string       `json:"type,omitempty" yaml:"type,omitempty"`
}

// AppGroupSpec defines model for AppGroupSpec.
type AppGroupSpec struct {
	Exclude []AppGroupSpecCriteria `json:"exclude,omitempty" yaml:"exclude,omitempty"`
	Match   []AppGroupSpecCriteria `json:"match,omitempty" yaml:"match,omitempty"`
}

// AppGroupSpecCriteria defines model for AppGroupSpecCriteria.
type AppGroupSpecCriteria struct {
	Clusters        []string         `json:"clusters,omitempty" yaml:"clusters,omitempty"`
	Containers      []string         `json:"containers,omitempty" yaml:"containers,omitempty"`
	Kinds           []string         `json:"kinds,omitempty" yaml:"kinds,omitempty"`
	Names           []string         `json:"names,omitempty" yaml:"names,omitempty"`
	Namespaces      []string         `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	Repositories    []string         `json:"repositories,omitempty" yaml:"repositories,omitempty"`
	NamespaceLabels []map[string]any `json:"namespaceLabels,omitempty" yaml:"namespaceLabels,omitempty"`
	Labels          []map[string]any `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func (a AppGroup) GetName() string {
	return a.Name
}
