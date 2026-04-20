package database

import "gorm.io/gorm"

type BalanceDbStruct struct {
	gorm.Model
	Balance int
	UserID  int
}
