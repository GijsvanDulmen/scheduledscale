package deploymentscaling

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/annotations"
)

type DeploymentScaling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentScalingSpec   `json:"spec"`
	Status DeploymentScalingStatus `json:"status"`
}

type DeploymentScalingSpec struct {
	Deployment DeploymentMatchLabels `json:"deployment"`
	ScaleTo    []ScaleTo             `json:"scaleTo"`
	OnDelete   *OnDelete             `json:"onDelete"`
}

type OnDelete struct {
	RemovePodDisruptionBudget *bool `json:"removePodDisruptionBudget"`
}

type DeploymentMatchLabels struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type ScaleTo struct {
	At                  string                     `json:"at"`
	Replicas            int32                      `json:"replicas"`
	PodDisruptionBudget *PodDisruptionBudgetEnable `json:"podDisruptionBudget,omitempty"`
	Annotations         *annotations.Annotations   `json:"annotations"`
}

type PodDisruptionBudgetEnable struct {
	MinAvailable *int32 `json:"minAvailable"`
	MaxAvailable *int32 `json:"maxAvailable"`
}

type DeploymentScalingStatus struct {
	ErrorMessage string `json:"errorMessage"`
}

type DeploymentScalingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeploymentScaling `json:"items"`
}
