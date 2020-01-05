package models

import (
	"github.com/corporateanon/my1562bot/pkg/config"
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

func NewDatabase(conf *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(conf.DBDriver, conf.DBConnection)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Session{})
	return db, nil
}
