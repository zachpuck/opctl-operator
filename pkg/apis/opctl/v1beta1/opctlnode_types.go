/*
Copyright 2018 Zach Puckett.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceSpec is used to define the k8s service created for a opctl node
type ServiceSpec struct {
	// Opctl Service Name
	Name string `json:"name,omitempty"`

	// Opctl node service type
	ServiceType corev1.ServiceType `json:"servicetype,omitempty"`

	// Opctl service annotations
	Annotations map[string]string `json:"annotations,omitempty"`
}

// OpctlNodeSpec defines the desired state of OpctlNode
type OpctlNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// What container image to use for a new opctl node
	Image string `json:"image,omitempty"`

	// Dictionary of environment variable values
	Env map[string]string `json:"env,omitempty"`

	// Opctl deployment annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Opctl service options
	Service *ServiceSpec `json:"service,omitempty"`
}

// OpctlNodeStatus defines the observed state of OpctlNode
type OpctlNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// state if opctl node
	Phase string `json:"phase"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpctlNode is the Schema for the opctlnodes API
// +k8s:openapi-gen=true
type OpctlNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpctlNodeSpec   `json:"spec,omitempty"`
	Status OpctlNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpctlNodeList contains a list of OpctlNode
type OpctlNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpctlNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpctlNode{}, &OpctlNodeList{})
}
