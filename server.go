package main

import (
	"bytes"
	"fmt"
	"io"
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

// HTTP Handlers -------------------------------------------------------------

// serveFile serves a single file from disk.
func serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

// loadResource expects 'uri' in the query-string, and fetches the given uri from
// the local RDF store. The application/sparql-results+json response from the
// SPARQL endpoint will be served.
func loadResource(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	uri := r.URL.Query()["uri"]
	if len(uri) == 0 || uri[0] == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "uninitialized RDF store", http.StatusInternalServerError)
		return
	}

	q := fmt.Sprintf(qGet, uri[0])
	res, err := db.Query(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, bytes.NewReader(res))
}

func createResource(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	uri := r.URL.Query()["uri"]
	if len(uri) == 0 || uri[0] == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "uninitialized RDF store", http.StatusInternalServerError)
		return
	}

	q := fmt.Sprintf(qGet, cfg.RDFStore.DefaultGraph, uri)
	res, err := db.Query(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, bytes.NewReader(res))
}

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
	mux.POST("/RDF/resource", createResource)
	//mux.PUT("/RDF/resource", updateResource)
	mux.HandlerFunc("GET", "/resource", serveFile("./data/html/resource.html"))
	mux.ServeFiles("/public/*filepath", http.Dir("./data/public/"))

	// Start server -------------------------------------------------------------
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServePort), mux))
}
