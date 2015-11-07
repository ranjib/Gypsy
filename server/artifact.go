package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (s *HttpServer) ListArtifacts(resp http.ResponseWriter, req *http.Request) {
	artifacts := []int{}
	p := mux.Vars(req)["pipeline_name"]
	err := s.db.View(func(tx *bolt.Tx) error {
		a := tx.Bucket([]byte("artifacts"))
		if a == nil {
			log.Errorf("Artifact bucket not found")
			return fmt.Errorf("Artifact bucket not found")
		}
		b := a.Bucket([]byte(p))
		if b == nil {
			log.Errorf("Artifact bucket for pipeline %s not found", p)
			return fmt.Errorf("Artifact bucket for pipeline %s not found", p)
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			artifacts = append(artifacts, int(btoi(k)))
		}
		log.Printf("List of artifacts %v", artifacts)
		return nil
	})
	if err != nil {
		log.Errorf("Failed to list pipelines: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(artifacts)
	if err != nil {
		log.Errorf("Failed to marshal json: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(js)
}

func (s *HttpServer) DownloadArtifact(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	a := mux.Vars(req)["artifact_id"]
	var artifact []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("artifacts"))
		if b == nil {
			log.Errorf("Artifacts bucket not found")
			return fmt.Errorf("Artifacts bucket not found")
		}
		pipeline := b.Bucket([]byte(p))
		if pipeline == nil {
			log.Errorf("Artifacts bucket for pipeline %s not found", p)
			return fmt.Errorf("Artifact bucket for pipeline %s not found", p)
		}
		log.Printf("Fetching artifact %s for pipeline %s", a, p)
		artifact = pipeline.Get([]byte(a))
		return nil
	})
	if err != nil {
		log.Warnf("Failed to show pipeline: %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	if artifact == nil {
		log.Warnf("No artifact found")
		http.Error(resp, "Not present", http.StatusNotFound)
		return
	}
	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-artifact_%s", p, a))
	resp.Write(artifact)
}

func (s *HttpServer) UploadArtifact(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warnf("Failed to read request body : %v", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	err1 := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("artifacts"))
		if b == nil {
			log.Errorf("Artifacts bucket not found")
			return fmt.Errorf("Artifacts bucket not found")
		}
		pipeline, err := b.CreateBucketIfNotExists([]byte(p))
		if err != nil {
			log.Errorf("Failed to create sub-bucket for pipeline %s under artifacts bucket. Error:%v", p, err)
			return err
		}
		if pipeline == nil {
			log.Errorf("Artifacts bucket for pipeline %s could not be created", p)
			return fmt.Errorf("Artifact bucket for pipeline %s could not be created", p)
		}
		id, _ := pipeline.NextSequence()
		log.Printf("Saving artifact '%d' for pipeline: %s", int(id), p)
		return pipeline.Put(itob(id), body)
	})
	if err1 != nil {
		log.Warnf("Failed to save artifact: %v", err1)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HttpServer) DeleteArtifact(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	a := mux.Vars(req)["artifact_id"]
	i, err1 := strconv.Atoi(a)
	if err1 != nil {
		log.Warnf("Failed to convert artifact if %s for pipeline '%s'. Error: %v", a, p, err1)
		http.Error(resp, err1.Error(), http.StatusBadRequest)
		return
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("artifacts"))
		if b == nil {
			log.Errorf("Artifacts bucket not found")
			return fmt.Errorf("Artifacts bucket not found")
		}
		pipeline := b.Bucket([]byte(p))
		if pipeline == nil {
			log.Errorf("Artifacts bucket for pipeline %s was not found", p)
			return fmt.Errorf("Artifact bucket for pipeline %s was not found", p)
		}
		log.Printf("Deleting artifact '%s' for pipeline: %s", a, p)
		return b.Delete(itob(uint64(i)))
	})
	if err != nil {
		log.Warnf("Failed to delete artifact '%s' for pipeline '%s'. Error: %v", a, p, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func itob(i uint64) []byte {
	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, i)
	return v
}

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
