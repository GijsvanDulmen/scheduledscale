package cronjobsuspend

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/annotations"
)

type CronJobSuspend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSuspendSpec   `json:"spec"`
	Status CronJobSuspendStatus `json:"status"`
}

type CronJobSuspendSpec struct {
	CronJob CronJobMatchLabels `json:"cronjob"`
	StateAt []StateAt          `json:"stateAt"`
}

type CronJobMatchLabels struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type StateAt struct {
	At          string                   `json:"at"`
	Suspend     bool                     `json:"suspend"`
	Annotations *annotations.Annotations `json:"annotations"`
}

type CronJobSuspendStatus struct {
	ErrorMessage string `json:"errorMessage"`
}

type CronJobSuspendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CronJobSuspend `json:"items"`
}
