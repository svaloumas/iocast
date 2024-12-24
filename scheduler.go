package iocast

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrEitherWhenOrInterval = errors.New("cannot use when and interval together")
)

// ScheduleDB represents the storage used for the Schedule entities.
type ScheduleDB interface {
	Store(string, *Schedule) error
	FetchDue(time.Time) ([]*Schedule, error)
}

type scheduleMemDB struct {
	db *sync.Map
}

// Schedule represents a task's schedule.
type Schedule struct {
	task     Task
	Days     []time.Weekday
	When     time.Time
	Interval time.Duration
	lastRun  time.Time
	nextRun  time.Time
	once     bool
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

// ScheduleRun schedules an once-off run for the task.
func (s *scheduler) ScheduleRun(t Task, when time.Time) error {
	today := time.Now().Weekday()
	schedule := &Schedule{
		task: t,
		Days: []time.Weekday{today},
		When: when,
		once: true,
	}
	return s.db.Store(t.Id(), schedule)
}

// Schedule schedules a task for periodic execution.
// NOTE: Interval min value is set to 5 mins.
func (s *scheduler) Schedule(t Task, given Schedule) error {
	if err := s.validate(given); err != nil {
		return err
	}
	schedule := &Schedule{
		task:     t,
		Days:     given.Days,
		When:     given.When,
		Interval: given.Interval,
		once:     false,
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
					} else {
						schedule.lastRun = now
						schedule.nextRun = now.Add(schedule.Interval)
						if schedule.once {
							// remove scheduled days
							schedule.Days = nil
						}
						if err := s.db.Store(schedule.task.Id(), schedule); err != nil {
							log.Printf("failed to update schedule for task with id: %s", schedule.task.Id())
						}
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

func (s *scheduler) validate(schedule Schedule) error {
	if !schedule.When.IsZero() && schedule.Interval.Nanoseconds() != 0 {
		return ErrEitherWhenOrInterval
	}
	if schedule.Interval.Nanoseconds() != 0 {
		minutes := schedule.Interval.Minutes()
		if minutes < 5 {
			// minimum interval value is 5 minutes
			schedule.Interval = 5 * time.Minute
		}
	}
	return nil
}

// Store stores the schedule in the database.
func (m *scheduleMemDB) Store(id string, s *Schedule) error {
	m.db.Store(id, s)
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

		today := now.Weekday()
		if contains(schedule.Days, today) {
			// handle whens
			if !schedule.When.IsZero() {
				runTime := time.Date(
					now.Year(), now.Month(), now.Day(),
					schedule.When.Hour(), schedule.When.Minute(),
					schedule.When.Second(), 0, now.Location(),
				)

				if now.After(runTime) && (schedule.lastRun.IsZero() || schedule.lastRun.Before(runTime)) {
					dueSchedules = append(dueSchedules, schedule)
				}
			} else {
				// handle intervals
				if now.After(schedule.nextRun) && (schedule.lastRun.IsZero() || schedule.lastRun.Before(schedule.nextRun)) {
					dueSchedules = append(dueSchedules, schedule)
				}
			}
		}
		return true
	})
	return dueSchedules, nil
}

func contains(days []time.Weekday, day time.Weekday) bool {
	for _, v := range days {
		if v == day {
			return true
		}
	}
	return false
}
