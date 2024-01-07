package deploymentscaling

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/common"
)

type DeploymentScaling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentScalingSpec   `json:"spec"`
	Status DeploymentScalingStatus `json:"status"`
}

type DeploymentScalingSpec struct {
	Deployment common.MatchLabels `json:"deployment"`
	ScaleTo    []ScaleTo          `json:"scaleTo"`
	OnDelete   *OnDelete          `json:"onDelete"`
}

type OnDelete struct {
	RemovePodDisruptionBudget *bool `json:"removePodDisruptionBudget"`
}

type ScaleTo struct {
	At                  string                     `json:"at"`
	Replicas            int32                      `json:"replicas"`
	PodDisruptionBudget *PodDisruptionBudgetEnable `json:"podDisruptionBudget,omitempty"`
	Annotations         *common.Annotations        `json:"annotations"`
}

type PodDisruptionBudgetEnable struct {
	MinAvailable   *int32 `json:"minAvailable"`
	MaxUnavailable *int32 `json:"maxUnavailable"`
}

type DeploymentScalingStatus struct {
	ErrorMessage string `json:"errorMessage"`
}

type DeploymentScalingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeploymentScaling `json:"items"`
}
