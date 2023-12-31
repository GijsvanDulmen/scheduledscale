package cron

import (
	"github.com/go-co-op/gocron/v2"
	"log"
	"sync"
)

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

	return cs.RemoveForGroup(groupKey)
}

func (cs *CronScheduler) removeForGroup(groupKey string) error {
	if jobs, ok := cs.jobDefinitions[groupKey]; ok {
		for _, job := range jobs {
			log.Printf("Removing job %s", job.ID())
			err := cs.scheduler.RemoveJob(job.ID())
			if err != nil {
				return err
			}
		}
	}

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
		job, err := cs.scheduler.NewJob(gocron.CronJob(handler.Cron, true), gocron.NewTask(handler.Handler))
		if err != nil {
			return err
		}
		cs.jobDefinitions[groupKey] = append(cs.jobDefinitions[groupKey], job)

		log.Printf("Added job %s", job.ID())
		log.Printf("Jobs for %s are %d", groupKey, len(cs.jobDefinitions[groupKey]))
	}

	return nil
}
