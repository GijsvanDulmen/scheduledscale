package deploymentscaling

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/common"
)

type DeploymentScaling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentScalingSpec `json:"spec"`
	Status common.Status         `json:"status"`
}

type DeploymentScalingSpec struct {
	Deployment common.MatchLabels  `json:"deployment"`
	ScaleTo    []ScaleTo           `json:"scaleTo"`
	OnDelete   *common.PdbOnDelete `json:"onDelete"`
}

type ScaleTo struct {
	At                  string                            `json:"at"`
	Replicas            int32                             `json:"replicas"`
	PodDisruptionBudget *common.PodDisruptionBudgetEnable `json:"podDisruptionBudget,omitempty"`
	Annotations         *common.Annotations               `json:"annotations"`
}

type DeploymentScalingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeploymentScaling `json:"items"`
}
