package informer

type PatchForCronJob struct {
	Spec CronJobSpec `json:"spec"`
}

type CronJobSpec struct {
	Suspend     bool               `json:"suspend"`
	JobTemplate CronJobJobTemplate `json:"jobTemplate"`
}

type CronJobJobTemplate struct {
	Spec CronJobJobTemplateSpec `json:"spec"`
}

type CronJobJobTemplateSpec struct {
	Template CronJobJobTemplateSpecTemplate `json:"template"`
}

type CronJobJobTemplateSpecTemplate struct {
	Metadata MetaDataAnnotations `json:"metadata"`
}

type MetaDataAnnotations struct {
	Annotations map[string]string `json:"annotations"`
}

type PatchForDeployment struct {
	Spec     DeploymentSpec      `json:"spec"`
	Metadata MetaDataAnnotations `json:"metadata"`
}

type DeploymentSpec struct {
	Replicas uint32 `json:"replicas"`
}

type PatchForHpa struct {
	Spec     HpaSpec             `json:"spec"`
	Metadata MetaDataAnnotations `json:"metadata"`
}

type HpaSpec struct {
	MinReplicas *uint32 `json:"minReplicas"`
	MaxReplicas *uint32 `json:"maxReplicas"`
}
