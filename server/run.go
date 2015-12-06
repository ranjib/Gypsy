package server

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	//	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/ranjib/gypsy/structs"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"

	"io/ioutil"
	"net/http"
	"strconv"
)

// REST: /pipelines/{pipeline_name}/runs
func (s *HttpServer) ListRuns(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	runs := []uint64{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("runs"))
		runBucket := b.Bucket([]byte(p))
		c := runBucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			runs = append(runs, util.Btoi(k))
		}
		log.Printf("List of runs: %v", runs)
		return nil
	})
	if err != nil {
		log.Errorf("Failed to list runs: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(runs)
	if err != nil {
		log.Errorf("Failed to marshal json: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(js)
}

// REST: /pipelines/{pipeline_name}/runs/{run_id}
func (s *HttpServer) ShowRun(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	r := mux.Vars(req)["run_id"]
	i, err := strconv.Atoi(r)
	if err != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}
	var run []byte
	err1 := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("runs"))
		runBucket := b.Bucket([]byte(p))
		run = runBucket.Get(util.Itob(uint64(i)))
		return nil
	})
	if err1 != nil {
		log.Errorf("Failed to list runs: %v", err1)
		http.Error(resp, err1.Error(), http.StatusInternalServerError)
		return
	}
	if run == nil {
		log.Warnf("No run found")
		http.Error(resp, "Not present", http.StatusNotFound)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(run)
}

// REST: /pipelines/{pipeline_name}/runs/{run_id}
func (s *HttpServer) UpdateRun(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	r := mux.Vars(req)["run_id"]
	i, err := strconv.Atoi(r)
	if err != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warnf("Failed to read request body : %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(string(body[:]))
	var run structs.Run
	if err := json.Unmarshal(body, &run); err != nil {
		log.Warnf("Failed to unmarshal request : %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}
	log.Infof("Run data validation succeeded, saving data")
	err1 := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("runs"))
		log.Infof("Bucket '%s' will be created if not exist", p)
		runBucket, e := b.CreateBucketIfNotExists([]byte(p))
		if e != nil {
			log.Errorln("Failed to create sub bucket")
			return e
		}
		return runBucket.Put(util.Itob(uint64(i)), body)
	})
	if err1 != nil {
		log.Warnf("Failed to store run details: %v", err1)
		http.Error(resp, err1.Error(), http.StatusInternalServerError)
		return
	}
}

// REST: /pipelines/{pipeline_name}/runs/{run_id}
func (s *HttpServer) DeleteRun(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	r := mux.Vars(req)["run_id"]
	i, err1 := strconv.Atoi(r)
	if err1 != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, err1)
		http.Error(resp, err1.Error(), http.StatusBadRequest)
		return
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("runs"))
		if b == nil {
			log.Errorf("Runs bucket not found")
			return fmt.Errorf("Runs bucket not found")
		}
		pipeline := b.Bucket([]byte(p))
		if pipeline == nil {
			log.Errorf("Run bucket for pipeline %s was not found", p)
			return fmt.Errorf("Run bucket for pipeline %s was not found", p)
		}
		log.Printf("Deleting run '%s' for pipeline: %s", r, p)
		return pipeline.Delete(util.Itob(uint64(i)))
	})
	if err != nil {
		log.Warnf("Failed to delete artifact '%s' for pipeline '%s'. Error: %v", r, p, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}
