package services

import (
	"encoder/domain"
	"encoder/infra/utils"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"os"
	"time"
)

type JobWorkerResult struct {
	Job     domain.Job
	Message *amqp.Delivery
	Error   error
}

func JobWorker(messageChannel chan amqp.Delivery,
	returnChannel chan JobWorkerResult,
	jobService JobService,
	job domain.Job,
	workerID int) {
	var err error

	for message := range messageChannel {

		if err = utils.IsJson(string(message.Body)); err != nil {
			returnChannel <- returnJobWorkerResult(domain.Job{}, message, err)
			continue
		}

		video := &jobService.VideoService.Video

		if err = json.Unmarshal(message.Body, video); err != nil {
			returnChannel <- returnJobWorkerResult(domain.Job{}, message, err)
			continue
		}

		(*video).ID = uuid.NewV4().String()

		if err = jobService.VideoService.Video.Validate(); err != nil {
			returnChannel <- returnJobWorkerResult(domain.Job{}, message, err)
			continue
		}

		if err = jobService.VideoService.Insert(); err != nil {
			returnChannel <- returnJobWorkerResult(domain.Job{}, message, err)
			continue
		}

		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("OUTPUT_BUCKET_NAME")
		job.ID = uuid.NewV4().String()
		job.Status = "STARTING"
		job.CreatedAt = time.Now()

		_, err = jobService.JobRepository.Insert(&job)

		if err != nil {
			returnChannel <- returnJobWorkerResult(domain.Job{}, message, err)
			continue
		}

		jobService.Job = &job
		err = jobService.Start()

		if err != nil {
			returnChannel <- returnJobWorkerResult(domain.Job{}, message, err)
			continue
		}
		returnChannel <- returnJobWorkerResult(job, message, nil)
	}
}

func returnJobWorkerResult(job domain.Job, message amqp.Delivery, err error) JobWorkerResult {
	return JobWorkerResult{
		Job:     job,
		Message: &message,
		Error:   err}
}
