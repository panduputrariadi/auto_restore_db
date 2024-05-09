package config

import (
	"encoding/json"
	"final-project/sekolahbeta-hacker/cli-service/model"
	"io/ioutil"
)

func BacaConfig(filePath string) ([]model.DatabaseConfig, error) {
	configData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var configs []model.DatabaseConfig
	if err := json.Unmarshal(configData, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}
