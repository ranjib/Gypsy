package command

import (
	"github.com/boltdb/bolt"
	"github.com/mitchellh/cli"
	"github.com/ranjib/gypsy/server"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type ServerCommand struct {
	Ui         cli.Ui
	httpServer *server.HttpServer
	poller     *server.Poller
}

func (c *ServerCommand) Help() string {
	return ""
}

func (c *ServerCommand) Synopsis() string {
	return "Runs Gypsy server"
}

func (c *ServerCommand) Run(args []string) int {
	config := c.readConfig()
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

func (c *ServerCommand) readConfig() *server.Config {
	return server.DefaultConfig()
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
