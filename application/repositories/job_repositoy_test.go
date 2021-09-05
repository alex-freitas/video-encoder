package repositories_test

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/infrastructure/database"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

func TestJobRepositoryDbInsert(t *testing.T) {
	db := database.NewDbTest()
	defer func(db *gorm.DB) {
		_ = db.Close()
	}(db)

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}
	_, _ = repo.Insert(video)

	job, err := domain.NewJob("output_path", "Pending", video)
	require.Nil(t, err)

	jobRepo := repositories.JobRepositoryDb{Db: db}
	_, _ = jobRepo.Insert(job)

	found, err := jobRepo.Find(job.ID)

	require.Nil(t, err)
	require.NotEmpty(t, job.ID)
	require.NotEmpty(t, job.VideoID)
	require.Equal(t, found.ID, job.ID)
	require.Equal(t, found.VideoID, video.ID)
}

func TestJobRepositoryDbUpdate(t *testing.T) {
	db := database.NewDbTest()
	defer func(db *gorm.DB) {
		_ = db.Close()
	}(db)

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}
	_, _ = repo.Insert(video)

	job, err := domain.NewJob("output_path", "Pending", video)
	require.Nil(t, err)

	jobRepo := repositories.JobRepositoryDb{Db: db}
	_, err = jobRepo.Insert(job)
	require.Nil(t, err)

	job.Status = "Complete"
	_, err = jobRepo.Update(job)
	require.Nil(t, err)

	found, err := jobRepo.Find(job.ID)
	require.Nil(t, err)
	require.NotEmpty(t, job.ID)
	require.Equal(t, found.Status, job.Status)
}
