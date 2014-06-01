package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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
		println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, bytes.NewReader(res))
}
