package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

// config holds the application configuration variables.
type config struct {
	ServePort           int
	LogFile             string
	RDFStore            source
	ExternalDataSources map[string]source
	Elasticsearch       string
	KohaPath            string
	KohaSyncUser        string
	KohaSyncPass        string
	SyncToKoha          bool
}

type source struct {
	Endpoint          string
	DefaultGraph      string
	DraftsGraph       string
	InternalNameSpace string
	Username          string
	Password          string
	Token             string
	Type              dataSourceType
}

// UnmarshalText implementation for dataSourceType.
func (t *dataSourceType) UnmarshalText(b []byte) error {
	s := strings.ToLower(strings.Trim(string(b), "\""))

	switch {
	case s == "sparql":
		*t = sourceSPARQL
	case s == "rest":
		*t = sourceREST
	case s == "get":
		*t = sourceGET
	default:
		*t = sourceUnknown
	}

	return nil
}

// loadConfig unmarshalls a JSON config file into a Config struct.
func loadConfig(filename string) (*config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := config{}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
