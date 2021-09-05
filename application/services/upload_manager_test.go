package services_test

import (
	"encoder/application/services"
	dotenv "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	err := dotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestVideoUploadProcessUpload(t *testing.T) {
	videoService := services.NewVideoService()
	videoService.Video, videoService.VideoRepository = prepare()

	err := videoService.Download(os.Getenv("INPUT_BUCKET_NAME"))
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	videoUploadManager := services.NewVideoUploadManager()
	videoUploadManager.OutputBucket = os.Getenv("OUTPUT_BUCKET_NAME")
	videoUploadManager.VideoPath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + videoService.Video.ID

	doneUpload := make(chan string)
	go videoUploadManager.ProcessUpload(50, doneUpload)

	result := <-doneUpload
	require.Equal(t, result, "completed")

	err = videoService.Finish()
	require.Nil(t, err)
}
