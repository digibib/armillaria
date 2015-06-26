var profile = {
  "overview": {
    "title": "Språk",
    "desc": "Beskriver et språk",
    "type": ["<http://lexvo.org/ontology#Language>"]
  },
  "views": [
    {
      "title": "",
      "desc": "",
      "elements": [
        {
          "id": "title",
          "label": "Navn",
          "desc": "Hva heter språket",
          "required": true,
          "repeatable": true,
          "predicate": "<http://purl.org/dc/terms/title>",
          "type": "langString"
        },
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