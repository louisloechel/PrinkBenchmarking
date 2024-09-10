package main

import (
	"log"
	"os"
	"prinkbenchmarking/src/config"
	"prinkbenchmarking/src/evaluation"
	"prinkbenchmarking/src/types"
)



func experimentDone() {
	// Create experiment_done.txt (to be used later in tf script for stopping the cluster)
	file, err := os.Create("../results/experiment_done.txt")
	if err != nil {
		log.Fatalf("Could not create experiment_done.txt: %v", err)
	}
	defer file.Close()
}

func getExperiments() []types.Experiment {
	// k = [5,10,20,40,80]
	// delta = [1250,5000,20000,80000]
	// l diversity = [0,2,4,8] (wenn null bedeutet, dass l diversity nicht beachtet wird)
	// beta (active clusters). Damit das keinen Einfluss hat, so hoch setzen wie Daten: beta= 321728
	// mu= 100 (wie im original paper)

	k := []int{5,10,20,40,80}
	delta := []int{1250,5000,20000,80000}
	l := []int{0,2,4,8}
	beta := []int{321728,}
	mu := []int{100,}


	experiments := []types.Experiment{}

	for _, k := range k {
		for _, delta := range delta {
			for _, l := range l {
				for _, beta := range beta {
					for _, mu := range mu {
						experiment := types.Experiment{
							K: k,
							Delta: delta,
							L: l,
							Beta: beta,
							Zeta: 0,
							Mu: mu,
						}
						experiments = append(experiments, experiment)
					}
				}
			}
		}
	}

	return experiments
}


func main() {
	// Load the config
	config := config.LoadConfig()

	if len(os.Args) > 1 {
		if os.Args[1] == "listen" {
			// just listen
			experiment := types.Experiment{
				K: 5,
				Delta: 20000,
				L: 0,
				Beta: 321728,
				Zeta: 0,
				Mu: 100,
				SutHost: config.Address,
				SutPortWrite: config.PortWrite,
				SutPortRead: config.PortRead,
			}

			evaluation.RunSockets(&experiment, *config)
			return
		}
	}

	for i, experiment := range getExperiments() {
		experiment.SutHost = config.Address
		experiment.SutPortWrite = config.PortWrite
		experiment.SutPortRead = config.PortRead
		// Start the experiment
		log.Printf("Starting %d experiment: %v", i,  experiment)

		evaluation.RunExperiment(experiment, *config)
	}

	experimentDone()
	log.Printf("Created output files in: %s", config.OutputFolder)
}
