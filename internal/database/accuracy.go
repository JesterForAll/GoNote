package database

import "gorm.io/gorm"

type AccuracyDbStruct struct {
	gorm.Model
	Note         string
	Octave       string
	NumTries     int
	CorrectCount int
	Accuracy     float32
}
