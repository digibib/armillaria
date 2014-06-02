package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Global variables and constants---------------------------------------------
var (
	db  *localRDFStore
	cfg config
	l   = log.New()
)

func main() {
	// Load configuration file --------------------------------------------------
	cfg, err := loadConfig("data/config.json")
	if err != nil {
		l.Error("failed to load config.json", log.Ctx{"details": err.Error()})
		os.Exit(1)
	}

	// Setup local repo ---------------------------------------------------------
	db = newLocalRDFStore(
		cfg.RDFStore.Endpoint,
		cfg.RDFStore.Username,
		cfg.RDFStore.Password)

	// Routing ------------------------------------------------------------------
	mux := httprouter.New()
	mux.POST("/RDF/resource", doResourceQuery)
	mux.HandlerFunc("GET", "/resource", serveFile("./data/html/resource.html"))
	mux.ServeFiles("/public/*filepath", http.Dir("./data/public/"))

	// Start server -------------------------------------------------------------
	l.Info("starting Armillaria server")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServePort), mux)
	if err != nil {
		l.Error("http server crashed", log.Ctx{"details": err.Error()})
	}
}
