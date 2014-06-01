package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/digibib/armillaria/digest"
)

// DataSource represents a data source which is queryable over HTTP.
type dataSource interface {
	// Name returns the name of the data source.
	Name() string
	// Query takes a request and returns unparsed response body as a
	// byte slice, or an error. The request can be of any type.
	Query(r interface{}) ([]byte, error)
}

// authType represents the authentication method for a data source.
type authType int

const (
	authNone = iota
	authBasic
	authDigest
)

// dataSourceType represents a supported type of data source.
type dataSourceType int

const (
	sourceUnknown = iota
	sourceSPARQL
	sourceREST
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

func (s *localRDFStore) Name() string {
	return s.name
}

func (s *localRDFStore) Query(q interface{}) ([]byte, error) {
	form := url.Values{}
	form.Set("query", q.(string))
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

	res, err := s.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SPARQL http request failed: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// TODO proxy the *http.Response directly?
	return body, nil
}
