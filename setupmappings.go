// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func urlify(s string) string { return fmt.Sprintf("<%s>", s) }

func main() {
	es := os.Getenv("ELASTICSEARCH_PORT_9200_TCP_ADDR")

	// clear any old indices
	// curl -XDELETE http://172.17.0.20:9200/public
	req, err := http.NewRequest(
		"DELETE",
		"http://"+es+":9200/public/",
		nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// setup analyzers
	//curl -XPUT http://172.17.0.20:9200/public -d @data/es_settings.json
	mb, err := ioutil.ReadFile("data/es_settings.json")
	if err != nil {
		log.Fatal(err)
	}
	req, err = http.NewRequest(
		"PUT",
		"http://"+es+":9200/public/",
		bytes.NewReader(mb),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode > 300 || resp.StatusCode < 200 {
		log.Fatal("Failed to setup indexing analyzer")
	}

	// ---
	cb, err := ioutil.ReadFile("data/mappings/_common")
	if err != nil {
		log.Fatal(err)
	}
	var common preMappings
	err = json.Unmarshal(cb, &common)
	if err != nil {
		log.Fatal(err)
	}

	files, err := filepath.Glob("data/mappings/*.json")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		properties := make(map[string]interface{})
		var m preMappings
		for _, v := range common {
			for ki, vi := range v {
				properties[ki] = vi
			}
		}

		b, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range m {
			for ki, vi := range v {
				properties[ki] = vi
			}
		}
		var out []byte

		profile := strings.TrimSuffix(strings.TrimPrefix(f, "data/mappings/"), ".json")
		out, err = json.Marshal(properties)
		var export bytes.Buffer
		export.WriteString("{\"")
		export.WriteString(profile)
		export.WriteString("\":{\"_id\": {\"path\": \"uri\"},\"properties\":")
		export.Write(out)
		export.WriteString("}}")
		//fmt.Println(string(export.Bytes()))

		req, err := http.NewRequest(
			"PUT",
			"http://"+es+":9200/public/"+profile+"/_mapping",
			bytes.NewReader(export.Bytes()))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode > 300 || resp.StatusCode < 200 {
			fmt.Printf("ERROR setting mappings for /public/%s\n", profile)
			fmt.Println(string(b))
		} else {
			fmt.Printf("OK set mappings for /public/%s\n", profile)
		}
	}

}
