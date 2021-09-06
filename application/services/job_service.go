package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"errors"
	"os"
	"strconv"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (j *JobService) Start() error {
	var err error

	if err = j.changeStatus("DOWNLOADING"); err != nil {
		return j.fail(err)
	}

	inputBucketName := os.Getenv("INPUT_BUCKET_NAME")
	if err = j.VideoService.Download(inputBucketName); err != nil {
		return j.fail(err)
	}

	if err = j.changeStatus("FRAGMENTING"); err != nil {
		return j.fail(err)
	}

	if err = j.VideoService.Fragment(); err != nil {
		return j.fail(err)
	}

	if err = j.changeStatus("ENCODING"); err != nil {
		return j.fail(err)
	}

	if err = j.VideoService.Encode(); err != nil {
		return j.fail(err)
	}

	if err = j.changeStatus("UPLOADING"); err != nil {
		return j.fail(err)
	}

	if err = j.upload(); err != nil {
		return j.fail(err)
	}

	if err = j.changeStatus("FINISHING"); err != nil {
		return j.fail(err)
	}

	if err = j.VideoService.Finish(); err != nil {
		return j.fail(err)
	}

	if err = j.changeStatus("COMPLETED"); err != nil {
		return j.fail(err)
	}

	return nil
}

func (j *JobService) upload() error {
	videoUpload := NewVideoUploadManager()
	videoUpload.OutputBucket = os.Getenv("OUTPUT_BUCKET_NAME")
	videoUpload.VideoPath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + j.VideoService.Video.ID

	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOADS"))
	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(concurrency, doneUpload)

	uploadResult := <-doneUpload

	if uploadResult != "completed" {
		return j.fail(errors.New(uploadResult))
	}

	return nil
}

func (j *JobService) changeStatus(status string) error {
	j.Job.Status = status
	_, err := j.JobRepository.Update(j.Job)
	if err != nil {
		return j.fail(err)
	}
	return nil
}

func (j *JobService) fail(error error) error {
	j.Job.Status = "FAILED"
	j.Job.Error = error.Error()
	_, err := j.JobRepository.Update(j.Job)
	if err != nil {
		return err
	}
	return error
}
