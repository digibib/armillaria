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

func (s *localRDFStore) Query(q string) ([]byte, error) {
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

	res, err := s.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		// log result body?
		b, _ := ioutil.ReadAll(res.Body)
		println(string(b))
		return nil, fmt.Errorf("SPARQL http request failed: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// TODO proxy the *http.Response directly?
	// look into ReverseProxy at http://golang.org/pkg/net/http/httputil/
	return body, nil
}
