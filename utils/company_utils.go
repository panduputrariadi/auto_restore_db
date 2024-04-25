package utils

import (
	"final-project/sekolah-beta/config"
	"final-project/sekolah-beta/model"
)

func ReadAllCompany()([]model.Company, error){
	var company model.Company
	return company.GetAllCompany(config.Mysql.DB)
}

func GetCompanyID(id uint) (model.Company, error) {
	company := model.Company{
		Model: model.Model{
			ID: id,
		},
	}
	return company.GetCompanyById(config.Mysql.DB)
}
func DownloadCompanyHistoryFile(companyID uint) (string, error) {
	company := model.Company{
		Model: model.Model{
			ID: companyID,
		},
	}
	return company.DownloadCompanyHistoryFile(config.Mysql.DB, companyID)
}