var profile = {
  "overview": {
    "title": "Manifestasjon",
    "desc": "En manifestasjon av et verk",
    "type": ["<http://purl.org/spar/fabio/Manifestation>", "<http://purl.org/ontology/bibo/book>"]
  },
  "externalRequired": ["isbn13"],
  "externalSources": [
    {
      "source": "bibsentralen",
      "genRequest": function( values ) {
        return cleanString( values.isbn13[0].value );
      },
      "parseRequest": function( response ) {
        // parse the marcxml response
        var parser = new DOMParser();
        var xml = parser.parseFromString( response, "text/xml")
        var root = xml.documentElement.nodeName;

        // return on xml parse error
        if ( root  === "parseerror" || root === "error while parsing" ) {
          return [];
        }

        if ( xml.getElementsByTagName("record").length > 1 ) {
          // Don't know how to handle multiple hits
          return [];
        }

        var getSubfield = function( dataField, code ) {
          for (var i=0; i<dataField.children.length; i++) {
            if ( dataField.children[0].getAttribute("code") === code) {
              return dataField.children[0].firstChild.nodeValue;
            }
          }
          return false;
        };

        var values = [];

        for (var i=0; i<xml.getElementsByTagName("datafield").length; i++) {
          var dataField = xml.getElementsByTagName("datafield")[i];

          switch ( dataField.getAttribute("tag") ) {
            case "245": // title
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": v,
                  "predicate": "<http://purl.org/dc/terms/title>",
                  "source": "BS"
                });
              }
              break;
            case "740": // alternative title
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": v,
                  "predicate": "<http://purl.org/dc/terms/alternative>",
                  "source": "BS"
                });
              }
              break;
            case "260": // publication issuer, place & year
              var v = getSubfield( dataField, "c");
              if ( parseInt( v ) ) {
                values.push({
                  "value": parseInt( v ),
                  "predicate": "<http://purl.org/spar/fabio/hasPublicationYear>",
                  "source": "BS"
                });
              }
              break;
            case "300": // number of pages
              var v = getSubfield( dataField, "a");
              if ( parseInt( v ) ) {
                values.push({
                  "value": parseInt( v ),
                  "predicate": "<http://purl.org/ontology/bibo/numPages>",
                  "source": "BS"
                });
              }
              break;
            case "250": // edition
              var v = getSubfield( dataField, "a");
              if ( parseInt( v ) ) {
                values.push({
                  "value": parseInt( v ),
                  "predicate": "<http://purl.org/ontology/bibo/edition>",
                  "source": "BS"
                });
              }
              break;
          }
        }
        return values;
      }
    },
    {
      "source": "OpenLibrary",
      "genRequest": function( values ) {
        return cleanString( values.isbn13[0].value );
      },
      "parseRequest": function( response ) {
        var data = JSON.parse(response);

        // Return if none, or more than one result.
        if ( Object.keys( data ).length != 1 ) {
          return [];
        }

        var key = Object.keys( data )[0];
        var book = data[key];
        var values = [];

        if ( book.number_of_pages ) {
          values.push({
            "value": parseInt( book.number_of_pages ),
            "predicate": "<http://purl.org/ontology/bibo/numPages>",
             "source": "Open Library"
          });
        }

         if ( book.publish_date ) {
          values.push({
            "value": parseInt( book.publish_date ),
            "predicate": "<http://purl.org/spar/fabio/hasPublicationYear>",
             "source": "Open Library"
          });
        }

        if ( book.title ) {
          values.push({
            "value": book.title,
            "predicate": "<http://purl.org/dc/terms/title>",
             "source": "Open Library"
          });
        }

        return values;
      }
    }
  ],
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
