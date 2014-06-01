package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	//log "gopkg.in/inconshreveable/log15.v2"
)

// Global variables and constants---------------------------------------------
var (
	db  *localRDFStore
	cfg config
)

const (
	qGet = `SELECT * WHERE { <%s> ?p ?o }`
)

func main() {
	// Load configuration file --------------------------------------------------
	cfg, err := loadConfig("data/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Setup local repo ---------------------------------------------------------
	db = newLocalRDFStore(
		cfg.RDFStore.Endpoint,
		cfg.RDFStore.Username,
		cfg.RDFStore.Password)

	// Routing ------------------------------------------------------------------
	mux := httprouter.New()
	mux.GET("/RDF/resource", loadResource)
	mux.POST("/RDF/resource", doResourceQuery)
	mux.HandlerFunc("GET", "/resource", serveFile("./data/html/resource.html"))
	mux.ServeFiles("/public/*filepath", http.Dir("./data/public/"))

	// Start server -------------------------------------------------------------
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServePort), mux))
}
