package main

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
