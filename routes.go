package main

import "net/http"

func (s *server) routes() {
	s.router.HandleFunc("/api/projects", s.handleAPIProjects())
	//s.router.HandleFunc("/", s.handleIndex())
}

func (s *server) handleAPIProjects() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		s.lock.RLock()
		defer s.lock.RUnlock()
		w.Write(s.projectsJSON)
	}
}
