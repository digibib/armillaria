var profile = {
  "overview": {
    "title": "Emne",
    "desc": "Bla bla bla emne, hva er det",
    "type": "<http://www.w3.org/2004/02/skos/core#Concept>"
  },
  "views": [
    {
      "title": "Beskrivelse",
      "desc": "",
      "elements": [
        {
          "id": "prefLabel",
          "label": "Emneord",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "emneord",
              "uri": "<http://www.w3.org/2004/02/skos/core#prefLabel>"
            }
          ],
          "type": "langString"
        },
        {
          "id": "altLabel",
          "label": "Alt. emneord",
          "desc": "Alternativ betegnelse på emnet.",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "alt. emneord",
              "uri": "<http://www.w3.org/2004/02/skos/core#altLabel>"
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
          "id": "class",
          "label": "Klassifikasjon",
          "desc": "Overordnet emne",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "klassifikasjon",
              "uri": "<http://www.w3.org/2004/02/skos/core#broadMatch>"
            }
          ],
          "type": "URI",
          "searchTypes": ["emne"]
        },
        {
          "id": "subdiv",
          "label": "Underaveling",
          "desc": "Underordnet emne?",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "underavdeling",
              "uri": "<http:/data.deichman.no/subdivision>"
            }
          ],
          "type": "URI",
          "searchTypes": ["emne"]
        },
        {
          "id": "qualic",
          "label": "Kvalifikator",
          "desc": "Du vet.. for ekstra kvalitet",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "kvalifikator",
              "uri": "<http://data.deichman.no/qualifier>"
            }
          ],
          "type": "URI",
          "searchTypes": ["emne"]
        },
        {
          "id": "seealso",
          "label": "Se også",
          "desc": "Relatert emne",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "se også",
              "uri": "<http://www.w3.org/2004/02/skos/core#related>"
            }
          ],
          "type": "URI",
          "searchTypes": ["emne"]
        },
        {
          "id": "focus",
          "label": "Fokus",
          "desc": "fokus, fokus, fokus",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "fokus",
              "uri": "<http://xmlns.com/foaf/0.1/focus>"
            }
          ],
          "type": "URI",
          "searchTypes": ["agent", "verk", "sted", "hendelse", "wgs84"]
        }
      ]
    }
  ],
  "displayLabel": function( values ) {
    var label = "";
    if ( values.prefLabel[0] ) {
      label = cleanString( values.prefLabel[0].value );
    }
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    var label = "";
    if ( values.prefLabel[0] ) {
      label = cleanString( values.prefLabel[0].value );
    }
    if ( values.altLabel.length > 0 ) {
      label += ' ';
      label += _.map( values.altLabel, function( e ) {
         return cleanString( e.value );
      } ).join( ' ' );
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
