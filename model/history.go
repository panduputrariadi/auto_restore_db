package model

import (
	"gorm.io/gorm"
)

type History struct {
	Model
	DatabaseName int `gorm:"not null" json:"database_name"`
	File string `gorm:"not null" json:"file"`
}

func (cr *History) CreateHistory(db *gorm.DB) error {
	err := db.
		Model(History{}).
		Create(&cr).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (cr *History) GetHistoryById(db *gorm.DB) (History, error) {
	res := History{}

	err := db.
		Model(History{}).
		Where("id = ?", cr.Model.ID).
		Take(&res).
		Error

	if err != nil {
		return History{}, err
	}

	return res, nil
}