package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (service *VideoService) Download(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(service.Video.FilePath)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}

	defer reader.Close()

	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	localStoragePath := os.Getenv("LOCAL_STORAGE_PATH") + "/"
	_ = ensureLocalStoragePath(localStoragePath)
	file, err := os.Create(localStoragePath + service.Video.ID + ".mp4")

	if err != nil {
		return err
	}

	_, err = file.Write(data)

	if err != nil {
		return err
	}

	defer file.Close()

	log.Printf("video %v has been stored", service.Video.ID)

	return nil
}

func (service *VideoService) Fragment() error {
	videoPath := os.Getenv("LOCAL_STORAGE_PATH") + "/" + service.Video.ID

	err := os.Mkdir(videoPath, os.ModePerm)

	if err != nil {
		return err
	}

	source := videoPath + ".mp4"
	target := videoPath + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (service *VideoService) Encode() error {
	videoPath := os.Getenv("LOCAL_STORAGE_PATH") + "/" + service.Video.ID

	var cmdArgs []string
	cmdArgs = append(cmdArgs, videoPath+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, videoPath)
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")

	cmd := exec.Command("mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (service *VideoService) Finish() error {
	videoPath := os.Getenv("LOCAL_STORAGE_PATH") + "/" + service.Video.ID

	err := os.Remove(videoPath + ".mp4")

	if err != nil {
		log.Println("Error removing file ", service.Video.ID, ".mp4")
	}

	err = os.Remove(videoPath + ".frag")

	if err != nil {
		log.Println("Error removing file ", service.Video.ID, ".frag")
	}

	err = os.RemoveAll(videoPath)

	if err != nil {
		log.Println("Error removing directory ", service.Video.ID)
	}

	log.Println("Files have been removed ", service.Video.ID)

	return nil
}

func (service *VideoService) Insert() error {
	_, err := service.VideoRepository.Insert(service.Video)
	return err
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("======> Output: %s\n", string(out))
	}
}

func ensureLocalStoragePath(localStoragePath string) error {
	baseDir := path.Dir(localStoragePath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}
