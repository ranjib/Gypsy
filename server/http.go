package server

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type HttpServer struct {
	router   *mux.Router
	listener net.Listener
	addr     string
	db       *bolt.DB
}

func NewHttpServer(addr string, db *bolt.DB) (*HttpServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Warnf("Failed to create http listener object. Error: %v", err)
		return nil, err
	}
	log.Printf("Gypsy server HTTP endpoint started on: %s", addr)
	r := mux.NewRouter()
	srv := &HttpServer{
		router:   r,
		listener: ln,
		addr:     addr,
		db:       db,
	}
	srv.registerHandlers()
	go http.Serve(ln, r)
	return srv, nil
}

func (s *HttpServer) registerHandlers() {
	// Pipeline API
	s.router.HandleFunc("/pipelines", s.ListPipelines).Methods("GET")
	s.router.HandleFunc("/pipelines/{pipeline_name}", s.ShowPipeline).Methods("GET")
	s.router.HandleFunc("/pipelines", s.CreatePipeline).Methods("PUT")
	s.router.HandleFunc("/pipelines/{pipeline_name}", s.DeletePipeline).Methods("DELETE")
	s.router.HandleFunc("/pipelines/{pipeline_name}", s.UpdatePipeline).Methods("POST")

	// Run API
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs", s.ListRuns).Methods("GET")
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}", s.ShowRun).Methods("GET")
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}", s.UpdateRun).Methods("POST")
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}", s.DeleteRun).Methods("DELETE")

	// Artifact API
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}/artifacts", s.ListArtifacts).Methods("GET")
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}", s.DownloadArtifact).Methods("GET")
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}", s.UploadArtifact).Methods("PUT")
	s.router.HandleFunc("/pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}", s.DeleteArtifact).Methods("DELETE")
}

func (s *HttpServer) Shutdown() {
	log.Println("Shutting down Gypsy server")
	if s != nil {
		s.listener.Close()
		if s.db != nil {
			s.db.Close()
		}
	}
}
