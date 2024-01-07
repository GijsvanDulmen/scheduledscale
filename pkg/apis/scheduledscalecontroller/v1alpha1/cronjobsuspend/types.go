package cronjobsuspend

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/common"
)

type CronJobSuspend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSuspendSpec `json:"spec"`
	Status common.Status      `json:"status"`
}

type CronJobSuspendSpec struct {
	CronJob common.MatchLabels `json:"cronjob"`
	StateAt []StateAt          `json:"stateAt"`
}

type StateAt struct {
	At          string              `json:"at"`
	Suspend     bool                `json:"suspend"`
	Annotations *common.Annotations `json:"annotations"`
}

type CronJobSuspendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CronJobSuspend `json:"items"`
}
