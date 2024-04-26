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

type DatabaseConfig struct {
	Name     string `json:"database_name"`
	Host     string `json:"db_host"`
	Port     string `json:"db_port"`
	Username string `json:"db_username"`
	ID       int    `json:"id"`
}

func main() {
	// Membaca file konfigurasi
	configs, err := readConfig("../config.json")
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return
	}

	// Membuat map untuk menyimpan status impor setiap database
	importStatus := make(map[string]bool)

	// Iterasi semua konfigurasi database
	for i := 0; i < len(configs); i++ {
		config := configs[i]
		dbHost := config.Host
		dbPort := config.Port
		dbName := config.Name
		dbUser := config.Username
		dbID := i + 1 // Menyesuaikan ID dengan iterasi dimulai dari 1

		// Membuat URL file dengan menggunakan ID database yang sesuai
		fileURL := fmt.Sprintf("http://localhost:3000/company/%d/download", dbID)
		saveDir := "../download/"

		if !importStatus[dbName] {
			// Jika belum diimpor, lakukan impor
			if err := executeWorkflow(dbUser, dbHost, dbPort, dbName, fileURL, saveDir); err != nil {
				fmt.Printf("Error executing workflow for %s: %s\n", dbName, err)
				return
			}

			// Setel status impor menjadi true setelah impor selesai
			importStatus[dbName] = true
		}

		if err := removeFiles(saveDir); err != nil {
			fmt.Printf("Error removing files: %s\n", err)
			return
		}

		zipDir := "../unzip/"
		if err := removeFiles(zipDir); err != nil {
			fmt.Printf("Error removing files: %s\n", err)
			return
		}
	}

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

func readConfig(filePath string) ([]DatabaseConfig, error) {
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

func executeWorkflow(dbUser, dbHost, dbPort, dbName, fileURL, saveDir string) error {
	// Membuat channel untuk mengoordinasikan aliran kerja
	fileChan := make(chan string)
	done := make(chan bool)

	// Goroutine untuk mengeksekusi aliran kerja
	go func() {
		for zipFile := range fileChan {
			destDir := "../unzip/"

			// Mengekstrak file ZIP
			if err := unzipFile(zipFile, destDir); err != nil {
				fmt.Println("Error unzipping file:", err)
				return
			}

			// Mencari file SQL dalam direktori tujuan
			sqlFile, err := findSQLFile(destDir)
			if err != nil {
				fmt.Println("Error finding SQL file:", err)
				return
			}

			// Mengimpor database
			if err := importDatabase(dbUser, dbHost, dbPort, dbName, sqlFile); err != nil {
				fmt.Println("Error importing database:", err)
				return
			}
		}
		done <- true
	}()

	// Trace link URL
	fmt.Println(fileURL)

	// Mengunduh dan mengirim file ke channel
	if err := downloadAndSend(fileURL, saveDir, fileChan); err != nil {
		return err
	}

	// Menutup channel setelah selesai
	close(fileChan)
	<-done

	return nil
}

func unzipFile(zipFile, destDir string) error {
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

func findSQLFile(dir string) (string, error) {
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

func importDatabase(dbUser, dbHost, dbPort, dbName, sqlFile string) error {
	importCmd := fmt.Sprintf("mysql -u %s -h %s -P %s %s < %s", dbUser, dbHost, dbPort, dbName, sqlFile)
	fmt.Printf("Importing database %s...\n", dbName)
	var stdErr bytes.Buffer
	cmd := exec.Command("bash", "-c", importCmd)
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing command: %s", stdErr.String())
	}
	return nil
}

func downloadAndSend(fileURL, saveDir string, fileChan chan<- string) error {
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

	// Membuat nama file berdasarkan URL, kecuali jika sudah ada file dengan nama yang sama
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

	fileChan <- filePath
	return nil
}
