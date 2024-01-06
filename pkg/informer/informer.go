package informer

import (
	"scheduledscale/pkg/clientset/v1alpha1"
	"scheduledscale/pkg/cron"

	"k8s.io/client-go/kubernetes"
)

type Informer struct {
	clientSet     v1alpha1.V1Alpha1Interface
	coreClientSet *kubernetes.Clientset
	cronScheduler *cron.CronScheduler
}

const finalizerName = "scheduledscale.io"

func NewInformer(clientSet v1alpha1.V1Alpha1Interface, coreClientSet *kubernetes.Clientset, cronScheduler *cron.CronScheduler) (informer *Informer) {
	return &Informer{
		clientSet:     clientSet,
		coreClientSet: coreClientSet,
		cronScheduler: cronScheduler,
	}
}
