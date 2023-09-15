package appgroups

// copied from FW Insights openAPI codegen

type AppGroup struct {
	Name string       `json:"name,omitempty" yaml:"name,omitempty"`
	Spec AppGroupSpec `json:"spec" yaml:"spec"`
	Type string       `json:"type,omitempty" yaml:"type,omitempty"`
}

// AppGroupSpec defines model for AppGroupSpec.
type AppGroupSpec struct {
	Exclude []AppGroupSpecCriteria `json:"exclude,omitempty" yaml:"exclude,omitempty"`
	Match   []AppGroupSpecCriteria `json:"match,omitempty" yaml:"match,omitempty"`
}

// AppGroupSpecCriteria defines model for AppGroupSpecCriteria.
type AppGroupSpecCriteria struct {
	Cluster         string         `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Container       string         `json:"container,omitempty" yaml:"container,omitempty"`
	Kind            string         `json:"kind,omitempty" yaml:"kind,omitempty"`
	Labels          map[string]any `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name            string         `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace       string         `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	NamespaceLabels map[string]any `json:"namespaceLabels,omitempty" yaml:"namespaceLabels,omitempty"`
	Repository      string         `json:"repository,omitempty" yaml:"repository,omitempty"`
}

func (a AppGroup) GetName() string {
	return a.Name
}
