package database

import "gorm.io/gorm"

type LoginDBStruct struct {
	gorm.Model
	UserName string `gorm:"uniqueIndex"`
}
