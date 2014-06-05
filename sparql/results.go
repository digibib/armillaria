package sparql

// Results holds the parsed results of a application/sparql-results+json response
type Results struct {
	Head    sparqlHeader
	Results sparqlResults
}

type sparqlHeader struct {
	Link []string
	Vars []string
}

type sparqlResults struct {
	Distinct bool
	Ordered  bool
	Bindings []map[string]sparqlBinding
}

type sparqlBinding struct {
	Type     string // "uri", "literal", "typed-literal" or "bnode"
	Value    string
	Lang     string `json:"xml:lang"`
	DataType string
}
