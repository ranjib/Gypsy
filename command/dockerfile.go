package command

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type DockerfileCommand struct {
	Meta
}

func (c *DockerfileCommand) Help() string {
	return "gypsy dockerfile [PIPELINE_NAME]"
}

func (c *DockerfileCommand) Synopsis() string {
	return "Create container from a dockerfile"
}

type BuilderState struct {
	Container *lxc.Container
	Env       []string
	Cwd       string
}

type Spec struct {
	ID         string
	File       string
	statements []string
	State      BuilderState
}

func (spec *Spec) Parse() error {
	fi, err := os.Open(spec.File)
	if err != nil {
		return err
	}
	defer fi.Close()
	scanner := bufio.NewScanner(fi)
	scanner.Split(bufio.ScanLines)
	var isComment = regexp.MustCompile(`^#`)
	var isExtendedStatement = regexp.MustCompile(`\\$`)
	previousStatement := ""
	for scanner.Scan() {
		line := scanner.Text()
		if isComment.MatchString(line) {
			log.Debug("Comment. bypassing")
			// dont process if line is comment
			continue
		} else if isExtendedStatement.MatchString(line) {
			log.Debug("Part of a multiline statement")
			// if line ends with \ then append statement
			if previousStatement != "" {
				previousStatement = previousStatement + " " + strings.TrimRight(line, "\\")
			} else {
				previousStatement = strings.TrimRight(line, "\\")
			}
		} else if strings.TrimSpace(line) == "" {
			log.Debug("Empty line. bypassing")
			// dont process if line empty
			continue
		} else {
			log.Debug("Statement completion. appending")
			// if line does not end with \ then append statement
			var statement string
			if previousStatement != "" {
				statement = previousStatement + " " + line
				previousStatement = ""
			} else {
				statement = line
			}
			spec.statements = append(spec.statements, statement)
		}
	}
	return nil
}

func (spec *Spec) Build() error {
	for _, statement := range spec.statements {
		log.Infof("Proecssing:|%s|\n", statement)
		words := strings.Fields(statement)
		switch words[0] {
		case "FROM":
			if spec.State.Container != nil {
				log.Errorf("Container already built. Multiple FROM declaration?\n")
				return errors.New("Container already built. Multiple FROM declaration?")
			}
			var err error
			spec.State.Container, err = util.CloneAndStartContainer(words[1], spec.ID)
			if err != nil {
				log.Errorf("Failed to clone container. Error: %s\n", err)
				return err
			}
		case "RUN":
			if spec.State.Container == nil {
				log.Error("No container has been created yet. Use FROM directive")
				return errors.New("No container has been created yet. Use FROM directive")
			}
			command := words[1:len(words)]
			log.Infof("Words: %#v\n", words)
			log.Infof("Attempting to execute: %#v\n", command)
			if err := spec.runCommand(command); err != nil {
				log.Errorf("Failed to run command inside container. Error: %s\n", err)
				return err
			}
		case "ENV":
			for i := 1; i < len(words); i++ {
				if strings.Contains(words[i], "=") {
					spec.State.Env = append(spec.State.Env, words[i])
				} else {
					spec.State.Env = append(spec.State.Env, words[i]+"="+words[i+1])
					i++
				}
			}
		case "WORKDIR":
			spec.State.Cwd = words[1]
		case "ADD":
			// setup bind mount
		case "COPY":
			// copy over files
		case "LABEL":
			// FIXME
		case "USER":
			// set exec user id
		case "VOLUME":
			// FIXME
		case "STOPSIGNAL":
			// FIXME
		case "MAINTAINER":
			// FIXME
		case "CMD":
			// FIXME
		case "ENTRYPOINT":
			// FIXME
		case "EXPOSE":
			// FIXME

		}
	}
	return nil
}

func (spec *Spec) runCommand(command []string) error {
	options := lxc.DefaultAttachOptions
	options.Cwd = "/root"
	options.Env = util.MinimalEnv()
	//spec.State.Env...)
	log.Infof("Exec environment: %#v\n", options.Env)
	rootfs := spec.State.Container.ConfigItem("lxc.rootfs")[0]
	var buffer bytes.Buffer
	buffer.WriteString("#!/bin/bash\n")
	for _, v := range spec.State.Env {
		if _, err := buffer.WriteString("export " + v + "\n"); err != nil {
			return err
		}
	}
	options.ClearEnv = true
	if spec.State.Cwd != "" {
		buffer.WriteString("cd " + spec.State.Cwd + "\n")
	}
	buffer.WriteString(strings.Join(command, " "))
	err := ioutil.WriteFile(filepath.Join(rootfs, "/tmp/dockerfile.sh"), buffer.Bytes(), 0755)
	if err != nil {
		log.Errorf("Failed to open file %s. Error: %v", err)
		return err
	}

	log.Infof("Executing:\n %s\n", buffer.String())
	exitCode, err := spec.State.Container.RunCommandStatus([]string{"/bin/bash", "/tmp/dockerfile.sh"}, options)
	if err != nil {
		log.Errorf("Failed to execute command: '%s'. Error: %v", command, err)
		return err
	}
	if exitCode != 0 {
		log.Warnf("Failed to execute command: '%s'. Exit code: %d", strings.Join(command, " "), exitCode)
	}
	return nil
}

func (c *DockerfileCommand) Run(args []string) int {
	id, err := util.UUID()
	if err != nil {
		log.Errorf("Failed to generate uuid. Error: %s\n", err)
		return -1
	}
	dockerfile := &Spec{File: "Dockerfile", statements: []string{}, ID: id}
	if err := dockerfile.Parse(); err != nil {
		log.Errorf("Failed to parse dockerfile. Error: %s\n", err)
		return -1
	}
	log.Debugf("Successfully parse dockerfile")
	if err := dockerfile.Build(); err != nil {
		log.Errorf("Failed to build container from dockerfile. Error: %s\n", err)
		return -1
	}
	log.Debugf("Successfully parse dockerfile")

	return 0
}
