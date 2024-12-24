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

// ScheduleDB represents the storage used for the Schedule entities.
type ScheduleDB interface {
	Store(string, *Schedule) error
	FetchDue(time.Time) ([]*Schedule, error)
	Delete(string) error
}

type scheduleMemDB struct {
	db *sync.Map
}

// Schedule is a task's schedule.
type Schedule struct {
	task  Task
	RunAt time.Time
}

type scheduler struct {
	db              ScheduleDB
	wp              *workerpool
	pollingInterval time.Duration
	done            chan struct{}
}

// NewScheduler creates and returns a new scheduler instance.
func NewScheduler(db ScheduleDB, wp *workerpool, pollingInterval time.Duration) *scheduler {
	return &scheduler{
		db:              db,
		wp:              wp,
		pollingInterval: pollingInterval,
		done:            make(chan struct{}),
	}
}

// NewScheduleMemDB creates and returns a new scheduleMemDB instance.
func NewScheduleMemDB(db *sync.Map) *scheduleMemDB {
	return &scheduleMemDB{
		db: db,
	}
}

// ScheduleRun schedules a run for the task.
func (s *scheduler) Schedule(t Task, runAt time.Time) error {
	if err := s.validate(runAt); err != nil {
		return err
	}
	schedule := &Schedule{
		task:  t,
		RunAt: runAt,
	}
	return s.db.Store(t.Id(), schedule)
}

// Dispatch polls the databases for any due schedules and enqueues their tasks for execution.
func (s *scheduler) Dispatch() {
	ticker := time.NewTicker(s.pollingInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				schedules, err := s.db.FetchDue(now)
				if err != nil {
					log.Printf("failed to fetch due schedules: %v", err)
					continue
				}

				for _, schedule := range schedules {
					ok := s.wp.Enqueue(schedule.task)
					if !ok {
						log.Printf("failed to enqueue task with id: %s", schedule.task.Id())
					}
					err := s.db.Delete(schedule.task.Id())
					if err != nil {
						log.Printf("failed to delete due schedule: %v", err)
						continue
					}
				}
			case <-s.done:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the scheduler.
func (s *scheduler) Stop() {
	close(s.done)
}

func (s *scheduler) validate(runAt time.Time) error {
	if runAt.Before(time.Now()) {
		return ErrScheduledRunInThePast
	}
	return nil
}

// Store stores the schedule in the database.
func (m *scheduleMemDB) Store(id string, s *Schedule) error {
	m.db.Store(id, s)
	return nil
}

// Delete removes a schedule from the database.
func (m *scheduleMemDB) Delete(id string) error {
	m.db.Delete(id)
	return nil
}

// FetchDue fetches the due schedules from the database.
func (m *scheduleMemDB) FetchDue(now time.Time) ([]*Schedule, error) {
	var dueSchedules []*Schedule
	m.db.Range(func(key, value any) bool {
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
