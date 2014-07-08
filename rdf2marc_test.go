package main

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/digibib/armillaria/sparql"
)

var res = []byte(`
{ "head": { "link": [], "vars": ["245_a", "r", "245_b", "020_a", "019_b", "020_b", "019_d", "090_b", "100_a", "245_j", "100_b", "100_j", "260_a", "260_b", "260_c", "300_a", "300_b", "c008_33", "c008_22", "c008_35", "c001_0"] },
  "results": { "distinct": false, "ordered": true, "bindings": [
    { "245_a": { "type": "literal", "value": "P\u00E5 tide" }	, "r": { "type": "uri", "value": "http://data.deichman.no/resource/tnr_214177" }	, "245_b": { "type": "literal", "value": "dikt" }	, "020_a": { "type": "literal", "value": "8270942790" }	, "019_b": { "type": "uri", "value": "http://data.deichman.no/format/Book" }	, "020_b": { "type": "uri", "value": "http://data.deichman.no/bindingInfo/h" }	, "019_d": { "type": "uri", "value": "http://dbpedia.org/resource/Poetry" }	, "090_b": { "type": "literal", "value": "D Nil" }	, "100_a": { "type": "literal", "value": "Bj\u00F8rn Nilsen" }	, "245_j": { "type": "literal", "value": "Nilsen, Bj\u00F8rn" }	, "100_j": { "type": "literal", "value": "n" }	, "260_a": { "type": "literal", "value": "Oslo" }	, "260_b": { "type": "literal", "value": "Oktober" }	, "260_c": { "type": "literal", "value": "1980" }	, "300_a": { "type": "literal", "value": "90" }	, "300_b": { "type": "literal", "value": "ill" }	, "c008_33": { "type": "uri", "value": "http://dbpedia.org/resource/Fiction" }	, "c008_22": { "type": "uri", "value": "http://data.deichman.no/audience/adult" }	, "c008_35": { "type": "uri", "value": "http://lexvo.org/id/iso639-3/nob" }	, "c001_0": { "type": "literal", "value": "214177" }},
    { "245_a": { "type": "literal", "value": "P\u00E5 tide" }	, "r": { "type": "uri", "value": "http://data.deichman.no/resource/tnr_214177" }	, "245_b": { "type": "literal", "value": "dikt" }	, "020_a": { "type": "literal", "value": "8270942790" }	, "019_b": { "type": "uri", "value": "http://data.deichman.no/format/Book" }	, "020_b": { "type": "uri", "value": "http://data.deichman.no/bindingInfo/h" }	, "019_d": { "type": "uri", "value": "http://dbpedia.org/resource/Poetry" }	, "090_b": { "type": "literal", "value": "D Nil" }	, "100_a": { "type": "literal", "value": "Bj\u00F8rn Nilsen" }	, "245_j": { "type": "literal", "value": "Nilsen, Bj\u00F8rn" }	, "100_j": { "type": "uri", "value": "http://data.deichman.no/nationality/n" }	, "260_a": { "type": "literal", "value": "Oslo" }	, "260_b": { "type": "literal", "value": "Oktober" }	, "260_c": { "type": "literal", "value": "1980" }	, "300_a": { "type": "literal", "value": "90" }	, "300_b": { "type": "literal", "value": "ill" }	, "c008_33": { "type": "uri", "value": "http://dbpedia.org/resource/Fiction" }	, "c008_22": { "type": "uri", "value": "http://data.deichman.no/audience/adult" }	, "c008_35": { "type": "uri", "value": "http://lexvo.org/id/iso639-3/nob" }	, "c001_0": { "type": "literal", "value": "214177" }} ] } }        `)

var wantMARCXML = `<record>
  <leader></leader>
  <controlfield tag="001">214177</controlfield>
  <controlfield tag="008">                      a          1 nob    </controlfield>
  <datafield tag="019" ind1=" " ind2=" ">
    <subfield code="b">l</subfield>
    <subfield code="d">D</subfield>
  </datafield>
  <datafield tag="020" ind1=" " ind2=" ">
    <subfield code="a">8270942790</subfield>
    <subfield code="b">h</subfield>
  </datafield>
  <datafield tag="090" ind1=" " ind2=" ">
    <subfield code="b">D Nil</subfield>
  </datafield>
  <datafield tag="100" ind1=" " ind2="0">
    <subfield code="a">Nilsen, Bjørn</subfield>
    <subfield code="d">1934-</subfield>
    <subfield code="j">n.</subfield>
  </datafield>
  <datafield tag="245" ind1="1" ind2="0">
    <subfield code="a">På tide</subfield>
    <subfield code="b">dikt</subfield>
    <subfield code="c">Bjørn Nilsen</subfield>
  </datafield>
  <datafield tag="260" ind1=" " ind2=" ">
    <subfield code="a">Oslo</subfield>
    <subfield code="b">Oktober</subfield>
    <subfield code="c">1980</subfield>
  </datafield>
  <datafield tag="300" ind1=" " ind2=" ">
    <subfield code="a">90</subfield>
    <subfield code="b">ill</subfield>
  </datafield>
</record>`

func TestConvertRDF2MARC(t *testing.T) {
	var parsedRes sparql.Results
	err := json.Unmarshal(res, &parsedRes)

	if err != nil {
		t.Fatal(err)
	}

	marc, err := convertRDF2MARC(parsedRes)
	if err != nil {
		t.Fatal(err)
	}

	gotMARCXML, err := xml.MarshalIndent(marc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if wantMARCXML != string(gotMARCXML) {
		t.Errorf("want:\n%s\ngot:\n%s", wantMARCXML, gotMARCXML)
	}
}
