package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Phase string

const (
	PhaseInit          Phase = ""
	PhaseEnterBuilding Phase = "PhaseEnterBuilding"
)

type Session struct {
	ChatID   int64 `gorm:"primary_key"`
	Phase    Phase
	Note     string
	StreetID int
}

func NewDatabase(conf *Config) (*gorm.DB, error) {
	db, err := gorm.Open(conf.dbDriver, conf.dbConnection)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Session{})
	return db, nil
}
