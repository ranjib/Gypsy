package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ranjib/gypsy/structs"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Builder struct {
	ServerURL string
	Run       structs.Run
}

func NewBuilder(url, name string, runId int) *Builder {
	return &Builder{
		ServerURL: url,
		Run: structs.Run{
			ID:           runId,
			PipelineName: name,
		},
	}
}

func BuildPipeline(name string, runId int) int {
	c := NewBuilder("http://127.0.0.1:5678", name, runId)
	pipeline, err1 := c.FetchPipeline(name)
	if err1 != nil {
		log.Errorf("Failed to fetch spec for pipeline %s. Error: %v", name, err1)
		return 1
	}
	//c.devBuild(name, pipeline)
	c.nomadicBuild(pipeline, runId)
	log.Info("Successfully downloaded pipeline spec. Creating container for ", pipeline.Name)
	return 0
}

func (c *Builder) nomadicBuild(pipeline *structs.Pipeline, runId int) int {
	job, err := c.CreateNomadJob(pipeline, runId)
	if err != nil {
		log.Errorf("Failed to create nomad job for pipeline %s. Error: %v", pipeline.Name, err)
		return 1
	}
	return job.Run()
}

func (c *Builder) devBuild(name string, pipeline *structs.Pipeline) int {
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
	c.Run.Success = true
	c.PostRunData()
	return 0
}

func (c *Builder) FetchPipeline(name string) (*structs.Pipeline, error) {
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

func (c *Builder) CreateContainer(original string) (*lxc.Container, error) {
	cloned, err := util.UUID()
	if err != nil {
		log.Errorf("Failed to generate uuid. Error: %v", err)
		return nil, err
	}
	ct, err := util.CloneAndStartContainer(original, cloned)
	if err != nil {
		log.Errorf("Failed to clone container %s as %s. Error: %v", original, cloned, err)
		return nil, err
	}
	return ct, nil
}

func (c *Builder) PerformBuild(container *lxc.Container, commands []structs.Command) error {
	for _, cmd := range commands {
		var wg sync.WaitGroup
		stdoutReader, stdoutWriter, err := os.Pipe()
		outWriter := new(bytes.Buffer)
		errWriter := new(bytes.Buffer)
		if err != nil {
			log.Errorf("Failed to create pipe: %v", err)
			return err
		}
		stderrReader, stderrWriter, err := os.Pipe()
		if err != nil {
			log.Errorf("Failed to create pipe: %v", err)
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(outWriter, stdoutReader)
			if err != nil {
				log.Errorf("Failed to copy stdout. Error: %v", err)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(errWriter, stderrReader)
			if err != nil {
				log.Errorf("Failed to copy stderr. Error: %v", err)
			}
		}()

		log.Infof("Executing command: '%s'", cmd.Command)
		cwd := "/root"
		if cmd.Cwd != "" {
			cwd = cmd.Cwd
		}
		options := lxc.DefaultAttachOptions
		options.Env = util.MinimalEnv()
		options.StdoutFd = stdoutWriter.Fd()
		options.StderrFd = stderrWriter.Fd()
		options.ClearEnv = true
		options.Cwd = cwd
		exitCode, err := container.RunCommandStatus(strings.Fields(cmd.Command), options)
		if e := stdoutWriter.Close(); e != nil {
			log.Errorf("Failed to close stdout pipe. Error: %v", e)
		}
		if e := stderrWriter.Close(); e != nil {
			log.Errorf("Failed to close stderr pipe. Error: %v", e)
		}
		wg.Wait()
		c.Run.Stdout = strings.Join([]string{c.Run.Stdout, outWriter.String()}, "\n")
		c.Run.Stderr = strings.Join([]string{c.Run.Stderr, errWriter.String()}, "\n")
		if err != nil {
			log.Errorf("Failed to execute command: '%s'. Error: %v", cmd.Command, err)
			return err
		}
		if exitCode != 0 {
			log.Errorf("Failed to execute command: '%s'. Exit code: %d", cmd.Command, exitCode)
			return fmt.Errorf("Exit code:%d", exitCode)
		}
	}
	return nil
}

func (c *Builder) UploadArtifacts(container *lxc.Container, artifacts []structs.Artifact) error {
	for _, artifact := range artifacts {
		url := c.ServerURL + "/pipelines/" + c.Run.PipelineName + "/runs/" + strconv.Itoa(c.Run.ID) + "/artifacts/" + artifact.Name
		log.Infof("Making http post request against '%s' with run data", url)
		if err := util.PostFileFromContainer(container, artifact.Path, url); err != nil {
			log.Errorf("Failed to post artifact. Error: %v", err)
			return err
		}
	}
	return nil
}

func (c *Builder) PostRunData() error {
	httpClient := &http.Client{}
	payload, err := json.Marshal(c.Run)
	if err != nil {
		log.Errorf("Failed to marshal run data. Error: %v", err)
		return err
	}
	log.Info(string(payload[:]))
	run_id := strconv.Itoa(c.Run.ID)
	url := c.ServerURL + "/pipelines/" + c.Run.PipelineName + "/runs/" + run_id
	log.Infof("Making http post request against '%s' with run data", url)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		log.Errorf("Failed to create http request. Error: %v", err)
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("Failed to make http put request. Error: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Builder) DestroyContainer(container *lxc.Container) error {
	if err := container.Stop(); err != nil {
		log.Errorf("Failed to stop container %s. Error: %v", container.Name(), err)
		return err
	}
	return container.Destroy()
}
