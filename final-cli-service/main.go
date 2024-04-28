package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	ch := make(chan DatabaseConfig)
	go func() {
		defer close(ch)
		for _, db := range configs {
			ch <- db
		}
	}()

	download := DownloadFileWithWorker(ch, 2, configs)
	for result := range download {
		fmt.Println(result)
	}

	unzip := UnzipFileWithWorker(download, 2, "../web-service/unzip/")
	for result := range unzip {
		if result.Error != nil {
			fmt.Println("Error:", result.Error)
		}
	}

	importDatabase := ImportFileWithWorker(unzip, 2, "../web-service/unzip/")
	// fmt.Println(importDatabase)
	for result := range importDatabase {
		if result.Error != nil {
			fmt.Println("Error:", result.Error)
		}
	}

	rmDownload := "../web-service/download/"

	deleteFileDownload := DeleteFileWithWorker(unzip, 2, rmDownload)
	for result := range deleteFileDownload {
		if result.Error != nil {
			fmt.Printf("Error removing files: %s\n", err)
		}
	}
	rmZip := "../web-service/unzip/"
	deleteFileZip := DeleteFileWithWorker(deleteFileDownload, 2, rmZip)
	for result := range deleteFileZip {
		if result.Error != nil {
			fmt.Printf("Error removing files: %s\n", err)
		}
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

func GoRoutineDownloadFile(chMasuk chan DatabaseConfig, configs []DatabaseConfig) chan DatabaseConfig {
	chKeluar := make(chan DatabaseConfig)
	go func() {
		defer close(chKeluar)
		for i, config := range configs {
			dbName := config.Name
			dbID := i + 1

			saveDir := "./download/"
			fileURL := fmt.Sprintf("http://localhost:3000/company/%d/download", dbID)

			if err := DownloadFile(fileURL, saveDir); err != nil {
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

func UnzipFileWithWorker(chin chan DatabaseConfig, worker int, destDir string) chan DatabaseConfig {
	chout := make(chan DatabaseConfig)
	wg := sync.WaitGroup{}
	files, err := ioutil.ReadDir("../web-service/download/")
	if err != nil {
		chout <- DatabaseConfig{Error: fmt.Errorf("error reading download directory: %s", err)}
		close(chout)
		return chout
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
			zipFile := filepath.Join("../web-service/download/", file.Name())
			wg.Add(1)
			go func(zipFile string) {
				defer wg.Done()
				unzipFile := GoRoutineUnzipFile(chin, destDir, zipFile)
				for result := range unzipFile {
					if result.Error != nil {
						chout <- result
					}
				}
			}(zipFile)
		}
	}

	go func() {
		wg.Wait()
		close(chout)
	}()

	return chout
}

func GoRoutineUnzipFile(chin chan DatabaseConfig, destDir, zipFile string) chan DatabaseConfig {
	chKeluar := make(chan DatabaseConfig)

	go func() {
		defer close(chKeluar)

		r, err := zip.OpenReader(zipFile)
		if err != nil {
			chKeluar <- DatabaseConfig{Error: fmt.Errorf("failed to open read ZIP file: %v", err)}
			return
		}
		defer r.Close()

		var dbName string
		for db := range chin {
			if db.Error != nil {
				chKeluar <- DatabaseConfig{Error: fmt.Errorf("failed to get database name: %v", db.Error)}
				return
			}
			dbName = db.Name
			continue
		}

		for _, f := range r.File {
			if filepath.Base(f.Name) == "__MACOSX" || strings.HasPrefix(filepath.Base(f.Name), "._") {
				continue
			}

			// code untuk ekstrak unzip file
			extractedFilePath := filepath.Join(destDir, dbName, filepath.Base(f.Name))

			if f.FileInfo().IsDir() {
				os.MkdirAll(extractedFilePath, f.Mode())
				continue
			}

			// jika file bukan sebuah direktori
			rc, err := f.Open()
			if err != nil {
				chKeluar <- DatabaseConfig{Error: fmt.Errorf("failed to open file in ZIP: %v", err)}
				return
			}
			defer rc.Close()

			// membuat destinasi file
			fDest, err := os.OpenFile(extractedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				chKeluar <- DatabaseConfig{Error: fmt.Errorf("failed to create destination file: %v", err)}
				return
			}
			defer fDest.Close()

			// copy file menuju destinasinya
			_, err = io.Copy(fDest, rc)
			if err != nil {
				chKeluar <- DatabaseConfig{Error: fmt.Errorf("failed to copy file contents: %v", err)}
				return
			}
		}

		chKeluar <- DatabaseConfig{Error: nil}
	}()

	return chKeluar
}

func ImportFileWithWorker(chin chan DatabaseConfig, worker int, destDir string) chan DatabaseConfig {
	channels := []chan DatabaseConfig{}

	chout := make(chan DatabaseConfig)

	wg := sync.WaitGroup{}

	wg.Add(worker)

	go func() {
		wg.Wait()
		close(chout)
	}()

	//Fan-in
	for i := 0; i < worker; i++ {
		channels = append(channels, GoroutineImportFile(chin, destDir))
	}

	//Fan-out
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

func GoroutineImportFile(chin chan DatabaseConfig, unzipDir string) chan DatabaseConfig {
	chout := make(chan DatabaseConfig)

	go func() {
		defer close(chout)

		for db := range chin {
			if db.Error != nil {
				chout <- db
				continue
			}

			sqlFile, err := readSql(unzipDir)
			if err != nil {
				chout <- DatabaseConfig{Error: fmt.Errorf("error reading SQL file: %v", err)}
				continue
			}

			importCmd := fmt.Sprintf("mysql -u %s -h %s -P %s %s < %s", db.Username, db.Host, db.Port, db.Name, sqlFile)
			var stdErr bytes.Buffer

			cmd := exec.Command("bash", "-c", importCmd)
			cmd.Stderr = &stdErr

			err = cmd.Run()
			if err != nil {
				chout <- DatabaseConfig{Error: fmt.Errorf("error executing command: %v\nstderr: %s", err, stdErr.String())}
				continue
			}

			chout <- db
		}
	}()

	return chout
}

func readSql(destDir string) (string, error) {
	// Mencari file SQL dalam direktori tujuan
	sqlFile, err := FindSQLFile(destDir)
	if err != nil {
		fmt.Println("Error finding SQL file:", err.Error())
		return "", err
	} else {
		fmt.Println(sqlFile)
	}

	return sqlFile, err
}

func FindSQLFile(dir string) (string, error) {
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, f := range fileList {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			return filepath.Join(dir, f.Name()), nil
		}
	}

	return "", fmt.Errorf("no SQL file found in the destination directory")
}

func DeleteFileWithWorker(chin chan DatabaseConfig, worker int, dirPath string) chan DatabaseConfig {
	channels := []chan DatabaseConfig{}

	chout := make(chan DatabaseConfig)

	wg := sync.WaitGroup{}

	wg.Add(worker)

	go func() {
		wg.Wait()
		close(chout)
	}()

	//Fan-in
	for i := 0; i < worker; i++ {
		channels = append(channels, GoroutineDeleteFile(chin, dirPath))
	}

	//Fan-out
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

func GoroutineDeleteFile(chin chan DatabaseConfig, dirPath string) chan DatabaseConfig {
	chout := make(chan DatabaseConfig)

	go func() {
		defer close(chout)

		for db := range chin {
			if db.Error != nil {
				chout <- db
				continue
			}
			dir, err := os.Open(dirPath)
			if err != nil {
				chout <- DatabaseConfig{Error: fmt.Errorf("error reading SQL file: %v", err)}
				return
			}
			defer dir.Close()

			files, err := dir.Readdir(-1)
			if err != nil {
				chout <- DatabaseConfig{Error: fmt.Errorf("error reading SQL file: %v", err)}
				return
			}

			for _, file := range files {
				err = os.RemoveAll(filepath.Join(dirPath, file.Name()))
				if err != nil {
					chout <- DatabaseConfig{Error: fmt.Errorf("error reading SQL file: %v", err)}
					return
				}
			}

			chout <- db
		}
	}()

	return chout
}
