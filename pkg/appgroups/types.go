package appgroups

// copied from FW Insights openAPI codegen

type AppGroup struct {
	Name string       `json:"name" yaml:"name"`
	Spec AppGroupSpec `json:"spec" yaml:"spec"`
	Type string       `json:"type" yaml:"type"`
}

// AppGroupSpec defines model for AppGroupSpec.
type AppGroupSpec struct {
	Exclude []AppGroupSpecCriteria `json:"exclude,omitempty" yaml:"exclude,omitempty"`
	Match   []AppGroupSpecCriteria `json:"match,omitempty" yaml:"match,omitempty"`
}

// AppGroupSpecCriteria defines model for AppGroupSpecCriteria.
type AppGroupSpecCriteria struct {
	Cluster         string         `json:"cluster" yaml:"cluster"`
	Container       string         `json:"container" yaml:"container"`
	Kind            string         `json:"kind" yaml:"kind"`
	Labels          map[string]any `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name            string         `json:"name" yaml:"name"`
	Namespace       string         `json:"namespace" yaml:"namespace"`
	NamespaceLabels map[string]any `json:"namespaceLabels,omitempty" yaml:"namespaceLabels,omitempty"`
	Repository      string         `json:"repository" yaml:"repository"`
}
