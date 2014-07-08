package main

import (
	"encoding/xml"

	"github.com/digibib/armillaria/sparql"
)

// Data structures for MARCXML marshalling
// =======================================

// marcRecord represents the top <record> node
type marcRecord struct {
	XMLName    xml.Name `xml:"record"`
	Leader     string   `xml:"leader"`
	CtrlFields []cField `xml:"controlfield"`
	DataFields []dField `xml:"datafield"`
}

// cField represents a <controlfield>
type cField struct {
	Tag   string `xml:"tag,attr"`
	Field string `xml:",chardata"`
}

// dField represents a <datafield>
type dField struct {
	Tag       string     `xml:"tag,attr"`
	Ind1      string     `xml:"ind1,attr"`
	Ind2      string     `xml:"ind2,attr"`
	SubFields []subField `xml:"subfield"`
}

// subField represents a <subfield> under a <datafield>
type subField struct {
	Code  string `xml:"code,attr"`
	Value string `xml:",chardata"`
}

// RDF2MARC Mappings
// =================

// literalMappings contains the mappings of URIs into string values
// used in DeichmanMARC/NORMARC.
var literalMappings = map[string]string{
	"http://data.deichman.no/format/Book":    "l",
	"http://dbpedia.org/resource/Poetry":     "D",
	"http://data.deichman.no/bindingInfo/h":  "h",
	"http://data.deichman.no/audience/adult": "a",
	"http://lexvo.org/id/iso639-3/nob":       "nob",
	"http://data.deichman.no/nationality/n":  "n",
}

type dMapping struct {
	dataField  string
	index1     string
	index2     string
	repeatable bool
	subFields  []sMapping
}

type sMapping struct {
	code       string
	repeatable bool
}

var marcMappings = []dMapping{
	{
		dataField: "019",
		subFields: []sMapping{
			{code: "b"},
			{code: "d", repeatable: true},
		},
	},
	{
		dataField: "020",
		subFields: []sMapping{
			{code: "a"},
			{code: "b"},
		},
	},
	{
		dataField: "090",
		subFields: []sMapping{
			{code: "b"},
		},
	},
	{
		dataField: "100", index2: "0",
		subFields: []sMapping{
			{code: "a"},
			{code: "b"},
			{code: "j"},
		},
	},
	{
		dataField: "245", index1: "1", index2: "0",
		subFields: []sMapping{
			{code: "a"},
			{code: "b"},
			{code: "j"},
		},
	},
	{
		dataField: "260",
		subFields: []sMapping{
			{code: "a"},
			{code: "b"},
			{code: "c"},
		},
	},
	{
		dataField: "300",
		subFields: []sMapping{
			{code: "a"},
			{code: "b"},
		},
	},
}

// Helper functions
// ================

// bindings takes a sparql.Reusults and returns a map where each
// bound variable has a key.
// TODO move to sparql package?
func bindings(rdf sparql.Results) map[string][]string {
	rb := make(map[string][]string)
	for _, k := range rdf.Head.Vars {
		for _, b := range rdf.Results.Bindings {
			if b[k].Value != "" {
				rb[k] = append(rb[k], b[k].Value)
			}
		}
	}
	return rb
}

// API
// ===

// convertRDF2MARC takes a SPARQL result response, and converts it into
// a marcRecord, which is easily serializable as marcxml.
func convertRDF2MARC(rdf sparql.Results) (marcRecord, error) {
	rec := marcRecord{}
	rec.CtrlFields = []cField{
		{Tag: "001"},
		{Tag: "008"},
	}

	bindings := bindings(rdf)

	for _, m := range marcMappings {
		field := dField{Tag: m.dataField, Ind1: m.index1, Ind2: m.index2}
		var foundMatch bool
		for _, s := range m.subFields {
			boundVar := m.dataField + "_" + s.code
			if v, ok := bindings[boundVar]; ok {
				val := v[0] // we only deal with non-repeatable fields for now
				if v2, ok := literalMappings[val]; ok {
					val = v2
				}
				field.SubFields = append(field.SubFields,
					subField{Code: s.code, Value: val})
				foundMatch = true
			}
		}
		if foundMatch {
			rec.DataFields = append(rec.DataFields, field)
		}
	}

	return rec, nil
}
