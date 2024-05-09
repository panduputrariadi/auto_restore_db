package controller

import (
	"final-project/sekolahbeta-hacker/cli-service/model"
	"final-project/sekolahbeta-hacker/cli-service/service"
	"sync"
)

func ImportFileWithWorker(chin chan model.DatabaseConfig, worker int) chan model.DatabaseConfig {
	channels := []chan model.DatabaseConfig{}

	chout := make(chan model.DatabaseConfig)

	wg := sync.WaitGroup{}

	wg.Add(worker)

	go func() {
		wg.Wait()
		close(chout)
	}()
	// Fan-in
	for i := 0; i < worker; i++ {
		channels = append(channels, service.GoroutineImportFile(chin))
	}

	// Fan-out
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
