var profile = {
  "overview": {
    "title": "Person",
    "desc": "Beskriver en historisk eller nålevende person, aka homo sapiens",
    "type": ["<http://xmlns.com/foaf/0.1/Person>"]
  },
  "views": [
    {
      "title": "Personopplysninger",
      "desc": "Her fyller du inn personalia. Sett i gang.",
      "elements": [
        {
          "id": "name",
          "label": "Navn",
          "desc": "Fylles inn på katalogform (siste etternavn, andre navn). Normalform genereres automatisk.",
          "required": true,
          "repeatable": false,
          "predicate": "<http://def.bibsys.no/xmlns/radatana/1.0#catalogueName>",
          "type": "string",
          "dependant": "nameNormalized",
          "dependantTransform": function( value ) {
            var name = cleanString( value );
            if ( name ) {
              var last, first;
              var split = name.match(/^(.*),\s?(.*)$/);
              if ( split && split.length == 3) {
                return '"' + split[2] + ' ' + split[1] + '"';
              }
              return name
            }
            return false
          }
        },
        {
          "id": "nameNormalized",
          "label": "Navn (normalisert)",
          "desc": "Genereres automatisk, men kan endres",
          "required": false,
          "repeatable": false,
          "predicate": "<http://xmlns.com/foaf/0.1/name>",
          "type": "string",
          "dependsOn": "name"
        },
        {
          "id": "birthyear",
          "label": "Fødselsår",
          "desc": "",
          "required": false,
          "repeatable": false,
          "predicate": "<http://dbpedia.org/ontology/birthYear>",
          "type": "integer"
        },
        {
          "id": "deathyear",
          "label": "Dødsår",
          "desc":"",
          "required": false,
          "repeatable": false,
          "predicate": "<http://dbpedia.org/ontology/deathYear>",
          "type": "integer"
        },
        {
          "id": "nationality",
          "label": "Nasjonalitet",
          "desc": "En person kan ha flere nasjonaliteter. Fyll inn alle du vet om.",
          "required": false,
          "repeatable": true,
          "predicate": "<http://dbpedia.org/ontology/nationality>",
          "type": "URI",
          "searchTypes": ["location"]
        },
                {
          "id": "tittel",
          "label": "Tittel",
          "desc": "Kongelige titler",
          "required": false,
          "repeatable": true,
          "hidden": true,
          "predicate": "<http://xmlns.com/foaf/0.1/title>",
          "type": "string"
        },
        {
          "id": "nummer",
          "label": "Nummer",
          "desc": "Ordensnummer for kongelige, paver osv.",
          "required": false,
          "repeatable": false,
          "hidden": true,
          "predicate": "<http://data.deichman.no/ordinal>",
          "type": "string"
        },
        {
          "id": "pseudo",
          "label": "Pseudonym",
          "desc": "Også kjent som...",
          "required": false,
          "repeatable": true,
          "hidden": true,
          "predicate": "<http://data.deichman.no/hasPseudonym>",
          "type": "URI",
          "searchTypes": ["person"]
        }
      ]
    }
  ],
  "displayLabel": function( values ) {
    var label = cleanString( values.nameNormalized[0].value );
    if ( values.birthyear[0] ) {
      label += " ("+ cleanString( values.birthyear[0].value ) + "-";
      if ( values.deathyear[0] ) {
        label += cleanString( values.deathyear[0].value ) + ")";
      } else {
        label += ")";
      }
    }
    return '"' + label + '"';
  },
  "searchLabel": function(values) {
    var label = cleanString( values.nameNormalized[0].value );
    return '"' + label + '"';
  },
  "rules": [
    "SPARQL ditt",
    "SPARQL datt"
  ]
};
