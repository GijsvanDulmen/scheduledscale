package v1alpha1

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/horizontalpodautoscalerscaling"
)

const horizontalpodautoscalerscalingsPlural = "horizontalpodautoscalerscalings"

type HorizontalPodAutoscalerScalingInterface interface {
	List(opts metav1.ListOptions) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScalingList, error)
	Get(name string, options metav1.GetOptions) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, error)
	Create(*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Update(*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, metav1.UpdateOptions) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, error)
}

type hpaScalingClient struct {
	restClient rest.Interface
	ns         string
}

func (c *hpaScalingClient) List(opts metav1.ListOptions) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScalingList, error) {
	result := horizontalpodautoscalerscaling.HorizontalPodAutoscalerScalingList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(horizontalpodautoscalerscalingsPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *hpaScalingClient) Get(name string, opts metav1.GetOptions) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, error) {
	result := horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource(horizontalpodautoscalerscalingsPlural).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *hpaScalingClient) Update(ds *horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, opts metav1.UpdateOptions) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, error) {
	result := horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling{}
	err := c.restClient.Put().
		Namespace(c.ns).
		Resource(horizontalpodautoscalerscalingsPlural).
		Name(ds.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(ds).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *hpaScalingClient) Create(ds *horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling) (*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, error) {
	result := horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource(horizontalpodautoscalerscalingsPlural).
		Body(ds).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *hpaScalingClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource(horizontalpodautoscalerscalingsPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.Background())
}
