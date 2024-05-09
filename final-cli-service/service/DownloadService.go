package service

import (
	"final-project/sekolahbeta-hacker/cli-service/model"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func GoRoutineDownloadFile(chMasuk chan model.DatabaseConfig, clientHeader string) chan model.DatabaseConfig {
	chKeluar := make(chan model.DatabaseConfig)
	go func() {
		defer close(chKeluar)

		for config := range chMasuk {
			if config.Error != nil {
				fmt.Println(config.Error.Error())
				chKeluar <- config
				continue
			}

			dbName := config.Name

			// dbID := config.ID

			// fmt.Println(dbID)

			saveDir := "./download/"
			// fileURL := fmt.Sprintf("http://localhost:3000/company/%d/download", dbID)
			fileURL := fmt.Sprintf("http://localhost:3000/company/download?company_name=%s", dbName)


			file, err := DownloadFile(fileURL, saveDir, config, clientHeader)

			if err != nil {
				fmt.Printf("Error downloading file on db name: %s\n", err)
				config.Error = err
				chKeluar <- config
				continue
			} else {
				fmt.Println("File downloaded successfully: ", dbName)
			}

			config.FileDownloaded = file
			chKeluar <- config
		}
	}()

	return chKeluar
}

func DownloadFile(fileURL, saveDir string, config model.DatabaseConfig, clientHeader string) (string, error) {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Client", clientHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	if resp.ContentLength == 0 {
		return "", fmt.Errorf("no content found at URL: %s", fileURL)
	}

	fileName := fmt.Sprintf("%s.zip", config.Name)
	filePath := filepath.Join(saveDir, fileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
