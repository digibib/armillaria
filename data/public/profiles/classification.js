var profile = {
  "overview": {
    "title": "Klassifikasjon",
    "desc": "bla bla",
    "type": ["<http://www.w3.org/2004/02/skos/core#Concept>"]
  },
  "externalRequired": ["num"],
  "externalSources": [
    {
      "source": "dewey.info",
      "genRequest": function( values ) {
        var q =
            'PREFIX skos: <http://www.w3.org/2004/02/skos/core#>\
             SELECT DISTINCT ?label WHERE {\
                _:loc skos:notation ?notation ;\
                      skos:prefLabel ?label .\
                FILTER(str(?notation) = "{loc}")\
                FILTER(langMatches(lang(?label), "no"))\
             }'.supplant({loc: values.num[0].value });
        return q;
      },
      "parseRequest": function( response ) {
        res = JSON.parse( response );
        var v = [];
        res.results.bindings.forEach(function (b) {
          v.push({value: '"'+ b.label.value + '"@' + b.label['xml:lang'],
                  predicate: "<http://www.w3.org/2004/02/skos/core#prefLabel>"
          });
        });
        return [v,[]];
      }
    }
  ],
  "views": [
    {
      "title": "",
      "desc": "",
      "elements": [
        {
          "id": "num",
          "label": "Klassenummer",
          "desc": "Nummer i Deweys klassifikasjonssystemet",
          "required": true,
          "repeatable": false,
          "predicate": "<http://www.w3.org/2004/02/skos/core#notation>",
          "type": "float"
        },
        {
          "id": "edition",
          "label": "Deweyutgave",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://www.w3.org/2004/02/skos/core#inScheme>",
          "values": [
            {
              "predicate": "<http://www.w3.org/2004/02/skos/core#inScheme>",
              "predicateLabel": "Deweyutgave",
              "value": "<http://data.deichman.no/class/DDK5>",
              "URILabel": "DDK5",
              "source": "local"
            }
          ],
          "type": "selectMust",
          "options": [
            {
              "predicate": "<http://www.w3.org/2004/02/skos/core#inScheme>",
              "predicateLabel": "Deweyutgave",
              "value": "<http://data.deichman.no/class/DDK4>",
              "URILabel": "DDK4",
              "source": "local"
            },
            {
              "predicate": "<http://www.w3.org/2004/02/skos/core#inScheme>",
              "predicateLabel": "Deweyutgave",
              "value": "<http://data.deichman.no/class/DDK5>",
              "URILabel": "DDK5",
              "source": "local"
            }
          ]
        },
        {
          "id": "prefLabel",
          "label": "Foretrukket betegnelse",
          "desc": "Kort (1-3 ord) beskrivelse av dette Dewey-nummeret",
          "required": false,
          "repeatable": false,
          "predicate": "<http://www.w3.org/2004/02/skos/core#prefLabel>",
          "type": "langString"
        },
        {
          "id": "related",
          "label": "Relatert emne",
          "desc": "Emne som beskriver dette Dewey-nummeret",
          "required": false,
          "repeatable": true,
          "predicate": "<http://www.w3.org/2004/02/skos/core#narrowMatch>",
          "type": "URI",
          "searchTypes": ["subject"]
        }
      ]
    }
  ],
  "uriNeedIds": ["num", "edition"],
  "uriFn": function(values) {
     return '<http://data.deichman.no/class/' + cleanString( values.edition[0].URILabel ) +
            '/' + values.num[0].value + ">";
  },
  "displayLabel": function( values ) {
    var label = "";
    if ( values.num[0] ) {
      label = values.num[0].value;
      if ( values.prefLabel[0] ) {
        label += " " + cleanString(values.prefLabel[0].value);
      }
    }
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    if ( values.num[0] ) {
      label = values.num[0].value;
      if ( values.prefLabel[0]) {
        label += " " + cleanString( values.prefLabel[0].value );
      }
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
