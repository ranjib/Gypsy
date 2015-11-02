package client

import (
	"github.com/ranjib/gypsy/structs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func BuildPipeline(name string) int {
	c := NewClient()
	pipeline, err1 := c.FetchPipeline(name)
	if err1 != nil {
		log.Errorf("Failed to fetch spec for pipeline %s. Error: %v", name, err1)
		return 1
	}
	container, err := c.CreateContainer(pipeline.Container)
	if err != nil {
		log.Errorf("Failed to create container for pipeline %s build. Error: %v", name, err)
		return 1
	}
	err = c.PerformBuild(pipeline.Scripts)
	if err != nil {
		log.Errorf("Failed to build pipeline %s. Error: %v", name, err)
		return 1
	}
	if len(pipeline.Artifacts) > 0 {
		err = c.UploadArtifacts(container, pipeline.Artifacts)
		if err != nil {
			log.Errorf("Failed to upload pipeline %s artifact. Error: %v", name, err)
			return 1
		}
	}
	err = c.DestroyContainer(container)
	if err != nil {
		log.Errorf("Failed to build pipeline %s. Error: %v", name, err)
		return 1
	}
	return 0
}

func (c *Client) FetchPipeline(name string) (*structs.Pipeline, error) {
	//TODO
	// use gorilla http client to get pipeline spec from server
	return nil, nil
}

func (c *Client) CreateContainer(name string) (*lxc.Container, error) {
	//TODO
	return nil, nil
}

func (c *Client) PerformBuild(scripts []string) error {
	//TODO
	return nil
}

func (c *Client) UploadArtifacts(container *lxc.Container, artifacts []structs.Artifact) error {
	//TODO
	return nil
}

func (c *Client) DestroyContainer(container *lxc.Container) error {
	//TODO
	return nil
}
