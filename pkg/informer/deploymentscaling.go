package informer

import (
	"context"
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
	v1alpha12 "vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
	"vandulmen.net/scheduledscale/pkg/cron"
)

func (informer *Informer) WatchDeploymentScaling() cache.Store {
	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return informer.clientSet.DeploymentScaling("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return informer.clientSet.DeploymentScaling("").Watch(lo)
			},
		},
		&v1alpha12.DeploymentScaling{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var ds = obj
				dsTyped := ds.(*v1alpha12.DeploymentScaling)
				log.Printf("%s - added scheduled deployment scaling for ", dsTyped.ObjectMeta.Namespace)

				informer.ReconcileDeploymentScaling(dsTyped)
			},
			UpdateFunc: func(old, new interface{}) {
				var ds = new
				dsTyped := ds.(*v1alpha12.DeploymentScaling)
				log.Printf("%s - reconciling updated scheduled deployment scaling for", dsTyped.ObjectMeta.Namespace)

				informer.ReconcileDeploymentScaling(dsTyped)
			},
			DeleteFunc: func(obj interface{}) {
				var ds = obj
				dsTyped := ds.(*v1alpha12.DeploymentScaling)
				log.Printf("%s - deleted scheduled deployment scaling for", dsTyped.ObjectMeta.Namespace)

			},
		},
	)

	go controller.Run(wait.NeverStop)
	return store
}

func (informer *Informer) ReconcileDeploymentScaling(ds *v1alpha12.DeploymentScaling) {
	keyForScheduler := "ds." + ds.Namespace + "." + ds.Name

	log.Printf("Reconciling for DS %s", keyForScheduler)

	if ds.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ds, finalizerName) {
			log.Println("Adding finalizer")
			controllerutil.AddFinalizer(ds, finalizerName)
			_, err := informer.clientSet.DeploymentScaling(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
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

			err = informer.DeletePodDisruptionBudgetsFor(ds)

			if err != nil {
				log.Println(err)
			}

			_, err = informer.clientSet.DeploymentScaling(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				log.Println(err)
			}
		}
		return
	}

	var groupFuncs []cron.AddFunc

	for _, scaleTo := range ds.Spec.ScaleTo {
		useThisScaleTo := scaleTo

		groupFuncs = append(groupFuncs, cron.AddFunc{
			Handler: func() {
				log.Printf("Scaling actions for %s and schedule %s", ds.Name, useThisScaleTo.At)

				listOptions := metav1.ListOptions{
					LabelSelector: labels.Set(ds.Spec.Deployment.MatchLabels).String(),
				}

				deploymentList, err := informer.coreClientSet.AppsV1().Deployments(ds.Namespace).
					List(context.TODO(), listOptions)
				if err != nil {
					log.Println(err.Error())
				} else {
					for _, deployment := range deploymentList.Items {
						useThisDeployment := deployment

						log.Printf("Updating deployment %s to %d replicas for %s in %s", useThisDeployment.Name, useThisScaleTo.Replicas, ds.Name, ds.Namespace)

						payloadBytes := informer.CreateDeploymentScalingPatch(&useThisScaleTo)

						_, err := informer.coreClientSet.
							AppsV1().Deployments(ds.Namespace).
							Patch(context.TODO(), useThisDeployment.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

						if err != nil {
							log.Printf("Patch deployment gone wrong for %s in %s", ds.Name, ds.Namespace)
							log.Println(err.Error())
							break
						}

						// do optional annotations add
						if useThisScaleTo.Annotations != nil {
							if useThisScaleTo.Annotations.Remove != nil {
								for _, anKey := range useThisScaleTo.Annotations.Remove {

									if _, ok := useThisDeployment.Annotations[anKey]; !ok {
										log.Printf("Deployment already has annotation %s for %s in %s", anKey, ds.Name, ds.Namespace)
										continue
									}

									removePayload := informer.CreateRemovePatch(anKey, "/metadata/annotations/%s")

									_, err := informer.coreClientSet.
										AppsV1().Deployments(ds.Namespace).
										Patch(context.TODO(), useThisDeployment.Name, types.JSONPatchType, removePayload, metav1.PatchOptions{})

									if err != nil {
										log.Printf("Remove annotations patch deployment gone wrong for %s in %s", ds.Name, ds.Namespace)
										log.Println(err.Error())
									}
								}
							}
						}

						informer.ReconcilePodDisruptionBudget(&useThisScaleTo, ds, &deployment)
					}
				}
			},
			Cron: scaleTo.At,
		})
	}

	err := informer.cronScheduler.ReplaceForGroup(keyForScheduler, groupFuncs)
	if err != nil {
		log.Println(err)
		return
	}
}
