package utils

import (
	"final-project/sekolah-beta/config"
	"final-project/sekolah-beta/models"
)

func ReadAllCompany()([]model.Company, error){
	var company model.Company
	return company.GetAllCompany(config.Mysql.DB)
}