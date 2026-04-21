package database

import (
	"context"
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

func (db *Database) Begin() *gorm.DB {
	return db.DB.Begin()
}

func (db *Database) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

func (db *Database) Upsert(data interface{}) error {
	return db.DB.Save(data).Error
}

func (db *Database) GetAll(result interface{}) error {
	return db.DB.Find(result).Error
}

func (db *Database) UpsertWithTx(tx *gorm.DB, data interface{}) error {
	return tx.Save(data).Error
}
