package informer

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	v1alpha12 "scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
	"scheduledscale/pkg/cron"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func LogForDeploymentScaling(ds v1alpha12.DeploymentScaling, line string, level zerolog.Level) {
	log.WithLevel(level).Msgf("ds %s/%s: %s", ds.Namespace, ds.Name, line)
}

func LogForDeploymentScalingScaleTo(ds v1alpha12.DeploymentScaling, scaleTo v1alpha12.ScaleTo, line string, level zerolog.Level) {
	LogForDeploymentScaling(ds, fmt.Sprintf("schedule %s: %s", scaleTo.At, line), level)
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
				LogForDeploymentScaling(*typed, "added", zerolog.DebugLevel)

				informer.ReconcileDeploymentScaling(typed)
			},
			UpdateFunc: func(old, new interface{}) {
				var ds = new
				typed := ds.(*v1alpha12.DeploymentScaling)
				LogForDeploymentScaling(*typed, "updated", zerolog.DebugLevel)

				informer.ReconcileDeploymentScaling(typed)
			},
			DeleteFunc: func(obj interface{}) {
				var ds = obj
				typed := ds.(*v1alpha12.DeploymentScaling)
				LogForDeploymentScaling(*typed, "deleted", zerolog.DebugLevel)
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return store
}

func (informer *Informer) ReconcileDeploymentScaling(ds *v1alpha12.DeploymentScaling) {
	keyForScheduler := "ds." + ds.Namespace + "." + ds.Name

	LogForDeploymentScaling(*ds, "reconciling", zerolog.InfoLevel)

	if ds.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ds, finalizerName) {
			LogForDeploymentScaling(*ds, "adding finalizer", zerolog.DebugLevel)

			boolTrue := true
			ds.Status.Registered = &boolTrue
			controllerutil.AddFinalizer(ds, finalizerName)
			_, err := informer.clientSet.DeploymentScaling(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				LogForDeploymentScaling(*ds, err.Error(), zerolog.ErrorLevel)
			}
			return
		}
	} else {
		if controllerutil.ContainsFinalizer(ds, finalizerName) {
			controllerutil.RemoveFinalizer(ds, finalizerName)

			LogForDeploymentScaling(*ds, "removing finalizer", zerolog.DebugLevel)

			err := informer.cronScheduler.RemoveForGroup(keyForScheduler)
			if err != nil {
				LogForDeploymentScaling(*ds, err.Error(), zerolog.ErrorLevel)
				return
			}

			err = informer.DeletePodDisruptionBudgetsFor(ds)

			if err != nil {
				LogForDeploymentScaling(*ds, "could not delete poddisruptionbudget(s)", zerolog.ErrorLevel)
				LogForDeploymentScaling(*ds, err.Error(), zerolog.ErrorLevel)
			}

			_, err = informer.clientSet.DeploymentScaling(ds.ObjectMeta.Namespace).Update(ds, metav1.UpdateOptions{})
			if err != nil {
				LogForDeploymentScaling(*ds, "could not remove finalizer", zerolog.ErrorLevel)
				LogForDeploymentScaling(*ds, err.Error(), zerolog.ErrorLevel)
			}
		}
		return
	}

	var groupFuncs []cron.AddFunc

	for _, scaleTo := range ds.Spec.ScaleTo {
		useThisScaleTo := scaleTo

		groupFuncs = append(groupFuncs, cron.AddFunc{
			Handler: func() {
				LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, "executing", zerolog.InfoLevel)

				listOptions := metav1.ListOptions{
					LabelSelector: labels.Set(ds.Spec.Deployment.MatchLabels).String(),
				}

				deploymentList, err := informer.coreClientSet.AppsV1().Deployments(ds.Namespace).
					List(context.TODO(), listOptions)
				if err != nil {
					LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, err.Error(), zerolog.ErrorLevel)
				} else {
					for _, deployment := range deploymentList.Items {
						useThisDeployment := deployment

						LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Updating deployment %s to %d", useThisDeployment.Name, useThisScaleTo.Replicas), zerolog.InfoLevel)

						payloadBytes := CreateDeploymentScalingPatch(&useThisScaleTo)

						_, err := informer.coreClientSet.
							AppsV1().Deployments(ds.Namespace).
							Patch(context.TODO(), useThisDeployment.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

						if err != nil {
							LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Patch deployment %s gone wrong", useThisDeployment.Name), zerolog.ErrorLevel)
							LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, err.Error(), zerolog.ErrorLevel)
							break
						}

						// do optional annotations add
						if useThisScaleTo.Annotations != nil {
							if useThisScaleTo.Annotations.Remove != nil {
								for _, anKey := range useThisScaleTo.Annotations.Remove {

									if _, ok := useThisDeployment.Annotations[anKey]; !ok {
										LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Deployment %s already has annotation: %s", useThisDeployment.Name, anKey), zerolog.DebugLevel)
										continue
									}

									removePayload := CreateRemovePatch(anKey, "/metadata/annotations/%s")

									_, err := informer.coreClientSet.
										AppsV1().Deployments(ds.Namespace).
										Patch(context.TODO(), useThisDeployment.Name, types.JSONPatchType, removePayload, metav1.PatchOptions{})

									if err != nil {
										LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, fmt.Sprintf("Deployment %s removal of annotations gone wrong", useThisDeployment.Name), zerolog.ErrorLevel)
										LogForDeploymentScalingScaleTo(*ds, useThisScaleTo, err.Error(), zerolog.ErrorLevel)
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
		LogForDeploymentScaling(*ds, "scheduled jobs not being replaced", zerolog.ErrorLevel)
		LogForDeploymentScaling(*ds, err.Error(), zerolog.ErrorLevel)
		return
	}
}
