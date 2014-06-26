var profile = {
  "overview": {
    "title": "Manifestasjon",
    "desc": "En manifestasjon av et litterært verk",
    "type": ["<http://purl.org/spar/fabio/Manifestation>", "<http://purl.org/ontology/bibo/book>"]
  },
  "externalRequired": ["isbn"],
  "externalSources": [
    {
      "source": "BS",
      "genRequest": function( values ) {
        return cleanString( values.isbn[0].value );
      },
      "parseRequest": function( response ) {
        // parse the marcxml response
        var parser = new DOMParser();
        var xml = parser.parseFromString( response, "text/xml")
        var root = xml.documentElement.nodeName;

        // return on xml parse error
        if ( root  === "parseerror" || root === "error while parsing" ) {
          return [[],[]];
        }

        if ( xml.getElementsByTagName("record").length > 1 ) {
          // Don't know how to handle multiple hits
          return [[],[]];
        }

        var getSubfield = function( dataField, code ) {
          for (var i=0; i<dataField.children.length; i++) {
            if ( dataField.children[i].getAttribute("code") === code) {
              return dataField.children[i].firstChild.nodeValue;
            }
          }
          return false;
        };

        var values = [];
        var suggestions = [];

        for (var i=0; i<xml.getElementsByTagName("datafield").length; i++) {
          var dataField = xml.getElementsByTagName("datafield")[i];

          switch ( dataField.getAttribute("tag") ) {
            case "082": // Dewey
              var v = getSubfield( dataField, "a");
              if ( v ) {
                /*var utgv = getSubfield( dataField, "2");
                if ( utgv ) {
                  v = v +  " (DDK" + utgv + ")";
                }*/
                suggestions.push({
                  "value": v,
                  "id": "class"
                });
              }
              break;
            case "100": // Forfatter
              var v = getSubfield( dataField, "a" );
              if ( v ) {
                var lifespan = getSubfield( dataField, "d" );
                var nationality = getSubfield( dataField, "j" );
                if ( lifespan || nationality ) {
                  v = v + " (";
                  if ( nationality ) {
                    v = v + nationality + " ";
                  }
                  if ( lifespan ) {
                    v = v + lifespan;
                  }
                  v = v + ")";
                }
              }
              suggestions.push({
                "value": v,
                "id": "creators",
              });
              break
            case "240": // originaltittel
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "work"
                });
              }
              break;
            case "245": // title
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": '"' + v + '"',
                  "predicate": "<http://purl.org/dc/terms/title>"
                });
              }

              // ansvarsangivelse; statementOfResponsibility
              v = getSubfield( dataField, "c");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "creators"
                });
              }
              break;
            case "740": // alternative title
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": '"' + v + '"',
                  "predicate": "<http://purl.org/dc/terms/alternative>"
                });
              }
              break;
            case "260": // publication issuer, place & year
              var v = getSubfield( dataField, "c");
              if ( parseInt( v ) ) {
                values.push({
                  "value": '' + parseInt( v ),
                  "predicate": "<http://purl.org/spar/fabio/hasPublicationYear>"
                });
              }

              v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "pubPlace"
                });
              }

              v = getSubfield( dataField, "b");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "issuer"
                });
              }
              break;
            case "300": // number of pages
              var v = getSubfield( dataField, "a");
              if ( parseInt( v ) ) {
                values.push({
                  "value": '' + parseInt( v ),
                  "predicate": "<http://purl.org/ontology/bibo/numPages>"
                });
              }
              break;
            case "250": // edition
              var v = getSubfield( dataField, "a");
              if ( parseInt( v ) ) {
                values.push({
                  "value": '' + parseInt( v ),
                  "predicate": "<http://purl.org/ontology/bibo/edition>"
                });
              }
              break;
            case "650": // generelt emneord
              var v = getSubfield( dataField, "a");
              if ( v ) {
                var overordnet = getSubfield( dataField, "0");
                var underordnet = getSubfield( dataField, "x");
                var dewey = getSubfield( dataField, "1");
                if ( overordnet ) {
                  v = overordnet + " > " + v;
                }
                if ( underordnet ) {
                  v = v + " > " + underordnet;
                }
                if ( dewey ) {
                  v = v + " (" + dewey + ")";
                }
              }
              suggestions.push({
                  "value": v,
                  "id": "subject"
                });
              break;
            case "440": // serie
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "partOf"
                });
              }
              break;
            case "700": // Biinførsler
              var v = getSubfield( dataField, "a");
              if ( v ) {
                var role = getSubfield( dataField, "e");
                if ( role ) {
                  v = v +  " (" + role + ")";
                }
                suggestions.push({
                  "value": v,
                  "id": "creators"
                });
              }
              break;
          }
        }
        return [values, suggestions];
      }
    },
    {
      "source": "OCLC",
      "genRequest": function( values ) {
        return cleanString( values.isbn[0].value );
      },
      "parseRequest": function( response ) {
        // http://classify.oclc.org/classify2/api_docs/classify.html#examples
        var parser = new DOMParser();
        var xml = parser.parseFromString( response, "text/xml")
        var root = xml.documentElement.nodeName;

        // return on xml parse error
        if ( root  === "parseerror" || root === "error while parsing" ) {
          return [[],[]];
        }

        if ( xml.getElementsByTagName("response")[0].getAttribute("code") != "2" ) {
          // Currently we only want single-work responses.
          // 0:	Success. Single-work summary response provided.
          // 2:	Success. Single-work detail response provided.
          // 4:	Success. Multi-work response provided.
          return [[],[]];
        }

        var values = [];
        var suggestions = [];

        // Title
        var title = xml.getElementsByTagName("work")[0].getAttribute("title");
        if ( title ) {
          values.push({
            "value": '"' + title + '"',
            "predicate": "<http://purl.org/dc/terms/title>"
          });
        }

        // Authors
        for (var i=0; i<xml.getElementsByTagName("author").length; i++) {
          var author = xml.getElementsByTagName("author")[i];
          var v = author.firstChild.nodeValue;
          if ( author.getAttribute("viaf") ) {
            v = v + ' (viaf: ' + author.getAttribute("viaf") + ')';
          }
          suggestions.push({
            "value": v,
            "id": "creators"
          })
        }

        // Most popular classification
        // TODO what is the difference sfa/nsfa
        var mp = xml.getElementsByTagName("ddc")[0].getElementsByTagName("mostPopular")[0];
        if ( mp ) {
          suggestions.push({
            "value": mp.getAttribute("sfa"),
            "id": "class"
          });
        }

        // Headings (subjects) with fast identificators
        for (var i=0; i<xml.getElementsByTagName("heading").length; i++) {
          var subject = xml.getElementsByTagName("heading")[i];
          var v = subject.firstChild.nodeValue;
          if ( subject.getAttribute("ident") ) {
            v = v + ' (fast: fst' + subject.getAttribute("ident") + ')';
          }
          suggestions.push({
            "value": v,
            "id": "subject"
          });
        }


        return [values, suggestions]
      }
    },
    {
      "source": "LOC",
      "genRequest": function( values ) {
        return cleanString( values.isbn[0].value );
      },
      "parseRequest": function( response ) {
        // parse the marcxml response
        var parser = new DOMParser();
        var xml = parser.parseFromString( response, "text/xml")
        var root = xml.documentElement.nodeName;

        // return on xml parse error
        if ( root  === "parseerror" || root === "error while parsing" ) {
          return [[],[]];
        }

        if ( xml.getElementsByTagName("record").length > 1 ) {
          // Don't know how to handle multiple hits
          return [[],[]];
        }

        var getSubfield = function( dataField, code ) {
          for (var i=0; i<dataField.children.length; i++) {
            if ( dataField.children[i].getAttribute("code") === code) {
              return dataField.children[i].firstChild.nodeValue;
            }
          }
          return false;
        };

        var values = [];
        var suggestions = [];

        for (var i=0; i<xml.getElementsByTagName("datafield").length; i++) {
          var dataField = xml.getElementsByTagName("datafield")[i];

          switch ( dataField.getAttribute("tag") ) {
            case "100": // Forfatter
              var v = getSubfield( dataField, "a" );
              if ( v ) {
                var lifespan = getSubfield( dataField, "d" );
                var nationality = getSubfield( dataField, "j" );
                if ( lifespan || nationality ) {
                  v = v + " (";
                  if ( nationality ) {
                    v = v + nationality + " ";
                  }
                  if ( lifespan ) {
                    v = v + lifespan;
                  }
                  v = v + ")";
                }
              }
              suggestions.push({
                "value": v,
                "id": "creators",
              });
              break;
            case "240": // original title
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "work"
                });
              }
              break;
            case "245": // title
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": '"' + v + '"',
                  "predicate": "<http://purl.org/dc/terms/title>"
                });
              }

              // ansvarsangivelse; statementOfResponsibility
              v = getSubfield( dataField, "c");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "creators"
                });
              }
              break;
            case "520": // summary
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": '"' + v + '"@en',
                  "predicate": "<http://purl.org/dc/terms/abstract>"
                });
              }

              break;
          }
        }
        return [values, suggestions];
      }
    },
    {
      "source": "Bibsys",
      "genRequest": function( values ) {
        return cleanString( values.isbn[0].value );
      },
      "parseRequest": function( response ) {
        // parse the marcxml response
        var parser = new DOMParser();
        var xml = parser.parseFromString( response, "text/xml")
        var root = xml.documentElement.nodeName;
        var ns = "info:lc/xmlns/marcxchange-v1";

        // return on xml parse error
        if ( root  === "parseerror" || root === "error while parsing" ) {
          return [[],[]];
        }

        if ( xml.getElementsByTagNameNS(ns, "record").length != 1 ) {
          // Don't know how to handle multiple (or none) hits
          return [[],[]];
        }

        var getSubfield = function( dataField, code ) {
          for (var i=0; i<dataField.children.length; i++) {
            if ( dataField.children[i].getAttribute("code") === code) {
              return dataField.children[i].firstChild.nodeValue;
            }
          }
          return false;
        };

        var values = [];
        var suggestions = [];

        for (var i=0; i<xml.getElementsByTagNameNS(ns, "datafield").length; i++) {
          var dataField = xml.getElementsByTagNameNS(ns, "datafield")[i];

          switch ( dataField.getAttribute("tag") ) {
            case "082": // Dewey
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "class"
                });
              }
              break;
            case "100": // hovedforfatter
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "creators"
                });
              }
              break
            case "245": // title, subtitle, statementOfResponosibility
              var v = getSubfield( dataField, "a");
              if ( v ) {
                values.push({
                  "value": '"' + v + '"',
                  "predicate": "<http://purl.org/dc/terms/title>"
                });
              }

              v = getSubfield( dataField, "b");
              if ( v ) {
                values.push({
                  "value": '"' + v + '"',
                  "predicate": "<http://purl.org/spar/fabio/hasSubtitle>"
                });
              }

              v = getSubfield( dataField, "c");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "creators"
                });
              }
              break;
            case "246": // originalutgave (paralelltittel)
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "work"
                });
              }
              break;
            case "260": // publication issuer, place & year
              var v = getSubfield( dataField, "c");
              if ( parseInt( v ) ) {
                values.push({
                  "value": '' + parseInt( v ),
                  "predicate": "<http://purl.org/spar/fabio/hasPublicationYear>"
                });
              }

              v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "pubPlace"
                });
              }

              v = getSubfield( dataField, "b");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "issuer"
                });
              }
              break;
            case "300": // number of pages
              var v = getSubfield( dataField, "a");
              if ( parseInt( v ) ) {
                values.push({
                  "value": '' + parseInt( v ),
                  "predicate": "<http://purl.org/ontology/bibo/numPages>"
                });
              }
              break;
            case "250": // edition
              var v = getSubfield( dataField, "a");
              if ( parseInt( v ) ) {
                values.push({
                  "value": '' + parseInt( v ),
                  "predicate": "<http://purl.org/ontology/bibo/edition>"
                });
              }
              break;
            case "830":
            case "490": // serie
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "partOf"
                });
              }
              break;
            case "650": // subject
              var v = getSubfield( dataField, "a");
              if ( v ) {
                suggestions.push({
                  "value": v,
                  "id": "subject"
                });
              }
              break;
            case "700": // Biinførsler
              var v = getSubfield( dataField, "a");
              if ( v ) {
                var role = getSubfield( dataField, "e");
                if ( role ) {
                  v = v +  " (" + role + ")";
                }
                suggestions.push({
                  "value": v,
                  "id": "creators"
                });
              }
              break;
          }
        }
        return [values, suggestions];
      }
    },
    {
      "source": "Open Library",
      "genRequest": function( values ) {
        return cleanString( values.isbn[0].value );
      },
      "parseRequest": function( response ) {
        var data = JSON.parse(response);

        // Return if none, or more than one result.
        if ( Object.keys( data ).length != 1 ) {
          return [[],[]];
        }

        var key = Object.keys( data )[0];
        var book = data[key];
        var values = [];
        var suggestions = [];

        if ( parseInt ( book.number_of_pages ) ) {
          values.push({
            "value": '' + parseInt( book.number_of_pages ),
            "predicate": "<http://purl.org/ontology/bibo/numPages>"
          });
        }

        if ( parseInt( book.publish_date ) ) {
          values.push({
            "value": '' + parseInt( book.publish_date ),
            "predicate": "<http://purl.org/spar/fabio/hasPublicationYear>"
          });
        }

        if ( book.title ) {
          values.push({
            "value": '"' + book.title + '"',
            "predicate": "<http://purl.org/dc/terms/title>"
          });
        }

        if ( book.authors ) {
          book.authors.forEach(function(author) {
            suggestions.push({
              "value": author.name,
              "id": "creators"
            });
          });
        }

        if ( book.publish_places ) {
          book.publish_places.forEach(function( place ) {
            suggestions.push({
              "value": place.name,
              "id": "pubPlace"
            });
          });
        }

        if ( book.publishers ) {
          book.publishers.forEach(function( issuer ) {
            suggestions.push({
              "value": issuer.name,
              "id": "issuer"
            });
          });
        }


        return [values, suggestions];
      }
    },
    {
      "source": "Google Books",
      "genRequest": function( values ) {
        return cleanString( values.isbn[0].value );
      },
      "parseRequest": function( response ) {
        var data = JSON.parse( response );

        // Return earyl if none, or more than 1 hits.
        if ( data.totalItems != 1) {
          return [[],[]];
        }

        var book = data.items[0].volumeInfo;
        var values = [];
        var suggestions = [];

        if ( book.title ) {
          values.push({
            "value": '"' + book.title + '"',
            "predicate": "<http://purl.org/dc/terms/title>"
          });
        }

        if ( book.publishedDate ) {
          values.push({
            "value": '' + parseInt( book.publishedDate ),
            "predicate": "<http://purl.org/spar/fabio/hasPublicationYear>"
          });
        }

        if ( book.pageCount ) {
          values.push({
            "value": '' + parseInt( book.pageCount ),
            "predicate": "<http://purl.org/ontology/bibo/numPages>",
          });
        }

        if ( book.description ) {
          values.push({
            "value": '"' + book.description + '"@en',
            "predicate": "<http://purl.org/dc/terms/description>",
          });
        }

        book.authors.forEach(function( author ) {
          suggestions.push({
            "value": author,
            "id": "creators"
          });
        });

       return [values, suggestions];
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
          "id": "isbn",
          "label": "ISBN",
          "desc": "ISBN-10 eller ISBN-13",
          "required": false,
          "repeatable": true,
          "predicateInferred": true,
          "predicateOptions": ["<http://purl.org/ontology/bibo/isbn10>", "<http://purl.org/ontology/bibo/isbn13>"],
          "predicateFn": function( value ) {
            value = value.replace(/[- ]/g,'');
            // We have to add 2 to length, becuase
            // the value is surrounded by quotes
            if ( value.length == 10 + 2 ) {
              return "<http://purl.org/ontology/bibo/isbn10>";
            }
            if ( value.length == 13 + 2 ) {
              return "<http://purl.org/ontology/bibo/isbn13>";
            }
            return false;
          },
          "type": "isbn"
        },
        {
          "id": "title",
          "label": "Tittel",
          "desc": "",
          "required": false,
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
          "required": false,
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
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/language>",
          "type": "URI",
          "searchTypes": ["language"]
        },
        {
          "id": "notation",
          "label": "Skriftsystem",
          "required": false,
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
        },
        {
          "id": "picture",
          "label": "Omslagsbilde",
          "required": false,
          "repeatable": true,
          "predicate": "<http://xmlns.com/foaf/0.1/depiction>",
          "type": "URL"
        },
        {
          "id": "elversion",
          "label": "Elektronisk utgave",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/spar/fabio/hasURL>",
          "type": "URL"
        },
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
          "required": false,
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
    var label = "";
    if ( values.title[0] ) {
      label = cleanString(values.title[0].value);
    }
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    var label = "";
    if ( values.title[0] ) {
      label = cleanString(values.title[0].value);
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
