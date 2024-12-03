/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppsSpec defines the desired state of Apps
type AppsSpec struct {
	Replicas    int32  `json:"replicas"`
	Image       string `json:"image"`
	Port        int32  `json:"port"`
	IngressHost string `json:"ingressHost,omitempty"` // 自定义域名，留空则不创建 Ingress
}

// AppsStatus defines the observed state of Apps
type AppsStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Apps is the Schema for the apps API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="Desired number of replicas"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image",description="Container image"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.port",description="Container port"
// +kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableReplicas",description="Available replicas"
// +kubebuilder:printcolumn:name="IngressHost",type="string",JSONPath=".spec.ingressHost",description="Ingress Hostname"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age of the resource"
type Apps struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppsSpec   `json:"spec,omitempty"`
	Status AppsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppsList contains a list of Apps
type AppsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Apps `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Apps{}, &AppsList{})
}
