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
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/isbn10>",
          "type": "string"
        },
        {
          "id": "isbn13",
          "label": "ISBN13",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://purl.org/ontology/bibo/isbn13>",
          "type": "integer"
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
          "searchTypes": ["agent", "person", "organisasjon"]
        },
        {
          "id": "pubPlace",
          "label": "Utgivelsessted",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://data.deichman.no/publicationPlace>",
          "type": "URI",
          "searchTypes": ["sted"]
        },
        {
          "id": "lang",
          "label": "Språk",
          "desc": "TODO nedtrekksliste?",
          "required": true,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/language>",
          "type": "string"
        },
        {
          "id": "notation",
          "label": "Skriftsystem",
          "desc": "Er dette det samme som alfabet?",
          "required": true,
          "repeatable": false,
          "predicate": "<http://www.ifomis.org/bfo/1.0#notation>",
          "type": "selectMust",
          "values": [
            {
              "predicate": "<http://www.ifomis.org/bfo/1.0#notation>",
              "predicateLabel": "Skriftsystem",
              "value": "<http://data.deichman.no/class/notation/1>",
              "URILabel": "Latinsk",
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
          "searchTypes": ["verk"]
        },
        {
          "id": "partOf",
          "label": "Del av",
          "desc": "Dersom manifestasjonen er en del av en samling, flerbindsverk eller serie",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/isPartOf>",
          "type": "URI",
          "searchTypes": ["flerbindsverk", "serie", "samling"]
        },
        {
          "id": "hasPart",
          "label": "Har del",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/hasPart>",
          "type": "URI",
          "searchTypes": ["manifestasjon", "delAvBok"]
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
      "title": "Skapere",
      "desc": "\"Skapere\" er kanskje ikke det riktige ordet...",
      "elements": [
        {
          "id": "creator",
          "label": "Opphavsperson",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/creator>",
          "type": "URI",
          "searchTypes": ["person", "agent", "organisasjon"]
        },
        {
          "id": "translator",
          "label": "Oversetter",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/ontology/bibo/translator>",
          "type": "URI",
          "searchTypes": ["person", "agent", "organisasjon"]
        },
        {
          "id": "illustrator",
          "label": "Illustratør",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://data.deichman.no/illustrator>",
          "type": "URI",
          "searchTypes": ["person", "agent", "organisasjon"]
        },
        {
          "id": "editor",
          "label": "Redaktør",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/ontology/bibo/editor>",
          "type": "URI",
          "searchTypes": ["person", "agent", "organisasjon"]
        },
        {
          "id": "contributor",
          "label": "Bidragsyter",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/contributor>",
          "type": "URI",
          "searchTypes": ["person", "agent", "organisasjon"]
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
          "searchTypes": ["literærForm"]
        },
        {
          "id": "genre",
          "label": "Sjanger",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://dbpedia.org/ontology/literaryGenre>",
          "type": "URI",
          "searchTypes": ["sjanger"]
        },
        {
          "id": "subject",
          "label": "Emne",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/subject>",
          "type": "URI",
          "searchTypes": ["emne"]
        },
        {
          "id": "class",
          "label": "Klassifikasjon",
          "desc": "Deweynummer",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/subjectDewey>",
          "type": "URI",
          "searchTypes": ["klassifikasjon"]
        },
        {
          "id": "tag",
          "label": "Stikkord",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://commontag.org/ns#Tag>",
          "type": "URI",
          "searchTypes": ["stikkord"]
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
