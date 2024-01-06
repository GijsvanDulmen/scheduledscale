package informer

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
	"testing"
)

func TestCreatePodDisruptionBudget(t *testing.T) {
	minAvailable := int32(2)
	maxAvailable := int32(3)

	st := deploymentscaling.ScaleTo{
		At:       "",
		Replicas: 2,
		PodDisruptionBudget: &deploymentscaling.PodDisruptionBudgetEnable{
			MinAvailable: &minAvailable,
			MaxAvailable: &maxAvailable,
		},
		Annotations: nil,
	}

	ds := deploymentscaling.DeploymentScaling{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
		},
		Spec: deploymentscaling.DeploymentScalingSpec{
			Deployment: deploymentscaling.DeploymentMatchLabels{},
			ScaleTo:    nil,
			OnDelete:   nil,
		},
		Status: deploymentscaling.DeploymentScalingStatus{},
	}

	deploy := v1.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "deployns",
		},
		Spec: v1.DeploymentSpec{
			Replicas: nil,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "app1",
					"sel2": "sel2",
				},
				MatchExpressions: nil,
			},
			Template:                v12.PodTemplateSpec{},
			Strategy:                v1.DeploymentStrategy{},
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},
		Status: v1.DeploymentStatus{},
	}

	result := CreatePodDisruptionBudget(&st, &ds, &deploy)

	assert.Equal(t, result.Namespace, "deployns")
	assert.Equal(t, result.Name, deploy.Name)
	assert.Equal(t, result.OwnerReferences[0].Name, ds.Name)
	assert.Equal(t, result.OwnerReferences[0].APIVersion, ds.APIVersion)
	assert.Equal(t, result.OwnerReferences[0].Kind, ds.Kind)
	assert.Equal(t, result.OwnerReferences[0].UID, ds.UID)

	assert.Equal(t, result.Spec.MinAvailable.IntValue(), 2)
	assert.Equal(t, result.Spec.MaxUnavailable.IntValue(), 3)
	assert.Equal(t, result.Spec.Selector.MatchLabels["app"], "app1")
	assert.Equal(t, result.Spec.Selector.MatchLabels["sel2"], "sel2")
}
