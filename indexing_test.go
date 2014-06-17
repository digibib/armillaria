package main

import (
	"encoding/json"
	"testing"
)

const (
	commonMapping = `{"<armillaria://internal/created>":{"created":{"type":"date","index":"not_analyzed"}},"<armillaria://internal/published>":{"published":{"type":"date","index":"not_analyzed"}},"<armillaria://internal/updated>":{"updated":{"type":"date","index":"not_analyzed"}},"<armillaria://internal/searchLabel>":{"searchLabel":{"type":"string"}},"<armillaria://internal/displayLabel>":{"displayLabel":{"type":"string","index":"not_analyzed"}},"dummy":{"uri":{"type":"string","index":"not_analyzed"}}}`
	workMapping   = `{"<http://purl.org/dc/terms/title>":{"title":{"type":"string"}},"<http://purl.org/spar/fabio/hasSubtitle>":{"subtitle":{"type":"string"}},"<http://purl.org/dc/terms/alternative>":{"altTitle":{"type":"string"}},"<http://purl.org/spar/fabio/hasPublicationYear>":{"pubYear":{"type":"integer"}},"<http://purl.org/dc/terms/language>":{"language":{"type":"nested","properties":{"uri":{"type":"string","index":"not_analyzed"},"label":{"type":"string"}}}},"<http://purl.org/dc/terms/creator>":{"creator":{"type":"nested","properties":{"uri":{"type":"string","index":"not_analyzed"},"label":{"type":"string"}}}},"<http://data.deichman.no/illustrator>":{"illustrator":{"type":"nested","properties":{"uri":{"type":"string","index":"not_analyzed"},"label":{"type":"string"}}}},"<http://purl.org/ontology/bibo/editor>":{"editor":{"type":"nested","properties":{"uri":{"type":"string","index":"not_analyzed"},"label":{"type":"string"}}}},"<http://purl.org/dc/terms/contributor>":{"contributor":{"type":"nested","properties":{"uri":{"type":"string","index":"not_analyzed"},"label":{"type":"string"}}}}}`
	jsonRes       = `{"head":{"link":[],"vars":["p","o","l"]},"results":{"distinct":false,"ordered":true,"bindings":[{"p":{"type":"uri","value":"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},"o":{"type":"typed-literal","datatype":"http://www.w3.org/2001/XMLSchema#integer","value":"0"}},{"p":{"type":"uri","value":"http://purl.org/dc/terms/alternative"},"o":{"type":"literal","value":"test1"}},{"p":{"type":"uri","value":"http://purl.org/dc/terms/alternative"},"o":{"type":"literal","value":"test2"}},{"p":{"type":"uri","value":"http://purl.org/dc/terms/title"},"o":{"type":"literal","xml:lang":"en","value":"Cat's Cradle"}},{"p":{"type":"uri","value":"armillaria://internal/displayLabel"},"o":{"type":"literal","value":"Cat's Cradle"}},{"p":{"type":"uri","value":"armillaria://internal/searchLabel"},"o":{"type":"literal","value":"Cat's Cradle"}},{"p":{"type":"uri","value":"armillaria://internal/updated"},"o":{"type":"typed-literal","datatype":"http://www.w3.org/2001/XMLSchema#dateTime","value":"2014-06-17T01:33:20.813Z"}},{"p":{"type":"uri","value":"armillaria://internal/profile"},"o":{"type":"literal","value":"work"}},{"p":{"type":"uri","value":"armillaria://internal/created"},"o":{"type":"typed-literal","datatype":"http://www.w3.org/2001/XMLSchema#dateTime","value":"2014-06-17T01:21:58.804Z"}},{"p":{"type":"uri","value":"armillaria://internal/published"},"o":{"type":"typed-literal","datatype":"http://www.w3.org/2001/XMLSchema#dateTime","value":"2014-06-17T01:21:58.804Z"}},{"p":{"type":"uri","value":"http://purl.org/spar/fabio/hasPublicationYear"},"o":{"type":"typed-literal","datatype":"http://www.w3.org/2001/XMLSchema#integer","value":"1963"}},{"p":{"type":"uri","value":"armillaria://internal/id"},"o":{"type":"typed-literal","datatype":"http://www.w3.org/2001/XMLSchema#integer","value":"8"}},{"p":{"type":"uri","value":"http://purl.org/dc/terms/creator"},"o":{"type":"uri","value":"http://data.deichman.no/person/17"},"l":{"type":"literal","value":"Kurt Vonnegut (1922-2007)"}},{"p":{"type":"uri","value":"http://purl.org/dc/terms/language"},"o":{"type":"uri","value":"http://lexvo.org/id/iso639-3/eng"},"l":{"type":"literal","xml:lang":"nb","value":"engelsk"}}]}}`
	esDocWant     = `{"altTitle":["test1","test2"],"created":"2014-06-17T01:21:58.804Z","creator":{"uri":"http://data.deichman.no/person/17","label":"Kurt Vonnegut (1922-2007)"},"displayLabel":"Cat's Cradle","language":{"uri":"http://lexvo.org/id/iso639-3/eng","label":"engelsk"},"pubYear":"1963","published":"2014-06-17T01:21:58.804Z","searchLabel":"Cat's Cradle","title":"Cat's Cradle","updated":"2014-06-17T01:33:20.813Z","uri":"http://data.deichman.no/work/8"}`
)

func TestCreateIndexingDoc(t *testing.T) {
	mappings := make(map[string]map[string]string)
	var common preMappings

	err := json.Unmarshal([]byte(commonMapping), &common)
	if err != nil {
		t.Fatal(err)
	}

	mappings["work"] = make(map[string]string)
	var m preMappings

	err = json.Unmarshal([]byte(workMapping), &m)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range common {
		for ki := range v {
			mappings["work"][k] = ki
		}
	}

	for k, v := range m {
		for ki := range v {
			mappings["work"][k] = ki
		}
	}

	esDocGot, profile, err := createIndexDoc(
		mappings,
		[]byte(jsonRes),
		"http://data.deichman.no/work/8",
	)

	if err != nil {
		t.Fatal(err)
	}

	if profile != "work" {
		t.Errorf("want \"work\", got %s", profile)
	}

	if string(esDocGot) != esDocWant {
		t.Errorf("want %s, got %s", esDocWant, string(esDocGot))
	}

}
