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
			for ki, _ := range v {
				allMappings[profile][k] = ki
			}
		}

		for k, v := range m {
			for ki, _ := range v {
				allMappings[profile][k] = ki
			}
		}
	}
	return allMappings, nil
}
