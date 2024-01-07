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
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/horizontalpodautoscalerscaling"
	"scheduledscale/pkg/cron"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func LogForHpaScaling(hs horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, line string, level zerolog.Level) {
	log.WithLevel(level).Msgf("hs %s/%s: %s", hs.Namespace, hs.Name, line)
}

func LogForHpaScalingScaleTo(hs horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling, scaleTo horizontalpodautoscalerscaling.ScaleTo, line string, level zerolog.Level) {
	LogForHpaScaling(hs, fmt.Sprintf("schedule %s: %s", scaleTo.At, line), level)
}

func (informer *Informer) WatchHpaScaling() cache.Store {
	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return informer.clientSet.HorizontalPodAutoscalerScaling("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return informer.clientSet.HorizontalPodAutoscalerScaling("").Watch(lo)
			},
		},
		&horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var ds = obj
				typed := ds.(*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling)
				LogForHpaScaling(*typed, "added", zerolog.DebugLevel)

				informer.ReconcileHpaScaling(typed)
			},
			UpdateFunc: func(old, new interface{}) {
				var ds = new
				typed := ds.(*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling)
				LogForHpaScaling(*typed, "updated", zerolog.DebugLevel)

				informer.ReconcileHpaScaling(typed)
			},
			DeleteFunc: func(obj interface{}) {
				var ds = obj
				typed := ds.(*horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling)
				LogForHpaScaling(*typed, "deleted", zerolog.DebugLevel)
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return store
}

func (informer *Informer) ReconcileHpaScaling(hs *horizontalpodautoscalerscaling.HorizontalPodAutoscalerScaling) {
	keyForScheduler := "hs." + hs.Namespace + "." + hs.Name

	LogForHpaScaling(*hs, "reconciling", zerolog.InfoLevel)

	if hs.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(hs, finalizerName) {
			LogForHpaScaling(*hs, "adding finalizer", zerolog.DebugLevel)

			boolTrue := true
			hs.Status.Registered = &boolTrue
			controllerutil.AddFinalizer(hs, finalizerName)
			_, err := informer.clientSet.HorizontalPodAutoscalerScaling(hs.ObjectMeta.Namespace).Update(hs, metav1.UpdateOptions{})
			if err != nil {
				LogForHpaScaling(*hs, err.Error(), zerolog.ErrorLevel)
			}
			return
		}
	} else {
		if controllerutil.ContainsFinalizer(hs, finalizerName) {
			controllerutil.RemoveFinalizer(hs, finalizerName)

			LogForHpaScaling(*hs, "removing finalizer", zerolog.DebugLevel)

			err := informer.cronScheduler.RemoveForGroup(keyForScheduler)

			if err != nil {
				LogForHpaScaling(*hs, "error removing finalizer", zerolog.DebugLevel)
				LogForHpaScaling(*hs, err.Error(), zerolog.ErrorLevel)
				return
			}

			_, err = informer.clientSet.HorizontalPodAutoscalerScaling(hs.ObjectMeta.Namespace).Update(hs, metav1.UpdateOptions{})
			if err != nil {
				LogForHpaScaling(*hs, "could not remove finalizer", zerolog.ErrorLevel)
				LogForHpaScaling(*hs, err.Error(), zerolog.ErrorLevel)
			}
			LogForHpaScaling(*hs, "removed finalizer", zerolog.DebugLevel)
		}
		return
	}

	var groupFuncs []cron.AddFunc

	for _, scaleTo := range hs.Spec.ScaleTo {
		useThisScaleTo := scaleTo

		groupFuncs = append(groupFuncs, cron.AddFunc{
			Handler: func() {
				LogForHpaScalingScaleTo(*hs, useThisScaleTo, "executing", zerolog.InfoLevel)

				listOptions := metav1.ListOptions{
					LabelSelector: labels.Set(hs.Spec.HorizontalPodAutoscaler.MatchLabels).String(),
				}

				hpaList, err := informer.coreClientSet.AutoscalingV2().HorizontalPodAutoscalers(hs.Namespace).
					List(context.TODO(), listOptions)
				if err != nil {
					LogForHpaScalingScaleTo(*hs, useThisScaleTo, err.Error(), zerolog.ErrorLevel)
				} else {
					for _, hpa := range hpaList.Items {
						useThisHpa := hpa

						LogForHpaScalingScaleTo(*hs, useThisScaleTo, fmt.Sprintf("Updating hpa %s to min %d and max %d", useThisHpa.Name, useThisScaleTo.MinReplicas, useThisScaleTo.MaxReplicas), zerolog.InfoLevel)

						payloadBytes := CreateHpaPatch(&useThisScaleTo)

						_, err := informer.coreClientSet.
							AutoscalingV2().HorizontalPodAutoscalers(hs.Namespace).
							Patch(context.TODO(), useThisHpa.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

						if err != nil {
							LogForHpaScalingScaleTo(*hs, useThisScaleTo, fmt.Sprintf("Patch hpa %s gone wrong", useThisHpa.Name), zerolog.ErrorLevel)
							LogForHpaScalingScaleTo(*hs, useThisScaleTo, err.Error(), zerolog.ErrorLevel)
							break
						}

						// do optional annotations add
						if useThisScaleTo.Annotations != nil {
							if useThisScaleTo.Annotations.Remove != nil {
								for _, anKey := range useThisScaleTo.Annotations.Remove {

									if _, ok := useThisHpa.Annotations[anKey]; !ok {
										LogForHpaScalingScaleTo(*hs, useThisScaleTo, fmt.Sprintf("HPA %s already has annotation: %s", useThisHpa.Name, anKey), zerolog.DebugLevel)
										continue
									}

									removePayload := CreateRemovePatch(anKey, "/metadata/annotations/%s")

									_, err := informer.coreClientSet.
										AutoscalingV2().HorizontalPodAutoscalers(hs.Namespace).
										Patch(context.TODO(), useThisHpa.Name, types.JSONPatchType, removePayload, metav1.PatchOptions{})

									if err != nil {
										LogForHpaScalingScaleTo(*hs, useThisScaleTo, fmt.Sprintf("HPA %s removal of annotations gone wrong", useThisHpa.Name), zerolog.ErrorLevel)
										LogForHpaScalingScaleTo(*hs, useThisScaleTo, err.Error(), zerolog.ErrorLevel)
									}
								}
							}
						}
					}
				}
			},
			Cron: scaleTo.At,
		})
	}

	err := informer.cronScheduler.ReplaceForGroup(keyForScheduler, groupFuncs)
	if err != nil {
		LogForHpaScaling(*hs, "scheduled jobs not being replaced", zerolog.ErrorLevel)
		LogForHpaScaling(*hs, err.Error(), zerolog.ErrorLevel)
		return
	}
}
