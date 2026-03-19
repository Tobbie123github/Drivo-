// internal/workers/email_worker.go

package workers

import (
	"drivo/internal/jobs"
	"log"
	"time"
)

var EmailQueue = make(chan jobs.EmailJob, 100)

func StartEmailWorkers(mailer jobs.Mailer) {
	workerCount := 5
	for i := 0; i < workerCount; i++ {
		go startEmailWorker(mailer)
	}
}

func startEmailWorker(mailer jobs.Mailer) {
	for job := range EmailQueue {
		processEmail(mailer, job)
	}
}

func processEmail(mailer jobs.Mailer, job jobs.EmailJob) {
	log.Printf("Processing %s email for: %s\n", job.Type, job.To)

	var err error

	for i := 0; i < 3; i++ {
		err = dispatch(mailer, job)
		if err == nil {
			log.Printf("Email sent to %s\n", job.To)
			return
		}
		log.Printf("Retry %d for %s: %v\n", i+1, job.To, err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("Failed to send email to %s after 3 retries: %v\n", job.To, err)
}

func dispatch(mailer jobs.Mailer, job jobs.EmailJob) error {
	switch job.Type {
	case jobs.EmailTypeOTP:
		return mailer.SendOTPEmail(job.To, job.Name, job.OTP)
	case jobs.EmailTypeWelcome:
		return mailer.SendWelcomeEmail(job.To, job.Name)
	case jobs.EmailTypeDriverWelcome:
		return mailer.SendDriverWelcomeEmail(job.To, job.Name)
	case jobs.EmailTypeRideConfirmation:
		return mailer.SendRideConfirmationEmail(job.To, job.RideConfirmationData)
	case jobs.EmailTypeRideCompleted: 
		return mailer.SendRideCompletedEmail(job.To, job.RideCompletedData)
	default:
		log.Printf("unknown email type: %s\n", job.Type)
		return nil
	}
}
