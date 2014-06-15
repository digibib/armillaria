var profile = {
  "overview": {
    "title": "Verk",
    "desc": "Beskriver et verk, forklar det om du kan",
    "type": ["<http://purl.org/spar/fabio/Work>"]
  },
  "views": [
    {
      "title": "Basisopplysninger",
      "desc": "",
      "elements": [
        {
          "id": "title",
          "label": "Tittel",
          "desc": "på orginalspråket?",
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
          "id": "pubyear",
          "label": "Førsteutgave",
          "desc": "Årstall for første utgivelse, dersom kjent.",
          "required": false,
          "repeatable": false,
          "predicate":"<http://purl.org/spar/fabio/hasPublicationYear>",
          "type": "integer"
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
      "title": "Manifestasjoner",
      "desc": "",
      "elements": [
        {
          "id": "edition",
          "label": "Manifestasjon",
          "desc": "Manifestasjon/utgave av verket",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/spar/fabio/hasManifestation>",
          "type": "URI",
          "searchTypes": ["manifestasjon"]
        },
        {
          "id": "partOf",
          "label": "Del av",
          "desc": "Dersom dette verket er en del av et annet verk",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/isPartOf>",
          "type": "URI",
          "searchTypes": ["verk"]
        },
        {
          "id": "hasPart",
          "label": "Har del",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/hasPart>",
          "type": "URI",
          "searchTypes": ["verk"]
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
        },
        {
          "id": "relatedWork",
          "label": "Relatert verk",
          "desc": "",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/vocab/frbr/core#relatedEndeavour>",
          "type": "URI",
          "searchTypes": ["verk"]
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
            {"uri": "<http://data.deichman.no/illustrator>", "predicateLabel": "illustratør"},
            {"uri": "<http://purl.org/ontology/bibo/editor>", "predicateLabel": "redaktør"},
            {"uri": "<http://purl.org/dc/terms/contributor>", "predicateLabel": "bidragsyter"}
          ],
          "type": "multiPredicateURI",
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
