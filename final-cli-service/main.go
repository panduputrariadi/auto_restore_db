package main

import (
	"final-project/sekolahbeta-hacker/cli-service/config"
	"final-project/sekolahbeta-hacker/cli-service/controller"
	"final-project/sekolahbeta-hacker/cli-service/model"
	"fmt"
)

func main() {
	configs, err := config.BacaConfig("config.json")
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return
	}

	ch := make(chan model.DatabaseConfig)
	go func() {
		defer close(ch)
		for _, db := range configs {
			ch <- db
		}
	}()

	download := controller.DownloadFileWithWorker(ch, 2, "Mobile")

	unzip := controller.UnzipFileWithWorker(download, 2, "./unzip/")

	importDatabase := controller.ImportFileWithWorker(unzip, 2)

	deleteDirektori := controller.DeleteFileWithWorker(importDatabase, 2)
	for result := range deleteDirektori {
		if result.Error != nil {
			fmt.Println("Error:", result.Error)
		}
	}

}
