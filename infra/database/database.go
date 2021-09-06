package database

import (
	"encoder/domain"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	Db            *gorm.DB
	Debug         bool
	AutoMigrateDb bool
	Env           string
	Dsn           string
	DsnTest       string
	DbType        string
	DbTypeTest    string
}

func NewDb() *Database {
	return &Database{}
}

func NewDbTest() *gorm.DB {
	dbInstance := NewDb()
	dbInstance.AutoMigrateDb = true
	dbInstance.Debug = true
	dbInstance.Env = "test"
	dbInstance.DbTypeTest = "sqlite3"
	dbInstance.DsnTest = ":memory:"

	connection, err := dbInstance.Connect()

	if err != nil {
		log.Fatalf("Test db error: %v", err)
	}

	return connection
}

func (database *Database) Connect() (*gorm.DB, error) {
	var err error

	if database.Env != "test" {
		database.Db, err = gorm.Open(database.DbType, database.Dsn)
	} else {
		database.Db, err = gorm.Open(database.DbTypeTest, database.DsnTest)
	}

	if err != nil {
		return nil, err
	}

	if database.Debug {
		database.Db.LogMode(true)
	}

	if database.AutoMigrateDb {
		database.Db.AutoMigrate(&domain.Video{}, &domain.Job{})

		database.Db.Model(domain.Job{}).AddForeignKey("video_id",
			"videos (id)",
			"CASCADE",
			"CASCADE",
		)
	}

	return database.Db, nil
}
