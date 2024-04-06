package model

import "gorm.io/gorm"

type History struct {
	Model
	DatabaseName string `gorm:"not null" json:"database_name"`
	FileName string `gorm:"not null" json:"file_name"`
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

func (cr *History) GetAllHistory(db *gorm.DB) ([]History, error) {
	res := []History{}

	err := db.
		Model(History{}).
		Find(&res).
		Error

	if err != nil {
		return []History{}, err
	}

	return res, nil
}

func (cr *History) UpdateHistory(db *gorm.DB) error {
	err := db.
		Model(&History{}).
		Where("id = ?", cr.ID).
		Updates(&cr).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (cr *History) DeleteHistory(db *gorm.DB) error {
	err := db.
		Model(&History{}).
		Where("id = ?", cr.ID).
		Delete(&cr).
		Error

	if err != nil {
		return err
	}

	return nil
}
