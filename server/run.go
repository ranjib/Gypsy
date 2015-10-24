package server

import (
	"net/http"
)

type Run struct {
	ID         string
	PipelineID string
	Stdout     string
	Stderr     string
	Status     bool
}

func (s *HttpServer) ListRuns(resp http.ResponseWriter, req *http.Request) {
}

func (s *HttpServer) ShowRun(resp http.ResponseWriter, req *http.Request) {
}

func (s *HttpServer) CreateRun(resp http.ResponseWriter, req *http.Request) {
}

func (s *HttpServer) DeleteRun(resp http.ResponseWriter, req *http.Request) {
}

func (s *HttpServer) UpdateRun(resp http.ResponseWriter, req *http.Request) {
}
