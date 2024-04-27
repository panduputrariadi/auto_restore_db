package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type DatabaseConfig struct {
	Name     string `json:"database_name"`
	Host     string `json:"db_host"`
	Port     string `json:"db_port"`
	Username string `json:"db_username"`
	ID       int    `json:"id"`
	Error    error
}

func main() {
	configs, err := BacaConfig("config.json")
    if err != nil {
        fmt.Printf("Error reading config file: %s\n", err)
        return
    }

	ch:= make(chan DatabaseConfig)
	go func() {
		defer close(ch)
		for _, db := range configs {
			ch <- db
		}
	}()

	download:= DownloadFileWithWorker(ch, 2, configs)
	for result := range download {
		fmt.Println(result)
	}
}


func BacaConfig(filePath string) ([]DatabaseConfig, error) {
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

func DownloadFileWithWorker(chin chan DatabaseConfig, worker int, configs []DatabaseConfig) chan DatabaseConfig {
	channels := []chan DatabaseConfig{}

	chout := make(chan DatabaseConfig)

	wg := sync.WaitGroup{}

	wg.Add(worker)

	go func() {
		wg.Wait()
		close(chout)
	}()

	for i := 0; i < worker; i++ {
		channels = append(channels, GoRoutineDownloadFile(chin, configs))
		
	}

	for _, ch := range channels {
		go func(channel chan DatabaseConfig) {
			for c := range channel {
				chout <- c
			}

			wg.Done()
		}(ch)

	}

	return chout
}

func GoRoutineDownloadFile(chMasuk chan DatabaseConfig, configs []DatabaseConfig) chan DatabaseConfig{
	chKeluar := make(chan DatabaseConfig)
	go func ()  {
		defer close(chKeluar)
		for i, config := range configs {
			dbName := config.Name
			dbID := i + 1

			saveDir := "./download/"
			fileURL := fmt.Sprintf("http://localhost:3000/company/%d/download", dbID)

			if err := DownloadFile(fileURL, saveDir);  err != nil {
				fmt.Printf("Error downloading file on db name: %s\n", err)
			} else {
				fmt.Println("File downloaded successfully: ", dbName)
			}
	
		}
	}()

	return chKeluar
}

func DownloadFile(fileURL, saveDir string) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	if resp.ContentLength == 0 {
		return fmt.Errorf("no content found at URL: %s", fileURL)
	}

	fileName := filepath.Base(fileURL)
	filePath := filepath.Join(saveDir, fileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
