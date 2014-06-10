var profile = {
  "overview": {
    "title": "Sted",
    "desc": "Beskriver et geografisk sted",
    "type": "<http://www.geonames.org/ontology#Feature>"
  },
  "views": [
    {
      "title": "Beskrivelse",
      "desc": "",
      "elements": [
        {
          "id": "name",
          "label": "Navn",
          "desc": "Hva heter stedet?",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "navn",
              "uri": "<http://www.geonames.org/ontology#name>"
            }
          ],
          "type": "string"
        },
        {
          "id": "geoId",
          "label": "Geonames-ID",
          "desc": "Denne finner du slik: søk i http://www.geonames.org/",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "Geonames-ID",
              "uri": "<http://www.geonames.org/ontology#geonamesID>"
            }
          ],
          "type": "integer"
        },
        {
          "id": "country",
          "label": "Landkode",
          "desc": "TODO denne bør kanskje velges fra en nedtrekksliste?",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "landkode",
              "uri": "<http://www.geonames.org/ontology#countryCode>"
            }
          ],
          "type": "string"
        },
        {
          "id": "lat",
          "label": "Breddegrad",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "breddegrad",
              "uri": "<http://www.w3.org/2003/01/geo/wgs84pos#lat>"
            }
          ],
          "type": "float"
        },
        {
          "id": "long",
          "label": "Lengdegrad",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "lengdegrad",
              "uri": "<http://www.w3.org/2003/01/geo/wgs84pos#long>"
            }
          ],
          "type": "float"
        }
      ]
    }
  ],
  "displayLabel": function( values ) {
    var label = '';
    if ( values.name[0] ) {
      label += cleanString( values.name[0].value );
    }
    if ( values.country[0] ) {
      label += ' (' + cleanString( values.country[0].value ) + ')';
    }
    return '"' + label + '"';
  },
  "searchLabel": function( values ) {
    var label = '';
    if ( values.name[0] ) {
      label = cleanString( values.name[0].value );
    }
    if ( values.country[0] ) {
      label += ' ' + cleanString( values.country[0].value );
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
