package cron

import (
	"github.com/go-co-op/gocron/v2"
	logger "scheduledscale/pkg/log"
	"sync"
)

var log = logger.Logger()

type CronScheduler struct {
	scheduler      gocron.Scheduler
	jobDefinitions map[string][]gocron.Job
	mu             sync.Mutex
}

func NewCronScheduler() (*CronScheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	scheduler.Start()

	return &CronScheduler{
		scheduler:      scheduler,
		jobDefinitions: map[string][]gocron.Job{},
	}, nil
}

func (cs *CronScheduler) RemoveForGroup(groupKey string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return cs.removeForGroup(groupKey)
}

func (cs *CronScheduler) GetCount() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return len(cs.scheduler.Jobs())
}

func (cs *CronScheduler) removeForGroup(groupKey string) error {
	cs.scheduler.RemoveByTags(groupKey)
	cs.jobDefinitions[groupKey] = []gocron.Job{}
	return nil
}

type AddFunc struct {
	Handler func()
	Cron    string
}

func (cs *CronScheduler) ReplaceForGroup(groupKey string, handlers []AddFunc) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	err := cs.removeForGroup(groupKey)
	if err != nil {
		return err
	}

	for _, handler := range handlers {
		job, err := cs.scheduler.NewJob(gocron.CronJob(handler.Cron, true), gocron.NewTask(handler.Handler), gocron.WithTags(groupKey))
		if err != nil {
			return err
		}
		cs.jobDefinitions[groupKey] = append(cs.jobDefinitions[groupKey], job)

		log.Debug().Msgf("Added job %s", job.ID())
		log.Debug().Msgf("Jobs for %s are %d", groupKey, len(cs.jobDefinitions[groupKey]))
	}

	return nil
}
