package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)


func main() {
	
	zipFile := "../uploads/mysql_2024-04-24_5007b2ab-13fc-4be1-b2d9-74d1da73ea75.zip"
    destDir := "../unzip"

    if err := unzipExtractor(zipFile, destDir); err != nil {
        fmt.Println("Error unzipping file:", err)
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
	dbPassword := config[0]["db_password"]
    filePath := "../unzip"

    cmdString := fmt.Sprintf("mysql -u %s -p%s -h %s -P %s %s < %s", dbUser, dbPassword, dbHost, dbPort, dbName, filePath)

    cmd := exec.Command("bash", "-c", cmdString)

    // Jalankan perintah dan tangani error
    err = cmd.Run()
    if err != nil {
        fmt.Printf("Error executing command: %s\n", err)
        return
    }

    fmt.Println("Database imported successfully")

	// URL layanan web, ganti {id} dengan ID yang sesuai.
    fileURL := "http://localhost:3000/3/download"

    // Path lokal tempat file akan disimpan
    savePath := "../unzip/"

    // Panggil fungsi download
    if err := downloadFile(fileURL, savePath); err != nil {
        fmt.Printf("Error downloading file: %s\n", err)
    } else {
        fmt.Println("File downloaded successfully.")
    }

}

// unzipExtraktor akan mengekstrak file ZIP ke direktori tujuan.
func unzipExtractor(zipFile, destDir string) error {
    // Buka file ZIP untuk dibaca.
    r, err := zip.OpenReader(zipFile)
    if err != nil {
        return err
    }
    defer r.Close()

    os.MkdirAll(destDir, 0755)
    for _, f := range r.File {
        // Tentukan path file tujuan.
        filePath := filepath.Join(destDir, f.Name)

        // Buat direktori untuk file jika diperlukan.
        if f.FileInfo().IsDir() {
            os.MkdirAll(filePath, f.Mode())
            continue
        }

        // Buka file dalam arsip ZIP.
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer rc.Close()

        // Buat file tujuan.
        fDest, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }
        defer fDest.Close()

        // Salin isi file dari ZIP ke file tujuan.
        _, err = io.Copy(fDest, rc)
        if err != nil {
            return err
        }
    }

    return nil
}

// downloadFile fungsi untuk mendownload file dari URL tertentu dan menyimpannya ke path lokal.
func downloadFile(fileURL, savePath string) error {
    // Buat HTTP request
    resp, err := http.Get(fileURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Pastikan kita mendapatkan respons HTTP 200 OK.
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("server return non-200 status: %d %s", resp.StatusCode, resp.Status)
    }

    // Buat file lokal
    outFile, err := os.Create(savePath)
    if err != nil {
        return err
    }
    defer outFile.Close()

    // Salin data dari HTTP response ke file lokal
    _, err = io.Copy(outFile, resp.Body)
    return err // akan nil jika tidak ada error
}
