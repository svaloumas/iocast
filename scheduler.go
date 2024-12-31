package iocast

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrScheduledRunInThePast = errors.New("cannot schedule run in the past: when < now")
)

type ScheduleDB struct {
	db *sync.Map
}

// Schedule is a task's schedule.
type Schedule struct {
	job   Job
	RunAt time.Time
}

type Scheduler struct {
	db              *ScheduleDB
	wp              *WorkerPool
	pollingInterval time.Duration
	done            chan struct{}
}

// NewScheduler creates and returns a new scheduler instance.
func NewScheduler(wp *WorkerPool, pollingInterval time.Duration) *Scheduler {
	return &Scheduler{
		db: &ScheduleDB{
			db: &sync.Map{},
		},
		wp:              wp,
		pollingInterval: pollingInterval,
		done:            make(chan struct{}),
	}
}

// ScheduleRun schedules a run for the task.
func (s *Scheduler) Schedule(j Job, runAt time.Time) error {
	if err := s.validate(runAt); err != nil {
		return err
	}
	schedule := &Schedule{
		job:   j,
		RunAt: runAt,
	}
	return s.db.Store(j.ID(), schedule)
}

// Dispatch polls the databases for any due schedules and enqueues their tasks for execution.
func (s *Scheduler) Dispatch() {
	ticker := time.NewTicker(s.pollingInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.dispatchDueTasks()
			case <-s.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) dispatchDueTasks() {
	schedules, err := s.db.FetchDue(time.Now())
	if err != nil {
		log.Printf("failed to fetch due schedules: %v", err)
		return
	}

	for _, schedule := range schedules {
		ok := s.wp.Enqueue(schedule.job)
		if !ok {
			log.Printf("failed to enqueue task with id: %s", schedule.job.ID())
			return
		}
		err = s.db.Delete(schedule.job.ID())
		if err != nil {
			log.Printf("failed to delete due schedule: %v", err)
			return
		}
	}
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	close(s.done)
}

func (s *Scheduler) validate(runAt time.Time) error {
	if runAt.Before(time.Now()) {
		return ErrScheduledRunInThePast
	}
	return nil
}

// Store stores the schedule in the database.
func (m *ScheduleDB) Store(id string, s *Schedule) error {
	m.db.Store(id, s)
	return nil
}

// Delete removes a schedule from the database.
func (m *ScheduleDB) Delete(id string) error {
	m.db.Delete(id)
	return nil
}

// FetchDue fetches the due schedules from the database.
func (m *ScheduleDB) FetchDue(now time.Time) ([]*Schedule, error) {
	var dueSchedules []*Schedule
	m.db.Range(func(_, value any) bool {
		schedule, ok := value.(*Schedule)
		if !ok {
			return true // skip
		}

		runTime := time.Date(
			now.Year(), now.Month(), now.Day(),
			schedule.RunAt.Hour(), schedule.RunAt.Minute(),
			schedule.RunAt.Second(), 0, now.Location(),
		)

		if now.After(runTime) {
			dueSchedules = append(dueSchedules, schedule)
		}
		return true
	})
	return dueSchedules, nil
}
