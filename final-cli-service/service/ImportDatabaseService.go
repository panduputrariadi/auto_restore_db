package service

import (
	"bytes"
	"final-project/sekolahbeta-hacker/cli-service/model"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

func GoroutineImportFile(chin chan model.DatabaseConfig) chan model.DatabaseConfig {
	chout := make(chan model.DatabaseConfig)

	go func() {
		defer close(chout)

		for db := range chin {
			if db.Error != nil {
				chout <- db
				continue
			}

			importCmd := fmt.Sprintf("mysql -u %s -h %s -P %s %s < %s", db.Username, db.Host, db.Port, db.Name, db.FileSQL)
			var stdErr bytes.Buffer

			cmd := exec.Command("bash", "-c", importCmd)
			cmd.Stderr = &stdErr

			err := cmd.Run()
			if err != nil {
				chout <- model.DatabaseConfig{Error: fmt.Errorf("error executing command: %v\nstderr: %s", err, stdErr.String())}
				continue
			}

			chout <- db
		}
	}()

	return chout
}

func ReadSql(destDir string) ([]string, error) {
	// Mencari file SQL dalam direktori tujuan
	fileList, err := ioutil.ReadDir(destDir)
	if err != nil {
		return nil, err
	}

	var sqlFiles []string

	for _, f := range fileList {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(destDir, f.Name()))
		}
	}

	if len(sqlFiles) == 0 {
		return nil, fmt.Errorf("no SQL files found in the destination directory")
	}

	return sqlFiles, nil
}
