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
)

func main() {

	// URL layanan web, ganti {id} dengan ID yang sesuai.
	fileURL := "http://localhost:3000/company/3/download"

	// Path lokal tempat file akan disimpan
	saveDir := "../download/"

	// Panggil fungsi download
	if err := downloadFile(fileURL, saveDir); err != nil {
		fmt.Printf("Error downloading file: %s\n", err)
		return
	} else {
		fmt.Println("File downloaded successfully.")
	}

	dirPath := "../download/"
	dir, err := os.Open(dirPath)
	if err != nil {
		fmt.Printf("Error opening directory: %s\n", err)
		return
	}
	defer dir.Close()

	files, err := dir.ReadDir(0)
	if err != nil {
		fmt.Printf("Error reading directory: %s\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			zipFile := filepath.Join(dirPath, file.Name())
			destDir := "../unzip/"

			if err := unzipExtractor(zipFile, destDir); err != nil {
				fmt.Println("Error unzipping file:", err)
				return
			} else {
				fmt.Println("Unzipping completed successfully")
			}

			configData, err := ioutil.ReadFile("../config.json")
			if err != nil {
				fmt.Printf("Error reading config file: %s\n", err)
				return
			}

			var config []map[string]string
			err = json.Unmarshal(configData, &config)
			if err != nil {
				fmt.Printf("Error decoding config JSON: %s\n", err)
				return
			}

			dbHost := config[0]["db_host"]
			dbPort := config[0]["db_port"]
			dbName := config[0]["database_name"]
			dbUser := config[0]["db_username"]
			// dbPassword := config[0]["db_password"]
			

			fileList, err := ioutil.ReadDir(destDir)
			if err != nil {
				fmt.Printf("Error reading directory: %s\n", err)
				return
			}

			var sqlFile string
			for _, f := range fileList {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
					sqlFile = filepath.Join(destDir, f.Name())
					break
				}
			}

			if sqlFile == "" {
				fmt.Println("No SQL file found in the destination directory")
				return
			}

			
			importCmd := fmt.Sprintf("mysql -u %s  -h %s -P %s %s < %s", dbUser, dbHost, dbPort, dbName, sqlFile)

			var stdErr bytes.Buffer
			cmd := exec.Command("bash", "-c", importCmd)

			cmd.Stderr = &stdErr
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error executing command: %s\n", stdErr.String())
				return
			}

			fmt.Println("Database imported successfully")

			if err := removeFiles(dirPath); err != nil {
				fmt.Printf("Error removing files: %s\n", err)
				return
			}
			if err := removeFiles(destDir); err != nil {
				fmt.Printf("Error removing files: %s\n", err)
				return
			}
		}
	}
}

func unzipExtractor(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(destDir, 0755)
	for _, f := range r.File {
		if filepath.Base(f.Name) == "__MACOSX" || strings.HasPrefix(filepath.Base(f.Name), "._") {
			continue
		}
		filePath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, f.Mode())
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fDest, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer fDest.Close()

		_, err = io.Copy(fDest, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadFile(fileURL, saveDir string) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server return non-200 status: %d %s", resp.StatusCode, resp.Status)
	}
	fileName := filepath.Base(fileURL)

	savePath := filepath.Join(saveDir, fileName)
	outFile, err := os.Create(savePath)
	if err != nil {
		return err
	}


	defer outFile.Close()
	_, err = io.Copy(outFile, resp.Body)
	return err
}

func removeFiles(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.RemoveAll(filepath.Join(dirPath, file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
