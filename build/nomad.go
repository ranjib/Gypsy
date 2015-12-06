package build

import (
	"bytes"
	"encoding/gob"
	"fmt"
	nomadApi "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/ranjib/gypsy/structs"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type NomadJob struct {
	Pipeline *structs.Pipeline
	Job      *nomadStructs.Job
}

func (c *Builder) CreateNomadJob(pipeline *structs.Pipeline, runId int) (*NomadJob, error) {
	config := make(map[string]interface{})
	config["container"] = pipeline.Container
	config["pipeline"] = pipeline.Name
	config["run_id"] = strconv.Itoa(runId)
	config["server_url"] = c.ServerURL
	resources := &nomadStructs.Resources{
		CPU:      1024,
		MemoryMB: 128,
	}
	task := &nomadStructs.Task{
		Name:      pipeline.Name,
		Driver:    "gypsy",
		Config:    config,
		Resources: resources,
	}
	group := &nomadStructs.TaskGroup{
		Name:          pipeline.Name,
		Count:         1,
		Tasks:         []*nomadStructs.Task{task},
		RestartPolicy: nomadStructs.NewRestartPolicy("batch"),
	}
	job := &nomadStructs.Job{
		ID:          pipeline.Name,
		Name:        pipeline.Name,
		Region:      "global",
		Priority:    50,
		Datacenters: []string{"dc1"},
		Type:        "batch",
		TaskGroups:  []*nomadStructs.TaskGroup{group},
	}
	if err := job.Validate(); err != nil {
		log.Errorf("Nomad job validation failed. Error: %s\n", err)
		return nil, err
	}
	apiJob, err := convertJob(job)
	if err != nil {
		log.Errorf("Failed to convert nomad job in api call. Error: %s\n", err)
		return nil, err
	}
	nomadConfig := nomadApi.DefaultConfig()
	nomadClient, err := nomadApi.NewClient(nomadConfig)
	if err != nil {
		log.Errorf("Error creating nomad api client: %s", err)
		return nil, fmt.Errorf(fmt.Sprintf("Error creating nomad api client: %s", err))
	}
	evalId, _, nomadErr := nomadClient.Jobs().Register(apiJob, nil)
	if nomadErr != nil {
		log.Errorf("Error submitting job: %s", nomadErr)
		return nil, fmt.Errorf(fmt.Sprintf("Error submitting job: %s", nomadErr))
	}
	log.Infof("Syccessfullt submitted nomad job. Eval id: %s\n", evalId)
	return &NomadJob{
		Pipeline: pipeline,
		Job:      job,
	}, nil
}

func (job *NomadJob) Run() int {
	log.Infof("Submitting nomad job for pipeline: %s\n", job.Pipeline.Name)
	return 0
}

func convertJob(in *nomadStructs.Job) (*nomadApi.Job, error) {
	gob.Register([]map[string]interface{}{})
	gob.Register([]interface{}{})
	var apiJob *nomadApi.Job
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(in); err != nil {
		return nil, err
	}
	if err := gob.NewDecoder(buf).Decode(&apiJob); err != nil {
		return nil, err
	}
	return apiJob, nil
}
