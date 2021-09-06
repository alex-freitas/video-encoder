package repositories_test

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/infra/database"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

func TestVideoRepositoryDbInsert(t *testing.T) {
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

	found, err := repo.Find(video.ID)

	require.NotEmpty(t, found.ID)
	require.Nil(t, err)
	require.Equal(t, found.ID, video.ID)
}
