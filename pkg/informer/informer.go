package informer

import (
	"vandulmen.net/scheduledscale/pkg/clientset/v1alpha1"
	"vandulmen.net/scheduledscale/pkg/cron"

	"k8s.io/client-go/kubernetes"
)

type Informer struct {
	clientSet     v1alpha1.V1Alpha1Interface
	coreClientSet *kubernetes.Clientset
	cronScheduler *cron.CronScheduler
}

const finalizerName = "vandulmen.net/scheduledscale"

func NewInformer(clientSet v1alpha1.V1Alpha1Interface, coreClientSet *kubernetes.Clientset, cronScheduler *cron.CronScheduler) (informer *Informer) {
	return &Informer{
		clientSet:     clientSet,
		coreClientSet: coreClientSet,
		cronScheduler: cronScheduler,
	}
}
