package v1alpha1

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
)

const deploymentscalingsPlural = "deploymentscalings"

type DeploymentScalingInterface interface {
	List(opts metav1.ListOptions) (*deploymentscaling.DeploymentScalingList, error)
	Get(name string, options metav1.GetOptions) (*deploymentscaling.DeploymentScaling, error)
	Create(*deploymentscaling.DeploymentScaling) (*deploymentscaling.DeploymentScaling, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Update(*deploymentscaling.DeploymentScaling, metav1.UpdateOptions) (*deploymentscaling.DeploymentScaling, error)
}

type deploymentScalingClient struct {
	restClient rest.Interface
	ns         string
}

func (c *deploymentScalingClient) List(opts metav1.ListOptions) (*deploymentscaling.DeploymentScalingList, error) {
	result := deploymentscaling.DeploymentScalingList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(deploymentscalingsPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *deploymentScalingClient) Get(name string, opts metav1.GetOptions) (*deploymentscaling.DeploymentScaling, error) {
	result := deploymentscaling.DeploymentScaling{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(deploymentscalingsPlural).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *deploymentScalingClient) Update(ds *deploymentscaling.DeploymentScaling, opts metav1.UpdateOptions) (*deploymentscaling.DeploymentScaling, error) {
	result := deploymentscaling.DeploymentScaling{}
	err := c.restClient.Put().
		Namespace(c.ns).
		Resource(deploymentscalingsPlural).
		Name(ds.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ds).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *deploymentScalingClient) Create(ds *deploymentscaling.DeploymentScaling) (*deploymentscaling.DeploymentScaling, error) {
	result := deploymentscaling.DeploymentScaling{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource(deploymentscalingsPlural).
		Body(ds).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *deploymentScalingClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource(deploymentscalingsPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.Background())
}
