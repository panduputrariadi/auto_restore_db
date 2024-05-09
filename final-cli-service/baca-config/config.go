package bacaconfig


func BacaConfig(filePath string) ([]data, error) {
	configData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var configs []DatabaseConfig
	if err := json.Unmarshal(configData, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}