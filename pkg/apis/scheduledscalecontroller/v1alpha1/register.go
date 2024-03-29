package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"scheduledscale/pkg/apis/scheduledscalecontroller"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/cronjobsuspend"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/horizontalpodautoscalerscaling"
)

var SchemeGroupVersion = schema.GroupVersion{Group: scheduledscalecontroller.GroupName, Version: "v1alpha1"}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&deploymentscaling.DeploymentScaling{},
		&deploymentscaling.DeploymentScalingList{},
		&cronjobsuspend.CronJobSuspend{},
		&cronjobsuspend.CronJobSuspendList{},
		&horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling{},
		&horizontalpodautoscalerscaling.HorizontalPodAutoscalerScalingList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
