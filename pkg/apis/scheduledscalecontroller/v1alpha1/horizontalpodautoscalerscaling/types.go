package horizontalpodautoscalerscaling

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/common"
)

type HorizontalPodAutoscalerScaling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HorizontalPodAutoscalerScalingSpec   `json:"spec"`
	Status HorizontalPodAutoscalerScalingStatus `json:"status"`
}

type HorizontalPodAutoscalerScalingSpec struct {
	HorizontalPodAutoscaler common.MatchLabels `json:"horizontalPodAutoscaler"`
	ScaleTo                 []ScaleTo          `json:"scaleTo"`
}

type ScaleTo struct {
	At          string              `json:"at"`
	MinReplicas *uint32             `json:"minReplicas"`
	MaxReplicas *uint32             `json:"maxReplicas"`
	Annotations *common.Annotations `json:"annotations"`
}

type HorizontalPodAutoscalerScalingStatus struct {
	ErrorMessage string `json:"errorMessage"`
}

type HorizontalPodAutoscalerScalingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []HorizontalPodAutoscalerScaling `json:"items"`
}
