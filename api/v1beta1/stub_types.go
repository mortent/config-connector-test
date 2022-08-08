/*
Copyright 2022.

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
	cc "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StubSpec defines the desired state of Stub
type StubSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ConditionSequence []ConditionSpec `json:"conditionSequence"`
}

type ConditionSpec struct {
	LatencySeconds int64 `json:"latencySeconds,omitempty"`

	Condition cc.Condition `json:"condition,omitempty"`
}

// StubStatus defines the observed state of Stub
type StubStatus struct {
	Conditions []cc.Condition `json:"conditions,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Stub is the Schema for the stubs API
type Stub struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StubSpec   `json:"spec,omitempty"`
	Status StubStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StubList contains a list of Stub
type StubList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Stub `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Stub{}, &StubList{})
}
