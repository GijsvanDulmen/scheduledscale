package informer

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/cronjobsuspend"
	"vandulmen.net/scheduledscale/pkg/cron"
)

func (informer *Informer) WatchCronJobSuspend() cache.Store {
	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return informer.clientSet.CronJobSuspend("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return informer.clientSet.CronJobSuspend("").Watch(lo)
			},
		},
		&cronjobsuspend.CronJobSuspend{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var ds = obj
				typed := ds.(*cronjobsuspend.CronJobSuspend)
				log.Printf("%s - added cronjob suspend for ", typed.ObjectMeta.Namespace)
				log.Printf("%s", typed.Spec.CronJob.MatchLabels)

				informer.ReconcileCronJobSuspend(typed)
			},
			UpdateFunc: func(old, new interface{}) {
				var ds = new
				typed := ds.(*cronjobsuspend.CronJobSuspend)
				log.Printf("%s - reconciling updated cronjob suspend for", typed.ObjectMeta.Namespace)

				informer.ReconcileCronJobSuspend(typed)
			},
			DeleteFunc: func(obj interface{}) {
				var ds = obj
				dsTyped := ds.(*cronjobsuspend.CronJobSuspend)
				log.Printf("%s - deleted scheduled deployment scaling for", dsTyped.ObjectMeta.Namespace)
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return store
}

func (informer *Informer) ReconcileCronJobSuspend(ds *cronjobsuspend.CronJobSuspend) {
	keyForScheduler := "cs." + ds.Namespace + "." + ds.Name

	log.Printf("Reconciling for CS %s", keyForScheduler)

	if ds.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ds, finalizerName) {
			log.Println("Adding finalizer")
			controllerutil.AddFinalizer(ds, finalizerName)
			_, err := informer.clientSet.CronJobSuspend(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				log.Println(err)
			}
			return
		}
	} else {
		if controllerutil.ContainsFinalizer(ds, finalizerName) {
			controllerutil.RemoveFinalizer(ds, finalizerName)

			err := informer.cronScheduler.RemoveForGroup(keyForScheduler)
			if err != nil {
				log.Println(err)
				log.Println("Could not reconcile")
				return
			}

			_, err = informer.clientSet.CronJobSuspend(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				log.Println(err)
			}
		}
		return
	}

	var groupFuncs []cron.AddFunc

	for _, stateAt := range ds.Spec.StateAt {
		useThisStateAt := stateAt

		groupFuncs = append(groupFuncs, cron.AddFunc{
			Handler: func() {
				log.Printf("Suspend actions for %s and schedule %s", ds.Name, useThisStateAt.At)

				listOptions := metav1.ListOptions{
					LabelSelector: labels.Set(ds.Spec.CronJob.MatchLabels).String(),
				}

				cronjobList, err := informer.coreClientSet.BatchV1().CronJobs(ds.Namespace).
					List(context.TODO(), listOptions)
				if err != nil {
					log.Println(err.Error())
				} else {
					log.Printf("Got %d cronjobs ", len(cronjobList.Items))
					for _, cronjob := range cronjobList.Items {
						useThisCronJob := cronjob

						log.Printf("Updating cronjob %s to %t for %s in %s", useThisCronJob.Name, useThisStateAt.Suspend, ds.Name, ds.Namespace)

						// do the standard patching
						payloadBytes := informer.CreateCronJobSuspendPatch(&useThisStateAt)

						log.Println(string(payloadBytes))

						_, err := informer.coreClientSet.
							BatchV1().CronJobs(ds.Namespace).
							Patch(context.TODO(), useThisCronJob.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

						if err != nil {
							log.Printf("Patch cronjob gone wrong for %s in %s", ds.Name, ds.Namespace)
							log.Println(err.Error())
							break
						}

						// do optional annotations add
						if useThisStateAt.Annotations != nil {
							if useThisStateAt.Annotations.Remove != nil {
								for _, anKey := range useThisStateAt.Annotations.Remove {

									fmt.Println(useThisCronJob.Spec.JobTemplate.Spec.Template.Annotations)
									if _, ok := useThisCronJob.Spec.JobTemplate.Spec.Template.Annotations[anKey]; !ok {
										log.Printf("Cronjob already has annotation %s for %s in %s", anKey, ds.Name, ds.Namespace)
										continue
									}

									removePayload := informer.CreateRemovePatch(anKey, "/spec/jobTemplate/spec/template/metadata/annotations/%s")
									log.Println(string(removePayload))

									_, err := informer.coreClientSet.
										BatchV1().CronJobs(ds.Namespace).
										Patch(context.TODO(), useThisCronJob.Name, types.JSONPatchType, removePayload, metav1.PatchOptions{})

									if err != nil {
										log.Printf("Remove annotations patch cronjob gone wrong for %s in %s", ds.Name, ds.Namespace)
										log.Println(err.Error())
									}
								}
							}
						}
					}
				}
			},
			Cron: stateAt.At,
		})
	}

	err := informer.cronScheduler.ReplaceForGroup(keyForScheduler, groupFuncs)
	if err != nil {
		log.Println(err)
		return
	}
}
