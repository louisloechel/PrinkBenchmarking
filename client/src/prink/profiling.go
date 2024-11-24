package prink

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prinkbenchmarking/src/types"
)



type JobOverview struct {
	Jobs []struct {
		JID            string `json:"jid"`
		Name           string `json:"name"`
		StartTime      int    `json:"start-time"`
		EndTime        int    `json:"end-time"`
		Duration       int    `json:"duration"`
		State          string `json:"state"`
		LastModification int    `json:"last-modification"`
		Tasks          struct {
			Running      int `json:"running"`
			Canceling    int `json:"canceling"`
			Canceled     int `json:"canceled"`
			Total        int `json:"total"`
			Created      int `json:"created"`
			Scheduled    int `json:"scheduled"`
			Deploying    int `json:"deploying"`
			Reconciling  int `json:"reconciling"`
			Finished     int `json:"finished"`
			Initializing int `json:"initializing"`
			Failed       int `json:"failed"`
		} `json:"tasks"`
	} `json:"jobs"`
}

type Vertex struct {
	ID string `json:"id"`
	SlotSharingGroupId string `json:"slotSharingGroupId"`
	Name string `json:"name"`
	MaxParallelism int `json:"maxParallelism"`
	Parallelism int `json:"parallelism"`
	Status string `json:"status"`
	StartTime int `json:"start-time"`
	EndTime int `json:"end-time"`
	Duration int `json:"duration"`
}

         
type JobDetails struct {
	JID string `json:"jid"`
	Name string `json:"name"`
	IsStoppable bool `json:"isStoppable"`
	State string `json:"state"`
	JobType string `json:"job-type"`
	StartTime int `json:"start-time"`
	EndTime int `json:"end-time"`
	Duration int `json:"duration"`
	MaxParallelism int `json:"maxParallelism"`
	Now int `json:"now"`
	Vertices []Vertex `json:"vertices"`
}

type Flamegraph struct {
	Name string `json:"name"`
	Value int `json:"value"`

	Children []Flamegraph `json:"children"`
}

type FlamegraphResponse struct {
	EndTimestamp int `json:"endTimestamp"`
	Data Flamegraph `json:"data"`
}


func GetProfilingData(experiment *types.Experiment, config types.Config) (*Flamegraph, error) {

	job_response, err := http.Get("http://" + experiment.SutHost + ":8081/jobs/overview")
	if err != nil {
		return nil, err
	}

	jobs := JobOverview{}
	// json
	json.NewDecoder(job_response.Body).Decode(&jobs)

	if len(jobs.Jobs) == 0 || jobs.Jobs[0].State != "RUNNING" {
		return nil, fmt.Errorf("job not finished")
	}
	
	// get the job id
	jobId := jobs.Jobs[0].JID

	job_details_response, err := http.Get("http://" + experiment.SutHost + ":8081/jobs/" + jobId)
	if err != nil {
		return nil, err
	}

	jobDetails := JobDetails{}
	// json
	json.NewDecoder(job_details_response.Body).Decode(&jobDetails)

	// get the vertex with the name starting with 'k'
	var vertex Vertex
	for _, v := range jobDetails.Vertices {
		if v.Name[0] == 'k' {
			vertex = v
			break
		}
	}

	//http://localhost:8081/jobs/145c3014963ee48bca4954e75c2ae369/vertices/4150b807e25f98bebfeb73f2fab67d53/flamegraph?type=on_cpu

	// get the flamegraph
	flamegraph_response, err := http.Get("http://" + experiment.SutHost + ":8081/jobs/" + jobId + "/vertices/" + vertex.ID + "/flamegraph?type=on_cpu")	
	if err != nil {
		return nil, err
	}

	// json
	flamegraph := FlamegraphResponse{}
	json.NewDecoder(flamegraph_response.Body).Decode(&flamegraph)

	return &flamegraph.Data, nil
}
