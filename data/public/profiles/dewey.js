var profile = {
  "overview": {
    "title": "Dewey-nummer",
    "desc": "Beskriver en plassering i Deweys desimalklassifikasjonssytem"
  },
  "views": [
    {
      "title": "Grunnverdier",
      "desc": "",
      "elements": [
        {
          "id": "loc",
          "label": "Plassering",
          "desc": "Nummer i Dewey klassifikasjonssystemet",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "plassering",
              "uri": "<http://www.w3.org/2004/02/skos/core#notation>"
            }
          ],
          "type": "float"
        },
        {
          "id": "label",
          "label": "Foretrukket betegnelse",
          "desc": "Kort (1-3 ord) beskrivelse av dette Dewey-nummeret",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "begenelse",
              "uri": "<http://www.w3.org/2004/02/skos/core#prefLabel>"
            }
          ],
          "type": "langString"
        }
      ]
    },
    {
      "title": "Relasjoner",
      "desc": "Linker til andre ressurser",
      "elements": [
        {
          "id": "exactMatch",
          "label": "Relatert emne (eksakt treff)",
          "desc": "Emne som beskriver dette Dewey-nummeret",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "relatert emne (nøyaktig)",
              "uri": "<http://www.w3.org/2004/02/skos/core#exactMatch>"
            }
          ],
          "type": "URI",
          "searchTypes": ["concept"],
          "values": []
        },
        {
          "id": "closeMatch",
          "label": "Relatert emne (omtrentlig treff)",
          "desc": "Emne som beskriver dette Dewey-nummeret",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "relatert emne (omtrentlig)",
              "uri": "<http://www.w3.org/2004/02/skos/core#closeMatch>"
            }
          ],
          "type": "URI",
          "searchTypes": ["concept"],
          "values": []
        }]
    }
  ],
  "uriNeedIds": ["loc"],
  "uriFn": function(values) {
     return "<http://dewey.info/class/" + values.loc[0].value + ">";
  },
  "displayLabel": function(values) {
    var label = "";
    if (values.loc[0] && values.label[0]) {
      label = values.loc[0].value + " " + cleanString(values.label[0].value);
    }
    return '"' + label + '"';
  },
  "searchLabel": function(values) {
    var label = "";
    if (values.loc[0] && values.label[0]) {
      label = values.loc[0].value + " " + cleanString(values.label[0].value);
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
