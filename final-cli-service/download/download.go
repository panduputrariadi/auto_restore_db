package download

import (
	"fmt"
	"io"
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
				fmt.Println("File downloaded successfully: ", dbName, "on url: ", fileURL)
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