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
	"time"
)

type DatabaseConfig struct {
	Name     string `json:"database_name"`
	Host     string `json:"db_host"`
	Port     string `json:"db_port"`
	Username string `json:"db_username"`
	ID       int    `json:"id"`
}

func main() {
    startTime := time.Now()
    configChan := make(chan []DatabaseConfig)
    
    go readConfigAsync("../config.json", configChan)
    configs := <-configChan
    
    importStatus := make(map[string]bool)
    saveDir := "../download/"
    
    if err := processDatabases(configs, importStatus, saveDir); err != nil {
        fmt.Printf("Error processing databases: %s\n", err)
        return
    }
    
    endTime := time.Now()
    executionTime := endTime.Sub(startTime).Seconds()
    fmt.Printf("Execution time: %.2f seconds\n", executionTime)
}

func readConfigAsync(filePath string, configChan chan<- []DatabaseConfig) {
    configs, err := readConfig(filePath)
    if err != nil {
        fmt.Printf("Error reading config file: %s\n", err)
        configChan <- nil
        return
    }
    configChan <- configs
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

func processDatabases(configs []DatabaseConfig, importStatus map[string]bool, saveDir string) error {
	for i, config := range configs {
		dbHost := config.Host
		dbPort := config.Port
		dbName := config.Name
		dbUser := config.Username
		dbID := i + 1

		fileURL := fmt.Sprintf("http://localhost:3000/company/%d/download", dbID)

		if !importStatus[dbName] {
			if err := executeWorkflow(dbUser, dbHost, dbPort, dbName, fileURL, saveDir); err != nil {
				return fmt.Errorf("error executing workflow for %s: %s", dbName, err)
			}
			importStatus[dbName] = true
		}

		if err := removeFiles(saveDir); err != nil {
			fmt.Printf("Error removing files: %s\n", err)
			return err
		}

		zipDir := "../unzip/"
		if err := removeFiles(zipDir); err != nil {
			fmt.Printf("Error removing files: %s\n", err)
			return err
		}
	}
	return nil
}

func executeWorkflow(dbUser, dbHost, dbPort, dbName, fileURL, saveDir string) error {
	fileChan := make(chan string)
	done := make(chan bool)

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
    // Membuat channel untuk mengirim error
    errChan := make(chan error)

    // Goroutine untuk mengekstrak file ZIP
    go func() {
        r, err := zip.OpenReader(zipFile)
        if err != nil {
            errChan <- err
            return
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
                errChan <- err
                return
            }
            defer rc.Close()

            fDest, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                errChan <- err
                return
            }
            defer fDest.Close()

            _, err = io.Copy(fDest, rc)
            if err != nil {
                errChan <- err
                return
            }
        }

        errChan <- nil // Mengirim sinyal bahwa tidak ada error
    }()

    // Menunggu goroutine selesai atau error
    if err := <-errChan; err != nil {
        return err
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
    // Membuat channel untuk mengirim error
    errChan := make(chan error)

    // Goroutine untuk mengimpor database
    go func() {
        importCmd := fmt.Sprintf("mysql -u %s -h %s -P %s %s < %s", dbUser, dbHost, dbPort, dbName, sqlFile)
        fmt.Printf("Importing database %s...\n", dbName)
        var stdErr bytes.Buffer
        cmd := exec.Command("bash", "-c", importCmd)
        cmd.Stderr = &stdErr
        err := cmd.Run()
        if err != nil {
            errChan <- fmt.Errorf("error executing command: %s", stdErr.String())
            return
        }

        errChan <- nil // Mengirim sinyal bahwa tidak ada error
    }()

    // Menunggu goroutine selesai atau error
    if err := <-errChan; err != nil {
        return err
    }

    return nil
}


func downloadAndSend(fileURL, saveDir string, fileChan chan<- string) error {
    // Membuat channel untuk mengirim error
    errChan := make(chan error)

    // Goroutine untuk download file
    go func() {
        resp, err := http.Get(fileURL)
        if err != nil {
            errChan <- err
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            errChan <- fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
            return
        }

        if resp.ContentLength == 0 {
            errChan <- fmt.Errorf("no content found at URL: %s", fileURL)
            return
        }

        fileName := filepath.Base(fileURL)
        filePath := filepath.Join(saveDir, fileName)

        outFile, err := os.Create(filePath)
        if err != nil {
            errChan <- err
            return
        }
        defer outFile.Close()

        _, err = io.Copy(outFile, resp.Body)
        if err != nil {
            errChan <- err
            return
        }

        fileChan <- filePath 
        errChan <- nil      
    }()

    // Menunggu goroutine selesai atau error
    if err := <-errChan; err != nil {
        return err
    }

    return nil
}
func removeFiles(dirPath string) error {
    // Membuat channel untuk mengirim error
    errChan := make(chan error)

    // Goroutine untuk menghapus file
    go func() {
        dir, err := os.Open(dirPath)
        if err != nil {
            errChan <- err
            return
        }
        defer dir.Close()

        files, err := dir.Readdir(-1)
        if err != nil {
            errChan <- err
            return
        }

        for _, file := range files {
            err = os.RemoveAll(filepath.Join(dirPath, file.Name()))
            if err != nil {
                errChan <- err
                return
            }
        }

        errChan <- nil // Mengirim sinyal bahwa tidak ada error
    }()

    // Menunggu goroutine selesai atau error
    if err := <-errChan; err != nil {
        return err
    }

    return nil
}

