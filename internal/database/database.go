package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

type DbStruct struct {
	gorm.Model
	Note         string
	Octave       string
	NumTries     int
	CorrectCount int
	Accuracy     float32
}

func New(filePath string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&DbStruct{})

	return &Database{DB: db}, nil
}

func (db *Database) Get(condition map[string]interface{}, result *DbStruct) bool {
	exist := false

	answr := db.DB.Where(condition).Find(result)

	if answr.RowsAffected == 0 {
		return exist
	}

	exist = true

	return exist
}

func (db *Database) Save(data *DbStruct) {
	db.DB.Save(data)
}
