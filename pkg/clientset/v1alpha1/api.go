package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"scheduledscale/pkg/apis/scheduledscalecontroller"

	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1"
)

type V1Alpha1Interface interface {
	DeploymentScaling(namespace string) DeploymentScalingInterface
	CronJobSuspend(namespace string) CronJobSuspendInterface
}

type V1Alpha1Client struct {
	restClient rest.Interface
}

func NewForScheduleScale(c *rest.Config) (*V1Alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: scheduledscalecontroller.GroupName, Version: v1alpha1.SchemeGroupVersion.Version}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &V1Alpha1Client{restClient: client}, nil
}

func (c *V1Alpha1Client) DeploymentScaling(namespace string) DeploymentScalingInterface {
	return &deploymentScalingClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}

func (c *V1Alpha1Client) CronJobSuspend(namespace string) CronJobSuspendInterface {
	return &cronjobSuspendClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
