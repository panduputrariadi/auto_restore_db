package utils

import (
	"final-project/sekolah-beta/config"
	"final-project/sekolah-beta/model"
	"time"
)

func InsertHistoryData(data model.History) (model.History, error) {

	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	err := data.CreateHistory(config.Mysql.DB)

	return data, err
}