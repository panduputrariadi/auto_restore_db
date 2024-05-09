package utils

import (
	"final-project/sekolah-beta/config"
	"final-project/sekolah-beta/model"
)

func ReadAllCompany()([]model.Company, error){
	var company model.Company
	return company.GetAllCompany(config.Mysql.DB)
}

func GetCompanyID(name string) (model.Company, error) {
	company := model.Company{
		CompanyName: name,
	}
	return company.GetCompanyById(config.Mysql.DB)
}
func DownloadCompanyHistoryFile(companyName string) (string, error) {
	company := model.Company{
		CompanyName: companyName,
	}
	return company.DownloadCompanyHistoryFile(config.Mysql.DB, companyName)
}