package main

import (
	"github.com/jtejido/hermes/cluster"
	"net/http"
)

type (
	server struct {
		logger cluster.Logging
		mux    *http.ServeMux
	}
)

func NewServer(options ...func(*server)) *server {
	s := &server{
		mux: http.NewServeMux(),
	}

	for _, f := range options {
		f(s)
	}

	s.mux.Handle(CachePath, loader(cacheIndexHandler(), logIt(s.logger)))
	s.mux.Handle(StatsCachePath, loader(statsIndexHandler(), logIt(s.logger)))
	s.mux.Handle(ClearCachePath, loader(clearIndexHandler(), logIt(s.logger)))
	s.mux.Handle(FilterClearCachePath, loader(clearFilterIndexHandler(), logIt(s.logger)))
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
