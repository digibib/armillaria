// +build ignore

/*

Basic script to index rdf data, using the mappings in data/mappings

To index all of type work, run:
go run indexrdf.go rdfstore.go indexing.go -t=work

By default 10,000 uris are fetched at a time
You can set ofsett & limit with -o & -l:
go run indexrdf.go rdfstore.go indexing.go -t=work -o=10000, -l=5000

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/digibib/armillaria/sparql"
)

const (
	qAll          = `SELECT DISTINCT ?res WHERE { ?res <armillaria://internal/profile> "%s" } OFFSET %d LIMIT %d`
	resourceQuery = `
SELECT * FROM <%s> WHERE {
   { <%s> ?p ?o .
     MINUS { <%s> ?p ?o . ?o <armillaria://internal/displayLabel> _:l . } }
   UNION
   { <%s> ?p ?o .
     ?o <armillaria://internal/displayLabel> ?l . }
}`
	head = `
{ "index" : { "_index" : "public", "_type" : "%s" } }
`
)

var (
	db = newLocalRDFStore("http://localhost:8890/sparql-auth", "dba", "dba")
)

func urlify(s string) string { return fmt.Sprintf("<%s>", s) }

func main() {
	offset := flag.Int("o", 0, "offset")
	limit := flag.Int("l", 10000, "limit")
	resType := flag.String("t", "", "resource type to index (the value of the <armillaria://internal/profile> predicate.)")
	graph := flag.String("g", "http://data.deichman.no/public", "graph from rdfstore to index from")

	flag.Parse()

	if resType == nil || *resType == "" {
		log.Println("Missing parameters:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	indexMappings, err := loadFromProfiles()
	if err != nil {
		log.Fatal(err)
	}

	b, err := db.Query(fmt.Sprintf(qAll, *resType, *offset, *limit))
	if err != nil {
		log.Fatal(err)
	}

	var res sparql.Results
	err = json.Unmarshal(b, &res)
	if err != nil {
		println(string(b))
		log.Fatal(err)
	}

	f, err := os.Create("out.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bulkHead := []byte(string(fmt.Sprintf(head, *resType)))
	for i, r := range res.Results.Bindings {
		fmt.Printf("%d resources processed\r", i)
		uri := r["res"].Value
		rb, err := db.Query(fmt.Sprintf(resourceQuery, graph, uri, uri, uri))
		if err != nil {
			log.Fatal(err)
		}

		resourceBody, _, err := createIndexDoc(indexMappings, rb, uri)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write(bulkHead)
		if err != nil {
			log.Fatal(err)
		}

		_, err = f.Write(resourceBody)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = f.Write([]byte("\n"))
	if err != nil {

		log.Fatal(err)
	}
	fmt.Println("\nTo index run:")
	fmt.Println("curl -s -XPOST localhost:9200/_bulk --data-binary @out.json")

}
