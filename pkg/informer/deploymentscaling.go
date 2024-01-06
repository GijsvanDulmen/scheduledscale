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
	v1alpha12 "scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
	"scheduledscale/pkg/cron"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func LogForDeploymentScaling(ds v1alpha12.DeploymentScaling, line string) {
	log.Printf("ds %s/%s: %s", ds.Namespace, ds.Name, line)
}

func LogForDeploymentScalingScaleTo(ds v1alpha12.DeploymentScaling, scaleTo v1alpha12.ScaleTo, line string) {
	LogForDeploymentScaling(ds, fmt.Sprintf("schedule %s: %s", scaleTo.At, line))
}

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
				typed := ds.(*v1alpha12.DeploymentScaling)
				LogForDeploymentScaling(*typed, "added")

				informer.ReconcileDeploymentScaling(typed)
			},
			UpdateFunc: func(old, new interface{}) {
				var ds = new
				typed := ds.(*v1alpha12.DeploymentScaling)
				LogForDeploymentScaling(*typed, "updated")

				informer.ReconcileDeploymentScaling(typed)
			},
			DeleteFunc: func(obj interface{}) {
				var ds = obj
				typed := ds.(*v1alpha12.DeploymentScaling)
				LogForDeploymentScaling(*typed, "deleted")
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return store
}

func (informer *Informer) ReconcileDeploymentScaling(ds *v1alpha12.DeploymentScaling) {
	keyForScheduler := "ds." + ds.Namespace + "." + ds.Name

	LogForDeploymentScaling(*ds, "reconciling")

	if ds.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ds, finalizerName) {
			LogForDeploymentScaling(*ds, "adding finalizer")

			controllerutil.AddFinalizer(ds, finalizerName)
			_, err := informer.clientSet.DeploymentScaling(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				LogForDeploymentScaling(*ds, err.Error())
			}
			return
		}
	} else {
		if controllerutil.ContainsFinalizer(ds, finalizerName) {
			controllerutil.RemoveFinalizer(ds, finalizerName)

			LogForDeploymentScaling(*ds, "removing finalizer")

			err := informer.cronScheduler.RemoveForGroup(keyForScheduler)
			if err != nil {
				LogForDeploymentScaling(*ds, err.Error())
				return
			}

			err = informer.DeletePodDisruptionBudgetsFor(ds)

			if err != nil {
				LogForDeploymentScaling(*ds, "could not delete poddisruptionbudget(s)")
				LogForDeploymentScaling(*ds, err.Error())
			}

			_, err = informer.clientSet.DeploymentScaling(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				LogForDeploymentScaling(*ds, "could not remove finalizer")
				LogForDeploymentScaling(*ds, err.Error())
			}
		}
		return
	}

	var groupFuncs []cron.AddFunc

	for _, scaleTo := range ds.Spec.ScaleTo {
		useThisScaleTo := scaleTo

		groupFuncs = append(groupFuncs, cron.AddFunc{
			Handler: func() {
				LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, "executing")

				listOptions := metav1.ListOptions{
					LabelSelector: labels.Set(ds.Spec.Deployment.MatchLabels).String(),
				}

				deploymentList, err := informer.coreClientSet.AppsV1().Deployments(ds.Namespace).
					List(context.TODO(), listOptions)
				if err != nil {
					LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, err.Error())
				} else {
					for _, deployment := range deploymentList.Items {
						useThisDeployment := deployment

						LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Updating deployment %s to %d", useThisDeployment.Name, useThisScaleTo.Replicas))

						payloadBytes := CreateDeploymentScalingPatch(&useThisScaleTo)

						_, err := informer.coreClientSet.
							AppsV1().Deployments(ds.Namespace).
							Patch(context.TODO(), useThisDeployment.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

						if err != nil {
							LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Patch deployment %s gone wrong", useThisDeployment.Name))
							LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, err.Error())
							break
						}

						// do optional annotations add
						if useThisScaleTo.Annotations != nil {
							if useThisScaleTo.Annotations.Remove != nil {
								for _, anKey := range useThisScaleTo.Annotations.Remove {

									if _, ok := useThisDeployment.Annotations[anKey]; !ok {
										LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Deployment %s already has annotation: %s", useThisDeployment.Name, anKey))
										continue
									}

									removePayload := CreateRemovePatch(anKey, "/metadata/annotations/%s")

									_, err := informer.coreClientSet.
										AppsV1().Deployments(ds.Namespace).
										Patch(context.TODO(), useThisDeployment.Name, types.JSONPatchType, removePayload, metav1.PatchOptions{})

									if err != nil {
										LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Deployment %s removal of annotations gone wrong", useThisDeployment.Name))
										LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, err.Error())
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
		LogForDeploymentScaling(*ds, "scheduled jobs not being replaced")
		LogForDeploymentScaling(*ds, err.Error())
		return
	}
}
