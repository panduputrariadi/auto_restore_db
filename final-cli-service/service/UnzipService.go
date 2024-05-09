package service

import (
	"archive/zip"
	"final-project/sekolahbeta-hacker/cli-service/model"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GoRoutineUnzipFile(chin chan model.DatabaseConfig, destDir string) chan model.DatabaseConfig {
	chKeluar := make(chan model.DatabaseConfig)

	go func() {
		defer close(chKeluar)

		for config := range chin {
			if config.Error != nil {
				fmt.Println(config.Error.Error())
				chKeluar <- config
				continue
			}
			// fmt.Println(config.FileDownloaded)
			r, err := zip.OpenReader(config.FileDownloaded)
			if err != nil {
				config.Error = err
				fmt.Println(config.Error.Error())
				chKeluar <- config
				r.Close()
				continue
			}
			// defer r.Close()

			for _, f := range r.File {
				if filepath.Base(f.Name) == "__MACOSX" || strings.HasPrefix(filepath.Base(f.Name), "._") {
					continue
				}

				// code untuk ekstrak unzip file
				extractedFilePath := filepath.Join(destDir, filepath.Base(f.Name))

				//memberikan informasi lokasi ekstrak
				config.FileSQL = extractedFilePath

				if f.FileInfo().IsDir() {
					os.MkdirAll(extractedFilePath, f.Mode())
					continue
				}

				// jika file bukan sebuah direktori
				rc, err := f.Open()
				if err != nil {
					config.Error = err
					fmt.Println(config.Error.Error())
					chKeluar <- config
					continue
				}
				// defer rc.Close()

				// membuat destinasi file
				fDest, err := os.OpenFile(extractedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					config.Error = err
					fmt.Println(config.Error.Error())
					chKeluar <- config
					continue
				}
				// defer fDest.Close()

				// copy file menuju destinasinya
				_, err = io.Copy(fDest, rc)
				if err != nil {
					config.Error = err
					fmt.Println(config.Error.Error())
					chKeluar <- config
					continue
				}
			}

			chKeluar <- config
		}
	}()

	return chKeluar
}
