package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodSecurityPolicyBindingSpec defines the desired state of PodSecurityPolicyBinding
// +k8s:openapi-gen=true
type PodSecurityPolicyBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Policy string `json:"policy"`

	// +kubebuilder:validation:MinItems=1
	Subjects []rbacv1.Subject `json:"subjects"`
}

// PodSecurityPolicyBindingStatus defines the observed state of PodSecurityPolicyBinding
// +k8s:openapi-gen=true
type PodSecurityPolicyBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSecurityPolicyBinding is the Schema for the podsecuritypolicybindings API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=podsecuritypolicybindings,scope=Namespaced
type PodSecurityPolicyBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodSecurityPolicyBindingSpec   `json:"spec,omitempty"`
	Status PodSecurityPolicyBindingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSecurityPolicyBindingList contains a list of PodSecurityPolicyBinding
type PodSecurityPolicyBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodSecurityPolicyBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodSecurityPolicyBinding{}, &PodSecurityPolicyBindingList{})
}
