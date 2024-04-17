package model

import "gorm.io/gorm"

type Company struct {
	Model
	CompanyName string    `gorm:"not null" json:"company_name"`
	Histories   []History `gorm:"foreignkey:database_name" json:"histories"`
}

func (cr *Company) CreateCompany(db *gorm.DB) error {
	err := db.
		Model(Company{}).
		Create(&cr).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (cr *Company) GetCompanyById(db *gorm.DB) (Company, error) {
	res := Company{}

	err := db.
		Model(Company{}).
		Where("id = ?", cr.Model.ID).
		Preload("Histories").
		Take(&res).
		Error

	if err != nil {
		return Company{}, err
	}

	return res, nil
}

func (cr *Company) GetAllCompany(db *gorm.DB) ([]Company, error) {
	res := []Company{}

	err := db.
		Model(Company{}).
		Preload("Histories", func(db *gorm.DB) *gorm.DB {
			return db.Order("histories.updated_at DESC").Limit(1)
		}).
		Find(&res).
		Error

	if err != nil {
		return []Company{}, err
	}

	return res, nil
}

func (cr *Company) UpdateCompany(db *gorm.DB) error {
	err := db.
		Model(&Company{}).
		Where("id = ?", cr.ID).
		Updates(&cr).
		Error

	if err != nil {
		return err
	}

	return nil
}

func (cr *Company) DeleteCompany(db *gorm.DB) error {
	err := db.
		Model(&Company{}).
		Where("id = ?", cr.ID).
		Delete(&cr).
		Error

	if err != nil {
		return err
	}

	return nil
}
