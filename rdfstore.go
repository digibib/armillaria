package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/digibib/armillaria/digest"
	"github.com/knakk/sparql"
)

// localRDFStore is the RDF store for local data.
type localRDFStore struct {
	name      string
	endpoint  string
	transport http.RoundTripper
}

func newLocalRDFStore(endpoint, username, password string) *localRDFStore {
	t := digest.NewTransport(username, password)
	l := localRDFStore{name: "local", endpoint: endpoint, transport: t}
	return &l
}

func query(s *localRDFStore, q string) (io.ReadCloser, error) {
	form := url.Values{}
	form.Set("query", q)
	form.Set("format", "application/sparql-results+json")
	b := form.Encode()

	req, err := http.NewRequest(
		"POST",
		s.endpoint,
		bytes.NewBufferString(b))
	if err != nil {
		// log err?
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))

	resp, err := s.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// temp log result body for debugging failed SPARQL requests
		// TODO remove when done
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		println(string(b))
		return nil, fmt.Errorf("SPARQL http request failed: %s", resp.Status)
	}
	return resp.Body, nil
}

// Query sends a query to localRDFStore's SPARQL endpoint, and returns the parsed
// results, or an error.
func (s *localRDFStore) Query(q string) (*sparql.Results, error) {
	body, err := query(s, q)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	results, err := sparql.ParseJSON(body)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Proxy sends a query to localRDFStore's SPARQL endpoint, and returns
// unparsed response body if the request was succesfull (200).
// It's the callers responsibility to close the response body.
func (s *localRDFStore) Proxy(q string) (io.ReadCloser, error) {
	body, err := query(s, q)
	if err != nil {
		return nil, err
	}
	return body, nil
}
