package main

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type server struct {
	db           *sql.DB
	router       mux.Router
	projectsJSON []byte
	lock         sync.RWMutex
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
