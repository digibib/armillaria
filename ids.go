package main

import (
	"strconv"
	"sync"

	"github.com/knakk/sparql"
)

const queryGetMax = `
SELECT DISTINCT ?profile, MAX(xsd:int(?num)) AS ?max
WHERE {
  _:s  <armillaria://internal/profile> ?profile ;
       <armillaria://internal/id> ?num .
}
`

// idService is responsible for generating integer IDs needed
// for generating URIs for RDF resources. Each RDF type, or profile,
// has it's own number series, which will increment by one for every
// ID request. This way we can ensure that we can create guaranted,
// unique reosource identifiers (URIs).
//
// The map of the next available IDs only exist in the application
// memory, so each time Armellaria starts, it queries the RDF store to
// get the maximum used integer for each type. The ID is stored as
// a RDF triple on each resource.
type idService struct {
	sync.Mutex
	ids map[string]int
}

// NextId returns the next ID for a given RDF type (profile).
func (s idService) NextId(t string) int {
	s.Lock()
	defer s.Unlock()
	s.ids[t] = s.ids[t] + 1
	return s.ids[t]
}

// Init initializes an idService with the maxiumum values for each type,
// by parsing the results from the query queryGetMax defined on top of this file.
// It takes the results as an unparsed application/sparql-results+json response.
func (s idService) Init(r *sparql.Results) error {
	s.Lock()
	defer s.Unlock()

	for _, b := range r.Results.Bindings {
		maxStr := b["max"].Value
		maxNum, err := strconv.Atoi(maxStr)
		if err != nil {
			return err
		}
		s.ids[b["profile"].Value] = maxNum
	}

	return nil
}

// newIdService returns a new idService.
func newIdService() idService {
	return idService{ids: make(map[string]int)}
}
