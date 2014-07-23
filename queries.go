package main

// Bank of all SPARQL queries used in the armillaria server:
const sparqlQueries = `
# tag: lastUpdated
PREFIX armillaria: <armillaria://internal/>
SELECT ?updated, ?kohaID
WHERE {
	{{.Res}} armillaria:updated ?updated
	OPTIONAL { {{.Res}} armillaria:kohaID ?kohaID }
}

# tag: hasDependant
PREFIX armillaria: <armillaria://internal/>
SELECT ?updated, ?kohaID, ?dependant
WHERE {
	{{.Res}} armillaria:updated ?updated
	OPTIONAL { {{.Res}} armillaria:kohaID ?kohaID }
	OPTIONAL { ?dependant _:p {{.Res}} }
} LIMIT 3

# tag: resource
PREFIX armillaria: <armillaria://internal/>
SELECT *
WHERE {
   { GRAPH <{{.Graph}}> {
     {{.Res}} ?p ?o .
     MINUS { {{.Res}} ?p ?o . ?o armillaria:displayLabel _:l . } } }
   UNION
   { {{.Res}} ?p ?o .
     ?o armillaria:displayLabel ?l . }
}

# tag: affectedResources
PREFIX armillaria: <armillaria://internal/>
SELECT ?resource, ?kohaID
FROM <{{.Graph}}>
WHERE {
	      { ?resource _:p {{.Res}}; armillaria:kohaID ?kohaID }
	UNION { {{.Res}} _:p ?resource . ?resource armillaria:kohaID ?kohaID }
	?resource armillaria:profile "manifestation" .
} LIMIT 100

# tag: insertKohaID
PREFIX armillaria: <armillaria://internal/>
WITH <{{.Graph}}>
DELETE { {{.Res}} armillaria:updated ?updated }
INSERT { {{.Res}} armillaria:updated ?now ;
                  armillaria:kohaID {{.KohaID}} }
WHERE {
	OPTIONAL { {{.Res}} armillaria:updated ?updated } .
    BIND( now() AS ?now )
}

# tag: rdf2marc
# queryRDF2MARC is the SPARQL query used to fetch the values
# needed for converting into MARC. This is only done on resources
# with type fabio:Manifestation.
# The binding variables indicates the MARC destination datafield
# and subfield for a given value. For example:
#   ?245_b   -> 245$b
#   ?c008_22 -> controlfield 008, position 22

PREFIX armillaria: <armillaria://internal/>
PREFIX radatana:   <http://def.bibsys.no/xmlns/radatana/1.0#>
PREFIX foaf:       <http://xmlns.com/foaf/0.1/>
PREFIX deich:      <http://data.deichman.no/>
PREFIX fabio:      <http://purl.org/spar/fabio/>
PREFIX dct:        <http://purl.org/dc/terms/>
PREFIX dbo:        <http://dbpedia.org/ontology/>
PREFIX bibo:       <http://purl.org/ontology/bibo/>
PREFIX gn:         <http://www.geonames.org/ontology#>

SELECT *
FROM <{{.Graph}}>
WHERE {
	{{.Res}} dct:title ?245_a .
    BIND({{.Res}} AS ?r)
    ?r armillaria:profile ?profile .
    OPTIONAL { ?r fabio:hasSubtitle ?245_b }
    OPTIONAL { ?r bibo:isbn ?020_a }
    OPTIONAL { ?r dct:format ?019_b }
    OPTIONAL { ?r deich:bindingInfo ?020_b }
    OPTIONAL { ?r deich:literaryFormat ?019_d
               FILTER(?019_d != <http://dbpedia.org/resource/Fiction>) }
    OPTIONAL { ?r deich:location_format ?090_b }
    OPTIONAL { ?r dct:creator _:creator .
                   _:creator foaf:name ?245_c ;
                   radatana:catalogueName ?100_a .
    OPTIONAL { _:creator deich:lifespan ?100_d }
               OPTIONAL { _:creator dbo:nationality ?100_j } }
    OPTIONAL { ?r deich:publicationPlace _:pubPlace .
               _:pubPlace gn:name ?260_a }
    OPTIONAL { ?r bibo:issuer _:issuer .
               _:issuer foaf:name ?260_b }
    OPTIONAL { ?r fabio:hasPublicationYear ?260_c }
    OPTIONAL { ?r bibo:numPages ?300_a }
    OPTIONAL { ?r dct:description ?300_b }

    OPTIONAL { ?r deich:literaryFormat ?c008_33 .
                FILTER(?c008_33 = <http://dbpedia.org/resource/Fiction> || ?c008_33 = <http://dbpedia.org/resource/Non-Fiction> ) }
    OPTIONAL { ?r dct:audience ?c008_22 }
    OPTIONAL { ?r dct:language ?c008_35 }
    OPTIONAL { ?r dct:identifier ?c001_0 }
}
`
