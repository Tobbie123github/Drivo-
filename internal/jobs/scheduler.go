package jobs

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron             *cron.Cron
	runScheduled     func()
	entryID          cron.EntryID
	recurringRideJob *RecurringRideJob
}

func NewScheduler(runScheduled func(), recurringRideJob *RecurringRideJob) *Scheduler {
	loc, _ := time.LoadLocation("Africa/Lagos")
	return &Scheduler{
		runScheduled:     runScheduled,
		cron:             cron.New(cron.WithSeconds(), cron.WithLocation(loc)),
		recurringRideJob: recurringRideJob,
	}
}

func (s *Scheduler) Start() {
	var err error
	s.entryID, err = s.cron.AddFunc("@every 1m", func() {
		fmt.Printf(" Running scheduled task at %s\n", time.Now().Format(time.RFC3339))
		s.runScheduled()
	})

	s.cron.AddFunc("0 0 0 * * *", func() {
		fmt.Printf("[Scheduler] booking recurring rides for tomorrow at %s\n",
			time.Now().Format("2006-01-02 15:04:05"))
		s.recurringRideJob.Run()
	})

	if err != nil {
		fmt.Printf("Error adding cron job: %v\n", err)
		return
	}

	s.cron.Start()
	fmt.Println("Scheduler Started")
}

func (s *Scheduler) Stop() {
	s.cron.Remove(s.entryID)
	s.cron.Stop()
	fmt.Println("Scheduler Stopped")
}
