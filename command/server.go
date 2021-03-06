// Copyright 2015 Ranjib Dey.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"github.com/boltdb/bolt"
	"github.com/ranjib/gypsy/server"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type ServerCommand struct {
	Meta
	httpServer *server.HttpServer
	poller     *server.Poller
}

func (c *ServerCommand) Help() string {
	helpText := `
   Usage: gypsy server [-config gypsy.yml]"

	 General Options:
	` + generalOptionsUsage()
	return strings.TrimSpace(helpText)
}

func (c *ServerCommand) Synopsis() string {
	return "Runs Gypsy server"
}

func (c *ServerCommand) Run(args []string) int {
	var configFile string
	flags := c.Meta.FlagSet("server", FlagSetLog)
	flags.StringVar(&configFile, "config", "", "")
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		log.Errorf("Failed to parse cli arguments. Error: %s\n", err)
		return 1
	}
	var logOutput io.Writer
	if c.Meta.logOutput != "" {
		fi, err := os.OpenFile(c.Meta.logOutput, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Failed to open log output file '%s'. Error: %s\n", c.Meta.logOutput, err)
			return -1
		}
		defer fi.Close()
		logOutput = fi
	} else {
		logOutput = os.Stdout
	}
	util.ConfigureLogging(c.Meta.logLevel, c.Meta.logFormat, logOutput)
	var config *server.Config
	if configFile != "" {
		var err error
		config, err = server.ConfigFomeFile(configFile)
		if err != nil {
			log.Errorf("Failed to parse config file. Error: %s\n", err)
			return 1
		}
	} else {
		config = server.DefaultConfig()
	}
	if config == nil {
		return 1
	}
	c.setup(config)
	defer func() {
		if c.httpServer != nil {
			c.httpServer.Shutdown()
		}
	}()
	log.Println("Running Gypsy server")
	return c.handleSignals()
}

func (c *ServerCommand) setup(config *server.Config) error {
	if err := os.MkdirAll(filepath.Join(config.DataDir, "artifacts"), 0777); err != nil {

		log.Errorln(err)
		return err
	}
	db, err := bolt.Open(filepath.Join(config.DataDir, "gypsy.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorln(err)
		return err
	}
	db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("pipelines")); err != nil {
			log.Errorln(err)
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("artifacts")); err != nil {
			log.Errorln(err)
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("runs")); err != nil {
			log.Errorln(err)
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("pollingStatus")); err != nil {
			log.Errorln(err)
			return err
		}
		return nil
	})
	s, err := server.NewHttpServer(config.BindAddr, config.ArtifactDir, db)
	if err != nil {
		log.Errorln(err)
		return err
	}
	c.httpServer = s
	c.poller = server.NewPoller(config.PollingFrequency, db)
	return nil
}

func (c *ServerCommand) handleSignals() int {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Received SIGTERM, shutting down")
	return 0
}
