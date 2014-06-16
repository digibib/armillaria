var profile = {
  "overview": {
    "title": "Manifestasjon",
    "desc": "En manifestasjon av et verk",
    "type": ["<http://purl.org/spar/fabio/Manifestation>", "<http://purl.org/ontology/bibo/book>"]
  },
  "views": [
    {
      "title": "Basisopplysninger",
      "desc": "",
      "elements": [
        {
          "id": "tnr",
          "label": "Tittelnummer",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/number>",
          "type": "integer"
        },
        {
          "id": "isbn10",
          "label": "ISBN10",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/ontology/bibo/isbn10>",
          "type": "string"
        },
        {
          "id": "isbn13",
          "label": "ISBN13",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/ontology/bibo/isbn13>",
          "type": "string"
        },
        {
          "id": "title",
          "label": "Tittel",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://purl.org/dc/terms/title>",
          "type": "langString"
        },
        {
          "id": "subtitle",
          "label": "Undertittel",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate":"<http://purl.org/spar/fabio/hasSubtitle>",
          "type": "langString"
        },
        {
          "id": "vartitle",
          "label": "Variant av tittel",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate":"<http://purl.org/dc/terms/alternative>",
          "type": "langString"
        },
        {
          "id": "edition",
          "label": "Utgave",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/edition>",
          "type": "integer"
        },
        {
          "id": "pubYear",
          "label": "Utgivelsesår",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate":"<http://purl.org/spar/fabio/hasPublicationYear>",
          "type": "integer"
        },
        {
          "id": "issuer",
          "label": "Utgiver",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/issuer>",
          "type": "URI",
          "searchTypes": ["agent", "person", "organization"]
        },
        {
          "id": "pubPlace",
          "label": "Utgivelsessted",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://data.deichman.no/publicationPlace>",
          "type": "URI",
          "searchTypes": ["location"]
        },
        {
          "id": "lang",
          "label": "Språk",
          "desc": "",
          "required": true,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/language>",
          "type": "URI",
          "searchTypes": ["language"]
        },
        {
          "id": "notation",
          "label": "Skriftsystem",
          "desc": "Er dette det samme som alfabet?",
          "required": true,
          "repeatable": true,
          "predicate": "<http://www.ifomis.org/bfo/1.0#notation>",
          "type": "URI",
          "searchTypes": ["script"],
          "values": [
            {
              "predicate": "<http://www.ifomis.org/bfo/1.0#notation>",
              "predicateLabel": "Skriftsystem",
              "value": "<http://lexvo.org/id/script/Latn>",
              "URILabel": "latinsk",
              "source": "local"
            }
          ],
          "options": [
            {
              "predicate": "<http://www.ifomis.org/bfo/1.0#notation>",
              "predicateLabel": "Skriftsystem",
              "value": "<http://data.deichman.no/class/notation/1>",
              "URILabel": "Latinsk",
              "source": "local"
            },
            {
              "predicate": "<http://www.ifomis.org/bfo/1.0#notation>",
              "predicateLabel": "Skriftsystem",
              "value": "<http://data.deichman.no/class/notation/2>",
              "URILabel": "Gresk",
              "source": "local"
            }
          ]
        },
        {
          "id": "numPages",
          "label": "Sidetall",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/numPages>",
          "type": "integer"
        },
        {
          "id": "format",
          "label": "Format",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://purl.org/dc/terms/MediaTypeOrExtent>",
          "type": "URI",
          "searchTypes": ["mediaType"]
        },
        {
          "id": "audience",
          "label": "Målgruppe",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/audience>",
          "type": "select",
          "options": [
            {
              "URILabel": "Voksne",
              "value": "<http://data.deichman.no/class/audience/1>",
              "predicate": "<http://purl.org/dc/terms/audience>",
              "source": "local"
            },
            {
              "URILabel": "Barn",
              "value": "<http://data.deichman.no/class/audience/2>",
              "predicate": "<http://purl.org/dc/terms/audience>",
              "source": "local"
            },
            {
              "URILabel": "Ungdom",
              "value": "<http://data.deichman.no/class/audience/3>",
              "predicate": "<http://purl.org/dc/terms/audience>",
              "source": "local"
            },
            {
              "URILabel": "Pensjonister",
              "value": "<http://data.deichman.no/class/audience/4>",
              "predicate": "<http://purl.org/dc/terms/audience>",
              "source": "local"
            }
          ]
        },
        {
          "id": "desc",
          "label": "Beskrivelse",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/description>",
          "type": "text"
        },
        {
          "id": "abs",
          "label": "Sammendrag",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/abstract>",
          "type": "text"
        }
      ]
    },
    {
      "title": "Verkstilknyttninger",
      "desc": "",
      "elements": [
        {
          "id": "work",
          "label": "Verk",
          "desc": "Dette er en manifestasjon av følgende verk",
          "required": true,
          "repeatable": true,
          "predicate": "<http://purl.org/spar/fabio/isManifestationOf>",
          "type": "URI",
          "searchTypes": ["work"]
        },
        {
          "id": "partOf",
          "label": "Del av",
          "desc": "Dersom manifestasjonen er en del av en samling, flerbindsverk eller serie",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/isPartOf>",
          "type": "URI",
          "searchTypes": ["multiVolumeWork", "serie", "collection"]
        },
        {
          "id": "hasPart",
          "label": "Har del",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/hasPart>",
          "type": "URI",
          "searchTypes": ["manifestation", "partOfBook"]
        },
        {
          "id": "bind",
          "label": "Bindnummer",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/volume>",
          "type": "integer"
        },
        {
          "id": "nr",
          "label": "Nummer i serie",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/locator>",
          "type": "integer"
        }
      ]
    },
    {
      "title": "Opphavspersoner",
      "desc": "",
      "elements": [
        {
          "id": "creators",
          "label": "Opphavspersoner",
          "desc": "et godt norsk ord for creator",
          "required": false,
          "repeatable": true,
          "predicates": [
            {"uri": "<http://purl.org/dc/terms/creator>", "predicateLabel": "opphavsperson"},
            {"uri": "<http://purl.org/ontology/bibo/translator>", "predicateLabel": "oversetter"},
            {"uri": "<http://data.deichman.no/illustrator>", "predicateLabel": "illustratør"},
            {"uri": "<http://purl.org/ontology/bibo/editor>", "predicateLabel": "redaktør"},
            {"uri": "<http://purl.org/dc/terms/contributor>", "predicateLabel": "bidragsyter"}
          ],
          "type": "multiPredicateURI",
          "searchTypes": ["person", "agent", "organization"]
        }
      ]
    },
    {
      "title": "Klassifisering",
      "desc": "",
      "elements": [
        {
          "id": "litform",
          "label": "Litterær form",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http:/data.deichman.no/literaryFormat>",
          "type": "URI",
          "searchTypes": ["literaryFormat"]
        },
        {
          "id": "genre",
          "label": "Sjanger",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://dbpedia.org/ontology/literaryGenre>",
          "type": "URI",
          "searchTypes": ["genre"]
        },
        {
          "id": "subject",
          "label": "Emne",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/subject>",
          "type": "URI",
          "searchTypes": ["subject"]
        },
        {
          "id": "class",
          "label": "Klassifikasjon",
          "desc": "Deweynummer",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/subjectDewey>",
          "type": "URI",
          "searchTypes": ["classification"]
        },
        {
          "id": "tag",
          "label": "Stikkord",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://commontag.org/ns#Tag>",
          "type": "URI",
          "searchTypes": ["tag"]
        }
      ]
    }
  ],
  "displayLabel": function( values ) {
    var label = cleanString(values.title[0].value);
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    var label = cleanString(values.title[0].value);
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
