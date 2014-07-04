var profile = {
  "overview": {
    "title": "Organisasjon",
    "desc": "",
    "type": ["<http://xmlns.com/foaf/0.1/Organization>"]
  },
  "views": [
    {
      "title": "Beskrivelse",
      "desc": "",
      "elements": [
        {
          "id": "name",
          "label": "Navn",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicate": "<http://xmlns.com/foaf/0.1/name>",
          "type": "string"
        }
      ]
    }
  ],
      "displayLabel": function( values ) {
        var label = '';
        if ( values.name[0] ) {
          label += cleanString( values.name[0].value );
        }
        return '"' + label + '"';
      },
      "searchLabel": function( values ) {
        var label = '';
        if ( values.name[0] ) {
          label = cleanString( values.name[0].value );
        }

        return '"' + label + '"';
      },
};
