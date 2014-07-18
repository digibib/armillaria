package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/digibib/armillaria/sparql"
	"github.com/julienschmidt/httprouter"
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
func doResourceQuery(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, bytes.NewReader(res))
}

// addToIndex enqueues the requested URI to the indexing queue.
func addToIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	uri := strings.TrimSpace(r.FormValue("uri"))
	if uri == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	if q, err := queues.Get("addToIndex"); err == nil {
		q.WorkQueue <- indexRequest{uri: uri}
	}
	w.WriteHeader(http.StatusCreated)
}

// rmFromIndex enqueues the requested URI to the remove-from-index queue.
func rmFromIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	uri := strings.TrimSpace(r.FormValue("uri"))
	if uri == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	if q, err := queues.Get("rmFromIndex"); err == nil {
		q.WorkQueue <- indexRequest{uri: uri}
	}
	w.WriteHeader(http.StatusCreated)
}

// searchHandler proxies request to Elasticsearch.
func searchHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/search") + "/_search"
		p.ServeHTTP(w, r)
	}
}

// getIdHandler returns the next available ID number for a given RDF type.
func getIdHandler(w http.ResponseWriter, r *http.Request, values httprouter.Params) {
	n := idGen.NextId(values.ByName("type"))
	w.Write([]byte(strconv.Itoa(n)))
}

// queryExternalSource acts as a proxy for querying external sources.
func queryExternalSource(w http.ResponseWriter, r *http.Request, values httprouter.Params) {
	sourceName := values.ByName("source")
	if _, ok := cfg.ExternalDataSources[sourceName]; !ok {
		http.Error(w, "unknown external source", http.StatusBadRequest)
		return
	}
	source := cfg.ExternalDataSources[sourceName]
	switch source.Type {
	case sourceSPARQL:
		// Query external data source via SPARQL.
		// The query is sent as a POST request and requesting a response in json.
		var q = r.FormValue("query")
		if q == "" {
			http.Error(w, "missing required parameter: query", http.StatusBadRequest)
			return
		}
		form := url.Values{}
		form.Set("query", q)
		b := form.Encode()

		req, err := http.NewRequest(
			"POST",
			source.Endpoint,
			bytes.NewBufferString(b))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Accept", "application/sparql-results+json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Content-Length", strconv.Itoa(len(b)))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("SPARQL http request failed: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case sourceGET:
		// Query external data source by means of a simple GET query.
		// The query endpoint will require one value to be interpolated into the adress.
		// TODO how to make this more flexible, with variable number of arguments?
		var q = r.FormValue("query")
		if q == "" {
			http.Error(w, "missing required parameter: query", http.StatusBadRequest)
			return
		}
		res, err := http.Get(fmt.Sprintf(source.Endpoint, q))
		if err != nil {
			http.Error(w, fmt.Sprintf("http request failed: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case sourceREST:
		panic("not implemented")
	default:
		l.Error("unknown external source type", log.Ctx{"source": sourceName})

	}

}

// rdf2marcHandler serves a marcxml record for a given URI (only for manifestation).
func rdf2marcHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	uri := strings.TrimSpace(r.FormValue("uri"))
	if uri == "" {
		http.Error(w, "missing required parameter: uri", http.StatusBadRequest)
		return
	}

	res, err := db.Query(
		fmt.Sprintf(queryRDF2MARC, cfg.RDFStore.DefaultGraph, uri, uri))
	if err != nil {
		l.Error("db.Query failed", log.Ctx{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var rdf sparql.Results
	err = json.Unmarshal(res, &rdf)
	if err != nil {
		l.Error("unmarshal sparql json response failed", log.Ctx{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(rdf.Results.Bindings) == 0 {
		http.Error(w, "no resource with that URI", http.StatusNotFound)
		return
	}

	if rdf.Results.Bindings[0]["profile"].Value != "manifestation" {
		http.Error(w, "can only convert resources of type fabio:manifestation", http.StatusBadRequest)
		return
	}

	rec, err := convertRDF2MARC(rdf)
	if err != nil {
		l.Error("rdf2marc coversion failed", log.Ctx{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marcxml, err := xml.Marshal(rec)
	if err != nil {
		l.Error("marcxml marhsalling failed", log.Ctx{"error": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	io.Copy(w, bytes.NewReader(marcxml))
}
