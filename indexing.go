package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

type details struct {
	Type  string `json:"type,omitempty"`
	Index string `json:"index,omitempty"`
}

type mapping map[string]details

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

	_, err = i.client.Do(req)
	if err != nil {
		return err
	}

	return nil

}

// Remove removes a resource from an index.
func (i Indexer) Remove(idx string, tp string, uri string) error {
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
