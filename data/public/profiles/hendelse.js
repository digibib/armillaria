var profile = {
  "overview": {
    "title": "Hendelse",
    "desc": "Beskriver en historisk hendelse",
    "type": "<http://purl.org/NET/c4dm/event.owl#Event>"
  },
  "views": [
    {
      "title": "Beskrivelse",
      "desc": "",
      "elements": [
        {
          "id": "title",
          "label": "Navn",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://purl.org/dc/terms/title>",
          "type": "langString"
        },
        {
          "id": "desc",
          "label": "Beskrivelse",
          "desc": "Hva skjedde egentlig",
          "required": false,
          "repeatable": false,
          "predicate": "<http://purl.org/dc/terms/description>",
          "type": "langString"
        },
        {
          "id": "place",
          "label": "Sted",
          "desc": "Hvor skjedde det?",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/NET/c4dm/event.owl#place>",
          "type": "URI",
          "searchTypes": ["sted"]
        },
        {
          "id": "time",
          "label": "Tid",
          "desc": "Når skjedde det?",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/NET/c4dm/event.owl#time>",
          "type": "URI",
          "searchTypes": ["temporalEntitet"]
        },
        {
          "id": "agent",
          "label": "Aktør",
          "desc": "Hvem gjorde det?",
          "required": false,
          "repeatable": true,
          "predicate": "<http://purl.org/NET/c4dm/event.owl#agent>",
          "type": "URI",
          "searchTypes": ["agent"]
        }
      ]
    }
  ],
  "displayLabel": function( values ) {
    var label = "";
    if ( values.title[0] ) {
      label = cleanString( values.title[0].value );
    }
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    var label = "";
    if ( values.title[0] ) {
      label = cleanString( values.title[0].value );
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
