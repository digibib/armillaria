// +build ignore

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
	qAll = `SELECT DISTINCT ?res WHERE { ?res <armillaria://internal/profile> "%s" } OFFSET %d LIMIT %d`
	qOne = `SELECT ?searchLabel, ?displayLabel, ?published
            WHERE {
               <%s> <armillaria://internal/searchLabel> ?searchLabel ;
                    <armillaria://internal/published> ?published ;
                    <armillaria://internal/displayLabel> ?displayLabel .
            }`
	head = `
{ "index" : { "_index" : "public", "_type" : "%s" } }
`
)

type Resource struct {
	Uri          string `json:"uri"`
	DisplayLabel string `json:"displayLabel"`
	SearchLabel  string `json:"searchLabel"`
	Published    string `json:"published"`
}

func main() {
	offset := flag.Int("o", 0, "offset")
	limit := flag.Int("l", 10000, "limit")
	resType := flag.String("t", "", "resource type to index (the value of the <armillaria://internal/profile> predicate.)")

	flag.Parse()

	if resType == nil || *resType == "" {
		log.Println("Missing parameters:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	db := newLocalRDFStore("http://localhost:8890/sparql-auth", "dba", "dba")
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

	resource := Resource{}
	bulkHead := []byte(string(fmt.Sprintf(head, *resType)))
	for i, r := range res.Results.Bindings {
		fmt.Printf("%d resources processed\r", i)
		rb, err := db.Query(fmt.Sprintf(qOne, r["res"].Value))
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(rb, &res)
		resource.Uri = r["res"].Value
		resource.DisplayLabel = res.Results.Bindings[0]["displayLabel"].Value
		resource.SearchLabel = res.Results.Bindings[0]["searchLabel"].Value
		resource.Published = res.Results.Bindings[0]["published"].Value

		_, err = f.Write(bulkHead)
		if err != nil {
			log.Fatal(err)
		}

		rm, err := json.Marshal(resource)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write(rm)
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
