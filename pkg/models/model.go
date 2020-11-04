package models

// KubernetesTarget represents a set of API Versions and Kinds to find Kubernetes objects.
type KubernetesTarget struct {
	APIGroups []string `json:"apiGroups" yaml:"apiGroups"`
	Kinds     []string
}

// OutputModel is a model to represent settings for Action Items created by OPA.
type OutputModel struct {
	Title       *string
	Severity    *float64
	Remediation *string
	Category    *string
}

// CustomCheckModel is a model for the API endpoint to receive a Custom Check for OPA
type CustomCheckModel struct {
	CheckName string `json:"-" yaml:"-"`
	Output    *OutputModel
	Rego      string
	Instances []CustomCheckInstanceModel `json:"-" yaml:"-"`
}

// CustomCheckInstanceModel is a model for the API endpoint to receive an Instance for a Custom Check in OPA
type CustomCheckInstanceModel struct {
	CheckName    string `json:"-" yaml:"-"`
	InstanceName string `json:"-" yaml:"-"`
	Targets      []KubernetesTarget
	Clusters     []string
	Parameters   map[string]interface{}
	Output       *OutputModel
}
