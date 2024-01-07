package informer

import (
	"github.com/stretchr/testify/assert"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/cronjobsuspend"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
	"testing"
)

func TestCreateRemovePatch(t *testing.T) {
	result := CreateRemovePatch("keya", "/path/%s")
	assert.Equal(t, string(result), "[{\"op\":\"remove\",\"path\":\"/path/keya\"}]")
}

func TestCreateRemovePatchOddKey(t *testing.T) {
	result := CreateRemovePatch("key/s", "/path/%s")
	assert.Equal(t, string(result), "[{\"op\":\"remove\",\"path\":\"/path/key~1s\"}]")
}

func TestCreateDeploymentScalingPatch(t *testing.T) {
	st := deploymentscaling.ScaleTo{
		At:                  "* * * * *",
		Replicas:            1,
		PodDisruptionBudget: nil,
		Annotations:         nil,
	}

	result := CreateDeploymentScalingPatch(&st)
	assert.Equal(t, string(result), "{\"spec\":{\"replicas\":1},\"metadata\":{\"annotations\":{}}}")
}

func TestCreateDeploymentScalingPatchWithAnnotations(t *testing.T) {
	st := deploymentscaling.ScaleTo{
		At:                  "* * * * *",
		Replicas:            1,
		PodDisruptionBudget: nil,
		Annotations: &common.Annotations{
			Add: map[string]string{
				"a": "b",
				"c": "d",
			},
			Remove: nil,
		},
	}

	result := CreateDeploymentScalingPatch(&st)
	assert.Equal(t, string(result), "{\"spec\":{\"replicas\":1},\"metadata\":{\"annotations\":{\"a\":\"b\",\"c\":\"d\"}}}")
}

func TestCreateCronJobSuspendPatch(t *testing.T) {
	st := cronjobsuspend.StateAt{
		At:          "",
		Suspend:     false,
		Annotations: nil,
	}

	result := CreateCronJobSuspendPatch(&st)
	assert.Equal(t, string(result), "{\"spec\":{\"suspend\":false,\"jobTemplate\":{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{}}}}}}}")
}

func TestCreateCronJobSuspendPatchWithAnnotations(t *testing.T) {
	st := cronjobsuspend.StateAt{
		At:      "",
		Suspend: true,
		Annotations: &common.Annotations{
			Add: map[string]string{
				"a": "b",
				"c": "d",
			},
			Remove: nil,
		},
	}

	result := CreateCronJobSuspendPatch(&st)
	assert.Equal(t, string(result), "{\"spec\":{\"suspend\":true,\"jobTemplate\":{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"a\":\"b\",\"c\":\"d\"}}}}}}}")
}
