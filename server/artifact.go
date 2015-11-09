package server

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// REST: /pipelines/{pipeline_name}/runs/{run_id}/artifacts
func (s *HttpServer) ListArtifacts(resp http.ResponseWriter, req *http.Request) {
	artifacts := []string{}
	p := mux.Vars(req)["pipeline_name"]
	r := mux.Vars(req)["run_id"]
	i, e := strconv.Atoi(r)
	if e != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, e)
		http.Error(resp, e.Error(), http.StatusBadRequest)
		return
	}
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
		runBucket := b.Bucket(util.Itob(uint64(i)))
		if runBucket == nil {
			log.Errorf("Sub0bucket for pipeline %s's run id %d  not found", p, i)
			return fmt.Errorf("Sub0bucket for pipeline %s's run id %d  not found", p, i)
		}
		c := runBucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			artifacts = append(artifacts, string(k[:]))
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

//REST: /pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}
func (s *HttpServer) DownloadArtifact(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	a := mux.Vars(req)["artifact_name"]
	r := mux.Vars(req)["run_id"]
	i, e := strconv.Atoi(r)
	if e != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, e)
		http.Error(resp, e.Error(), http.StatusBadRequest)
		return
	}
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
		runBucket := pipeline.Bucket(util.Itob(uint64(i)))
		if runBucket == nil {
			log.Errorf("Artifact bucket for pipeline %s for run id %d not found", p, i)
			return fmt.Errorf("Artifact bucket for pipeline %s for run id %d not found", p, i)
		}
		log.Printf("Fetching artifact %s for pipeline %s", a, p)
		artifact = pipeline.Get([]byte(a))
		return nil
	})
	if err != nil {
		log.Warnf("Failed to fetch artifact : %v", err)
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

// REST: 	/pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}
func (s *HttpServer) UploadArtifact(resp http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(32 << 20)
	p := mux.Vars(req)["pipeline_name"]
	r := mux.Vars(req)["run_id"]
	a := mux.Vars(req)["artifact_name"]
	i, err := strconv.Atoi(r)
	if err != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := req.FormFile("artifact")
	if err != nil {
		log.Warnf("Failed in form file invocation '%s'. Error: %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	fmt.Fprintf(resp, "%v", handler.Header)
	artifactPath := filepath.Join(s.artifactLocation, p, r, a)
	if e := os.MkdirAll(filepath.Dir(artifactPath), 0775); e != nil {
		log.Errorf("Failed to create artifact directory %s. Error: %v", filepath.Dir(artifactPath), e)
		http.Error(resp, e.Error(), http.StatusInternalServerError)
		return
	}
	fw, e1 := os.OpenFile(artifactPath, os.O_WRONLY|os.O_CREATE, 0666)
	if e1 != nil {
		log.Warnf("Failed in create file '%s'. Error: %v", artifactPath, e1)
		http.Error(resp, e1.Error(), http.StatusInternalServerError)
		return
	}
	defer fw.Close()
	_, e := io.Copy(fw, file)
	if e != nil {
		log.Warnf("Failed in copy file '%s'. Error: %v", artifactPath, e)
		http.Error(resp, e.Error(), http.StatusInternalServerError)
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
		runBucket, err := pipeline.CreateBucketIfNotExists(util.Itob(uint64(i)))
		if err != nil {
			log.Errorf("Failed to create sub-bucket for pipeline %s under artifacts bucket. Error:%v", p, err)
			return err
		}
		log.Printf("Saving artifact '%s' for pipeline: %s run id %d", a, p, i)
		return runBucket.Put([]byte(a), []byte(artifactPath))
	})
	if err1 != nil {
		log.Warnf("Failed to save artifact: %v", err1)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

// REST: /pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}
func (s *HttpServer) DeleteArtifact(resp http.ResponseWriter, req *http.Request) {
	p := mux.Vars(req)["pipeline_name"]
	r := mux.Vars(req)["run_id"]
	a := mux.Vars(req)["artifact_name"]
	i, e := strconv.Atoi(r)
	if e != nil {
		log.Warnf("Failed to convert run id %s for pipeline '%s'. Error: %v", r, p, e)
		http.Error(resp, e.Error(), http.StatusBadRequest)
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
		return pipeline.Delete(util.Itob(uint64(i)))
	})
	if err != nil {
		log.Warnf("Failed to delete artifact '%s' for pipeline '%s'. Error: %v", a, p, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}
