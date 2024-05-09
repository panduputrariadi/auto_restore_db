package model_test

import (
	"final-project/sekolah-beta/config"
	"final-project/sekolah-beta/model"
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func Init() {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Print(err)
	}
	config.OpenDB()
}
func TestGetAll(t *testing.T) {
	Init()
	companies := []model.Company{
		{
			CompanyName: "Test Company 1",
			Histories:   []model.History{},
		},
		{
			CompanyName: "Test Company 2",
			Histories:   []model.History{},
		},
	}

	res, err := companies.GetAllCompany(config.Mysql.DB)
	assert.Nil(t, err)
}
