package v1alpha1

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/cronjobsuspend"
)

const cronjobsuspendPlural = "cronjobsuspends"

type CronJobSuspendInterface interface {
	List(opts metav1.ListOptions) (*cronjobsuspend.CronJobSuspendList, error)
	Get(name string, options metav1.GetOptions) (*cronjobsuspend.CronJobSuspend, error)
	Create(*cronjobsuspend.CronJobSuspend) (*cronjobsuspend.CronJobSuspend, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Update(*cronjobsuspend.CronJobSuspend, metav1.UpdateOptions) (*cronjobsuspend.CronJobSuspend, error)
}

type cronjobSuspendClient struct {
	restClient rest.Interface
	ns         string
}

func (c *cronjobSuspendClient) List(opts metav1.ListOptions) (*cronjobsuspend.CronJobSuspendList, error) {
	result := cronjobsuspend.CronJobSuspendList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(cronjobsuspendPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *cronjobSuspendClient) Get(name string, opts metav1.GetOptions) (*cronjobsuspend.CronJobSuspend, error) {
	result := cronjobsuspend.CronJobSuspend{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(cronjobsuspendPlural).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *cronjobSuspendClient) Update(ds *cronjobsuspend.CronJobSuspend, opts metav1.UpdateOptions) (*cronjobsuspend.CronJobSuspend, error) {
	result := cronjobsuspend.CronJobSuspend{}
	err := c.restClient.Put().
		Namespace(c.ns).
		Resource(cronjobsuspendPlural).
		Name(ds.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ds).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *cronjobSuspendClient) Create(ds *cronjobsuspend.CronJobSuspend) (*cronjobsuspend.CronJobSuspend, error) {
	result := cronjobsuspend.CronJobSuspend{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource(cronjobsuspendPlural).
		Body(ds).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *cronjobSuspendClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource(cronjobsuspendPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.Background())
}
