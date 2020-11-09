// Copyright 2020 FairwindsOps Inc
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
	Output    OutputModel
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
	Output       OutputModel
}
