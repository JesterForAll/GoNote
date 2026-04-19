package database

import (
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func New(filePath string, logger *slog.Logger, model ...interface{}) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(model...)
	if err != nil {
		logger.Error("error migrating database", slog.Any("err", err))
	}

	return &Database{DB: db}, nil
}

func (db *Database) CheckIfExistAndGetFirst(condition map[string]interface{}, result interface{}) bool {
	res := db.DB.Where(condition).First(result)

	return res.RowsAffected != 0
}

func (db *Database) Upsert(data interface{}) error {
	return db.DB.Save(data).Error
}

func (db *Database) GetAll(result interface{}) error {
	return db.DB.Find(result).Error
}
