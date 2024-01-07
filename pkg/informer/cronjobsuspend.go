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
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/cronjobsuspend"
	"scheduledscale/pkg/cron"
	logger "scheduledscale/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var log = logger.Logger()

func LogForCronJobSuspend(suspend cronjobsuspend.CronJobSuspend, line string, level zerolog.Level) {
	log.WithLevel(level).Msgf("cs %s/%s: %s", suspend.Namespace, suspend.Name, line)
}

func LogForCronJobSuspendState(suspend cronjobsuspend.CronJobSuspend, stateAt cronjobsuspend.StateAt, line string, level zerolog.Level) {
	LogForCronJobSuspend(suspend, fmt.Sprintf("schedule %s: %s", stateAt.At, line), level)
}

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
				var cs = obj
				typed := cs.(*cronjobsuspend.CronJobSuspend)
				LogForCronJobSuspend(*typed, "added", zerolog.DebugLevel)
				informer.ReconcileCronJobSuspend(typed)
			},
			UpdateFunc: func(old, new interface{}) {
				var cs = new
				typed := cs.(*cronjobsuspend.CronJobSuspend)
				LogForCronJobSuspend(*typed, "updated", zerolog.DebugLevel)

				informer.ReconcileCronJobSuspend(typed)
			},
			DeleteFunc: func(obj interface{}) {
				var cs = obj
				typed := cs.(*cronjobsuspend.CronJobSuspend)
				LogForCronJobSuspend(*typed, "deleted", zerolog.DebugLevel)
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return store
}

func (informer *Informer) ReconcileCronJobSuspend(cs *cronjobsuspend.CronJobSuspend) {
	keyForScheduler := "cs." + cs.Namespace + "." + cs.Name

	LogForCronJobSuspend(*cs, "reconciling", zerolog.InfoLevel)

	if cs.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(cs, finalizerName) {
			LogForCronJobSuspend(*cs, "adding finalizer", zerolog.DebugLevel)
			controllerutil.AddFinalizer(cs, finalizerName)
			_, err := informer.clientSet.CronJobSuspend(cs.ObjectMeta.Namespace).Update(cs, metav1.UpdateOptions{})
			if err != nil {
				LogForCronJobSuspend(*cs, err.Error(), zerolog.ErrorLevel)
			}
			return
		}
	} else {
		if controllerutil.ContainsFinalizer(cs, finalizerName) {
			controllerutil.RemoveFinalizer(cs, finalizerName)

			LogForCronJobSuspend(*cs, "removing finalizer", zerolog.DebugLevel)

			err := informer.cronScheduler.RemoveForGroup(keyForScheduler)
			if err != nil {
				LogForCronJobSuspend(*cs, err.Error(), zerolog.ErrorLevel)
				return
			}

			_, err = informer.clientSet.CronJobSuspend(cs.ObjectMeta.Namespace).Update(cs, metav1.UpdateOptions{})
			if err != nil {
				LogForCronJobSuspend(*cs, "could not remove finalizer", zerolog.ErrorLevel)
				LogForCronJobSuspend(*cs, err.Error(), zerolog.ErrorLevel)
			}
		}
		return
	}

	var groupFuncs []cron.AddFunc

	for _, stateAt := range cs.Spec.StateAt {
		useThisStateAt := stateAt

		groupFuncs = append(groupFuncs, cron.AddFunc{
			Handler: func() {
				LogForCronJobSuspendState(*cs, useThisStateAt, "executing", zerolog.InfoLevel)

				listOptions := metav1.ListOptions{
					LabelSelector: labels.Set(cs.Spec.CronJob.MatchLabels).String(),
				}

				cronjobList, err := informer.coreClientSet.BatchV1().CronJobs(cs.Namespace).
					List(context.TODO(), listOptions)
				if err != nil {
					LogForCronJobSuspendState(*cs, useThisStateAt, err.Error(), zerolog.ErrorLevel)
				} else {
					LogForCronJobSuspendState(*cs, useThisStateAt, fmt.Sprintf("got %d cronjobs", len(cronjobList.Items)), zerolog.InfoLevel)

					for _, cronjob := range cronjobList.Items {
						useThisCronJob := cronjob

						LogForCronJobSuspendState(*cs, useThisStateAt, fmt.Sprintf("Updating cronjob %s to %t", useThisCronJob.Name, useThisStateAt.Suspend), zerolog.InfoLevel)

						// do the standard patching
						payloadBytes := CreateCronJobSuspendPatch(&useThisStateAt)

						_, err := informer.coreClientSet.
							BatchV1().CronJobs(cs.Namespace).
							Patch(context.TODO(), useThisCronJob.Name, types.MergePatchType, payloadBytes, metav1.PatchOptions{})

						if err != nil {
							LogForCronJobSuspendState(*cs, useThisStateAt, fmt.Sprintf("Patch cronjob %s gone wrong", useThisCronJob.Name), zerolog.ErrorLevel)
							LogForCronJobSuspendState(*cs, useThisStateAt, err.Error(), zerolog.ErrorLevel)
							break
						}

						// do optional annotations add
						if useThisStateAt.Annotations != nil {
							if useThisStateAt.Annotations.Remove != nil {
								for _, anKey := range useThisStateAt.Annotations.Remove {

									if _, ok := useThisCronJob.Spec.JobTemplate.Spec.Template.Annotations[anKey]; !ok {
										LogForCronJobSuspendState(*cs, useThisStateAt, fmt.Sprintf("Cronjob %s already has annotation: %s", useThisCronJob.Name, anKey), zerolog.DebugLevel)
										continue
									}

									removePayload := CreateRemovePatch(anKey, "/spec/jobTemplate/spec/template/metadata/annotations/%s")

									_, err := informer.coreClientSet.
										BatchV1().CronJobs(cs.Namespace).
										Patch(context.TODO(), useThisCronJob.Name, types.JSONPatchType, removePayload, metav1.PatchOptions{})

									if err != nil {
										LogForCronJobSuspendState(*cs, useThisStateAt, fmt.Sprintf("Cronjob %s removal of annotations gone wrong", useThisCronJob.Name), zerolog.ErrorLevel)
										LogForCronJobSuspendState(*cs, useThisStateAt, err.Error(), zerolog.ErrorLevel)
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
		LogForCronJobSuspend(*cs, "scheduled jobs not being replaced", zerolog.ErrorLevel)
		LogForCronJobSuspend(*cs, err.Error(), zerolog.ErrorLevel)
		return
	}
}
