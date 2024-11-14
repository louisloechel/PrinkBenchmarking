package evaluation

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	cfg "prinkbenchmarking/src/config"
	"prinkbenchmarking/src/prink"
	"prinkbenchmarking/src/types"
	"sync"
	"time"
)

func RunSockets(experiment* types.Experiment, config types.Config) {
	dataset := cfg.LoadDataset(config.InputData)

	var wg sync.WaitGroup

	// Increment the WaitGroup counter
	wg.Add(2)

	// write socket connection
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		if err := socketConnection(experiment, dataset); err != nil {
			log.Println("Error in socket connection: ", err)
		}
	}()

	// read socket connection
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		if err := readSocketConnection(experiment, config); err != nil {
			log.Println("Error in socket connection: ", err)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
}

func RunExperiment(experiment types.Experiment, config types.Config) {

	var wg sync.WaitGroup
	// Increment the WaitGroup counter
	wg.Add(2)

	// start prink
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		if err := prink.StartPrink(&experiment, config); err != nil {
			log.Println("Error in prink: ", err)
		}
	}()
	

	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		RunSockets(&experiment, config)
	}()

	ticker := time.NewTicker(time.Second)
	done := make(chan bool)

	go func() {
		var fg *prink.Flamegraph
		for {
			select {
			case <-done:
			  SaveFlamegraph(fg, &experiment, config)
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				flamegraph, err := prink.GetProfilingData(&experiment, config);
				if err != nil {
					log.Println("Error in prink profiling: ", err)
					continue
				}
				fg = flamegraph
			}
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	ticker.Stop()
	done <- true
}


func SaveFlamegraph(fg *prink.Flamegraph, experiment *types.Experiment, config types.Config) {
	// save flamegraph
	filename := fmt.Sprintf("%s/flamegraph-%s.%s.json", config.OutputFolder, time.Now().Format("2006-01-02.15:04:05"), experiment.ToFileName())
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Could not open flamegraph file: %v", err)
	}
	defer file.Close()

	json.NewEncoder(file).Encode(fg)
}