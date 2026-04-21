package database

import "gorm.io/gorm"

type InventoryDbStruct struct {
	gorm.Model
	NumOfSafeFails int
	UserID         int
}
