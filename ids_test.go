package main

import (
	"sync"
	"testing"
)

var sparqlRes = []byte(`
{ "head": { "link": [], "vars": ["profile", "max"] },
  "results": { "distinct": false, "ordered": true, "bindings": [
    { "profile": { "type": "literal", "value": "dewey" }	, "max": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#integer", "value": "2" }},
    { "profile": { "type": "literal", "value": "person" }	, "max": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#integer", "value": "743" }} ] } }
`)

func TestNextId(t *testing.T) {
	s := newIdService()
	s.ids["a"] = 1
	s.ids["b"] = 2
	s.ids["c"] = 999

	tests := []struct{ got, want int }{
		{s.NextId("a"), 2},
		{s.NextId("b"), 3},
		{s.NextId("c"), 1000},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("NextId() => %d, want %d", tt.got, tt.want)
		}
	}
}

func TestConcurrentNextId(t *testing.T) {
	s := newIdService()
	num := 1000
	var wg sync.WaitGroup
	res := make(map[int]*struct{})

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res[s.NextId("a")] = nil
		}()
	}

	wg.Wait()
	if len(res) != num || s.NextId("a") != num+1 {
		t.Error("NextId doesn't handle multiple concurrent requests")
	}
}

func TestInitIdServiceFromSparqlResponse(t *testing.T) {
	s := newIdService()
	err := s.Init(sparqlRes)

	if err != nil {
		t.Errorf("idService.Init failed with: %v", err)
	}

	if s.NextId("dewey") != 3 || s.NextId("person") != 744 {
		t.Errorf("idService.Init failed to parse the sparql json response correctly")
	}

}
