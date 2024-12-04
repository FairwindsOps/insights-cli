package policymappings

// copied from FW Insights openAPI codegen

type PolicyMapping struct {
	Name string            `json:"name,omitempty" yaml:"name,omitempty"`
	Spec PolicyMappingSpec `json:"spec" yaml:"spec"`
	Type string            `json:"type,omitempty" yaml:"type,omitempty"`
}

type PolicyMappingSpec struct {
	Enabled   *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AppGroups []string `json:"appGroups,omitempty" yaml:"appGroups,omitempty"`
	Contexts  []string `json:"contexts,omitempty" yaml:"contexts,omitempty"`
	Policies  []string `json:"policies,omitempty" yaml:"policies,omitempty"`
	Block     *bool    `json:"block,omitempty" yaml:"block,omitempty"`
}

func (pm PolicyMapping) GetName() string {
	return pm.Name
}
