package models

import (
	"github.com/corporateanon/my1562bot/pkg/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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

type Subscription struct {
	gorm.Model
	ChatID       int64
	AddressIDAr  uint32
	StreetID1562 uint32
}

type Address struct {
	gorm.Model
	Subscriptions []Subscription
	StreetID      int
	Building      string
	StreetName    string
}

//NewDatabase creates database connection
func NewDatabase(conf *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(conf.DBDriver, conf.DBConnection)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Session{}, &Subscription{}, &Address{})
	return db, nil
}
