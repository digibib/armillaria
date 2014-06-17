package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/digibib/armillaria/sparql"
)

// esError represents an error returned from a Elasticsearch REST endpoint.
type esError struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}
type mapping map[string]interface{}

type preMappings map[string]mapping

// Indexer is a simple wrapping around Elasticsearch
type Indexer struct {
	host   string
	client *http.Client
}

// Add adds anresource to an index.
func (i Indexer) Add(idx string, tp string, b []byte) error {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/%s/%s", i.host, idx, tp),
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 300 || resp.StatusCode < 200 {
		var esErr esError
		e := "unparsable error response from Elasticsearch"
		err := json.Unmarshal(b, &esErr)
		if err == nil {
			e = esErr.Error
		}
		return errors.New(e)
	}
	return nil

}

// Remove removes a resource from an index.
func (i Indexer) Remove(uri string) error {
	// Delete resource by query
	var queryData bytes.Buffer
	queryData.Write([]byte(`{"query":{"ids":{"values":["`))
	queryData.Write([]byte(uri))
	queryData.Write([]byte(`"]}}}`))

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/public/_query", i.host),
		&queryData,
	)
	if err != nil {
		return err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 300 || resp.StatusCode < 200 {
		var esErr esError
		e := "unparsable error response from Elasticsearch"
		err := json.Unmarshal(b, &esErr)
		if err == nil {
			e = esErr.Error
		}
		return errors.New(e)
	}
	return nil
}

// loadFromProfiles loads the profile mapping files and constructs a map
// which holds mappings from predicates to elasticsearch properties,
// for all types (profiles).
func loadFromProfiles() (map[string]map[string]string, error) {
	cb, err := ioutil.ReadFile("data/mappings/_common")
	if err != nil {
		return nil, err
	}
	var common preMappings
	err = json.Unmarshal(cb, &common)
	if err != nil {
		return nil, err
	}

	files, err := filepath.Glob("data/mappings/*.json")
	if err != nil {
		return nil, err
	}
	//                  map[profile]map[predicate]field
	allMappings := make(map[string]map[string]string)
	for _, f := range files {
		profile := strings.TrimSuffix(strings.TrimPrefix(f, "data/mappings/"), ".json")
		allMappings[profile] = make(map[string]string)

		var m preMappings
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}

		for k, v := range common {
			for ki := range v {
				allMappings[profile][k] = ki
			}
		}

		for k, v := range m {
			for ki := range v {
				allMappings[profile][k] = ki
			}
		}
	}
	return allMappings, nil
}

// createIndexedDoc creates a json document for a RDF resource, given
// a set of predicate to field mappings and a sparql json result.
func createIndexDoc(mappings map[string]map[string]string, sparqlRes []byte, uri string) ([]byte, string, error) {
	var res *sparql.Results
	var profile string
	err := json.Unmarshal(sparqlRes, &res)
	if err != nil {
		return nil, profile, errors.New("failed to parse sparql response")

	}
	// fetch the resource profile from the SPARQL responsevar profile string
	for _, b := range res.Results.Bindings {
		if b["p"].Value == "armillaria://internal/profile" {
			profile = b["o"].Value
			break
		}
	}
	if profile == "" {
		return nil, profile, errors.New("resource lacks profile information")
	}

	resource := make(map[string]interface{})
	type uriField struct {
		URI   string `json:"uri"`
		Label string `json:"label"`
	}
	var pred string
	uf := uriField{}
	println(profile)

	for _, b := range res.Results.Bindings {
		pred = urlify(b["p"].Value)
		if mappings[profile][pred] == "" {
			continue // if not in mapping, we don't want to index it
		}
		if _, ok := b["l"]; ok {
			uf.URI = b["o"].Value
			uf.Label = b["l"].Value
			switch resource[mappings[profile][pred]].(type) {
			case []interface{}:
				resource[mappings[profile][pred]] =
					append(resource[mappings[profile][pred]].([]interface{}), uf)
			case uriField:
				var s []interface{}
				s = append(s, resource[mappings[profile][pred]])
				resource[mappings[profile][pred]] = append(s, uf)
			default:
				resource[mappings[profile][pred]] = uf
			}
			continue
		}

		val := b["o"].Value
		switch resource[mappings[profile][pred]].(type) {
		case []interface{}:
			resource[mappings[profile][pred]] =
				append(resource[mappings[profile][pred]].([]interface{}), val)
		case interface{}:
			var s []interface{}
			s = append(s, resource[mappings[profile][pred]])
			resource[mappings[profile][pred]] = append(s, val)
		default:
			resource[mappings[profile][pred]] = val
		}

	}

	// We want to use the URI as the elasticsearch document ID
	resource["uri"] = uri

	resourceBody, err := json.Marshal(resource)
	if err != nil {
		return nil, profile, errors.New("failed to marshal json")
	}

	return resourceBody, profile, nil
}
