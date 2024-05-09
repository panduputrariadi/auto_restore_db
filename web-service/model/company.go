package model

import "gorm.io/gorm"

type Company struct {
	Model
	CompanyName string    `gorm:"not null" json:"company_name"`
	Histories   []History `gorm:"foreignkey:database_name" json:"histories"`
}

func (cr *Company) GetCompanyById(db *gorm.DB) (Company, error) {
	res := Company{}

	err := db.
		Model(Company{}).
		Where("company_name = ?", cr.CompanyName).
		Preload("Histories").
		Take(&res).
		Error

	if err != nil {
		return Company{}, err
	}

	return res, nil
}

func (cr *Company) GetAllCompany(db *gorm.DB) ([]Company, error) {
	var companies []Company

	// query dapetin update terbaru
	subQuery := db.Model(&History{}).
		Select("database_name, MAX(updated_at) AS latest_updated_at").
		Group("database_name")

	// Mengambil semua companies dengan history terbaru
	if err := db.
		Preload("Histories", func(db *gorm.DB) *gorm.DB {
			return db.Joins("INNER JOIN (?) AS latest_histories ON histories.database_name = latest_histories.database_name AND histories.updated_at = latest_histories.latest_updated_at", subQuery)
		}).
		Find(&companies).Error; err != nil {
		return nil, err
	}

	return companies, nil
}


func (cr *Company) DownloadCompanyHistoryFile(db *gorm.DB, companyName string) (string, error) {
    var history History
    err := db.
        Model(&History{}).
        Joins("JOIN companies ON companies.id = histories.database_name").
        Where("companies.company_name = ?", companyName).
        Select("histories.file").
        Order("histories.updated_at desc").
        Limit(1).
        First(&history).
        Error

    if err != nil {
        return "", err
    }

    return history.File, nil
}

