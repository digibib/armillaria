var profile = {
  "overview": {
    "title": "Person",
    "desc": "Beskriver en historisk eller nålevende person, aka homo sapiens",
    "type": "<http://xmlns.com/foaf/0.1/Person>"
  },
  "views": [
    {
      "title": "Personopplysninger",
      "desc": "Her fyller du inn personalia. Sett i gang.",
      "elements": [
        {
          "id": "firstname",
          "label": "Fornavn",
          "desc": "",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "fornavn",
              "uri": "<http://xmlns.com/foaf/0.1/givenName>"
            }
          ],
          "type": "string"
        },
        {
          "id": "lastname",
          "label": "Etternavn",
          "desc": "Etternavn kan være så mangt.",
          "required": true,
          "repeatable": false,
          "predicates": [
            {
              "label": "etternavn",
              "uri": "<http://xmlns.com/foaf/0.1/familyName>"
            }
          ],
          "type": "string"
        },
        {
          "id": "tittel",
          "label": "Tittel",
          "desc": "herr fru professor doktor ærbødigst?",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "tittel",
              "uri": "<http://xmlns.com/foaf/0.1/title>"
            }
          ],
          "type": "string"
        },
        {
          "id": "nummer",
          "label": "Nummer",
          "desc": "Hva er dette, tro?",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "nummer",
              "uri": "<http://data.deichman.no/ordinal>"
            }
          ],
          "type": "integer"
        },
        {
          "id": "birthyear",
          "label": "Fødselsår",
          "desc": "Anno domini",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "fødselssår",
              "uri": "<http://dbpedia.org/ontology/birthYear>"
            }
          ],
          "type": "integer"
        },
        {
          "id": "deathyear",
          "label": "Dødsår",
          "desc": "Anno domini",
          "required": false,
          "repeatable": false,
          "predicates": [
            {
              "label": "dødsår",
              "uri": "<http://dbpedia.org/ontology/deathYear>"
            }
          ],
          "type": "integer"
        },
        {
          "id": "nationality",
          "label": "Nasjonalitet",
          "desc": "En person kan ha flere nasjonaliteter. Fyll inn alle du vet om.",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "nasjonalitet",
              "uri": "<http://dbpedia.org/ontology/nationality>"
            }
          ],
          "type": "URI",
          "searchTypes": ["sted"]
        },
        {
          "id": "pseudo",
          "label": "Pseudonym",
          "desc": "Også kjent som...",
          "required": false,
          "repeatable": true,
          "predicates": [
            {
              "label": "pseudonym",
              "uri": "<http://data.deichman.no/hasPseudonym>"
            }
          ],
          "type": "URI",
          "searchTypes": ["agent"]
        }
      ]
    }
  ],
  "displayLabel": function( values ) {
    var label = "";
    if ( values.firstname[0] && values.lastname[0] ) {
      label = cleanString( values.firstname[0].value ) + " " +
              cleanString( values.lastname[0].value );
      if ( values.birthyear[0] ) {
        label += " ("+ cleanString( values.birthyear[0].value );
        if ( values.deathyear[0] ) {
          label += "-" + cleanString( values.deathyear[0].value ) + ")";
        } else {
          label += ")";
        }
      }
    }
    return '"' + label + '"';
  },
  "searchLabel": function(values) {
    var label = "";
    if ( values.firstname[0] && values.lastname[0] ) {
      label = cleanString( values.firstname[0].value ) + " " +
              cleanString( values.lastname[0].value );
    }
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
