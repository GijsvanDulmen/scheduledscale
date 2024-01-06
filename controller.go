package main

import (
	"flag"
	"log"
	"os"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1"
	v1alpha12 "scheduledscale/pkg/clientset/v1alpha1"
	"scheduledscale/pkg/cron"
	"time"

	"scheduledscale/pkg/informer"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to Kubernetes config file")
	flag.Parse()
}

func main() {

	var restConfig *rest.Config
	var err error

	if kubeconfig == "" {
		log.Printf("using in-cluster configuration")
		restConfig, err = rest.InClusterConfig()
	} else {
		log.Printf("using configuration from '%s'", kubeconfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err)
	}

	_ = v1alpha1.AddToScheme(scheme.Scheme)

	clientSet, err := v1alpha12.NewForScheduleScale(restConfig)
	if err != nil {
		panic(err)
	}

	coreClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Println(err)
		os.Exit(3)
		return
	}

	scheduler, err := cron.NewCronScheduler()

	if err != nil {
		log.Println(err)
		os.Exit(3)
		return
	}

	newInformer := informer.NewInformer(clientSet, coreClientSet, scheduler)
	deploymentScalingStore := newInformer.WatchDeploymentScaling()
	cronjobsuspendStore := newInformer.WatchCronJobSuspend()

	for {
		dsFromStore := deploymentScalingStore.List()
		log.Printf("number of ds watching: %d\n", len(dsFromStore))

		csFromStore := cronjobsuspendStore.List()
		log.Printf("number of cs watching: %d\n", len(csFromStore))

		log.Printf("number of cronjobs: %d\n", scheduler.GetCount())

		time.Sleep(10 * time.Second)
	}
}
