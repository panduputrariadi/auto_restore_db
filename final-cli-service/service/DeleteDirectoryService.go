package service

import (
	"final-project/sekolahbeta-hacker/cli-service/model"
	"fmt"
	"os"
)

func GoroutineDeleteFile(chin chan model.DatabaseConfig) chan model.DatabaseConfig {
	chout := make(chan model.DatabaseConfig)

	go func() {
		defer close(chout)

		for db := range chin {

			if db.Error != nil {
				fmt.Println(db.Error.Error())
				chout <- db
				continue
			}

			err := os.RemoveAll(db.FileDownloaded)
			if err != nil {
				fmt.Printf("Error deleting file %s: %s\n", db.FileDownloaded, err)
				chout <- db
				continue
			}

			err = os.RemoveAll(db.FileSQL)
			if err != nil {
				fmt.Printf("Error deleting file %s: %s\n", db.FileSQL, err)
				chout <- db
				continue
			}

			chout <- db
		}
	}()

	return chout
}
