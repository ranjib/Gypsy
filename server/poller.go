package server

import (
	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	"github.com/ranjib/gypsy/build"
	"github.com/ranjib/gypsy/structs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"strings"
	"time"
)

type Poller struct {
	Splay time.Duration
	db    *bolt.DB
}

func (p *Poller) Start() {
	for {
		log.Println("Beginning polling")
		p.poll()
		log.Println("Polling finished")
		time.Sleep(p.Splay)
	}
}

func NewPoller(splay int, db *bolt.DB) *Poller {
	poller := Poller{
		Splay: time.Duration(splay) * time.Second,
		db:    db,
	}
	go poller.Start()
	return &poller
}

func (p *Poller) poll() error {
	err := p.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pipelines"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Infof("Checking pipeline '%s' for changes", string(k[:]))
			var pipeline structs.Pipeline
			if err := yaml.Unmarshal(v, &pipeline); err != nil {
				log.Errorf("Failed to unmarshal yaml definition for pipeline %s. Error:%v", string(k[:]), v)
				continue
			}
			go p.checkMaterial(pipeline)
		}
		return nil
	})
	if err != nil {
		log.Errorf("Failed to list pipelines: %v", err)
		return err
	}
	return nil
}

func (p *Poller) checkMaterial(pipeline structs.Pipeline) {
	for _, material := range pipeline.Materials {
		switch material.Type {
		case "github":
			log.Infof("Checking github changes for pipeline %s", pipeline.Name)
			p.checkGithubMaterial(pipeline, material)
		default:
			log.Errorf("Unknown meterial type: %s", material.Type)
		}
	}
}

func (p *Poller) checkGithubMaterial(pipeline structs.Pipeline, material structs.Material) {
	fields := strings.Split(material.URI, "/")
	client := github.NewClient(nil)
	log.Infof("Getting current sha at %s/%s for '%s' pipeline", fields[0], fields[1], pipeline.Name)
	ref, _, err := client.Git.GetRef(fields[0], fields[1], "heads/master")
	if err != nil {
		log.Errorf("Error checking github material ref for %s pipeline. Error %v", pipeline.Name, err)
		return
	}
	var prevSHA []byte
	err1 := p.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pollingStatus"))
		prevSHA = b.Get([]byte(pipeline.Name))
		return nil
	})
	if err1 != nil {
		log.Errorf("Failed to fetch previous head SHA for pipeline: %s. Error: %v", pipeline.Name, err1)
		return
	}
	if prevSHA == nil || string(prevSHA[:]) != *ref.Object.SHA {
		log.Infof("Current SHA (%s) is different than  previously built SHA(%s). Triggering build", *ref.Object.SHA, string(prevSHA[:]))
		var runId uint64
		err := p.db.Update(func(tx *bolt.Tx) error {
			status := tx.Bucket([]byte("pollingStatus"))
			r := tx.Bucket([]byte("runs"))
			p, e := r.CreateBucketIfNotExists([]byte(pipeline.Name))
			if e != nil {
				log.Errorf("Failed to create pipeline specific run bucket")
				return e
			}
			runId, _ = p.NextSequence()
			return status.Put([]byte(pipeline.Name), []byte(*ref.Object.SHA))
		})
		if err != nil {
			log.Errorf("Failed to store current head SHA for pipeline: %s. Error: %v", pipeline.Name, err)
			return
		}
		exitCode := build.BuildPipeline(pipeline.Name, int(runId))
		log.Infof("Build exit code: %d", exitCode)
		return
	}
	log.Infof("Current SHA (%s) is same as previously built SHA. Skipping build", *ref.Object.SHA)
}
