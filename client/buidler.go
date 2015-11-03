package client

import (
	"fmt"
	"github.com/ranjib/gypsy/structs"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client struct {
	ServerURL string
}

func NewClient(url string) *Client {
	return &Client{ServerURL: url}
}

func BuildPipeline(name string) int {
	c := NewClient("http://127.0.0.1:5678")
	pipeline, err1 := c.FetchPipeline(name)
	if err1 != nil {
		log.Errorf("Failed to fetch spec for pipeline %s. Error: %v", name, err1)
		return 1
	}
	log.Info("Successfully downloaded pipeline spec for ", pipeline.Name)
	container, err := c.CreateContainer(pipeline.Container)
	if err != nil {
		log.Errorf("Failed to create container for pipeline %s build. Error: %v", name, err)
		return 1
	}
	err = c.PerformBuild(container, pipeline.Scripts)
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
	resp, err := http.Get(c.ServerURL + "/pipelines/" + name)
	if err != nil {
		log.Errorf("Failed to fetch pipeline spec from server. Error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body. Error: %v", err)
		return nil, err
	}
	pipeline := new(structs.Pipeline)
	if err := yaml.Unmarshal(body, pipeline); err != nil {
		log.Warnf("Failed to unmarshal request : %v", err)
		return nil, err
	}
	return pipeline, nil
}

func (c *Client) CreateContainer(original string) (*lxc.Container, error) {
	cloned, err := util.UUID()
	if err != nil {
		log.Errorf("Failed to generate uuid. Error: %v", err)
		return nil, err
	}
	orig, err := lxc.NewContainer(original)
	if err != nil {
		log.Errorf("Failed to initialize container object. Error: %v", err)
		return nil, err
	}
	if err := orig.Clone(cloned, lxc.CloneOptions{}); err != nil {
		log.Errorf("Failed to clone container %s as %s. Error: %v", original, cloned, err)
		return nil, err
	}
	ct, err := lxc.NewContainer(cloned)
	if err != nil {
		log.Errorf("Failed to clone container %s as %s. Error: %v", original, cloned, err)
		return nil, err
	}
	if err := ct.Start(); err != nil {
		log.Errorf("Failed to start cloned container %s. Error: %v", cloned, err)
		return nil, err
	}
	return ct, nil
}

func (c *Client) PerformBuild(container *lxc.Container, scripts []string) error {
	for _, cmd := range scripts {
		log.Infof("Executing command: '%s'", cmd)
		exitCode, err := container.RunCommandStatus(strings.Fields(cmd), lxc.AttachOptions{})
		if err != nil {
			log.Infof("Failed to execute command: '%s'. Error: %v", cmd, err)
			return err
		}
		if exitCode != 0 {
			log.Infof("Failed to execute command: '%s'. Exit code: %d", cmd, exitCode)
			return fmt.Errorf("Exit code:%d", exitCode)
		}
	}
	return nil
}

func (c *Client) UploadArtifacts(container *lxc.Container, artifacts []structs.Artifact) error {
	//TODO
	return nil
}

func (c *Client) DestroyContainer(container *lxc.Container) error {
	if err := container.Stop(); err != nil {
		log.Errorf("Failed to stop container %s. Error: %v", container.Name(), err)
		return err
	}
	return container.Destroy()
}
