package controller

import (
	"final-project/sekolahbeta-hacker/cli-service/model"
	"final-project/sekolahbeta-hacker/cli-service/service"
	"sync"
)

func DownloadFileWithWorker(chin chan model.DatabaseConfig, worker int, clientHeader string) chan model.DatabaseConfig {
	channels := []chan model.DatabaseConfig{}

	chout := make(chan model.DatabaseConfig)

	wg := sync.WaitGroup{}

	wg.Add(worker)

	go func() {
		wg.Wait()
		close(chout)
	}()

	for i := 0; i < worker; i++ {
		channels = append(channels, service.GoRoutineDownloadFile(chin, clientHeader))

	}

	for _, ch := range channels {
		go func(channel chan model.DatabaseConfig) {
			for c := range channel {
				chout <- c
			}

			wg.Done()
		}(ch)

	}

	return chout
}
