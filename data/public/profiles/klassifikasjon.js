var profile = {
  "overview": {
    "title": "Klassifikasjon",
    "desc": "bla bla",
    "type": "<http://www.w3.org/2004/02/skos/core#Concept>"
  },
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
          "predicates": [
            {
              "label": "klassenummer",
              "uri": "<http://www.w3.org/2004/02/skos/core#notation>"
            }
          ],
          "type": "float"
        },
        {
          "id": "edition",
          "label": "Deweyutgave",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "Deweyutgave",
              "uri": "<http://www.w3.org/2004/02/skos/core#inScheme>"
            }
          ],
          "values": [
            {
              "predicate": "<http://www.w3.org/2004/02/skos/core#inScheme>",
              "predicateLabel": "Deweyutgave",
              "value": "<http://data.deichman.no/class/DDK5>",
              "URILabel": "DDK5",
              "source": "local"
            }
          ],
          "type": "select",
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
          "id": "label",
          "label": "Foretrukket betegnelse",
          "desc": "Kort (1-3 ord) beskrivelse av dette Dewey-nummeret",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "begenelse",
              "uri": "<http://www.w3.org/2004/02/skos/core#prefLabel>"
            }
          ],
          "type": "langString"
        },
        {
          "id": "related",
          "label": "Relatert emne",
          "desc": "Emne som beskriver dette Dewey-nummeret",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "relatert emne",
              "uri": "<http://www.w3.org/2004/02/skos/core#narrowMatch>"
            }
          ],
          "type": "URI",
          "searchTypes": ["emne"]
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
      if ( values.label[0] ) {
        label += " " + cleanString(values.label[0].value);
      }
    }
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    if ( values.num[0] ) {
      label = values.num[0].value;
      if ( values.label[0]) {
        label += " " + cleanString( values.label[0].value );
      }
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
