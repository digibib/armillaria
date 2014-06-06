package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	log "gopkg.in/inconshreveable/log15.v2"
)

// HTTP Handlers -------------------------------------------------------------

// serveFile serves a single file from disk.
func serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

// doResourceQuery sends a query to the RDF store SPARQL endpoint and returns the
// application/sparql-results+json  response.
func doResourceQuery(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	q := r.FormValue("query")
	if q == "" {
		http.Error(w, "missing required parameter: query", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "uninitialized RDF store", http.StatusInternalServerError)
		return
	}

	res, err := db.Query(q)
	if err != nil {
		l.Error("db.Query failed", log.Ctx{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, bytes.NewReader(res))
}

// addToIndex enqueues the requested URI to the indexing queue.
func addToIndex(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	uri := strings.TrimSpace(r.FormValue("uri"))
	if uri == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	queueAdd.WorkQueue <- indexRequest(uri)
	w.WriteHeader(http.StatusCreated)
}

// rmFromIndex enqueues the requested URI to the remove-from-index queue.
func rmFromIndex(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	uri := strings.TrimSpace(r.FormValue("uri"))
	if uri == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	queueRemove.WorkQueue <- indexRequest(uri)
	w.WriteHeader(http.StatusCreated)
}

// searchHandler proxies request to Elasticsearch.
func searchHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request, map[string]string) {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/search") + "/_search"
		p.ServeHTTP(w, r)
	}
}
