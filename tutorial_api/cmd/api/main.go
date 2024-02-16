package main

import (
	"fmt"
	"net/http"

	"github.com/chhoumann/goapi/internal/handlers"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus" // alias to log
)

func main() {
	log.SetReportCaller(true)
	
	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	fmt.Println("Starting GO API service...")
	if err := http.ListenAndServe("localhost:8000", r); err != nil {
		log.Error(err)
	}
}