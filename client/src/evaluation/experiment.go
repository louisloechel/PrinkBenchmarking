package evaluation

import (
	"log"
	cfg "prinkbenchmarking/src/config"
	"prinkbenchmarking/src/prink"
	"prinkbenchmarking/src/types"
	"sync"
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

	// Wait for all goroutines to finish
	wg.Wait()
}