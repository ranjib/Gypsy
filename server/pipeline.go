package server

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/ranjib/gypsy/structs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

func (s *HttpServer) ListPipelines(resp http.ResponseWriter, req *http.Request) {
	pipelines := []string{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pipelines"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			pipelines = append(pipelines, string(k[:]))
		}
		log.Printf("List of pipelines: %v", pipelines)
		return nil
	})
	if err != nil {
		log.Errorf("Failed to list pipelines: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(pipelines)
	if err != nil {
		log.Errorf("Failed to marshal json: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(js)
}

func (s *HttpServer) ShowPipeline(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	var pipeline []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pipelines"))
		pipeline = b.Get([]byte(p))
		log.Printf("Showing pipeline: %v", p)
		return nil
	})
	if err != nil {
		log.Warnf("Failed to show pipeline: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	if pipeline == nil {
		log.Warnf("No pipeline found")
		http.Error(resp, "Not present", http.StatusNotFound)
		return
	}
	resp.Header().Set("Content-Type", "application/yaml")
	resp.Write(pipeline)
}

func (s *HttpServer) CreatePipeline(resp http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warnf("Failed to read request body : %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(string(body[:]))
	var pipeline structs.Pipeline
	if err := yaml.Unmarshal(body, &pipeline); err != nil {
		log.Warnf("Failed to unmarshal request : %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.ParseForm(); err != nil {
		log.Warnf("Failed to parse form: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(req.FormValue("pipeline_name"))
	err1 := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pipelines"))
		log.Printf("Creating pipeline: %s", pipeline.Name)
		return b.Put([]byte(pipeline.Name), body)
	})
	if err1 != nil {
		log.Warnf("Failed to create pipeline: %v", err1)
		http.Error(resp, err1.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HttpServer) DeletePipeline(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pipelines"))
		log.Printf("Deleting pipeline: %s", p)
		return b.Delete([]byte(p))
	})
	if err != nil {
		log.Warnf("Failed to delete pipeline: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HttpServer) UpdatePipeline(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	p := vars["pipeline_name"]
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pipelines"))
		log.Printf("Updating pipeline: %s", p)
		return b.Put([]byte(p), []byte("42"))
	})
	if err != nil {
		log.Warnf("Failed to update pipeline: %v", err)
	}
}
