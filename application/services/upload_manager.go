package services

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
}

func NewVideoUploadManager() *VideoUpload {
	return &VideoUpload{}
}

func (vu *VideoUpload) UploadObject(objectPath string, client *storage.Client, ctx context.Context) error {
	l := os.Getenv("LOCAL_STORAGE_PATH") + "/"
	name := strings.Split(objectPath, l)[1]

	f, err := os.Open(objectPath)

	if err != nil {
		return err
	}

	defer f.Close()

	wc := client.Bucket(vu.OutputBucket).Object(name).NewWriter(ctx)
	//wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	_, err = io.Copy(wc, f)

	if err != nil {
		return err
	}

	err = wc.Close()

	if err != nil {
		return err
	}

	return nil
}

func (vu *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
	in := make(chan int, runtime.NumCPU())
	returnChannel := make(chan string)

	err := vu.loadPaths()
	if err != nil {
		return err
	}

	uploadClient, ctx, err := getClientUpload()
	if err != nil {
		return err
	}

	for process := 0; process < concurrency; process++ {
		go vu.uploadWorker(in, returnChannel, uploadClient, ctx)
	}

	go func() {
		for i := 0; i < len(vu.Paths); i++ {
			in <- i
		}
		close(in)
	}()

	for r := range returnChannel {
		if r != "" {
			doneUpload <- r
			break
		}
	}

	return nil
}

func (vu *VideoUpload) uploadWorker(in chan int, out chan string, client *storage.Client, ctx context.Context) {
	for i := range in {
		err := vu.UploadObject(vu.Paths[i], client, ctx)
		if err != nil {
			vu.Errors = append(vu.Errors, vu.Paths[i])
			log.Printf("Error during the upload: %v. Error: %v", vu.Paths[i], err)
			out <- err.Error()
		}
		out <- ""
	}
	out <- "completed"
}

func (vu *VideoUpload) loadPaths() error {

	fn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			vu.Paths = append(vu.Paths, path)
		}
		return nil
	}

	if err := filepath.Walk(vu.VideoPath, fn); err != nil {
		return err
	}

	return nil
}

func getClientUpload() (*storage.Client, context.Context, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return client, ctx, nil
}
