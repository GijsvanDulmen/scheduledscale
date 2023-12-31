package informer

import (
	"encoding/json"
	"fmt"
	"strings"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/cronjobsuspend"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
)

type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

func (informer *Informer) CreateRemovePatch(key string, path string) []byte {
	key = strings.ReplaceAll(key, "/", "~1")
	payload := []patchUInt32Value{{
		Op:   "remove",
		Path: fmt.Sprintf(path, key),
	}}

	removePayload, _ := json.Marshal(payload)
	return removePayload
}

func (informer *Informer) CreateDeploymentScalingPatch(scaleTo *deploymentscaling.ScaleTo) []byte {
	deploymentPatch := PatchForDeployment{
		Spec: DeploymentSpec{
			Replicas: uint32(scaleTo.Replicas),
		},
		Metadata: MetaDataAnnotations{
			Annotations: map[string]string{},
		},
	}

	if scaleTo.Annotations != nil {
		if scaleTo.Annotations.Add != nil {
			for anKey, anValue := range scaleTo.Annotations.Add {
				deploymentPatch.Metadata.Annotations[anKey] = anValue
			}
		}
	}

	payloadBytes, _ := json.Marshal(deploymentPatch)
	return payloadBytes
}

func (informer *Informer) CreateCronJobSuspendPatch(stateAt *cronjobsuspend.StateAt) []byte {
	cronjobPatch := PatchForCronJob{
		Spec: CronJobSpec{
			Suspend: stateAt.Suspend,
			JobTemplate: CronJobJobTemplate{
				Spec: CronJobJobTemplateSpec{
					Template: CronJobJobTemplateSpecTemplate{
						Metadata: MetaDataAnnotations{
							Annotations: map[string]string{},
						},
					},
				},
			},
		},
	}

	if stateAt.Annotations != nil {
		if stateAt.Annotations.Add != nil {
			for anKey, anValue := range stateAt.Annotations.Add {
				cronjobPatch.Spec.JobTemplate.Spec.Template.Metadata.Annotations[anKey] = anValue
			}
		}
	}

	payloadBytes, _ := json.Marshal(cronjobPatch)
	return payloadBytes
}
